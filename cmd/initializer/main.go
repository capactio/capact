// Maybe we can get rid of that and replace with some bash script
// Scenario:
// 1. Each Helm Chart produces a ConfigMap with TypeInstance value
// 1.2. Voltron Helm Charts produces additionally the capact-config
// 2. Voltron post-install jobs collects all ConfigMap using a given label selector
// 3. Uploads them into a local OCH using curl -X POST ...

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/machinebox/graphql"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"projectvoltron.dev/voltron/internal/ptr"
	"projectvoltron.dev/voltron/pkg/httputil"
	gqllocalapi "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	"projectvoltron.dev/voltron/pkg/och/client/local"
	"projectvoltron.dev/voltron/pkg/runner/helm"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/yaml"
)

type Config struct {
	LoggerDevMode      bool   `envconfig:"default=false"`
	LocalOCHEndpoint   string `envconfig:"default=http://voltron-och-local.voltron-system/graphql"`
	HelmReleasesLookup []string
	CapactReleaseName  string `envconfig:"default=voltron"`
	HelmRepositoryPath string `envconfig:"default=https://capactio-awesome-charts.storage.googleapis.com"`
}

func main() {
	var cfg Config
	err := envconfig.Init(&cfg)
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

	// HELM
	lookupNS := map[string]struct{}{}
	for _, ns := range cfg.HelmReleasesLookup {
		lookupNS[ns] = struct{}{}
	}

	k8sCfg, err := config.GetConfig()
	exitOnError(err, "while creating k8s config")
	actionConfig := new(action.Configuration)
	helmCfg := &genericclioptions.ConfigFlags{
		APIServer:   &k8sCfg.Host,
		Insecure:    &k8sCfg.Insecure,
		CAFile:      &k8sCfg.CAFile,
		BearerToken: &k8sCfg.BearerToken,
	}

	debugLog := func(format string, v ...interface{}) {
		logger.Debug(fmt.Sprintf(format, v...), zap.String("source", "Helm"))
	}

	err = actionConfig.Init(helmCfg, "", "secrets", debugLog)
	exitOnError(err, "while initializing Helm configuration")

	listAct := action.NewList(actionConfig)
	releases, err := listAct.Run()
	exitOnError(err, "")

	renderer := helm.NewRenderer()
	outputter := helm.NewOutputter(logger, renderer)

	var ti []*gqllocalapi.CreateTypeInstanceInput

	capactConfigTIName := fmt.Sprintf("%s-config", cfg.CapactReleaseName)

	for _, r := range releases {
		if _, found := lookupNS[r.Namespace]; !found {
			continue
		}

		releaseOut, err := outputter.ProduceHelmRelease(cfg.HelmRepositoryPath, r)
		exitOnError(err, "")

		var values interface{}
		err = yaml.Unmarshal(releaseOut, &values)
		exitOnError(err, "while unmarshaling bytes")

		ti = append(ti, &gqllocalapi.CreateTypeInstanceInput{
			Alias: ptr.String(r.Name),
			Attributes: []*gqllocalapi.AttributeReferenceInput{
				{
					Path:     "cap.core.attribute.system.capact-managed",
					Revision: "0.1.0",
				},
			},
			TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
				Path:     "cap.type.helm.chart.release",
				Revision: "0.1.0",
			},
			Value: values,
		})

		if r.Name == cfg.CapactReleaseName {
			tpl, err := yaml.YAMLToJSON([]byte(heredoc.Doc(`
							gateway:
							  url: "https://{{ .Values.gateway.ingress.host}}.{{ .Values.global.domainName }}"
							  username: "{{ .Values.global.gateway.auth.username}}"
							  password: "{{ .Values.global.gateway.auth.password }}"
						`)))
			exitOnError(err, "")

			args := helm.OutputArgs{
				GoTemplate: tpl,
			}
			data, err := outputter.ProduceAdditional(args, r.Chart, r)
			exitOnError(err, "")

			var values interface{}
			err = yaml.Unmarshal(data, &values)
			exitOnError(err, "while unmarshaling bytes")
			ti = append(ti, &gqllocalapi.CreateTypeInstanceInput{
				Alias: ptr.String(capactConfigTIName),
				Attributes: []*gqllocalapi.AttributeReferenceInput{
					{
						Path:     "cap.core.attribute.system.capact-managed",
						Revision: "0.1.0",
					},
				},
				TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
					Path:     "cap.type.capactio.capact.config",
					Revision: "0.1.0",
				},
				Value: values,
			})
		}
	}

	uploadOutput, err := client.CreateTypeInstances(context.Background(), createTypeInstancesInput(capactConfigTIName, ti))
	exitOnError(err, "")

	for _, ti := range uploadOutput {
		logger.Info("TypeInstance uploaded", zap.String("alias", ti.Alias), zap.String("ID", ti.ID))
	}
}

func createTypeInstancesInput(capactConfigTIName string, ti []*gqllocalapi.CreateTypeInstanceInput) *gqllocalapi.CreateTypeInstancesInput {
	var rel []*gqllocalapi.TypeInstanceUsesRelationInput
	for _, item := range ti {
		if capactConfigTIName != *item.Alias {
			rel = append(rel, &gqllocalapi.TypeInstanceUsesRelationInput{
				From: capactConfigTIName,
				To:   *item.Alias,
			})
		}
	}

	return &gqllocalapi.CreateTypeInstancesInput{
		TypeInstances: ti,
		UsesRelations: rel,
	}
}

// TODO unify:

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
