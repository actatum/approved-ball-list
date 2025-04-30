// Package usbc provides an implementation of the USBCClient using the USBC json API.
package usbc

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/actatum/approved-ball-list/internal/balls"
	"github.com/rs/zerolog"
)

const layoutUS = "January 2, 2006"

const (
	ballListURL = "https://bowl.com/api/approvedballs?brandName="
	noImageURL  = "https://images.bowl.com/bowl/media/legacy/internap/bowl/equipandspecs/images/approvedballs/noimage.jpg"
)

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
	cfg.HTTPClient.Timeout = 10 * time.Second
	return &Client{
		client: cfg.HTTPClient,
		logger: cfg.Logger,
	}
}

// ListBalls lists balls from the USBC approved ball list by brand.
func (c *Client) ListBalls(ctx context.Context, brand balls.Brand) ([]balls.Ball, error) {
	brandKey := base64.URLEncoding.EncodeToString([]byte(brand))
	r, err := http.NewRequestWithContext(ctx, "GET", ballListURL+brandKey, nil)
	if err != nil {
		return nil, fmt.Errorf("creating http request: %w", err)
	}

	resp, err := c.client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("making http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status: %d", resp.StatusCode)
	}

	var items []ball
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	result := make([]balls.Ball, 0, len(items))
	for _, i := range items {
		if i.Name == "" || i.DateApproved == "" {
			continue
		}

		i.Brand = strings.TrimSpace(i.Brand)
		i.Name = strings.TrimSpace(i.Name)
		i.ImageURL = strings.TrimSpace(i.ImageURL)
		i.DateApproved = strings.TrimSpace(i.DateApproved)

		var approvedAt time.Time
		approvedAt, err = parseDate(i.DateApproved)
		if err != nil {
			return nil, fmt.Errorf("parsing date: %s: %w", i.DateApproved, err)
		}

		if strings.Contains(i.ImageURL, "getmedia") {
			i.ImageURL = fixImageURL(i.ImageURL)
		}

		parsedURL, err := url.Parse(i.ImageURL)
		if err != nil {
			parsedURL, _ = url.Parse(noImageURL)
		}
		result = append(result, balls.Ball{
			Brand:        balls.Brand(i.Brand),
			Name:         i.Name,
			ApprovalDate: approvedAt,
			ImageURL:     parsedURL,
		})
	}

	return result, nil
}

// Close shuts down idle connections
func (c *Client) Close() {
	c.client.CloseIdleConnections()
}

func (c *Client) writeToJSONFile(balls []balls.Ball) error {
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
	}

	return time.Time{}, fmt.Errorf("unexpected date format: %s", date)
}
