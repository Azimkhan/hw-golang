package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(packed string) (string, error) {
	builder := strings.Builder{}

	if len(packed) == 0 {
		return "", nil
	}

	// hold the character to be repeated
	var char rune = -1
	var escapeNext bool

	for _, current := range packed {
		n, err := strconv.Atoi(string(current))
		isNum := err == nil

		// case 4 - previous char was a backslash
		if escapeNext {
			// if current not a number or backslash return error
			if !isNum && current != '\\' {
				return "", ErrInvalidString
			}
			char = current
			escapeNext = false
			continue
		}

		// case 5 - current char is a backslash
		if current == '\\' {
			if char != -1 {
				builder.WriteRune(char)
			}
			escapeNext = true
			continue
		}

		if char != -1 && isNum {
			// case 1 - previous char is a letter and current is a number
			if n > 0 {
				builder.WriteString(strings.Repeat(string(char), n))
			}
			char = -1
			continue
		}
		if char != -1 {
			// case 2 - previous char is a letter and current is a letter
			builder.WriteRune(char)
			char = current
			continue
		}

		// case 3 - previous char is a number and current is a number
		if isNum {
			return "", ErrInvalidString
		}
		char = current
	}

	// case 6 - last char is a letter
	if char != -1 {
		builder.WriteRune(char)
	}
	return builder.String(), nil
}
