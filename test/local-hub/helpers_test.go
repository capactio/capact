//go:build localhub
// +build localhub

package localhub

import (
	"os"
	"testing"

	"capact.io/capact/pkg/hub/client/local"
)

const localhubAddrEnv = "LOCAL_HUB_ADDR"

func getLocalClient(t *testing.T) *local.Client {
	localhubAddr := os.Getenv(localhubAddrEnv)
	if localhubAddr == "" {
		t.Skipf("skipping running test as the env %s is not provided", localhubAddrEnv)
	}
	return local.NewDefaultClient(localhubAddr)
}
