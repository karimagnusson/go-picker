# go-picker

**Type-safe JSON validation for Go with clear error messages**

Validate all required fields and collect every error in one step. No repetitive error checking, no silent failures with zero values.

## Installation

```bash
go get github.com/karimagnusson/go-picker
```

## Quick Example

```go
type User struct {
    Name  string
    Email string
    Age   int64
}

jsonStr := `{"name": "John", "email": "john@example.com", "age": 30}`

user, err := picker.PickFromJson(jsonStr, func(p *picker.Picker) User {
    return User{
        Name:  p.GetString("name"),
        Email: p.GetString("email"),
        Age:   p.GetInt("age"),
    }
})
```

## The Problem

Traditional JSON parsing in Go has two options, both problematic:

**Option 1: `json.Unmarshal`** - Silent failures

```go
var user User
json.Unmarshal([]byte(jsonStr), &user)
// Missing fields become zero values: "", 0, nil
// No way to know if data was actually validated
```

**Option 2: Manual validation** - Verbose and repetitive

```go
name, ok := data["name"].(string)
if !ok {
    return errors.New("invalid name")
}
email, ok := data["email"].(string)
if !ok {
    return errors.New("invalid email")
}
// ... repeat for every field
```

## With go-picker

go-picker validates **all required fields** and collects **all errors** in one pass:

```go
user, err := picker.PickFromJson(jsonStr, func(p *picker.Picker) User {
    return User{
        Name:  p.GetString("name"),
        Email: p.GetString("email"),
        Age:   p.GetInt("age"),
    }
})
// Automatically validates - no forgotten Confirm() calls
// Returns detailed errors for ALL missing/invalid fields
```

## Features

-   **Automatic validation** - All fields validated in one operation, no manual error checking
-   **Detailed error reporting** - Returns a map of all validation errors with field paths
-   **Type-safe parsing** - Direct conversion from JSON to your structs
-   **Nested objects and arrays** - Clean API for complex JSON structures
-   **Customizable error messages** - Built-in support for internationalization

## API

### Pick Functions

```go
// Pick from parsed data
picker.Pick(data, func(p *picker.Picker) T { ... })

// Pick from JSON string
picker.PickFromJson(jsonStr, func(p *picker.Picker) T { ... })

// Pick from HTTP request body
picker.PickFromRequestBody(r, func(p *picker.Picker) T { ... })
```

### Helper Functions

```go
// Parse JSON string into map
data, err := picker.ParseJson(jsonStr)  // returns map[string]interface{}

// Parse HTTP request body into map
data, err := picker.ParseRequestBody(r)  // returns map[string]interface{}
```

### Getter Methods

#### Required Fields

```go
p.GetString("name")                    // string
p.GetInt("age")                        // int64 (from JSON number)
p.GetFloat("price")                    // float64
p.GetBool("active")                    // bool
p.GetDate("created_at")                // time.Time (supports RFC3339, date-only, and RFC3339 without timezone)
p.GetObject("metadata")                // map[string]interface{}
p.GetArray("items")                    // []interface{}
```

#### Optional Fields with Fallbacks

```go
p.GetStringOr("email", "default@example.com")
p.GetIntOr("age", 18)
p.GetFloatOr("price", 0.0)
p.GetBoolOr("active", false)
p.GetDateOr("updated", time.Now())
p.GetObjectOr("metadata", map[string]interface{}{})
p.GetArrayOr("tags", []interface{}{})
```

#### Nested Objects and Arrays

```go
p.Nested("user")                            // *Picker for nested object
array := p.NestedArray("users")             // *NestedPickerArray for array of objects
picker.GetTypedArray[T](p, "items")         // []T for typed arrays
picker.Map[T](array, func(*Picker) T)       // []T - map array items through a function
array.At(index)                             // *Picker - get item at index with bounds checking
```

### Error Handling

Use `HasDetail(err)` to check if the error is a validation error, then `Detail(err)` to get the error map:

```go
user, err := picker.PickFromJson(jsonStr, func(p *picker.Picker) User {
    return User{
        Name: p.GetString("name"),
        Age:  p.GetInt("age"),
    }
})

if err != nil {
    // Check if it's a validation error
    if picker.HasDetail(err) {
        // Get detailed error map
        errors := picker.Detail(err)
        // errors = map[string]string{
        //     "name": "missing",
        //     "age": "invalid"
        // }

        for field, reason := range errors {
            fmt.Printf("%s: %s\n", field, reason)
        }
        return
    }

    // Otherwise, it's a JSON parse error
    fmt.Printf("Parse error: %v\n", err)
    return
}
```

### Nested Objects

Use `Nested(key)` to access nested objects. Errors will include the full path:

```go
jsonStr := `{
    "user": {
        "name": "John",
        "profile": {
            "email": "john@example.com"
        }
    }
}`

result, err := picker.PickFromJson(jsonStr, func(p *picker.Picker) Result {
    user := p.Nested("user")
    profile := user.Nested("profile")

    return Result{
        Name:  user.GetString("name"),
        Email: profile.GetString("email"),
    }
})

// Errors from nested fields show full path: "user.profile.email"
```

### Arrays

Use `GetTypedArray[T]` for arrays of primitive types, or `NestedArray` with `Map` to transform arrays of objects:

```go
// Typed arrays of primitives
tags := picker.GetTypedArray[string](p, "tags")     // []string
scores := picker.GetTypedArray[float64](p, "scores") // []float64

// Array of objects - using Map
jsonStr := `{
    "users": [
        {"name": "John", "age": 30},
        {"name": "Jane", "age": 25}
    ]
}`

result, err := picker.PickFromJson(jsonStr, func(p *picker.Picker) Result {
    users := p.NestedArray("users")

    // Map to extract just the names
    names := picker.Map(users, func(user *picker.Picker) string {
        return user.GetString("name")
    })

    return Result{Names: names}
})

// Or map to structs
userList := picker.Map(users, func(user *picker.Picker) User {
    return User{
        Name: user.GetString("name"),
        Age:  user.GetInt("age"),
    }
})

// Access specific item by index with bounds checking
firstUser := users.At(0).GetString("name")  // "John"
```

### Custom Validation

Use `SetError(key, message)` to add custom validation errors or `SetInvalid(key)` to mark a field as invalid:

```go
user, err := picker.Pick(data, func(p *picker.Picker) User {
    age := p.GetInt("age")
    if age < 18 {
        p.SetError("age", "must be 18 or older")
    }

    email := p.GetString("email")
    if !strings.Contains(email, "@") {
        p.SetInvalid("email")
    }

    return User{Age: age, Email: email}
})
```

### Customizable Error Messages

By default, validation errors will be either `"missing"` (field not present in JSON) or `"invalid"` (field has wrong type). You can customize these messages:

```go
// Default values
picker.ErrorMissing = "missing"
picker.ErrorInvalid = "invalid"

// Customize for your application
picker.ErrorMissing = "required"
picker.ErrorInvalid = "wrong type"

// Or for internationalization
picker.ErrorMissing = "faltante"  // Spanish
picker.ErrorInvalid = "inválido"
```

## HTTP Handler Example

```go
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    user, err := picker.PickFromRequestBody(r, func(p *picker.Picker) User {
        return User{
            Name:  p.GetString("name"),
            Email: p.GetString("email"),
            Age:   p.GetInt("age"),
        }
    })

    if err != nil {
        if picker.HasDetail(err) {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "errors": picker.Detail(err),
            })
            return
        }
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // user is validated and ready to save
    saveUser(user)
    w.WriteHeader(http.StatusCreated)
}
```

## License

MIT License - see LICENSE file for details.
