package validate_test

import (
	"context"
	"io/ioutil"
	"path"
	"strings"
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
	files, err := ioutil.ReadDir(pathToExamples)
	require.NoError(t, err)

	var filePaths []string
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		filePaths = append(filePaths, path.Join(pathToExamples, file.Name()))
	}

	require.True(t, len(filePaths) > 0)
	t.Log(filePaths)

	// when
	err = validation.Run(context.Background(), filePaths)

	// then
	assert.NoError(t, err)
}

func TestValidation_NoFiles(t *testing.T) {
	// given
	validation, err := validate.New(ioutil.Discard, validate.Options{MaxConcurrency: 5})
	require.NoError(t, err)

	filePaths := []string{"/this/file/doesnt/exist", "/same/here"}

	require.True(t, len(filePaths) > 0)
	t.Log(filePaths)

	// when
	err = validation.Run(context.Background(), filePaths)

	// then
	assert.Error(t, err)
	assert.EqualError(t, err, "detected 2 validation errors")
}
