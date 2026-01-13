package picker

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	ErrorMissing = "missing"
	ErrorInvalid = "invalid"
)

func Pick[T any](data map[string]interface{}, fn func(*Picker) T) (T, error) {
	inst := newPicker(data)
	result := fn(inst)
	err := inst.Confirm()
	if err != nil {
		var zero T
		return zero, err
	}
	return result, nil
}

func PickFromJson[T any](jsonStr string, fn func(*Picker) T) (T, error) {
	data, err := ParseJson(jsonStr)
	if err != nil {
		var zero T
		return zero, err
	}
	result, pickerErr := Pick(data, fn)
	if pickerErr != nil {
		var zero T
		return zero, pickerErr
	}
	return result, nil
}

func PickFromRequestBody[T any](r *http.Request, fn func(*Picker) T) (T, error) {
	data, err := ParseRequestBody(r)
	if err != nil {
		var zero T
		return zero, err
	}
	result, pickerErr := Pick(data, fn)
	if pickerErr != nil {
		var zero T
		return zero, pickerErr
	}
	return result, nil
}

func ParseJson(jsonStr string) (map[string]interface{}, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ParseRequestBody(r *http.Request) (map[string]interface{}, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	return ParseJson(string(data))
}

func newPicker(data map[string]interface{}) *Picker {
	return &Picker{
		data:         data,
		errors:       map[string]string{},
		parentPicker: nil,
		parentKey:    "",
	}
}

func newNestedPicker(data map[string]interface{}, parent *Picker, key string) *Picker {
	return &Picker{
		data:         data,
		errors:       map[string]string{},
		parentPicker: parent,
		parentKey:    key,
	}
}

type Picker struct {
	data         map[string]interface{}
	errors       map[string]string
	parentPicker *Picker
	parentKey    string
}

func (p *Picker) addError(key string) {
	var reason string
	if p.HasKey(key) {
		reason = ErrorInvalid
	} else {
		reason = ErrorMissing
	}
	p.SetError(key, reason)
}

func (p *Picker) SetInvalid(key string) {
	p.SetError(key, ErrorInvalid)
}

func (p *Picker) SetError(key string, reason string) {
	if p.parentPicker != nil {
		p.parentPicker.errors[p.parentKey+"."+key] = reason
	} else {
		p.errors[key] = reason
	}
}

func (p *Picker) Confirm() *PickerError {
	if len(p.errors) > 0 {
		return &PickerError{Errors: p.errors}
	}
	return nil
}

func (p *Picker) HasKey(key string) bool {
	_, ok := p.data[key]
	return ok
}

func (p *Picker) Nested(key string) *Picker {
	value, ok := p.data[key].(map[string]interface{})
	if !ok {
		p.addError(key)
		return newNestedPicker(map[string]interface{}{}, p, key)
	}
	return newNestedPicker(value, p, key)
}

func (p *Picker) NestedArray(key string) *NestedPickerArray {
	value, ok := p.data[key].([]interface{})
	if !ok {
		p.addError(key)
		return newNestedPickerArray(p, make([]*Picker, 0))
	}
	pickers := make([]*Picker, len(value))
	for i, item := range value {
		if itemMap, ok := item.(map[string]interface{}); ok {
			pickers[i] = newPicker(itemMap)
		} else {
			p.addError(key)
			pickers[i] = newPicker(map[string]interface{}{})
		}
	}
	return newNestedPickerArray(p, pickers)
}

func (p *Picker) GetString(key string) string {
	value, ok := p.data[key].(string)
	if !ok {
		p.addError(key)
		return ""
	}
	return value
}

func (p *Picker) GetStringOr(key string, fallback string) string {
	value, ok := p.data[key].(string)
	if !ok {
		return fallback
	}
	return value
}

func (p *Picker) GetInt(key string) int64 {
	value, ok := p.data[key].(float64)
	if !ok {
		p.addError(key)
		return 0
	}
	return int64(value)
}

func (p *Picker) GetIntOr(key string, fallback int64) int64 {
	value, ok := p.data[key].(float64)
	if !ok {
		return fallback
	}
	return int64(value)
}

func (p *Picker) GetFloat(key string) float64 {
	value, ok := p.data[key].(float64)
	if !ok {
		p.addError(key)
		return 0
	}
	return value
}

func (p *Picker) GetFloatOr(key string, fallback float64) float64 {
	value, ok := p.data[key].(float64)
	if !ok {
		return fallback
	}
	return value
}

func (p *Picker) GetBool(key string) bool {
	value, ok := p.data[key].(bool)
	if !ok {
		p.addError(key)
		return false
	}
	return value
}

func (p *Picker) GetBoolOr(key string, fallback bool) bool {
	value, ok := p.data[key].(bool)
	if !ok {
		return fallback
	}
	return value
}

func (p *Picker) GetDate(key string) time.Time {
	value, ok := p.data[key].(string)
	if !ok {
		p.addError(key)
		return time.Time{}
	}
	parsedTime, err := time.Parse(time.RFC3339, value)
	if err != nil {
		p.addError(key)
		return time.Time{}
	}
	return parsedTime
}
func (p *Picker) GetDateOr(key string, fallback time.Time) time.Time {
	value, ok := p.data[key].(string)
	if !ok {
		return fallback
	}
	parsedTime, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return fallback
	}
	return parsedTime
}

func (p *Picker) GetObject(key string) map[string]interface{} {
	value, ok := p.data[key].(map[string]interface{})
	if !ok {
		p.addError(key)
		return nil
	}
	return value
}

func (p *Picker) GetObjectOr(key string, fallback map[string]interface{}) map[string]interface{} {
	value, ok := p.data[key].(map[string]interface{})
	if !ok {
		return fallback
	}
	return value
}

func (p *Picker) GetArray(key string) []interface{} {
	value, ok := p.data[key].([]interface{})
	if !ok {
		p.addError(key)
		return nil
	}
	return value
}

func (p *Picker) GetArrayOr(key string, fallback []interface{}) []interface{} {
	value, ok := p.data[key].([]interface{})
	if !ok {
		return fallback
	}
	return value
}

// errors

type PickerError struct {
	Errors map[string]string
}

func (pe *PickerError) Error() string {
	keys := make([]string, 0, len(pe.Errors))
	for key := range pe.Errors {
		keys = append(keys, key)
	}
	return "missing or invalid: " + strings.Join(keys, ", ")
}

func Detail(err error) map[string]string {
	if pickerErr, ok := err.(*PickerError); ok {
		return pickerErr.Errors
	}
	return map[string]string{}
}

// array

type NestedPickerArray struct {
	nestedKey string
	parent    *Picker
	Items     []*Picker
}

func newNestedPickerArray(parent *Picker, items []*Picker) *NestedPickerArray {
	return &NestedPickerArray{
		parent: parent,
		Items:  items,
	}
}

func (npa *NestedPickerArray) GetItem(index int) *Picker {
	if index < 0 || index >= len(npa.Items) {
		npa.parent.addError(npa.nestedKey + "[" + strconv.Itoa(index) + "]")
		return newPicker(map[string]interface{}{})
	}
	return npa.Items[index]
}

func convert[T any](items []interface{}) ([]T, error) {
	result := make([]T, 0, len(items))
	for _, item := range items {
		if typedItem, ok := item.(T); ok {
			result = append(result, typedItem)
		} else {
			return nil, errors.New("error")
		}
	}
	return result, nil
}

func GetTypedArray[T any](p *Picker, key string) []T {
	value, ok := p.data[key].([]interface{})
	if !ok {
		p.addError(key)
		return []T{}
	}

	result, err := convert[T](value)
	if err != nil {
		p.addError(key)
		return []T{}
	}

	return result
}
