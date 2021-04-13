package argo

import (
	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/clusterpolicy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

func fixGCPClusterPolicy() clusterpolicy.ClusterPolicy {
	return clusterpolicy.ClusterPolicy{
		APIVersion: "0.1.0",
		Rules: clusterpolicy.RulesMap{
			"cap.*": {
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{},
					},
				},
			},
			"cap.interface.database.postgresql.install:0.1.0": {
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Requires: &[]types.TypeRefWithOptRevision{
								{
									Path:     "cap.type.gcp.auth.service-account",
									Revision: ptr.String("0.1.0"),
								},
							},
							Attributes: &[]types.AttributeRef{
								{
									Path: "cap.attribute.cloud.provider.gcp",
								},
							},
						},
						InjectTypeInstances: []clusterpolicy.TypeInstanceToInject{
							{
								ID: "c268d3f5-8834-434b-bea2-b677793611c5",
								TypeRef: types.TypeRefWithOptRevision{
									Path:     "cap.type.gcp.auth.service-account",
									Revision: ptr.String("0.1.0"),
								},
							},
						},
					},
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Path: ptr.String("cap.implementation.bitnami.postgresql.install"),
						},
					},
				},
			},
		},
	}
}

func fixAWSClusterPolicy() clusterpolicy.ClusterPolicy {
	return clusterpolicy.ClusterPolicy{
		APIVersion: "0.1.0",
		Rules: clusterpolicy.RulesMap{
			"cap.*": {
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{},
					},
				},
			},
			"cap.interface.database.postgresql.install:0.1.0": {
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Requires: &[]types.TypeRefWithOptRevision{
								{
									Path:     "cap.type.aws.auth.credentials",
									Revision: ptr.String("0.1.0"),
								},
							},
							Attributes: &[]types.AttributeRef{
								{
									Path: "cap.attribute.cloud.provider.aws",
								},
							},
						},
					},
				},
			},
			"cap.interface.aws.rds.postgresql.provision": {
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Requires: &[]types.TypeRefWithOptRevision{
								{
									Path:     "cap.type.aws.auth.credentials",
									Revision: ptr.String("0.1.0"),
								},
							},
							Attributes: &[]types.AttributeRef{
								{
									Path: "cap.attribute.cloud.provider.aws",
								},
							},
						},
						InjectTypeInstances: []clusterpolicy.TypeInstanceToInject{
							{
								ID: "517cf827-233c-4bf1-8fc9-48534424dd58",
								TypeRef: types.TypeRefWithOptRevision{
									Path:     "cap.type.aws.auth.credentials",
									Revision: ptr.String("0.1.0"),
								},
							},
						},
					},
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Path: ptr.String("cap.implementation.bitnami.postgresql.install"),
						},
					},
				},
			},
		},
	}
}

func fixClusterPolicyForFallback() clusterpolicy.ClusterPolicy {
	return clusterpolicy.ClusterPolicy{
		APIVersion: "0.1.0",
		Rules: clusterpolicy.RulesMap{
			"cap.*": {
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{},
					},
				},
			},
			"cap.interface.database.postgresql.install:0.1.0": {
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Attributes: &[]types.AttributeRef{
								{
									Path:     "cap.attribute.not-existing",
									Revision: ptr.String("0.1.0"),
								},
							},
						},
						InjectTypeInstances: []clusterpolicy.TypeInstanceToInject{
							{
								ID: "gcp-sa-uuid",
								TypeRef: types.TypeRefWithOptRevision{
									Path:     "cap.type.gcp.auth.service-account",
									Revision: ptr.String("0.1.0"),
								},
							},
						},
					},
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Path: ptr.String("cap.implementation.bitnami.postgresql.install"),
						},
					},
				},
			},
		},
	}
}

func fixTerraformPolicy() clusterpolicy.ClusterPolicy {
	return clusterpolicy.ClusterPolicy{
		APIVersion: "0.1.0",
		Rules: clusterpolicy.RulesMap{
			"cap.interface.database.postgresql.install:0.1.0": {
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{
							Path: ptr.String("cap.implementation.terraform.gcp.cloudsql.postgresql.install"),
						},
						InjectTypeInstances: []clusterpolicy.TypeInstanceToInject{
							{
								ID: "c268d3f5-8834-434b-bea2-b677793611c5",
								TypeRef: types.TypeRefWithOptRevision{
									Path:     "cap.type.gcp.auth.service-account",
									Revision: ptr.String("0.1.0"),
								},
							},
						},
					},
				},
			},
			"cap.*": {
				OneOf: []clusterpolicy.Rule{
					{
						ImplementationConstraints: clusterpolicy.ImplementationConstraints{},
					},
				},
			},
		},
	}
}
