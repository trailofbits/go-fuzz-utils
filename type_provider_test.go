package go_fuzz_utils_test

import (
	"github.com/stretchr/testify/assert"
	goFuzzUtils "go-fuzz-utils"
	"sync"
	"testing"
)

func generateTestData(length uint) []byte {
	// Create our test data
	b := make([]byte, length)

	// Loop through and set our bytes accordingly. We do it in descending order so the values we test will have
	// negative bits set if they're signed integers first. If tests don't use all the data, this will at least test
	// integer width properties better since more bits will be set.
	for i := 0; i < len(b); i++ {
		b[i] = 255 - byte(i % 256)
	}

	return b
}

func TestSimpleTypes(t *testing.T) {
	// Create our fuzz data
	b := generateTestData(256)

	// Create our type provider
	tp := goFuzzUtils.NewTypeProvider(b)

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
	assert.EqualValues(t, 7, len(bytesDynamic))
	assert.EqualValues(t, 0XD8, bytesDynamic[0])
	assert.EqualValues(t, 0xD7, bytesDynamic[1])

	strDynamic, err := tp.GetString(0) // pos 46
	assert.Nil(t, err)
	assert.EqualValues(t, 62, len(strDynamic))

	strFixed, err := tp.GetFixedString(7) // pos 112
	assert.Nil(t, err)
	assert.EqualValues(t, 7, len(strFixed))

	// Previously forgot the float64 test case, but we'll add it at the end so it doesn't shift our read values
	f64, err := tp.GetFloat64() // pos 119
	assert.Nil(t, err)
	assert.EqualValues(t, -1.4249914579614907e-267, f64)

	// Architecture dependent integer type tests
	i, err := tp.GetInt() // pos 127
	assert.Nil(t, err)
	assert.EqualValues(t, -9187485637388043655, i)

	u, err := tp.GetUint() // pos 135
	assert.Nil(t, err)
	assert.EqualValues(t, 8680537053616894577, u)
}

func TestPositionReachedEnd(t *testing.T) {
	// Create our fuzz data
	b := generateTestData(1)

	// Create our type provider
	tp := goFuzzUtils.NewTypeProvider(b)

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

	_, err = tp.GetInt()
	assert.NotNil(t, err)

	_, err = tp.GetUint()
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

type testStruct struct {
	s1 string
	st1 struct {
		s string
		s2 string
		i int
	}
	sArr []string
	bArr []byte
	stArr [] struct {
		b byte
		s string
		i8 int8
		i16 int16
		i32 int32
		i64 int64
		f32 float32
		f64 float64
	}
	PublicString string
	PublicByte byte
	testMutex sync.Mutex
	PublicBytes []byte
}

func TestFillStructs(t *testing.T) {
	// Create our fuzz data
	b := generateTestData(0x1000)

	// Create our type provider
	tp := goFuzzUtils.NewTypeProvider(b)

	// Create a test structure and fill it.
	st := testStruct{}
	err := tp.Fill(&st, 15, 15, 0, true)

	// Ensure no error was encountered and private variables were filled in this instance.
	assert.Nil(t, err)
	assert.NotNil(t, st.sArr) // private variable, filled
	assert.NotNil(t, st.bArr) // private variable, filled
	assert.False(t, st.st1.s == "" && st.st1.s2 == "" && st.st1.i == 0) // depth 2, something should be non-default value.

	// Reset our position to use the same data, but without filling private variables.
	tp.Position = 0

	// Create a test structure and fill it.
	st2 := testStruct{}
	err = tp.Fill(&st2, 15, 15, 0, false)

	// Ensure no error was encountered and private variables weren't filled in this instance.
	assert.Nil(t, err)
	assert.Nil(t, st2.sArr) // private variable, unfilled
	assert.Nil(t, st2.bArr) // private variable, unfilled
	assert.EqualValues(t, "", st2.st1.s) // private variable, unfilled
	assert.EqualValues(t, "", st2.st1.s2) // private variable, unfilled
	assert.EqualValues(t, 0, st2.st1.i) // private variable, unfilled

	// Reset our position to use the same data, but without filling private variables.
	tp.Position = 0

	// Create a test structure and fill it.
	st3 := testStruct{}
	err = tp.Fill(&st3, 15, 15, 1, true)

	// Ensure no error was encountered and private variables weren't filled in this instance.
	assert.Nil(t, err)
	assert.NotNil(t, st3.sArr) // private variable, unfilled
	assert.NotNil(t, st3.bArr)
	assert.EqualValues(t, "", st3.st1.s)// depth 2, not filled
	assert.EqualValues(t, "", st3.st1.s2)// depth 2, not filled
	assert.EqualValues(t, 0, st3.st1.i)// depth 2, not filled
}

func TestFillBasicTypes(t *testing.T) {
	// Create our fuzz data
	b := generateTestData(0x1000)

	// Create our type provider
	tp := goFuzzUtils.NewTypeProvider(b)

	// Create a test structure and fill it.
	st := testStruct{}
	err := tp.Fill(&st, 15, 15, 0, true)

	// Ensure no error was encountered.
	assert.Nil(t, err)
	assert.NotNil(t, st.sArr) // private variable, filled
	assert.NotNil(t, st.bArr) // private variable, filled

	// Reset our position
	tp.Position = 0

	// Fill a int16
	var i16 int16
	err = tp.Fill(&i16, 0, 0, 0, false)
	assert.Nil(t, err)
	assert.EqualValues(t, -2, i16)

	// Fill a uint16
	var u16 uint16
	err = tp.Fill(&u16, 0, 0, 0, false)
	assert.Nil(t, err)
	assert.EqualValues(t, 65020, u16)

	// Fill a int32
	var i32 int32
	err = tp.Fill(&i32, 0, 0, 0, false)
	assert.Nil(t, err)
	assert.EqualValues(t, -67438088, i32)

	// Fill a uint32
	var u32 uint32
	err = tp.Fill(&u32, 0, 0, 0, false)
	assert.Nil(t, err)
	assert.EqualValues(t, 4160157172, u32)

	// Fill a int64
	var i64 int64
	err = tp.Fill(&i64, 0, 0, 0, false)
	assert.Nil(t, err)
	assert.EqualValues(t, -868365761009226260, i64)

	// Fill a uint64
	var u64 uint64
	err = tp.Fill(&u64, 0, 0, 0, false)
	assert.Nil(t, err)
	assert.EqualValues(t, uint64(16999656929995711972), u64)

	// Fill a float32
	var f32 float32
	err = tp.Fill(&f32, 0, 0, 0, false)
	assert.Nil(t, err)
	assert.EqualValues(t, -8.3704803e+021, f32)

	// Fill a float64
	var f64 float64
	err = tp.Fill(&f64, 0, 0, 0, false)
	assert.Nil(t, err)
	assert.EqualValues(t, -6.466470811086963e+153, f64)
}

func TestFillComplexTypes(t *testing.T) {
	// Create our fuzz data
	b := generateTestData(0x1000)

	// Create our type provider
	tp := goFuzzUtils.NewTypeProvider(b)

	// Create a mapping and fill it.
	m := make(map[string]int)
	err := tp.Fill(&m, 15, 15, 0, true)

	// Ensure something was generated.
	assert.Nil(t, err)
	assert.Greater(t, len(m), 0)
	assert.LessOrEqual(t, len(m), 15) // no more than 15 entries based on our args

	// Reset our position we extract data from in our array.
	tp.Position = 0

	// Create an array and fill it.
	u64Arr := make([]uint64, 20)
	err = tp.Fill(&u64Arr, 15, 15, 0, true)

	// Ensure something was generated.
	assert.Nil(t, err)
	assert.Greater(t, len(u64Arr), 0)
	assert.LessOrEqual(t, len(u64Arr), 15) // no more than 15 entries based on our args

	// Reset our position we extract data from in our array.
	tp.Position = 0

	// Create an array and fill it.
	mappingArr := make([]map[string]int, 15)
	err = tp.Fill(&mappingArr, 15, 15, 0, true)

	// Ensure something was generated.
	assert.Nil(t, err)
	assert.Greater(t, len(mappingArr), 0)
	assert.LessOrEqual(t, len(mappingArr), 15) // no more than 15 entries based on our args
}