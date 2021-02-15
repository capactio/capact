package public

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"projectvoltron.dev/voltron/internal/ptr"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
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

			getOpts := &getImplementationOptions{}
			getOpts.Apply(WithImplementationFilter(filter))

			allRevs := append(tt.expRevision, tt.revisionToFilterOut...)

			// when
			gotRevs := filterImplementationRevisions(allRevs, getOpts)

			// then
			assert.Len(t, gotRevs, len(tt.expRevision))
			for idx := range tt.expRevision {
				assert.Contains(t, gotRevs, tt.expRevision[idx])
			}
		})
	}
}

func TestImplementationRequirementFilters(t *testing.T) {
	tests := []struct {
		name                    string
		expRevision             []gqlpublicapi.ImplementationRevision
		revisionToFilterOut     []gqlpublicapi.ImplementationRevision
		requirementsSatisfiedBy []*gqlpublicapi.TypeInstanceValue
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
					Prefix: "cap.core.type.platform",
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
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			// given
			filter := gqlpublicapi.ImplementationRevisionFilter{
				RequirementsSatisfiedBy: tt.requirementsSatisfiedBy,
			}

			getOpts := &getImplementationOptions{}
			getOpts.Apply(WithImplementationFilter(filter))

			allRevs := append(tt.expRevision, tt.revisionToFilterOut...)

			// when
			gotRevs := filterImplementationRevisions(allRevs, getOpts)

			// then
			assert.Len(t, gotRevs, len(tt.expRevision))
			for idx := range tt.expRevision {
				assert.Contains(t, gotRevs, tt.expRevision[idx])
			}
		})
	}
}

func TestImplementationRequirementPrefixPattern(t *testing.T) {
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
		PrefixPattern: ptr.String("cap.implementation.db.postgres.*"),
	}

	getOpts := &getImplementationOptions{}
	getOpts.Apply(WithImplementationFilter(filter))

	allRevs := append(expRevision, revisionToFilterOut...)

	// when
	gotRevs := filterImplementationRevisions(allRevs, getOpts)

	// then
	assert.Len(t, gotRevs, len(expRevision))
	for idx := range expRevision {
		assert.Contains(t, gotRevs, expRevision[idx])
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
				Path: ptr.String(attrPath),
			},
			Revision: attrRev,
		},
	}

	return impl
}

func fixImplementationRevision(path, rev string) gqlpublicapi.ImplementationRevision {
	return gqlpublicapi.ImplementationRevision{
		Metadata: &gqlpublicapi.ImplementationMetadata{
			Path:   path,
			Prefix: ptr.String(path),
		},
		Spec:     &gqlpublicapi.ImplementationSpec{},
		Revision: rev,
	}
}
