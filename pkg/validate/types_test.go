package validate

import (
	"testing"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"

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

func TestMergeTypeRefCollectionHappyPath(t *testing.T) {
	tests := map[string]struct {
		givenCollections []TypeRefCollection
		expCollection    TypeRefCollection
	}{
		"empty input gives empty result": {
			givenCollections: []TypeRefCollection{},
			expCollection:    TypeRefCollection{},
		},
		"two collection merged into one": {
			givenCollections: []TypeRefCollection{
				{
					"input1": {
						TypeRef: types.TypeRef{
							Path:     "path1",
							Revision: "rev1",
						},
						Required: false,
					},
				},
				{
					"input2": {
						TypeRef: types.TypeRef{
							Path:     "path2",
							Revision: "rev2",
						},
						Required: true,
					},
				},
			},
			expCollection: TypeRefCollection{
				"input1": {
					TypeRef: types.TypeRef{
						Path:     "path1",
						Revision: "rev1",
					},
					Required: false,
				},
				"input2": {
					TypeRef: types.TypeRef{
						Path:     "path2",
						Revision: "rev2",
					},
					Required: true,
				},
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// when
			got, err := MergeTypeRefCollection(tc.givenCollections...)

			// then
			require.NoError(t, err)
			assert.Equal(t, got, tc.expCollection)
		})
	}
}

func TestMergeTypeRefCollectionFailure(t *testing.T) {
	// given
	givenCollections := []TypeRefCollection{
		{
			"input1": TypeRef{
				TypeRef: types.TypeRef{
					Path:     "path1",
					Revision: "rev1",
				},
				Required: false,
			},
		},
		{
			"input1": TypeRef{
				TypeRef: types.TypeRef{
					Path:     "path1",
					Revision: "rev2",
				},
				Required: true,
			},
		},
	}

	// when
	got, err := MergeTypeRefCollection(givenCollections...)

	// then
	assert.EqualError(t, err, `cannot merge input TypeRef collection, found name collision for "input1"`)
	assert.Nil(t, got)
}
