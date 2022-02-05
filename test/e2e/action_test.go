//go:build integration
// +build integration

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"capact.io/capact/internal/ptr"
	enginegraphql "capact.io/capact/pkg/engine/api/graphql"
	engine "capact.io/capact/pkg/engine/client"
	hublocalgraphql "capact.io/capact/pkg/hub/api/graphql/local"
	hubclient "capact.io/capact/pkg/hub/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"
)

const (
	actionPassingPath = "cap.interface.capactio.capact.validation.action.passing"
)

func getActionName() string {
	return fmt.Sprintf("e2e-test-%d-%s", GinkgoParallelNode(), strconv.Itoa(rand.Intn(10000)))
}

var _ = Describe("Action", func() {
	var engineClient *engine.Client
	var hubClient *hubclient.Client
	var actionName string

	failingActionName := fmt.Sprintf("e2e-failing-test-%d", GinkgoParallelNode())
	ctx := context.Background()

	BeforeEach(func() {
		engineClient = getEngineGraphQLClient()
		hubClient = getHubGraphQLClient()
		actionName = getActionName()

		// Ensure Test Policy
		updateGlobalPolicy(ctx, engineClient, nil)
	})

	AfterEach(func() {
		// cleanup Action
		engineClient.DeleteAction(ctx, actionName)
		engineClient.DeleteAction(ctx, failingActionName)
	})

	Context("Action execution", func() {
		It("should pick proper Implementation and inject TypeInstance based on cluster policy", func() {
			actionPath := "cap.interface.capactio.capact.validation.action.passing"
			testValue := "Implementation A"

			By("1. Preparing input Type Instances")

			By("1.1 Creating TypeInstance which will be downloaded")
			download := getTypeInstanceInputForDownload(testValue)
			downloadTI, downloadTICleanup := createTypeInstance(ctx, hubClient, download)
			defer downloadTICleanup()

			By("1.2 Creating TypeInstance which will be downloaded and updated")
			update := getTypeInstanceInputForUpdate()
			updateTI, updateTICleanup := createTypeInstance(ctx, hubClient, update)
			defer updateTICleanup()

			typeInstances := []*enginegraphql.InputTypeInstanceData{
				{Name: "testUpdate", ID: updateTI.ID},
				{Name: "testInput", ID: downloadTI.ID},
			}

			inputData := &enginegraphql.ActionInputData{
				TypeInstances: typeInstances,
			}

			By("1.3 Create TypeInstance which is required for Implementation B to be picked based on Policy")
			typeInstanceValue := getTypeInstanceInputForPolicy()
			typeInstance, tiCleanupFn := createTypeInstance(ctx, hubClient, typeInstanceValue)
			defer tiCleanupFn()
			injectedTypeInstanceID := typeInstance.ID

			By("2. Expecting Implementation A is picked based on test policy and requirements met...")

			action := createActionAndWaitForReadyToRunPhase(ctx, engineClient, actionName, actionPath, inputData)
			assertActionRenderedWorkflowContains(action, "echo 'Implementation A'")
			runActionAndWaitForSucceeded(ctx, engineClient, actionName)

			By("3. Check TypeInstances")
			By("3.1 Check uploaded TypeInstances")
			assertUploadedTypeInstance(ctx, hubClient, testValue)

			assertOutputTypeInstancesInActionStatus(ctx, engineClient, action.Name, And(ContainElement(
				&enginegraphql.OutputTypeInstanceDetails{
					ID: updateTI.ID,
					TypeRef: &enginegraphql.ManifestReference{
						Path:     updateTI.TypeRef.Path,
						Revision: updateTI.TypeRef.Revision,
					},
				},
			), HaveLen(2)))

			By("3.2 Check updated TypeInstances")
			updateTI, err := hubClient.FindTypeInstance(ctx, updateTI.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(updateTI).ToNot(BeNil())

			_, err = getTypeInstanceWithValue([]hublocalgraphql.TypeInstance{*updateTI}, testValue)
			Expect(err).ToNot(HaveOccurred())

			By("3.3 Deleting Action...")
			err = engineClient.DeleteAction(ctx, actionName)
			Expect(err).ToNot(HaveOccurred())

			By("3.4 Waiting for Action deleted")
			waitForActionDeleted(ctx, engineClient, actionName)

			By("4. Modifying Policy to make Implementation B picked for next run...")
			globalPolicyRequiredTypeInstances := enginegraphql.RequiredTypeInstanceReferenceInput{
				ID:          injectedTypeInstanceID,
				Description: ptr.String("Test TypeInstance"),
			}
			updateGlobalPolicy(ctx, engineClient, &globalPolicyRequiredTypeInstances)

			By("5. Expecting Implementation B is picked based on test policy...")
			action = createActionAndWaitForReadyToRunPhase(ctx, engineClient, actionName, actionPath, inputData)
			assertActionRenderedWorkflowContains(action, "echo 'Implementation B'")
			runActionAndWaitForSucceeded(ctx, engineClient, actionName)

			By("6. Check Uploaded TypeInstances")
			assertUploadedTypeInstance(ctx, hubClient, testValue)

			By("7. Check output TypeInstances in Action status")
			assertOutputTypeInstancesInActionStatus(ctx, engineClient, action.Name, HaveLen(1))
		})

		It("should have failed status after a failed workflow", func() {
			_, err := engineClient.CreateAction(ctx, &enginegraphql.ActionDetailsInput{
				Name: failingActionName,
				ActionRef: &enginegraphql.ManifestReferenceInput{
					Path:     "cap.interface.capactio.capact.validation.action.failing",
					Revision: ptr.String("0.1.0"),
				},
			})

			Expect(err).ToNot(HaveOccurred())

			Eventually(
				getActionStatusFunc(ctx, engineClient, failingActionName),
				cfg.PollingTimeout, cfg.PollingInterval,
			).Should(Equal(enginegraphql.ActionStatusPhaseReadyToRun))

			err = engineClient.RunAction(ctx, failingActionName)

			Expect(err).ToNot(HaveOccurred())

			Eventually(
				getActionStatusFunc(ctx, engineClient, failingActionName),
				cfg.PollingTimeout, cfg.PollingInterval,
			).Should(Equal(enginegraphql.ActionStatusPhaseFailed))
		})

		DescribeTable("Should lock and unlock updated TypeInstances", func(inputParameters map[string]interface{}, expectedStatus enginegraphql.ActionStatusPhase) {
			const actionPath = "cap.interface.capactio.capact.validation.action.update"

			By("Prepare TypeInstance to update")

			update := getTypeInstanceInputForUpdate()
			updateTI, updateTICleanup := createTypeInstance(ctx, hubClient, update)
			defer updateTICleanup()

			typeInstances := []*enginegraphql.InputTypeInstanceData{
				{Name: "testUpdate", ID: updateTI.ID},
			}

			parameters, err := mapToInputParameters(inputParameters)
			Expect(err).ToNot(HaveOccurred())

			inputData := &enginegraphql.ActionInputData{
				TypeInstances: typeInstances,
				Parameters:    parameters,
			}

			By("Create and run Action")

			createActionAndWaitForReadyToRunPhase(ctx, engineClient, actionName, actionPath, inputData)
			defer func() {
				err := engineClient.DeleteAction(ctx, actionName)
				Expect(err).ToNot(HaveOccurred())
			}()

			err = engineClient.RunAction(ctx, actionName)
			Expect(err).ToNot(HaveOccurred())

			By("Verify the TypeInstance is locked")
			Eventually(func() error {
				updateTI, err := hubClient.FindTypeInstance(ctx, updateTI.ID)
				if err != nil {
					return err
				}

				if updateTI.LockedBy == nil {
					return errors.New("TypeInstance is not locked")
				}

				return nil
			}, 30*time.Second).Should(BeNil())

			By("Wait for Action completion")
			runActionAndWaitForStatus(ctx, engineClient, actionName, expectedStatus)

			By("Verify the TypeInstance is unlock after the action passes")
			Eventually(func() error {
				updateTI, err := hubClient.FindTypeInstance(ctx, updateTI.ID)
				if err != nil {
					return err
				}

				if updateTI.LockedBy != nil {
					return errors.New("TypeInstance is locked")
				}

				return nil
			}, cfg.PollingTimeout, cfg.PollingInterval).Should(BeNil())
		},
			Entry("Passing action", map[string]interface{}{
				"input-parameters": map[string]interface{}{
					"testString": "success",
				},
			}, enginegraphql.ActionStatusPhaseSucceeded),
			Entry("Failing action", map[string]interface{}{
				"input-parameters": map[string]interface{}{
					"testString": "failure",
				},
			}, enginegraphql.ActionStatusPhaseFailed),
		)
	})
})

func getActionStatusFunc(ctx context.Context, cl *engine.Client, name string) func() (enginegraphql.ActionStatusPhase, error) {
	return func() (enginegraphql.ActionStatusPhase, error) {
		action, err := cl.GetAction(ctx, name)
		if err != nil {
			return "", err
		}
		if action == nil || action.Status == nil {
			return "", errors.New("Action and its status cannot be nil")
		}

		return action.Status.Phase, err
	}
}

func getActionFunc(ctx context.Context, cl *engine.Client, name string) func() (*enginegraphql.Action, error) {
	return func() (*enginegraphql.Action, error) {
		action, err := cl.GetAction(ctx, name)
		if err != nil {
			return nil, err
		}
		return action, err
	}
}

func getTypeInstanceInputForPolicy() *hublocalgraphql.CreateTypeInstanceInput {
	return &hublocalgraphql.CreateTypeInstanceInput{
		TypeRef: &hublocalgraphql.TypeInstanceTypeReferenceInput{
			Path:     "cap.type.capactio.capact.validation.single-key",
			Revision: "0.1.0",
		},
		Attributes: []*hublocalgraphql.AttributeReferenceInput{
			{
				Path:     "cap.attribute.capactio.capact.attribute",
				Revision: "0.1.0",
			},
		},
		Value: map[string]interface{}{
			"key": true,
		},
	}
}

func getTypeInstanceInputForDownload(testValue string) *hublocalgraphql.CreateTypeInstanceInput {
	return &hublocalgraphql.CreateTypeInstanceInput{
		TypeRef: &hublocalgraphql.TypeInstanceTypeReferenceInput{
			Path:     "cap.type.capactio.capact.validation.download",
			Revision: "0.1.0",
		},
		Value: map[string]interface{}{"key": testValue},
		Attributes: []*hublocalgraphql.AttributeReferenceInput{
			{
				Path:     "cap.attribute.capactio.capact.attribute1",
				Revision: "0.1.0",
			},
		},
	}
}

func getTypeInstanceInputForUpdate() *hublocalgraphql.CreateTypeInstanceInput {
	return &hublocalgraphql.CreateTypeInstanceInput{
		TypeRef: &hublocalgraphql.TypeInstanceTypeReferenceInput{
			Path:     "cap.type.capactio.capact.validation.update",
			Revision: "0.1.0",
		},
		Value: map[string]interface{}{"key": "random text to update"},
		Attributes: []*hublocalgraphql.AttributeReferenceInput{
			{
				Path:     "cap.attribute.capactio.capact.attribute1",
				Revision: "0.1.0",
			},
		},
	}
}

func createActionAndWaitForReadyToRunPhase(ctx context.Context, engineClient *engine.Client, actionName, actionPath string, input *enginegraphql.ActionInputData) *enginegraphql.Action {
	_, err := engineClient.CreateAction(ctx, &enginegraphql.ActionDetailsInput{
		Name: actionName,
		ActionRef: &enginegraphql.ManifestReferenceInput{
			Path: actionPath,
		},
		Input: input,
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
	runActionAndWaitForStatus(ctx, engineClient, actionName,
		enginegraphql.ActionStatusPhaseSucceeded)
}

func waitForActionDeleted(ctx context.Context, engineClient *engine.Client, actionName string) {
	Eventually(
		getActionFunc(ctx, engineClient, actionName),
		cfg.PollingTimeout, cfg.PollingInterval,
	).Should(BeNil())
}

func runActionAndWaitForStatus(ctx context.Context, engineClient *engine.Client, actionName string, statuses ...enginegraphql.ActionStatusPhase) {
	err := engineClient.RunAction(ctx, actionName)
	Expect(err).ToNot(HaveOccurred())

	Eventually(
		getActionStatusFunc(ctx, engineClient, actionName),
		cfg.PollingTimeout, cfg.PollingInterval,
	).Should(BeElementOf(statuses))
}

func createTypeInstance(ctx context.Context, hubClient *hubclient.Client, in *hublocalgraphql.CreateTypeInstanceInput) (*hublocalgraphql.TypeInstance, func()) {
	createdTypeInstance, err := hubClient.CreateTypeInstance(ctx, in)
	Expect(err).ToNot(HaveOccurred())

	Expect(createdTypeInstance).NotTo(BeNil())
	typeInstanceID := createdTypeInstance.ID

	cleanupFn := func() {
		err := hubClient.DeleteTypeInstance(ctx, typeInstanceID)
		if err != nil {
			log(errors.Wrapf(err, "while deleting TypeInstance with ID %s", typeInstanceID).Error())
		}
	}

	return createdTypeInstance, cleanupFn
}

func updateGlobalPolicy(ctx context.Context, client *engine.Client, reqTypeInstance *enginegraphql.RequiredTypeInstanceReferenceInput) {
	var reqInput []*enginegraphql.RequiredTypeInstanceReferenceInput
	// nils element are not handled by GraphQL
	if reqTypeInstance != nil {
		reqInput = append(reqInput, reqTypeInstance)
	}
	p := PolicyInputTestFixture(reqInput)
	_, err := client.UpdatePolicy(ctx, p)
	Expect(err).ToNot(HaveOccurred())
}

func assertUploadedTypeInstance(ctx context.Context, hubClient *hubclient.Client, testValue string) {
	uploaded, err := hubClient.ListTypeInstances(ctx, &hublocalgraphql.TypeInstanceFilter{
		TypeRef: &hublocalgraphql.TypeRefFilterInput{
			Path:     "cap.type.capactio.capact.validation.upload",
			Revision: ptr.String("0.1.0"),
		},
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(len(uploaded)).Should(BeNumerically(">", 0))

	ti, err := getTypeInstanceWithValue(uploaded, testValue)
	Expect(err).ToNot(HaveOccurred())

	err = hubClient.DeleteTypeInstance(ctx, ti.ID)
	Expect(err).ToNot(HaveOccurred())
}

func assertOutputTypeInstancesInActionStatus(ctx context.Context, engineClient *engine.Client, actionName string,
	match types.GomegaMatcher,
) {
	Eventually(func() ([]*enginegraphql.OutputTypeInstanceDetails, error) {
		action, err := engineClient.GetAction(ctx, actionName)
		if err != nil {
			return nil, err
		}

		if action.Output == nil {
			return nil, errors.New(".output.typeInstances not populated")
		}

		return action.Output.TypeInstances, nil
	}, 10*time.Second).Should(match)
}

func getTypeInstanceWithValue(typeInstances []hublocalgraphql.TypeInstance, testValue string) (*hublocalgraphql.TypeInstance, error) {
	for _, ti := range typeInstances {
		values, ok := ti.LatestResourceVersion.Spec.Value.(map[string]interface{})
		if !ok {
			continue
		}
		value, ok := values["key"].(string)
		if !ok {
			continue
		}
		if value == testValue {
			return &ti, nil
		}
	}
	return nil, fmt.Errorf("No TypeInstance with value %s", testValue)
}

func mapToInputParameters(params map[string]interface{}) (*enginegraphql.JSON, error) {
	marshaled, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	res := enginegraphql.JSON(marshaled)
	return &res, nil
}

func PolicyInputTestFixture(reqInput []*enginegraphql.RequiredTypeInstanceReferenceInput) *enginegraphql.PolicyInput {
	manifestRef := func(path string) []*enginegraphql.ManifestReferenceInput {
		return []*enginegraphql.ManifestReferenceInput{
			{
				Path: path,
			},
		}
	}

	return &enginegraphql.PolicyInput{
		Interface: &enginegraphql.InterfacePolicyInput{
			Rules: []*enginegraphql.RulesForInterfaceInput{
				{
					Interface: manifestRef(actionPassingPath)[0],
					OneOf: []*enginegraphql.PolicyRuleInput{
						{
							ImplementationConstraints: &enginegraphql.PolicyRuleImplementationConstraintsInput{
								Requires:   manifestRef("cap.type.capactio.capact.validation.single-key"),
								Attributes: manifestRef("cap.attribute.capactio.capact.validation.policy.most-preferred"),
							},
							Inject: &enginegraphql.PolicyRuleInjectDataInput{
								RequiredTypeInstances: reqInput,
							},
						},
						{
							ImplementationConstraints: &enginegraphql.PolicyRuleImplementationConstraintsInput{
								Path: ptr.String("cap.implementation.capactio.capact.validation.action.passing-a"),
							},
						},
					},
				},
				// allow all others
				{
					Interface: manifestRef("cap.*")[0],
					OneOf: []*enginegraphql.PolicyRuleInput{
						{
							ImplementationConstraints: &enginegraphql.PolicyRuleImplementationConstraintsInput{},
						},
					},
				},
			},
		},
	}
}
