package discord

import (
	"fmt"
	"html"

	"github.com/actatum/approved-ball-list/models"
	"github.com/bwmarrin/discordgo"
	"github.com/gomarkdown/markdown"
)

var token = "OTExMDgwODE0ODgyNzI1OTM5.YZcMIQ.kbp-gqqkGdmyNuGbS4vSaKuz6SA"
var channelID = "77164508690264064"

// SendNewBalls sends the most recently added balls to discord
func SendNewBalls(balls []models.Ball) error {
	ballMap := make(map[string][]models.Ball)
	for _, b := range balls {
		ballMap[b.Brand] = append(ballMap[b.Brand], b)
	}

	fmt.Println(ballMap)

	content := ""
	for k, v := range ballMap {
		content += fmt.Sprintf("**%s**", k)
		// ![alt text](https://github.com/adam-p/markdown-here/raw/master/src/common/images/icon48.png "Logo Title Text 1")
		for idx, b := range v {
			content += "\n"
			content += fmt.Sprintf("%d. %s\n", idx+1, b.Name)
			// content += b.ImageURL
			// content += "\n```hi```"
			content += fmt.Sprintf("![%s](%s)", b.Brand+" "+b.Name, b.ImageURL)
			// fmt.Println(b)
		}
	}
	content = html.EscapeString(content)

	fmt.Println(content)

	output := markdown.ToHTML([]byte(content), nil, nil)
	fmt.Println(string(output))

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("discordgo.New: %w", err)
	}

	// msg, err := dg.ChannelMessageSendEmbed(ChannelID, &discordgo.MessageEmbed{
	// 	Title: "New Approved Balls",
	// 	Type:  discordgo.EmbedTypeImage,
	// 	Fields: []*discordgo.MessageEmbedField{
	// 		{
	// 			Name:  "hi",
	// 			Value: "hi",
	// 		},
	// 	},
	// })
	// if err != nil {
	// 	return fmt.Errorf("dg.ChannelMessageSendEmbed: %w", err)
	// }

	msg, err := dg.ChannelMessageSend(channelID, content)
	if err != nil {
		return fmt.Errorf("dg.ChannelMessageSend: %w", err)
	}

	fmt.Println(msg.ID)

	return nil
}
