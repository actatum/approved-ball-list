// Package discord provides an implementation of the Notifier interface using discord as the notification medium.
package discord

import (
	"context"
	"fmt"

	"github.com/actatum/approved-ball-list/internal/balls"
	"github.com/bwmarrin/discordgo"
)

// const layoutUS = "January 2, 2006"

// Notifier handles sending notification messages to discord.
type Notifier struct {
	dg       *discordgo.Session
	channels []string
}

// NewNotifier returns a new instance of Notifier.
func NewNotifier(token string, channels []string) (*Notifier, error) {
	dg, err := discordgo.New(fmt.Sprintf("Bot %s", token))
	if err != nil {
		return nil, err
	}

	return &Notifier{
		dg:       dg,
		channels: channels,
	}, nil
}

// Close shuts down the underlying discord client.
func (n *Notifier) Close() error {
	return n.dg.Close()
}

// SendNotification sends notifications to discord.
func (n *Notifier) SendNotification(ctx context.Context, approvedBalls []balls.Ball) error {
	if len(approvedBalls) == 0 {
		return nil
	}

	embeds := make([]*discordgo.MessageEmbed, 0, len(approvedBalls))
	for _, b := range approvedBalls {
		embeds = append(embeds, &discordgo.MessageEmbed{
			Type:  discordgo.EmbedTypeImage,
			Title: fmt.Sprintf("%s %s", b.Brand, b.Name),
			Image: &discordgo.MessageEmbedImage{
				URL: b.ImageURL.String(),
			},
		})
	}

	for _, id := range n.channels {
		if _, err := n.dg.ChannelMessageSendEmbeds(id, embeds); err != nil {
			return fmt.Errorf("sending embeds: %w", err)
		}
	}

	return nil
}
