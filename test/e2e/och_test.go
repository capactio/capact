// +build integration

package e2e

import (
	"context"
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
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

	Context("Local OCH", func() {
		It("should create, find and delete TypeInstance", func() {
			cli := getOCHGraphQLClient()

			// create TypeInstance
			createdTypeInstance, err := cli.CreateTypeInstance(ctx, &gqllocalapi.CreateTypeInstanceInput{
				TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
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

			// check create TypeInstance
			typeInstance, err := cli.FindTypeInstance(ctx, createdTypeInstance.ID)

			Expect(err).ToNot(HaveOccurred())
			rev := &gqllocalapi.TypeInstanceResourceVersion{
				ResourceVersion: 1,
				Metadata: &gqllocalapi.TypeInstanceResourceVersionMetadata{
					Attributes: []*gqllocalapi.AttributeReference{
						{
							Path:     "com.voltron.attribute1",
							Revision: "0.1.0",
						},
					},
				},
				Spec: &gqllocalapi.TypeInstanceResourceVersionSpec{
					Value: map[string]interface{}{
						"foo": "bar",
					},
				},
			}
			Expect(typeInstance).To(Equal(&gqllocalapi.TypeInstance{
				ID: createdTypeInstance.ID,
				TypeRef: &gqllocalapi.TypeInstanceTypeReference{
					Path:     "com.voltron.ti",
					Revision: "0.1.0",
				},
				Uses:                    []*gqllocalapi.TypeInstance{},
				UsedBy:                  []*gqllocalapi.TypeInstance{},
				LatestResourceVersion:   rev,
				FirstResourceVersion:    rev,
				PreviousResourceVersion: nil,
				ResourceVersion:         rev,
				ResourceVersions:        []*gqllocalapi.TypeInstanceResourceVersion{rev},
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
			expectedChild.UsedBy = []*gqllocalapi.TypeInstance{expectedParentTypeInstance(*parentTiID)}
			expectedChild.Uses = []*gqllocalapi.TypeInstance{}
			expectedParent.Uses = []*gqllocalapi.TypeInstance{expectedChildTypeInstance(*childTiID)}
			expectedParent.UsedBy = []*gqllocalapi.TypeInstance{}

			assertTypeInstance(ctx, cli, *childTiID, expectedChild)
			assertTypeInstance(ctx, cli, *parentTiID, expectedParent)
		})

		It("should lock TypeInstances based on all edge cases", func() {
			const (
				fooOwnerID = "namespace/Foo"
				barOwnerID = "namespace/Bar"
			)
			localCli := getOCHGraphQLClient()

			var createdTIIDs []string

			for _, ver := range []string{"id1", "id2", "id3"} {
				out, err := localCli.CreateTypeInstance(ctx, typeInstance(ver))
				Expect(err).NotTo(HaveOccurred())
				createdTIIDs = append(createdTIIDs, out.ID)
			}
			defer func() {
				for _, id := range createdTIIDs {
					_ = localCli.DeleteTypeInstance(ctx, id)
				}

			}()

			scenario("id1 and id2 are not locked")
			firstTwoInstances := createdTIIDs[:2]
			lastInstances := createdTIIDs[2:]

			when("Foo tries to locks them")
			err := localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstanceInput{
				Ids:     firstTwoInstances,
				OwnerID: fooOwnerID,
			})
			Expect(err).NotTo(HaveOccurred())

			then("should success")
			got, err := localCli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{})
			Expect(err).NotTo(HaveOccurred())

			for _, instance := range got {
				if includes(firstTwoInstances, instance.ID) {
					Expect(instance.LockedBy).NotTo(BeNil())
					Expect(*instance.LockedBy).To(Equal(fooOwnerID))
				} else if includes(lastInstances, instance.ID) {
					Expect(instance.LockedBy).To(BeNil())
				}
			}

			scenario("id1 and id2 are locked by Foo, id3: not locked")
			when("Foo tries to locks them")
			err = localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstanceInput{
				Ids:     createdTIIDs, // lock all 3 instances, when the first two are already locked
				OwnerID: fooOwnerID,
			})
			Expect(err).NotTo(HaveOccurred())

			then("should success")
			got, err = localCli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{})
			Expect(err).NotTo(HaveOccurred())

			for _, instance := range got {
				Expect(instance.LockedBy).NotTo(BeNil())
				Expect(*instance.LockedBy).To(Equal(fooOwnerID))
			}

			scenario("id1, id2, id3 are locked by Foo, id4: not found")
			lockingIDs := createdTIIDs
			lockingIDs = append(lockingIDs, "123-not-found")

			when("Foo tries to locks id1,id2,id3,id4")
			err = localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstanceInput{
				Ids:     lockingIDs,
				OwnerID: fooOwnerID,
			})

			then("should failed with id4 not found error")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(heredoc.Doc(`while executing mutation to lock TypeInstances: All attempts fail:
							#1: graphql: failed to lock TypeInstances: 1 error occurred: TypeInstances with IDs 123-not-found were not found`)))

			when("Bar tries to locks id1,id2,id3,id4")
			err = localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstanceInput{
				Ids:     lockingIDs,
				OwnerID: barOwnerID,
			})

			then("should failed with id4 not found and already locked error for id1,id2,id3")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(heredoc.Docf(`while executing mutation to lock TypeInstances: All attempts fail:
				#1: graphql: failed to lock TypeInstances: 2 errors occurred: [TypeInstances with IDs 123-not-found were not found, TypeInstances with IDs %s are locked by other owner]`, strings.Join(createdTIIDs, ", "))))

			scenario("id1, id2, id3 are locked by Foo, id4: not locked")
			when("Bar tries to locks all of them")
			id4, err := localCli.CreateTypeInstance(ctx, typeInstance("id4"))
			Expect(err).ToNot(HaveOccurred())

			defer localCli.DeleteTypeInstance(ctx, id4.ID)

			lockingIDs = createdTIIDs
			lockingIDs = append(lockingIDs, id4.ID)
			err = localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstanceInput{
				Ids:     lockingIDs,
				OwnerID: barOwnerID,
			})

			then("should failed with error id1,id2,id3 already locked by Foo")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(heredoc.Docf(`while executing mutation to lock TypeInstances: All attempts fail:
						#1: graphql: failed to lock TypeInstances: 1 error occurred: TypeInstances with IDs %s are locked by other owner`, strings.Join(createdTIIDs, ", "))))

			scenario("id1, id2, id3 are locked by Foo, id4: not locked")

			when("Bar tries to locks all of them")
			err = localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstanceInput{
				Ids:     lockingIDs,
				OwnerID: barOwnerID,
			})

			then("should failed with error id1,id2,id3 already locked by Foo")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(heredoc.Docf(`while executing mutation to lock TypeInstances: All attempts fail:
						#1: graphql: failed to lock TypeInstances: 1 error occurred: TypeInstances with IDs %s are locked by other owner`, strings.Join(createdTIIDs, ", "))))
		})
	})
})

func includes(ids []string, expID string) bool {
	for _, i := range ids {
		if i == expID {
			return true
		}
	}

	return false
}

func typeInstance(ver string) *gqllocalapi.CreateTypeInstanceInput {
	return &gqllocalapi.CreateTypeInstanceInput{
		TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
			Path:     "cap.type.sample-v" + ver,
			Revision: "0.1.0",
		},
		Attributes: []*gqllocalapi.AttributeReferenceInput{
			{
				Path:     "cap.type.sample-v" + ver,
				Revision: "0.1.0",
			},
		},
		Value: map[string]interface{}{
			"sample-v" + ver: true,
		},
	}
}

func assertTypeInstance(ctx context.Context, cli *ochclient.Client, ID string, expected *gqllocalapi.TypeInstance) {
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

func deleteTypeInstance(ctx context.Context, cli *ochclient.Client, ID string) {
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

func expectedChildTypeInstance(ID string) *gqllocalapi.TypeInstance {
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
		},
	}

	return &gqllocalapi.TypeInstance{
		ID: ID,
		TypeRef: &gqllocalapi.TypeInstanceTypeReference{
			Path:     "com.child",
			Revision: "0.1.0",
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

func expectedParentTypeInstance(ID string) *gqllocalapi.TypeInstance {
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
		},
	}

	return &gqllocalapi.TypeInstance{
		ID: ID,
		TypeRef: &gqllocalapi.TypeInstanceTypeReference{
			Path:     "com.parent",
			Revision: "0.1.0",
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

func scenario(format string, args ...interface{}) {
	fmt.Fprintf(GinkgoWriter, "[Scenario]: "+format+"\n", args...)
}

func when(format string, args ...interface{}) {
	fmt.Fprintf(GinkgoWriter, "\t[when]: "+format+"\n", args...)
}

func then(format string, args ...interface{}) {
	fmt.Fprintf(GinkgoWriter, "\t[then]: "+format+"\n", args...)
}
