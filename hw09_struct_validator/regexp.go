package hw09structvalidator

import (
	"reflect"
	"regexp"
)

type RegexpValidationError struct {
	Regexp  string
	Current string
}

func (v RegexpValidationError) Error() string {
	return "regexp validation failed: " + v.Current
}

func regexpValidator(v reflect.Value, rawArgs string) error {
	if v.Kind() != reflect.String {
		return ErrUnsupportedFieldType
	}
	re, err := regexp.Compile(rawArgs)
	if err != nil {
		return ValidationRuleParseError{
			Tag:     "regexp",
			Message: "invalid regexp value",
		}
	}

	if !re.MatchString(v.String()) {
		return RegexpValidationError{
			Regexp:  rawArgs,
			Current: v.String(),
		}
	}

	return nil
}
