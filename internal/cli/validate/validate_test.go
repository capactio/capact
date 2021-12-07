package validate_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"capact.io/capact/internal/cli/validate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidation_Run_SmokeTest(t *testing.T) {
	// given
	validation, err := validate.New(ioutil.Discard, validate.Options{MaxConcurrency: 5})
	require.NoError(t, err)

	pathToExamples := "../../../ocf-spec/0.0.1/examples"

	// when
	err = validation.Run(context.Background(), []string{pathToExamples})

	// then
	assert.NoError(t, err)
}

func TestValidation_NoFiles(t *testing.T) {
	// given
	validation, err := validate.New(ioutil.Discard, validate.Options{MaxConcurrency: 5})
	require.NoError(t, err)

	filePaths := []string{"/this/file/doesnt/exist", "/same/here"}

	// when
	err = validation.Run(context.Background(), filePaths)

	// then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestValidation_NoRecursive(t *testing.T) {
	// given
	var buff = &bytes.Buffer{}
	validation, err := validate.New(buff, validate.Options{MaxConcurrency: 5, RecursiveSearch: false})
	require.NoError(t, err)

	pathToExamples := "../../../ocf-spec/0.0.1"

	// when
	err = validation.Run(context.Background(), []string{pathToExamples})

	// then
	assert.NoError(t, err)
	assert.Contains(t, buff.String(), "Validated 0 files in total")
}

func TestValidation_Recursive(t *testing.T) {
	// given
	var buff = &bytes.Buffer{}
	validation, err := validate.New(buff, validate.Options{MaxConcurrency: 5, RecursiveSearch: true})
	require.NoError(t, err)

	pathToExamples := "../../../ocf-spec/0.0.1/"

	// when
	err = validation.Run(context.Background(), []string{pathToExamples})

	// then
	assert.NoError(t, err)
	assert.Contains(t, buff.String(), "Validated 7 files in total")
}
