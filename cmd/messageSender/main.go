package p

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/actatum/approved-ball-list/alerter"
	"github.com/actatum/approved-ball-list/config"
	"github.com/actatum/approved-ball-list/core"
	"github.com/actatum/approved-ball-list/log"
	"github.com/actatum/approved-ball-list/repository"
	"github.com/actatum/approved-ball-list/usbc"
	"go.uber.org/zap"
)

// FirestoreEvent is the payload of a Firestore event.
type FirestoreEvent struct {
	OldValue   FirestoreValue `json:"oldValue"`
	Value      FirestoreValue `json:"value"`
	UpdateMask struct {
		FieldPaths []string `json:"fieldPaths"`
	} `json:"updateMask"`
}

// FirestoreValue holds Firestore fields.
type FirestoreValue struct {
	CreateTime time.Time `json:"createTime"`
	// Fields is the data for this value. The type depends on the format of your
	// database. Log the interface{} value and inspect the result to see a JSON
	// representation of your database fields.
	Fields     core.Ball `json:"fields"`
	Name       string    `json:"name"`
	UpdateTime time.Time `json:"updateTime"`
}

var svc core.Service

func init() {
	logger, err := log.NewLogger()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cfg, err := config.NewAppConfig()
	if err != nil {
		logger.Fatal("failed to initialize app config", zap.Error(err))
	}

	a, err := alerter.NewAlerter(cfg.DiscordToken)
	if err != nil {
		logger.Fatal("failed to initialize alerter", zap.Error(err))
	}

	usbcClient := usbc.NewClient(&usbc.Config{
		Logger:     logger,
		HTTPClient: nil,
	})

	repo, err := repository.NewRepository(context.Background(), cfg.GCPProjectID)
	if err != nil {
		logger.Fatal("failed to initialize repository", zap.Error(err))
	}

	svc = core.NewService(&core.Config{
		Logger: logger,
		DiscordChannels: map[string]core.DiscordChannel{
			"motivated": {
				Name:   "motivated",
				ID:     cfg.Channels["motivated"],
				Brands: []string{"Motiv"},
			},
			"panda-pack": {
				Name:   "panda-pack",
				ID:     cfg.Channels["panda-pack"],
				Brands: []string{"Storm", "Roto Grip", "900 Global"},
			},
			"brunswick-central": {
				Name:   "brunswick-central",
				ID:     cfg.Channels["brunswick-central"],
				Brands: []string{"Brunswick", "Columbia", "DV8", "Ebonite", "Hammer", "Radical", "Track"},
			},
			"personal": {
				Name:   "personal channel",
				ID:     cfg.Channels["personal"],
				Brands: []string{"900 Global", "BIG Bowling", "Brunswick", "Columbia", "DV8", "Ebonite", "Hammer", "Motiv", "Radical", "Roto Grip", "Storm", "Track Inc."},
			},
		},
		Repository: repo,
		Alerter:    a,
		USBC:       usbcClient,
	})
}

// MessageSender is the entry point for the message sender cloud function.
// This function receives events when a new entry is added to the database and sends messages to the corresponding channels
func MessageSender(ctx context.Context, e FirestoreEvent) error {
	return svc.AlertNewBall(ctx, e.Value.Fields)
}
