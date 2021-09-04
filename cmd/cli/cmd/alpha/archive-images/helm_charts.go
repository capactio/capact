package archiveimages

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"capact.io/capact/internal/cli"
	"capact.io/capact/internal/cli/capact"
	"capact.io/capact/internal/cli/heredoc"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var image = regexp.MustCompile(`image:(.*)`)
var prometheusReloader = regexp.MustCompile(`--prometheus-config-reloader=(.*)`)

func NewFromHelmCharts() *cobra.Command {
	opts := capact.Options{
		DryRun:     true,
		Replace:    true,
		ClientOnly: true,
	}
	var outputPath string

	cmd := &cobra.Command{
		Use:   "helm",
		Short: "Archive all the container images listed in Capact Helm charts.",
		Example: heredoc.WithCLIName(`
			# List all the Hub configuration contexts 
			<cli> alpha archive-images helm
		`, cli.Name),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := opts.Parameters.ResolveVersion()
			if err != nil {
				return errors.Wrap(err, "while resolving version")
			}

			images := map[string]struct{}{}
			//logger, err := logger.New(logger.Config{})
			//if err != nil {
			//	return errors.Wrap(err, "while creating zap logger")
			//}
			//helmOutputter := helm.NewOutputter(logger, helm.NewRenderer())
			//var additionalOutput = map[string]helm.OutputArgs{
			//	"ingress-nginx": {
			//		GoTemplate: `image: "{{- if .repository -}}{{ .repository }}{{ else }}{{ .registry }}/{{ .image }}{{- end -}}:{{ .tag }}{{- if (.digest) -}} @{{.digest}} {{- end -}}"`,
			//	},
			//}
			for _, component := range capact.Components {
				component.WithOptions(&opts)

				rel, err := component.RunInstall(opts.Parameters.Version, map[string]interface{}{})
				if err != nil {
					return err
				}
				out := image.FindAllStringSubmatch(rel.Manifest, -1)
				for _, e := range out {
					if len(e) == 2 {
						images[sanitizeImageString(e[1])] = struct{}{}
					}
				}
				out = prometheusReloader.FindAllStringSubmatch(rel.Manifest, -1)
				for _, e := range out {
					if len(e) == 2 {
						images[sanitizeImageString(e[1])] = struct{}{}
					}
				}
			}

			for _, i := range additionalImages {
				images[i] = struct{}{}
			}

			return CacheDepImages(cmd.Context(), images, outputPath)
		},
	}

	cmd.Flags().StringVar(&opts.Parameters.Version, "version", capact.LatestVersionTag, "Capact version. Possible values @latest, @local, 0.3.0, ...")
	cmd.Flags().StringVar(&opts.Parameters.Override.HelmRepoURL, "helm-repo-url", capact.HelmRepoStable, fmt.Sprintf("Capact Helm chart repository URL. Use %s tag to select repository which holds the latest Helm chart versions.", capact.LatestVersionTag))

	cmd.Flags().StringVar(&outputPath, "output-path", "", "Defines")
	cmd.MarkFlagRequired("output-path")
	return cmd
}

// Is hard without that being implemented: https://github.com/helm/helm/issues/7754
var missing = `
quay.io/prometheus-operator/prometheus-config-reloader:v0.48.1 # https://github.com/prometheus-operator/kube-prometheus/blob/075875e8aaaeb2bc8a6b76c3776272a0b98ac86a/manifests/setup/prometheus-operator-deployment.yaml#L31

image: "{{- if .repository -}}{{ .repository }}{{ else }}{{ .registry }}/{{ .image }}{{- end -}}:{{ .tag }}{{- if (.digest) -}} @{{.digest}} {{- end -}}"
4m7s        Normal    Pulling                 pod/ingress-nginx-admission-create-wj44h # webhook
Pulling image "docker.io/jettech/kube-webhook-certgen:v1.5.1"

3m51s       Normal    Pulling                 pod/ingress-nginx-controller-758d79bf79-bpgpl    
Pulling image "k8s.gcr.io/ingress-nginx/controller:v0.47.0@sha256:a1e4efc107be0bb78f32eaec37bef17d7a0c81bec8066cdf2572508d21351d0b"
k8s.gcr.io/ingress-nginx/controller:v0.47.0@sha256:a1e4efc107be0bb78f32eaec37bef17d7a0c81bec8066cdf2572508d21351d0b

3m14s       Normal    Pulling                 pod/monitoring-kube-prometheus-admission-create-djq6n     
image: {{ .Values.prometheusOperator.admissionWebhooks.patch.image.repository }}:{{ .Values.prometheusOperator.admissionWebhooks.patch.image.tag }}
Pulling image "jettech/kube-webhook-certgen:v1.5.2"
`

var additionalImages = []string{"docker.io/jettech/kube-webhook-certgen:v1.5.1", "jettech/kube-webhook-certgen:v1.5.2"}

func sanitizeImageString(s string) string {
	s = strings.Replace(s, "\"", "", 2)
	return strings.TrimSpace(s)
}

func CacheDepImages(ctx context.Context, images map[string]struct{}, outPath string) error {
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	var pulled []string
	for img := range images {
		reader, err := dockerCli.ImagePull(ctx, img, types.ImagePullOptions{})
		if err != nil {
			return err
		}

		err = jsonmessage.DisplayJSONMessagesStream(reader, os.Stdout, os.Stdout.Fd(), cli.IsSmartTerminal(os.Stdout), nil)
		if err != nil {
			return err
		}
		// todo:
		reader.Close()
		pulled = append(pulled, img)
	}

	responseBody, err := dockerCli.ImageSave(ctx, pulled)
	if err != nil {
		return err
	}
	defer responseBody.Close()

	err = command.CopyToFile(outPath, responseBody)
	if err != nil {
		return err
	}

	return nil
}
