package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	Tag struct {
		Key   string `validate:"lowercase"`
		Value string
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		// unsupported type
		{
			"string",
			ErrUnsupportedType,
		},
		// unsupported validator
		{
			Tag{
				Key:   "foo",
				Value: "bar",
			},
			ValidationErrors{
				ValidationError{
					Field: "Key",
					Err:   UnsupportedValidatorError{Name: "lowercase"},
				},
			},
		},
		// invalid length
		{
			App{
				Version: "123456",
			},
			ValidationErrors{
				ValidationError{
					Field: "Version",
					Err:   LenValidationError{Target: 5, Current: 6},
				},
			},
		},
		// no validation rules
		{
			Token{
				Header:    []byte("X-Token"),
				Payload:   []byte("abc"),
				Signature: []byte("0x1328bcda"),
			},
			nil,
		},
		// not in list of ints
		{
			Response{
				Code: 201,
				Body: "OK",
			},
			ValidationErrors{
				ValidationError{
					Field: "Code",
					Err: InValidationError{
						In:      "200,404,500",
						Current: int64(201),
					},
				},
			},
		},
		// all invalid
		{
			User{
				ID:    "123456789012345678901234567890",
				Age:   51,
				Email: "",
				Role:  "expert",
				Phones: []string{
					"1234567890",
				},
			},
			ValidationErrors{
				ValidationError{
					Field: "ID",
					Err:   LenValidationError{Target: 36, Current: 30},
				},
				ValidationError{
					Field: "Age",
					Err:   MaxValidationError{Max: 50, Current: 51},
				},
				ValidationError{
					Field: "Email",
					Err:   RegexpValidationError{Regexp: "^\\w+@\\w+\\.\\w+$", Current: ""},
				},
				ValidationError{
					Field: "Role",
					Err:   InValidationError{In: "admin,stuff", Current: "expert"},
				},
				ValidationError{
					Field: "Phones",
					Err:   LenValidationError{Target: 11, Current: 10},
				},
			},
		},
		// all valid
		{
			User{
				ID:    "123456789012345678901234567890123456",
				Age:   23,
				Email: "user1@mail.org",
				Role:  "admin",
				Phones: []string{
					"12345678901",
				},
			},
			nil,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)

			if tt.expectedErr == nil && err == nil {
				return
			}

			if tt.expectedErr == nil && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			var expectedErrors ValidationErrors
			if errors.As(tt.expectedErr, &expectedErrors) {
				var actualErrors ValidationErrors
				if !errors.As(err, &actualErrors) {
					t.Errorf("expected error: %v, got: %v", expectedErrors, err)
					return
				}
				if len(expectedErrors) != len(actualErrors) {
					t.Errorf("expected error: %v, got: %v", expectedErrors, err)
					return
				}
				for i := range expectedErrors {
					expectedError := expectedErrors[i]
					actualError := actualErrors[i]
					compareValidationErrors(t, expectedError, actualError)
				}
			} else if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}
		})
	}
}

func compareValidationErrors(t *testing.T, expectedError ValidationError, actualError ValidationError) {
	t.Helper()
	if expectedError.Field != actualError.Field {
		t.Errorf("expected error: %v, got: %v", expectedError, actualError)
	}
	if !errors.Is(actualError.Err, expectedError.Err) {
		t.Errorf("expected error: %v, got: %v", expectedError, actualError)
	}
}
