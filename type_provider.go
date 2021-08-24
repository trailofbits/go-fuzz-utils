package go_fuzz_utils

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

type TypeProvider struct {
	data []byte
	position uint
}

func New(data []byte) *TypeProvider {
	// Create a new type provider from the provided data.
	t := &TypeProvider{data: data}
	return t
}

func (t *TypeProvider) ValidateBounds(expectedCount uint) error {
	// If our position is out of bounds, return an error.
	if uint(len(t.data)) < t.position {
		return errors.New(fmt.Sprintf("Position out of bounds (pos: %d)", t.position))
	}

	// If there aren't enough bytes left, return an error.
	bytesLeft := uint(len(t.data)) - t.position
	if bytesLeft < expectedCount {
		return errors.New(fmt.Sprintf("End of stream reached. Tried to read %d bytes but only had %d left.", expectedCount, bytesLeft))
	}

	// Return no error
	return nil
}

func (t *TypeProvider) GetByte() (byte, error) {
	// Validate our boundaries
	err := t.ValidateBounds(1)
	if err != nil {
		return 0, err
	}

	// Obtain a slice of our data, advance position, and return the data.
	b := t.data[t.position]
	t.position += 1
	return b, nil
}

func (t *TypeProvider) GetBool() (bool, error) {
	// Obtain a byte and return a bool depending on if its even or odd.
	b, err := t.GetByte()
	return b % 2 == 0, err
}

func (t *TypeProvider) GetNBytes(length uint) ([]byte, error) {
	// Validate our boundaries
	err := t.ValidateBounds(length)
	if err != nil {
		return nil, err
	}

	// Obtain a slice of our data, advance position, and return the data.
	b := t.data[t.position:t.position + length]
	t.position += length
	return b, nil
}

func (t *TypeProvider) GetFixedString(length uint) (string, error) {
	// Obtain bytes to convert to a string.
	b, err := t.GetNBytes(length)
	if err != nil {
		return "", err
	}

	// Return a string from the bytes
	return string(b), nil
}

func (t *TypeProvider) GetBytes(maxLength uint) ([]byte, error) {
	// Obtain an uint32 which will represent the length we will read.
	x, err := t.GetUint32()
	if err != nil {
		return nil, err
	}

	// If a max length of zero is provided, it is a special case indicating we can read to the end of the data.
	if maxLength == 0 {
		maxLength = uint(len(t.data)) - t.position
	}

	// Use the previously read uint32 to determine how many bytes to read, then obtain them and return.
	return t.GetNBytes(uint(x) % maxLength)
}

func (t *TypeProvider) GetString(maxLength uint) (string, error) {
	// Obtain a byte array of random length and convert it to a string.
	b, err := t.GetBytes(maxLength)
	if err != nil {
		return "", err
	} else {
		return string(b), err
	}
}

func (t *TypeProvider) GetUint8() (uint8, error) {
	// Obtain a byte and return it as the requested type.
	b, err := t.GetByte()
	return uint8(b), err
}

func (t *TypeProvider) GetInt8() (int8, error) {
	// Obtain a byte and return it as the requested type.
	b, err := t.GetByte()
	return int8(b), err
}

func (t *TypeProvider) GetUint16() (uint16, error) {
	// Obtain the data to back our value
	b, err := t.GetNBytes(2)
	if err != nil {
		return 0, err
	}

	// Convert our data to an uint16 and return
	return binary.BigEndian.Uint16(b), nil
}

func (t *TypeProvider) GetInt16() (int16, error) {
	// Obtain an uint16 and convert it to an int16
	x, err := t.GetUint16()
	return int16(x), err
}

func (t *TypeProvider) GetUint32() (uint32, error) {
	// Obtain the data to back our value
	b, err := t.GetNBytes(4)
	if err != nil {
		return 0, err
	}

	// Convert our data to an uint32 and return
	return binary.BigEndian.Uint32(b), nil
}

func (t *TypeProvider) GetInt32() (int32, error) {
	// Obtain an uint32 and convert it to an int32
	x, err := t.GetUint32()
	return int32(x), err
}

func (t *TypeProvider) GetUint64() (uint64, error) {
	// Obtain the data to back our value
	b, err := t.GetNBytes(8)
	if err != nil {
		return 0, err
	}

	// Convert our data to an uint64 and return
	return binary.BigEndian.Uint64(b), nil
}

func (t *TypeProvider) GetInt64() (int64, error) {
	// Obtain an uint64 and convert it to an int64
	x, err := t.GetUint64()
	return int64(x), err
}

func (t *TypeProvider) GetFloat32() (float32, error) {
	// Obtain an uint32 and convert it to a float32
	x, err := t.GetUint32()
	return math.Float32frombits(x), err
}

func (t *TypeProvider) GetFloat64() (float64, error) {
	// Obtain an uint64 and convert it to a float64
	x, err := t.GetUint64()
	return math.Float64frombits(x), err
}
