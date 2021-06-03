package wutils

import "strings"

// TestStrSliceEq returns if two string slices are equal.
func TestStrSliceEq(a, b []string) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestIntSliceEq returns if two int slices are equal.
func TestIntSliceEq(a, b []int) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestStrEqCaseInsensitive returns if two string are equal under case-insensitivity form.
func TestStrEqCaseInsensitive(a, b string) bool {
	return len(a) == len(b) && strings.EqualFold(a, b)
}

