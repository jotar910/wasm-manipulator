package wkeyword

type KeywordType int

const (
	KeywordTypeUnknown KeywordType = iota
	KeywordTypeString
	KeywordTypeObject
	KeywordTypeTemplate
)

// KeywordsMap is implemented by any value that can retrieve a keyword.
type KeywordsMap interface {
	Is(k string) KeywordType
	Get(k string) (interface{}, KeywordType, bool)
}

// StringValuesMap represents a keyword map of strings.
type StringValuesMap map[string]string

// NewStringValuesMap is a constructor for StringValuesMap.
func NewStringValuesMap(values ...[]string) StringValuesMap {
	m := make(StringValuesMap)
	for _, v := range values {
		m[v[0]] = v[1]
	}
	return m
}

// Is returns the keyword type.
// KeywordTypeUnknown if not found.
func (m StringValuesMap) Is(k string) KeywordType {
	if _, ok := m[k]; ok {
		return KeywordTypeString
	}
	return KeywordTypeUnknown
}

// Get return the keyword if presented in the map.
func (m StringValuesMap) Get(k string) (interface{}, KeywordType, bool) {
	if v, ok := m[k]; ok {
		return v, KeywordTypeString, ok
	}
	return nil, KeywordTypeUnknown, false
}

// ObjectValuesMap represents a keyword map of objects.
type ObjectValuesMap map[string]Object

// KeyValueObject is a key value structure for objects.
type KeyValueObject struct {
	key   string
	value Object
}

// NewKeyValueObject is a constructor for KeyValueObject.
func NewKeyValueObject(k string, v Object) KeyValueObject {
	return KeyValueObject{k, v}
}

// NewObjectValuesMap is a constructor for ObjectValuesMap.
func NewObjectValuesMap(values ...KeyValueObject) ObjectValuesMap {
	m := make(ObjectValuesMap)
	for _, v := range values {
		m[v.key] = v.value
	}
	return m
}

// Is returns the keyword type.
// KeywordTypeUnknown if not found.
func (m ObjectValuesMap) Is(k string) KeywordType {
	if _, ok := m[k]; ok {
		return KeywordTypeObject
	}
	return KeywordTypeUnknown
}

// Get return the keyword if presented in the map.
func (m ObjectValuesMap) Get(k string) (interface{}, KeywordType, bool) {
	if v, ok := m[k]; ok {
		return v, KeywordTypeObject, ok
	}
	return nil, KeywordTypeUnknown, false
}
