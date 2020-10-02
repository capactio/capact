// +build integration

package e2e

import (
	"testing"
	"time"

	"projectvoltron.dev/voltron/pkg/httputil"
	"projectvoltron.dev/voltron/pkg/iosafety"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	StatusEndpoints []string
}

func TestStatusEndpoints(t *testing.T) {
	var cfg Config
	err := envconfig.Init(&cfg)
	require.NoError(t, err)

	cli := httputil.NewClient(30*time.Second, true)

	for _, endpoint := range cfg.StatusEndpoints {
		resp, err := cli.Get(endpoint)
		assert.NoErrorf(t, err, "Get on %s", endpoint)

		err = iosafety.DrainReader(resp.Body)
		assert.NoError(t, err)
		err = resp.Body.Close()
		assert.NoError(t, err)
	}
}
