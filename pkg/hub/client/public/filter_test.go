package public

import (
	"testing"

	"capact.io/capact/internal/ptr"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"github.com/stretchr/testify/assert"
)

func TestImplementationAttributeFilters(t *testing.T) {
	include := gqlpublicapi.FilterRuleInclude
	exclude := gqlpublicapi.FilterRuleExclude

	tests := []struct {
		name                string
		expRevision         []gqlpublicapi.ImplementationRevision
		revisionToFilterOut []gqlpublicapi.ImplementationRevision
		filterAttr          *gqlpublicapi.AttributeFilterInput
	}{
		{
			name: "Return revisions without attr cap.attr.foo 0.0.1",
			expRevision: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevision("without-attr", "0.0.1"),
				fixImplementationRevisionWithAttr("with-attr-foo2", "0.1.1", "cap.attr.foo2", "0.0.1"),
				fixImplementationRevisionWithAttr("with-attr-foo-0.1.0", "0.1.0", "cap.attr.foo", "0.1.0"),
			},
			revisionToFilterOut: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevisionWithAttr("with-attr-foo-0.0.1", "0.1.0", "cap.attr.foo", "0.0.1"),
			},
			filterAttr: &gqlpublicapi.AttributeFilterInput{
				Path:     "cap.attr.foo",
				Rule:     &exclude,
				Revision: ptr.String("0.0.1"),
			},
		},
		{
			name: "Return revisions with attr cap.attr.foo 0.0.1",
			expRevision: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevisionWithAttr("with-attr-foo-0.0.1", "0.1.0", "cap.attr.foo", "0.0.1"),
			},
			revisionToFilterOut: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevision("without-attr", "0.0.1"),
				fixImplementationRevisionWithAttr("with-attr-foo2", "0.1.0", "cap.attr.foo2", "0.0.1"),
				fixImplementationRevisionWithAttr("with-attr-foo-0.1.0", "0.1.0", "cap.attr.foo", "0.1.0"),
			},
			filterAttr: &gqlpublicapi.AttributeFilterInput{
				Path:     "cap.attr.foo",
				Rule:     &include,
				Revision: ptr.String("0.0.1"),
			},
		},
		{
			name: "Return revisions without attr cap.attr.foo (revision is not checked)",
			expRevision: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevision("without-attr", "0.0.1"),
				fixImplementationRevisionWithAttr("with-attr-foo2", "0.1.0", "cap.attr.foo2", "0.0.1"),
			},
			revisionToFilterOut: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevisionWithAttr("with-attr-foo-0.0.1", "0.1.0", "cap.attr.foo", "0.0.1"),
				fixImplementationRevisionWithAttr("with-attr-foo-0.1.0", "0.1.0", "cap.attr.foo", "0.1.0"),
			},
			filterAttr: &gqlpublicapi.AttributeFilterInput{
				Path: "cap.attr.foo",
				Rule: &exclude,
			},
		},
		{
			name: "Return revisions with attr cap.attr.foo (revision is not checked)",
			expRevision: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevisionWithAttr("with-attr-foo-0.0.1", "0.1.0", "cap.attr.foo", "0.0.1"),
				fixImplementationRevisionWithAttr("with-attr-foo-0.1.0", "0.1.0", "cap.attr.foo", "0.1.0"),
			},
			revisionToFilterOut: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevision("without-attr", "0.0.1"),
				fixImplementationRevisionWithAttr("with-attr-foo2", "0.1.0", "cap.attr.foo2", "0.0.1"),
			},
			filterAttr: &gqlpublicapi.AttributeFilterInput{
				Path: "cap.attr.foo",
				Rule: &include,
			},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			// given
			filter := gqlpublicapi.ImplementationRevisionFilter{
				Attributes: []*gqlpublicapi.AttributeFilterInput{tt.filterAttr},
			}

			getOpts := &ListImplementationRevisionsForInterfaceOptions{}
			getOpts.Apply(WithFilter(filter))

			allRevs := append(tt.expRevision, tt.revisionToFilterOut...)

			// when
			gotRevs := FilterImplementationRevisions(allRevs, getOpts)

			// then
			assert.Len(t, gotRevs, len(tt.expRevision))
			for idx := range tt.expRevision {
				assert.Contains(t, gotRevs, tt.expRevision[idx])
			}
		})
	}
}

func TestImplementationRequirementsSatisfiedByFilters(t *testing.T) {
	tests := []struct {
		name                           string
		expRevision                    []gqlpublicapi.ImplementationRevision
		revisionToFilterOut            []gqlpublicapi.ImplementationRevision
		requirementsSatisfiedBy        []*gqlpublicapi.TypeInstanceValue
		requiredTIInjectionSatisfiedBy []*gqlpublicapi.TypeInstanceValue
	}{
		{
			name: "Return all Implementations as they are without requirements",
			expRevision: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevision("without-attr", "0.0.1"),
				fixImplementationRevisionWithAttr("with-attr-foo2", "0.1.1", "cap.attr.foo2", "0.0.1"),
				fixImplementationRevisionWithAttr("with-attr-foo-0.1.0", "0.1.0", "cap.attr.foo", "0.1.0"),
			},
			requirementsSatisfiedBy: []*gqlpublicapi.TypeInstanceValue{
				{
					TypeRef: &gqlpublicapi.TypeReferenceInput{
						Path:     "cap.type.gcp.sa",
						Revision: "0.1.1",
					},
				},
			},
		},
		{
			name: "Return Implementations satisfied by GCP SA",
			expRevision: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevision("without-any-requirements", "0.0.1"),
				fixImplementationRevisionWithRequire("with-gcp-sa-requirement", "0.1.0", gqlpublicapi.ImplementationRequirement{
					AllOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.gcp.sa",
								Revision: "0.1.1",
							},
						},
					},
				}),
			},
			revisionToFilterOut: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevisionWithRequire("with-cf-requirement", "0.1.0", gqlpublicapi.ImplementationRequirement{
					AllOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.core.type.platform.cf",
								Revision: "0.1.1",
							},
						},
					},
				}),
			},
			requirementsSatisfiedBy: []*gqlpublicapi.TypeInstanceValue{
				{
					TypeRef: &gqlpublicapi.TypeReferenceInput{
						Path:     "cap.type.gcp.sa",
						Revision: "0.1.1",
					},
				},
			},
		},
		{
			name: "Return Implementations satisfied by injected RequiredTypeInstances and requires satisfied by",
			expRevision: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevision("without-any-requirements", "0.0.1"),
				fixImplementationRevisionWithRequire("with-gcp-sa-requirement", "0.1.0", gqlpublicapi.ImplementationRequirement{
					AllOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							Alias: ptr.String("gcp-sa"),
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.gcp.sa",
								Revision: "0.1.1",
							},
						},
					},
				}),
				fixImplementationRevisionWithRequire("with-gcp-sa-requirement", "0.1.0", gqlpublicapi.ImplementationRequirement{
					AllOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.sample",
								Revision: "0.1.0",
							},
						},
					},
				}),
				fixImplementationRevisionWithRequire("any-of", "0.1.0", gqlpublicapi.ImplementationRequirement{
					AnyOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.not-existing",
								Revision: "0.1.0",
							},
						},
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.sample",
								Revision: "0.1.0",
							},
						},
					},
					OneOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.not-existing",
								Revision: "0.1.0",
							},
						},
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.sample",
								Revision: "0.1.0",
							},
						},
					},
				}),
			},
			revisionToFilterOut: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevisionWithRequire("with-gcp-sa-requirement-alias", "0.1.0", gqlpublicapi.ImplementationRequirement{
					AllOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							Alias: ptr.String("foo"),
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.core.type.platform.cf",
								Revision: "0.1.1",
							},
						},
					},
				}),
				fixImplementationRevisionWithRequire("any-of", "0.1.0", gqlpublicapi.ImplementationRequirement{
					OneOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.gcp.sa",
								Revision: "0.1.1",
							},
						},
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.sample",
								Revision: "0.1.0",
							},
						},
					},
				}),
			},
			requiredTIInjectionSatisfiedBy: []*gqlpublicapi.TypeInstanceValue{
				{
					TypeRef: &gqlpublicapi.TypeReferenceInput{
						Path:     "cap.type.gcp.sa",
						Revision: "0.1.1",
					},
				},
			},
			requirementsSatisfiedBy: []*gqlpublicapi.TypeInstanceValue{
				{
					TypeRef: &gqlpublicapi.TypeReferenceInput{
						Path:     "cap.type.gcp.sa",
						Revision: "0.1.1",
					},
				},
				{
					TypeRef: &gqlpublicapi.TypeReferenceInput{
						Path:     "cap.type.sample",
						Revision: "0.1.0",
					},
				},
				{
					TypeRef: &gqlpublicapi.TypeReferenceInput{
						Path:     "cap.core.type.platform.cf",
						Revision: "0.1.1",
					},
				},
			},
		},
		{
			name: "Return Implementations satisfied by injected RequiredTypeInstances without using `requirementsSatisfiedBy`",
			expRevision: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevision("without-any-requirements", "0.0.1"),
				fixImplementationRevisionWithRequire("with-gcp-sa-requirement", "0.1.0", gqlpublicapi.ImplementationRequirement{
					AllOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							Alias: ptr.String("gcp-sa"),
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.gcp.sa",
								Revision: "0.1.1",
							},
						},
						{
							Alias: ptr.String("aws-credentials"),
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.aws.credentials",
								Revision: "0.1.0",
							},
						},
					},
				}),
			},
			requiredTIInjectionSatisfiedBy: []*gqlpublicapi.TypeInstanceValue{
				{
					TypeRef: &gqlpublicapi.TypeReferenceInput{
						Path:     "cap.type.gcp.sa",
						Revision: "0.1.1",
					},
				},
				{
					TypeRef: &gqlpublicapi.TypeReferenceInput{
						Path:     "cap.type.aws.credentials",
						Revision: "0.1.0",
					},
				},
			},
		},
		{
			name: "Return Implementations satisfied by injected RequiredTypeInstances without including them in `requirementsSatisfiedBy`",
			expRevision: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevision("without-any-requirements", "0.0.1"),
				fixImplementationRevisionWithRequire("with-gcp-sa-requirement", "0.1.0", gqlpublicapi.ImplementationRequirement{
					AllOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							Alias: ptr.String("gcp-sa"),
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.gcp.sa",
								Revision: "0.1.1",
							},
						},
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.core.type.platform.kubernetes",
								Revision: "0.1.1",
							},
						},
					},
				}),
			},
			requiredTIInjectionSatisfiedBy: []*gqlpublicapi.TypeInstanceValue{
				{
					TypeRef: &gqlpublicapi.TypeReferenceInput{
						Path:     "cap.type.gcp.sa",
						Revision: "0.1.1",
					},
				},
			},
			requirementsSatisfiedBy: []*gqlpublicapi.TypeInstanceValue{
				{
					TypeRef: &gqlpublicapi.TypeReferenceInput{
						Path:     "cap.core.type.platform.kubernetes",
						Revision: "0.1.1",
					},
				},
			},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			// given
			filter := gqlpublicapi.ImplementationRevisionFilter{
				RequirementsSatisfiedBy:                   tt.requirementsSatisfiedBy,
				RequiredTypeInstancesInjectionSatisfiedBy: tt.requiredTIInjectionSatisfiedBy,
			}

			getOpts := &ListImplementationRevisionsForInterfaceOptions{}
			getOpts.Apply(WithFilter(filter))

			allRevs := append(tt.expRevision, tt.revisionToFilterOut...)

			// when
			gotRevs := FilterImplementationRevisions(allRevs, getOpts)

			// then
			assert.Len(t, gotRevs, len(tt.expRevision))
			for idx := range tt.expRevision {
				assert.Contains(t, gotRevs, tt.expRevision[idx])
			}
		})
	}
}

func TestImplementationPathPatternFilters(t *testing.T) {
	// given
	expRevision := []gqlpublicapi.ImplementationRevision{
		fixImplementationRevision("cap.implementation.db.postgres.install", "0.0.1"),
		fixImplementationRevision("cap.implementation.db.postgres.uninstall", "0.0.1"),
	}

	revisionToFilterOut := []gqlpublicapi.ImplementationRevision{
		fixImplementationRevision("cap.implementation.db.rds.install", "0.0.1"),
		fixImplementationRevision("cap.implementation.db.rds.uninstall", "0.0.1"),
	}

	filter := gqlpublicapi.ImplementationRevisionFilter{
		PathPattern: ptr.String("cap.implementation.db.postgres.*"),
	}

	getOpts := &ListImplementationRevisionsForInterfaceOptions{}
	getOpts.Apply(WithFilter(filter))

	allRevs := append(expRevision, revisionToFilterOut...)

	// when
	gotRevs := FilterImplementationRevisions(allRevs, getOpts)

	// then
	assert.Len(t, gotRevs, len(expRevision))
	for idx := range expRevision {
		assert.Contains(t, gotRevs, expRevision[idx])
	}
}

func TestImplementationRequiresFilters(t *testing.T) {
	tests := []struct {
		name                string
		expRevision         []gqlpublicapi.ImplementationRevision
		revisionToFilterOut []gqlpublicapi.ImplementationRevision
		requires            []*gqlpublicapi.TypeReferenceWithOptionalRevision
	}{
		{
			name: "Return Implementations that requires GCP SA",
			expRevision: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevisionWithRequire("with-gcp-sa-requirement", "0.1.1", gqlpublicapi.ImplementationRequirement{
					AnyOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.gcp.sa",
								Revision: "0.1.1",
							},
						},
					},
				}),
				fixImplementationRevisionWithRequire("with-multiple-requirements", "0.1.0", gqlpublicapi.ImplementationRequirement{
					AnyOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.aws.subscription",
								Revision: "0.1.0",
							},
						},
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.gcp.sa",
								Revision: "0.1.1",
							},
						},
					},
				}),
			},
			revisionToFilterOut: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevision("without-any-requirements", "0.0.1"),
				fixImplementationRevisionWithRequire("with-cf-requirement", "0.1.0", gqlpublicapi.ImplementationRequirement{
					AllOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.core.type.platform.cf",
								Revision: "0.1.1",
							},
						},
					},
				}),
				fixImplementationRevisionWithRequire("with-gcp-sa-requirement-diff-version", "0.1.0", gqlpublicapi.ImplementationRequirement{
					OneOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.gcp.sa",
								Revision: "0.1.2",
							},
						},
					},
				}),
			},
			requires: []*gqlpublicapi.TypeReferenceWithOptionalRevision{
				{
					Path:     "cap.type.gcp.sa",
					Revision: ptr.String("0.1.1"),
				},
			},
		},
		{
			name: "Return Implementations that has GCP SA and AWS subscription without revision in `requires` section",
			expRevision: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevisionWithRequire("with-multiple-requirements", "0.1.0", gqlpublicapi.ImplementationRequirement{
					AnyOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path: "cap.type.aws.subscription",
							},
						},
					},
					OneOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.gcp.sa",
								Revision: "0.1.1",
							},
						},
					},
				}),
			},
			revisionToFilterOut: []gqlpublicapi.ImplementationRevision{
				fixImplementationRevisionWithRequire("with-aws-subscription-requirement", "0.1.0", gqlpublicapi.ImplementationRequirement{
					AnyOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.aws.subscription",
								Revision: "0.1.1",
							},
						},
					},
				}),
				fixImplementationRevisionWithRequire("with-gcp-sa-requirement", "0.1.0", gqlpublicapi.ImplementationRequirement{
					AnyOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.type.gcp.sa",
								Revision: "0.1.1",
							},
						},
					},
				}),
				fixImplementationRevisionWithRequire("with-cf-requirement", "0.1.0", gqlpublicapi.ImplementationRequirement{
					AllOf: []*gqlpublicapi.ImplementationRequirementItem{
						{
							TypeRef: &gqlpublicapi.TypeReference{
								Path:     "cap.core.type.platform.cf",
								Revision: "0.1.1",
							},
						},
					},
				}),
			},
			requires: []*gqlpublicapi.TypeReferenceWithOptionalRevision{
				{
					Path:     "cap.type.gcp.sa",
					Revision: ptr.String("0.1.1"),
				},
				{
					Path: "cap.type.aws.subscription",
				},
			},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			// given
			filter := gqlpublicapi.ImplementationRevisionFilter{
				Requires: tt.requires,
			}

			getOpts := &ListImplementationRevisionsForInterfaceOptions{}
			getOpts.Apply(WithFilter(filter))

			allRevs := append(tt.expRevision, tt.revisionToFilterOut...)

			// when
			gotRevs := FilterImplementationRevisions(allRevs, getOpts)

			// then
			assert.Len(t, gotRevs, len(tt.expRevision))
			for idx := range tt.expRevision {
				assert.Contains(t, gotRevs, tt.expRevision[idx])
			}
		})
	}
}

func fixImplementationRevisionWithRequire(implPath, implRev string, req gqlpublicapi.ImplementationRequirement) gqlpublicapi.ImplementationRevision {
	impl := fixImplementationRevision(implPath, implRev)
	impl.Spec.Requires = []*gqlpublicapi.ImplementationRequirement{
		&req,
	}

	return impl
}

func fixImplementationRevisionWithAttr(implPath, implRev, attrPath, attrRev string) gqlpublicapi.ImplementationRevision {
	impl := fixImplementationRevision(implPath, implRev)
	impl.Metadata.Attributes = []*gqlpublicapi.AttributeRevision{
		{
			Metadata: &gqlpublicapi.GenericMetadata{
				Path: attrPath,
			},
			Revision: attrRev,
		},
	}

	return impl
}

func fixImplementationRevision(path, rev string) gqlpublicapi.ImplementationRevision {
	return gqlpublicapi.ImplementationRevision{
		Metadata: &gqlpublicapi.ImplementationMetadata{
			Path: path,
		},
		Spec:     &gqlpublicapi.ImplementationSpec{},
		Revision: rev,
	}
}
