package hw09structvalidator

import (
	"fmt"
	"reflect"
)

type fieldValidator func(v reflect.Value, rawArgs string) error

var fieldValidators = map[string]fieldValidator{
	"len":    wrapValidator(lenValidator),
	"min":    wrapValidator(minValidator),
	"max":    wrapValidator(maxValidator),
	"regexp": wrapValidator(regexpValidator),
	"in":     wrapValidator(inValidator),
}

// wrapValidator is a function that wraps a fieldValidator to handle arrays and slices if needed.
func wrapValidator(validator fieldValidator) fieldValidator {
	return func(v reflect.Value, rawArgs string) error {
		if v.Kind() == reflect.Array || v.Kind() == reflect.Slice {
			for i := range v.Len() {
				if err := validator(v.Index(i), rawArgs); err != nil {
					return fmt.Errorf("index %d: %w", i, err)
				}
			}
		} else {
			return validator(v, rawArgs)
		}
		return nil
	}
}
