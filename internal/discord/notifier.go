// Package discord provides an implementation of the Notifier interface using discord as the notification medium.
package discord

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/actatum/approved-ball-list/internal/abl"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const layoutUS = "January 2, 2006"

// Notifier handles sending notification messages to discord.
type Notifier struct {
	dg       *discordgo.Session
	channels []string
}

// NewNotifier returns a new instance of Notifier.
func NewNotifier(token string, channels []string) (*Notifier, error) {
	dg, err := discordgo.New(token)
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

// Notify sends notifications to the configured channels.
func (n *Notifier) Notify(ctx context.Context, notifications []abl.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	embeds := make([]*discordgo.MessageEmbed, 0, len(notifications))
	for _, notif := range notifications {
		e := &discordgo.MessageEmbed{
			Type:  discordgo.EmbedTypeImage,
			Title: fmt.Sprintf("%s %s", notif.Ball.Brand, notif.Ball.Name),
			Image: &discordgo.MessageEmbedImage{
				URL: notif.Ball.ImageURL,
			},
		}

		embeds = append(embeds, e)
	}

	for _, id := range n.channels {
		msgType := notifications[0].Type
		caser := cases.Title(language.AmericanEnglish)
		_, err := n.dg.ChannelMessageSend(
			id,
			fmt.Sprintf("%s Balls: %s", caser.String(string(msgType)), time.Now().Format(layoutUS)),
		)
		if err != nil {
			return fmt.Errorf("dg.ChannelMessageSend: %w", err)
		}
		_, err = n.dg.ChannelMessageSendEmbeds(id, embeds)
		if err != nil {
			if !strings.Contains(err.Error(), "405") {
				return fmt.Errorf("dg.ChannelMessageSendEmbeds: %w", err)
			}
		}
	}

	return nil
}
