package picker

import (
	"errors"
	"math/big"
)

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

func typedArray(items []interface{}, valueType ValueType) (interface{}, error) {
	switch valueType {
	case ValueTypeString:
		return convert[string](items)
	case ValueTypeInt:
		return convert[int64](items)
	case ValueTypeFloat:
		return convert[float64](items)
	case ValueTypeBool:
		return convert[bool](items)
	case ValueTypeBigInt:
		return convert[*big.Int](items)
	case ValueTypeBigFloat:
		return convert[*big.Float](items)
	case ValueTypeBigRat:
		return convert[*big.Rat](items)
	default:
		return nil, errors.New("error")
	}
}