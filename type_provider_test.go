package go_fuzz_utils_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/trailofbits/go-fuzz-utils"
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
		b[i] = 255 - byte(i%256)
	}

	return b
}

func TestSimpleTypes(t *testing.T) {
	// Create our fuzz data
	b := generateTestData(256)

	// Create our type provider
	tp, err := go_fuzz_utils.NewTypeProvider(b)
	assert.Nil(t, err)

	// Basic values
	b1, err := tp.GetByte()
	assert.Nil(t, err)
	assert.EqualValues(t, 0xF7, b1)

	i16, err := tp.GetInt16()
	assert.Nil(t, err)
	assert.EqualValues(t, -2315, i16)

	u16, err := tp.GetUint16()
	assert.Nil(t, err)
	assert.EqualValues(t, 0xF4F3, u16)

	i32, err := tp.GetInt32()
	assert.Nil(t, err)
	assert.EqualValues(t, -219025169, i32)

	u32, err := tp.GetUint32()
	assert.Nil(t, err)
	assert.EqualValues(t, 0xEEEDECEB, u32)

	i64, err := tp.GetInt64()
	assert.Nil(t, err)
	assert.EqualValues(t, -1519427316551916317, i64)

	u64, err := tp.GetUint64()
	assert.Nil(t, err)
	assert.EqualValues(t, uint64(0xE2E1E0DFDEDDDCDB), u64)

	f32, err := tp.GetFloat32()
	assert.Nil(t, err)
	assert.EqualValues(t, -3.0659244e+16, f32)

	f64, err := tp.GetFloat64()
	assert.Nil(t, err)
	assert.EqualValues(t, -2.050874878169571e+110, f64)

	// Architecture dependent integer type tests
	i, err := tp.GetInt()
	assert.Nil(t, err)
	assert.EqualValues(t, -3544952156018063161, i)

	u, err := tp.GetUint()
	assert.Nil(t, err)
	assert.EqualValues(t, uint(0xC6C5C4C3C2C1C0BF), u)

	// Slices/strings
	bytesFixed, err := tp.GetNBytes(2)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, len(bytesFixed))
	assert.EqualValues(t, 0xBE, bytesFixed[0])
	assert.EqualValues(t, 0xBD, bytesFixed[1])

	bytesDynamic, err := tp.GetBytes()
	assert.Nil(t, err)
	assert.EqualValues(t, 3, len(bytesDynamic))
	assert.EqualValues(t, 0xBC, bytesDynamic[0])
	assert.EqualValues(t, 0xBB, bytesDynamic[1])

	strDynamic, err := tp.GetString()
	assert.Nil(t, err)
	assert.EqualValues(t, 11, len(strDynamic))

	strFixed, err := tp.GetFixedString(7)
	assert.Nil(t, err)
	assert.EqualValues(t, 7, len(strFixed))
}

func TestPositionReachedEnd(t *testing.T) {
	// Create our fuzz data
	b := generateTestData(1)

	// Create our type provider. We should encounter an error since we need at least 64-bits to read a random seed from.
	tp, err := go_fuzz_utils.NewTypeProvider(b)
	assert.NotNil(t, err)

	// Create more fuzz data
	b = generateTestData(9)

	// Recreate our type provider, this time it should succeed, reading 8 bytes as a random seed, leaving 1 byte left.
	tp, err = go_fuzz_utils.NewTypeProvider(b)
	assert.Nil(t, err)

	// Assert the values are as expected
	b1, err := tp.GetByte()
	assert.Nil(t, err)
	assert.EqualValues(t, 0xF7, b1)

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

	_, err = tp.GetString()
	assert.NotNil(t, err)

	_, err = tp.GetBytes()
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
	s1  string
	st1 struct {
		s  string
		s2 string
		i  int
	}
	sArr  []string
	bArr  []byte
	stArr []struct {
		b   byte
		s   string
		i8  int8
		i16 int16
		i32 int32
		i64 int64
		f32 float32
		f64 float64
	}
	PublicString string
	PublicByte   byte
	testMutex    sync.Mutex
	PublicBytes  []byte
}

func TestFillStructs(t *testing.T) {
	// Create our fuzz data
	b := generateTestData(0x1000)

	// Create our type provider
	tp, err := go_fuzz_utils.NewTypeProvider(b)
	assert.Nil(t, err)

	// Create a test structure and fill it.
	st := testStruct{}
	err = tp.SetParamsBiasesCommon(0, 0)
	assert.Nil(t, err)
	tp.SetParamsFillUnexportedFields(true)
	err = tp.Fill(&st)

	// Ensure no error was encountered and private variables were filled in this instance.
	assert.Nil(t, err)
	assert.NotNil(t, st.sArr)                                           // private variable, filled
	assert.NotNil(t, st.bArr)                                           // private variable, filled
	assert.False(t, st.st1.s == "" && st.st1.s2 == "" && st.st1.i == 0) // depth 2, something should be non-default value.

	// Reset our provider state
	err = tp.Reset()
	assert.Nil(t, err)

	// Create a test structure and fill it.
	st2 := testStruct{}
	tp.SetParamsFillUnexportedFields(false)
	err = tp.Fill(&st2)

	// Ensure no error was encountered and private variables weren't filled in this instance.
	assert.Nil(t, err)
	assert.Nil(t, st2.sArr)               // private variable, unfilled
	assert.Nil(t, st2.bArr)               // private variable, unfilled
	assert.EqualValues(t, "", st2.st1.s)  // private variable, unfilled
	assert.EqualValues(t, "", st2.st1.s2) // private variable, unfilled
	assert.EqualValues(t, 0, st2.st1.i)   // private variable, unfilled

	// Reset our provider state
	err = tp.Reset()
	assert.Nil(t, err)

	// Create a test structure and fill it.
	st3 := testStruct{}
	assert.Nil(t, tp.SetParamsDepthLimit(1))
	tp.SetParamsFillUnexportedFields(true)
	err = tp.Fill(&st3)

	// Ensure no error was encountered and private variables weren't filled in this instance.
	assert.Nil(t, err)
	assert.NotNil(t, st3.sArr) // private variable, unfilled
	assert.NotNil(t, st3.bArr)
	assert.EqualValues(t, "", st3.st1.s)  // depth 2, not filled
	assert.EqualValues(t, "", st3.st1.s2) // depth 2, not filled
	assert.EqualValues(t, 0, st3.st1.i)   // depth 2, not filled
}

func TestFillBasicTypes(t *testing.T) {
	// Create our fuzz data
	b := generateTestData(0x1000)

	// Create our type provider
	tp, err := go_fuzz_utils.NewTypeProvider(b)
	assert.Nil(t, err)

	// Create a test structure and fill it.
	st := testStruct{}
	err = tp.Fill(&st)

	// Ensure no error was encountered.
	assert.Nil(t, err)
	assert.NotNil(t, st.sArr) // private variable, filled
	assert.NotNil(t, st.bArr) // private variable, filled

	// Reset our provider state
	err = tp.Reset()
	assert.Nil(t, err)

	// Fill a int16
	var i16 int16
	err = tp.Fill(&i16)
	assert.Nil(t, err)
	assert.EqualValues(t, -2058, i16)

	// Fill a uint16
	var u16 uint16
	err = tp.Fill(&u16)
	assert.Nil(t, err)
	assert.EqualValues(t, 0xF5F4, u16)

	// Fill a int32
	var i32 int32
	err = tp.Fill(&i32)
	assert.Nil(t, err)
	assert.EqualValues(t, -202182160, i32)

	// Fill a uint32
	var u32 uint32
	err = tp.Fill(&u32)
	assert.Nil(t, err)
	assert.EqualValues(t, 0xEFEEEDEC, u32)

	// Fill a int64
	var i64 int64
	err = tp.Fill(&i64)
	assert.Nil(t, err)
	assert.EqualValues(t, -1447087143713839644, i64)

	// Fill a uint64
	var u64 uint64
	err = tp.Fill(&u64)
	assert.Nil(t, err)
	assert.EqualValues(t, uint64(0xE3E2E1E0DFDEDDDC), u64)

	// Fill a float32
	var f32 float32
	err = tp.Fill(&f32)
	assert.Nil(t, err)
	assert.EqualValues(t, -1.2320213e+17, f32)

	// Fill a float64
	var f64 float64
	err = tp.Fill(&f64)
	assert.Nil(t, err)
	assert.EqualValues(t, -1.405868428700574e+115, f64)
}

func TestFillComplexTypes(t *testing.T) {
	// Create our fuzz data
	b := generateTestData(0x1000)

	// Create our type provider
	tp, err := go_fuzz_utils.NewTypeProvider(b)
	assert.Nil(t, err)
	assert.Nil(t, tp.SetParamsBiasesCommon(0, 0))

	// Create a mapping and fill it.
	m := make(map[string]int)
	err = tp.Fill(&m)

	// Ensure something was generated.
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(m), 0)
	assert.LessOrEqual(t, len(m), 15) // no more than 15 entries based on our args

	// Create a slice and fill it.
	u64Arr := make([]uint64, 0)
	err = tp.Fill(&u64Arr)

	// Ensure something was generated.
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(u64Arr), 0)
	assert.LessOrEqual(t, len(u64Arr), 15) // no more than 15 entries based on our args

	// Create a slice and fill it.
	mappingArr := make([]map[string]int, 0)
	assert.Nil(t, tp.SetParamsMapBounds(1, 1))
	assert.Nil(t, tp.SetParamsSliceBounds(3, 3))
	err = tp.Fill(&mappingArr)

	// Ensure something was generated.
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(mappingArr), 0)
	assert.LessOrEqual(t, len(mappingArr), 15) // no more than 15 entries based on our args
}

func TestNilBiases(t *testing.T) {
	// Create our fuzz data
	b := generateTestData(0x1000)

	// Create our type provider
	tp, err := go_fuzz_utils.NewTypeProvider(b)
	assert.Nil(t, err)

	// Define a struct to test our nil biases for slice
	type TestStruct struct {
		x []uint8
	}
	var nestedSlices [20]TestStruct

	// Fill our slice and verify length
	assert.Nil(t, tp.SetParamsSliceBounds(15, 15))
	assert.Nil(t, tp.SetParamsBiasesCommon(0, 0))
	err = tp.Fill(&nestedSlices)
	assert.Nil(t, err)

	// Verify nothing in the struct is nil based on our 0 bias.
	for i := 0; i < len(nestedSlices); i++ {
		assert.NotNil(t, nestedSlices[i].x)
		assert.EqualValues(t, 15, len(nestedSlices[i].x))
	}

	// Try to fill our slice again, this time with a full nil bias.
	var nestedSlices2 [20]TestStruct
	assert.Nil(t, tp.SetParamsBiasesCommon(1, 0))
	err = tp.Fill(&nestedSlices2)
	assert.Nil(t, err)

	// Verify everything in the struct is nil based on our bias.
	for i := 0; i < len(nestedSlices2); i++ {
		assert.Nil(t, nestedSlices2[i].x)
	}
}

func TestByteArrayFilling(t *testing.T) {
	// We have a special optimization case for byte slices that is separate from other slice types, so we offer a quick
	// test to ensure that those types are populated without issue.

	// Create our fuzz data (the test data isn't great for this test, so we tweak it a bit)
	b := append([]byte{6}, generateTestData(0x1000)...)

	// Create our type provider
	tp, err := go_fuzz_utils.NewTypeProvider(b)
	assert.Nil(t, err)

	// Create a slice and fill it.
	bSlice := make([]byte, 0)
	err = tp.Fill(&bSlice)

	// Ensure something was generated.
	assert.Nil(t, err)

	assert.GreaterOrEqual(t, len(bSlice), 0)
	assert.LessOrEqual(t, len(bSlice), 15) // no more than 15 entries based on our args
}

func TestSkipBiases(t *testing.T) {
	// Create our fuzz data
	b := generateTestData(0x1000)

	// Create our type provider
	tp, err := go_fuzz_utils.NewTypeProvider(b)
	assert.Nil(t, err)

	// Define a struct to test our skip biases. We define every type that is considered for skipping.
	type TestStruct struct {
		sliceVal []byte
		mapVal   map[int]int
		ptrVal   *int
	}
	x := 7
	skipStruct := TestStruct{
		sliceVal: []byte{0, 1, 2, 3},
		mapVal:   map[int]int{0: 0, 1: 1, 2: 2, 3: 3},
		ptrVal:   &x,
	}

	// We'll try to fill all maps/ptr/slices with nil, but also set skip to a full bias so it shouldn't ever actually
	// happen.
	assert.Nil(t, tp.SetParamsBiasesCommon(1, 1))
	assert.Nil(t, tp.SetParamsSliceBounds(0, 0))
	err = tp.Fill(&skipStruct)
	assert.Nil(t, err)

	// Ensure all values are not nil
	assert.NotNil(t, skipStruct.mapVal)
	assert.NotNil(t, skipStruct.ptrVal)
	assert.NotNil(t, skipStruct.sliceVal)

	// Now fill with the no skip bias, and they should all be nil
	assert.Nil(t, tp.SetParamsBiasesCommon(1, 0))
	err = tp.Fill(&skipStruct)
	assert.Nil(t, err)

	// Ensure all values are now nil given we have full nil bias and no skip bias
	assert.Nil(t, skipStruct.mapVal)
	assert.Nil(t, skipStruct.ptrVal)
	assert.Nil(t, skipStruct.sliceVal)
}
