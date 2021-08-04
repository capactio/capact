package cmd

import (
	"context"
	"fmt"
	"github.com/rancher/k3d/v4/pkg/runtimes"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test(t *testing.T) {
	imgs, err := runtimes.SelectedRuntime.GetImages(context.Background())
	require.NoError(t, err)
	fmt.Println(imgs)
}

