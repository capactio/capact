//go:build integration
// +build integration

package e2e

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/internal/regexutil"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	hubclient "capact.io/capact/pkg/hub/client"
	"capact.io/capact/pkg/hub/client/public"

	"github.com/MakeNowJust/heredoc"
	prmt "github.com/gitchander/permutation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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

type StorageSpec struct {
	URL           *string `json:"url,omitempty"`
	AcceptValue   *bool   `json:"acceptValue,omitempty"`
	ContextSchema *string `json:"contextSchema,omitempty"`
}

func includes(ids []string, expID string) bool {
	for _, i := range ids {
		if i == expID {
			return true
		}
	}

	return false
}

func registerExternalStorage(ctx context.Context, cli *hubclient.Client, value interface{}) (string, func()) {
	storage := &gqllocalapi.CreateTypeInstanceInput{
		TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
			Path:     "cap.type.example.filesystem.storage",
			Revision: "0.1.0",
		},
		Value: value,
	}

	externalStorageID, err := cli.CreateTypeInstance(ctx, storage)
	Expect(err).NotTo(HaveOccurred())
	Expect(externalStorageID).NotTo(BeEmpty())

	return externalStorageID, func() {
		_ = cli.DeleteTypeInstance(ctx, externalStorageID)
	}
}

func typeInstance(ver string) *gqllocalapi.CreateTypeInstanceInput {
	return &gqllocalapi.CreateTypeInstanceInput{
		TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
			Path:     "cap.type.capactio.capact.validation.single-key",
			Revision: "0.1.0",
		},
		Attributes: []*gqllocalapi.AttributeReferenceInput{
			{
				Path:     "cap.type.sample-v" + ver,
				Revision: "0.1.0",
			},
		},
		Value: map[string]interface{}{
			"key": "sample-v" + ver,
		},
	}
}

func assertTypeInstance(ctx context.Context, cli *hubclient.Client, ID string, expected *gqllocalapi.TypeInstance) {
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

func findCreatedTypeInstanceID(alias string, instances []gqllocalapi.CreateTypeInstanceOutput) *string {
	for _, el := range instances {
		if el.Alias != alias {
			continue
		}
		return &el.ID
	}

	return nil
}

func deleteTypeInstance(ctx context.Context, cli *hubclient.Client, ID string) {
	err := cli.DeleteTypeInstance(ctx, ID)
	Expect(err).ToNot(HaveOccurred())
}

func createTypeInstancesInput() *gqllocalapi.CreateTypeInstancesInput {
	return &gqllocalapi.CreateTypeInstancesInput{
		TypeInstances: []*gqllocalapi.CreateTypeInstanceInput{
			{
				Alias: ptr.String("parent"),
				TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
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
					"parent": true,
				},
			},
			{
				Alias: ptr.String("child"),
				TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
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
					"child": true,
				},
			},
		},
		UsesRelations: []*gqllocalapi.TypeInstanceUsesRelationInput{
			{
				From: "parent",
				To:   "child",
			},
		},
	}
}

func expectedChildTypeInstance(tiID, backendID string) *gqllocalapi.TypeInstance {
	tiRev := &gqllocalapi.TypeInstanceResourceVersion{
		ResourceVersion: 1,
		Metadata: &gqllocalapi.TypeInstanceResourceVersionMetadata{
			Attributes: []*gqllocalapi.AttributeReference{
				{
					Path:     "com.attr",
					Revision: "0.1.0",
				},
			},
		},
		Spec: &gqllocalapi.TypeInstanceResourceVersionSpec{
			Value: map[string]interface{}{
				"child": true,
			},
			Backend: &gqllocalapi.TypeInstanceResourceVersionSpecBackend{},
		},
	}

	return &gqllocalapi.TypeInstance{
		ID: tiID,
		TypeRef: &gqllocalapi.TypeInstanceTypeReference{
			Path:     "com.child",
			Revision: "0.1.0",
		},

		Backend: &gqllocalapi.TypeInstanceBackendReference{
			ID:       backendID,
			Abstract: true,
		},
		LatestResourceVersion:   tiRev,
		FirstResourceVersion:    tiRev,
		PreviousResourceVersion: nil,
		ResourceVersion:         tiRev,
		ResourceVersions:        []*gqllocalapi.TypeInstanceResourceVersion{tiRev},
		UsedBy:                  nil,
		Uses:                    nil,
	}
}

func expectedParentTypeInstance(tiID, backendID string) *gqllocalapi.TypeInstance {
	tiRev := &gqllocalapi.TypeInstanceResourceVersion{
		ResourceVersion: 1,
		Metadata: &gqllocalapi.TypeInstanceResourceVersionMetadata{
			Attributes: []*gqllocalapi.AttributeReference{
				{
					Path:     "com.attr",
					Revision: "0.1.0",
				},
			},
		},
		Spec: &gqllocalapi.TypeInstanceResourceVersionSpec{
			Value: map[string]interface{}{
				"parent": true,
			},
			Backend: &gqllocalapi.TypeInstanceResourceVersionSpecBackend{},
		},
	}

	return &gqllocalapi.TypeInstance{
		ID: tiID,
		TypeRef: &gqllocalapi.TypeInstanceTypeReference{
			Path:     "com.parent",
			Revision: "0.1.0",
		},

		Backend: &gqllocalapi.TypeInstanceBackendReference{
			ID:       backendID,
			Abstract: true,
		},
		LatestResourceVersion:   tiRev,
		FirstResourceVersion:    tiRev,
		PreviousResourceVersion: nil,
		ResourceVersion:         tiRev,
		ResourceVersions:        []*gqllocalapi.TypeInstanceResourceVersion{tiRev},
		UsedBy:                  nil,
		Uses:                    nil,
	}
}

// We cannot write specs using Context and It which are connected with each other
// see: https://github.com/onsi/ginkgo/issues/246
// this functions just adds syntax sugar which can be used when we have just one single `It` block
// with sequential test-cases
func scenario(format string, args ...interface{}) {
	fmt.Fprintf(GinkgoWriter, "[Scenario]: "+format+"\n", args...)
}

func when(format string, args ...interface{}) {
	fmt.Fprintf(GinkgoWriter, "\t[when]: "+format+"\n", args...)
}

func then(format string, args ...interface{}) {
	fmt.Fprintf(GinkgoWriter, "\t[then]: "+format+"\n", args...)
}

// allPermutations returns all possible permutations in regex format.
// For such input
//	a := []string{"alpha", "beta"}
// returns
//	("alpha", "beta"|"beta", "alpha")
//
// This function allows you to match list of words in any order using regex.
func allPermutations(in []string) string {
	p := prmt.New(prmt.StringSlice(in))
	var opts []string
	for p.Next() {
		opts = append(opts, fmt.Sprintf(`"%s"`, strings.Join(in, `", "`)))
	}
	return regexutil.OrStringSlice(opts)
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
