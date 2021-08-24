package go_fuzz_utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimpleTypes(t *testing.T) {
	// Create our fuzz data
	b := make([]byte, 256)

	// Loop through and set our bytes accordingly. We do it in descending order so the values we test will have
	// negative bits set if they're signed integers.
	for i := 0; i < len(b); i++ {
		b[i] = 255 - byte(i)
	}

	// Create our type provider
	tp := New(b)

	// Assert the values are as expected reading from the above buffer.
	b1, err := tp.GetByte() // pos 0
	assert.Nil(t, err)
	assert.EqualValues(t, 0xFF, b1)

	i16, err := tp.GetInt16() // pos 1
	assert.Nil(t, err)
	assert.EqualValues(t, -259, i16)

	u16, err := tp.GetUint16() // pos 3
	assert.Nil(t, err)
	assert.EqualValues(t, 64763, u16)

	i32, err := tp.GetInt32() // pos 5
	assert.Nil(t, err)
	assert.EqualValues(t, -84281097, i32)

	u32, err := tp.GetUint32() // pos 9
	assert.Nil(t, err)
	assert.EqualValues(t, 4143314163, u32)

	i64, err := tp.GetInt64() // pos 13
	assert.Nil(t, err)
	assert.EqualValues(t, -940705933847302933, i64)

	u64, err := tp.GetUint64() // pos 21
	assert.Nil(t, err)
	assert.EqualValues(t, uint64(16927316757157635299), u64)

	f32, err := tp.GetFloat32() // pos 29
	assert.Nil(t, err)
	assert.EqualValues(t, -2.0833605e+021, f32)

	bytesFixed, err := tp.GetNBytes(2) // pos 33
	assert.Nil(t, err)
	assert.EqualValues(t, 2, len(bytesFixed))
	assert.EqualValues(t, 0xDE, bytesFixed[0])
	assert.EqualValues(t, 0xDD, bytesFixed[1])

	bytesDynamic, err := tp.GetBytes(0) // pos 35
	assert.Nil(t, err)
	assert.EqualValues(t, 60, len(bytesDynamic))
	assert.EqualValues(t, 0XD8, bytesDynamic[0])
	assert.EqualValues(t, 0xD7, bytesDynamic[1])

	strDynamic, err := tp.GetString(0) // pos 35
	assert.Nil(t, err)
	assert.EqualValues(t, 57, len(strDynamic))

	strFixed, err := tp.GetFixedString(7) // pos 35
	assert.Nil(t, err)
	assert.EqualValues(t, 7, len(strFixed))

	// Previously forgot the float64 test case, but we'll add it at the end so it doesn't shift our read values
	f64, err := tp.GetFloat64() // pos 29
	assert.Nil(t, err)
	assert.EqualValues(t, 3.6781362252089753e+117, f64)
}

func TestReachedEnd(t *testing.T) {
	// Create our fuzz data
	b := make([]byte, 1)
	b[0] = 0xFF

	// Create our type provider
	tp := New(b)

	// Assert the values are as expected
	b1, err := tp.GetByte() // pos 0
	assert.Nil(t, err)
	assert.EqualValues(t, 0xFF, b1)

	// Now expect errors reading any type.
	_, err = tp.GetByte()
	assert.NotNil(t, err)

	_, err = tp.GetInt8()
	assert.NotNil(t, err)

	_, err = tp.GetUint8()
	assert.NotNil(t, err)

	_, err = tp.GetInt16()
	assert.NotNil(t, err)

	_, err = tp.GetUint16()
	assert.NotNil(t, err)

	_, err = tp.GetInt32()
	assert.NotNil(t, err)

	_, err = tp.GetUint32()
	assert.NotNil(t, err)

	_, err = tp.GetInt64()
	assert.NotNil(t, err)

	_, err = tp.GetUint64()
	assert.NotNil(t, err)

	_, err = tp.GetFloat32()
	assert.NotNil(t, err)

	_, err = tp.GetFloat64()
	assert.NotNil(t, err)

	_, err = tp.GetString(100)
	assert.NotNil(t, err)

	_, err = tp.GetBytes(100)
	assert.NotNil(t, err)

	_, err = tp.GetFixedString(1)
	assert.NotNil(t, err)

	_, err = tp.GetNBytes(1)
	assert.NotNil(t, err)

	// Validate zero length strings and bytes will return successfully.
	s, err := tp.GetFixedString(0)
	assert.Nil(t, err)
	assert.EqualValues(t, 0, len(s))

	b, err = tp.GetNBytes(0)
	assert.Nil(t, err)
	assert.EqualValues(t, 0, len(b))
}