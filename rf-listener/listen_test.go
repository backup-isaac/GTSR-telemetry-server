package main

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadWriteBytes(t *testing.T) {
	testInput, err := os.Open("test_input.bin")
	assert.NoError(t, err)
	defer testInput.Close()
	testOutput, err := os.Create("test_output.bin")
	assert.NoError(t, err)
	defer testOutput.Close()
	err = readWriteBytes(testInput, testOutput)
	if err != io.EOF {
		assert.Fail(t, err.Error())
	}
	inBytes, err := ioutil.ReadAll(testInput)
	assert.NoError(t, err)
	outBytes, err := ioutil.ReadAll(testOutput)
	assert.NoError(t, err)
	assert.EqualValues(t, inBytes, outBytes)
}
