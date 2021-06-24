package graphql

import (
	"encoding/json"
	"io"
	"strconv"

	"capact.io/capact/internal/graphqlutil"

	"log"
)

// JSON represents a GraphQL scalar, which holds a JSON string.
type JSON string

// UnmarshalGQL unmarshals the GraphQL input to JSON.
func (j *JSON) UnmarshalGQL(v interface{}) error {
	val, err := graphqlutil.ScalarToString(v)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(val), new(interface{}))
	if err != nil {
		return err
	}

	*j = JSON(val)
	return nil
}

// MarshalGQL writes the JSON to the provided writer.
func (j JSON) MarshalGQL(w io.Writer) {
	_, err := io.WriteString(w, strconv.Quote(string(j)))
	if err != nil {
		log.Printf("while writing %T: %s", j, err)
	}
}
