package headers

import (
	"bytes"
	"errors"
	"strings"
)

type Headers map[string]string

var crlf = "\r\n"

var ErrorInvalidFieldLine = errors.New("invalid field line")
var ErrorInvalidKeyChars = errors.New("invalid character in field name")
var ErrorLeadingOrTrailingHeader = errors.New("field name cannot have leading or trailing space")

func NewHeaders() Headers {
	return make(Headers)
}

// IsValidToken checks if the string contains only alphanumeric
// and the specified HTTP token special characters.
func isValidToken(s string) bool {
	if len(s) == 0 {
		return false
	}

	for i := 0; i < len(s); i++ {
		b := s[i]
		switch {
		// Alphanumeric characters
		case b >= 'A' && b <= 'Z':
		case b >= 'a' && b <= 'z':
		case b >= '0' && b <= '9':
		// Allowed special characters
		case b == '!', b == '#', b == '$', b == '%', b == '&', b == '\'',
			b == '*', b == '+', b == '-', b == '.', b == '^', b == '_',
			b == '`', b == '|', b == '~':
			// Character is valid, continue checking
		default:
			// Invalid character found
			return false
		}
	}
	return true
}

// process key
//  1. should throw error if key contains anything other than:
//     Uppercase letters: A-Z
//     Lowercase letters: a-z
//     Digits: 0-9
//     Special characters: !, #, $, %, &, ', *, +, -, ., ^, _, `, |, ~
//  2. should throw error if key has leading or trailing spaces
//  3. should return the lower case of the key to store in headers map
func processKey(key string) (string, error) {
	if strings.ContainsAny(key, " \t") {
		return "", ErrorLeadingOrTrailingHeader
	}

	if !isValidToken(key) {
		return "", ErrorInvalidKeyChars
	}

	return strings.ToLower(key), nil
}

func (h Headers) Parse(data []byte) (int, bool, error) {
	endIdx := bytes.Index(data, []byte(crlf))
	if endIdx == -1 {
		// crlf not found, hence need more data before parsing
		// hence return nil
		return 0, false, nil
	}

	if endIdx == 0 {
		return len(crlf), true, nil
	}

	line := string(data[:endIdx])
	key, value, found := strings.Cut(line, ":")
	if !found {
		return 0, false, ErrorInvalidFieldLine
	}

	processedKey, err := processKey(key)
	if err != nil {
		return 0, false, err
	}
	key = processedKey
	value = strings.TrimSpace(value)

	// append to headers map seperated by comma (,)
	// else add the value
	_, keyPresent := h[key]

	if keyPresent {
		h[key] = h[key] + "," + value
	} else {
		h[key] = value
	}

	return endIdx + len(crlf), false, nil
}
