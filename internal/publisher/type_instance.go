package publisher

import (
	"context"
	"fmt"

	"projectvoltron.dev/voltron/internal/logger"
	"projectvoltron.dev/voltron/internal/ptr"
	gqllocalapi "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	"projectvoltron.dev/voltron/pkg/och/client/local"
	"projectvoltron.dev/voltron/pkg/runner/helm"

	"github.com/MakeNowJust/heredoc"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/yaml"
)

const (
	actionNameFormat       = "%s-config"
	helmReleaseTypeRefPath = "cap.type.helm.chart.release"
	capactTypeRefPath      = "cap.type.capactio.capact.config"
)

var voltronAdditionalOutput = heredoc.Doc(`
							gateway:
							  url: "https://{{ .Values.gateway.ingress.host}}.{{ .Values.global.domainName }}"
							  username: "{{ .Values.global.gateway.auth.username}}"
							  password: "{{ .Values.global.gateway.auth.password }}"
						`)

// TypeInstances provides functionality to produce and upload TypeInstances
type TypeInstances struct {
	k8sCfg        *rest.Config
	logger        *zap.Logger
	localOCHCli   *local.Client
	cfg           TypeInstancesConfig
	helmOutputter *helm.Outputter
}

func NewTypeInstances() (*TypeInstances, error) {
	var cfg TypeInstancesConfig
	err := envconfig.Init(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "while loading configuration")
	}

	logger, err := logger.New(cfg.Logger)
	if err != nil {
		return nil, errors.Wrap(err, "while creating zap logger")
	}

	k8sCfg, err := config.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s config")
	}

	client := local.NewDefaultClient(cfg.LocalOCHEndpoint)

	return &TypeInstances{
		k8sCfg:        k8sCfg,
		logger:        logger,
		localOCHCli:   client,
		cfg:           cfg,
		helmOutputter: helm.NewOutputter(logger, helm.NewRenderer()),
	}, nil
}

// PublishVoltronInstallTypeInstances produces and uploads TypeInstances which describe Voltron installation.
func (i *TypeInstances) PublishVoltronInstallTypeInstances(ctx context.Context) error {
	listAct, err := i.newListAction()
	if err != nil {
		return errors.Wrap(err, "while creating Helm list action")
	}

	releases, err := listAct.Run()
	if err != nil {
		return errors.Wrap(err, "while listing all Helm releases")
	}

	var (
		ownerName = fmt.Sprintf(actionNameFormat, i.cfg.VoltronReleaseName)
		ti        []*gqllocalapi.CreateTypeInstanceInput
	)

	for _, r := range releases {
		if i.cfg.HelmReleasesNSLookup.Has(r.Namespace) {
			continue
		}

		helmReleaseTI, err := i.produceHelmReleaseTypeInstance(r)
		if err != nil {
			return errors.Wrap(err, "while producing Helm release TypeInstance")
		}
		ti = append(ti, helmReleaseTI)

		if r.Name == i.cfg.VoltronReleaseName {
			configTI, err := i.produceConfigTypeInstance(ownerName, r)
			if err != nil {
				return errors.Wrap(err, "while producing config TypeInstance")
			}
			ti = append(ti, configTI)
		}
	}

	tiToCreate := i.createTypeInstancesInput(ownerName, ti)
	uploadOutput, err := i.localOCHCli.CreateTypeInstances(ctx, tiToCreate)
	if err != nil {
		return errors.Wrap(err, "while uploading TypeInstances to OCH")
	}

	for _, ti := range uploadOutput {
		i.logger.Info("TypeInstance uploaded", zap.String("alias", ti.Alias), zap.String("ID", ti.ID))
	}

	return nil
}

func (i *TypeInstances) createTypeInstancesInput(owner string, ti []*gqllocalapi.CreateTypeInstanceInput) *gqllocalapi.CreateTypeInstancesInput {
	var rel []*gqllocalapi.TypeInstanceUsesRelationInput
	for _, item := range ti {
		if owner != *item.Alias {
			rel = append(rel, &gqllocalapi.TypeInstanceUsesRelationInput{
				From: owner,
				To:   *item.Alias,
			})
		}
	}

	return &gqllocalapi.CreateTypeInstancesInput{
		TypeInstances: ti,
		UsesRelations: rel,
	}
}

func (i *TypeInstances) newListAction() (*action.List, error) {
	actionConfig := new(action.Configuration)
	helmCfg := &genericclioptions.ConfigFlags{
		APIServer:   &i.k8sCfg.Host,
		Insecure:    &i.k8sCfg.Insecure,
		CAFile:      &i.k8sCfg.CAFile,
		BearerToken: &i.k8sCfg.BearerToken,
	}

	debugLog := func(format string, v ...interface{}) {
		i.logger.Debug(fmt.Sprintf(format, v...), zap.String("source", "Helm"))
	}

	err := actionConfig.Init(helmCfg, "", "secrets", debugLog)
	if err != nil {
		return nil, err
	}

	return action.NewList(actionConfig), nil
}

func (i *TypeInstances) produceHelmReleaseTypeInstance(helmRelease *release.Release) (*gqllocalapi.CreateTypeInstanceInput, error) {
	releaseOut, err := i.helmOutputter.ProduceHelmRelease(i.cfg.HelmRepositoryPath, helmRelease)
	if err != nil {
		return nil, errors.Wrap(err, "while producing Helm release definition")
	}

	var unmarshaled interface{}
	err = yaml.Unmarshal(releaseOut, &unmarshaled)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshaling bytes")
	}

	return &gqllocalapi.CreateTypeInstanceInput{
		Alias: ptr.String(helmRelease.Name),
		TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
			Path:     helmReleaseTypeRefPath,
			Revision: "0.1.0",
		},
		Value: unmarshaled,
	}, nil
}

func (i *TypeInstances) produceConfigTypeInstance(ownerName string, helmRelease *release.Release) (*gqllocalapi.CreateTypeInstanceInput, error) {
	tpl, err := yaml.YAMLToJSON([]byte(voltronAdditionalOutput))
	if err != nil {
		return nil, errors.Wrap(err, "while converting YAML to JSON")
	}

	args := helm.OutputArgs{
		GoTemplate: tpl,
	}
	data, err := i.helmOutputter.ProduceAdditional(args, helmRelease.Chart, helmRelease)
	if err != nil {
		return nil, errors.Wrap(err, "while producing additional info")
	}

	var unmarshaled interface{}
	err = yaml.Unmarshal(data, &unmarshaled)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshaling bytes")
	}
	return &gqllocalapi.CreateTypeInstanceInput{
		Alias: ptr.String(ownerName),
		TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
			Path:     capactTypeRefPath,
			Revision: "0.1.0",
		},
		Value: unmarshaled,
	}, nil
}
