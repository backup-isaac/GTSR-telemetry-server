package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	// Generate all required test input files.
	generateTestInput()
	generateCompleteCRCTestInput()
	generateIncompleteCRCTestInput()
	fmt.Print("> Setup completed\n")
}

func teardown() {
	fmt.Print("> Teardown completed\n")
}

// TestReadWriteBytes checks whether readWriteBytes writes the bytes from test_input.bin to test_output.bin correctly.
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
	// Reset readers
	testInput.Close()
	testOutput.Close()
	testInput, err = os.Open("test_input.bin")
	assert.NoError(t, err)
	defer testInput.Close()
	testOutput, err = os.Open("test_output.bin")
	assert.NoError(t, err)
	defer testOutput.Close()
	inBytes, err := ioutil.ReadAll(testInput)
	assert.NoError(t, err)
	outBytes, err := ioutil.ReadAll(testOutput)
	assert.NoError(t, err)
	assert.EqualValues(t, inBytes, outBytes)
}

// TestReadWriteBytesCompleteCRC checks whether readWriteBytesCRC strips the CRC checksum from the frames in
// test_input_complete_crc.bin such that test_output.bin = test_input.bin.
func TestReadWriteBytesCompleteCRC(t *testing.T) {
	testInput, err := os.Open("test_input_complete_crc.bin")
	assert.NoError(t, err)
	defer testInput.Close()
	testOutput, err := os.Create("test_output.bin")
	assert.NoError(t, err)
	defer testOutput.Close()
	err = readWriteBytesCRC(testInput, testOutput)
	if err != io.EOF {
		assert.Fail(t, err.Error())
	}
	expectedOutput, err := os.Open("test_input.bin")
	assert.NoError(t, err)
	defer expectedOutput.Close()
	// Reset reader
	testOutput.Close()
	testOutput, err = os.Open("test_output.bin")
	assert.NoError(t, err)
	defer testOutput.Close()
	expectedBytes, err := ioutil.ReadAll(expectedOutput)
	assert.NoError(t, err)
	actualBytes, err := ioutil.ReadAll(testOutput)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedBytes, actualBytes)
}

// TestReadWriteBytesIncompleteCRC checks whether readWriteBytesCRC throws out incomplete frames from its datastream
// as expected.
func TestReadWriteBytesIncompleteCRC(t *testing.T) {
	testInput, err := os.Open("test_input_incomplete_crc.bin")
	assert.NoError(t, err)
	defer testInput.Close()
	testOutput, err := os.Create("test_output.bin")
	assert.NoError(t, err)
	defer testOutput.Close()
	err = readWriteBytesCRC(testInput, testOutput)
	if err != io.EOF {
		assert.Fail(t, err.Error())
	}
	expectedOutput, err := os.Open("test_expected_output_incomplete_crc.bin")
	assert.NoError(t, err)
	defer expectedOutput.Close()
	// Reset reader
	testOutput.Close()
	testOutput, err = os.Open("test_output.bin")
	assert.NoError(t, err)
	defer testOutput.Close()
	expectedBytes, err := ioutil.ReadAll(expectedOutput)
	assert.NoError(t, err)
	actualBytes, err := ioutil.ReadAll(testOutput)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedBytes, actualBytes)
}

////////////////////////////////////////////////////////
// Test input generation functions, called in setup() //
////////////////////////////////////////////////////////

// generateTestInput creates a file with sample frames, following the format used by the solar car.
func generateTestInput() {
	outFile, err := os.Create("test_input.bin")
	if err != nil {
		log.Fatalf("Creation of test input failed: %s", err.Error())
	}
	defer outFile.Close()
	GTAndCANID := make([]byte, 4)
	binary.BigEndian.PutUint32(GTAndCANID, 0x4754FFFF)
	var floatValue uint16 = 0xC842
	valBuf := make([]byte, 2)
	zeroBuf := make([]byte, 6)
	for i := 0; i < 100; i++ {
		binary.BigEndian.PutUint16(valBuf, floatValue)
		buf := make([]byte, 0)
		buf = append(buf, GTAndCANID...)
		buf = append(buf, zeroBuf...)
		buf = append(buf, valBuf...)
		_, err = outFile.Write(buf)
		if err != nil {
			log.Fatalf("Write to test input failed: %s", err.Error())
		}
		floatValue -= 512
	}
}

// generateCompleteCRCTestInput creates a file with sample frames and appended CRC checksums.
func generateCompleteCRCTestInput() {
	table := crc32.MakeTable(crc32.IEEE)
	outFile, err := os.Create("test_input_complete_crc.bin")
	if err != nil {
		log.Fatalf("Creation of complete CRC test input failed: %s", err.Error())
	}
	defer outFile.Close()
	GTAndCANID := make([]byte, 4)
	binary.BigEndian.PutUint32(GTAndCANID, 0x4754FFFF)
	var floatValue uint16 = 0xC842
	zeroBuf := make([]byte, 6)
	valBuf := make([]byte, 2)
	checksumBuf := make([]byte, 4)
	for i := 0; i < 100; i++ {
		binary.BigEndian.PutUint16(valBuf, floatValue)
		buf := make([]byte, 0)
		buf = append(buf, GTAndCANID...)
		buf = append(buf, zeroBuf...)
		buf = append(buf, valBuf...)
		checksum := crc32.Checksum(buf, table)
		binary.LittleEndian.PutUint32(checksumBuf, checksum)
		buf = append(buf, checksumBuf...)
		_, err = outFile.Write(buf)
		if err != nil {
			log.Fatalf("Write to complete CRC test input failed: %s", err.Error())
		}
		floatValue -= 512
	}
}

// Mock corruption of data by randomly removing a byte from the frame every now and then
func generateIncompleteCRCTestInput() {
	table := crc32.MakeTable(crc32.IEEE)
	outFile, err := os.Create("test_input_incomplete_crc.bin")
	if err != nil {
		log.Fatalf("Creation of incomplete CRC test input failed: %s", err.Error())
	}
	defer outFile.Close()
	postParseFile, err := os.Create("test_expected_output_incomplete_crc.bin")
	if err != nil {
		log.Fatalf("Creation of incomplete CRC expected output failed: %s", err.Error())
	}
	defer outFile.Close()
	GTAndCANID := make([]byte, 4)
	binary.BigEndian.PutUint32(GTAndCANID, 0x4754FFFF)
	var floatValue uint16 = 0xC842
	zeroBuf := make([]byte, 6)
	valBuf := make([]byte, 2)
	checksumBuf := make([]byte, 4)
	// Seed rand with current time, so that test is unique
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Was the previous frame corrupted?
	wasCorrupted := false
	for i := 0; i < 100; i++ {
		binary.BigEndian.PutUint16(valBuf, floatValue)
		buf := make([]byte, 0)
		buf = append(buf, GTAndCANID...)
		buf = append(buf, zeroBuf...)
		buf = append(buf, valBuf...)
		checksum := crc32.Checksum(buf, table)
		binary.LittleEndian.PutUint32(checksumBuf, checksum)
		if !wasCorrupted {
			randNum := r.Intn(100)
			if randNum > 11 {
				// This file should only be written to if the frame is not thrown out by readWriteBytesCRC
				_, err = postParseFile.Write(buf)
				if err != nil {
					log.Fatalf("Write to incomplete CRC expected output failed: %s", err.Error())
				}
			} else {
				// The next frame is not thrown out if 'G' or 'T' is never seen
				if randNum > 1 {
					wasCorrupted = true
				}
				// Remove the byte in buf corresponding to randNum
				buf = append(buf[:randNum], buf[randNum+1:]...)
			}
		} else {
			wasCorrupted = false
		}
		buf = append(buf, checksumBuf...)
		_, err = outFile.Write(buf)
		if err != nil {
			log.Fatalf("Write to incomplete CRC test input failed: %s", err.Error())
		}
		floatValue -= 512
	}
}
