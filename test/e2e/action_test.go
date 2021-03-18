// +build integration

package e2e

import (
	"context"
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"projectvoltron.dev/voltron/internal/ptr"
	ochlocalgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	ochclient "projectvoltron.dev/voltron/pkg/och/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	enginegraphql "projectvoltron.dev/voltron/pkg/engine/api/graphql"
	engine "projectvoltron.dev/voltron/pkg/engine/client"
)

const clusterPolicyConfigMapKey = "cluster-policy.yaml"
const clusterPolicyTokenToReplace = "{typeInstanceUUID}"

var _ = Describe("Action", func() {
	var engineClient *engine.Client
	var ochClient *ochclient.Client

	actionName := "e2e-test"
	ctx := context.Background()

	BeforeEach(func() {
		engineClient = getEngineGraphQLClient()
		ochClient = getOCHGraphQLClient()
	})

	AfterEach(func() {
		// cleanup Action
		engineClient.DeleteAction(ctx, actionName)
	})

	Context("Action execution", func() {
		It("should have succeeded status after a passed workflow", func() {
			_, err := engineClient.CreateAction(ctx, &enginegraphql.ActionDetailsInput{
				Name: actionName,
				ActionRef: &enginegraphql.ManifestReferenceInput{
					Path:     "cap.interface.voltron.e2e.passing",
					Revision: ptr.String("0.1.0"),
				},
			})

			Expect(err).ToNot(HaveOccurred())

			Eventually(
				getActionStatusFunc(ctx, engineClient, actionName),
				cfg.PollingTimeout, cfg.PollingInterval,
			).Should(Equal(enginegraphql.ActionStatusPhaseReadyToRun))

			err = engineClient.RunAction(ctx, actionName)

			Expect(err).ToNot(HaveOccurred())

			Eventually(
				getActionStatusFunc(ctx, engineClient, actionName),
				cfg.PollingTimeout, cfg.PollingInterval,
			).Should(Equal(enginegraphql.ActionStatusPhaseSucceeded))
		})

		It("should pick proper Implementation and inject TypeInstance based on cluster policy", func() {
			actionPath := "cap.interface.voltron.policy.test"

			log("1. Expecting Implementation A is picked based on test policy and requirements met...")

			action := createActionAndWaitForReadyToRunPhase(ctx, engineClient, actionName, actionPath)
			assertActionRenderedWorkflowContains(action, "echo 'Implementation A'")
			runActionAndWaitForSucceeded(ctx, engineClient, actionName)

			log("2. Deleting Action...")

			err := engineClient.DeleteAction(ctx, actionName)
			Expect(err).ToNot(HaveOccurred())

			log("3. Creating TypeInstance and modifying Policy to make Implementation B picked for next run...")

			// 3.1. Create TypeInstance which is required for second Implementation to be picked
			typeInstanceValue := getTypeInstanceInputForPolicy()
			typeInstance, tiCleanupFn := createTypeInstance(ctx, ochClient, typeInstanceValue)
			defer tiCleanupFn()

			// 3.2. Update cluster policy with the TypeInstance ID to inject for the most preferred Implementation (Implementation B)
			typeInstanceID := typeInstance.ID
			cfgMapCleanupFn := updateClusterPolicyConfigMap(clusterPolicyTokenToReplace, typeInstanceID)
			defer cfgMapCleanupFn()

			log("4. Expecting Implementation B is picked based on test policy...")

			action = createActionAndWaitForReadyToRunPhase(ctx, engineClient, actionName, actionPath)
			assertActionRenderedWorkflowContains(action, "echo 'Implementation B'")
			runActionAndWaitForSucceeded(ctx, engineClient, actionName)
		})

		It("should have failed status after a failed workflow", func() {
			_, err := engineClient.CreateAction(ctx, &enginegraphql.ActionDetailsInput{
				Name: actionName,
				ActionRef: &enginegraphql.ManifestReferenceInput{
					Path:     "cap.interface.voltron.e2e.failing",
					Revision: ptr.String("0.1.0"),
				},
			})

			Expect(err).ToNot(HaveOccurred())

			Eventually(
				getActionStatusFunc(ctx, engineClient, actionName),
				cfg.PollingTimeout, cfg.PollingInterval,
			).Should(Equal(enginegraphql.ActionStatusPhaseReadyToRun))

			err = engineClient.RunAction(ctx, actionName)

			Expect(err).ToNot(HaveOccurred())

			Eventually(
				getActionStatusFunc(ctx, engineClient, actionName),
				cfg.PollingTimeout, cfg.PollingInterval,
			).Should(Equal(enginegraphql.ActionStatusPhaseFailed))
		})

		It("should download input TypeInstance", func() {
			var typeInstances []*enginegraphql.InputTypeInstanceData
			input := &ochlocalgraphql.CreateTypeInstanceInput{
				TypeRef: &ochlocalgraphql.TypeInstanceTypeReferenceInput{
					Path:     "cap.type.simple.single-key",
					Revision: "0.1.0",
				},
				Value: map[string]interface{}{"key": true},
				Attributes: []*ochlocalgraphql.AttributeReferenceInput{
					{
						Path:     "com.voltron.attribute1",
						Revision: "0.1.0",
					},
				},
			}
			simpleTI, simpleTICleanupFn := createTypeInstance(ctx, ochClient, input)
			defer simpleTICleanupFn()

			typeInstances = append(typeInstances,
				&enginegraphql.InputTypeInstanceData{Name: "simple-key-value", ID: simpleTI.ID})

			input = &ochlocalgraphql.CreateTypeInstanceInput{
				TypeRef: &ochlocalgraphql.TypeInstanceTypeReferenceInput{
					Path:     "cap.type.gcp.auth.service-account",
					Revision: "0.1.0",
				},
				Value: map[string]string{"project": "voltron"},
				Attributes: []*ochlocalgraphql.AttributeReferenceInput{
					{
						Path:     "com.voltron.attribute1",
						Revision: "0.1.0",
					},
				},
			}
			saTypeInstance, saTICleanupFn := createTypeInstance(ctx, ochClient, input)
			defer saTICleanupFn()
			typeInstances = append(typeInstances,
				&enginegraphql.InputTypeInstanceData{Name: "gcp", ID: saTypeInstance.ID})

			_, err := engineClient.CreateAction(ctx, &enginegraphql.ActionDetailsInput{
				Name: actionName,
				ActionRef: &enginegraphql.ManifestReferenceInput{
					Path:     "cap.interface.voltron.e2e.type-instance-download",
					Revision: ptr.String("0.1.0"),
				},
				Input: &enginegraphql.ActionInputData{
					TypeInstances: typeInstances,
				},
			})

			Expect(err).ToNot(HaveOccurred())

			Eventually(
				getActionStatusFunc(ctx, engineClient, actionName),
				cfg.PollingTimeout, cfg.PollingInterval,
			).Should(Equal(enginegraphql.ActionStatusPhaseReadyToRun))

			err = engineClient.RunAction(ctx, actionName)

			Expect(err).ToNot(HaveOccurred())

			Eventually(
				getActionStatusFunc(ctx, engineClient, actionName),
				cfg.PollingTimeout, cfg.PollingInterval,
			).Should(Equal(enginegraphql.ActionStatusPhaseSucceeded))
		})

		It("should upload output TypeInstances", func() {
			_, err := engineClient.CreateAction(ctx, &enginegraphql.ActionDetailsInput{
				Name: actionName,
				ActionRef: &enginegraphql.ManifestReferenceInput{
					Path:     "cap.interface.voltron.e2e.type-instance-upload",
					Revision: ptr.String("0.1.0"),
				},
				Input: &enginegraphql.ActionInputData{},
			})

			Expect(err).ToNot(HaveOccurred())

			Eventually(
				getActionStatusFunc(ctx, engineClient, actionName),
				cfg.PollingTimeout, cfg.PollingInterval,
			).Should(Equal(enginegraphql.ActionStatusPhaseReadyToRun))

			err = engineClient.RunAction(ctx, actionName)

			Expect(err).ToNot(HaveOccurred())

			Eventually(
				getActionStatusFunc(ctx, engineClient, actionName),
				cfg.PollingTimeout, cfg.PollingInterval,
			).Should(Equal(enginegraphql.ActionStatusPhaseSucceeded))

			typeInstances, err := ochClient.ListTypeInstances(ctx, &ochlocalgraphql.TypeInstanceFilter{
				TypeRef: &ochlocalgraphql.TypeRefFilterInput{
					Path:     "cap.type.upload-test",
					Revision: ptr.String("0.1.0"),
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(typeInstances).To(HaveLen(1))

			uploadedTypeInstance := typeInstances[0]
			defer func() {
				err := ochClient.DeleteTypeInstance(ctx, uploadedTypeInstance.ID)
				if err != nil {
					log(errors.Wrapf(err, "while deleting TypeInstance with ID %s", uploadedTypeInstance.ID).Error())
				}
			}()

			Expect(uploadedTypeInstance).To(Equal(getExpectedUploadActionTypeInstance(uploadedTypeInstance.ID)))
		})

		It("should update a TypeInstance", func() {
			input := &ochlocalgraphql.CreateTypeInstanceInput{
				TypeRef: &ochlocalgraphql.TypeInstanceTypeReferenceInput{
					Path:     "cap.type.upload-test",
					Revision: "0.1.0",
				},
				Attributes: []*ochlocalgraphql.AttributeReferenceInput{},
				Value:      map[string]interface{}{"hello": "world"},
			}
			simpleTI, simpleTICleanupFn := createTypeInstance(ctx, ochClient, input)
			defer simpleTICleanupFn()

			_, err := engineClient.CreateAction(ctx, &enginegraphql.ActionDetailsInput{
				Name: actionName,
				ActionRef: &enginegraphql.ManifestReferenceInput{
					Path:     "cap.interface.voltron.e2e.type-instance-update",
					Revision: ptr.String("0.1.0"),
				},
				Input: &enginegraphql.ActionInputData{
					TypeInstances: []*enginegraphql.InputTypeInstanceData{
						{
							Name: "updateTestTypeInstance",
							ID:   simpleTI.ID,
						},
					},
				},
			})

			Expect(err).ToNot(HaveOccurred())

			Eventually(
				getActionStatusFunc(ctx, engineClient, actionName),
				cfg.PollingTimeout, cfg.PollingInterval,
			).Should(Equal(enginegraphql.ActionStatusPhaseReadyToRun))

			err = engineClient.RunAction(ctx, actionName)

			Expect(err).ToNot(HaveOccurred())

			Eventually(
				getActionStatusFunc(ctx, engineClient, actionName),
				cfg.PollingTimeout, cfg.PollingInterval,
			).Should(Equal(enginegraphql.ActionStatusPhaseSucceeded))

			typeInstances, err := ochClient.ListTypeInstances(ctx, &ochlocalgraphql.TypeInstanceFilter{
				TypeRef: &ochlocalgraphql.TypeRefFilterInput{
					Path:     "cap.type.upload-test",
					Revision: ptr.String("0.1.0"),
				},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(typeInstances).To(HaveLen(1))

			uploadedTypeInstance := typeInstances[0]
			defer func() {
				err := ochClient.DeleteTypeInstance(ctx, uploadedTypeInstance.ID)
				if err != nil {
					log(errors.Wrapf(err, "while deleting TypeInstance with ID %s", uploadedTypeInstance.ID).Error())
				}
			}()

			Expect(uploadedTypeInstance).To(Equal(getExpectedUpdateActionTypeInstance(uploadedTypeInstance.ID)))
		})
	})
})

func getExpectedUploadActionTypeInstance(ID string) ochlocalgraphql.TypeInstance {
	revision := getExpectedUploadActionTypeInstanceRevision()

	return ochlocalgraphql.TypeInstance{
		ID: ID,
		TypeRef: &ochlocalgraphql.TypeInstanceTypeReference{
			Path:     "cap.type.upload-test",
			Revision: "0.1.0",
		},
		Uses:                    []*ochlocalgraphql.TypeInstance{},
		UsedBy:                  []*ochlocalgraphql.TypeInstance{},
		LatestResourceVersion:   revision,
		FirstResourceVersion:    revision,
		PreviousResourceVersion: nil,
		ResourceVersion:         revision,
		ResourceVersions:        []*ochlocalgraphql.TypeInstanceResourceVersion{revision},
	}
}

func getExpectedUpdateActionTypeInstance(ID string) ochlocalgraphql.TypeInstance {
	firstRevision := getExpectedUploadActionTypeInstanceRevision()
	secondRevision := getExpectedUpdateActionTypeInstanceRevision()

	return ochlocalgraphql.TypeInstance{
		ID: ID,
		TypeRef: &ochlocalgraphql.TypeInstanceTypeReference{
			Path:     "cap.type.upload-test",
			Revision: "0.1.0",
		},
		Uses:                    []*ochlocalgraphql.TypeInstance{},
		UsedBy:                  []*ochlocalgraphql.TypeInstance{},
		LatestResourceVersion:   secondRevision,
		FirstResourceVersion:    firstRevision,
		PreviousResourceVersion: firstRevision,
		ResourceVersion:         firstRevision,
		ResourceVersions:        []*ochlocalgraphql.TypeInstanceResourceVersion{secondRevision, firstRevision},
	}
}

func getExpectedUploadActionTypeInstanceRevision() *ochlocalgraphql.TypeInstanceResourceVersion {
	return &ochlocalgraphql.TypeInstanceResourceVersion{
		ResourceVersion: 1,
		Metadata: &ochlocalgraphql.TypeInstanceResourceVersionMetadata{
			Attributes: []*ochlocalgraphql.AttributeReference{},
		},
		Spec: &ochlocalgraphql.TypeInstanceResourceVersionSpec{
			Value:           map[string]interface{}{"hello": "world"},
			Instrumentation: nil,
		},
	}
}

func getExpectedUpdateActionTypeInstanceRevision() *ochlocalgraphql.TypeInstanceResourceVersion {
	return &ochlocalgraphql.TypeInstanceResourceVersion{
		ResourceVersion: 2,
		Metadata: &ochlocalgraphql.TypeInstanceResourceVersionMetadata{
			Attributes: []*ochlocalgraphql.AttributeReference{},
		},
		Spec: &ochlocalgraphql.TypeInstanceResourceVersionSpec{
			Value:           map[string]interface{}{"hello": "world2"},
			Instrumentation: nil,
		},
	}
}

func getActionStatusFunc(ctx context.Context, cl *engine.Client, name string) func() (enginegraphql.ActionStatusPhase, error) {
	return func() (enginegraphql.ActionStatusPhase, error) {
		action, err := cl.GetAction(ctx, name)
		if err != nil {
			return "", err
		}
		return action.Status.Phase, err
	}
}

func getTypeInstanceInputForPolicy() *ochlocalgraphql.CreateTypeInstanceInput {
	return &ochlocalgraphql.CreateTypeInstanceInput{
		TypeRef: &ochlocalgraphql.TypeInstanceTypeReferenceInput{
			Path:     "cap.type.simple.single-key",
			Revision: "0.1.0",
		},
		Attributes: []*ochlocalgraphql.AttributeReferenceInput{
			{
				Path:     "com.voltron.attribute",
				Revision: "0.1.0",
			},
		},
		Value: map[string]interface{}{
			"key": true,
		},
	}
}

func createActionAndWaitForReadyToRunPhase(ctx context.Context, engineClient *engine.Client, actionName, actionPath string) *enginegraphql.Action {
	_, err := engineClient.CreateAction(ctx, &enginegraphql.ActionDetailsInput{
		Name: actionName,
		ActionRef: &enginegraphql.ManifestReferenceInput{
			Path: actionPath,
		},
	})
	Expect(err).ToNot(HaveOccurred())

	// Wait for Action Ready to Run
	Eventually(
		getActionStatusFunc(ctx, engineClient, actionName),
		cfg.PollingTimeout, cfg.PollingInterval,
	).Should(Equal(enginegraphql.ActionStatusPhaseReadyToRun))

	action, err := engineClient.GetAction(ctx, actionName)
	Expect(err).ToNot(HaveOccurred())
	Expect(action).ToNot(BeNil())

	return action
}

func assertActionRenderedWorkflowContains(action *enginegraphql.Action, stringToFind string) {
	jsonBytes, err := json.Marshal(action.RenderedAction)
	Expect(err).ToNot(HaveOccurred())
	Expect(
		strings.Contains(string(jsonBytes), stringToFind),
	).To(BeTrue())
}

func runActionAndWaitForSucceeded(ctx context.Context, engineClient *engine.Client, actionName string) {
	err := engineClient.RunAction(ctx, actionName)
	Expect(err).ToNot(HaveOccurred())

	// Wait for Action Succeeded
	Eventually(
		getActionStatusFunc(ctx, engineClient, actionName),
		cfg.PollingTimeout, cfg.PollingInterval,
	).Should(Equal(enginegraphql.ActionStatusPhaseSucceeded))
}

func createTypeInstance(ctx context.Context, ochClient *ochclient.Client, in *ochlocalgraphql.CreateTypeInstanceInput) (*ochlocalgraphql.TypeInstance, func()) {
	createdTypeInstance, err := ochClient.CreateTypeInstance(ctx, in)
	Expect(err).ToNot(HaveOccurred())

	Expect(createdTypeInstance).NotTo(BeNil())
	typeInstanceID := createdTypeInstance.ID

	cleanupFn := func() {
		err := ochClient.DeleteTypeInstance(ctx, typeInstanceID)
		if err != nil {
			log(errors.Wrapf(err, "while deleting TypeInstance with ID %s", typeInstanceID).Error())
		}
	}

	return createdTypeInstance, cleanupFn
}

func updateClusterPolicyConfigMap(stringToFind, stringToReplace string) func() {
	err := replaceInClusterPolicyConfigMap(stringToFind, stringToReplace)
	Expect(err).ToNot(HaveOccurred())

	cleanupFn := func() {
		err := replaceInClusterPolicyConfigMap(stringToReplace, stringToFind)
		if err != nil {
			log(errors.Wrap(err, "while cleaning up ConfigMap with cluster policy").Error())
		}
	}

	return cleanupFn
}

func replaceInClusterPolicyConfigMap(stringToFind, stringToReplace string) error {
	k8sCfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(k8sCfg)
	if err != nil {
		return err
	}

	cfgMapCli := clientset.CoreV1().ConfigMaps(cfg.ClusterPolicy.Namespace)

	clusterPolicyCfgMap, err := cfgMapCli.Get(cfg.ClusterPolicy.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	oldContent := clusterPolicyCfgMap.Data[clusterPolicyConfigMapKey]
	newContent := strings.ReplaceAll(oldContent, stringToFind, stringToReplace)
	clusterPolicyCfgMap.Data[clusterPolicyConfigMapKey] = newContent

	_, err = cfgMapCli.Update(clusterPolicyCfgMap)
	if err != nil {
		return err
	}

	return nil
}
