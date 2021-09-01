# go-fuzz-utils
`go-fuzz-utils` is a helper package for use with [go-fuzz](https://github.com/dvyukov/go-fuzz) or other fuzzing utilities. It provides a simple interface to produce random values for various data types and can recursively populate complex structures from raw fuzz data generated by `go-fuzz`. Spend more time writing property tests, and less time with ugly data type conversions, edge cases supporting full value ranges, `nil` cases, etc. Simply feed `go-fuzz` data into `go-fuzz-utils` to produce fuzzed objects and use them in your property tests as needed.

When populating variables, you can configure a number of parameters:
- Minimum/maximum sizes of strings, maps, slices
- Probability of `nil` for maps, slices, pointers
- Depth limit for nested structures
- Toggle for filling unexported fields in structures
- Probability of skipping a field when filling (to randomly fuzz over valid structure fields)

## Setup
Import this package into your `go-fuzz` tests:
```go
import "github.com/trailofbits/go-fuzz-utils"
```

As `go-fuzz` provides `[]byte` data for fuzzing campaigns, simply ensure enough data is provided to convert into all the necessary data types, then construct a new `TypeProvider` using `NewTypeProvider(...)`.  
```go
func Fuzz(data []byte) int {
	// Verify we have a sufficient amount of data to produce all the variables we need.
	if len(data) < 0x100 {
		return 0
	}

	// Create a new type provider
	tp, err := go_fuzz_utils.NewTypeProvider(data)
	if err != nil {
		return 0 // shouldn't happen, only if not enough data is supplied.
	}
[...]
```

## Simple data types
You can obtain the necessary type of data with exported functions such as:
```go
	// Obtain a byte
	b, err := tp.GetByte()
...
	// Obtain a bool
	bl, err := tp.GetBool()
...
	// Obtain an int16
	i16, err := tp.GetInt16()
...
	// Obtain a float32
	f32, err := tp.GetFloat32()
...
	// Obtain a fixed-length string
	strFixed, err := tp.GetFixedString(7)
...
	// Obtain a dynamic-length string
	strDynamic, err := tp.GetString() // uses TypeProvider parameters to determine length/nil possibility
...
	// Obtain a fixed-length byte array
	bytesFixed, err := tp.GetNBytes(2)
...
	// Obtain a dynamic-length byte array
	bytesDynamic, err := tp.GetBytes() // uses TypeProvider parameters to determine length/nil possibility
```


## Structures
`go-fuzz-utils` exposes a generic `Fill(...)` method which can populate simple data types, mappings, arrays, and arbitrary structures recursively via reflection. 

For example, given the following structure:
```go
	type Person struct {
		ID uint64
		Name string
		Photo []byte
		Employed bool
		EmergencyContact *Person
	}
```

You can simply perform a `Fill` call to populate it with the fuzz data. Even though `Person` has a circular reference in `EmergencyContact`, you can configure depth limits and `nil` bias settings to prevent infinite loops while giving us various deeply nested structures.
```go
	// Create a person struct and fill it recursively. 
	var p Person
	err := tp.Fill(&p)    
```

Similarly, you can fill other data types as needed:
```go
	// Create an array of mappings and fill them
	mappingArr := make([]map[string]int, 15)
	err = tp.Fill(&mappingArr)
```
