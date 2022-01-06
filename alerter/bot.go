package alerter

import (
	"context"
	"fmt"
	"strings"

	"github.com/actatum/approved-ball-list/core"
	"github.com/bwmarrin/discordgo"
)

// Alerter handles sending messages to discord
type Alerter struct {
	dg discordSession
}

type discordSession interface {
	ChannelMessageSendEmbed(channelID string, embed *discordgo.MessageEmbed) (*discordgo.Message, error)
	Close() error
}

// NewAlerter returns an alerter using the given discord bot token
func NewAlerter(token string) (*Alerter, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("discordgo.New: %w", err)
	}

	return &Alerter{
		dg: dg,
	}, nil
}

// SendMessage sends the message for the new ball to the list of channels provided
func (a *Alerter) SendMessage(ctx context.Context, channelIDs []string, ball core.Ball) error {
	for _, id := range channelIDs {
		_, err := a.dg.ChannelMessageSendEmbed(id, &discordgo.MessageEmbed{
			Title: fmt.Sprintf("%s %s", ball.Brand, ball.Name),
			Type:  discordgo.EmbedTypeImage,
			Image: &discordgo.MessageEmbedImage{
				URL: ball.ImageURL,
			},
		})
		if err != nil {
			if !strings.Contains(err.Error(), "405") {
				return fmt.Errorf("dg.ChannelMessageSendEmbed: %w", err)
			}
		}
	}

	return nil
}

// Close shuts down the underlying discordgo session
func (a *Alerter) Close() error {
	return a.dg.Close()
}
