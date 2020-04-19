package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	// Do something here.
	generateTestInput()
	generateCompleteCRCTestInput()
	fmt.Printf("\033[1;36m%s\033[0m", "> Setup completed\n")
}

func teardown() {
	// Do something here.

	fmt.Printf("\033[1;36m%s\033[0m", "> Teardown completed")
	fmt.Printf("\n")
}

// TestReadWriteBytes checks whether readWriteBytes writes the bytes from test_input.bin to test_output.bin correctly.
func TestReadWriteBytes(t *testing.T) {
	testInput, err := os.Open("test_files/test_input.bin")
	assert.NoError(t, err)
	defer testInput.Close()
	testOutput, err := os.Create("test_files/test_output.bin")
	assert.NoError(t, err)
	defer testOutput.Close()
	err = readWriteBytes(testInput, testOutput)
	if err != io.EOF {
		assert.Fail(t, err.Error())
	}
	// Reset readers
	testInput.Close()
	testOutput.Close()
	testInput, err = os.Open("test_files/test_input.bin")
	assert.NoError(t, err)
	defer testInput.Close()
	testOutput, err = os.Open("test_files/test_output.bin")
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
	testInput, err := os.Open("test_files/test_input_complete_crc.bin")
	assert.NoError(t, err)
	defer testInput.Close()
	testOutput, err := os.Create("test_files/test_output.bin")
	assert.NoError(t, err)
	defer testOutput.Close()
	err = readWriteBytesCRC(testInput, testOutput)
	if err != io.EOF {
		assert.Fail(t, err.Error())
	}
	expectedOutput, err := os.Open("test_files/test_input.bin")
	assert.NoError(t, err)
	defer expectedOutput.Close()
	// Reset reader
	testOutput.Close()
	testOutput, err = os.Open("test_files/test_output.bin")
	assert.NoError(t, err)
	defer testOutput.Close()
	expectedBytes, err := ioutil.ReadAll(expectedOutput)
	assert.NoError(t, err)
	actualBytes, err := ioutil.ReadAll(testOutput)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedBytes, actualBytes)
}

// // TestReadWriteBytesCRC checks whether readWriteBytes reads bytes from test_input_crc.bin and outputs only frames that pass CRC to test_output_crc.bin.
// func TestReadWriteBytesCRC(t *testing.T) {
// 	testInput, err := os.Open("test_files/test_input_crc.bin")
// 	assert.NoError(t, err)
// 	defer testInput.Close()
// 	testOutput, err := os.Create("test_files/test_output_crc.bin")
// 	assert.NoError(t, err)
// 	defer testOutput.Close()
// 	err = readWriteBytesCRC(testInput, testOutput)
// 	if err != io.EOF {
// 		assert.Fail(t, err.Error())
// 	}
// 	expectedOutput, err := os.Open("expected_output_crc.bin")
// 	inBytes, err := ioutil.ReadAll(expectedOutput)
// 	assert.NoError(t, err)
// 	outBytes, err := ioutil.ReadAll(testOutput)
// 	assert.NoError(t, err)
// 	assert.EqualValues(t, inBytes, outBytes)
// }

// TestReadWriteBytesCRCComplete checks whether readWriteBytesCRC reads bytes from test_input_crc_complete.bin and outputs only frames that pass CRC to test_output_crc_complete.bin.
// func TestReadWriteBytesCRCComplete(t *testing.T) {
// 	testInput, err := os.Open("test_files/test_input_crc_complete.bin")
// 	assert.NoError(t, err)
// 	defer testInput.Close()
// 	testOutput, err := os.Create("test_files/test_output_crc_complete.bin")
// 	assert.NoError(t, err)
// 	defer testOutput.Close()
// 	err = readWriteBytesCRC(testInput, testOutput)
// 	if err != io.EOF {
// 		assert.Fail(t, err.Error())
// 	}
// 	testOutput.Close()
// 	expectedOutput, err := os.Open("test_files/test_input.bin")
// 	assert.NoError(t, err)
// 	defer expectedOutput.Close()
// 	expectedBytes, err := ioutil.ReadAll(expectedOutput)
// 	assert.NoError(t, err)
// 	actualOutput, err := os.Open("test_files/test_output_crc_complete.bin")
// 	assert.NoError(t, err)
// 	defer actualOutput.Close()
// 	actualBytes, err := ioutil.ReadAll(actualOutput)
// 	assert.NoError(t, err)
// 	assert.EqualValues(t, expectedBytes, actualBytes)
// }

// Test input generation functions (called in setup)

// generateTestInput creates a file with sample frames, following the format used by the solar car.
func generateTestInput() {
	outFile, err := os.Create("test_files/test_input.bin")
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
			log.Fatalf("Creation of test input failed: %s", err.Error())
		}
		floatValue -= 512
	}
}

// generateCompleteCRCTestInput creates a file with sample frames and appended CRC checksums.
func generateCompleteCRCTestInput() {
	table := crc32.MakeTable(crc32.IEEE)
	outFile, err := os.Create("test_files/test_input_complete_crc.bin")
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
			log.Fatalf("Creation of complete CRC test input failed: %s", err.Error())
		}
		floatValue -= 512
	}
}
