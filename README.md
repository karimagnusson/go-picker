# go-picker

A lightweight, type-safe JSON data picker library for Go that provides fluent API for extracting and validating JSON data with comprehensive error tracking.

## Features

- **Type-safe data extraction** - Extract strings, integers, floats, booleans, and big numbers
- **Nested object support** - Navigate through nested JSON structures  
- **Array handling** - Work with arrays of objects and primitive types
- **Struct mapping** - Automatically map JSON to Go structs using reflection
- **Error tracking** - Collect and validate all extraction errors
- **Fallback values** - Provide default values for missing or invalid data
- **HTTP integration** - Direct parsing from HTTP requests

## Installation

```bash
go get github.com/karimagnusson/picker
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/karimagnusson/picker"
)

func main() {
    jsonStr := `{
        "name": "John Doe",
        "age": 30,
        "email": "john@example.com",
        "profile": {
            "location": "New York",
            "verified": true
        },
        "scores": [85.5, 92.0, 78.5]
    }`
    
    p, err := picker.NewPickerFromJson(jsonStr)
    if err != nil {
        panic(err)
    }
    
    // Extract basic values
    name := p.GetString("name")
    age := p.GetInt("age")
    email := p.GetStringOr("email", "no-email")
    
    // Navigate nested objects
    profile := p.Nested("profile")
    location := profile.GetString("location")
    verified := profile.GetBool("verified")
    
    // Get typed arrays
    scores := picker.GetTypedArray[float64](p, "scores")
    
    // Validate all operations succeeded
    if err := p.Confirm(); err != nil {
        fmt.Printf("Validation errors: %v\n", err)
    }
    
    fmt.Printf("Name: %s, Age: %d, Location: %s\n", name, age, location)
}
```

## API Reference

### Creating Pickers

```go
// From JSON string
p, err := picker.NewPickerFromJson(jsonStr)

// From map
data := map[string]interface{}{"key": "value"}
p := picker.NewPicker(data)

// From HTTP request
p, err := picker.NewPickerFromRequest(r)
```

### Basic Data Extraction

```go
// Required values (adds error if missing/wrong type)
name := p.GetString("name")
age := p.GetInt("age")
price := p.GetFloat("price")
active := p.GetBool("active")
date := p.GetDate("created_at") // RFC3339 format

// Optional values with fallbacks
name := p.GetStringOr("name", "Unknown")
age := p.GetIntOr("age", 0)
price := p.GetFloatOr("price", 0.0)
active := p.GetBoolOr("active", false)
```

### Nested Objects

```go
// Navigate to nested object
user := p.Nested("user")
name := user.GetString("name")

// Chain navigation
email := p.Nested("user").Nested("contact").GetString("email")

// Create new picker for nested object (isolates errors)
userPicker := p.GetNewPicker("user")
```

### Arrays

```go
// Get raw array
items := p.GetArray("items")

// Get typed arrays
names := picker.GetTypedArray[string](p, "names")
scores := picker.GetTypedArray[float64](p, "scores")
flags := picker.GetTypedArray[bool](p, "flags")

// Array of objects
users := p.NestedArray("users")
for i, user := range users.Items {
    name := user.GetString("name")
    fmt.Printf("User %d: %s\n", i, name)
}
```

### Struct Mapping

Automatically map JSON to Go structs using reflection:

```go
type User struct {
    ID       int64     `json:"id"`
    Name     string    `json:"name"`
    Email    string    `json:"email"`
    Profile  *Profile  `json:"profile"`
    Tags     []string  `json:"tags"`
}

type Profile struct {
    Location string `json:"location"`
    Verified bool   `json:"verified"`
}

var user User
err := picker.PickToStruct(jsonStr, &user)
if err != nil {
    log.Fatal(err)
}
```

### Error Handling

```go
// Check for any errors
if err := p.Confirm(); err != nil {
    fmt.Printf("Validation errors: %v\n", err)
}

// Validate all operations (fails if any errors)
if err := p.Confirm(); err != nil {
    fmt.Printf("Validation failed: %v\n", err)
}
```

### Utility Methods

```go
// Check key existence
if p.HasKey("optional_field") {
    value := p.GetString("optional_field")
}

// Get all keys
keys := p.Keys()

// Manipulate data
p.AddKey("new_field", "value")
p.DelKey("unwanted_field")

// Copy and transform
copy := p.Copy()
flat := p.FlatCopy() // Flattens nested objects with dot notation

// Serialization
jsonStr := p.ToJsonString()
prettyJson := p.ToPrettyJsonString()
```

## Typed Arrays

The generic `GetTypedArray` function provides type-safe array extraction:

```go
// Basic types
strings := picker.GetTypedArray[string](p, "names")        // []string
ints := picker.GetTypedArray[int64](p, "numbers")          // []int64  
floats := picker.GetTypedArray[float64](p, "scores")       // []float64
bools := picker.GetTypedArray[bool](p, "flags")            // []bool

// Big number types
bigints := picker.GetTypedArray[*big.Int](p, "large")      // []*big.Int
bigfloats := picker.GetTypedArray[*big.Float](p, "precise") // []*big.Float
bigrats := picker.GetTypedArray[*big.Rat](p, "ratios")     // []*big.Rat
```

## Error Handling Philosophy

go-picker uses an error collection approach rather than failing fast:

1. **Collection Phase** - All `Get*()` operations collect errors without stopping
2. **Validation Phase** - Call `Confirm()` to check if any operations failed
3. **Graceful Degradation** - Invalid operations return zero values but track errors

This allows you to extract all valid data even when some fields are missing or invalid.

## License

MIT License - see LICENSE file for details.