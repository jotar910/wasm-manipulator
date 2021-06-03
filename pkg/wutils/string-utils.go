package wutils

import (
	"encoding/json"
	"unicode"

	"github.com/sirupsen/logrus"
)

// CapitalizeFirstLetter capitalizes the first character of a string.
func CapitalizeFirstLetter(str string) string {
	if len(str) == 0 {
		return ""
	}
	tmp := []rune(str)
	tmp[0] = unicode.ToUpper(tmp[0])
	return string(tmp)
}

// LowerFirstLetter lowers the first character of a string.
func LowerFirstLetter(str string) string {
	if len(str) == 0 {
		return ""
	}
	tmp := []rune(str)
	tmp[0] = unicode.ToLower(tmp[0])
	return string(tmp)
}

// PrintJSON transforms object into json.
func PrintJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	res, err := json.Marshal(v)
	if err != nil {
		logrus.Fatalf("printing json: %v", err)
	}
	return string(res)
}

// ContainsString returns if an array contains a string value.
func ContainsString(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// CloneString clones a string value.
func CloneString(s string) string {
	b := make([]byte, len(s))
	copy(b, s)
	n := string(b)
	return n
}

// IsAlphaNumeric returns if the character rune is alpha numeric.
func IsAlphaNumeric(s rune) bool {
	return unicode.IsNumber(s) || unicode.IsLetter(s)
}

// IsIdentifier returns if the character rune is valid for identifiers.
func IsIdentifier(s rune) bool {
	return IsAlphaNumeric(s) || s == '_'
}
