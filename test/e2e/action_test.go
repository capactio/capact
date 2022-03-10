//go:build integration
// +build integration

package e2e

import (
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/ptr"
	enginegraphql "capact.io/capact/pkg/engine/api/graphql"
	engine "capact.io/capact/pkg/engine/client"
	hublocalgraphql "capact.io/capact/pkg/hub/api/graphql/local"
	hubclient "capact.io/capact/pkg/hub/client"
	"capact.io/capact/pkg/hub/client/local"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	"github.com/pkg/errors"
)

const (
	actionPassingInterfacePath = "cap.interface.capactio.capact.validation.action.passing"
	uploadTypePath             = "cap.type.capactio.capact.validation.upload"
	singleKeyTypePath          = "cap.type.capactio.capact.validation.single-key"
	testStorageBackendPath     = "cap.type.capactio.capact.validation.storage"
)

func getActionName() string {
	rand.Seed(time.Now().UTC().UnixNano())
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

		// Ensure default Test Policy
		setGlobalTestPolicy(ctx, engineClient)
	})

	AfterEach(func() {
		// cleanup Action
		engineClient.DeleteAction(ctx, actionName)
		engineClient.DeleteAction(ctx, failingActionName)
	})

	Context("Action execution", func() {

		It("should pick Implementation A", func() {
			implIndicatorValue := "Implementation A"

			// TODO: This can be extracted after switching to ginkgo v2
			// see: https://github.com/onsi/ginkgo/issues/70#issuecomment-924250145
			By("1. Preparing input Type Instances")

			By("1.1 Creating TypeInstance which will be downloaded")
			download := getTypeInstanceInputForDownload(map[string]interface{}{"key": implIndicatorValue})
			downloadTI, downloadTICleanup := createTypeInstance(ctx, hubClient, download)
			defer downloadTICleanup()

			By("1.2 Creating TypeInstance which will be downloaded and updated")
			update := getTypeInstanceInputForUpdate()
			updateTI, updateTICleanup := createTypeInstance(ctx, hubClient, update)
			defer updateTICleanup()

			By("1.3 Creating TypeInstance that describes Helm storage")
			helmStorage := fixHelmStorageTypeInstanceCreateInput()
			helmStorageTI, helmStorageTICleanup := createTypeInstance(ctx, hubClient, helmStorage)
			defer helmStorageTICleanup()

			inputData := &enginegraphql.ActionInputData{
				TypeInstances: []*enginegraphql.InputTypeInstanceData{
					{Name: "testInput", ID: downloadTI.ID},
					{Name: "testUpdate", ID: updateTI.ID},
				},
			}

			builtinStorage := getBuiltinStorageTypeInstance(ctx, hubClient)
			expUpdatedTIOutput := mapToOutputTypeInstanceDetails(updateTI, builtinStorage.Backend)

			By("2. Expecting Implementation A is picked and builtin storage is used...")

			action := createActionAndWaitForReadyToRunPhase(ctx, engineClient, actionName, actionPassingInterfacePath, inputData)
			assertActionRenderedWorkflowContains(action, "echo '%s'", implIndicatorValue)
			runActionAndWaitForSucceeded(ctx, engineClient, actionName)

			By("3.1 Check uploaded TypeInstances")
			expUploadTIBackend := &hublocalgraphql.TypeInstanceBackendReference{ID: builtinStorage.ID, Abstract: true}
			uploadedTI, cleanupUploaded := getUploadedTypeInstanceByValue(ctx, hubClient, implIndicatorValue)
			Expect(uploadedTI.Backend).Should(Equal(expUploadTIBackend))

			By("3.2 Check updated TypeInstances")
			getTypeInstanceByIDAndValue(ctx, hubClient, updateTI.ID, implIndicatorValue)

			By("3.3 Check Action output TypeInstances")
			uploadedTIOutput := mapToOutputTypeInstanceDetails(uploadedTI, expUploadTIBackend)
			assertOutputTypeInstancesInActionStatus(ctx, engineClient, action.Name, And(ContainElements(expUpdatedTIOutput, uploadedTIOutput), HaveLen(2)))

			By("4. Deleting Action...")
			err := engineClient.DeleteAction(ctx, actionName)
			cleanupUploaded() // We need to clean it up as it's not deleted when Action is deleted.
			Expect(err).ToNot(HaveOccurred())

			By("5. Waiting for Action deleted")
			waitForActionDeleted(ctx, engineClient, actionName)

			By("6. Modifying Policy to change backend storage for uploaded TypeInstance via TypeRef...")
			setGlobalTestPolicy(ctx, engineClient, withHelmBackendForUploadTypeRef(helmStorageTI.ID))

			By("7. Expecting Implementation A is picked and the Helm storage is used for uploaded TypeInstance...")
			action = createActionAndWaitForReadyToRunPhase(ctx, engineClient, actionName, actionPassingInterfacePath, inputData)
			assertActionRenderedWorkflowContains(action, "echo '%s'", implIndicatorValue)
			runActionAndWaitForSucceeded(ctx, engineClient, actionName)

			By("8.1 Check uploaded TypeInstances")
			expUploadTIBackend = &hublocalgraphql.TypeInstanceBackendReference{ID: helmStorageTI.ID, Abstract: false}
			uploadedTI, cleanupUploaded = getUploadedTypeInstanceByValue(ctx, hubClient, implIndicatorValue)
			defer cleanupUploaded() // We need to clean it up as it's not deleted when Action is deleted.
			Expect(uploadedTI.Backend).Should(Equal(expUploadTIBackend))

			By("8.2 Check Action output TypeInstances")
			uploadedTIOutput = mapToOutputTypeInstanceDetails(uploadedTI, expUploadTIBackend)
			assertOutputTypeInstancesInActionStatus(ctx, engineClient, action.Name, And(ContainElements(expUpdatedTIOutput, uploadedTIOutput), HaveLen(2)))
		})

		It("should pick Implementation B based on Policy rule", func() {
			implIndicatorValue := "Implementation B"

			// TODO: This can be extracted after switching to ginkgo v2
			// see: https://github.com/onsi/ginkgo/issues/70#issuecomment-924250145
			By("1. Preparing input Type Instances")
			By("1.1 Creating TypeInstance which will be downloaded")
			download := getTypeInstanceInputForDownload(map[string]interface{}{"key": implIndicatorValue})
			downloadTI, downloadTICleanup := createTypeInstance(ctx, hubClient, download)
			defer downloadTICleanup()

			By("1.2 Creating TypeInstance which will be downloaded and updated")
			update := getTypeInstanceInputForUpdate()
			updateTI, updateTICleanup := createTypeInstance(ctx, hubClient, update)
			defer updateTICleanup()

			By("1.3 Creating TypeInstance that describes Helm storage")
			helmStorage := fixHelmStorageTypeInstanceCreateInput()
			helmStorageTI, helmStorageTICleanup := createTypeInstance(ctx, hubClient, helmStorage)
			defer helmStorageTICleanup()

			By("1.4 Create TypeInstance which is required for Implementation B to be picked based on Policy")
			typeInstanceValue := getTypeInstanceInputForPolicy()
			injectTypeInstance, tiCleanupFn := createTypeInstance(ctx, hubClient, typeInstanceValue)
			defer tiCleanupFn()

			inputData := &enginegraphql.ActionInputData{
				TypeInstances: []*enginegraphql.InputTypeInstanceData{
					{Name: "testInput", ID: downloadTI.ID},
					{Name: "testUpdate", ID: updateTI.ID},
				},
			}

			By("2. Modifying rule Policy to pick Implementation B...")
			globalPolicyRequiredTypeInstances := []*enginegraphql.RequiredTypeInstanceReferenceInput{
				{
					ID:          injectTypeInstance.ID,
					Description: ptr.String("Test TypeInstance"),
				},
				{
					ID:          helmStorageTI.ID,
					Description: ptr.String("Helm backend TypeInstance"),
				},
			}
			setGlobalTestPolicy(ctx, engineClient, prependInjectRuleForPassingActionInterface(globalPolicyRequiredTypeInstances))

			By("3. Expecting Implementation B is picked and injected Helm storage is used...")
			action := createActionAndWaitForReadyToRunPhase(ctx, engineClient, actionName, actionPassingInterfacePath, inputData)
			assertActionRenderedWorkflowContains(action, "echo '%s'", implIndicatorValue)
			runActionAndWaitForSucceeded(ctx, engineClient, actionName)

			By("4.1 Check uploaded TypeInstances")
			expUploadTIBackend := &hublocalgraphql.TypeInstanceBackendReference{ID: helmStorageTI.ID, Abstract: false}
			uploadedTI, cleanupUploaded := getUploadedTypeInstanceByValue(ctx, hubClient, implIndicatorValue)
			defer cleanupUploaded() // We need to clean it up as it's not deleted when Action is deleted.
			Expect(uploadedTI.Backend).Should(Equal(expUploadTIBackend))

			By("4.2 Check Action output TypeInstances")
			uploadedTIOutput := mapToOutputTypeInstanceDetails(uploadedTI, expUploadTIBackend)
			assertOutputTypeInstancesInActionStatus(ctx, engineClient, action.Name, And(ContainElements(uploadedTIOutput), HaveLen(1)))
		})

		It("should pick Implementation B based on Interface default", func() {
			implIndicatorValue := "Implementation B"

			// TODO: This can be extracted after switching to ginkgo v2
			// see: https://github.com/onsi/ginkgo/issues/70#issuecomment-924250145
			By("1. Preparing input Type Instances")
			By("1.1 Creating TypeInstance which will be downloaded")
			download := getTypeInstanceInputForDownload(map[string]interface{}{"key": implIndicatorValue})
			downloadTI, downloadTICleanup := createTypeInstance(ctx, hubClient, download)
			defer downloadTICleanup()

			By("1.2 Creating TypeInstance which will be downloaded and updated")
			update := getTypeInstanceInputForUpdate()
			updateTI, updateTICleanup := createTypeInstance(ctx, hubClient, update)
			defer updateTICleanup()

			By("1.3 Creating TypeInstance that describes Helm storage")
			helmStorage := fixHelmStorageTypeInstanceCreateInput()
			helmStorageTI, helmStorageTICleanup := createTypeInstance(ctx, hubClient, helmStorage)
			defer helmStorageTICleanup()

			By("1.4 Create TypeInstance which is required for Implementation B to be picked based on Policy")
			typeInstanceValue := getTypeInstanceInputForPolicy()
			injectTypeInstance, tiCleanupFn := createTypeInstance(ctx, hubClient, typeInstanceValue)
			defer tiCleanupFn()

			inputData := &enginegraphql.ActionInputData{
				TypeInstances: []*enginegraphql.InputTypeInstanceData{
					{Name: "testInput", ID: downloadTI.ID},
					{Name: "testUpdate", ID: updateTI.ID},
				},
			}

			By("2. Modifying default Policy to pick Implementation B...")
			globalPolicyRequiredTypeInstances := []*enginegraphql.RequiredTypeInstanceReferenceInput{
				{
					ID:          injectTypeInstance.ID,
					Description: ptr.String("Test TypeInstance"),
				},
				{
					ID:          helmStorageTI.ID,
					Description: ptr.String("Helm backend TypeInstance"),
				},
			}
			setGlobalTestPolicy(ctx, engineClient, addInterfacePolicyDefaultInjectionForPassingActionInterface(globalPolicyRequiredTypeInstances))

			By("3. Expecting Implementation B is picked and injected Helm storage is used...")
			action := createActionAndWaitForReadyToRunPhase(ctx, engineClient, actionName, actionPassingInterfacePath, inputData)
			assertActionRenderedWorkflowContains(action, "echo '%s'", implIndicatorValue)
			runActionAndWaitForSucceeded(ctx, engineClient, actionName)

			By("4.1 Check uploaded TypeInstances")
			expUploadTIBackend := &hublocalgraphql.TypeInstanceBackendReference{ID: helmStorageTI.ID, Abstract: false}
			uploadedTI, cleanupUploaded := getUploadedTypeInstanceByValue(ctx, hubClient, implIndicatorValue)
			defer cleanupUploaded() // We need to clean it up as it's not deleted when Action is deleted.
			Expect(uploadedTI.Backend).Should(Equal(expUploadTIBackend))

			By("4.2 Check Action output TypeInstances")
			uploadedTIOutput := mapToOutputTypeInstanceDetails(uploadedTI, expUploadTIBackend)
			assertOutputTypeInstancesInActionStatus(ctx, engineClient, action.Name, And(ContainElements(uploadedTIOutput), HaveLen(1)))
		})

		It("should propagate context provider to storage backend", func() {
			implIndicatorValue := "Implementation C"
			testStorageBackendTI := getDefaultTestStorageTypeInstance(ctx, hubClient)

			// TODO: This can be extracted after switching to ginkgo v2
			// see: https://github.com/onsi/ginkgo/issues/70#issuecomment-924250145
			By("1. Preparing input Type Instances")
			By("1.1 Creating TypeInstance which will be downloaded")
			download := getTypeInstanceInputForDownload(map[string]interface{}{
				"value": map[string]interface{}{
					"key": implIndicatorValue,
				},
				"backend": map[string]interface{}{
					"context": map[string]interface{}{
						"provider": "dotenv",
					},
				},
			})
			downloadTI, downloadTICleanup := createTypeInstance(ctx, hubClient, download)
			defer downloadTICleanup()

			By("1.2 Creating TypeInstance which will be downloaded and updated")
			update := getTypeInstanceInputForUpdate()
			updateTI, updateTICleanup := createTypeInstance(ctx, hubClient, update)
			defer updateTICleanup()

			By("1.3 Create TypeInstance which is required for Implementation C to be picked based on Policy")
			typeInstanceValue := getTypeInstanceInputForPolicy()
			injectTypeInstance, tiCleanupFn := createTypeInstance(ctx, hubClient, typeInstanceValue)
			defer tiCleanupFn()

			inputData := &enginegraphql.ActionInputData{
				TypeInstances: []*enginegraphql.InputTypeInstanceData{
					{Name: "testInput", ID: downloadTI.ID},
					{Name: "testUpdate", ID: updateTI.ID},
				},
			}

			By("2. Modifying default Policy to pick Implementation C...")
			globalPolicyRequiredTypeInstances := []*enginegraphql.RequiredTypeInstanceReferenceInput{
				{
					ID:          injectTypeInstance.ID,
					Description: ptr.String("Test TypeInstance"),
				},
				{
					ID:          testStorageBackendTI.ID,
					Description: ptr.String("Validation storage backend TypeInstance"),
				},
			}
			setGlobalTestPolicy(ctx, engineClient, addInterfacePolicyDefaultInjectionForPassingActionInterface(globalPolicyRequiredTypeInstances))

			By("3. Expecting Implementation C is picked and injected Validation storage is used...")
			action := createActionAndWaitForReadyToRunPhase(ctx, engineClient, actionName, actionPassingInterfacePath, inputData)
			assertActionRenderedWorkflowContains(action, "echo '%s'", implIndicatorValue)
			runActionAndWaitForSucceeded(ctx, engineClient, actionName)

			By("4.1 Check uploaded TypeInstances")
			expUploadTIBackend := &hublocalgraphql.TypeInstanceBackendReference{ID: testStorageBackendTI.ID, Abstract: false}
			fmt.Println("expUploadTIBackend", expUploadTIBackend.ID)
			uploadedTI, cleanupUploaded := getUploadedTypeInstanceByValue(ctx, hubClient, implIndicatorValue)
			defer cleanupUploaded()
			Expect(uploadedTI.Backend).Should(Equal(expUploadTIBackend))

			By("4.2 Check Action output TypeInstances")
			uploadedTIOutput := mapToOutputTypeInstanceDetails(uploadedTI, expUploadTIBackend)
			assertOutputTypeInstancesInActionStatus(ctx, engineClient, action.Name, And(ContainElements(uploadedTIOutput), HaveLen(1)))
		})

		It("should fail due to incorrect storage provider", func() {
			implIndicatorValue := "Implementation C"
			testStorageBackendTI := getDefaultTestStorageTypeInstance(ctx, hubClient)

			// TODO: This can be extracted after switching to ginkgo v2
			// see: https://github.com/onsi/ginkgo/issues/70#issuecomment-924250145
			By("1. Preparing input Type Instances")
			By("1.1 Creating TypeInstance which will be downloaded")
			download := getTypeInstanceInputForDownload(map[string]interface{}{
				"value": map[string]interface{}{
					"key": implIndicatorValue,
				},
				"backend": map[string]interface{}{
					"context": map[string]interface{}{
						"provider": "incorrect",
					},
				},
			})
			downloadTI, downloadTICleanup := createTypeInstance(ctx, hubClient, download)
			defer downloadTICleanup()

			By("1.2 Creating TypeInstance which will be downloaded and updated")
			update := getTypeInstanceInputForUpdate()
			updateTI, updateTICleanup := createTypeInstance(ctx, hubClient, update)
			defer updateTICleanup()

			By("1.3 Create TypeInstance which is required for Implementation C to be picked based on Policy")
			typeInstanceValue := getTypeInstanceInputForPolicy()
			injectTypeInstance, tiCleanupFn := createTypeInstance(ctx, hubClient, typeInstanceValue)
			defer tiCleanupFn()

			inputData := &enginegraphql.ActionInputData{
				TypeInstances: []*enginegraphql.InputTypeInstanceData{
					{Name: "testInput", ID: downloadTI.ID},
					{Name: "testUpdate", ID: updateTI.ID},
				},
			}

			By("2. Modifying default Policy to pick Implementation C...")
			globalPolicyRequiredTypeInstances := []*enginegraphql.RequiredTypeInstanceReferenceInput{
				{
					ID:          injectTypeInstance.ID,
					Description: ptr.String("Test TypeInstance"),
				},
				{
					ID:          testStorageBackendTI.ID,
					Description: ptr.String("Validation storage backend TypeInstance"),
				},
			}
			setGlobalTestPolicy(ctx, engineClient, addInterfacePolicyDefaultInjectionForPassingActionInterface(globalPolicyRequiredTypeInstances))

			By("3. Create action and wait for a status phase failed")
			createActionAndWaitForReadyToRunPhase(ctx, engineClient, actionName, actionPassingInterfacePath, inputData)
			runActionAndWaitForStatus(ctx, engineClient, actionName, enginegraphql.ActionStatusPhaseFailed)
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

func mapToOutputTypeInstanceDetails(ti *hublocalgraphql.TypeInstance, backend *hublocalgraphql.TypeInstanceBackendReference) *enginegraphql.OutputTypeInstanceDetails {
	return &enginegraphql.OutputTypeInstanceDetails{
		ID: ti.ID,
		TypeRef: &enginegraphql.ManifestReference{
			Path:     ti.TypeRef.Path,
			Revision: ti.TypeRef.Revision,
		},
		Backend: &enginegraphql.TypeInstanceBackendDetails{
			ID:       backend.ID,
			Abstract: backend.Abstract,
		},
	}
}

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
			Path:     singleKeyTypePath,
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

func getTypeInstanceInputForDownload(testValues map[string]interface{}) *hublocalgraphql.CreateTypeInstanceInput {
	return &hublocalgraphql.CreateTypeInstanceInput{
		TypeRef: &hublocalgraphql.TypeInstanceTypeReferenceInput{
			Path:     "cap.type.capactio.capact.validation.download",
			Revision: "0.1.0",
		},
		Value: testValues,
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

func fixHelmStorageTypeInstanceCreateInput() *hublocalgraphql.CreateTypeInstanceInput {
	return &hublocalgraphql.CreateTypeInstanceInput{
		TypeRef: &hublocalgraphql.TypeInstanceTypeReferenceInput{
			Path:     "cap.type.helm.storage",
			Revision: "0.1.0",
		},
		Attributes: []*hublocalgraphql.AttributeReferenceInput{},
		Value: map[string]interface{}{
			"url":         "e2e-test-backend-mock-url:50051",
			"acceptValue": true,
			"contextSchema": heredoc.Doc(`
				{
					"$id": "#/properties/contextSchema",
					"type": "object",
					"required": [
						"name",
						"namespace"
					],
					"properties": {
						"name": {
							"$id": "#/properties/contextSchema/properties/name",
							"type": "string"
						},
						"namespace": {
							"$id": "#/properties/contextSchema/properties/namespace",
							"type": "string"
						}
					},
					"additionalProperties": false
				}`),
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

func assertActionRenderedWorkflowContains(action *enginegraphql.Action, toFindFormat string, toFindArgs ...interface{}) {
	jsonBytes, err := json.Marshal(action.RenderedAction)
	Expect(err).ToNot(HaveOccurred())
	Expect(
		strings.Contains(string(jsonBytes), fmt.Sprintf(toFindFormat, toFindArgs...)),
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
	typeInstanceID, err := hubClient.CreateTypeInstance(ctx, in)
	Expect(err).ToNot(HaveOccurred())

	Expect(typeInstanceID).NotTo(BeEmpty())

	typeInstance, err := hubClient.FindTypeInstance(ctx, typeInstanceID)
	Expect(err).ToNot(HaveOccurred())

	cleanupFn := func() {
		err := hubClient.DeleteTypeInstance(ctx, typeInstanceID)
		if err != nil {
			log(errors.Wrapf(err, "while deleting TypeInstance with ID %s", typeInstanceID).Error())
		}
	}

	return typeInstance, cleanupFn
}

type policyOption func(*enginegraphql.PolicyInput)

func withHelmBackendForUploadTypeRef(backendID string) policyOption {
	return func(policy *enginegraphql.PolicyInput) {
		policy.TypeInstance = &enginegraphql.TypeInstancePolicyInput{
			Rules: []*enginegraphql.RulesForTypeInstanceInput{
				{
					TypeRef: &enginegraphql.ManifestReferenceInput{
						Path:     uploadTypePath,
						Revision: ptr.String("0.1.0"),
					},
					Backend: &enginegraphql.TypeInstanceBackendRuleInput{
						ID:          backendID,
						Description: ptr.String("Default Hub backend storage via TypeRef"),
					},
				},
			},
		}
	}
}

func prependInjectRuleForPassingActionInterface(reqInput []*enginegraphql.RequiredTypeInstanceReferenceInput) policyOption {
	manifestRef := func(path string) []*enginegraphql.ManifestReferenceInput {
		return []*enginegraphql.ManifestReferenceInput{
			{
				Path: path,
			},
		}
	}
	return func(policy *enginegraphql.PolicyInput) {
		for idx, rule := range policy.Interface.Rules {
			if rule.Interface.Path != actionPassingInterfacePath {
				continue
			}
			policy.Interface.Rules[idx].OneOf = append([]*enginegraphql.PolicyRuleInput{
				{
					ImplementationConstraints: &enginegraphql.PolicyRuleImplementationConstraintsInput{
						Requires:   manifestRef(singleKeyTypePath),
						Attributes: manifestRef("cap.attribute.capactio.capact.validation.policy.most-preferred"),
					},
					Inject: &enginegraphql.PolicyRuleInjectDataInput{
						RequiredTypeInstances: reqInput,
					},
				},
			}, policy.Interface.Rules[idx].OneOf...)
		}
	}
}

func addInterfacePolicyDefaultInjectionForPassingActionInterface(reqInput []*enginegraphql.RequiredTypeInstanceReferenceInput) policyOption {
	manifestRef := func(path string) []*enginegraphql.ManifestReferenceInput {
		return []*enginegraphql.ManifestReferenceInput{
			{
				Path: path,
			},
		}
	}
	return func(policy *enginegraphql.PolicyInput) {
		for idx, rule := range policy.Interface.Rules {
			if rule.Interface.Path != actionPassingInterfacePath {
				continue
			}
			policy.Interface.Rules[idx].OneOf = append([]*enginegraphql.PolicyRuleInput{
				{
					ImplementationConstraints: &enginegraphql.PolicyRuleImplementationConstraintsInput{
						Requires:   manifestRef(singleKeyTypePath),
						Attributes: manifestRef("cap.attribute.capactio.capact.validation.policy.most-preferred"),
					},
				},
			}, policy.Interface.Rules[idx].OneOf...)
		}
		policy.Interface.Default = &enginegraphql.DefaultForInterfaceInput{
			Inject: &enginegraphql.DefaultInjectForInterfaceInput{
				RequiredTypeInstances: reqInput,
			},
		}
	}
}

func setGlobalTestPolicy(ctx context.Context, client *engine.Client, opts ...policyOption) {
	p := fixGQLTestPolicyInput()

	for _, opt := range opts {
		opt(p)
	}

	_, err := client.UpdatePolicy(ctx, p)
	Expect(err).ToNot(HaveOccurred())
}

func getTypeInstanceByIDAndValue(ctx context.Context, hubClient *hubclient.Client, id, expValue string) *hublocalgraphql.TypeInstance {
	updateTI, err := hubClient.FindTypeInstance(ctx, id)
	Expect(err).ToNot(HaveOccurred())
	Expect(updateTI).ToNot(BeNil())
	_, err = getTypeInstanceWithValue([]hublocalgraphql.TypeInstance{*updateTI}, expValue)
	Expect(err).ToNot(HaveOccurred())

	return updateTI
}

func getDefaultTestStorageTypeInstance(ctx context.Context, hubClient *hubclient.Client) *hublocalgraphql.TypeInstance {
	storage, err := hubClient.ListTypeInstances(ctx, &hublocalgraphql.TypeInstanceFilter{
		TypeRef: &hublocalgraphql.TypeRefFilterInput{
			Path:     testStorageBackendPath,
			Revision: ptr.String("0.1.0"),
		},
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(len(storage)).Should(Equal(1))
	return &storage[0]
}

func getUploadedTypeInstanceByValue(ctx context.Context, hubClient *hubclient.Client, expValue string) (*hublocalgraphql.TypeInstance, func()) {
	uploaded, err := hubClient.ListTypeInstances(ctx, &hublocalgraphql.TypeInstanceFilter{
		TypeRef: &hublocalgraphql.TypeRefFilterInput{
			Path:     uploadTypePath,
			Revision: ptr.String("0.1.0"),
		},
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(len(uploaded)).Should(BeNumerically(">", 0))

	ti, err := getTypeInstanceWithValue(uploaded, expValue)
	Expect(err).ToNot(HaveOccurred())

	return ti, func() {
		err = hubClient.DeleteTypeInstance(ctx, ti.ID)
		Expect(err).ToNot(HaveOccurred())
	}
}

func getBuiltinStorageTypeInstance(ctx context.Context, hubClient *hubclient.Client) hublocalgraphql.TypeInstance {
	coreStorage, err := hubClient.ListTypeInstances(ctx, &hublocalgraphql.TypeInstanceFilter{
		TypeRef: &hublocalgraphql.TypeRefFilterInput{
			Path:     types.BuiltinHubStorageTypePath,
			Revision: ptr.String("0.1.0"),
		},
	}, local.WithFields(local.TypeInstanceAllFields))
	Expect(err).ToNot(HaveOccurred())
	Expect(coreStorage).Should(HaveLen(1))

	return coreStorage[0]
}

func assertOutputTypeInstancesInActionStatus(ctx context.Context, engineClient *engine.Client, actionName string,
	match gomegatypes.GomegaMatcher,
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

func fixGQLTestPolicyInput() *enginegraphql.PolicyInput {
	manifestRef := func(path string) *enginegraphql.ManifestReferenceInput {
		return &enginegraphql.ManifestReferenceInput{
			Path: path,
		}
	}

	return &enginegraphql.PolicyInput{
		Interface: &enginegraphql.InterfacePolicyInput{
			Rules: []*enginegraphql.RulesForInterfaceInput{
				{
					Interface: manifestRef(actionPassingInterfacePath),
					OneOf: []*enginegraphql.PolicyRuleInput{
						{
							ImplementationConstraints: &enginegraphql.PolicyRuleImplementationConstraintsInput{
								Path: ptr.String("cap.implementation.capactio.capact.validation.action.passing-a"),
							},
						},
					},
				},
				// allow all others
				{
					Interface: manifestRef("cap.*"),
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
