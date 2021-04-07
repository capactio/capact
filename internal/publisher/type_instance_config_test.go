package publisher

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
)

func TestLookupNS_Unmarshal_Success_Set(t *testing.T) {
	// given
	var conf struct {
		HelmReleasesNSLookup LookupNS
	}

	os.Setenv("HELM_RELEASES_NS_LOOKUP", "ns1,ns2")
	defer os.Unsetenv("HELM_RELEASES_NS_LOOKUP")

	// when
	err := envconfig.Init(&conf)

	// then
	require.NoError(t, err)
	assert.True(t, conf.HelmReleasesNSLookup.Has("ns1"))
	assert.True(t, conf.HelmReleasesNSLookup.Has("ns2"))
}
