package hw09structvalidator

import (
	"fmt"
	"reflect"
)

type MinValidationError struct {
	Min     int64
	Current int64
}

type MaxValidationError struct {
	Max     int64
	Current int64
}

func (v MinValidationError) Error() string {
	return fmt.Sprintf("value is out of range: min=%d, current=%d", v.Min, v.Current)
}

func (v MaxValidationError) Error() string {
	return fmt.Sprintf("value is out of range: max=%d, current=%d", v.Max, v.Current)
}

func minValidator(v reflect.Value, rawArgs string) error {
	var err error
	var minVal int64

	if v.Kind() != reflect.Int {
		return ValidationRuleParseError{
			Tag:     "min",
			Message: "unsupported field type",
		}
	}

	_, err = fmt.Sscanf(rawArgs, "%d", &minVal)
	if err != nil {
		return ValidationRuleParseError{
			Tag:     "min",
			Message: "invalid min value",
		}
	}

	val := v.Int()
	if val < minVal {
		return MinValidationError{
			Min:     minVal,
			Current: val,
		}
	}

	return nil
}

func maxValidator(v reflect.Value, rawArgs string) error {
	var err error
	var maxVal int64

	if v.Kind() != reflect.Int {
		return ValidationRuleParseError{
			Tag:     "min",
			Message: "unsupported field type",
		}
	}

	_, err = fmt.Sscanf(rawArgs, "%d", &maxVal)
	if err != nil {
		return ValidationRuleParseError{
			Tag:     "max",
			Message: "invalid min value",
		}
	}

	val := v.Int()
	if val > maxVal {
		return MaxValidationError{
			Max:     maxVal,
			Current: val,
		}
	}

	return nil
}
