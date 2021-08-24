package maps_test

import (
	"testing"

	"capact.io/capact/internal/maps"
	"github.com/stretchr/testify/assert"
)

func TestMergeMaps(t *testing.T) {
	tests := []struct {
		name       string
		current    map[string]interface{}
		overwrites map[string]interface{}
		expected   map[string]interface{}
	}{
		{
			name:       "Nothing",
			current:    map[string]interface{}{},
			overwrites: map[string]interface{}{},
			expected:   map[string]interface{}{},
		},
		{
			name:       "empty overwrites",
			current:    map[string]interface{}{"A": 1},
			overwrites: map[string]interface{}{},
			expected:   map[string]interface{}{"A": 1},
		},
		{
			name:       "empty current",
			current:    map[string]interface{}{},
			overwrites: map[string]interface{}{"A": 1},
			expected:   map[string]interface{}{"A": 1},
		},
		{
			name:       "merge different",
			current:    map[string]interface{}{"A": 1},
			overwrites: map[string]interface{}{"B": 2},
			expected:   map[string]interface{}{"A": 1, "B": 2},
		},
		{
			name:       "simple overwrite",
			current:    map[string]interface{}{"A": 1, "B": 2},
			overwrites: map[string]interface{}{"A": 2},
			expected:   map[string]interface{}{"A": 2, "B": 2},
		},
		{
			name:       "nested overwrite",
			current:    map[string]interface{}{"A": 1, "B": map[string]interface{}{"C": 2, "E": 5}},
			overwrites: map[string]interface{}{"A": 1, "B": map[string]interface{}{"C": 3, "D": 4}},
			expected:   map[string]interface{}{"A": 1, "B": map[string]interface{}{"C": 3, "D": 4, "E": 5}},
		},
		{
			name:       "change type",
			current:    map[string]interface{}{"A": 1, "B": 2},
			overwrites: map[string]interface{}{"A": "1"},
			expected:   map[string]interface{}{"A": "1", "B": 2},
		},
		{
			name:       "replace list",
			current:    map[string]interface{}{"A": []int{0, 1, 2}},
			overwrites: map[string]interface{}{"A": []int{1, 2, 3}},
			expected:   map[string]interface{}{"A": []int{1, 2, 3}},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result := maps.Merge(test.current, test.overwrites)
			assert.Equalf(t, test.expected, result, "Merged map is different from expected")
		})
	}
}
