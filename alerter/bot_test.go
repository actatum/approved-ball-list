package alerter

import (
	"context"
	"errors"
	"testing"

	"github.com/actatum/approved-ball-list/core"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

type mockDiscordSession struct {
	wantErr bool
}

func (s *mockDiscordSession) ChannelMessageSendEmbed(channelID string, embed *discordgo.MessageEmbed) (*discordgo.Message, error) {
	if s.wantErr {
		return nil, errors.New("message send error")
	}

	return &discordgo.Message{
		ID:        "123",
		ChannelID: channelID,
	}, nil
}

func (s *mockDiscordSession) Close() error {
	return nil
}

func TestAlerter_SendMessage(t *testing.T) {
	type fields struct {
		dg discordSession
	}
	type args struct {
		ctx        context.Context
		channelIDs []string
		ball       core.Ball
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "send embedded discord message",
			fields: fields{
				&mockDiscordSession{wantErr: false},
			},
			args: args{
				ctx:        context.Background(),
				channelIDs: []string{"900"},
				ball: core.Ball{
					Brand: "900 Global",
					Name:  "Altered Reality",
				},
			},
			wantErr: false,
		},
		{
			name: "send message error",
			fields: fields{
				&mockDiscordSession{wantErr: true},
			},
			args: args{
				ctx:        context.Background(),
				channelIDs: []string{"900"},
				ball: core.Ball{
					Brand: "Motiv",
					Name:  "Venom Shock",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Alerter{
				dg: tt.fields.dg,
			}

			t.Cleanup(func() {
				closeErr := a.Close()
				if closeErr != nil {
					t.Fatalf("failed to close alerter: %v", closeErr)
				}
			})

			err := a.SendMessage(tt.args.ctx, tt.args.channelIDs, tt.args.ball)

			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
		})
	}
}

func TestNewAlerter(t *testing.T) {
	type args struct {
		token string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "new alerter",
			args: args{
				token: "test-token",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAlerter(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAlerter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
