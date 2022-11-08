// Package main provides the entrypoint to the service when its run as a server.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/actatum/approved-ball-list/internal/app"
)

func main() {
	var cfg app.Config
	cfg.DiscordChannels = strings.Split(lookupEnv("DISCORD_CHANNELS", ""), ",")
	flag.StringVar(&cfg.Env, "env", lookupEnv("ENV", "local"), "environment service is running in")
	flag.StringVar(&cfg.Port, "port", lookupEnv("PORT", "8080"), "http server port")
	flag.StringVar(&cfg.StorageBucket, "storage-bucket", lookupEnv("STORAGE_BUCKET", ""), "gcp storage bucket for backups")
	flag.StringVar(&cfg.DiscordToken, "discord-token", lookupEnv("DISCORD_TOKEN", ""), "discord bot token")
	flag.Var(&cfg.DiscordChannels, "discord-channels", "discord channels to notify")
	flag.Parse()

	application, err := app.NewApplication(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func lookupEnv(key string, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}

	return defaultValue
}
