package capact

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"

	"capact.io/capact/internal/cli"

	"capact.io/capact/internal/cli/printer"
	"github.com/pkg/errors"

	"k8s.io/utils/strings/slices"
)

type image struct {
	Dir    string
	Target string

	ExtraBuildArgs  []string
	DisableBuildKit bool
}

type images map[string]image

func (i images) All() []string {
	var all []string
	for img := range i {
		all = append(all, img)
	}

	// We generate doc automatically, so it needs to be deterministic
	sort.Strings(all)
	return all
}

// Images is a list of all Capact Docker images available to build
var Images = images{
	"gateway": {
		Dir:    ".",
		Target: "generic",
	},
	"k8s-engine": {
		Dir:    ".",
		Target: "generic",
	},
	"hub-js": {
		Dir: "hub-js",

		DisableBuildKit: true,
	},
	"argo-runner": {
		Dir:    ".",
		Target: "generic",
	},
	"argo-actions": {
		Dir:    ".",
		Target: "generic",
	},
	"populator": {
		Dir:    ".",
		Target: "generic-alpine",
	},
	"e2e-test": {
		Dir:    ".",
		Target: "e2e",
		ExtraBuildArgs: []string{
			"BUILD_CMD=go test -v -c",
			"SOURCE_PATH=./test/e2e/*_test.go",
		},
	},
}

func buildImage(ctx context.Context, status *printer.Status, imgName string, img image, repository, version string) (string, error) {
	// docker build --build-arg COMPONENT=$(APP) --target generic -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG)
	imageTag := fmt.Sprintf("%s/%s:%s", repository, imgName, version)
	// #nosec G204
	cmd := exec.CommandContext(ctx, "docker",
		"build",
		"--build-arg", fmt.Sprintf("COMPONENT=%s", imgName),
		"-t", imageTag, ".")

	if img.Target != "" {
		cmd.Args = append(cmd.Args, "--target", img.Target)
	}

	cmd.Env = os.Environ()

	if !img.DisableBuildKit {
		cmd.Env = append(cmd.Env,
			// enable the BuildKit builder in the Docker CLI.
			"DOCKER_BUILDKIT=1")
	}

	if len(img.ExtraBuildArgs) != 0 {
		for _, arg := range img.ExtraBuildArgs {
			cmd.Args = append(cmd.Args, "--build-arg", arg)
		}
	}
	cmd.Dir = img.Dir
	if err := runCMD(cmd, status, "Building image %s", imageTag); err != nil {
		return "", err
	}

	return imageTag, nil
}

// BuildImages builds passed images setting passed repository and version
func BuildImages(ctx context.Context, status *printer.Status, repository, version string, names []string) ([]string, error) {
	var created []string

	for _, image := range Images.All() {
		if !slices.Contains(names, image) {
			continue
		}
		imageTag, err := buildImage(ctx, status, image, Images[image], repository, version)
		if err != nil {
			return nil, errors.Wrapf(err, "while building image %s", image)
		}
		created = append(created, imageTag)
	}
	return created, nil
}

// PushImages pushes passed images to a given registry
func PushImages(ctx context.Context, status *printer.Status, names []string) error {
	for _, image := range names {
		// #nosec G204
		cmd := exec.CommandContext(ctx, "docker", "push", image)

		if err := runCMD(cmd, status, "Pushing %s", image); err != nil {
			return err
		}
	}
	return nil
}

func runCMD(cmd *exec.Cmd, status *printer.Status, stageFmt string, args ...interface{}) error {
	if cli.VerboseMode.IsEnabled() {
		cmd.Stdout = status.Writer()
		cmd.Stderr = status.Writer()
		return cmd.Run()
	}

	var buff bytes.Buffer
	status.Step(stageFmt, args...)
	cmd.Stderr = &buff
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(err, "while running cmd [stderr: %s]", buff.String())
	}

	return nil
}
