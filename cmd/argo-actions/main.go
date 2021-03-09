package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/machinebox/graphql"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	argoactions "projectvoltron.dev/voltron/pkg/argo-actions"
	"projectvoltron.dev/voltron/pkg/httputil"
	"projectvoltron.dev/voltron/pkg/och/client/local/v2"
)

type Config struct {
	Action           string
	DownloadConfig   []argoactions.DownloadConfig `envconfig:"optional"`
	UploadConfig     argoactions.UploadConfig     `envconfig:"optional"`
	UpdateConfig     argoactions.UpdateConfig     `envconfig:"optional"`
	LocalOCHEndpoint string                       `envconfig:"default=http://voltron-och-local.voltron-system/graphql"`
	LoggerDevMode    bool                         `envconfig:"default=false"`
}

func main() {
	start := time.Now()

	var cfg Config
	var action argoactions.Action

	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading configuration")

	// setup logger
	var logCfg zap.Config
	if cfg.LoggerDevMode {
		logCfg = zap.NewDevelopmentConfig()
	} else {
		logCfg = zap.NewProductionConfig()
	}

	logger, err := logCfg.Build()
	exitOnError(err, "while creating zap logger")

	client := NewOCHLocalClient(cfg.LocalOCHEndpoint)

	switch cfg.Action {
	case argoactions.DownloadAction:
		log := logger.With(zap.String("Action", argoactions.DownloadAction))
		action = argoactions.NewDownloadAction(log, client, cfg.DownloadConfig)

	case argoactions.UploadAction:
		log := logger.With(zap.String("Action", argoactions.UploadAction))
		action = argoactions.NewUploadAction(log, client, cfg.UploadConfig)

	case argoactions.UpdateAction:
		log := logger.With(zap.String("Action", argoactions.UpdateAction))
		action = argoactions.NewUpdateAction(log, client, cfg.UpdateConfig)

	default:
		err := fmt.Errorf("Invalid action: %s", cfg.Action)
		exitOnError(err, "while selecting action")
	}

	ctx := context.Background()
	err = action.Do(ctx)
	exitOnError(err, "while executing action")

	// Argo doesn't like when a Pod exits too fast
	// See https://cshark.atlassian.net/browse/SV-236
	minTime := start.Add(time.Second)
	if time.Now().Before(minTime) {
		time.Sleep(time.Second)
	}
}

func NewOCHLocalClient(endpoint string) *local.Client {
	httpClient := httputil.NewClient(
		30*time.Second,
		true,
	)
	clientOpt := graphql.WithHTTPClient(httpClient)
	client := graphql.NewClient(endpoint, clientOpt)

	return local.NewClient(client)
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
