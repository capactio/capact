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
	"projectvoltron.dev/voltron/pkg/och/client/local"
)

type Config struct {
	Action           string
	DownloadConfig   []argoactions.DownloadConfig `envconfig:"optional"`
	UploadConfig     argoactions.UploadConfig     `envconfig:"optional"`
	LocalOCHEndpoint string                       `envconfig:"default=http://voltron-och-local.voltron-system/graphql"`
}

func main() {
	var cfg Config
	var action argoactions.Action

	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading configuration")

	logger, err := zap.NewProductionConfig().Build()
	exitOnError(err, "while creating logger")

	client := NewOCHLocalClient(cfg.LocalOCHEndpoint)

	switch cfg.Action {
	case argoactions.DownloadAction:
		action = argoactions.NewDownloadAction(client, cfg.DownloadConfig)

	case argoactions.UploadAction:
		action = argoactions.NewUploadAction(logger, client, cfg.UploadConfig)

	default:
		err := fmt.Errorf("Invalid action: %s", cfg.Action)
		exitOnError(err, "while selecting action")
	}

	ctx := context.Background()
	err = action.Do(ctx)
	exitOnError(err, "while executing action")
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
