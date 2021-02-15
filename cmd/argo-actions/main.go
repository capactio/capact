package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/machinebox/graphql"
	"github.com/vrischmann/envconfig"
	argoactions "projectvoltron.dev/voltron/pkg/argo-actions"
	"projectvoltron.dev/voltron/pkg/httputil"
	"projectvoltron.dev/voltron/pkg/och/client/local"
)

type Config struct {
	Action           string
	DownloadConfig   []argoactions.DownloadConfig
	LocalOCHEndpoint string `envconfig:"default=https://voltron-och-local.voltron.local/graphql"`
}

func main() {
	var cfg Config
	var action argoactions.Action

	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading configuration")

	client := NewOCHLocalClient(cfg.LocalOCHEndpoint)

	if cfg.Action == argoactions.DownloadAction {
		action = argoactions.NewDownloadAction(client, cfg.DownloadConfig)
	} else {
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
