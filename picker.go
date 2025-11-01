package picker

import (
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"
)

type ValueType int

const (
	ValueTypeString ValueType = iota
	ValueTypeInt
	ValueTypeFloat
	ValueTypeBool
	ValueTypeBigInt
	ValueTypeBigFloat
	ValueTypeBigRat
)

type Picker struct {
	data         map[string]interface{}
	errorKeys    []string
	isNested     bool
	nestedPicker *Picker
	nestedKey    string
}

func NewPickerFromJson(jsonStr string) (*Picker, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return nil, err
	}
	return NewPicker(data), nil
}

func NewPickerFromRequest(r *http.Request) (*Picker, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	return NewPickerFromJson(string(body))
}

func NewPicker(data map[string]interface{}) *Picker {
	return &Picker{
		data:         data,
		errorKeys:    []string{},
		isNested:     false,
		nestedPicker: nil,
	}
}

func newNestedPicker(data map[string]interface{}, parent *Picker, key string) *Picker {
	return &Picker{
		data:         data,
		errorKeys:    []string{},
		isNested:     true,
		nestedPicker: parent,
		nestedKey:    key,
	}
}

func (p *Picker) addError(key string) {
	if p.isNested {
		p.nestedPicker.addError(p.nestedKey + "." + key)
	} else {
		p.errorKeys = append(p.errorKeys, key)
	}
}

func (p *Picker) GetNewPicker(key string) *Picker {
	value, ok := p.data[key].(map[string]interface{})
	if !ok {
		p.addError(key)
		return NewPicker(map[string]interface{}{})
	}
	return NewPicker(value)
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
			pickers[i] = NewPicker(itemMap)
		} else {
			p.addError(key)
			pickers[i] = NewPicker(map[string]interface{}{})
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
	value, ok := p.data[key].(int64)
	if !ok {
		p.addError(key)
		return 0
	}
	return value
}

func (p *Picker) GetIntOr(key string, fallback int64) int64 {
	value, ok := p.data[key].(int64)
	if !ok {
		return fallback
	}
	return value
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

func (p *Picker) GetBigInt(key string) *big.Int {
	value, ok := p.data[key].(*big.Int)
	if !ok {
		p.addError(key)
		return nil
	}
	return value
}

func (p *Picker) GetBigIntOr(key string, fallback *big.Int) *big.Int {
	value, ok := p.data[key].(*big.Int)
	if !ok {
		return fallback
	}
	return value
}

func (p *Picker) GetBigFloat(key string) *big.Float {
	value, ok := p.data[key].(*big.Float)
	if !ok {
		p.addError(key)
		return nil
	}
	return value
}

func (p *Picker) GetBigFloatOr(key string, fallback *big.Float) *big.Float {
	value, ok := p.data[key].(*big.Float)
	if !ok {
		return fallback
	}
	return value
}

func (p *Picker) GetBigRat(key string) *big.Rat {
	value, ok := p.data[key].(*big.Rat)
	if !ok {
		p.addError(key)
		return nil
	}
	return value
}

func (p *Picker) GetBigRatOr(key string, fallback *big.Rat) *big.Rat {
	value, ok := p.data[key].(*big.Rat)
	if !ok {
		return fallback
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
	value, ok := p.data[key].(time.Time)
	if !ok {
		return fallback
	}
	return value
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

func (p *Picker) GetTypedArray(key string, valueType ValueType) interface{} {
	value, ok := p.data[key].([]interface{})
	if !ok {
		p.addError(key)
		return nil
	}

	result, err := typedArray(value, valueType)
	if err != nil {
		p.addError(key)
		return nil
	}

	return result
}

func (p *Picker) HasErrors() bool {
	return len(p.errorKeys) > 0
}

func (p *Picker) Confirm() error {
	if p.isNested {
		return errors.New("cannot confirm a nested picker directly")
	}
	if len(p.errorKeys) > 0 {
		return errors.New("errors in keys: " + strings.Join(p.errorKeys, ", "))
	}
	return nil
}

func (p *Picker) ErrorKeys() []string {
	return p.errorKeys
}

func (p *Picker) HasKey(key string) bool {
	_, ok := p.data[key]
	return ok
}

func (p *Picker) Keys() []string {
	keys := make([]string, 0, len(p.data))
	for k := range p.data {
		keys = append(keys, k)
	}
	return keys
}

func (p *Picker) AddKey(key string, value interface{}) {
	p.data[key] = value
}

func (p *Picker) DelKey(key string) {
	delete(p.data, key)
}

func (p *Picker) GetData() map[string]interface{} {
	return p.data
}

func (p *Picker) Extend(other *Picker) {
	for k, v := range other.data {
		p.data[k] = v
	}
}

func (p *Picker) Copy() *Picker {
	data := make(map[string]interface{})
	for k, v := range p.data {
		data[k] = v
	}
	return NewPicker(data)
}

func (p *Picker) FlatCopy() *Picker {
	flat := make(map[string]interface{})
	var flatten func(prefix string, data map[string]interface{})
	flatten = func(prefix string, data map[string]interface{}) {
		for k, v := range data {
			switch v := v.(type) {
			case map[string]interface{}:
				flatten(prefix+k+".", v)
			default:
				flat[prefix+k] = v
			}
		}
	}
	flatten("", p.data)
	return NewPicker(flat)
}

func (p *Picker) ToJsonString() string {
	jsonBytes, err := json.Marshal(p.data)
	if err != nil {
		return "{}"
	}
	return string(jsonBytes)
}

func (p *Picker) ToPrettyJsonString() string {
	jsonBytes, err := json.MarshalIndent(p.data, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(jsonBytes)
}
