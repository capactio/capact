package capact

import (
	"fmt"
	"os"
	"path/filepath"

	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/nodes"
	"sigs.k8s.io/kind/pkg/cluster/nodeutils"
	"sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/errors"
	"sigs.k8s.io/kind/pkg/exec"
	"sigs.k8s.io/kind/pkg/fs"
)

// LoadKindImage loads local docker images into a kind cluster
// Based on https://github.com/kubernetes-sigs/kind/blob/942198293a4cc7938bf759039fd6447c4b38ad1c/pkg/cmd/kind/load/docker-image/docker-image.go
func LoadKindImage(envName string, image string) error {
	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(cmd.NewLogger()),
		cluster.ProviderWithDocker(),
	)

	// Check that the image exists locally and gets its ID, if not return error
	imageID, err := imageID(image)
	if err != nil {
		return fmt.Errorf("image: %q not present locally", image)
	}

	// Check if the cluster nodes exist
	nodeList, err := provider.ListInternalNodes(envName)
	if err != nil {
		return err
	}
	if len(nodeList) == 0 {
		return fmt.Errorf("no nodes found for cluster %q", envName)
	}

	candidateNodes := nodeList

	// pick only the nodes that don't have the image
	var selectedNodes []nodes.Node
	for _, node := range candidateNodes {
		id, err := nodeutils.ImageID(node, image)
		if err != nil || id != imageID {
			selectedNodes = append(selectedNodes, node)
		}
	}
	if len(selectedNodes) == 0 {
		return nil
	}

	// Setup the tar path where the images will be saved
	dir, err := fs.TempDir("", "images-tar")
	if err != nil {
		return errors.Wrap(err, "failed to create tempdir")
	}
	defer os.RemoveAll(dir)
	imagesTarPath := filepath.Join(dir, "images.tar")
	// Save the images into a tar
	err = save([]string{image}, imagesTarPath)
	if err != nil {
		return err
	}

	var fns []func() error
	// Load the images on the selected nodes
	for _, selectedNode := range selectedNodes {
		selectedNode := selectedNode // capture loop variable
		fns = append(fns, func() error {
			return loadImage(imagesTarPath, selectedNode)
		})
	}
	return errors.UntilErrorConcurrent(fns)
}

// loads an image tarball onto a node
func loadImage(imageTarName string, node nodes.Node) error {
	f, err := os.Open(filepath.Clean(imageTarName))
	if err != nil {
		return errors.Wrap(err, "failed to open image")
	}
	defer f.Close()
	return nodeutils.LoadImageArchive(node, f)
}

// save saves images to dest, as in `docker save`
func save(images []string, dest string) error {
	commandArgs := append([]string{"save", "-o", dest}, images...)
	return exec.Command("docker", commandArgs...).Run()
}

// imageID return the Id of the container image
func imageID(containerNameOrID string) (string, error) {
	cmd := exec.Command("docker", "image", "inspect",
		"-f", "{{ .Id }}",
		containerNameOrID, // ... against the container
	)
	lines, err := exec.OutputLines(cmd)
	if err != nil {
		return "", err
	}
	if len(lines) != 1 {
		return "", errors.Errorf("Docker image ID should only be one line, got %d lines", len(lines))
	}
	return lines[0], nil
}
