package graphql

import (
	"io"
	"log"
	"strconv"
	"time"

	"projectvoltron.dev/voltron/internal/graphqlutil"
)

type Timestamp time.Time

func (t *Timestamp) UnmarshalGQL(v interface{}) error {
	tmpStr, err := graphqlutil.ScalarToString(v)
	if err != nil {
		return err
	}

	parse, err := time.Parse(time.RFC3339, tmpStr)
	if err != nil {
		return err
	}

	*t = Timestamp(parse)
	return nil
}

func (t Timestamp) MarshalGQL(w io.Writer) {
	_, err := w.Write([]byte(strconv.Quote(time.Time(t).Format(time.RFC3339))))
	if err != nil {
		log.Printf("while writing %T: %v", t, err)
	}
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	return time.Time(t).MarshalJSON()
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	return (*time.Time)(t).UnmarshalJSON(data)
}
