package graphql

import (
	"bytes"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON_UnmarshalGQL(t *testing.T) {
	for name, tc := range map[string]struct {
		input         interface{}
		expected      JSON
		expectedError error
	}{
		//given
		"correct input: object": {
			input:         `{"schema":"schema"}`,
			expected:      JSON(`{"schema":"schema"}`),
			expectedError: nil,
		},
		"correct input: string": {
			input:         `"test"`,
			expected:      JSON(`"test"`),
			expectedError: nil,
		},
		"correct input: number": {
			input:         `11`,
			expected:      JSON(`11`),
			expectedError: nil,
		},
		"error: empty input": {
			input:         nil,
			expected:      "",
			expectedError: errors.New("input should not be nil"),
		},
		"error: invalid input": {
			input:         123,
			expected:      "",
			expectedError: errors.New("unexpected input type: int, should be string"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			//when
			var j JSON
			err := j.UnmarshalGQL(tc.input)

			if tc.expectedError == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedError.Error())
			}
			assert.Equal(t, tc.expected, j)
		})
	}
}

func TestJSON_MarshalGQL(t *testing.T) {
	for name, tc := range map[string]struct {
		input    JSON
		expected string
	}{
		//given
		"object": {
			input:    JSON(`{"schema":"schema"}`),
			expected: `"{\"schema\":\"schema\"}"`,
		},
		"number": {
			input:    JSON(`11`),
			expected: "\"11\"",
		},
	} {
		t.Run(name, func(t *testing.T) {
			buf := bytes.Buffer{}

			tc.input.MarshalGQL(&buf)

			require.Equal(t, tc.expected, buf.String())
		})
	}
}
