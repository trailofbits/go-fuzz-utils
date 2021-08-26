package go_fuzz_utils

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"unsafe"
)

type TypeProvider struct {
	data []byte
	position int
	randomProvider *rand.Rand // initialized after seed is obtained from first few bytes of data

	// Fill-related fields
	SliceMinSize int
	SliceMaxSize int
	SliceNilBias float32
	MapMinSize   int
	MapMaxSize int
	MapNilBias float32
	StringMinLength int
	StringMaxLength int
	DepthLimit int // zero indicates infinite depth
	FillUnexportedFields bool
}

func NewTypeProvider(data []byte) (*TypeProvider, error) {
	// Create a new type provider from the provided data and default settings
	t := &TypeProvider{
		data:                 data,
		SliceMinSize:         0,
		SliceMaxSize:         15,
		SliceNilBias:         0.05,
		MapMinSize:           0,
		MapMaxSize:           15,
		MapNilBias:           0.05,
		StringMinLength:      0,
		StringMaxLength:      15,
		DepthLimit:           0,
		FillUnexportedFields: true,
	}

	// Call reset to create our random provider from this data.
	err := t.Reset()
	if err != nil {
		return nil, err
	}

	return t, nil
}

// validateBounds checks if the remaining data in the buffer can satisfy an expected amount of bytes to be read.
// Returns an error if the provided number of bytes left at the current position cannot satisfy the expected count.
func (t *TypeProvider) validateBounds(expectedCount int) error {
	// If our expected count of bytes to read is negative, return an error as the caller likely had an arithmetic issue.
	if expectedCount < 0 {
		return fmt.Errorf("attempted to read a negative amount of bytes: %d", expectedCount)
	}

	// If our position is out of bounds, return an error.
	if t.position < 0 || len(t.data) < t.position {
		return fmt.Errorf("position out of bounds: (position: %d / length: %d)", t.position, len(t.data))
	}

	// If there aren't enough bytes left, return an error.
	bytesLeft := len(t.data) - t.position
	if bytesLeft < expectedCount {
		return fmt.Errorf("end of stream reached: could not read %d bytes (position: %d / length: %d)", expectedCount, t.position, len(t.data))
	}

	// Return no error
	return nil
}

// validateFillSettings checks if the fill settings provided in the TypeProvider are valid.
// Returns an error if the TypeProvider's fill settings are invalid.
func (t *TypeProvider) validateFillSettings() error {
	// Validate our min and max values
	if t.SliceMinSize < 0 || t.SliceMaxSize < 0 || t.SliceMinSize > t.SliceMaxSize {
		return errors.New("fill settings for slice size represent an invalid range")
	}
	if t.StringMinLength < 0 || t.StringMaxLength < 0 || t.StringMinLength > t.StringMaxLength {
		return errors.New("fill settings for string length represent an invalid range")
	}
	if t.MapMinSize < 0 || t.MapMaxSize < 0 || t.MapMinSize > t.MapMaxSize {
		return errors.New("fill settings for map size represent an invalid range")
	}
	if t.SliceNilBias < 0 || t.SliceNilBias > 1 {
		return errors.New("fill setting for slice nil bias is invalid. it must be between 0 and 1")
	}
	if t.MapNilBias < 0 || t.MapNilBias > 1 {
		return errors.New("fill setting for map nil bias is invalid. it must be between 0 and 1")
	}
	if t.DepthLimit < 0 {
		return errors.New("fill setting for depth limit cannot be less than zero")
	}
	return nil
}

// getRandomSize obtains a random int in the positive int range.
func (t *TypeProvider) getRandomSize(min int, max int) int {
	// Obtain a random size.
	return t.randomProvider.Intn((max - min) + 1)  + min
}

// getRandomBool obtains a random boolean given a probability between 0 and 1.
func (t *TypeProvider) getRandomBool(probability float32) bool {
	return t.randomProvider.Float32() < probability
}

// Reset resets the position to extract data from in the stream and reconstructs the random provider with the seed
// read from the first few bytes. This puts the TypeProvider in the same state as when it was created, unless the
// underlying TypeProviderConfig was changed.
func (t *TypeProvider) Reset() error {
	// Set the position to zero.
	t.position = 0
	t.randomProvider = nil

	// Read our random seed from the first int64
	seed, err := t.GetInt64()
	if err != nil {
		return err
	}

	// Create our random provider from the seed.
	t.randomProvider = rand.New(rand.NewSource(seed))
	return nil
}

// GetNBytes obtains the requested number of bytes from the current position in the buffer.
// This advances the position the provided length.
// Returns the requested bytes, or an error if the end of stream has been reached.
func (t *TypeProvider) GetNBytes(length int) ([]byte, error) {
	// Validate our boundaries
	err := t.validateBounds(length)
	if err != nil {
		return nil, err
	}

	// Obtain a slice of our data, advance position, and return the data.
	b := t.data[t.position:t.position + length]
	t.position += length
	return b, nil
}

// GetByte obtains a single byte from the current position in the buffer.
// This advances the position by 1.
// Returns the single read byte, or an error if the end of stream has been reached.
func (t *TypeProvider) GetByte() (byte, error) {
	// Validate our boundaries
	err := t.validateBounds(1)
	if err != nil {
		return 0, err
	}

	// Obtain our single byte, advance position, and return the data.
	b := t.data[t.position]
	t.position += 1
	return b, nil
}

// GetBool obtains a bool from the current position in the buffer.
// This advances the position by 1.
// Returns the read bool, or an error if the end of stream has been reached.
func (t *TypeProvider) GetBool() (bool, error) {
	// Obtain a byte and return a bool depending on if its even or odd.
	b, err := t.GetByte()
	return b % 2 == 0, err
}

// GetUint8 obtains an uint8 from the current position in the buffer.
// This advances the position by 1.
// Returns the read uint8, or an error if the end of stream has been reached.
func (t *TypeProvider) GetUint8() (uint8, error) {
	// Obtain a byte and return it as the requested type.
	b, err := t.GetByte()
	return uint8(b), err
}

// GetInt8 obtains an int8 from the current position in the buffer.
// This advances the position by 1.
// Returns the read int8, or an error if the end of stream has been reached.
func (t *TypeProvider) GetInt8() (int8, error) {
	// Obtain a byte and return it as the requested type.
	b, err := t.GetByte()
	return int8(b), err
}

// GetUint16 obtains an uint16 from the current position in the buffer.
// This advances the position by 2.
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

// GetInt16 obtains an int16 from the current position in the buffer.
// This advances the position by 2.
// Returns the read int16, or an error if the end of stream has been reached.
func (t *TypeProvider) GetInt16() (int16, error) {
	// Obtain an uint16 and convert it to an int16
	x, err := t.GetUint16()
	return int16(x), err
}

// GetUint32 obtains an uint32 from the current position in the buffer.
// This advances the position by 4.
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

// GetInt32 obtains an int32 from the current position in the buffer.
// This advances the position by 4.
// Returns the read int32, or an error if the end of stream has been reached.
func (t *TypeProvider) GetInt32() (int32, error) {
	// Obtain an uint32 and convert it to an int32
	x, err := t.GetUint32()
	return int32(x), err
}

// GetUint64 obtains an uint64 from the current position in the buffer.
// This advances the position by 8.
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

// GetInt64 obtains an int64 from the current position in the buffer.
// This advances the position by 64.
// Returns the read int64, or an error if the end of stream has been reached.
func (t *TypeProvider) GetInt64() (int64, error) {
	// Obtain an uint64 and convert it to an int64
	x, err := t.GetUint64()
	return int64(x), err
}

// GetUint obtains an uint from the current position in the buffer.
// This advances the position by 8, reading an uint64 and casting it to the architecture-dependent width.
// Returns the read uint, or an error if the end of stream has been reached.
func (t *TypeProvider) GetUint() (uint, error) {
	// Obtain an uint64 and convert it to an uint
	x, err := t.GetUint64()
	return uint(x), err
}

// GetInt obtains an int from the current position in the buffer.
// This advances the position by 8, reading an int64 and casting it to the architecture-dependent width.
// Returns the read int, or an error if the end of stream has been reached.
func (t *TypeProvider) GetInt() (int, error) {
	// Obtain an uint64 and convert it to an int
	x, err := t.GetUint64()
	return int(x), err
}

// GetFloat32 obtains a float32 from the current position in the buffer.
// This advances the position by 4.
// Returns the read float32, or an error if the end of stream has been reached.
func (t *TypeProvider) GetFloat32() (float32, error) {
	// Obtain an uint32 and convert it to a float32
	x, err := t.GetUint32()
	return math.Float32frombits(x), err
}

// GetFloat64 obtains a float64 from the current position in the buffer.
// This advances the position by 8.
// Returns the read float64, or an error if the end of stream has been reached.
func (t *TypeProvider) GetFloat64() (float64, error) {
	// Obtain an uint64 and convert it to a float64
	x, err := t.GetUint64()
	return math.Float64frombits(x), err
}

// GetFixedString obtains a string of the requested length from the current position in the buffer.
// This advances the position the provided length.
// Returns a string of the requested length, or an error if the end of stream has been reached.
func (t *TypeProvider) GetFixedString(length int) (string, error) {
	// Obtain bytes to convert to a string.
	b, err := t.GetNBytes(length)
	if err != nil {
		return "", err
	}

	// Return a string from the bytes
	return string(b), nil
}

// GetBytes obtains a number of bytes of length within the range settings provided in the TypeProvider.
// This advances the position by len(result)
// Returns the read bytes, or an error if the end of stream has been reached.
func (t *TypeProvider) GetBytes() ([]byte, error) {
	// Obtain a random size to read
	x := t.getRandomSize(t.SliceMinSize, t.SliceMaxSize)

	// Use the random size to determine how many bytes to read, then obtain them and return.
	return t.GetNBytes(x)
}

// GetString obtains a string of length within the range settings provided in the TypeProvider.
// This advances the position by len(result)
// Returns the read string, or an error if the end of stream has been reached.
func (t *TypeProvider) GetString() (string, error) {
	// Obtain a random size to read
	x := t.getRandomSize(t.StringMinLength, t.StringMaxLength)

	// Use the random to determine how many bytes to read, then obtain them and return.
	b, err := t.GetNBytes(x)
	if err != nil {
		return "", err
	}

	return string(b), err
}

// Fill populates data into a variable at a provided pointer. This can be used for structs or basic types.
// Returns an error if one is encountered.
func (t *TypeProvider) Fill(i interface{}) error {
	// Validate fill settings
	err := t.validateFillSettings()
	if err != nil {
		return err
	}

	// We should have been provided a pointer, so we obtain reflect pkg values and dereference.
	v := reflect.Indirect(reflect.ValueOf(i))

	// Next we fill the value.
	return t.fillValue(v, 0)
}

// fillValue populates data into a variable based on reflection. Given the provided parameters, structures and simple
// types can be recursively populated. See documentation surrounding the Fill method for more details.
// Returns an error if one is encountered.
func (t *TypeProvider) fillValue(v reflect.Value, currentDepth int) error {
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
	} else if v.Kind() == reflect.Complex64 {
		f, err := t.GetFloat32()
		if err != nil {
			return err
		}
		f2, err := t.GetFloat32()
		if err != nil {
			return err
		}
		v.SetComplex(complex128(complex(f, f2)))
	} else if v.Kind() == reflect.Complex128 {
		f, err := t.GetFloat64()
		if err != nil {
			return err
		}
		f2, err := t.GetFloat64()
		if err != nil {
			return err
		}
		v.SetComplex(complex(f, f2))
	}else if v.Kind() == reflect.String {
		s, err := t.GetString()
		if err != nil {
			return err
		}
		v.SetString(s)
	} else if v.Kind() == reflect.Slice && !t.getRandomBool(t.SliceNilBias) {
		// Obtain a random size
		sliceSize := t.getRandomSize(t.SliceMinSize, t.SliceMaxSize)

		// Typically, we just create a slice here and loop for each element and fill it. But we add a special case here
		// for byte arrays, as they're very common. Setting each element individually will take too long, so we read
		// a slice of bytes and set them all at once if we can detect the type is a []byte
		sliceElementType := v.Type().Elem()
		if sliceElementType.Kind() == reflect.Uint8 {
			b, err := t.GetNBytes(sliceSize)
			if err != nil {
				return err
			}
			v.SetBytes(b)
		} else {
			// If this isn't a byte array, create a generic slice of the correct type and fill it.
			slice := reflect.MakeSlice(v.Type(), sliceSize, sliceSize)
			for i := 0; i < sliceSize; i++ {
				err := t.fillValue(slice.Index(i), currentDepth)
				if err != nil {
					return err
				}
			}
			// Set our slice value
			v.Set(slice)
		}
	} else if v.Kind() == reflect.Map && !t.getRandomBool(t.MapNilBias) {
		// Obtain a random size
		mapSize := t.getRandomSize(t.MapMinSize, t.MapMaxSize)

		// Create our map and set it now, so we can proceed to create key-value pairs for it.
		v.Set(reflect.MakeMap(v.Type()))

		// Loop for each element we wish to create
		for i := 0; i < mapSize; i++ {
			// First we need to create our key, depending on the key type
			mKey := reflect.New(v.Type().Key()).Elem()
			mValue := reflect.New(v.Type().Elem()).Elem()

			// Populate the key and value
			err := t.fillValue(mKey, currentDepth)
			if err != nil {
				return err
			}
			err = t.fillValue(mValue, currentDepth)
			if err != nil {
				return err
			}

			// Set the key-value pair in our dictionary
			v.SetMapIndex(mKey, mValue)
		}
	} else if v.Kind() == reflect.Ptr {
		// If it's a pointer, we need to create a new underlying type to live at the pointer, then populate it.
		v.Set(reflect.New(v.Type().Elem()))
		err := t.fillValue(v.Elem(), currentDepth)
		if err != nil {
			return err
		}
	} else if v.Kind() == reflect.Array && !t.getRandomBool(t.SliceNilBias) {
		// Loop through each element and fill it recursively.
		for i := 0; i < v.Len(); i++ {
			err := t.fillValue(v.Index(i), currentDepth)
			if err != nil {
				return err
			}
		}
	} else if v.Kind() == reflect.Struct && (t.DepthLimit == 0 || t.DepthLimit > currentDepth) {
		// For structs we need to recursively populate every field
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)

			// If it's private and we're not setting private fields, skip it
			if !field.CanSet() {
				if !t.FillUnexportedFields {
					continue
				}
				// If we are filling private fields, we continue by creating a new one here.
				// Reference: https://stackoverflow.com/questions/42664837/how-to-access-unexported-struct-fields
				field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
			}

			// Now we're ready to set our data, so fill it accordingly.
			err := t.fillValue(field, currentDepth + 1)
			if err != nil {
				return err
			}
		}
	}

	// Unknown value types are simply skipped/ignored, so we continue to fuzz what we're able to.
	return nil
}
