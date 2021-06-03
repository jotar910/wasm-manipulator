package wtemplate

import "strings"

// SearchValue is the value for a template search result variable.
type SearchValue struct {
	Key   string             `json:"key"`
	Templ string             `json:"template"`
	Iter  []*SearchIteration `json:"iterations"`
}

// NewSearchValue is a constructor for SearchValue.
func NewSearchValue(id, templKey string, iterations []*SearchIteration) *SearchValue {
	return &SearchValue{
		Key:   id,
		Templ: templKey,
		Iter:  iterations,
	}
}

// Get returns the search value that matches the variable key.
func (s *SearchValue) Get(key string) (*SearchValue, bool) {
	if s.Key == key {
		return s, true
	}
	for _, it := range s.Iter {
		if it.Values == nil {
			continue
		}
		for _, sv := range it.Values {
			if v, ok := sv.Get(key); ok {
				return v, true
			}
		}
	}
	return nil, false
}

// Remove removes a variable from the search value.
func (s *SearchValue) Remove(key string) *SearchValue {
	if child, removed := removeSearch(s, key); removed && child == nil {
		return nil
	}
	return s
}

// Replace replaces the variable value from the search value.
func (s *SearchValue) Replace(old, new *ReplaceOpArg) *SearchValue {
	res, _ := replaceSearch(s, old, new)
	return res
}

// Clone clones the search result value.
func (s *SearchValue) Clone() *SearchValue {
	cS := *s
	cS.Iter = []*SearchIteration{}
	for _, iter := range s.Iter {
		cS.Iter = append(cS.Iter, iter.Clone())
	}
	return &cS
}

// SearchIteration is an iteration for some search result variable.
type SearchIteration struct {
	Found  string         `json:"found"`
	Values []*SearchValue `json:"values"`
}

// Clone clones the search result iteration.
func (si *SearchIteration) Clone() *SearchIteration {
	cSi := *si
	cSi.Values = []*SearchValue{}
	if len(si.Values) > 0 {
		for _, val := range si.Values {
			cSi.Values = append(cSi.Values, val.Clone())
		}
	}
	return &cSi
}

// removeSearch removes a variable (and its value) from a search result.
func removeSearch(s *SearchValue, key string) (*SearchValue, bool) {
	if s.Key == key {
		return nil, true
	}
	for _, it := range s.Iter {
		if len(it.Values) == 0 {
			continue
		}
		accumIndex := 0
		for i, value := range it.Values {
			if len(value.Iter) == 0 {
				continue
			}
			start := accumIndex + strings.Index(it.Found[accumIndex:], value.Iter[0].Found)
			end := start + len(value.Iter[0].Found)
			accumIndex = end
			if child, removed := removeSearch(value, key); removed {
				if child == nil {
					if i == len(it.Values)-1 {
						it.Values = it.Values[:i]
					} else {
						it.Values = append(it.Values[:i], it.Values[i+1:]...)
					}
					if len(it.Values) == 0 {
						it.Values = nil
					}
					it.Found = ClearString(it.Found[:start] + it.Found[end:])
				} else {
					it.Found = ClearString(it.Found[:start] + value.Iter[0].Found + it.Found[end:])
				}
				return value, true
			}
		}
	}
	return s, false
}

// replaceSearch replaces a variable on a search result.
func replaceSearch(s *SearchValue, old, new *ReplaceOpArg) (*SearchValue, bool) {
	value := new.value
	if new.isReference {
		result, ok := s.Get(value)
		if !ok || len(result.Iter) == 0 {
			return s, false
		}
		value = result.Iter[0].Found
	}
	if old.isReference {
		return replaceByReference(s, old.value, value)
	}
	return replaceByString(s, old.value, value)
}

// replaceByReference replaces a variable of type reference on a search result.
func replaceByReference(s *SearchValue, key string, newValue string) (*SearchValue, bool) {
	if len(s.Iter) == 0 {
		return s, false
	}
	if s.Key == key {
		s.Iter[0].Found = newValue
		s.Iter[0].Values = nil
		return s, true
	}
	it := s.Iter[0]
	if len(it.Values) == 0 {
		return s, false
	}

	var changed bool
	var accumIndex int
	for valueI, value := range it.Values {
		if len(value.Iter) == 0 {
			continue
		}
		start := accumIndex + strings.Index(it.Found[accumIndex:], value.Iter[0].Found)
		end := start + len(value.Iter[0].Found)
		if res, ok := replaceByReference(value, key, newValue); ok {
			it.Found = it.Found[:start] + res.Iter[0].Found + it.Found[end:]
			it.Values[valueI] = res
			changed = true
		}
		accumIndex = start + len(value.Iter[0].Found)
	}
	return s, changed
}

// replaceByString replaces a variable of type string on a search result.
func replaceByString(s *SearchValue, old string, new string) (*SearchValue, bool) {
	if len(s.Iter) == 0 {
		return s, false
	}
	it := s.Iter[0]
	if len(it.Values) == 0 {
		prevFoundValue := it.Found
		it.Found = strings.ReplaceAll(it.Found, old, new)
		return s, prevFoundValue != it.Found
	}

	var changed bool
	var accumIndex int
	for valueI, value := range it.Values {
		if len(value.Iter) == 0 {
			continue
		}
		start := accumIndex + strings.Index(it.Found[accumIndex:], value.Iter[0].Found)
		end := start + len(value.Iter[0].Found)
		if res, ok := replaceByString(value, old, new); ok {
			it.Found = it.Found[:start] + res.Iter[0].Found + it.Found[end:]
			it.Values[valueI] = res
			changed = true
		}
		accumIndex = start + len(value.Iter[0].Found)
	}
	return s, changed
}
