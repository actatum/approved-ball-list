package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// AppConfig is the configuration for the entire application
type AppConfig struct {
	Channels     map[string]string
	DiscordToken string `split_words:"true"`
	GCPProjectID string `split_words:"true"`
}

// NewAppConfig returns a new app config loaded from the environment
func NewAppConfig() (AppConfig, error) {
	var c AppConfig
	if err := envconfig.Process("", &c); err != nil {
		return AppConfig{}, fmt.Errorf("envconfig.Process: %w", err)
	}

	return c, nil
}
