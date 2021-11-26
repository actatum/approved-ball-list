package discord

import (
	"fmt"
	"os"

	"github.com/actatum/approved-ball-list/models"
	"github.com/bwmarrin/discordgo"
)

var token = os.Getenv("DISCORD_TOKEN")
var channelID = os.Getenv("CHANNEL_ID")

// SendNewBalls sends the most recently added balls to discord
func SendNewBalls(balls []models.Ball) error {
	ballMap := make(map[string][]models.Ball)
	for _, b := range balls {
		ballMap[b.Brand] = append(ballMap[b.Brand], b)
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		return fmt.Errorf("discordgo.New: %w", err)
	}

	for _, v := range ballMap {
		for _, b := range v {
			_, err := dg.ChannelMessageSendEmbed(channelID, &discordgo.MessageEmbed{
				Title: fmt.Sprintf("%s %s", b.Brand, b.Name),
				Type:  discordgo.EmbedTypeImage,
				Image: &discordgo.MessageEmbedImage{
					URL: b.ImageURL,
				},
			})
			if err != nil {
				return fmt.Errorf("dg.ChannelMessageSendEmbed: %w", err)
			}
		}
	}

	return nil
}
