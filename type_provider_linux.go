//go:build linux

package go_fuzz_utils

import (
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
)

// GetFilePath returns a path (/proc/self/fd/) to the anonymous file that is filled with
// bytes from the current position in the buffer
func (t *TypeProvider) GetFilePath() (string, error) {
	// Obtain a random size to read
	x := t.getRandomSize(t.sliceMinSize, t.sliceMaxSize)

	// Use the random to determine how many bytes to read, then obtain them and return
	b, err := t.GetNBytes(x)
	if err != nil {
		return "", err
	}

	// Create an anonymous file with the debug go-fuzz-utils-filepath name, fill with the bytes b
	// and obtain a file descriptor to it
	fd, err := memoryFile("go-fuzz-utils-filepath", b)
	if err != nil {
		return "", err
	}

	// Prepare a path to the specific file descriptor
	fp := fmt.Sprintf("/proc/self/fd/%d", fd)

	// Return a path to the created file
	return fp, err
}

// GetFile returns *os.File
func (t *TypeProvider) GetFile() (*os.File, error) {
	// Obtain a random size to read
	x := t.getRandomSize(t.sliceMinSize, t.sliceMaxSize)

	// Use the random to determine how many bytes to read, then obtain them and return
	b, err := t.GetNBytes(x)
	if err != nil {
		return nil, err
	}

	// Create an anonymous file with the debug go-fuzz-utils-file name, fill with the bytes b
	// and obtain a file descriptor to it
	fd, err := memoryFile("go-fuzz-utils-file", b)
	if err != nil {
		return nil, err
	}

	// Prepare a path to the specific file descriptor
	fp := fmt.Sprintf("/proc/self/fd/%d", fd)

	// NewFile returns a new File with the given file descriptor and name
	f := os.NewFile(uintptr(fd), fp)

	// Return a new File object
	return f, err
}

// memoryFile is a helper function that creates an anonymous file
// The file behaves like a regular file, but it lives in memory
// The name supplied in the filename parameter is used as a filename and is displayed
// as the target of the corresponding symbolic link in the /proc/self/fd/. The name is
// prefixed with the memfd: and serves only for debugging purposes. Names do not affect
// the behavior of the file descriptor, and as such multiple files can have the same name
// without any side effects
// Returns the file descriptor to an anonymous file
func memoryFile(filename string, b []byte) (int, error) {
	// Do not handle "empty" files
	if len(b) == 0 {
		return 0, errors.New("empty input to memoryFile")
	}

	// Calls the memfd_create syscall that creates an anonymous file and returns a file descriptor that refers to it
	fd, err := unix.MemfdCreate(filename, 0)
	if err != nil {
		return 0, err
	}

	// Calls the ftruncate syscall and truncates to a size of bytes referenced by a file descriptor
	err = unix.Ftruncate(fd, int64(len(b)))
	if err != nil {
		return 0, err
	}

	// Creates a new mapping in the virtual address space
	data, err := unix.Mmap(fd, 0, len(b), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		return 0, err
	}

	// Copy from a buffer b to the mapping
	copy(data, b)

	// Delete the mapping
	err = unix.Munmap(data)
	if err != nil {
		return 0, err
	}

	// Return a file descriptor
	return fd, nil
}
