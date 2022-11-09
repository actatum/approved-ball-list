// Package usbc provides an implementation of the USBCClient using the USBC json API.
package usbc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/actatum/approved-ball-list/internal/abl"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

const layoutUS = "January 2, 2006"

const ballListURL = "https://bowl.com/api/approvedballs?brandName="

var monthMap = map[string]int{
	"Jan":       1,
	"January":   1,
	"Feb":       2,
	"February":  2,
	"Mar":       3,
	"March":     3,
	"Apr":       4,
	"April":     4,
	"May":       5,
	"Jun":       6,
	"June":      6,
	"Jul":       7,
	"July":      7,
	"Aug":       8,
	"August":    8,
	"Sep":       9,
	"Sept":      9,
	"September": 9,
	"Oct":       10,
	"October":   10,
	"Nov":       11,
	"November":  11,
	"Dec":       12,
	"December":  12,
}

type ball struct {
	Brand        string `json:"brandName"`
	Name         string `json:"name"`
	DateApproved string `json:"dateApproved"`
	ImageURL     string `json:"image"`
}

// Client handles interfacing with the usbc approved ball list xml api
type Client struct {
	client *http.Client
	logger *zerolog.Logger
}

// Config is the configuration for the usbc api client
type Config struct {
	Logger     *zerolog.Logger
	HTTPClient *http.Client
}

// NewClient returns a new usbc api client
func NewClient(cfg *Config) *Client {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{}
	}
	return &Client{
		client: cfg.HTTPClient,
		logger: cfg.Logger,
	}
}

// GetApprovedBallList retrieves the current approved ball list from the usbc xml api
func (c *Client) GetApprovedBallList(ctx context.Context) ([]abl.Ball, error) {
	ch := make(chan []abl.Ball, len(abl.ActiveBrands))
	g, gCtx := errgroup.WithContext(ctx)

	for brand := range abl.ActiveBrands {
		b := brand
		g.Go(func() error {
			return c.getBallsByBrand(gCtx, base64.StdEncoding.EncodeToString([]byte(b)), ch)
		})
	}

	var balls []abl.Ball
	for i := 0; i < len(abl.ActiveBrands); i++ {
		b := <-ch
		balls = append(balls, b...)
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return balls, nil
}

// Close shuts down idle connections
func (c *Client) Close() {
	c.client.CloseIdleConnections()
}

func (c *Client) getBallsByBrand(ctx context.Context, brand string, result chan<- []abl.Ball) error {
	req, err := http.NewRequestWithContext(ctx, "GET", ballListURL+brand, nil)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("client.Do: %w", err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			c.logger.Warn().Err(closeErr).Msg("error closing response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	var list []ball
	err = json.Unmarshal(data, &list)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	balls := filterBrands(list)

	res := make([]abl.Ball, 0, len(balls))
	for i := range balls {
		var approvedAt time.Time
		approvedAt, err = parseDate(balls[i].DateApproved)
		if err != nil {
			return err
		}

		if strings.Contains(balls[i].ImageURL, "getmedia") {
			balls[i].ImageURL = fixImageURL(balls[i].ImageURL)
		}

		res = append(res, abl.Ball{
			Brand:      balls[i].Brand,
			Name:       balls[i].Name,
			ApprovedAt: approvedAt,
			ImageURL:   balls[i].ImageURL,
		})
	}

	result <- res
	return nil
}

func (c *Client) writeToJSONFile(balls []abl.Ball) error {
	data, err := json.MarshalIndent(balls, "", "  ")
	if err != nil {
		return fmt.Errorf("json.MarshalIndent: %w", err)
	}

	file, err := os.Create("approvedBalls.json")
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("file.Write: %w", err)
	}

	defer func() {
		fileErr := file.Close()
		if fileErr != nil {
			c.logger.Warn().Err(fileErr).Msg("error closing filer")
		}
	}()

	return nil
}

func filterBrands(balls []ball) []ball {
	filtered := make([]ball, 0)

	for _, b := range balls {
		if _, ok := abl.ActiveBrands[b.Brand]; ok {
			filtered = append(filtered, b)
		}
	}

	return filtered
}

func fixImageURL(dirty string) string {
	trimmed := strings.TrimLeft(dirty, "~")
	return fmt.Sprintf("https://bowl.com%s", trimmed)
}

func parseDate(date string) (time.Time, error) {
	date = strings.ReplaceAll(date, "'", "") // Remove apostrophes

	if strings.Contains(date, ",") {
		t, err := time.Parse(layoutUS, date)
		if err != nil {
			return time.Time{}, fmt.Errorf("time.Parse: %w", err)
		}
		return t, err
	} else if strings.Contains(date, "-") {
		sp := strings.Split(strings.TrimSpace(date), "-")
		if len(sp) != 2 {
			return time.Time{}, fmt.Errorf("invalid month-year combo")
		}

		month, ok := monthMap[strings.TrimSpace(sp[0])]
		if !ok {
			return time.Time{}, fmt.Errorf("invalid month string: %s", sp[0])
		}

		yrString := strings.TrimSpace(sp[1])
		if string(sp[1][0]) == "9" {
			yrString = fmt.Sprintf("%s%s", "19", yrString)
		} else if len(yrString) == 2 {
			yrString = fmt.Sprintf("%s%s", "20", yrString)
		}

		yr, err := strconv.Atoi(yrString)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid year: %s", yrString)
		}

		if yr > 9999 {
			fmt.Println(date, yr)
		}

		return time.Date(yr, time.Month(month), 0, 0, 0, 0, 0, time.UTC), nil
	} else if strings.Contains(date, "00") {
		date = strings.ReplaceAll(date, "00", "")

		month, ok := monthMap[strings.TrimSpace(date)]
		if !ok {
			return time.Time{}, fmt.Errorf("invalid month string: %s", date)
		}

		return time.Date(2000, time.Month(month), 0, 0, 0, 0, 0, time.UTC), nil
	} else {
		return time.Time{}, fmt.Errorf("unexpected date format: %s", date)
	}
}
