package capact

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

type image struct {
	Dir    string
	Target string

	ExtraBuildArgs  []string
	DisableBuildKit bool
}

var images = map[string]image{
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
	"helm-runner": {
		Dir:    ".",
		Target: "generic",
	},
	"cloudsql-runner": {
		Dir:    ".",
		Target: "generic",
	},
	"terraform-runner": {
		Dir:    ".",
		Target: "terraform-runner",
	},
	"populator": {
		Dir:    ".",
		Target: "generic-alpine",
	},
	"e2e": {
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

	for _, name := range names {
		image, ok := images[name]
		if !ok {
			return nil, fmt.Errorf("cannot find image %s", name)
		}
		imageTag, err := buildImage(w, name, image, repository, version)
		if err != nil {
			return nil, errors.Wrapf(err, "while building image %s", name)
		}
		created = append(created, imageTag)
	}
	return created, nil
}

// SelectImages returns a list of images calculated from focus and skip lists
func SelectImages(focus, skip []string) ([]string, error) {
	var selected []string
	if len(focus) > 0 {
		for _, name := range focus {
			_, ok := images[name]
			if !ok {
				return nil, fmt.Errorf("focused image does not exist: %s", name)
			}
			selected = append(selected, name)
		}
		return selected, nil
	}

	for image := range images {
		if shouldSkipImage(image, skip) {
			continue
		}
		selected = append(selected, image)
	}
	return selected, nil
}

func shouldSkipImage(name string, skipList []string) bool {
	for _, image := range skipList {
		if image == name {
			return true
		}
	}
	return false
}
