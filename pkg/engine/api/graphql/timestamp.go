package graphql

import (
	"io"
	"log"
	"strconv"
	"time"

	"capact.io/capact/internal/graphqlutil"
)

// Timestamp is a GraphQL scalar, which represents time. It can be used like the Go time.Time struct.
type Timestamp struct {
	time.Time
}

// UnmarshalGQL unmarshals the provided GraphQL scalar to this Timestamp struct. The scalar must be a string representing time formatted in RFC3339.
func (t *Timestamp) UnmarshalGQL(v interface{}) error {
	tmpStr, err := graphqlutil.ScalarToString(v)
	if err != nil {
		return err
	}

	parse, err := time.Parse(time.RFC3339, tmpStr)
	if err != nil {
		return err
	}

	*t = Timestamp{parse}
	return nil
}

// MarshalGQL writes the RFC3339 formatted Timestamp to the provided writer.
func (t Timestamp) MarshalGQL(w io.Writer) {
	_, err := w.Write([]byte(strconv.Quote(t.Format(time.RFC3339))))
	if err != nil {
		log.Printf("while writing %T: %v", t, err)
	}
}
