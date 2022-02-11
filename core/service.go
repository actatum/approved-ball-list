package core

import (
	"context"

	"github.com/rs/zerolog"
)

// Service is the interface for the business logic
type Service interface {
	FilterAndAddBalls(ctx context.Context) error
	AlertNewBall(ctx context.Context, ball Ball) error
}

type service struct {
	alerter         Alerter
	repository      Repository
	usbc            USBC
	logger          *zerolog.Logger
	discordChannels map[string]DiscordChannel
}

// NewService returns a new service with the given config
func NewService(cfg *Config) Service {
	return &service{
		logger:          cfg.Logger,
		discordChannels: cfg.DiscordChannels,
		repository:      cfg.Repository,
		alerter:         cfg.Alerter,
		usbc:            cfg.USBC,
	}
}

func (s service) FilterAndAddBalls(ctx context.Context) error {
	fromRepo := make(chan RetrieveBallResult)
	fromUSBC := make(chan RetrieveBallResult)

	go func(ch chan RetrieveBallResult) {
		balls, err := s.repository.GetAllBalls(ctx)
		ch <- RetrieveBallResult{Balls: balls, Err: err}
	}(fromRepo)

	go func(ch chan RetrieveBallResult) {
		balls, err := s.usbc.GetApprovedBallList(ctx)
		ch <- RetrieveBallResult{Balls: balls, Err: err}
	}(fromUSBC)

	repoResult := <-fromRepo
	if repoResult.Err != nil {
		return repoResult.Err
	}
	s.logger.Info().Msgf("number of balls in database: %d", len(repoResult.Balls))

	usbcResult := <-fromUSBC
	if usbcResult.Err != nil {
		return usbcResult.Err
	}
	s.logger.Info().Msgf("number of approved balls from USBC: %d", len(usbcResult.Balls))

	filteredBalls := s.filter(repoResult.Balls, usbcResult.Balls)
	s.logger.Info().Msgf("number of newly approved balls: %d", len(filteredBalls))

	if err := s.repository.InsertNewBalls(ctx, filteredBalls); err != nil {
		return err
	}

	return nil
}

func (s service) AlertNewBall(ctx context.Context, ball Ball) error {
	var channelIDs []string
	for _, channel := range s.discordChannels {
		if channel.containsBrand(ball.Brand) {
			channelIDs = append(channelIDs, channel.ID)
		}
	}
	return s.alerter.SendMessage(ctx, channelIDs, ball)
}

func (s service) filter(fromRepo []Ball, fromUSBC []Ball) []Ball {
	var unique []Ball

	for _, b := range fromUSBC { // each ball from the usbc
		found := false
		for i := 0; i < len(fromRepo); i++ {
			if b.Name == fromRepo[i].Name && b.Brand == fromRepo[i].Brand {
				found = true
				break
			}
		}
		if !found {
			unique = append(unique, b)
		}
	}

	return unique
}
