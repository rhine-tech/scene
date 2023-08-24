# upartial

The `upartial` package provides functionality to update one Go struct with the values of another struct, using reflection. This package is particularly useful for scenarios where you want to partially update a destination struct with non-nil values from a source struct.

## Features

- Automatically updates matching fields between source and destination structs.
- Supports nested structs.
- Allows specifying default values using the `upartial` tag, which will be used if the corresponding source field is nil.


## Usage

```go
package main

type A struct {
	Value0 *string `upartial:"default0"`
	Value1 *int    `upartial:"42"`
}

type B struct {
	Value0 string
	Value1 int
}

func main() {
	src := &A{}
	dest := &B{}
	err := upartial.UpdateStruct(src, dest)
	if err != nil {
		log.Fatal(err)
	}
	// dest will now have Value0 = "default0" and Value1 = 42
}
```

## Author

@Aynakeya (Yiyang Lu)