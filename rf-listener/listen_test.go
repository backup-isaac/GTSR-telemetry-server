package main

import (
	"encoding/binary"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"hash/crc32"

	"github.com/stretchr/testify/assert"
)

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
	inBytes, err := ioutil.ReadAll(testInput)
	assert.NoError(t, err)
	outBytes, err := ioutil.ReadAll(testOutput)
	assert.NoError(t, err)
	assert.EqualValues(t, inBytes, outBytes)
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
func TestReadWriteBytesCRCComplete(t *testing.T) {
	testInput, err := os.Open("test_files/test_input_crc_complete.bin")
	assert.NoError(t, err)
	defer testInput.Close()
	testOutput, err := os.Create("test_files/test_output_crc_complete.bin")
	assert.NoError(t, err)
	defer testOutput.Close()
	err = readWriteBytesCRC(testInput, testOutput)
	if err != io.EOF {
		assert.Fail(t, err.Error())
	}
	testOutput.Close()
	expectedOutput, err := os.Open("test_files/test_input.bin")
	assert.NoError(t, err)
	defer expectedOutput.Close()
	expectedBytes, err := ioutil.ReadAll(expectedOutput)
	assert.NoError(t, err)
	actualOutput, err := os.Open("test_files/test_output_crc_complete.bin")
	assert.NoError(t, err)
	defer actualOutput.Close()
	actualBytes, err := ioutil.ReadAll(actualOutput)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedBytes, actualBytes)
}

func TestGenerateCompleteCRCTestInput(t *testing.T) {
	inFile, err := os.Open("test_files/test_input.bin")
	assert.NoError(t, err)
	defer inFile.Close()
	buf := make([]byte, 12)
	table := crc32.MakeTable(crc32.Castagnoli)
	testOutput, err := os.Create("test_files/test_input_crc_complete.bin")
	if err != nil {
		log.Fatalf("Output file creation error: %s", err)
	}
	defer testOutput.Close()
	for {
		_, err = inFile.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("Read error: %s", err)
		}
		checksum := crc32.Checksum(buf, table)
		checksumBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(checksumBytes, checksum)
		print(checksumBytes)
		CRCBuf := append(buf, checksumBytes...)
		_, err = testOutput.Write(CRCBuf)
		if err != nil {
			log.Fatalf("Write error: %s", err)
		}
	}
}

func TestGenerateTestInput(t *testing.T) {
	outFile, err := os.Create("test_files/test_input.bin")
	assert.NoError(t, err)
	defer outFile.Close()
	var frame1 uint32 = 0x4754FFFF
	var frame2 uint16 = 0xC842
	zeroBuf := make([]byte, 6)
	for i := 0; i < 100; i++ {
		buf1 := make([]byte, 4)
		buf2 := make([]byte, 2)
		binary.BigEndian.PutUint32(buf1, frame1)
		binary.BigEndian.PutUint16(buf2, frame2)
		_, err = outFile.Write(buf1)
		if err != nil {
			assert.NoError(t, err)
		}
		_, err = outFile.Write(zeroBuf)
		if err != nil {
			assert.NoError(t, err)
		}
		_, err = outFile.Write(buf2)
		if err != nil {
			assert.NoError(t, err)
		}
		frame2 -= 512
	}
}
