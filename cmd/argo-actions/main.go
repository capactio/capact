package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"capact.io/capact/internal/logger"
	argoactions "capact.io/capact/pkg/argo-actions"
	"capact.io/capact/pkg/hub/client/local"
	"capact.io/capact/pkg/hub/client/public"

	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
)

// Config for the argo-actions command.
type Config struct {
	Action            string
	DownloadConfig    []argoactions.DownloadConfig `envconfig:"optional"`
	UploadConfig      argoactions.UploadConfig     `envconfig:"optional"`
	UpdateConfig      argoactions.UpdateConfig     `envconfig:"optional"`
	LocalHubEndpoint  string                       `envconfig:"default=http://capact-hub-local.capact-system/graphql"`
	PublicHubEndpoint string                       `envconfig:"default=http://capact-hub-public.capact-system/graphql"`
	Logger            logger.Config
}

func main() {
	start := time.Now()

	var cfg Config
	var action argoactions.Action

	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading configuration")

	// setup logger
	logger, err := logger.New(cfg.Logger)
	exitOnError(err, "while creating zap logger")

	localClient := local.NewDefaultClient(cfg.LocalHubEndpoint)
	publicClient := public.NewDefaultClient(cfg.PublicHubEndpoint)

	switch cfg.Action {
	case argoactions.DownloadAction:
		log := logger.With(zap.String("Action", argoactions.DownloadAction))
		action = argoactions.NewDownloadAction(log, localClient, cfg.DownloadConfig)

	case argoactions.UploadAction:
		log := logger.With(zap.String("Action", argoactions.UploadAction))
		action = argoactions.NewUploadAction(log, localClient, publicClient, cfg.UploadConfig)

	case argoactions.UpdateAction:
		log := logger.With(zap.String("Action", argoactions.UpdateAction))
		action = argoactions.NewUpdateAction(log, localClient, cfg.UpdateConfig)

	default:
		err := fmt.Errorf("Invalid action: %s", cfg.Action)
		exitOnError(err, "while selecting action")
	}

	ctx := context.Background()
	err = action.Do(ctx)
	exitOnError(err, "while executing action")

	// Argo doesn't like when a Pod exits too fast
	minTime := start.Add(time.Second)
	if time.Now().Before(minTime) {
		time.Sleep(time.Second)
	}
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
