// +build integration

package e2e

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"projectvoltron.dev/voltron/internal/ptr"
	gqllocalapi "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
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
				revisionsForInterface, err := cli.GetImplementationRevisionsForInterface(ctx, gqlpublicapi.InterfaceReference{
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
				revisionsForInterface, err := cli.GetImplementationRevisionsForInterface(ctx, gqlpublicapi.InterfaceReference{
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

			It("with specified prefixPattern", func() {
				ref := gqlpublicapi.InterfaceReference{
					Path:     interfacePath,
					Revision: revision,
				}
				prefixPattern := "cap.implementation.voltron.*"
				filter := gqlpublicapi.ImplementationRevisionFilter{
					PrefixPattern: ptr.String(prefixPattern),
				}

				revisionsForInterface, err := cli.GetImplementationRevisionsForInterface(ctx, ref, public.WithImplementationFilter(filter))

				Expect(err).ToNot(HaveOccurred())
				Expect(revisionsForInterface).To(HaveLen(1))

				Expect(revisionsForInterface[0].Spec.Implements).To(HaveLen(1))
				Expect(revisionsForInterface[0].Spec.Implements[0].Path).To(Equal(ref.Path))
				Expect(revisionsForInterface[0].Spec.Implements[0].Revision).To(Equal(ref.Revision))

				Expect(*revisionsForInterface[0].Metadata.Prefix).To(MatchRegexp(prefixPattern))

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

				revisionsForInterface, err := cli.GetImplementationRevisionsForInterface(ctx, ref, public.WithImplementationFilter(filter))

				Expect(err).ToNot(HaveOccurred())
				Expect(revisionsForInterface).To(HaveLen(1))

				Expect(revisionsForInterface[0].Spec.Implements).To(HaveLen(1))
				Expect(revisionsForInterface[0].Spec.Implements[0].Path).To(Equal(ref.Path))
				Expect(revisionsForInterface[0].Spec.Implements[0].Revision).To(Equal(ref.Revision))

				Expect(revisionsForInterface[0].Metadata.Attributes).To(HaveLen(1))
				Expect(*revisionsForInterface[0].Metadata.Attributes[0].Metadata.Path).To(Equal(attr.Path))
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

				revisionsForInterface, err := cli.GetImplementationRevisionsForInterface(ctx, ref, public.WithImplementationFilter(filter))

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
							TypeRef: &gqlpublicapi.TypeReferenceInput{
								Path:     "cap.core.type.platform.kubernetes",
								Revision: ptr.String("0.1.0"),
							},
						},
					},
				}

				revisionsForInterface, err := cli.GetImplementationRevisionsForInterface(ctx, ref, public.WithImplementationFilter(filter))

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

	Context("Local OCH", func() {
		typeInstancesToCleanup := []string{}

		AfterEach(func() {
			cli := getOCHGraphQLClient()
			for _, ID := range typeInstancesToCleanup {
				deleteTypeInstance(ctx, cli, ID)
			}
		})

		It("creates and deletes TypeInstance", func() {
			cli := getOCHGraphQLClient()

			createdTypeInstance, err := cli.CreateTypeInstance(ctx, &gqllocalapi.CreateTypeInstanceInput{
				TypeRef: &gqllocalapi.TypeReferenceInput{
					Path:     "com.voltron.ti",
					Revision: "0.1.0",
				},
				Attributes: []*gqllocalapi.AttributeReferenceInput{
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

			typeInstance, err := cli.GetTypeInstance(ctx, createdTypeInstance.Metadata.ID)

			Expect(err).ToNot(HaveOccurred())
			Expect(typeInstance).To(Equal(&gqllocalapi.TypeInstance{
				ResourceVersion: 1,
				Metadata: &gqllocalapi.TypeInstanceMetadata{
					ID: createdTypeInstance.Metadata.ID,
					Attributes: []*gqllocalapi.AttributeReference{
						{
							Path:     "com.voltron.attribute1",
							Revision: "0.1.0",
						},
					},
				},
				Spec: &gqllocalapi.TypeInstanceSpec{
					TypeRef: &gqllocalapi.TypeReference{
						Path:     "com.voltron.ti",
						Revision: "0.1.0",
					},
					Value: map[string]interface{}{
						"foo": "bar",
					},
				},
				Uses: []*gqllocalapi.TypeInstance{},
			}))

			deleteTypeInstance(ctx, cli, typeInstance.Metadata.ID)
		})

		It("creates multiple TypeInstances with uses relations", func() {
			cli := getOCHGraphQLClient()

			createdTypeInstanceIDs, err := cli.CreateTypeInstances(ctx, &gqllocalapi.CreateTypeInstancesInput{
				TypeInstances: []*gqllocalapi.CreateTypeInstanceInput{
					{
						Alias: "parent",
						TypeRef: &gqllocalapi.TypeReferenceInput{
							Path:     "com.parent",
							Revision: "0.1.0",
						},
						Attributes: []*gqllocalapi.AttributeReferenceInput{
							{
								Path:     "com.attr",
								Revision: "0.1.0",
							},
						},
						Value: map[string]interface{}{
							"foo": "bar",
						},
					},
					{
						Alias: "child",
						TypeRef: &gqllocalapi.TypeReferenceInput{
							Path:     "com.child",
							Revision: "0.1.0",
						},
						Attributes: []*gqllocalapi.AttributeReferenceInput{
							{
								Path:     "com.attr",
								Revision: "0.1.0",
							},
						},
						Value: map[string]interface{}{
							"foo": "bar",
						},
					},
				},
				UsesRelations: []*gqllocalapi.TypeInstanceUsesRelationInput{
					{
						From: "child",
						To:   "parent",
					},
				},
			})

			Expect(err).NotTo(HaveOccurred())
			typeInstancesToCleanup = append(typeInstancesToCleanup, createdTypeInstanceIDs...)

			typeInstances, err := cli.ListTypeInstances(ctx, gqllocalapi.TypeInstanceFilter{
				TypeRef: &gqllocalapi.TypeRefFilterInput{
					Path: "com.child",
				},
			})

			Expect(err).NotTo(HaveOccurred())

			Expect(typeInstances[0]).To(Equal(gqllocalapi.TypeInstance{
				ResourceVersion: 1,
				Metadata: &gqllocalapi.TypeInstanceMetadata{
					ID: createdTypeInstanceIDs[1],
					Attributes: []*gqllocalapi.AttributeReference{
						{
							Path:     "com.attr",
							Revision: "0.1.0",
						},
					},
				},
				Spec: &gqllocalapi.TypeInstanceSpec{
					TypeRef: &gqllocalapi.TypeReference{
						Path:     "com.child",
						Revision: "0.1.0",
					},
					Value: map[string]interface{}{
						"foo": "bar",
					},
				},
				Uses: []*gqllocalapi.TypeInstance{
					{
						Metadata: &gqllocalapi.TypeInstanceMetadata{
							ID: createdTypeInstanceIDs[0],
						},
					},
				},
			}))
		})
	})
})

func attributeFilterInput(path, rev string, rule gqlpublicapi.FilterRule) gqlpublicapi.AttributeFilterInput {
	return gqlpublicapi.AttributeFilterInput{
		Path:     path,
		Rule:     &rule,
		Revision: ptr.String(rev),
	}
}

func deleteTypeInstance(ctx context.Context, cli *ochclient.Client, ID string) {
	err := cli.DeleteTypeInstance(ctx, ID)
	Expect(err).ToNot(HaveOccurred())
}
