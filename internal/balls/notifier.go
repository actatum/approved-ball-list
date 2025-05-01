package balls

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// Notifier handles sending notifications of newly approved balls.
//
//go:generate moq -fmt goimports -out notifier_moq_test.go . Notifier
type Notifier interface {
	// Notify notifies configured recipients of newly approved balls.
	Notify(ctx context.Context, approvedBalls []Ball) error
}

// DiscordNotifier implements the Notifier interface and sends notifications of newly approved balls to the
// configured discord channels.
type DiscordNotifier struct {
	dg       *discordgo.Session
	channels []string
}

func NewDiscordNotifier(dg *discordgo.Session, channels []string) *DiscordNotifier {
	return &DiscordNotifier{dg: dg, channels: channels}
}

func (n *DiscordNotifier) Notify(ctx context.Context, approvedBalls []Ball) error {
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

	batches := batchSlice(embeds, 3)

	for _, id := range n.channels {
		for _, batch := range batches {
			if _, err := n.dg.ChannelMessageSendEmbeds(id, batch, discordgo.WithContext(ctx)); err != nil {
				return fmt.Errorf("sending embeds: %w", err)
			}
		}
	}

	return nil
}

func batchSlice[T any](sl []T, batchSize int) [][]T {
	batches := make([][]T, 0)
	for i := 0; i < len(sl); i += batchSize {
		end := i + batchSize
		if end > len(sl) {
			end = len(sl)
		}
		batches = append(batches, sl[i:end])
	}

	return batches
}

// LocalNotifier implements the Notifier interface and prints out newly approved balls to stdout.
type LocalNotifier struct{}

func (n LocalNotifier) Notify(_ context.Context, approvedBalls []Ball) error {
	if len(approvedBalls) == 0 {
		fmt.Println("NOTIFIER: no approved balls to notify")
		return nil
	}

	fmt.Println("NOTIFIER:")
	for _, ball := range approvedBalls {
		fmt.Printf("NEWLY APPROVED BALL: %s %s", ball.Brand, ball.Name)
	}

	return nil
}
