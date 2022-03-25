package helmstoragebackend

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"sigs.k8s.io/yaml"

	"capact.io/capact/internal/ptr"
	pb "capact.io/capact/pkg/hub/api/grpc/storage_backend"
	"capact.io/capact/pkg/runner/helm"
)

// repositoryCache Helm cache for repositories
const repositoryCache = "/tmp/helm"

var _ pb.StorageBackendServer = &TemplateHandler{}

type (
	// TemplateContext holds context used by Helm template storage backend.
	TemplateContext struct {
		// GoTemplate specifies Go template which is used to render returned value.
		GoTemplate string `json:"goTemplate"`
		// HelmRelease specifies Helm release details against which the render logic should be executed.
		HelmRelease HelmRelease `json:"release"`
	}
)

// TemplateHandler handles incoming requests to the Helm template storage backend gRPC server.
type TemplateHandler struct {
	pb.UnimplementedStorageBackendServer

	log           *zap.Logger
	fetcher       *HelmReleaseFetcher
	helmOutputter *helm.Outputter
}

// NewTemplateHandler returns new TemplateHandler.
func NewTemplateHandler(log *zap.Logger, helmRelFetcher *HelmReleaseFetcher) *TemplateHandler {
	return &TemplateHandler{
		log:           log,
		fetcher:       helmRelFetcher,
		helmOutputter: helm.NewOutputter(log, helm.NewRenderer()),
	}
}

// GetValue returns a value for a given TypeInstance.
func (h *TemplateHandler) GetValue(_ context.Context, req *pb.GetValueRequest) (*pb.GetValueResponse, error) {
	h.log.Info("getting entry", zap.String("id", req.TypeInstanceId))
	rel, relCtx, err := h.fetchHelmRelease(req.TypeInstanceId, req.Context)
	if err != nil {
		return nil, err
	}

	value, err := h.renderOutputValue(relCtx, rel)
	if err != nil {
		return nil, err
	}

	return &pb.GetValueResponse{
		Value: value,
	}, nil
}

// OnCreate only validates if provided goTemplate can be later rendered on GetValue request.
func (h *TemplateHandler) OnCreate(ctx context.Context, req *pb.OnCreateRequest) (*pb.OnCreateResponse, error) {
	h.log.Info("creating entry", zap.String("id", req.TypeInstanceId))
	// dry-run that get will work
	_, err := h.GetValue(ctx, &pb.GetValueRequest{
		TypeInstanceId: req.TypeInstanceId,
		Context:        req.Context,
	})

	if err != nil {
		return nil, err
	}
	return &pb.OnCreateResponse{}, nil
}

// OnUpdate only validates if provided goTemplate can be later rendered on GetValue request.
func (h *TemplateHandler) OnUpdate(ctx context.Context, req *pb.OnUpdateRequest) (*pb.OnUpdateResponse, error) {
	h.log.Info("updating entry", zap.String("id", req.TypeInstanceId))
	// dry-run that get will work
	_, err := h.GetValue(ctx, &pb.GetValueRequest{
		TypeInstanceId: req.TypeInstanceId,
		Context:        req.Context,
	})

	if err != nil {
		return nil, err
	}
	h.log.Info("return entry up")
	return &pb.OnUpdateResponse{}, nil
}

// OnDelete does nothing.
func (*TemplateHandler) OnDelete(context.Context, *pb.OnDeleteRequest) (*pb.OnDeleteResponse, error) {
	return &pb.OnDeleteResponse{}, nil
}

// GetLockedBy does nothing.
func (*TemplateHandler) GetLockedBy(context.Context, *pb.GetLockedByRequest) (*pb.GetLockedByResponse, error) {
	return &pb.GetLockedByResponse{}, nil
}

// OnLock does nothing.
func (*TemplateHandler) OnLock(context.Context, *pb.OnLockRequest) (*pb.OnLockResponse, error) {
	return &pb.OnLockResponse{}, nil
}

// OnUnlock does nothing.
func (*TemplateHandler) OnUnlock(context.Context, *pb.OnUnlockRequest) (*pb.OnUnlockResponse, error) {
	return &pb.OnUnlockResponse{}, nil
}

func (h *TemplateHandler) getReleaseContext(contextBytes []byte) (*TemplateContext, error) {
	var ctx TemplateContext
	err := json.Unmarshal(contextBytes, &ctx)
	if err != nil {
		return nil, gRPCInternalError(errors.Wrap(err, "while unmarshaling context"))
	}

	if ctx.HelmRelease.Driver == nil {
		ctx.HelmRelease.Driver = ptr.String(defaultHelmDriver)
	}

	return &ctx, nil
}

func (h *TemplateHandler) fetchHelmRelease(ti string, ctx []byte) (*release.Release, *TemplateContext, error) {
	relCtx, err := h.getReleaseContext(ctx)
	if err != nil {
		return nil, nil, gRPCInternalError(err)
	}

	rel, err := h.fetcher.FetchHelmRelease(ti, relCtx.HelmRelease)
	if err != nil {
		return nil, nil, err // it already handles grpc errors properly
	}

	return rel, relCtx, nil
}

func (h *TemplateHandler) renderOutputValue(relCtx *TemplateContext, helmRelease *release.Release) ([]byte, error) {
	if helmRelease.Chart == nil {
		return nil, gRPCInternalError(errors.New("Helm release doesn't have associated Helm chart"))
	}
	args := helm.OutputArgs{
		Additional: helm.AdditionalOutputArgs{
			UseHelmTemplateStorage: false,
			GoTemplate:             relCtx.GoTemplate,
		},
	}

	// It is important to load the chart dependencies as by default they are not
	// available and if chart was using that we may miss some value and template funcs.
	chartDeps, err := h.loadHelmChartDependencies(helmRelease.Chart)
	if err != nil {
		return nil, gRPCInternalError(errors.Wrap(err, "while ensuring dependency charts"))
	}

	helmRelease.Chart.SetDependencies(chartDeps...)
	if err := chartutil.ProcessDependencies(helmRelease.Chart, helmRelease.Config); err != nil {
		return nil, gRPCInternalError(errors.Wrap(err, "while processing dependency charts"))
	}

	data, err := h.helmOutputter.ProduceAdditional(args, helmRelease.Chart, ptr.StringPtrToString(relCtx.HelmRelease.Driver), helmRelease)
	if err != nil {
		return nil, gRPCInternalError(errors.Wrap(err, "while rendering output value"))
	}

	var outputFile helm.OutputFile
	err = yaml.Unmarshal(data, &outputFile)
	if err != nil {
		return nil, gRPCInternalError(errors.Wrap(err, "while unmarshaling output bytes"))
	}

	valueBytes, err := json.Marshal(outputFile.Value)
	if err != nil {
		return nil, gRPCInternalError(errors.Wrap(err, "while marshaling output value to JSON"))
	}

	return valueBytes, nil
}

func (h *TemplateHandler) loadHelmChartDependencies(chrt *chart.Chart) ([]*chart.Chart, error) {
	out := []*chart.Chart{}
	if chrt == nil || chrt.Lock == nil || chrt.Lock.Dependencies == nil || len(chrt.Lock.Dependencies) == 0 {
		return out, nil
	}

	defer h.log.Debug("Loading dependency charts finished")
	for _, dep := range chrt.Lock.Dependencies {
		cpo := action.ChartPathOptions{
			RepoURL: dep.Repository,
			Version: dep.Version,
		}

		chartLocation, err := cpo.LocateChart(dep.Name, &cli.EnvSettings{
			RepositoryCache: repositoryCache,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "while locating %q Helm chart", dep.Name)
		}

		h.log.Debug("Loading chart", zap.String("repo", dep.Repository), zap.String("version", dep.Version), zap.String("name", dep.Name))

		ch, err := loader.Load(chartLocation)
		if err != nil {
			return nil, errors.Wrapf(err, "while loading %q Helm chart", dep.Name)
		}

		out = append(out, ch)
	}

	return out, nil
}
