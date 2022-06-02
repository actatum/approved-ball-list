package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config is the configuration for the entire application
type Config struct {
	CockroachDSN                  string `split_words:"true"`
	DiscordToken                  string `split_words:"true"`
	GCPProject                    string `split_words:"true"`
	USBCApprovedBallListChannelID string `split_words:"true"`
	PersonalChannelID             string `split_words:"true"`
}

// NewConfig returns a new app config loaded from the environment
func NewConfig() (Config, error) {
	var c Config
	if err := envconfig.Process("", &c); err != nil {
		return Config{}, fmt.Errorf("envconfig.Process: %w", err)
	}

	return c, nil
}
