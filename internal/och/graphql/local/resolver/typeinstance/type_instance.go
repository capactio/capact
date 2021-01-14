package typeinstance

import (
	"context"

	"projectvoltron.dev/voltron/internal/ptr"
	gqllocalapi "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
)

type TypeInstanceResolver struct{}

func NewResolver() *TypeInstanceResolver {
	return &TypeInstanceResolver{}
}

func (r *TypeInstanceResolver) TypeInstances(ctx context.Context, filter *gqllocalapi.TypeInstanceFilter) ([]*gqllocalapi.TypeInstance, error) {
	return []*gqllocalapi.TypeInstance{dummyTypeInstance("cc4926cc-4255-4b67-b7c4-cb38a28f3ec5"), dummyTypeInstance("cc4926cc-4255-4b67-b7c4-cb38a28f3ec3")}, nil
}

func (r *TypeInstanceResolver) TypeInstance(ctx context.Context, id string) (*gqllocalapi.TypeInstance, error) {
	return dummyTypeInstance(id), nil
}

func (r *TypeInstanceResolver) CreateTypeInstance(ctx context.Context, in *gqllocalapi.CreateTypeInstanceInput) (*gqllocalapi.TypeInstance, error) {
	return dummyTypeInstance("5cc47865-3339-4f6d-902e-fc59f2c61943"), nil
}

func (r *TypeInstanceResolver) UpdateTypeInstance(ctx context.Context, id string, in *gqllocalapi.UpdateTypeInstanceInput) (*gqllocalapi.TypeInstance, error) {
	return dummyTypeInstance(id), nil
}

func (r *TypeInstanceResolver) DeleteTypeInstance(ctx context.Context, id string) (*gqllocalapi.TypeInstance, error) {
	return dummyTypeInstance(id), nil
}

func dummyTypeInstance(id string) *gqllocalapi.TypeInstance {
	methodGET := gqllocalapi.HTTPRequestMethodGet
	return &gqllocalapi.TypeInstance{
		Metadata: &gqllocalapi.TypeInstanceMetadata{
			ID: id,
			Attributes: []*gqllocalapi.AttributeReference{
				{
					Path:     "cap.attribute.platform.kubernetes",
					Revision: "1.0.0",
				},
			},
		},
		ResourceVersion: 1410,
		Spec: &gqllocalapi.TypeInstanceSpec{
			TypeRef: &gqllocalapi.TypeReference{
				Path:     "cap.type.database.mysql.config",
				Revision: "0.1.0",
			},
			Value: struct {
				Hostname string
			}{
				Hostname: "mysql.svc.cluster.local:3306",
			},
			Instrumentation: &gqllocalapi.TypeInstanceInstrumentation{
				Metrics: &gqllocalapi.TypeInstanceInstrumentationMetrics{
					Endpoint: ptr.String("https://foo.bar:3000/metrics"),
					Regex:    ptr.String("^(go_gc_duration_seconds|go_goroutines|go_memstats_alloc_bytes|go_memstats_heap_alloc_bytes)$"),
					Dashboards: []*gqllocalapi.TypeInstanceInstrumentationMetricsDashboard{
						{URL: "https://grafana.foo.bar/d/foo/bar"},
						{URL: "https://grafana.foo.bar/d/baz/qux"},
					},
				},
				Health: &gqllocalapi.TypeInstanceInstrumentationHealth{
					URL:    ptr.String("https://foo.bar/healthz"),
					Method: &methodGET,
				},
			},
		},
	}
}
