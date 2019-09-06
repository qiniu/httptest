package httptest

import (
	"reflect"

	"github.com/qiniu/x/jsonutil"
)

// ---------------------------------------------------------------------------

func castFloat(v interface{}) (float64, bool) {

	t := reflect.ValueOf(v)

	kind := t.Kind()
	if kind < reflect.Int || kind > reflect.Float64 {
		return 0, false
	}

	if kind <= reflect.Int64 {
		return float64(t.Int()), true
	}
	if kind <= reflect.Uintptr {
		return float64(t.Uint()), true
	}
	return t.Float(), true
}

func Equal(v1, v2 interface{}) bool {

	f1, ok1 := castFloat(v1)
	f2, ok2 := castFloat(v2)
	if ok1 != ok2 {
		return false
	}
	if ok1 && f1 == f2 {
		return true
	}
	return reflect.DeepEqual(v1, v2)
}

func EqualSet(obj1, obj2 interface{}) bool {

	var v interface{}
	if text, ok := obj1.(string); ok {
		err := jsonutil.Unmarshal(text, &v)
		if err != nil {
			return false
		}
		obj1 = v
	}

	v1 := reflect.ValueOf(obj1)
	if v1.Kind() != reflect.Slice {
		return false
	}

	if text, ok := obj2.(string); ok {
		err := jsonutil.Unmarshal(text, &v)
		if err != nil {
			return false
		}
		obj2 = v
	}

	v2 := reflect.ValueOf(obj2)
	if v2.Kind() != reflect.Slice {
		return false
	}

	if v1.Len() != v2.Len() {
		return false
	}
	for i := 0; i < v1.Len(); i++ {
		item1 := v1.Index(i)
		if !hasElem(item1.Interface(), v2) {
			return false
		}
	}
	return true
}

func hasElem(item1 interface{}, v2 reflect.Value) bool {

	for j := 0; j < v2.Len(); j++ {
		item2 := v2.Index(j)
		if Equal(item1, item2.Interface()) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------

type Var struct {
	Data interface{}
	Ok   bool
}

func (p Var) Equal(v interface{}) bool {

	return p.Ok && Equal(p.Data, v)
}

func (p Var) EqualObject(obj string) bool {

	if !p.Ok {
		return false
	}

	var v interface{}
	err := jsonutil.Unmarshal(obj, &v)
	if err != nil {
		return false
	}

	return Equal(p.Data, v)
}

func (p Var) EqualSet(obj interface{}) bool {

	if !p.Ok {
		return false
	}
	return EqualSet(p.Data, obj)
}

// ---------------------------------------------------------------------------
