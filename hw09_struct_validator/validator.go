package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	ErrUnsupportedType      = errors.New("unsupported type")
	ErrUnsupportedFieldType = errors.New("unsupported field type")
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var errStr strings.Builder
	for _, err := range v {
		errStr.WriteString(fmt.Sprintf("field %s: %s\n", err.Field, err.Err))
	}
	return errStr.String()
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("field %s: %s", v.Field, v.Err)
}

type ValidationRuleParseError struct {
	Tag     string
	Message string
}

func (v ValidationRuleParseError) Error() string {
	return fmt.Sprintf("tag %s: %s", v.Tag, v.Message)
}

type UnsupportedValidatorError struct {
	Name string
}

func (v UnsupportedValidatorError) Error() string {
	return fmt.Sprintf("validator %s not found", v.Name)
}

func Validate(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		return ErrUnsupportedType
	}

	var validationErrors ValidationErrors

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)

		// Check if the field is public
		if !(field.PkgPath == "") {
			continue
		}

		// Check if the field has a validate tag
		tagValue := field.Tag.Get("validate")
		if tagValue == "" {
			continue
		}

		tagParts := strings.Split(tagValue, "|")

		for _, tagPart := range tagParts {
			// split the tag into key and value
			parts := strings.SplitN(tagPart, ":", 2)

			validatorKey := parts[0]
			validatorArgs := ""
			if len(parts) > 1 {
				validatorArgs = parts[1]
			}

			// Check if the field has a validator
			validator, ok := fieldValidators[validatorKey]
			if !ok {
				validationErrors = append(validationErrors, ValidationError{
					Field: field.Name,
					Err:   UnsupportedValidatorError{validatorKey},
				})
				continue
			}

			// Validate the field
			err := validator(val.Field(i), validatorArgs)
			if err != nil {
				validationErrors = append(validationErrors, ValidationError{
					Field: field.Name,
					Err:   err,
				})
			}
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}

	return nil
}
