package usbc

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/actatum/approved-ball-list/abl"
	"github.com/rs/zerolog"
)

const layoutUS = "January 2, 2006"

const ballListURL = "https://www.bowl.com/approvedballlist/netballs.xml"

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
	XMLName      xml.Name `xml:"Brand"`
	Brand        string   `xml:"name,attr"`
	Name         string   `xml:"BallName"`
	DateApproved string   `xml:"DateApproved"`
	ImageURL     string   `xml:"link"`
}

type ballList struct {
	XMLName xml.Name `xml:"BallList"`
	Balls   []ball   `xml:"Brand"`
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
	req, err := http.NewRequestWithContext(ctx, "GET", ballListURL, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client.Do: %w", err)
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			c.logger.Warn().Err(closeErr).Msg("error closing response body")
		}
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}

	var list ballList
	err = xml.Unmarshal(data, &list)
	if err != nil {
		return nil, fmt.Errorf("xml.Unmarshal: %w", err)
	}

	list.Balls = filterBrands(list.Balls)

	result := make([]abl.Ball, 0, len(list.Balls))
	for i := range list.Balls {
		var approvedAt time.Time
		approvedAt, err = parseDate(list.Balls[i].DateApproved)
		if err != nil {
			return nil, err
		}

		result = append(result, abl.Ball{
			Brand:      list.Balls[i].Brand,
			Name:       list.Balls[i].Name,
			ApprovedAt: approvedAt,
			ImageURL:   list.Balls[i].ImageURL,
		})
	}

	return result, nil
}

// Close shuts down idle connections
func (c *Client) Close() {
	c.client.CloseIdleConnections()
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

func parseDate(date string) (time.Time, error) {
	date = strings.ReplaceAll(date, "'", "") // Remove apostrophes

	if strings.Contains(date, ",") {
		t, err := time.Parse(layoutUS, date)
		if err != nil {
			fmt.Println(err)
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
