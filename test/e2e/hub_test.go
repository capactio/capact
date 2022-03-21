//go:build integration
// +build integration

package e2e

import (
	"capact.io/capact/internal/ptr"
	"capact.io/capact/internal/regexutil"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	hubclient "capact.io/capact/pkg/hub/client"
	"capact.io/capact/pkg/hub/client/public"
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GraphQL API", func() {
	ctx := context.Background()

	Context("Public Hub", func() {
		var cli *hubclient.Client

		BeforeEach(func() {
			cli = getHubGraphQLClient()
		})

		Describe("should return Types", func() {
			It("based on full path name", func() {
				const fullTypePath = "cap.core.type.platform.kubernetes"

				gotTypes, err := cli.ListTypes(ctx, public.WithTypeFilter(gqlpublicapi.TypeFilter{
					PathPattern: ptr.String(fullTypePath),
				}))

				Expect(err).ToNot(HaveOccurred())

				Expect(gotTypes).To(HaveLen(1))
				Expect(gotTypes[0].Path).To(Equal(fullTypePath))
			})

			It("based on a prefix of the parent node", func() {
				const parentNode = "cap.core.type.platform.*"
				expAssociatedPaths := []string{"cap.core.type.platform.kubernetes", "cap.type.platform.nomad"}

				gotTypes, err := cli.ListTypes(ctx, public.WithTypeFilter(gqlpublicapi.TypeFilter{
					PathPattern: ptr.String(parentNode),
				}))

				Expect(err).ToNot(HaveOccurred())

				HasOnlyExpectTypePaths(gotTypes, expAssociatedPaths)
			})
			It("only child node if full path name specified", func() {
				const fullTypePath = "cap.type.platform.nomad"

				gotTypes, err := cli.ListTypes(ctx, public.WithTypeFilter(gqlpublicapi.TypeFilter{
					PathPattern: ptr.String(fullTypePath),
				}))

				Expect(err).ToNot(HaveOccurred())

				Expect(gotTypes).To(HaveLen(1))
				Expect(gotTypes[0].Path).To(Equal(fullTypePath))
			})
			It("entries matching or regex (cap.core.type.generic.value|cap.type.platform.nomad)", func() {
				expTypePaths := []string{"cap.core.type.generic.value", "cap.type.platform.nomad"}
				typePathORFilter := regexutil.OrStringSlice(expTypePaths)

				gotTypes, err := cli.ListTypes(ctx, public.WithTypeFilter(gqlpublicapi.TypeFilter{
					PathPattern: ptr.String(typePathORFilter),
				}))

				Expect(err).ToNot(HaveOccurred())

				HasOnlyExpectTypePaths(gotTypes, expTypePaths)
			})
			It("all entries if there is no filter", func() {
				gotTypes, err := cli.ListTypes(ctx, public.WithTypeFilter(gqlpublicapi.TypeFilter{
					// no path filter
				}))

				Expect(err).ToNot(HaveOccurred())

				Expect(len(gotTypes)).Should(BeNumerically(">=", 50))
			})
			It("no entries if prefix is not a regex and there is no Type with such explicit path", func() {
				const parentNode = "cap.core.type.platform"

				gotTypes, err := cli.ListTypes(ctx, public.WithTypeFilter(gqlpublicapi.TypeFilter{
					PathPattern: ptr.String(parentNode),
				}))

				Expect(err).ToNot(HaveOccurred())
				Expect(gotTypes).To(HaveLen(0))
			})
		})
		Describe("should return ImplementationRevision", func() {
			const (
				interfacePath  = "cap.interface.capactio.capact.validation.hub.install"
				latestRevision = "2.0.0"
				revision       = "1.0.0"
			)

			It("for latest Interface when revision is not specified", func() {
				revisionsForInterface, err := cli.ListImplementationRevisionsForInterface(ctx, gqlpublicapi.InterfaceReference{
					Path: interfacePath,
				})

				Expect(err).ToNot(HaveOccurred())
				Expect(revisionsForInterface).To(HaveLen(3))
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
				pathPattern := "cap.implementation.capactio.capact.validation.own.*"
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
							TypeRef: &gqlpublicapi.TypeReferenceInput{
								Path:     "cap.core.type.platform.kubernetes",
								Revision: "0.1.0",
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
})

func attributeFilterInput(path, rev string, rule gqlpublicapi.FilterRule) gqlpublicapi.AttributeFilterInput {
	return gqlpublicapi.AttributeFilterInput{
		Path:     path,
		Rule:     &rule,
		Revision: ptr.String(rev),
	}
}

func HasOnlyExpectTypePaths(gotTypes []*gqlpublicapi.Type, expectedPaths []string) {
	Expect(gotTypes).To(HaveLen(len(expectedPaths)))
	var gotPaths []string
	for _, t := range gotTypes {
		if t == nil {
			continue
		}
		gotPaths = append(gotPaths, t.Path)
	}
	Expect(gotPaths).To(ConsistOf(expectedPaths))
}
