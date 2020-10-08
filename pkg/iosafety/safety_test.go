package iosafety_test

import (
	"io"
	"testing"

	"projectvoltron.dev/voltron/pkg/iosafety"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type readerMock struct {
	mock.Mock
}

func (rm *readerMock) Read(p []byte) (int, error) {
	args := rm.Called(p)
	return args.Int(0), args.Error(1)
}

func TestDrainReadsAll(t *testing.T) {
	// given
	reader := &readerMock{}
	reader.On("Read", mock.Anything).Return(500, nil).Twice()
	reader.On("Read", mock.Anything).Return(0, io.EOF)

	// when
	err := iosafety.DrainReader(reader)

	// then
	assert.NoError(t, err)
	reader.AssertExpectations(t)
}

func TestDrainWhenLimitReached(t *testing.T) {
	// given
	reader := &readerMock{}
	reader.On("Read", mock.Anything).Return(5000, nil).Once()

	// when
	err := iosafety.DrainReader(reader)

	// then
	reader.AssertExpectations(t)
	require.NoError(t, err)
}


func TestDrainEmptyReader(t *testing.T) {
	// when
	err := iosafety.DrainReader(nil)

	// then
	require.NoError(t, err)
}
