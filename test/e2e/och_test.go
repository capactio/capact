// +build integration

package e2e

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"projectvoltron.dev/voltron/internal/ptr"
	gqllocalapiv2 "projectvoltron.dev/voltron/pkg/och/api/graphql/local-v2"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
	ochclient "projectvoltron.dev/voltron/pkg/och/client"
	"projectvoltron.dev/voltron/pkg/och/client/public"
)

var _ = Describe("GraphQL API", func() {
	ctx := context.Background()

	Context("Public OCH", func() {
		var cli *ochclient.Client

		BeforeEach(func() {
			cli = getOCHGraphQLClient()
		})

		Describe("should return ImplementationRevision", func() {
			const (
				interfacePath  = "cap.interface.voltron.ochtests.install"
				latestRevision = "2.0.0"
				revision       = "1.0.0"
			)

			It("for latest Interface when revision is not specified", func() {
				revisionsForInterface, err := cli.ListImplementationRevisionsForInterface(ctx, gqlpublicapi.InterfaceReference{
					Path: interfacePath,
				})

				Expect(err).ToNot(HaveOccurred())
				Expect(revisionsForInterface).To(HaveLen(2))
				for _, rev := range revisionsForInterface {
					Expect(rev.Spec.Implements).To(HaveLen(1))
					Expect(rev.Spec.Implements[0].Path).To(Equal(interfacePath))
					Expect(rev.Spec.Implements[0].Revision).To(Equal(latestRevision))
				}
			})

			It("for a specific Interface when revision is defined", func() {
				revisionsForInterface, err := cli.ListImplementationRevisionsForInterface(ctx, gqlpublicapi.InterfaceReference{
					Path:     interfacePath,
					Revision: revision,
				})

				Expect(err).ToNot(HaveOccurred())
				Expect(revisionsForInterface).To(HaveLen(2))
				for _, rev := range revisionsForInterface {
					Expect(rev.Spec.Implements).To(HaveLen(1))
					Expect(rev.Spec.Implements[0].Path).To(Equal(interfacePath))
					Expect(rev.Spec.Implements[0].Revision).To(Equal(revision))
				}

			})

			It("with specified pathPattern", func() {
				ref := gqlpublicapi.InterfaceReference{
					Path:     interfacePath,
					Revision: revision,
				}
				pathPattern := "cap.implementation.voltron.*"
				filter := gqlpublicapi.ImplementationRevisionFilter{
					PathPattern: ptr.String(pathPattern),
				}

				revisionsForInterface, err := cli.ListImplementationRevisionsForInterface(ctx, ref, public.WithFilter(filter))

				Expect(err).ToNot(HaveOccurred())
				Expect(revisionsForInterface).To(HaveLen(1))

				Expect(revisionsForInterface[0].Spec.Implements).To(HaveLen(1))
				Expect(revisionsForInterface[0].Spec.Implements[0].Path).To(Equal(ref.Path))
				Expect(revisionsForInterface[0].Spec.Implements[0].Revision).To(Equal(ref.Revision))

				Expect(*revisionsForInterface[0].Metadata.Prefix).To(MatchRegexp(pathPattern))

			})

			It("with attribute", func() {
				ref := gqlpublicapi.InterfaceReference{
					Path:     interfacePath,
					Revision: revision,
				}

				attr := attributeFilterInput("cap.attribute.cloud.provider.gcp", "0.1.0", gqlpublicapi.FilterRuleInclude)
				filter := gqlpublicapi.ImplementationRevisionFilter{
					Attributes: []*gqlpublicapi.AttributeFilterInput{&attr},
				}

				revisionsForInterface, err := cli.ListImplementationRevisionsForInterface(ctx, ref, public.WithFilter(filter))

				Expect(err).ToNot(HaveOccurred())
				Expect(revisionsForInterface).To(HaveLen(1))

				Expect(revisionsForInterface[0].Spec.Implements).To(HaveLen(1))
				Expect(revisionsForInterface[0].Spec.Implements[0].Path).To(Equal(ref.Path))
				Expect(revisionsForInterface[0].Spec.Implements[0].Revision).To(Equal(ref.Revision))

				Expect(revisionsForInterface[0].Metadata.Attributes).To(HaveLen(1))
				Expect(revisionsForInterface[0].Metadata.Attributes[0].Metadata.Path).To(Equal(attr.Path))
				Expect(revisionsForInterface[0].Metadata.Attributes[0].Revision).To(Equal(*attr.Revision))

			})
			It("without attribute", func() {
				ref := gqlpublicapi.InterfaceReference{
					Path:     interfacePath,
					Revision: revision,
				}

				attr := attributeFilterInput("cap.attribute.cloud.provider.gcp", "0.1.0", gqlpublicapi.FilterRuleExclude)
				filter := gqlpublicapi.ImplementationRevisionFilter{
					Attributes: []*gqlpublicapi.AttributeFilterInput{&attr},
				}

				revisionsForInterface, err := cli.ListImplementationRevisionsForInterface(ctx, ref, public.WithFilter(filter))

				Expect(err).ToNot(HaveOccurred())
				Expect(revisionsForInterface).To(HaveLen(1))

				Expect(revisionsForInterface[0].Spec.Implements).To(HaveLen(1))
				Expect(revisionsForInterface[0].Spec.Implements[0].Path).To(Equal(ref.Path))
				Expect(revisionsForInterface[0].Spec.Implements[0].Revision).To(Equal(ref.Revision))

				Expect(revisionsForInterface[0].Metadata.Attributes).To(BeEmpty())

			})

			It("satisfied by Kubernetes platform", func() {
				ref := gqlpublicapi.InterfaceReference{
					Path:     interfacePath,
					Revision: revision,
				}

				filter := gqlpublicapi.ImplementationRevisionFilter{
					RequirementsSatisfiedBy: []*gqlpublicapi.TypeInstanceValue{
						{
							TypeRef: &gqlpublicapi.TypeReferenceWithOptionalRevision{
								Path:     "cap.core.type.platform.kubernetes",
								Revision: ptr.String("0.1.0"),
							},
						},
					},
				}

				revisionsForInterface, err := cli.ListImplementationRevisionsForInterface(ctx, ref, public.WithFilter(filter))

				Expect(err).ToNot(HaveOccurred())
				Expect(revisionsForInterface).To(HaveLen(1))

				Expect(revisionsForInterface[0].Spec.Implements).To(HaveLen(1))
				Expect(revisionsForInterface[0].Spec.Implements[0].Path).To(Equal(ref.Path))
				Expect(revisionsForInterface[0].Spec.Implements[0].Revision).To(Equal(ref.Revision))

				Expect(revisionsForInterface[0].Spec.Requires).To(HaveLen(1))
				Expect(revisionsForInterface[0].Spec.Requires[0].Prefix).To(Equal("cap.core.type.platform"))

			})
		})
	})

	Context("Local OCH v2", func() {
		It("should create, find and delete TypeInstance", func() {
			cli := getOCHGraphQLClient()

			// create TypeInstance
			createdTypeInstance, err := cli.CreateTypeInstance(ctx, &gqllocalapiv2.CreateTypeInstanceInput{
				TypeRef: &gqllocalapiv2.TypeInstanceTypeReferenceInput{
					Path:     "com.voltron.ti",
					Revision: "0.1.0",
				},
				Attributes: []*gqllocalapiv2.AttributeReferenceInput{
					{
						Path:     "com.voltron.attribute1",
						Revision: "0.1.0",
					},
				},
				Value: map[string]interface{}{
					"foo": "bar",
				},
			})
			Expect(err).ToNot(HaveOccurred())

			// check create TypeInstance
			typeInstance, err := cli.FindTypeInstance(ctx, createdTypeInstance.ID)

			Expect(err).ToNot(HaveOccurred())
			rev := &gqllocalapiv2.TypeInstanceResourceVersion{
				ResourceVersion: 1,
				Metadata: &gqllocalapiv2.TypeInstanceResourceVersionMetadata{
					Attributes: []*gqllocalapiv2.AttributeReference{
						{
							Path:     "com.voltron.attribute1",
							Revision: "0.1.0",
						},
					},
				},
				Spec: &gqllocalapiv2.TypeInstanceResourceVersionSpec{
					Value: map[string]interface{}{
						"foo": "bar",
					},
				},
			}
			Expect(typeInstance).To(Equal(&gqllocalapiv2.TypeInstance{
				ID: createdTypeInstance.ID,
				TypeRef: &gqllocalapiv2.TypeInstanceTypeReference{
					Path:     "com.voltron.ti",
					Revision: "0.1.0",
				},
				Uses:                    []*gqllocalapiv2.TypeInstance{},
				UsedBy:                  []*gqllocalapiv2.TypeInstance{},
				LatestResourceVersion:   rev,
				FirstResourceVersion:    rev,
				PreviousResourceVersion: nil,
				ResourceVersion:         rev,
				ResourceVersions:        []*gqllocalapiv2.TypeInstanceResourceVersion{rev},
			}))

			// check delete TypeInstance
			err = cli.DeleteTypeInstance(ctx, createdTypeInstance.ID)
			Expect(err).ToNot(HaveOccurred())

			got, err := cli.FindTypeInstance(ctx, createdTypeInstance.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(BeNil())
		})

		It("creates multiple TypeInstances with uses relations", func() {
			cli := getOCHGraphQLClient()

			createdTypeInstanceIDs, err := cli.CreateTypeInstances(ctx, createTypeInstancesInput())

			Expect(err).NotTo(HaveOccurred())
			for _, ti := range createdTypeInstanceIDs {
				defer deleteTypeInstance(ctx, cli, ti.ID)
			}

			parentTiID := findCreatedTypeInstanceID("parent", createdTypeInstanceIDs)
			Expect(parentTiID).ToNot(BeNil())

			childTiID := findCreatedTypeInstanceID("child", createdTypeInstanceIDs)
			Expect(childTiID).ToNot(BeNil())

			expectedChild := expectedChildTypeInstance(*childTiID)
			expectedParent := expectedParentTypeInstance(*parentTiID)
			expectedChild.UsedBy = []*gqllocalapiv2.TypeInstance{expectedParentTypeInstance(*parentTiID)}
			expectedChild.Uses = []*gqllocalapiv2.TypeInstance{}
			expectedParent.Uses = []*gqllocalapiv2.TypeInstance{expectedChildTypeInstance(*childTiID)}
			expectedParent.UsedBy = []*gqllocalapiv2.TypeInstance{}

			assertTypeInstance(ctx, cli, *childTiID, expectedChild)
			assertTypeInstance(ctx, cli, *parentTiID, expectedParent)
		})
	})

})

type OCHMode string

func assertTypeInstance(ctx context.Context, cli *ochclient.Client, ID string, expected *gqllocalapiv2.TypeInstance) {
	actual, err := cli.FindTypeInstance(ctx, ID)
	Expect(err).NotTo(HaveOccurred())
	Expect(actual).NotTo(BeNil())
	Expect(expected).NotTo(BeNil())
	Expect(*actual).To(Equal(*expected))
}

func attributeFilterInput(path, rev string, rule gqlpublicapi.FilterRule) gqlpublicapi.AttributeFilterInput {
	return gqlpublicapi.AttributeFilterInput{
		Path:     path,
		Rule:     &rule,
		Revision: ptr.String(rev),
	}
}

func findCreatedTypeInstanceID(alias string, instances []gqllocalapiv2.CreateTypeInstanceOutput) *string {
	for _, el := range instances {
		if el.Alias != alias {
			continue
		}
		return &el.ID
	}

	return nil
}

func deleteTypeInstance(ctx context.Context, cli *ochclient.Client, ID string) {
	err := cli.DeleteTypeInstance(ctx, ID)
	Expect(err).ToNot(HaveOccurred())
}

func createTypeInstancesInput() *gqllocalapiv2.CreateTypeInstancesInput {
	return &gqllocalapiv2.CreateTypeInstancesInput{
		TypeInstances: []*gqllocalapiv2.CreateTypeInstanceInput{
			{
				Alias: ptr.String("parent"),
				TypeRef: &gqllocalapiv2.TypeInstanceTypeReferenceInput{
					Path:     "com.parent",
					Revision: "0.1.0",
				},
				Attributes: []*gqllocalapiv2.AttributeReferenceInput{
					{
						Path:     "com.attr",
						Revision: "0.1.0",
					},
				},
				Value: map[string]interface{}{
					"parent": true,
				},
			},
			{
				Alias: ptr.String("child"),
				TypeRef: &gqllocalapiv2.TypeInstanceTypeReferenceInput{
					Path:     "com.child",
					Revision: "0.1.0",
				},
				Attributes: []*gqllocalapiv2.AttributeReferenceInput{
					{
						Path:     "com.attr",
						Revision: "0.1.0",
					},
				},
				Value: map[string]interface{}{
					"child": true,
				},
			},
		},
		UsesRelations: []*gqllocalapiv2.TypeInstanceUsesRelationInput{
			{
				From: "parent",
				To:   "child",
			},
		},
	}
}

func expectedChildTypeInstance(ID string) *gqllocalapiv2.TypeInstance {
	tiRev := &gqllocalapiv2.TypeInstanceResourceVersion{
		ResourceVersion: 1,
		Metadata: &gqllocalapiv2.TypeInstanceResourceVersionMetadata{
			Attributes: []*gqllocalapiv2.AttributeReference{
				{
					Path:     "com.attr",
					Revision: "0.1.0",
				},
			},
		},
		Spec: &gqllocalapiv2.TypeInstanceResourceVersionSpec{
			Value: map[string]interface{}{
				"child": true,
			},
		},
	}

	return &gqllocalapiv2.TypeInstance{
		ID: ID,
		TypeRef: &gqllocalapiv2.TypeInstanceTypeReference{
			Path:     "com.child",
			Revision: "0.1.0",
		},
		LatestResourceVersion:   tiRev,
		FirstResourceVersion:    tiRev,
		PreviousResourceVersion: nil,
		ResourceVersion:         tiRev,
		ResourceVersions:        []*gqllocalapiv2.TypeInstanceResourceVersion{tiRev},
		UsedBy:                  nil,
		Uses:                    nil,
	}
}

func expectedParentTypeInstance(ID string) *gqllocalapiv2.TypeInstance {
	tiRev := &gqllocalapiv2.TypeInstanceResourceVersion{
		ResourceVersion: 1,
		Metadata: &gqllocalapiv2.TypeInstanceResourceVersionMetadata{
			Attributes: []*gqllocalapiv2.AttributeReference{
				{
					Path:     "com.attr",
					Revision: "0.1.0",
				},
			},
		},
		Spec: &gqllocalapiv2.TypeInstanceResourceVersionSpec{
			Value: map[string]interface{}{
				"parent": true,
			},
		},
	}

	return &gqllocalapiv2.TypeInstance{
		ID: ID,
		TypeRef: &gqllocalapiv2.TypeInstanceTypeReference{
			Path:     "com.parent",
			Revision: "0.1.0",
		},
		LatestResourceVersion:   tiRev,
		FirstResourceVersion:    tiRev,
		PreviousResourceVersion: nil,
		ResourceVersion:         tiRev,
		ResourceVersions:        []*gqllocalapiv2.TypeInstanceResourceVersion{tiRev},
		UsedBy:                  nil,
		Uses:                    nil,
	}
}
