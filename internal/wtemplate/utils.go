package wtemplate

import (
	"regexp"
)

var (
	commentRegex                     = regexp.MustCompile(`(\s*;;[^\n]*|\s*\(;[\s\S]*?;\)\s*)`)
	redundantLeadingSpacesRegex      = regexp.MustCompile(`^[\s\p{Zs}]+|[\s\p{Zs}]+$`)
	redundantLParenthesesSpacesRegex = regexp.MustCompile(`\([\s\p{Zs}]+`)
	redundantRParenthesesSpacesRegex = regexp.MustCompile(`[\s\p{Zs}]+\)`)
	redundantInsideSpacesRegex       = regexp.MustCompile(`[\s\p{Zs}]{2,}`)
)

// ClearString formats code by cleaning redundant characters.
func ClearString(val string) string {
	return redundantInsideSpacesRegex.ReplaceAllString(
		redundantRParenthesesSpacesRegex.ReplaceAllString(
			redundantLParenthesesSpacesRegex.ReplaceAllString(
				redundantLeadingSpacesRegex.ReplaceAllString(
					commentRegex.ReplaceAllString(val, ""),
					""),
				"("),
			")"),
		" ")
}
