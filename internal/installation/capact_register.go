package installation

import (
	"context"
	"fmt"

	"capact.io/capact/internal/logger"
	"capact.io/capact/internal/ptr"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/hub/client/local"
	"capact.io/capact/pkg/runner/helm"

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

var capactAdditionalOutput = heredoc.Doc(`
							version: "{{ .Values.global.containerRegistry.overrideTag | default .Chart.AppVersion }}"
							gateway:
							  url: "https://{{ .Values.gateway.ingress.host}}.{{ .Values.global.domainName }}"
							  username: "{{ .Values.global.gateway.auth.username}}"
							  password: "{{ .Values.global.gateway.auth.password }}"
						`)

// CapactRegister provides functionality to produce and upload Capact TypeInstances
type CapactRegister struct {
	k8sCfg        *rest.Config
	logger        *zap.Logger
	localHubCli   *local.Client
	cfg           TypeInstancesConfig
	helmOutputter *helm.Outputter
}

func NewCapactRegister() (*CapactRegister, error) {
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

	client := local.NewDefaultClient(cfg.LocalHubEndpoint)

	return &CapactRegister{
		k8sCfg:        k8sCfg,
		logger:        logger,
		localHubCli:   client,
		cfg:           cfg,
		helmOutputter: helm.NewOutputter(logger, helm.NewRenderer()),
	}, nil
}

// RegisterTypeInstances produces and uploads TypeInstances which describe Capact installation.
func (i *CapactRegister) RegisterTypeInstances(ctx context.Context) error {
	listAct, err := i.newHelmListAction()
	if err != nil {
		return errors.Wrap(err, "while creating Helm list action")
	}

	releases, err := listAct.Run()
	if err != nil {
		return errors.Wrap(err, "while listing all Helm releases")
	}

	var (
		ownerName = fmt.Sprintf(actionNameFormat, i.cfg.CapactReleaseName)
		ti        []*gqllocalapi.CreateTypeInstanceInput
	)

	for _, r := range releases {
		if !i.cfg.HelmReleasesNSLookup.Contains(r.Namespace) {
			continue
		}

		helmReleaseTI, err := i.produceHelmReleaseTypeInstance(r)
		if err != nil {
			return errors.Wrap(err, "while producing Helm release TypeInstance")
		}
		ti = append(ti, helmReleaseTI)

		if r.Name == i.cfg.CapactReleaseName {
			configTI, err := i.produceConfigTypeInstance(ownerName, r)
			if err != nil {
				return errors.Wrap(err, "while producing config TypeInstance")
			}
			ti = append(ti, configTI)
		}
	}

	tiToCreate := i.createTypeInstancesInput(ownerName, ti)
	uploadOutput, err := i.localHubCli.CreateTypeInstances(ctx, tiToCreate)
	if err != nil {
		return errors.Wrap(err, "while uploading TypeInstances to Hub")
	}

	for _, ti := range uploadOutput {
		i.logger.Info("TypeInstance uploaded", zap.String("alias", ti.Alias), zap.String("ID", ti.ID))
	}

	return nil
}

func (i *CapactRegister) createTypeInstancesInput(owner string, ti []*gqllocalapi.CreateTypeInstanceInput) *gqllocalapi.CreateTypeInstancesInput {
	var rel []*gqllocalapi.TypeInstanceUsesRelationInput
	for _, item := range ti {
		if owner == *item.Alias {
			continue
		}
		rel = append(rel, &gqllocalapi.TypeInstanceUsesRelationInput{
			From: owner,
			To:   *item.Alias,
		})
	}

	return &gqllocalapi.CreateTypeInstancesInput{
		TypeInstances: ti,
		UsesRelations: rel,
	}
}

func (i *CapactRegister) newHelmListAction() (*action.List, error) {
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

	actList := action.NewList(actionConfig)

	// We do not wait for all Helm chart to finish before Capact is deployed
	// Additionally, this command is executed as a post-install hook, which means that Capact itself is still
	// in `pending-install`.
	actList.All = true
	actList.SetStateMask()

	return actList, nil
}

func (i *CapactRegister) produceHelmReleaseTypeInstance(helmRelease *release.Release) (*gqllocalapi.CreateTypeInstanceInput, error) {
	releaseOut, err := i.helmOutputter.ProduceHelmRelease(i.cfg.HelmRepositoryPath, helmRelease)
	if err != nil {
		return nil, errors.Wrap(err, "while producing Helm release definition")
	}

	var unmarshalled interface{}
	err = yaml.Unmarshal(releaseOut, &unmarshalled)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshaling bytes")
	}

	return &gqllocalapi.CreateTypeInstanceInput{
		Alias: ptr.String(helmRelease.Name),
		TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
			Path:     helmReleaseTypeRefPath,
			Revision: "0.1.0",
		},
		Value: unmarshalled,
	}, nil
}

func (i *CapactRegister) produceConfigTypeInstance(ownerName string, helmRelease *release.Release) (*gqllocalapi.CreateTypeInstanceInput, error) {
	tpl, err := yaml.YAMLToJSON([]byte(capactAdditionalOutput))
	if err != nil {
		return nil, errors.Wrap(err, "while converting YAML to JSON")
	}

	args := helm.OutputArgs{
		GoTemplate: string(tpl),
	}
	data, err := i.helmOutputter.ProduceAdditional(args, helmRelease.Chart, helmRelease)
	if err != nil {
		return nil, errors.Wrap(err, "while producing additional info")
	}

	var unmarshalled interface{}
	err = yaml.Unmarshal(data, &unmarshalled)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshaling bytes")
	}
	return &gqllocalapi.CreateTypeInstanceInput{
		Alias: ptr.String(ownerName),
		TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
			Path:     capactTypeRefPath,
			Revision: "0.1.0",
		},
		Value: unmarshalled,
	}, nil
}
