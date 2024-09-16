package hw09structvalidator

import (
	"fmt"
	"reflect"
	"strings"
)

type InValidationError struct {
	In      string
	Current interface{}
}

func (v InValidationError) Error() string {
	return fmt.Sprintf("value is not in the list: current=%v", v.Current)
}

func inStringValidator(v string, rawArgs string) error {
	inValues := strings.Split(rawArgs, ",")

	for _, inValue := range inValues {
		if v == inValue {
			return nil
		}
	}

	return InValidationError{
		In:      rawArgs,
		Current: v,
	}
}

func inIntValidator(val int64, rawArgs string) error {
	inValues := strings.Split(rawArgs, ",")

	for _, inValue := range inValues {
		var inInt int64
		_, err := fmt.Sscanf(inValue, "%d", &inInt)
		if err != nil {
			return ValidationRuleParseError{
				Tag:     "in",
				Message: "invalid in value",
			}
		}
		if val == inInt {
			return nil
		}
	}

	return InValidationError{
		In:      rawArgs,
		Current: val,
	}
}

func inValidator(v reflect.Value, rawArgs string) error {
	if v.Kind() == reflect.String {
		return inStringValidator(v.String(), rawArgs)
	}

	if v.Kind() == reflect.Int64 || v.Kind() == reflect.Int {
		return inIntValidator(v.Int(), rawArgs)
	}
	return ErrUnsupportedFieldType
}
