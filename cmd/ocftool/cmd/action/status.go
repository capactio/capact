package action

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"projectvoltron.dev/voltron/internal/ocftool"
	"projectvoltron.dev/voltron/internal/ocftool/client"
	"projectvoltron.dev/voltron/internal/ocftool/config"
	"projectvoltron.dev/voltron/internal/ocftool/heredoc"

	"github.com/argoproj/argo/cmd/argo/commands"
	argocli "github.com/argoproj/argo/cmd/argo/commands/client"
	workflowpkg "github.com/argoproj/argo/pkg/apiclient/workflow"
	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

var hiddenFlags = []string{
	"as",
	"as-group",
	"certificate-authority",
	"client-certificate",
	"client-key",
	"cluster",
	"context",
	"help",
	"insecure-skip-tls-verify",
	"kubeconfig",
	"no-utf8",
	"node-field-selector",
	"password",
	"request-timeout",
	"server",
	"token",
	"user",
	"username",
	"node-field-selector",
}

func NewStatus() *cobra.Command {
	status := commands.NewGetCommand()

	status.RunE = wrapRun(status.Run)
	status.Run = nil
	status.Args = cobra.MaximumNArgs(1)
	status.Use = "status ACTION"
	status.Short = "Show Action status"
	status.Example = heredoc.WithCLIName(`
		# Get information about a workflow:
		<cli> action status my-action
		
		# Get the latest workflow:
		<cli> action status @latest
	`, ocftool.CLIName)

	argocli.AddKubectlFlagsToCmd(status)

	for _, hide := range hiddenFlags {
		// set flags exits
		_ = status.PersistentFlags().MarkHidden(hide)
	}

	return status
}

type cobraRunFn func(cmd *cobra.Command, args []string)
type cobraRunEFn func(cmd *cobra.Command, args []string) error

func wrapRun(underlying cobraRunFn) cobraRunEFn {
	return func(cmd *cobra.Command, args []string) error {
		underlying(cmd, args)

		ctx, apiClient := argocli.NewAPIClient()
		serviceClient := apiClient.NewWorkflowServiceClient()
		namespace := argocli.Namespace()

		server, err := config.GetDefaultContext()
		if err != nil {
			return err
		}

		actionCli, err := client.NewCluster(server)
		if err != nil {
			return err
		}

		name := args[0]
		wf, err := serviceClient.GetWorkflow(ctx, &workflowpkg.WorkflowGetRequest{
			Name:      name,
			Namespace: namespace,
		})

		if err != nil {
			return err
		}

		if wf.Status.Phase != wfv1.NodeSucceeded {
			return nil
		}

		podName := getUploadPodName(wf.Status.Nodes)

		stream, err := serviceClient.WorkflowLogs(ctx, &workflowpkg.WorkflowLogRequest{
			Name:      name,
			Namespace: namespace,
			PodName:   podName,
			LogOptions: &v1.PodLogOptions{
				Container: "main",
				Follow:    false,
				Previous:  false,
			},
		})
		if err != nil {
			return err
		}

		outputTI, err := getUploadedTypeInstance(ctx, actionCli, stream)
		if err != nil {
			return err
		}

		data, err := yaml.Marshal(outputTI)
		if err != nil {
			return errors.Wrap(err, "while marshaling TypeInstance to YAML")
		}

		fmt.Printf("Output TypeInstance:\n %s", data)
		return nil
	}
}

func getUploadPodName(nodes wfv1.Nodes) string {
	for key, node := range nodes {
		if node.DisplayName == "upload-output-type-instances-step" {
			return key
		}
	}
	return ""
}

type logMsg struct {
	Alias string `json:"alias"`
	ID    string `json:"id"`
}

func getUploadedTypeInstance(ctx context.Context, actionCli client.ClusterClient, stream workflowpkg.WorkflowService_WorkflowLogsClient) (map[string]interface{}, error) {
	outputTI := map[string]interface{}{}
	for {
		event, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		msg := logMsg{}
		if err := json.Unmarshal([]byte(event.Content), &msg); err != nil {
			return nil, err
		}
		if msg.Alias == "" || msg.ID == "" {
			continue
		}

		typeInstance, err := actionCli.FindTypeInstance(ctx, msg.ID)
		if err != nil {
			return nil, err
		}
		if typeInstance == nil {
			return nil, fmt.Errorf("failed to find TypeInstance with ID %q", msg.ID)
		}

		outputTI[msg.Alias] = typeInstance.Spec
	}

	return outputTI, nil
}
