package wkeyword

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/structs"
	"github.com/sirupsen/logrus"

	"joao/wasm-manipulator/pkg/wutils"
)

// Object is implemented by any keyword object.
// it provides some object oriented features to simulate a real object.
type Object interface {
	fmt.Stringer
	Value() string
	Slice() []Object
	StringSlice() []string
	Join(Object)
	Prop(string) Object
	Index(int) Object
	Len() int
}

// KwObject is the object type for a keyword object.
type KwObject struct {
	Val interface{} `json:"-"`
}

// NewKwObject is a constructor for KwObject.
func NewKwObject(v interface{}) *KwObject {
	return &KwObject{v}
}

// Value returns the value of this object when printed.
func (to *KwObject) Value() string {
	return fmt.Sprintf("%+v", to.Val)
}

// String returns the value formatted as string.
func (to KwObject) String() string {
	rv, err := to.assertObject()
	if err != nil {
		logrus.Fatalf("accessing string format in non-object element: %v", err)
	}
	if rv.Kind() == reflect.Map {
		return to.stringMap(rv)
	}
	return to.stringStruct(rv)
}

func (to KwObject) stringMap(rv reflect.Value) string {
	var res []string
	rKeys := rv.MapKeys()
	for _, key := range rKeys {
		res = append(res, fmt.Sprintf("%s:%s", returnValue(key).String(), returnValue(rv.MapIndex(key)).String()))
	}
	return fmt.Sprintf("{%s}", strings.Join(res, ","))
}

func (to KwObject) stringStruct(rv reflect.Value) string {
	var res []string
	for i := rv.NumField() - 1; i > -1; i-- {
		res = append(res, fmt.Sprintf("%s:%s", rv.Type().Field(i).Name, returnValue(rv.Field(i)).String()))
	}
	return fmt.Sprintf("{%s}", strings.Join(res, ","))
}

// Slice returns the value formatted as a slice of objects.
func (to KwObject) Slice() []Object {
	rv, err := to.assertObject()
	if err != nil {
		logrus.Fatalf("accessing property in non-object element: %v", err)
	}
	if rv.Kind() == reflect.Map {
		return to.sliceMap(rv)
	}
	return to.sliceStruct(rv)
}

// StringSlice returns the value formatted as a slice of strings.
func (to KwObject) StringSlice() []string {
	rv, err := to.assertObject()
	if err != nil {
		logrus.Fatalf("accessing property in non-object element: %v", err)
	}
	if rv.Kind() == reflect.Map {
		return to.stringSliceMap(rv)
	}
	return to.stringSliceStruct(rv)
}

func (to KwObject) sliceMap(rv reflect.Value) []Object {
	var res []Object
	keys := rv.MapKeys()
	for _, k := range keys {
		res = append(res, returnValue(rv.MapIndex(k)))
	}
	return res
}

func (to KwObject) stringSliceMap(rv reflect.Value) []string {
	var res []string
	keys := rv.MapKeys()
	for _, k := range keys {
		res = append(res, returnValue(rv.MapIndex(k)).String())
	}
	return res
}

func (to KwObject) sliceStruct(rv reflect.Value) []Object {
	var res []Object
	rNumFields := rv.NumField()
	for i := 0; i < rNumFields; i++ {
		res = append(res, returnValue(rv.Field(i)))
	}
	return res
}

func (to KwObject) stringSliceStruct(rv reflect.Value) []string {
	var res []string
	rNumFields := rv.NumField()
	for i := 0; i < rNumFields; i++ {
		res = append(res, returnValue(rv.Field(i)).String())
	}
	return res
}

// KeysSlice returns the value formatted as a slice of keys.
func (to KwObject) KeysSlice() []string {
	rv, err := to.assertObject()
	if err != nil {
		logrus.Fatalf("accessing property in non-object element: %v", err)
	}
	if rv.Kind() == reflect.Map {
		return to.keysSliceMap(rv)
	}
	return to.keysSliceStruct(rv)
}

func (to KwObject) keysSliceMap(rv reflect.Value) []string {
	var res []string
	keys := rv.MapKeys()
	for _, k := range keys {
		res = append(res, wutils.LowerFirstLetter(returnValue(k).String()))
	}
	return res
}

func (to KwObject) keysSliceStruct(rv reflect.Value) []string {
	var res []string
	rNumFields := rv.NumField()
	for i := 0; i < rNumFields; i++ {
		res = append(res, wutils.LowerFirstLetter(rv.Type().Field(i).Name))
	}
	return res
}

// Join joins an object to the corrent one.
func (to *KwObject) Join(o Object) {
	kwO, ok := o.(*KwObject)
	if !ok {
		logrus.Fatalf("objects must always be joint with another object")
	}
	rv, err := to.assertObject()
	if err != nil {
		logrus.Fatalf("joining object in non-object element: %v", err)
	}
	rvO, err := kwO.assertObject()
	if err != nil {
		logrus.Fatalf("joining non-object element: %v", err)
	}
	if rv.Kind() != reflect.Map {
		rv = reflect.ValueOf(structs.Map(to.Val))
	}
	if rvO.Kind() != reflect.Map {
		rvO = reflect.ValueOf(structs.Map(kwO.Val))
	}
	indirectMap := reflect.Indirect(rv)
	aMap := reflect.MakeMap(indirectMap.Type())
	for _, key := range rv.MapKeys() {
		aMap.SetMapIndex(key, rv.MapIndex(key))
	}
	for _, key := range rvO.MapKeys() {
		aMap.SetMapIndex(key, rvO.MapIndex(key))
	}
	to.Val = aMap.Interface()
}

// Prop returns the property value of the current object as a keyword object itself.
func (to *KwObject) Prop(k string) Object {
	rv, err := to.assertObject()
	if err != nil {
		logrus.Fatalf("accessing property in non-object element: %v", err)
	}
	if rv.Kind() == reflect.Map {
		return to.propMap(rv, k)
	}
	return to.propStruct(rv, k)
}

func (to *KwObject) propMap(rv reflect.Value, k string) Object {
	rKeys := rv.MapKeys()
	for _, key := range rKeys {
		if key.Interface() == k {
			return returnValue(rv.MapIndex(key))
		}
	}
	return NewKwNil()
}

func (to *KwObject) propStruct(rv reflect.Value, k string) Object {
	for i := rv.NumField() - 1; i > -1; i-- {
		if rv.Type().Field(i).Name == k {
			return returnValue(rv.Field(i))
		}
	}
	return NewKwNil()
}

// RemoveProp returns a new object without the provided property.
func (to *KwObject) RemoveProp(k string) Object {
	rv, err := to.assertObject()
	if err != nil {
		logrus.Fatalf("removing property in non-object element: %v", err)
	}
	if rv.Kind() == reflect.Map {
		return to.removePropMap(rv, k)
	}
	return to.removePropStruct(rv, k)
}

func (to *KwObject) removePropMap(rv reflect.Value, k string) Object {
	totalFields := rv.Len()
	if totalFields == 0 {
		return NewKwObject(to.Val)
	}
	var mapValue reflect.Value = reflect.ValueOf(to.Val)
	indirectMap := reflect.Indirect(mapValue)
	aMap := reflect.MakeMap(indirectMap.Type())
	keys := rv.MapKeys()
	for _, key := range keys {
		if key.Interface() != k {
			aMap.SetMapIndex(key, rv.MapIndex(key))
		}
	}
	return NewKwObject(aMap.Interface())
}

func (to *KwObject) removePropStruct(rv reflect.Value, k string) Object {
	totalFields := rv.NumField()
	if totalFields == 0 {
		return NewKwObject(to.Val)
	}
	var mapValue reflect.Value = reflect.ValueOf(structs.Map(to.Val))
	indirectMap := reflect.Indirect(mapValue)
	aMap := reflect.MakeMap(indirectMap.Type())
	for i := totalFields - 1; i > -1; i-- {
		key := rv.Type().Field(i).Name
		if key != k {
			aMap.SetMapIndex(reflect.ValueOf(key), rv.Field(i))
		}
	}
	return NewKwObject(aMap.Interface())
}

// ReplacePropValue returns a new object with the property changed.
func (to *KwObject) ReplacePropValue(k, v string) Object {
	rv, err := to.assertObject()
	if err != nil {
		logrus.Fatalf("replacing property value in non-object element: %v", err)
	}
	if rv.Kind() == reflect.Map {
		return to.replacePropValueMap(rv, k, v)
	}
	return to.replacePropValueStruct(rv, k, v)
}

func (to *KwObject) replacePropValueMap(rv reflect.Value, k, v string) Object {
	totalFields := rv.Len()
	if totalFields == 0 {
		return NewKwObject(to.Val)
	}
	var mapValue reflect.Value = reflect.ValueOf(to.Val)
	indirectMap := reflect.Indirect(mapValue)
	aMap := reflect.MakeMap(indirectMap.Type())
	keys := rv.MapKeys()
	for _, key := range keys {
		if key.Interface() == k {
			aMap.SetMapIndex(key, reflect.ValueOf(v))
		} else {
			aMap.SetMapIndex(key, rv.MapIndex(key))
		}
	}
	return NewKwObject(aMap.Interface())
}

func (to *KwObject) replacePropValueStruct(rv reflect.Value, k, v string) Object {
	totalFields := rv.NumField()
	if totalFields == 0 {
		return NewKwObject(to.Val)
	}
	var mapValue reflect.Value = reflect.ValueOf(structs.Map(to.Val))
	indirectMap := reflect.Indirect(mapValue)
	aMap := reflect.MakeMap(indirectMap.Type())
	for i := totalFields - 1; i > -1; i-- {
		key := rv.Type().Field(i).Name
		if key == k {
			aMap.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(v))
		} else {
			aMap.SetMapIndex(reflect.ValueOf(key), rv.Field(i))

		}
	}
	return NewKwObject(aMap.Interface())
}

// Index returns the value underlying the passed index as a keyword object itself.
func (to *KwObject) Index(int) Object {
	logrus.Fatalf("accessing index in non-array element: expected array but got object")
	return NewKwNil()
}

// Len returns the length that the object has.
func (to *KwObject) Len() int {
	rv, err := to.assertObject()
	if err != nil {
		logrus.Fatalf("error getting the object length: %v", err)
	}
	if rv.Kind() == reflect.Map {
		return to.Len()
	}
	return rv.NumField()
}

// assertObject ensures that the value type is a struct or a map
func (to *KwObject) assertObject() (reflect.Value, error) {
	rv := reflect.Indirect(reflect.ValueOf(to.Val))
	if rkind := rv.Kind(); rkind != reflect.Struct && rkind != reflect.Map {
		return rv, fmt.Errorf("expected struct/map but got %s", rkind)
	}
	return rv, nil
}

// KwArray is the array type for a keyword object.
type KwArray struct {
	Val interface{} `json:"-"`
}

// NewKwArray is a constructor for KwArray.
func NewKwArray(v interface{}) *KwArray {
	return &KwArray{v}
}

// Value returns the value of this object when printed.
func (ta *KwArray) Value() string {
	return fmt.Sprintf("%v", ta.Val)
}

// String returns the value formatted as string.
func (ta KwArray) String() string {
	return fmt.Sprintf("[%s]", strings.Join(ta.StringSlice(), ","))
}

// String returns the value formatted as a slice of objects.
func (ta KwArray) Slice() []Object {
	rv, err := ta.assertArray()
	if err != nil {
		logrus.Fatalf("retrieving string slice from non-array element: %v", err)
	}
	var res []Object
	rLen := rv.Len()
	for i := 0; i < rLen; i++ {
		res = append(res, returnValue(rv.Index(i)))
	}
	return res
}

// StringSlice returns the value formatted as a slice of strings.
func (ta KwArray) StringSlice() []string {
	rv, err := ta.assertArray()
	if err != nil {
		logrus.Fatalf("retrieving string slice from non-array element: %v", err)
	}
	var res []string
	rLen := rv.Len()
	for i := 0; i < rLen; i++ {
		res = append(res, returnValue(rv.Index(i)).String())
	}
	return res
}

// RemoveValue returns a new object without the provided property.
func (ta *KwArray) RemoveValue(k string) Object {
	rv, err := ta.assertArray()
	if err != nil {
		logrus.Fatalf("removing value from non-array element: %v", err)
	}
	rLen := rv.Len()
	if rLen == 0 {
		return NewKwArray(ta.Val)
	}
	var res []interface{}
	for i := 0; i < rLen; i++ {
		value := rv.Index(i).Interface()
		if value != k {
			res = append(res, value)
		}
	}
	return NewKwArray(res)
}

// ReplaceValue returns a new object with the value replaced.
func (ta *KwArray) ReplaceValue(k, v string) Object {
	rv, err := ta.assertArray()
	if err != nil {
		logrus.Fatalf("replacing value from non-array element: %v", err)
	}
	rLen := rv.Len()
	if rLen == 0 {
		return NewKwArray(ta.Val)
	}
	var res []interface{}
	for i := 0; i < rLen; i++ {
		value := rv.Index(i).Interface()
		if value == k {
			res = append(res, v)
		} else {
			res = append(res, value)
		}
	}
	return NewKwArray(res)
}

// Join joins an object to the corrent one.
func (ta *KwArray) Join(a Object) {
	kwA, ok := a.(*KwArray)
	if !ok {
		logrus.Fatalf("array must always be joint with another array")
	}
	rv, err := ta.assertArray()
	if err != nil {
		logrus.Fatalf("joining array in non-array element: %v", err)
	}
	rvA, err := kwA.assertArray()
	if err != nil {
		logrus.Fatalf("joining non-array element: %v", err)
	}
	var res []interface{}
	rLen := rv.Len()
	for i := 0; i < rLen; i++ {
		res = append(res, rv.Index(i).Interface())
	}
	rALen := rvA.Len()
	for i := 0; i < rALen; i++ {
		res = append(res, rvA.Index(i).Interface())
	}
	ta.Val = res
}

// Prop returns the property value of the current object as a keyword object itself.
func (ta *KwArray) Prop(string) Object {
	logrus.Fatalf("accessing property in non-object element: expected object but got array")
	return NewKwNil()
}

// Index returns the value underlying the passed index as a keyword object itself.
func (ta *KwArray) Index(i int) Object {
	rv, err := ta.assertArray()
	if err != nil {
		logrus.Fatalf("accessing index in non-array element: %v", err)
	}
	if rlen := rv.Len(); i >= rlen {
		return NewKwNil()
	}
	return returnValue(rv.Index(i))
}

// Len returns the length that the object has.
func (ta *KwArray) Len() int {
	rv, err := ta.assertArray()
	if err != nil {
		logrus.Fatalf("error getting the array length: %v", err)
	}
	return rv.Len()
}

// assertObject ensures that the value type is a slice or an array
func (ta *KwArray) assertArray() (reflect.Value, error) {
	rv := reflect.Indirect(reflect.ValueOf(ta.Val))
	if rkind := rv.Kind(); rkind != reflect.Slice && rkind != reflect.Array {
		return rv, fmt.Errorf("expected array but got %s", rkind)
	}
	return rv, nil
}

// KwPrimitive is the primitive type for a keyword object.
type KwPrimitive struct {
	Val interface{} `json:"-"`
}

// NewKwPrimitive is a constructor for KwPrimitive.
func NewKwPrimitive(v interface{}) *KwPrimitive {
	return &KwPrimitive{v}
}

// Value returns the value of this object when printed.
func (tp *KwPrimitive) Value() string {
	return fmt.Sprintf("%v", tp.Val)
}

// String returns the value formatted as string.
func (tp KwPrimitive) String() string {
	return fmt.Sprintf("%v", tp.Val)
}

// Slice returns the value formatted as a slice of objects.
func (tp KwPrimitive) Slice() []Object {
	return []Object{NewKwPrimitive(fmt.Sprintf("%v", tp.Val))}
}

// StringSlice returns the value formatted as a slice of strings.
func (tp KwPrimitive) StringSlice() []string {
	return []string{fmt.Sprintf("%v", tp.Val)}
}

// Join joins an object to the corrent one.
func (tp KwPrimitive) Join(a Object) {
	logrus.Fatalf("primitives cannot be joint")
}

// Prop returns the property value of the current object as a keyword object itself.
func (tp *KwPrimitive) Prop(string) Object {
	logrus.Fatalf("accessing property in non-object element: expected object but got primitive")
	return NewKwNil()
}

// Index returns the value underlying the passed index as a keyword object itself.
func (tp *KwPrimitive) Index(i int) Object {
	rv := reflect.Indirect(reflect.ValueOf(tp.Val))
	if rkind := rv.Kind(); rkind != reflect.String {
		logrus.Fatalf("invalid index access in primitive element: expected string but got %s", rkind)
	}
	val := rv.String()
	if valLen := len(val); i >= valLen {
		logrus.Fatalf("accessing index in string element: index out of range (length=%d, index=%d)", valLen, i)
	}
	return NewKwPrimitive(string(val[i]))
}

// Len returns the length that the object has.
func (tp *KwPrimitive) Len() int {
	rv := reflect.Indirect(reflect.ValueOf(tp.Val))
	if rkind := rv.Kind(); rkind != reflect.String {
		logrus.Fatalf("unable to get the length of an element with type %s", rkind)
	}
	return len(rv.Interface().(string))
}

// KwPrimitive is the nil type for a keyword object.
type KwNil struct {
}

// NewKwNil is a constructor for KwNil.
func NewKwNil() *KwNil {
	return &KwNil{}
}

// Value returns the value of this object when printed.
func (tn *KwNil) Value() string {
	return ""
}

// String returns the value formatted as string.
func (tn KwNil) String() string {
	return ""
}

// Slice returns the value formatted as a slice of objects.
func (tn KwNil) Slice() []Object {
	return []Object{}
}

// StringSlice returns the value formatted as a slice of strings.
func (tn KwNil) StringSlice() []string {
	return []string{}
}

// Join joins an object to the corrent one.
func (tn KwNil) Join(a Object) {
	logrus.Fatalf("nil values cannot be joint")
}

// Prop returns the property value of the current object as a keyword object itself.
func (tn *KwNil) Prop(string) Object {
	logrus.Fatalf("accessing property in nil element")
	return tn
}

// Index returns the value underlying the passed index as a keyword object itself.
func (tn *KwNil) Index(int) Object {
	logrus.Fatalf("accessing index in nil element")
	return tn
}

// Len returns the length that the object has.
func (tn *KwNil) Len() int {
	return 0
}

// IsPrimitive returns if the keyword object type is primitive.
func IsPrimitive(o Object) bool {
	_, ok := o.(*KwPrimitive)
	return ok
}

// IsArray returns if the keyword object type is array.
func IsArray(o Object) bool {
	_, ok := o.(*KwArray)
	return ok
}

// IsObject returns if the keyword object type is object.
func IsObject(o Object) bool {
	_, ok := o.(*KwObject)
	return ok
}

// IsNil returns if the keyword object type is nil.
func IsNil(o Object) bool {
	_, ok := o.(*KwNil)
	return ok
}

// returnValue returns the keyword object value for some abstract value.
// it validates and cleans the abstract value.
func returnValue(rfield reflect.Value) Object {
	if !rfield.IsValid() {
		return NewKwNil()
	}
	for rfield.Kind() == reflect.Ptr || rfield.Kind() == reflect.Interface {
		if rfield.IsNil() {
			return NewKwNil()
		}
		rfield = rfield.Elem()
	}
	val := rfield.Interface()
	if res, ok := val.(KwObject); ok {
		return &res
	}
	if res, ok := val.(KwArray); ok {
		return &res
	}
	if res, ok := val.(KwPrimitive); ok {
		return &res
	}
	kind := rfield.Kind()
	switch kind {
	case reflect.Struct, reflect.Map:
		return NewKwObject(val)
	case reflect.Slice, reflect.Array:
		return NewKwArray(val)
	default:
		return NewKwPrimitive(val)
	}
}
