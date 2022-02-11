package usbc

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/actatum/approved-ball-list/core"
	"github.com/rs/zerolog"
)

const ballListURL = "https://www.bowl.com/approvedballlist/netballs.xml"

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
func (c *Client) GetApprovedBallList(ctx context.Context) ([]core.Ball, error) {
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

	var balls core.BallList
	err = xml.Unmarshal(data, &balls)
	if err != nil {
		return nil, fmt.Errorf("xml.Unmarshal: %w", err)
	}

	return c.filterBallList(balls), nil
}

// Close shuts down idle connections
func (c *Client) Close() {
	c.client.CloseIdleConnections()
}

func (c *Client) filterBallList(ballList core.BallList) []core.Ball {
	n := 0
	for _, ball := range ballList.Balls {
		if _, ok := core.CurrentBrands[ball.Brand]; ok {
			ballList.Balls[n] = ball
			n++
		}
	}

	return ballList.Balls[:n]
}

func (c *Client) writeToJSONFile(balls []core.Ball) error {
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
