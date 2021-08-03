package capact

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"

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

func buildImage(w io.Writer, imgName string, img image, repository, version string) (string, error) {
	// docker build --build-arg COMPONENT=$(APP) --target generic -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG)
	imageTag := fmt.Sprintf("%s/%s:%s", repository, imgName, version)

	// #nosec G204
	cmd := exec.Command("docker",
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
	cmd.Stdout = w
	cmd.Stderr = w
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return imageTag, nil
}

// DeleteImages deletes passed image
func DeleteImages(images []string) error {
	for _, image := range images {
		// #nosec G204
		cmd := exec.Command("docker", "rmi", image)
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

// BuildImages builds passed images setting passed repository and version
func BuildImages(w io.Writer, repository, version string, names []string) ([]string, error) {
	var created []string

	for _, image := range Images.All() {
		ok := slices.Contains(names, image)
		if !ok {
			return nil, fmt.Errorf("cannot find image %s", image)
		}
		imageTag, err := buildImage(w, image, Images[image], repository, version)
		if err != nil {
			return nil, errors.Wrapf(err, "while building image %s", image)
		}
		created = append(created, imageTag)
	}
	return created, nil
}
