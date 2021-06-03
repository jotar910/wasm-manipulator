package waspect

import (
	"sync"

	"joao/wasm-manipulator/internal/wkeyword"
)

// contextVariables contains all the context zone data.
// it is the entry point to obtain the zone values as keywords.
type contextVariables struct {
	zone      *contextVariablesZone
	zoneCache map[string]string
	mutex     sync.Mutex
}

// newContextVariables is a constructor for contextVariables.
func newContextVariables(zone *contextVariablesZone) *contextVariables {
	return &contextVariables{
		zone:      zone,
		zoneCache: make(map[string]string),
	}
}

// Is returns the type of keyword of some context variable.
func (cr *contextVariables) Is(k string) wkeyword.KeywordType {
	for z := cr.zone; z != nil; z = z.contextVariablesZone {
		if _, ok := z.VariablesMap[k]; ok {
			return wkeyword.KeywordTypeString
		}
		if _, ok := z.FunctionsMap[k]; ok {
			return wkeyword.KeywordTypeString
		}
	}
	return wkeyword.KeywordTypeUnknown
}

// Get returns the context variable, for a given key, as a keyword value.
func (cr *contextVariables) Get(k string) (interface{}, wkeyword.KeywordType, bool) {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()
	if cacheVal, ok := cr.zoneCache[k]; ok {
		return cacheVal, wkeyword.KeywordTypeString, true
	}
	if val, ok := cr.zone.Value(k); ok {
		cr.zoneCache[k] = val
		return val, wkeyword.KeywordTypeString, true
	}
	return nil, wkeyword.KeywordTypeUnknown, false
}

// functionZoneValue represents the value of a function for the context zone.
type functionZoneValue struct {
	index int
	name  string
}

// newFunctionZone is a constructor for functionZoneValue.
func newFunctionZone(index int, name string) *functionZoneValue {
	return &functionZoneValue{
		index: index,
		name:  name,
	}
}

// contextVariablesZone represents a context zone that contains the value for variables and function.
// zones have depth and so it may have a reference of its parent zone.
type contextVariablesZone struct {
	*contextVariablesZone
	VariablesMap map[string]string
	FunctionsMap map[string]*functionZoneValue
}

// newContextVariablesZone is a constructor for contextVariablesZone.
func newContextVariablesZone(parent *contextVariablesZone) *contextVariablesZone {
	return &contextVariablesZone{
		contextVariablesZone: parent,
		VariablesMap:         make(map[string]string),
		FunctionsMap:         make(map[string]*functionZoneValue),
	}
}

// AddVariable adds a variable value to the zone.
func (zone *contextVariablesZone) AddVariable(k, value string) {
	zone.VariablesMap[k] = value
}

// AddFunction adds a function value to the zone.
func (zone *contextVariablesZone) AddFunction(k string, value *functionZoneValue) {
	zone.FunctionsMap[k] = value
}

// Value returns the context value for a given key.
func (zone *contextVariablesZone) Value(k string) (string, bool) {
	for z := zone; z != nil; z = z.contextVariablesZone {
		if val, ok := z.VariablesMap[k]; ok {
			return val, true
		}
		if val, ok := z.FunctionsMap[k]; ok {
			return val.name, true
		}
	}
	return "", false
}
