package client

import (
	"context"

	"projectvoltron.dev/voltron/pkg/engine/k8s/clusterpolicy"

	ochlocalgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	ochpublicgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
	"projectvoltron.dev/voltron/pkg/och/client/public"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
)

type OCHClient interface {
	GetImplementationRevisionsForInterface(ctx context.Context, ref ochpublicgraphql.InterfaceReference, opts ...public.GetImplementationOption) ([]ochpublicgraphql.ImplementationRevision, error)
	GetTypeInstance(ctx context.Context, id string) (*ochlocalgraphql.TypeInstance, error)
	GetInterfaceRevision(ctx context.Context, ref ochpublicgraphql.InterfaceReference) (*ochpublicgraphql.InterfaceRevision, error)
}

type PolicyEnforcedClient struct {
	ochCli OCHClient
	policy clusterpolicy.ClusterPolicy
}

func NewPolicyEnforcedClient(ochCli OCHClient) *PolicyEnforcedClient {
	return &PolicyEnforcedClient{ochCli: ochCli}
}

// Temporary solution to maintain compatibility before cluster policy implementation
func (e *PolicyEnforcedClient) ListImplementationRevisionForInterface(ctx context.Context, interfaceRef ochpublicgraphql.InterfaceReference) ([]ochpublicgraphql.ImplementationRevision, clusterpolicy.Rule, error) {
	impls, err := e.ochCli.GetImplementationRevisionsForInterface(ctx, interfaceRef, public.WithImplementationFilter(e.defaultOCHImplementationFilter()))
	if err != nil {
		return nil, clusterpolicy.Rule{}, err
	}

	return impls, clusterpolicy.Rule{}, err
}

func (e *PolicyEnforcedClient) GetInterfaceRevision(ctx context.Context, ref ochpublicgraphql.InterfaceReference) (*ochpublicgraphql.InterfaceRevision, error) {
	return e.ochCli.GetInterfaceRevision(ctx, ref)
}

func (e *PolicyEnforcedClient) ListTypeInstancesToInjectBasedOnPolicy(ctx context.Context, policyRule clusterpolicy.Rule, implRev ochpublicgraphql.ImplementationRevision) ([]types.InputTypeInstanceRef, error) {
	// TODO: Implement as a part of SV-185
	return []types.InputTypeInstanceRef{}, nil
}

func (e *PolicyEnforcedClient) SetPolicy(policy clusterpolicy.ClusterPolicy) {
	e.policy = policy
}

// TODO: Remove it
func (e *PolicyEnforcedClient) defaultOCHImplementationFilter() ochpublicgraphql.ImplementationRevisionFilter {
	exclude := ochpublicgraphql.FilterRuleExclude

	return ochpublicgraphql.ImplementationRevisionFilter{
		RequirementsSatisfiedBy: []*ochpublicgraphql.TypeInstanceValue{
			{TypeRef: &ochpublicgraphql.TypeReferenceInput{Path: "cap.core.type.platform.kubernetes"}},
		},
		// currently we do not have any policies, so GCP solutions are not supported
		Attributes: []*ochpublicgraphql.AttributeFilterInput{
			{
				Path: "cap.attribute.cloud.provider.gcp",
				Rule: &exclude,
			},
		},
	}
}
