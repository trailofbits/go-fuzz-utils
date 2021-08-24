package go_fuzz_utils

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"unsafe"
)

type TypeProvider struct {
	data []byte
	Position uint
}

func NewTypeProvider(data []byte) *TypeProvider {
	// Create a new type provider from the provided data.
	t := &TypeProvider{data: data}
	return t
}

// validateBounds checks if the remaining data in the buffer can satisfy an expected amount of bytes to be read.
// Returns an error if the provided number of bytes left at the current Position cannot satisfy the expected count.
func (t *TypeProvider) validateBounds(expectedCount uint) error {
	// If our position is out of bounds, return an error.
	length := uint(len(t.data))
	if length < t.Position {
		return errors.New(fmt.Sprintf("position out of bounds: (position: %d / length: %d)", t.Position, length))
	}

	// If there aren't enough bytes left, return an error.
	bytesLeft := length - t.Position
	if bytesLeft < expectedCount {
		return errors.New(fmt.Sprintf("end of stream reached: could not read %d bytes (position: %d / length: %d)", expectedCount, t.Position, length))
	}

	// Return no error
	return nil
}

// GetNBytes obtains the requested number of bytes from the current Position in the buffer.
// This advances the Position the provided length.
// Returns the requested bytes, or an error if the end of stream has been reached.
func (t *TypeProvider) GetNBytes(length uint) ([]byte, error) {
	// Validate our boundaries
	err := t.validateBounds(length)
	if err != nil {
		return nil, err
	}

	// Obtain a slice of our data, advance position, and return the data.
	b := t.data[t.Position:t.Position + length]
	t.Position += length
	return b, nil
}

// GetByte obtains a single byte from the current Position in the buffer.
// This advances the Position by 1.
// Returns the single read byte, or an error if the end of stream has been reached.
func (t *TypeProvider) GetByte() (byte, error) {
	// Validate our boundaries
	err := t.validateBounds(1)
	if err != nil {
		return 0, err
	}

	// Obtain our single byte, advance position, and return the data.
	b := t.data[t.Position]
	t.Position += 1
	return b, nil
}

// GetBool obtains a bool from the current Position in the buffer.
// This advances the Position by 1.
// Returns the read bool, or an error if the end of stream has been reached.
func (t *TypeProvider) GetBool() (bool, error) {
	// Obtain a byte and return a bool depending on if its even or odd.
	b, err := t.GetByte()
	return b % 2 == 0, err
}

// GetUint8 obtains an uint8 from the current Position in the buffer.
// This advances the Position by 1.
// Returns the read uint8, or an error if the end of stream has been reached.
func (t *TypeProvider) GetUint8() (uint8, error) {
	// Obtain a byte and return it as the requested type.
	b, err := t.GetByte()
	return uint8(b), err
}

// GetInt8 obtains an int8 from the current Position in the buffer.
// This advances the Position by 1.
// Returns the read int8, or an error if the end of stream has been reached.
func (t *TypeProvider) GetInt8() (int8, error) {
	// Obtain a byte and return it as the requested type.
	b, err := t.GetByte()
	return int8(b), err
}

// GetUint16 obtains an uint16 from the current Position in the buffer.
// This advances the Position by 2.
// Returns the read uint16, or an error if the end of stream has been reached.
func (t *TypeProvider) GetUint16() (uint16, error) {
	// Obtain the data to back our value
	b, err := t.GetNBytes(2)
	if err != nil {
		return 0, err
	}

	// Convert our data to an uint16 and return
	return binary.BigEndian.Uint16(b), nil
}

// GetInt16 obtains an int16 from the current Position in the buffer.
// This advances the Position by 2.
// Returns the read int16, or an error if the end of stream has been reached.
func (t *TypeProvider) GetInt16() (int16, error) {
	// Obtain an uint16 and convert it to an int16
	x, err := t.GetUint16()
	return int16(x), err
}

// GetUint32 obtains an uint32 from the current Position in the buffer.
// This advances the Position by 4.
// Returns the read uint32, or an error if the end of stream has been reached.
func (t *TypeProvider) GetUint32() (uint32, error) {
	// Obtain the data to back our value
	b, err := t.GetNBytes(4)
	if err != nil {
		return 0, err
	}

	// Convert our data to an uint32 and return
	return binary.BigEndian.Uint32(b), nil
}

// GetInt32 obtains an int32 from the current Position in the buffer.
// This advances the Position by 4.
// Returns the read int32, or an error if the end of stream has been reached.
func (t *TypeProvider) GetInt32() (int32, error) {
	// Obtain an uint32 and convert it to an int32
	x, err := t.GetUint32()
	return int32(x), err
}

// GetUint64 obtains an uint64 from the current Position in the buffer.
// This advances the Position by 8.
// Returns the read uint64, or an error if the end of stream has been reached.
func (t *TypeProvider) GetUint64() (uint64, error) {
	// Obtain the data to back our value
	b, err := t.GetNBytes(8)
	if err != nil {
		return 0, err
	}

	// Convert our data to an uint64 and return
	return binary.BigEndian.Uint64(b), nil
}

// GetInt64 obtains an int64 from the current Position in the buffer.
// This advances the Position by 64.
// Returns the read int64, or an error if the end of stream has been reached.
func (t *TypeProvider) GetInt64() (int64, error) {
	// Obtain an uint64 and convert it to an int64
	x, err := t.GetUint64()
	return int64(x), err
}

// GetUint obtains an uint from the current Position in the buffer.
// This advances the Position by 8, reading an uint64 and casting it to the architecture-dependent width.
// Returns the read uint, or an error if the end of stream has been reached.
func (t *TypeProvider) GetUint() (uint, error) {
	// Obtain an uint64 and convert it to an uint
	x, err := t.GetUint64()
	return uint(x), err
}

// GetInt obtains an int from the current Position in the buffer.
// This advances the Position by 8, reading an int64 and casting it to the architecture-dependent width.
// Returns the read int, or an error if the end of stream has been reached.
func (t *TypeProvider) GetInt() (int, error) {
	// Obtain an uint64 and convert it to an int
	x, err := t.GetUint64()
	return int(x), err
}

// GetFloat32 obtains a float32 from the current Position in the buffer.
// This advances the Position by 4.
// Returns the read float32, or an error if the end of stream has been reached.
func (t *TypeProvider) GetFloat32() (float32, error) {
	// Obtain an uint32 and convert it to a float32
	x, err := t.GetUint32()
	return math.Float32frombits(x), err
}

// GetFloat64 obtains a float64 from the current Position in the buffer.
// This advances the Position by 8.
// Returns the read float64, or an error if the end of stream has been reached.
func (t *TypeProvider) GetFloat64() (float64, error) {
	// Obtain an uint64 and convert it to a float64
	x, err := t.GetUint64()
	return math.Float64frombits(x), err
}

// GetFixedString obtains a string of the requested length from the current Position in the buffer.
// This advances the Position the provided length.
// Returns a string of the requested length, or an error if the end of stream has been reached.
func (t *TypeProvider) GetFixedString(length uint) (string, error) {
	// Obtain bytes to convert to a string.
	b, err := t.GetNBytes(length)
	if err != nil {
		return "", err
	}

	// Return a string from the bytes
	return string(b), nil
}

// GetBytes obtains a number of bytes no more than the provided maxLength from the buffer.
// This advances the Position by 4 + len(result), as a 32-bit unsigned integer is read to determine the buffer size
// to subsequently read.
// Returns the read bytes, or an error if the end of stream has been reached.
func (t *TypeProvider) GetBytes(maxLength uint) ([]byte, error) {
	// Obtain an uint32 which will represent the length we will read.
	x, err := t.GetUint32()
	if err != nil {
		return nil, err
	}

	// If a max length of zero is provided, it is a special case indicating we can read to the end of the data.
	if maxLength == 0 {
		maxLength = uint(len(t.data)) - t.Position
	}

	// Use the previously read uint32 to determine how many bytes to read, then obtain them and return.
	return t.GetNBytes(uint(x) % (maxLength + 1))
}

// GetString obtains a string of length no more than the provided maxLength from the buffer.
// This advances the Position by 4 + len(result), as a 32-bit unsigned integer is read to determine the buffer size
// to subsequently read.
// Returns the read string, or an error if the end of stream has been reached.
func (t *TypeProvider) GetString(maxLength uint) (string, error) {
	// Obtain a byte array of random length and convert it to a string.
	b, err := t.GetBytes(maxLength)
	if err != nil {
		return "", err
	}
	return string(b), err
}

// Fill populates data into a variable at a provided pointer. This can be used for structs or basic types.
// String lengths are bounded by maxStringLength, while array lengths are bounded by maxArrayLength.
// When filling a structure, the structDepthLimit defines how deep nested structures can be filled. A depth of one will
// only fill the immediately provided structure and not any nested structures. A depth of zero can be provided for
// unlimited depth.
// Returns the read string, or an error if encountered.
func (t *TypeProvider) Fill(i interface{}, maxStringLength uint, maxArrayLength uint, structDepthLimit uint, fillPrivateFields bool) error {
	// If we are given a depth limit of zero, it is a special case where we allow infinite depth.
	if structDepthLimit == 0 {
		structDepthLimit = ^uint(0)
	}

	// We should have been provided a pointer, so we obtain reflect pkg values and dereference.
	v := reflect.Indirect(reflect.ValueOf(i))

	// Next we fill the value.
	return t.fillValue(v, maxStringLength, maxArrayLength, structDepthLimit, fillPrivateFields)
}

// fillValue populates data into a variable based on reflection. Given the provided parameters, structures and simple
// types can be recursively populated. See documentation surrounding the Fill method for more details.
// Returns an error if one is encountered.
func (t *TypeProvider) fillValue(v reflect.Value, maxStringLength uint, maxArrayLength uint, structDepthLimit uint, fillPrivateFields bool) error {
	// If we can't set the value, we can stop immediately.
	if !v.CanSet() {
		return nil
	}

	// Determine how to set our value based on its type.
	if v.Kind() == reflect.Bool {
		bl, err := t.GetBool()
		if err != nil {
			return err
		}
		v.SetBool(bl)
	} else if v.Kind() == reflect.Int8 {
		i8, err := t.GetInt8()
		if err != nil {
			return err
		}
		v.SetInt(int64(i8))
	} else if v.Kind() == reflect.Uint8 {
		u8, err := t.GetUint8()
		if err != nil {
			return err
		}
		v.SetUint(uint64(u8))
	}  else if v.Kind() == reflect.Int16 {
		i16, err := t.GetInt16()
		if err != nil {
			return err
		}
		v.SetInt(int64(i16))
	} else if v.Kind() == reflect.Uint16 {
		u16, err := t.GetUint16()
		if err != nil {
			return err
		}
		v.SetUint(uint64(u16))
	} else if v.Kind() == reflect.Int32 {
		i32, err := t.GetInt32()
		if err != nil {
			return err
		}
		v.SetInt(int64(i32))
	} else if v.Kind() == reflect.Uint32 {
		u32, err := t.GetUint32()
		if err != nil {
			return err
		}
		v.SetUint(uint64(u32))
	} else if v.Kind() == reflect.Int64 {
		i64, err := t.GetInt64()
		if err != nil {
			return err
		}
		v.SetInt(i64)
	} else if v.Kind() == reflect.Uint64 {
		u64, err := t.GetUint64()
		if err != nil {
			return err
		}
		v.SetUint(u64)
	} else if v.Kind() == reflect.Int {
		i, err := t.GetInt()
		if err != nil {
			return err
		}
		v.SetInt(int64(i))
	} else if v.Kind() == reflect.Uint {
		u, err := t.GetUint()
		if err != nil {
			return err
		}
		v.SetUint(uint64(u))
	} else if v.Kind() == reflect.Float32 {
		f32, err := t.GetFloat32()
		if err != nil {
			return err
		}
		v.SetFloat(float64(f32))
	} else if v.Kind() == reflect.Float64 {
		f64, err := t.GetFloat64()
		if err != nil {
			return err
		}
		v.SetFloat(f64)
	} else if v.Kind() == reflect.String {
		s, err := t.GetString(maxStringLength)
		if err != nil {
			return err
		}
		v.SetString(s)
	} else if v.Kind() == reflect.Slice {
		// Read an uint32 for the size of the slice we will create
		// (we modulo divide by max bytes + 1 to get a random value in our range)
		x, err := t.GetUint32()
		if err != nil {
			return err
		}
		sliceSize := int(uint(x) % (maxArrayLength + 1))

		// Create a slice and recursively populate it, as its type may be complex.
		slice := reflect.MakeSlice(v.Type(), sliceSize, sliceSize)
		for i := 0; i < sliceSize; i++ {
			err = t.fillValue(slice.Index(i), maxStringLength, maxArrayLength, structDepthLimit, fillPrivateFields)
			if err != nil {
				return err
			}
		}

		// Set our slice value
		v.Set(slice)
	} else if v.Kind() == reflect.Map {
		// Read an uint32 for the size of the slice we will create
		// (we modulo divide by max bytes + 1 to get a random value in our range)
		x, err := t.GetUint32()
		if err != nil {
			return err
		}
		mapSize := int(uint(x) % (maxArrayLength + 1))

		// Create our map and set it now, so we can proceed to create key-value pairs for it.
		v.Set(reflect.MakeMap(v.Type()))

		// Loop for each element we wish to create
		for i := 0; i < mapSize; i++ {
			// First we need to create our key, depending on the key type
			mKey := reflect.New(v.Type().Key()).Elem()
			mValue := reflect.New(v.Type().Elem()).Elem()

			// Populate the key and value
			err = t.fillValue(mKey, maxStringLength, maxArrayLength, structDepthLimit, fillPrivateFields)
			if err != nil {
				return err
			}
			err = t.fillValue(mValue, maxStringLength, maxArrayLength, structDepthLimit, fillPrivateFields)
			if err != nil {
				return err
			}

			// Set the key-value pair in our dictionary
			v.SetMapIndex(mKey, mValue)
		}
	} else if v.Kind() == reflect.Ptr {
		// If it's a pointer, we need to create a new underlying type to live at the pointer, then populate it.
		v.Set(reflect.New(v.Type().Elem()))
		err := t.fillValue(v.Elem(), maxStringLength, maxArrayLength, structDepthLimit, fillPrivateFields)
		if err != nil {
			return err
		}
	} else if v.Kind() == reflect.Struct && structDepthLimit != 0 {
		// For structs we need to recursively populate every field
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)

			// If it's private and we're not setting private fields, skip it
			if !field.CanSet() {
				if !fillPrivateFields {
					continue
				}
				// If we are filling private fields, we continue by creating a new one here.
				// Reference: https://stackoverflow.com/questions/42664837/how-to-access-unexported-struct-fields
				field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
			}

			// Now we're ready to set our data, so fill it accordingly.
			err := t.fillValue(field, maxStringLength, maxArrayLength, structDepthLimit - 1, fillPrivateFields)
			if err != nil {
				return err
			}
		}
	}

	// Unknown value types are simply skipped/ignored so we can fuzz what we're able to.
	return nil
}
