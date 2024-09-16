package hw09structvalidator

import (
	"fmt"
	"reflect"
)

type LenValidationError struct {
	Target  int
	Current int
}

func (v LenValidationError) Error() string {
	return fmt.Sprintf("invalid length: target=%d, current=%d", v.Target, v.Current)
}

func lenValidator(v reflect.Value, rawArgs string) error {
	if v.Kind() != reflect.String {
		return ErrUnsupportedFieldType
	}
	var parsedLen int
	_, err := fmt.Sscanf(rawArgs, "%d", &parsedLen)
	if err != nil {
		return ValidationRuleParseError{
			Tag:     "len",
			Message: "invalid len value",
		}
	}

	if v.Len() != parsedLen {
		return LenValidationError{
			Target:  parsedLen,
			Current: v.Len(),
		}
	}
	return nil
}
