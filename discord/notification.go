package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/actatum/approved-ball-list/abl"
	"github.com/bwmarrin/discordgo"
)

// NotificationService represents a service for managing distribution of notifications.
type NotificationService struct {
	dg       *discordgo.Session
	channels []string
}

// NewNotificationService returns a new instance of NotificationService.
func NewNotificationService(dg *discordgo.Session, channels []string) *NotificationService {
	return &NotificationService{
		dg:       dg,
		channels: channels,
	}
}

// SendNotifications sends notifications to the configured channels.
func (s *NotificationService) SendNotifications(ctx context.Context, n ...abl.Notification) error {
	embeds := make([]*discordgo.MessageEmbed, 0, len(n))
	for i := range n {
		e := &discordgo.MessageEmbed{
			Type:  discordgo.EmbedTypeImage,
			Title: fmt.Sprintf("%s %s", n[i].Ball.Brand, n[i].Ball.Name),
			Image: &discordgo.MessageEmbedImage{
				URL: n[i].Ball.ImageURL,
			},
		}

		embeds = append(embeds, e)
	}

	if len(embeds) == 0 {
		return nil
	}

	for _, id := range s.channels {
		_, err := s.dg.ChannelMessageSendEmbeds(id, embeds)
		if err != nil {
			if !strings.Contains(err.Error(), "405") {
				return fmt.Errorf("dg.ChannelMessageSendEmbeds")
			}
		}
	}

	return nil
}
