package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty"
)

func TestLoadVariablesFromFiles(t *testing.T) {
	files := []string{"testdata/current.tfvars", "testdata/override.tfvars"}

	values, err := LoadVariablesFromFiles(files...)
	assert.NoError(t, err)

	assert.Equal(t, values["username"], cty.StringVal("testuser"))
	assert.Equal(t, values["password"], cty.StringVal("secret"))
	assert.Equal(t, values["region"], cty.StringVal("eu-central-1"))
}

func TestMarshalVariables(t *testing.T) {
	variables := map[string]cty.Value{
		"username": cty.StringVal("testuser"),
		"count":    cty.NumberIntVal(10),
		"enabled":  cty.BoolVal(true),
	}

	data := MarshalVariables(variables)
	assert.Contains(t, string(data), `username = "testuser"`)
	assert.Contains(t, string(data), `count    = 10`)
	assert.Contains(t, string(data), `enabled  = true`)
}
