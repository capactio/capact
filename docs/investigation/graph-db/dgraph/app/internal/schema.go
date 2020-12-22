package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// curl -X POST localhost:8080/admin/schema --data-binary '@assets/schema.graphql'
func MustInitSchema(schemaPath string) {
	f, err := os.Open(schemaPath)
	requireNoErr(err)
	defer f.Close()

	// using dgo client creates ONLY RDF schemas and GraphQL queries are not supported
	req, err := http.NewRequest("POST", "http://localhost:8080/admin/schema", f)
	requireNoErr(err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	requireNoErr(err)
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)

	body := struct {
		Errors []map[string]interface{} `json:"errors"`
	}{}
	err = json.Unmarshal(respBody, &body)
	requireNoErr(err)

	if len(body.Errors) > 0 {
		requireNoErr(fmt.Errorf("cannot upload schema: response: %s", respBody))
	}
}
