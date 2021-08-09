package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeSchemaCollectionHappyPath(t *testing.T) {
	tests := map[string]struct {
		givenCollections []SchemaCollection
		expCollection    SchemaCollection
	}{
		"empty input gives empty result": {
			givenCollections: []SchemaCollection{},
			expCollection:    SchemaCollection{},
		},
		"two collection merged into one": {
			givenCollections: []SchemaCollection{
				{
					"input1": {
						Value:    "val1",
						Required: false,
					},
				},
				{
					"input2": {
						Value:    "val2",
						Required: true,
					},
				},
			},
			expCollection: SchemaCollection{
				"input1": {
					Value:    "val1",
					Required: false,
				},
				"input2": {
					Value:    "val2",
					Required: true,
				},
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// when
			got, err := MergeSchemaCollection(tc.givenCollections...)

			// then
			require.NoError(t, err)
			assert.Equal(t, got, tc.expCollection)
		})
	}
}

func TestMergeSchemaCollectionFailure(t *testing.T) {
	// given
	givenCollections := []SchemaCollection{
		{
			"input1": Schema{
				Value:    "in1",
				Required: false,
			},
		},
		{
			"input1": Schema{
				Value:    "in2",
				Required: false,
			},
		},
	}

	// when
	got, err := MergeSchemaCollection(givenCollections...)

	// then
	assert.EqualError(t, err, `cannot merge schema collections, found name collision for "input1"`)
	assert.Nil(t, got)
}
