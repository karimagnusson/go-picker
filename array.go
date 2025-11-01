package picker

import (
	"errors"
	"strconv"
)

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
		return NewPicker(map[string]interface{}{})
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
