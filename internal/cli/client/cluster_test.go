package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeServerEndpoint(t *testing.T) {
	tt := []struct {
		input            string
		expectedEndpoint string
	}{
		{
			input:            "https://capact.local",
			expectedEndpoint: "https://capact.local/graphql",
		},
		{
			input:            "http://capact.local",
			expectedEndpoint: "http://capact.local/graphql",
		},
		{
			input:            "capact.local",
			expectedEndpoint: "https://capact.local/graphql",
		},
	}

	for _, tc := range tt {
		normalized := normalizeServerEndpoint(tc.input)
		assert.Equal(t, tc.expectedEndpoint, normalized)
	}
}
