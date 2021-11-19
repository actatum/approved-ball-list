package usbc

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/actatum/approved-ball-list/models"
)

const ballListURL = "https://www.bowl.com/approvedballlist/netballs.xml"

func GetBalls(ctx context.Context) ([]models.Ball, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", ballListURL, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client.Do: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}

	var balls models.BallList
	err = xml.Unmarshal(data, &balls)
	if err != nil {
		return nil, fmt.Errorf("xml.Unmarshal: %w", err)
	}
	balls.Filter()

	return balls.Balls, nil
}
