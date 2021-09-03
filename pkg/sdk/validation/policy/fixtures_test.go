package policy_test

import (
	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

func fixPolicyWithoutTypeRef() policy.Policy {
	return policy.Policy{
		Rules: policy.RulesList{
			{
				Interface: types.ManifestRefWithOptRevision{
					Path: "cap.*",
				},
				OneOf: []policy.Rule{
					{
						ImplementationConstraints: policy.ImplementationConstraints{},
						Inject: &policy.InjectData{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID: "id",
									},
								},
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID:          "id2",
										Description: ptr.String("ID 2"),
									},
								},
							},
							AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
								{
									AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
										ID:   "id3",
										Name: "id-3",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func fixPolicyWithTypeRef() policy.Policy {
	return policy.Policy{
		Rules: policy.RulesList{
			{
				Interface: types.ManifestRefWithOptRevision{
					Path: "cap.*",
				},
				OneOf: []policy.Rule{
					{
						ImplementationConstraints: policy.ImplementationConstraints{},
						Inject: &policy.InjectData{
							RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID: "id",
									},
									TypeRef: &types.ManifestRef{
										Path:     "cap.type.sample",
										Revision: "0.1.0",
									},
								},
								{
									RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
										ID:          "id2",
										Description: ptr.String("ID 2"),
									},
									TypeRef: &types.ManifestRef{
										Path:     "cap.type.sample2",
										Revision: "0.2.0",
									},
								},
							},
							AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
								{
									AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
										ID:   "id3",
										Name: "name",
									},
									TypeRef: &types.ManifestRef{
										Path:     "cap.type.sample3",
										Revision: "0.3.0",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func fixAWSElasticsearchTypeRev() *gqlpublicapi.TypeRevision {
	return &gqlpublicapi.TypeRevision{
		Metadata: &gqlpublicapi.TypeMetadata{
			Path: "cap.type.aws.elasticsearch.install-input",
		},
		Revision: "0.1.0",
		Spec: &gqlpublicapi.TypeSpec{
			JSONSchema: heredoc.Doc(`
                    {
                      "$schema": "http://json-schema.org/draft-07/schema",
                      "type": "object",
                      "title": "The schema for Elasticsearch input parameters.",
                      "required": ["replicas"],
                      "properties": {
                        "replicas": {
                          "type": "string",
                          "title": "Replica count for the Elasticsearch"
                        }
                      },
                      "additionalProperties": false
                    }`),
		},
	}
}

func fixAWSCredsTypeRev() *gqlpublicapi.TypeRevision {
	return &gqlpublicapi.TypeRevision{
		Metadata: &gqlpublicapi.TypeMetadata{
			Path: "cap.type.aws.auth.creds",
		},
		Revision: "0.1.0",
		Spec: &gqlpublicapi.TypeSpec{
			JSONSchema: heredoc.Doc(`
                    {
                      "$schema": "http://json-schema.org/draft-07/schema",
                      "type": "object",
                      "required": [ "key" ],
                      "properties": {
                        "key": {
                          "type": "string"
                        }
                      }
                    }`),
		},
	}
}

var implementationRevisionRaw = []byte(`
revision: 0.1.0
spec:
  additionalInput:
    parameters:
    - name: additional-parameters
      typeRef:
        path: cap.type.aws.auth.creds
        revision: 0.1.0
    - name: impl-specific-config
      typeRef:
        path: cap.type.aws.elasticsearch.install-input
        revision: 0.1.0
`)
