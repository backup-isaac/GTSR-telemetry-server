package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"testing"
)

func TestReadWriteBytes(t *testing.T) {
	testInput, err := os.Open("test_input.bin")
	if err != nil {
		log.Fatalf("Test input open error: %s", err)
	}
	defer testInput.Close()
	testOutput, err := os.Create("test_output.bin")
	if err != nil {
		log.Fatalf("Test output open error: %s", err)
	}
	defer testOutput.Close()
	err = readWriteBytes(testInput, testOutput)
	if err != io.EOF {
		log.Fatalf("Read/write failed: %s", err)
	}
	fmt.Printf("Output stored in %s", testOutput.Name())
}
