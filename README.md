# go-picker

**Clean, readable JSON data extraction for Go without the bloat of repetitive error checking.**

Go-picker lets you write clean, readable code for extracting values from JSON data. Instead of checking errors after every operation, you focus on your business logic and validate everything at the end with a single `Confirm()` call.

## The Problem

Traditional JSON parsing in Go is verbose and error-prone:

```go
// Traditional approach - bloated with error checking
userObj, ok := data["user"].(map[string]interface{})
if !ok {
    return errors.New("invalid user")
}

name, ok := userObj["name"].(string)
if !ok {
    return errors.New("invalid name")
}

age, ok := userObj["age"].(float64)
if !ok {
    return errors.New("invalid age") 
}

profileObj, ok := userObj["profile"].(map[string]interface{})
if !ok {
    return errors.New("invalid profile")
}

email, ok := profileObj["email"].(string)
if !ok {
    return errors.New("invalid email")
}
```

## The Solution

With go-picker, the same logic becomes clean and readable:

```go
// Clean, readable approach
p := picker.NewPicker(data)

name := p.Nested("user").GetString("name")
age := p.Nested("user").GetInt("age")  
email := p.Nested("user").Nested("profile").GetString("email")

// Single validation check at the end
if err := p.Confirm(); err != nil {
    return fmt.Errorf("validation failed: %v", err)
}
```

## Key Benefits

- **🧹 Clean Code** - No repetitive error checking cluttering your logic
- **📖 Readable** - Code reads like the JSON structure you're extracting
- **🔒 Type-safe** - Generic functions return exact types without casting
- **🛡️ Strict Validation** - Unlike `json.Unmarshal`, requires ALL fields present and valid
- **🎯 Precise Errors** - Know exactly which fields failed with detailed paths
- **🌳 Nested Support** - Errors from nested pickers are collected in the original
- **⚡ Efficient** - Collect all errors in one pass, validate once

## Installation

```bash
go get github.com/karimagnusson/go-picker@v0.1.0
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
    
    // Parse JSON string into a picker
    p, err := picker.NewPickerFromJson(jsonStr)
    if err != nil {
        panic(err)
    }
    
    // Extract basic values
    name := p.GetString("name")
    age := p.GetInt("age")
    email := p.GetStringOr("email", "no-email")
    
    // Navigate nested objects with chaining
    location := p.Nested("profile").GetString("location")
    verified := p.Nested("profile").GetBool("verified")
    
    // Get typed arrays
    scores := picker.GetTypedArray[float64](p, "scores")
    
    // Validate all operations succeeded
    if err := p.Confirm(); err != nil {
        fmt.Printf("Validation errors: %v\n", err)
    }
    
    fmt.Printf("Name: %s, Age: %d, Location: %s, Verified: %t\n", name, age, location, verified)
}
```

## Strict Struct Validation

**When you need guaranteed data completeness (unlike permissive `json.Unmarshal`)**

`PickToStruct` provides strict validation - **ALL fields must be present and valid**. Perfect for APIs, configuration parsing, and data integrity checks where missing or invalid data should fail fast.

```go
type User struct {
    ID       int64     `json:"id"`       // Must exist and be valid int64
    Name     string    `json:"name"`     // Must exist and be valid string
    Email    string    `json:"email"`    // Must exist and be valid string
    Profile  *Profile  `json:"profile"`  // Must exist and be valid object
    Tags     []string  `json:"tags"`     // Must exist and be valid string array
}

type Profile struct {
    Location string `json:"location"`   // Must exist in profile object
    Verified bool   `json:"verified"`   // Must exist in profile object
}

var user User
err := picker.PickToStruct(jsonStr, &user)
if err != nil {
    // Will fail if ANY field is missing or wrong type
    log.Fatal(err) 
}
```

### vs Standard JSON Parsing

| **Scenario** | **`json.Unmarshal`** | **`picker.PickToStruct`** |
|--------------|---------------------|---------------------------|
| Missing field | ✅ Zero value (`""`, `0`, `false`) | ❌ **Error with exact path** |
| Wrong type | ✅ Zero value | ❌ **Error with exact path** |
| Incomplete data | ✅ Partial success | ❌ **Complete failure** |
| Error details | ❌ Generic message | ✅ **Precise field paths** |

```go
// This JSON succeeds with json.Unmarshal but FAILS with picker:
{
    "name": "John"
    // Missing required fields: id, email, profile, tags
}

// json.Unmarshal result: User{Name: "John", ID: 0, Email: "", Profile: nil, Tags: nil}
// picker.PickToStruct error: "Missing required fields: id, email, profile, tags"
```

**Use `PickToStruct` when:**
- 🛡️ **API validation** - Reject incomplete requests immediately
- ⚙️ **Configuration parsing** - Ensure all settings are present
- 🔍 **Data integrity** - No silent failures or missing data
- 📊 **Contract enforcement** - JSON must match struct exactly

## API Reference

### Creating Pickers

```go
// From JSON string
p, err := picker.NewPickerFromJson(jsonStr)

// From map
data := map[string]interface{}{"key": "value"}
p := picker.NewPicker(data)

// From HTTP request body
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
// Get raw array (returns []interface{})
items := p.GetArray("items")

// Get typed arrays (returns exact types)
names := picker.GetTypedArray[string](p, "names")     // returns []string
scores := picker.GetTypedArray[float64](p, "scores")   // returns []float64
flags := picker.GetTypedArray[bool](p, "flags")       // returns []bool

// Array of objects - NestedArray returns array of nested pickers
users := p.NestedArray("users")  // Returns *NestedPickerArray
for i, user := range users.Items {  // Each 'user' is a *Picker
    name := user.GetString("name")   // Errors collected in original 'p'
    email := user.GetString("email") // Errors collected in original 'p'
    fmt.Printf("User %d: %s (%s)\n", i, name, email)
}

// Safe item access with bounds checking and chaining
firstName := users.GetItem(0).GetString("name")  // Errors collected in 'p'

// ALL errors from nested pickers are validated in the original picker
if err := p.Confirm(); err != nil {
    // Catches errors like "users[0].name", "users[1].email", etc.
    fmt.Printf("Validation failed: %v\n", err)
}
```


### Error Key Examples

The `ErrorKeys()` method returns the exact paths that failed:

```go
data := map[string]interface{}{
    "user": map[string]interface{}{
        "name": "John",
        "age":  "invalid", // Should be int64
    },
    "scores": []interface{}{1.1, "invalid", 3.3}, // Mixed types
    "users": []interface{}{
        map[string]interface{}{"name": "Alice"},
        map[string]interface{}{"missing": "email"}, // Missing "name" field
    },
}

p := picker.NewPicker(data)

// These operations will fail:
p.Nested("user").GetInt("age")                    // Error key: "user.age"
picker.GetTypedArray[float64](p, "scores")        // Error key: "scores" 
p.NestedArray("users").GetItem(1).GetString("name") // Error key: "users[1].name"
p.GetString("nonexistent")                        // Error key: "nonexistent"

if err := p.Confirm(); err != nil {
    errorKeys := p.ErrorKeys()
    // Output: ["user.age", "scores", "users[1].name", "nonexistent"]
}
```

### Utility Methods

```go
// Check key existence
if p.HasKey("optional_field") {
    value := p.GetString("optional_field")
}

// Get detailed error information
errorKeys := p.ErrorKeys()  // Returns []string with exact paths
fmt.Printf("Failed keys: %v\n", errorKeys)

// Access underlying data if needed
data := p.GetData()  // Returns map[string]interface{}

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

**The key to clean, readable code: collect errors, validate once.**

go-picker uses an error collection approach that eliminates repetitive error checking:

### How It Works

1. **📝 Collection Phase** - All `Get*()` operations collect errors silently
   ```go
   // No errors returned here - just collect and continue
   name := p.GetString("name")
   age := p.GetInt("age")
   email := p.Nested("profile").GetString("email")
   ```

2. **✅ Validation Phase** - **You MUST call `Confirm()` after reading**
   ```go
   // Single validation check catches all errors
   if err := p.Confirm(); err != nil {
       return fmt.Errorf("validation failed: %v", err)
   }
   ```

3. **🛡️ Graceful Degradation** - Invalid operations return zero values but track errors
   ```go
   age := p.GetInt("invalid_field")  // Returns 0, tracks error
   ```

### Critical: Nested Error Collection

**Errors from nested pickers are automatically collected in the original picker:**

```go
p := picker.NewPicker(data)

// Errors from these nested operations...
userPicker := p.Nested("user")
name := userPicker.GetString("name")
age := userPicker.GetInt("age")

profilePicker := userPicker.Nested("profile")  
email := profilePicker.GetString("email")

// ...are ALL collected in the original picker 'p'
if err := p.Confirm(); err != nil {
    // This catches errors from user.name, user.age, AND user.profile.email
    errorKeys := p.ErrorKeys()
    fmt.Printf("Failed paths: %v\n", errorKeys)
    // Output: ["user.name", "user.age", "user.profile.email"]
}
```

### Why This Approach?

- **🎯 Focus on Logic** - Write business logic without error noise
- **📚 Readable Code** - Code structure mirrors JSON structure  
- **🔍 Complete Validation** - Catch ALL issues in one place
- **🏎️ Efficient** - Process entire structure, fail fast only when needed

## License

MIT License - see LICENSE file for details.