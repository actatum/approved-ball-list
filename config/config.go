package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// AppConfig is the configuration for the entire application
type AppConfig struct {
	DiscordToken              string `split_words:"true"`
	GCPProject                string `split_words:"true"`
	PandapackChannelID        string `split_words:"true"`
	MotivatedChannelID        string `split_words:"true"`
	BrunswickCentralChannelID string `split_words:"true"`
	PersonalChannelID         string `split_words:"true"`
}

// NewAppConfig returns a new app config loaded from the environment
func NewAppConfig() (AppConfig, error) {
	var c AppConfig
	if err := envconfig.Process("", &c); err != nil {
		return AppConfig{}, fmt.Errorf("envconfig.Process: %w", err)
	}

	return c, nil
}
