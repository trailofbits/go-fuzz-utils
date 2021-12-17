//go:build linux

package go_fuzz_utils_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/trailofbits/go-fuzz-utils"
	"io/ioutil"
	"os"
	"testing"
)

func TestFileCreate(t *testing.T) {
	// Create our fuzz data
	b := GenerateTestData(0x1000)

	// Create our type provider
	tp, err := go_fuzz_utils.NewTypeProvider(b)
	assert.Nil(t, err)

	// Create a file and get the path to it
	filepath, err := tp.GetFilePath()

	// Ensure the filepath was generated
	assert.NotEmpty(t, filepath)

	// Open a file
	fileOpenTest, err := os.OpenFile(filepath, os.O_RDONLY, 0755)

	// Ensure the file exists
	assert.NotNil(t, fileOpenTest)

	// Ensure the file is not empty
	body, err := ioutil.ReadFile(filepath)
	assert.NotEmpty(t, body)

	// Reset our type provider
	err = tp.Reset()
	assert.Nil(t, err)

	// Create a file and get the *os.File
	tp, err = go_fuzz_utils.NewTypeProvider(b)
	assert.Nil(t, err)

	// Get a file
	file, err := tp.GetFile()

	// Ensure the file (*os.File) was generated
	assert.NotNil(t, file)

	// Ensure that the file is not empty
	fileinfo, err := file.Stat()
	assert.Nil(t, err)

	// Get a size of the file
	filesize := fileinfo.Size()
	assert.Greater(t, filesize, int64(0))
}

func TestPositionReachedEndFile(t *testing.T) {
	// Create our fuzz data
	b := GenerateTestData(1)

	// Create our type provider. We should encounter an error since we need at least 64-bits to read a random seed from.
	tp, err := go_fuzz_utils.NewTypeProvider(b)
	assert.NotNil(t, err)

	// Create more fuzz data
	b = GenerateTestData(9)

	// Recreate our type provider, this time it should succeed, reading 8 bytes as a random seed, leaving 1 byte left.
	tp, err = go_fuzz_utils.NewTypeProvider(b)
	assert.Nil(t, err)

	// Assert the values are as expected
	b1, err := tp.GetByte()
	assert.Nil(t, err)
	assert.EqualValues(t, 0xF7, b1)

	// Now expect errors reading any type
	_, err = tp.GetFile()
	assert.NotNil(t, err)

	_, err = tp.GetFilePath()
	assert.NotNil(t, err)
}
