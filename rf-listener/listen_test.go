package main

import (
	"testing"

	"hash/crc32"
)

var host string
var serialPort string
var table = crc32.MakeTable(0x1EDC6F41)

func TestCRCFullFrames(t *testing.T) {

}
