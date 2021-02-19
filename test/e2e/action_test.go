// +build integration

package e2e

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"projectvoltron.dev/voltron/internal/ptr"
	enginegraphql "projectvoltron.dev/voltron/pkg/engine/api/graphql"
	client "projectvoltron.dev/voltron/pkg/engine/client"
	ochlocalgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	ochclient "projectvoltron.dev/voltron/pkg/och/client"
)

var _ = Describe("Action", func() {
	var engineClient *client.Client
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

		FIt("should download input TypeInstance", func() {
			fmt.Println("Running!!!")
			input := &ochlocalgraphql.CreateTypeInstanceInput{
				TypeRef: &ochlocalgraphql.TypeReferenceInput{
					Path:     "cap.type.e2e.test",
					Revision: "0.1.0",
				},
				Value: map[string]string{"key": "e2e test"},
				Attributes: []*ochlocalgraphql.AttributeReferenceInput{
					{
						Path:     "com.voltron.attribute1",
						Revision: "0.1.0",
					},
				},
			}
			typeInstance, err := ochClient.CreateTypeInstance(ctx, input)
			Expect(err).ToNot(HaveOccurred())

			inputTypeInstance := enginegraphql.InputTypeInstanceData{Name: "e2e", ID: typeInstance.Metadata.ID}

			_, err = engineClient.CreateAction(ctx, &enginegraphql.ActionDetailsInput{
				Name: actionName,
				ActionRef: &enginegraphql.ManifestReferenceInput{
					Path:     "cap.interface.voltron.e2e.passingWithTypeInstance",
					Revision: ptr.String("0.1.0"),
				},
				Input: &enginegraphql.ActionInputData{
					TypeInstances: []*enginegraphql.InputTypeInstanceData{&inputTypeInstance},
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

			// TODO when upload TypeInstances is ready add uploading result of cat and compare with input TypeInstance

		})

	})
})

func getActionStatusFunc(ctx context.Context, cl *client.Client, name string) func() (enginegraphql.ActionStatusPhase, error) {
	return func() (enginegraphql.ActionStatusPhase, error) {
		action, err := cl.GetAction(ctx, name)
		if err != nil {
			return "", err
		}
		return action.Status.Phase, err
	}
}
