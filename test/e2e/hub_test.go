//go:build integration
// +build integration

package e2e

import (
	"context"
	"fmt"
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

	Context("Local Hub", func() {
		It("should create, find and delete TypeInstance", func() {
			cli := getHubGraphQLClient()
			builtinStorage := getBuiltinStorageTypeInstance(ctx, cli)

			// create TypeInstance
			createdTypeInstance, err := cli.CreateTypeInstance(ctx, &gqllocalapi.CreateTypeInstanceInput{
				TypeRef: &gqllocalapi.TypeInstanceTypeReferenceInput{
					Path:     "cap.type.capactio.capact.validation.single-key",
					Revision: "0.1.0",
				},
				Attributes: []*gqllocalapi.AttributeReferenceInput{
					{
						Path:     "cap.type.capactio.capact.attribute1",
						Revision: "0.1.0",
					},
				},
				Value: map[string]interface{}{
					"key": "bar",
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
							Path:     "cap.type.capactio.capact.attribute1",
							Revision: "0.1.0",
						},
					},
				},
				Spec: &gqllocalapi.TypeInstanceResourceVersionSpec{
					Value: map[string]interface{}{
						"key": "bar",
					},
				},
			}
			Expect(typeInstance).To(Equal(&gqllocalapi.TypeInstance{
				ID: createdTypeInstance.ID,
				TypeRef: &gqllocalapi.TypeInstanceTypeReference{
					Path:     "cap.type.capactio.capact.validation.single-key",
					Revision: "0.1.0",
				},
				Backend: &gqllocalapi.TypeInstanceBackendReference{
					ID:       builtinStorage.ID,
					Abstract: true,
				},
				Uses:                    []*gqllocalapi.TypeInstance{&builtinStorage},
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
			cli := getHubGraphQLClient()
			builtinStorage := getBuiltinStorageTypeInstance(ctx, cli)

			createdTypeInstanceIDs, err := cli.CreateTypeInstances(ctx, createTypeInstancesInput())

			Expect(err).NotTo(HaveOccurred())
			for _, ti := range createdTypeInstanceIDs {
				defer deleteTypeInstance(ctx, cli, ti.ID)
			}

			parentTiID := findCreatedTypeInstanceID("parent", createdTypeInstanceIDs)
			Expect(parentTiID).ToNot(BeNil())

			childTiID := findCreatedTypeInstanceID("child", createdTypeInstanceIDs)
			Expect(childTiID).ToNot(BeNil())

			expectedChild := expectedChildTypeInstance(*childTiID, builtinStorage.ID)
			expectedParent := expectedParentTypeInstance(*parentTiID, builtinStorage.ID)
			expectedChild.UsedBy = []*gqllocalapi.TypeInstance{expectedParentTypeInstance(*parentTiID, builtinStorage.ID)}
			expectedChild.Uses = []*gqllocalapi.TypeInstance{&builtinStorage}
			expectedParent.Uses = []*gqllocalapi.TypeInstance{&builtinStorage, expectedChildTypeInstance(*childTiID, builtinStorage.ID)}
			expectedParent.UsedBy = []*gqllocalapi.TypeInstance{}

			assertTypeInstance(ctx, cli, *childTiID, expectedChild)
			assertTypeInstance(ctx, cli, *parentTiID, expectedParent)
		})

		It("should lock TypeInstances based on all edge cases", func() {
			const (
				fooOwnerID = "namespace/Foo"
				barOwnerID = "namespace/Bar"
			)
			localCli := getHubGraphQLClient()

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
			err := localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
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
			err = localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
				Ids:     createdTIIDs, // lock all 3 instances, when the first two are already locked
				OwnerID: fooOwnerID,
			})
			Expect(err).NotTo(HaveOccurred())

			then("should success")
			got, err = localCli.ListTypeInstances(ctx, &gqllocalapi.TypeInstanceFilter{})
			Expect(err).NotTo(HaveOccurred())

			for _, instance := range got {
				if !includes(createdTIIDs, instance.ID) {
					continue
				}
				Expect(instance.LockedBy).NotTo(BeNil())
				Expect(*instance.LockedBy).To(Equal(fooOwnerID))
			}

			scenario("id1, id2, id3 are locked by Foo, id4: not found")
			lockingIDs := createdTIIDs
			lockingIDs = append(lockingIDs, "123-not-found")

			when("Foo tries to locks id1,id2,id3,id4")
			err = localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
				Ids:     lockingIDs,
				OwnerID: fooOwnerID,
			})

			then("should failed with id4 not found error")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(heredoc.Doc(`while executing mutation to lock TypeInstances: All attempts fail:
							#1: graphql: failed to lock TypeInstances: 1 error occurred: TypeInstances with IDs "123-not-found" were not found`)))

			when("Bar tries to locks id1,id2,id3,id4")
			err = localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
				Ids:     lockingIDs,
				OwnerID: barOwnerID,
			})

			then("should failed with id4 not found and already locked error for id1,id2,id3")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(heredoc.Docf(`while executing mutation to lock TypeInstances: All attempts fail:
				#1: graphql: failed to lock TypeInstances: 2 errors occurred: \[TypeInstances with IDs "123-not-found" were not found, TypeInstances with IDs %s are locked by different owner\]`, allPermutations(createdTIIDs))))

			scenario("id1, id2, id3 are locked by Foo, id4: not locked")
			when("Bar tries to locks all of them")
			id4, err := localCli.CreateTypeInstance(ctx, typeInstance("id4"))
			Expect(err).ToNot(HaveOccurred())

			defer localCli.DeleteTypeInstance(ctx, id4.ID)

			lockingIDs = createdTIIDs
			lockingIDs = append(lockingIDs, id4.ID)
			err = localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
				Ids:     lockingIDs,
				OwnerID: barOwnerID,
			})

			then("should failed with error id1,id2,id3 already locked by Foo")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(heredoc.Docf(`while executing mutation to lock TypeInstances: All attempts fail:
						#1: graphql: failed to lock TypeInstances: 1 error occurred: TypeInstances with IDs %s are locked by different owner`, allPermutations(createdTIIDs))))

			scenario("id1, id2, id3 are locked by Foo, id4: not locked")

			when("Bar tries to locks all of them")
			err = localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
				Ids:     lockingIDs,
				OwnerID: barOwnerID,
			})

			then("should failed with error id1,id2,id3 already locked by Foo")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(heredoc.Docf(`while executing mutation to lock TypeInstances: All attempts fail:
						#1: graphql: failed to lock TypeInstances: 1 error occurred: TypeInstances with IDs %s are locked by different owner`, allPermutations(createdTIIDs))))

			then("should unlock id1,id2,id3")
			err = localCli.UnlockTypeInstances(ctx, &gqllocalapi.UnlockTypeInstancesInput{
				Ids:     createdTIIDs,
				OwnerID: fooOwnerID,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should test update TypeInstances based on all edge cases", func() {
			const (
				fooOwnerID = "namespace/Foo"
				barOwnerID = "namespace/Bar"
			)
			localCli := getHubGraphQLClient()

			var createdTIIDs []string

			for _, ver := range []string{"id1", "id2"} {
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
			expUpdateTI := &gqllocalapi.UpdateTypeInstanceInput{
				Attributes: []*gqllocalapi.AttributeReferenceInput{
					{Path: "cap.update.not.locked", Revision: "0.0.1"},
				},
			}

			when("try to update them")
			updatedTI, err := localCli.UpdateTypeInstances(ctx, []gqllocalapi.UpdateTypeInstancesInput{
				{
					ID:           createdTIIDs[0],
					TypeInstance: expUpdateTI,
				},
				{
					ID:           createdTIIDs[1],
					TypeInstance: expUpdateTI,
				},
			})

			then("should success")
			Expect(err).NotTo(HaveOccurred())
			for _, instance := range updatedTI {
				Expect(instance.LatestResourceVersion.Metadata.Attributes).To(HaveLen(1))
				Expect(instance.LatestResourceVersion.Metadata.Attributes[0]).To(BeEquivalentTo(expUpdateTI.Attributes[0]))
			}

			scenario("id1 and id2 are locked by Foo")
			expUpdateTI = &gqllocalapi.UpdateTypeInstanceInput{
				Attributes: []*gqllocalapi.AttributeReferenceInput{
					{Path: "cap.update.locked.by.foo", Revision: "0.0.1"},
				},
			}

			err = localCli.LockTypeInstances(ctx, &gqllocalapi.LockTypeInstancesInput{
				Ids:     createdTIIDs,
				OwnerID: fooOwnerID,
			})
			Expect(err).NotTo(HaveOccurred())

			when("update them as Foo owner")
			updatedTI, err = localCli.UpdateTypeInstances(ctx, []gqllocalapi.UpdateTypeInstancesInput{
				{
					ID:           createdTIIDs[0],
					OwnerID:      ptr.String(fooOwnerID),
					TypeInstance: expUpdateTI,
				},
				{
					ID:           createdTIIDs[1],
					OwnerID:      ptr.String(fooOwnerID),
					TypeInstance: expUpdateTI,
				},
			})

			then("should success")
			Expect(err).NotTo(HaveOccurred())
			for _, instance := range updatedTI {
				Expect(instance.LatestResourceVersion.Metadata.Attributes).To(HaveLen(1))
				Expect(instance.LatestResourceVersion.Metadata.Attributes[0]).To(BeEquivalentTo(expUpdateTI.Attributes[0]))
			}

			when("update them as Bar owner")
			_, err = localCli.UpdateTypeInstances(ctx, []gqllocalapi.UpdateTypeInstancesInput{
				{
					ID:           createdTIIDs[0],
					OwnerID:      ptr.String(barOwnerID),
					TypeInstance: expUpdateTI,
				},
				{
					ID:           createdTIIDs[1],
					OwnerID:      ptr.String(barOwnerID),
					TypeInstance: expUpdateTI,
				},
			})

			then("should failed with error id1,id2 already locked by different owner")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(heredoc.Docf(`while executing mutation to update TypeInstances: All attempts fail:
        				#1: graphql: failed to update TypeInstances: TypeInstances with IDs %s are locked by different owner`, allPermutations(createdTIIDs))))

			when("update them without owner")
			_, err = localCli.UpdateTypeInstances(ctx, []gqllocalapi.UpdateTypeInstancesInput{
				{
					ID:           createdTIIDs[0],
					TypeInstance: expUpdateTI,
				},
				{
					ID:           createdTIIDs[1],
					TypeInstance: expUpdateTI,
				},
			})

			then("should failed with error id1,id2 already locked by different owner")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp(heredoc.Docf(`while executing mutation to update TypeInstances: All attempts fail:
        				#1: graphql: failed to update TypeInstances: TypeInstances with IDs %s are locked by different owner`, allPermutations(createdTIIDs))))

			when("update one property with Foo owner, and second without owner")
			_, err = localCli.UpdateTypeInstances(ctx, []gqllocalapi.UpdateTypeInstancesInput{
				{
					ID:           createdTIIDs[0],
					OwnerID:      ptr.String(fooOwnerID),
					TypeInstance: expUpdateTI,
				},
				{
					ID:           createdTIIDs[1],
					TypeInstance: expUpdateTI,
				},
			})

			then("should failed with error id2 already locked by different owner")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(heredoc.Docf(`while executing mutation to update TypeInstances: All attempts fail:
        				#1: graphql: failed to update TypeInstances: TypeInstances with IDs "%s" are locked by different owner`, createdTIIDs[1])))

			scenario("id3 does not exist")
			when("try to update it")
			_, err = localCli.UpdateTypeInstances(ctx, []gqllocalapi.UpdateTypeInstancesInput{
				{
					ID:           "id3",
					TypeInstance: expUpdateTI,
				},
			})

			then("should failed with error id3 not found")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(heredoc.Doc(`while executing mutation to update TypeInstances: All attempts fail:
        			#1: graphql: failed to update TypeInstances: TypeInstances with IDs "id3" were not found`)))

			then("should unlock id1,id2,id3")
			err = localCli.UnlockTypeInstances(ctx, &gqllocalapi.UnlockTypeInstancesInput{
				Ids:     createdTIIDs,
				OwnerID: fooOwnerID,
			})
			Expect(err).NotTo(HaveOccurred())
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
