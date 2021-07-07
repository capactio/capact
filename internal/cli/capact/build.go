package capact

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

type image struct {
	Name   string
	Dir    string
	Target string

	ExtraBuildArgs  []string
	DisableBuildKit bool
}

var images = []image{
	{
		Name:   "gateway",
		Dir:    ".",
		Target: "generic",
	},
	{
		Name:   "k8s-engine",
		Dir:    ".",
		Target: "generic",
	},
	{
		Name: "hub-js",
		Dir:  "hub-js",

		DisableBuildKit: true,
	},
	{
		Name:   "argo-runner",
		Dir:    ".",
		Target: "generic",
	},
	{
		Name:   "helm-runner",
		Dir:    ".",
		Target: "generic",
	},
	{
		Name:   "cloudsql-runner",
		Dir:    ".",
		Target: "generic",
	},
	{
		Name:   "terraform-runner",
		Dir:    ".",
		Target: "terraform-runner",
	},
	{
		Name:   "populator",
		Dir:    ".",
		Target: "generic-alpine",
	},
	{
		Name:   "e2e",
		Dir:    ".",
		Target: "e2e",
		ExtraBuildArgs: []string{
			"BUILD_CMD=go test -v -c",
			"SOURCE_PATH=./test/e2e/*_test.go",
		},
	},
}

func buildImage(w io.Writer, img image, repository, version string) (string, error) {
	// docker build --build-arg COMPONENT=$(APP) --target generic -t $(DOCKER_REPOSITORY)/$(APP):$(DOCKER_TAG)
	imageTag := fmt.Sprintf("%s/%s:%s", repository, img.Name, version)

	// #nosec G204
	cmd := exec.Command("docker",
		"build",
		"--build-arg", fmt.Sprintf("COMPONENT=%s", img.Name),
		"-t", imageTag, ".")

	if img.Target != "" {
		cmd.Args = append(cmd.Args, "--target", img.Target)
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, []string{
		// enable module support across all go commands.
		"GO111MODULE=on",
		// enable consistent Go 1.12/1.13 GOPROXY behavior.
		"GOPROXY=https://proxy.golang.org",
	}...)

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

// DeleteImage deletes passed image
func DeleteImage(images []string) error {
	// #nosec G204
	cmd := exec.Command("docker",
		"rmi", images[0]) //TODO
	stdoutStderr, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", stdoutStderr)
	if err != nil {
		return err
	}
	return nil
}

// BuildImages builds passed images setting passed repository and version
func BuildImages(w io.Writer, repository, version string, names []string) ([]string, error) {
	imagesMap := mappedImages()
	var created []string

	for _, name := range names {
		image, ok := imagesMap[name]
		if !ok {
			return nil, fmt.Errorf("can not find image %s", name)
		}
		imageTag, err := buildImage(w, image, repository, version)
		if err != nil {
			return nil, errors.Wrapf(err, "while building image %s", image.Name)
		}
		created = append(created, imageTag)
	}
	return created, nil
}

// SelectImages returns a list of images calculated from focus and skip lists
func SelectImages(focus, skip []string) ([]string, error) {
	if len(focus) > 0 && len(skip) > 0 {
		return nil, errors.New("can not skip and focus images at the same time")
	}

	imagesMap := mappedImages()

	var selected []string
	if len(focus) > 0 {
		for _, name := range focus {
			_, ok := imagesMap[name]
			if !ok {
				return nil, fmt.Errorf("focused image does not exist: %s", name)
			}
			selected = append(selected, name)
		}
		return selected, nil
	}

	for _, image := range images {
		if shouldSkipImage(image.Name, skip) {
			continue
		}
		selected = append(selected, image.Name)
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

func mappedImages() map[string]image {
	imagesMap := map[string]image{}
	for _, image := range images {
		imagesMap[image.Name] = image
	}
	return imagesMap
}
