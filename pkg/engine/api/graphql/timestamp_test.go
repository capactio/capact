package graphql

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimestampUnmarshalGQL(t *testing.T) {
	//given
	var timestamp Timestamp
	fixTime := "2002-10-02T10:00:00-05:00"
	parsedTime, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	assert.NoError(t, err)
	expectedTimestamp := Timestamp(parsedTime)

	//when
	err = timestamp.UnmarshalGQL(fixTime)

	//then
	require.NoError(t, err)
	assert.Equal(t, expectedTimestamp, timestamp)
}

func TestTimestampUnmarshalGQLError(t *testing.T) {
	t.Run("invalid input", func(t *testing.T) {
		//given
		var timestamp Timestamp
		invalidInput := 123

		//when
		err := timestamp.UnmarshalGQL(invalidInput)

		//then
		require.Error(t, err)
		assert.Empty(t, timestamp)

	})

	t.Run("can't parse time", func(t *testing.T) {
		//given
		var timestamp Timestamp
		invalidTime := "invalid time"

		//when
		err := timestamp.UnmarshalGQL(invalidTime)

		//then
		require.Error(t, err)
		assert.Empty(t, timestamp)

	})
}

func TestTimestampMarshalGQL(t *testing.T) {
	//given
	parsedTime, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	assert.NoError(t, err)
	fixTimestamp := Timestamp(parsedTime)
	expectedTimestamp := `"2002-10-02T10:00:00-05:00"`
	buf := bytes.Buffer{}

	//when
	fixTimestamp.MarshalGQL(&buf)

	//then
	assert.Equal(t, expectedTimestamp, buf.String())
}

func TestTimestampUmarshalJSON(t *testing.T) {
	// given
	ts := &Timestamp{}

	// when
	err := ts.UnmarshalJSON([]byte(`"2002-10-02T10:00:00-05:00"`))

	// then
	require.NoError(t, err)
	tm := time.Time(*ts)
	assert.Equal(t, 2002, tm.Year())
}
