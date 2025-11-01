package picker

import (
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"
)

// PickerToStruct maps Picker data to struct using Picker methods and reflection
func PickToStruct(jsonStr string, target interface{}) error {
	picker, err := NewPickerFromJson(jsonStr)
	if err != nil {
		return err
	}
	return pickerToStructWithDepth(picker, target, 0, 10) // Max depth of 10
}

// pickerToStructWithDepth maps Picker data to struct with recursion depth tracking
func pickerToStructWithDepth(picker *Picker, target interface{}, currentDepth, maxDepth int) error {
	// Prevent infinite recursion
	if currentDepth >= maxDepth {
		return fmt.Errorf("maximum recursion depth (%d) exceeded", maxDepth)
	}
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to a struct")
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Get JSON tag - require explicit JSON tags
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" {
			return fmt.Errorf("field %s.%s missing required json tag", typ.Name(), fieldType.Name)
		}
		if jsonTag == "-" {
			continue // Explicitly ignored field
		}

		// Remove options like ",omitempty"
		if idx := strings.Index(jsonTag, ","); idx != -1 {
			jsonTag = jsonTag[:idx]
		}

		if !field.CanSet() {
			continue
		}

		// Map based on field type - use strict methods that track errors
		switch field.Kind() {
		case reflect.String:
			field.SetString(picker.GetString(jsonTag))
		case reflect.Int, reflect.Int64:
			// JSON numbers come as float64, so try that first
			errorCountBefore := len(picker.errorKeys)
			floatVal := picker.GetFloat(jsonTag)
			if len(picker.errorKeys) == errorCountBefore {
				// No new error, conversion successful
				field.SetInt(int64(floatVal))
			} else {
				// Remove the error from GetFloat and try GetInt
				picker.errorKeys = picker.errorKeys[:errorCountBefore]
				field.SetInt(picker.GetInt(jsonTag))
			}
		case reflect.Float64:
			field.SetFloat(picker.GetFloat(jsonTag))
		case reflect.Bool:
			field.SetBool(picker.GetBool(jsonTag))
		case reflect.Struct:
			if fieldType.Type == reflect.TypeOf(time.Time{}) {
				// Special handling for time.Time
				field.Set(reflect.ValueOf(picker.GetDate(jsonTag)))
			} else {
				// Handle other nested structs (embedded, not pointers)
				if picker.HasKey(jsonTag) {
					nestedPicker := picker.Nested(jsonTag)
					newStruct := reflect.New(field.Type())
					if err := pickerToStructWithDepth(nestedPicker, newStruct.Interface(), currentDepth+1, maxDepth); err != nil {
						return fmt.Errorf("failed to convert nested struct %s.%s: %w", typ.Name(), fieldType.Name, err)
					}
					field.Set(newStruct.Elem())
				}
			}
		case reflect.Slice:
			// Handle slices
			if picker.HasKey(jsonTag) {
				sliceType := field.Type()
				elemType := sliceType.Elem()
				
				// Handle different slice element types
				if elemType.Kind() == reflect.Struct {
					// Handle slice of structs using NestedArray
					pickerArray := picker.NestedArray(jsonTag)
					newSlice := reflect.MakeSlice(sliceType, len(pickerArray.Items), len(pickerArray.Items))
					
					for i, elemPicker := range pickerArray.Items {
						elemVal := newSlice.Index(i)
						newElem := reflect.New(elemType)
						if err := pickerToStructWithDepth(elemPicker, newElem.Interface(), currentDepth+1, maxDepth); err != nil {
							return fmt.Errorf("failed to convert slice element %d in %s.%s: %w", i, typ.Name(), fieldType.Name, err)
						}
						elemVal.Set(newElem.Elem())
					}
					field.Set(newSlice)
				} else {
					// Handle primitive type slices using GetTypedArray
					var arrayValue reflect.Value
					
					switch elemType.Kind() {
					case reflect.String:
						result := GetTypedArray[string](picker, jsonTag)
						if len(result) > 0 {
							arrayValue = reflect.ValueOf(result)
						}
					case reflect.Int64:
						result := GetTypedArray[int64](picker, jsonTag)
						if len(result) > 0 {
							arrayValue = reflect.ValueOf(result)
						}
					case reflect.Float64:
						result := GetTypedArray[float64](picker, jsonTag)
						if len(result) > 0 {
							arrayValue = reflect.ValueOf(result)
						}
					case reflect.Bool:
						result := GetTypedArray[bool](picker, jsonTag)
						if len(result) > 0 {
							arrayValue = reflect.ValueOf(result)
						}
					default:
						// Handle pointer types for big numbers
						switch elemType {
						case reflect.TypeOf((*big.Int)(nil)):
							result := GetTypedArray[*big.Int](picker, jsonTag)
							if len(result) > 0 {
								arrayValue = reflect.ValueOf(result)
							}
						case reflect.TypeOf((*big.Float)(nil)):
							result := GetTypedArray[*big.Float](picker, jsonTag)
							if len(result) > 0 {
								arrayValue = reflect.ValueOf(result)
							}
						case reflect.TypeOf((*big.Rat)(nil)):
							result := GetTypedArray[*big.Rat](picker, jsonTag)
							if len(result) > 0 {
								arrayValue = reflect.ValueOf(result)
							}
						default:
							return fmt.Errorf("unsupported slice element type %s in %s.%s", elemType.String(), typ.Name(), fieldType.Name)
						}
					}
					
					if arrayValue.IsValid() {
						field.Set(arrayValue)
					}
				}
			}
		case reflect.Ptr:
			if field.Type().Elem().Kind() == reflect.Struct {
				// Handle pointer to nested struct
				if picker.HasKey(jsonTag) {
					newVal := reflect.New(field.Type().Elem())
					field.Set(newVal)

					// Get nested picker and recursively populate the struct
					nestedPicker := picker.Nested(jsonTag)
					if err := pickerToStructWithDepth(nestedPicker, newVal.Interface(), currentDepth+1, maxDepth); err != nil {
						return fmt.Errorf("failed to convert nested struct pointer %s.%s: %w", typ.Name(), fieldType.Name, err)
					}
				}
			} else if field.Type() == reflect.TypeOf((*big.Int)(nil)) {
				// Handle *big.Int
				field.Set(reflect.ValueOf(picker.GetBigInt(jsonTag)))
			} else if field.Type() == reflect.TypeOf((*big.Float)(nil)) {
				// Handle *big.Float
				field.Set(reflect.ValueOf(picker.GetBigFloat(jsonTag)))
			} else if field.Type() == reflect.TypeOf((*big.Rat)(nil)) {
				// Handle *big.Rat
				field.Set(reflect.ValueOf(picker.GetBigRat(jsonTag)))
			}
		}
	}

	if err := picker.Confirm(); err != nil {
		return fmt.Errorf("errors occurred during mapping: %w", err)
	}

	return nil
}
