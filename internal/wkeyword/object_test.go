package wkeyword

import (
	"fmt"
	"testing"
)

type TestingInput struct {
	Obj struct {
		Prop1 int
		Prop2 string
	}            // KwObject
	Ar  []string // KwArray
	Str string   // KwPrimitive
}

func newTestingInput() Object {
	return &KwObject{
		TestingInput{
			Obj: struct {
				Prop1 int
				Prop2 string
			}{
				Prop1: 123,
			},
			Ar:  []string{"Testing", "String", "Slice"},
			Str: "TestingString",
		},
	}
}

func TestObject(t *testing.T) {
	fmt.Println(newTestingInput().Value())
	fmt.Println(newTestingInput().Prop("Obj").Value())
	fmt.Println(newTestingInput().Prop("Obj").Prop("Prop1").Value())
	fmt.Println(newTestingInput().Prop("Ar").Value())
	fmt.Println(newTestingInput().Prop("Ar").Index(2).Value())
	fmt.Println(newTestingInput().Prop("Str").Value())
	fmt.Println(newTestingInput().Prop("Str").Index(3).Value())
}
