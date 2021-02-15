package implementations

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"

	corev1 "k8s.io/api/core/v1"
)

type ImplementationResolver struct {
}

func NewResolver() *ImplementationResolver {
	return &ImplementationResolver{}
}

type ImplementationRevisionResolver struct {
}

func NewRevisionResolver() *ImplementationRevisionResolver {
	return &ImplementationRevisionResolver{}
}

func (i *ImplementationResolver) Implementations(ctx context.Context, filter *gqlpublicapi.ImplementationFilter) ([]*gqlpublicapi.Implementation, error) {
	return []*gqlpublicapi.Implementation{dummyImplementation("cap.implementation.cms.wordpress.install")}, nil
}

func (i ImplementationResolver) Implementation(ctx context.Context, path string) (*gqlpublicapi.Implementation, error) {
	return dummyImplementation(path), nil
}

func (i *ImplementationResolver) Revision(ctx context.Context, obj *gqlpublicapi.Implementation, revision string) (*gqlpublicapi.ImplementationRevision, error) {
	return &gqlpublicapi.ImplementationRevision{}, fmt.Errorf("No Implementation with revision %s", revision)
}

func (i *ImplementationRevisionResolver) Interfaces(ctx context.Context, obj *gqlpublicapi.ImplementationRevision) ([]*gqlpublicapi.Interface, error) {
	return []*gqlpublicapi.Interface{}, nil
}

func dummyImplementation(path string) *gqlpublicapi.Implementation {
	var (
		name   = filepath.Ext(path)
		prefix = strings.TrimSuffix(path, fmt.Sprintf(".%s", name))

		wf = &v1alpha1.WorkflowSpec{
			Entrypoint: "whalesay",
			Templates: []v1alpha1.Template{
				{
					Name: "whalesay",
					Container: &corev1.Container{
						Image:   "docker/whalesay:latest",
						Command: []string{"sh", "-c", "cowsay 'Never gonna give you up... Never gonna let you down'"},
					},
				},
			},
		}
	)

	if strings.Contains(path, "failure") {
		wf.Templates[0].Container.Command = []string{"sh", "-c", "cowsay 'Oops! ...I Did It Again'; exit 1"}
	}

	return &gqlpublicapi.Implementation{
		Name:   name,
		Prefix: prefix,
		Path:   path,
		LatestRevision: &gqlpublicapi.ImplementationRevision{
			Spec: &gqlpublicapi.ImplementationSpec{
				Action: &gqlpublicapi.ImplementationAction{
					RunnerInterface: "cap.interface.runner.argo",
					Args: map[string]interface{}{
						"workflow": wf,
					},
				},
			},
		},
	}
}
