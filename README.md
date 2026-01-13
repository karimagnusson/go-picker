# go-picker

**Type-safe JSON validation for Go with clear error messages**

Parse JSON data and validate all required fields in one operation. No repetitive error checking, no silent failures with zero values.

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

if err != nil {
    // Get detailed error information
    errors := picker.Detail(err)
    // errors = map[string]string{"email": "missing", "age": "invalid"}
    for field, reason := range errors {
        fmt.Printf("%s: %s\n", field, reason)
    }
    return
}

// user is validated and ready to use
fmt.Printf("User: %s, %s, %d\n", user.Name, user.Email, user.Age)
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

## The Solution

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

- ✅ **Automatic validation** - No manual `Confirm()` needed
- 📊 **Detailed errors** - Know exactly which fields failed and why
- 🎯 **Type-safe** - Returns your exact struct type
- 🔍 **All errors at once** - Collects all validation errors in one pass
- 🌐 **Customizable messages** - Support for internationalization
- 🌳 **Nested support** - Clean handling of nested objects and arrays

## API

### Pick Functions

```go
// From parsed JSON data
data, _ := picker.ParseJson(jsonStr)
user, err := picker.Pick(data, func(p *picker.Picker) User {
    return User{
        Name: p.GetString("name"),
        Age:  p.GetInt("age"),
    }
})

// From JSON string (combines parsing + validation)
user, err := picker.PickFromJson(jsonStr, func(p *picker.Picker) User {
    return User{
        Name: p.GetString("name"),
        Age:  p.GetInt("age"),
    }
})

// From HTTP request body
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    user, err := picker.PickFromRequestBody(r, func(p *picker.Picker) User {
        return User{
            Name: p.GetString("name"),
            Age:  p.GetInt("age"),
        }
    })

    if err != nil {
        // Handle error
        return
    }

    // user is validated
}
```

### Getter Methods

#### Required Fields

```go
p.GetString("name")                    // string
p.GetInt("age")                        // int64 (from JSON number)
p.GetFloat("price")                    // float64
p.GetBool("active")                    // bool
p.GetDate("created_at")                // time.Time (RFC3339 format)
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

### Error Handling

```go
user, err := picker.PickFromJson(jsonStr, func(p *picker.Picker) User {
    return User{
        Name: p.GetString("name"),
        Age:  p.GetInt("age"),
    }
})

if err != nil {
    // Get detailed error map
    errors := picker.Detail(err)
    // errors = map[string]string{
    //     "name": "missing",
    //     "age": "invalid"
    // }

    // If empty, it was a JSON parse error
    if len(errors) == 0 {
        fmt.Printf("Parse error: %v\n", err)
        return
    }

    for field, reason := range errors {
        fmt.Printf("%s: %s\n", field, reason)
    }
}
```

### Nested Objects

```go
jsonStr := `{
    "user": {
        "name": "John",
        "profile": {
            "email": "john@example.com"
        }
    }
}`

type Profile struct {
    Email string
}

type User struct {
    Name    string
    Profile Profile
}

type Response struct {
    User User
}

response, err := picker.PickFromJson(jsonStr, func(p *picker.Picker) Response {
    user := p.Nested("user")
    profile := user.Nested("profile")

    return Response{
        User: User{
            Name: user.GetString("name"),
            Profile: Profile{
                Email: profile.GetString("email"),
            },
        },
    }
})

// Errors from nested fields show full path: "user.profile.email"
```

### Arrays

```go
// Typed arrays
tags := picker.GetTypedArray[string](p, "tags")     // []string
scores := picker.GetTypedArray[float64](p, "scores") // []float64

// Array of objects
jsonStr := `{
    "users": [
        {"name": "John"},
        {"name": "Jane"}
    ]
}`

type User struct {
    Name string
}

type Response struct {
    Users []User
}

response, err := picker.PickFromJson(jsonStr, func(p *picker.Picker) Response {
    users := p.NestedArray("users")
    userList := make([]User, len(users.Items))

    for i, userPicker := range users.Items {
        userList[i] = User{
            Name: userPicker.GetString("name"),
        }
    }

    return Response{Users: userList}
})
```

### Custom Validation

```go
user, err := picker.Pick(data, func(p *picker.Picker) User {
    age := p.GetInt("age")

    // Add custom validation
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

```go
// Default messages
picker.ErrorMissing = "missing"  // Field not present in JSON
picker.ErrorInvalid = "invalid"  // Field has wrong type

// Customize for your needs
picker.ErrorMissing = "required"
picker.ErrorInvalid = "wrong type"

// Or internationalization
picker.ErrorMissing = "faltante"  // Spanish
picker.ErrorInvalid = "inválido"
```

## HTTP Handler Example

```go
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
    type User struct {
        Name  string
        Email string
        Age   int64
    }

    user, err := picker.PickFromRequestBody(r, func(p *picker.Picker) User {
        return User{
            Name:  p.GetString("name"),
            Email: p.GetString("email"),
            Age:   p.GetInt("age"),
        }
    })

    if err != nil {
        // Return structured validation errors
        errors := picker.Detail(err)
        if len(errors) > 0 {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "errors": errors,
            })
            return
        }

        // JSON parse error
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // user is validated and ready to save
    saveUser(user)
    w.WriteHeader(http.StatusCreated)
}
```

## Why go-picker?

| Feature | `json.Unmarshal` | **go-picker** |
|---------|------------------|---------------|
| Missing fields | ✅ Zero values (`""`, `0`, `nil`) | ❌ Error with field path |
| Wrong types | ✅ Zero values | ❌ Error with field path |
| Validation | ❌ Manual checks needed | ✅ Automatic |
| Error details | ❌ Generic messages | ✅ Field-by-field map |
| All errors | ❌ Stops at first | ✅ Collects all |
| Custom validation | ❌ Write your own | ✅ Built-in `SetError()` |

## License

MIT License - see LICENSE file for details.
