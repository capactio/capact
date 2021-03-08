// +build integration

package e2e

import (
	"context"
	"fmt"
	"time"

	"github.com/machinebox/graphql"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	cliappsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"projectvoltron.dev/voltron/internal/ptr"
	"projectvoltron.dev/voltron/pkg/httputil"
	gqllocalapi "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	gqllocalapiv2 "projectvoltron.dev/voltron/pkg/och/api/graphql/local-v2"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
	ochclient "projectvoltron.dev/voltron/pkg/och/client"
	ochv2cli "projectvoltron.dev/voltron/pkg/och/client/local/v2"
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
			defer deleteTypeInstance(ctx, cli, createdTypeInstance.Metadata.ID)

			typeInstance, err := cli.FindTypeInstance(ctx, createdTypeInstance.Metadata.ID)

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
				Uses:   []*gqllocalapi.TypeInstance{},
				UsedBy: []*gqllocalapi.TypeInstance{},
			}))
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

			assertTypeInstance(ctx, cli, *childTiID, expectedChildTypeInstance(*childTiID, *parentTiID))
			assertTypeInstance(ctx, cli, *parentTiID, expectedParentTypeInstance(*parentTiID, *childTiID))
		})
	})

	Context("Local OCH v2", func() {
		// TODO(SV-266): temporary solution
		It("should switch to Local OCH v2 mode", func() {
			clientset, err := kubernetes.NewForConfig(config.GetConfigOrDie())
			Expect(err).NotTo(HaveOccurred())
			cli := clientset.AppsV1().Deployments(cfg.OCHLocalDeployNamespace)

			By("setting OCH_MODE to local-v2")
			mergePatch := ochLocalModePatch(OCHModeLocalV2)
			newDeploy, err := cli.Patch(cfg.OCHLocalDeployName, types.StrategicMergePatchType, mergePatch)
			Expect(err).NotTo(HaveOccurred())

			err = statusReady(cli, cfg.OCHLocalDeployName, newDeploy.Generation)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create, find and delete TypeInstance", func() {
			cli := newOCHLocalV2Client()

			// create TypeInstance
			createdTypeInstance, err := cli.CreateTypeInstance(ctx, &gqllocalapiv2.CreateTypeInstanceInput{
				TypeRef: &gqllocalapiv2.TypeReferenceInput{
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
				TypeRef: &gqllocalapiv2.TypeReference{
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
			Skip("Not yet implemented")
		})

		// TODO(SV-266): temporary solution
		It("should switch back the Local OCH v1 mode", func() {
			clientset, err := kubernetes.NewForConfig(config.GetConfigOrDie())
			Expect(err).NotTo(HaveOccurred())
			cli := clientset.AppsV1().Deployments(cfg.OCHLocalDeployNamespace)

			By("setting OCH_MODE to local")
			mergePatch := ochLocalModePatch(OCHModeLocal)
			newDeploy, err := cli.Patch(cfg.OCHLocalDeployName, types.StrategicMergePatchType, mergePatch)
			Expect(err).NotTo(HaveOccurred())

			err = statusReady(cli, cfg.OCHLocalDeployName, newDeploy.Generation)
			Expect(err).NotTo(HaveOccurred())
		})
	})

})

// Same configuration as we have for argo-actions app
// Note: skip gateway as it still has wrong schema, additional pod restart will cost more time
// TODO(SV-266): temporary solution
func newOCHLocalV2Client() *ochv2cli.Client {
	httpClient := httputil.NewClient(
		30*time.Second,
		true,
	)

	clientOpt := graphql.WithHTTPClient(httpClient)
	endpoint := fmt.Sprintf("http://%s.%s/graphql", cfg.OCHLocalDeployName, cfg.OCHLocalDeployNamespace)
	gcli := graphql.NewClient(endpoint, clientOpt)

	return ochv2cli.NewClient(gcli)
}

// TODO(SV-266): temporary solution
func statusReady(cli cliappsv1.DeploymentInterface, deployName string, expGen int64) error {
	return wait.Poll(cfg.PollingInterval, cfg.PollingTimeout, func() (done bool, err error) {
		dep, err := cli.Get(deployName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return dep.Status.ObservedGeneration == expGen && dep.Status.Replicas == dep.Status.ReadyReplicas, nil
	})
}

type OCHMode string

const (
	OCHModeLocalV2 OCHMode = "local-v2"
	OCHModeLocal           = "local"
)

// TODO(SV-266): temporary solution
func ochLocalModePatch(mode OCHMode) []byte {
	return []byte(fmt.Sprintf(`{
		  "spec": {
			"template": {
			  "spec": {
				"containers": [
				  {
					"env": [
					  {
						"name": "APP_OCH_MODE",
						"value": "%s"
					  }
					],
					"name": "och-local"
				  }
				]
			  }
			}
		  }
		}`, mode))
}

func assertTypeInstance(ctx context.Context, cli *ochclient.Client, ID string, expected *gqllocalapi.TypeInstance) {
	childTI, err := cli.FindTypeInstance(ctx, ID)
	Expect(err).NotTo(HaveOccurred())
	Expect(childTI).To(Equal(expected))
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
		if el.Alias == alias {
			return &el.ID
		}
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
				Alias: ptr.String("child"),
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
				From: "parent",
				To:   "child",
			},
		},
	}
}

func expectedChildTypeInstance(ID string, parentID string) *gqllocalapi.TypeInstance {
	return &gqllocalapi.TypeInstance{
		ResourceVersion: 1,
		Metadata: &gqllocalapi.TypeInstanceMetadata{
			ID: ID,
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
		Uses: []*gqllocalapi.TypeInstance{},
		UsedBy: []*gqllocalapi.TypeInstance{
			{
				Metadata: &gqllocalapi.TypeInstanceMetadata{
					ID: parentID,
				},
				Spec: &gqllocalapi.TypeInstanceSpec{
					TypeRef: &gqllocalapi.TypeReference{
						Path:     "com.parent",
						Revision: "0.1.0",
					},
				},
			},
		},
	}
}

func expectedParentTypeInstance(ID string, childID string) *gqllocalapi.TypeInstance {
	return &gqllocalapi.TypeInstance{
		ResourceVersion: 1,
		Metadata: &gqllocalapi.TypeInstanceMetadata{
			ID: ID,
			Attributes: []*gqllocalapi.AttributeReference{
				{
					Path:     "com.attr",
					Revision: "0.1.0",
				},
			},
		},
		Spec: &gqllocalapi.TypeInstanceSpec{
			TypeRef: &gqllocalapi.TypeReference{
				Path:     "com.parent",
				Revision: "0.1.0",
			},
			Value: map[string]interface{}{
				"foo": "bar",
			},
		},
		Uses: []*gqllocalapi.TypeInstance{
			{
				Metadata: &gqllocalapi.TypeInstanceMetadata{
					ID: childID,
				},
				Spec: &gqllocalapi.TypeInstanceSpec{
					TypeRef: &gqllocalapi.TypeReference{
						Path:     "com.child",
						Revision: "0.1.0",
					},
				},
			},
		},
		UsedBy: []*gqllocalapi.TypeInstance{},
	}
}
