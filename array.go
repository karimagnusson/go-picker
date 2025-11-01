package picker

import "strconv"

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
