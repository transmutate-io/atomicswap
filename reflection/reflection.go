package reflection

import (
	"errors"
	"reflect"
)

var (
	ErrNilPointer  = errors.New("nil pointer")
	ErrNotAPointer = errors.New("not a pointer")
	ErrNotAStruct  = errors.New("not a struct")
)

// copy fields by name
func CopyFieldsByName(src, dst interface{}) error {
	if src == nil || dst == nil {
		return ErrNilPointer
	}
	vd := reflect.ValueOf(dst)
	if vd.Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	if vd.IsNil() {
		return ErrNilPointer
	}
	if vd = vd.Elem(); vd.Kind() != reflect.Struct {
		return ErrNotAStruct
	}
	vs := reflect.ValueOf(src)
	if vs.Kind() == reflect.Ptr {
		if vs.IsNil() {
			return ErrNilPointer
		}
		vs = vs.Elem()
	}
	if vs.Kind() != reflect.Struct {
		return ErrNotAStruct
	}
	if vs.Kind() != reflect.Struct || vd.Kind() != reflect.Struct {
		return ErrNotAStruct
	}
	vst := vs.Type()
	vdt := vd.Type()
	for i := 0; i < vst.NumField(); i++ {
		sfld := vst.Field(i)
		dfld, ok := vdt.FieldByName(sfld.Name)
		if !ok {
			continue
		}
		if sfld.Type != dfld.Type {
			if dfld.Type.Kind() != reflect.Interface {
				continue
			}
			if !sfld.Type.Implements(dfld.Type) {
				continue
			}
		}
		vd.FieldByName(sfld.Name).Set(vs.Field(i))
	}
	return nil
}

type FieldReplacementMap = map[string]interface{}

func ReplaceTypeFields(t interface{}, rm FieldReplacementMap) (interface{}, error) {
	if t == nil {
		return nil, ErrNilPointer
	}
	tt := reflect.TypeOf(t)
	notPointer := true
	if tt.Kind() == reflect.Ptr {
		tt = tt.Elem()
		notPointer = false
	}
	if tt.Kind() != reflect.Struct {
		return nil, ErrNotAStruct
	}
	newFields := make([]reflect.StructField, 0, tt.NumField())
	for i := 0; i < tt.NumField(); i++ {
		fld := tt.Field(i)
		if newType, ok := rm[fld.Name]; ok {
			fld.Type = reflect.TypeOf(newType)
		}
		newFields = append(newFields, fld)
	}
	tr := reflect.New(reflect.StructOf(newFields))
	if notPointer {
		tr = tr.Elem()
	}
	return tr.Interface(), nil
}

var ErrFieldNotFound = errors.New("field not found")

func FilterFields(t interface{}, flds ...string) (interface{}, error) {
	if t == nil {
		return nil, ErrNilPointer
	}
	tv := reflect.ValueOf(t)
	notPointer := true
	if tv.Kind() == reflect.Ptr {
		if tv.IsNil() {
			return nil, ErrNilPointer
		}
		tv = tv.Elem()
		notPointer = false
	}
	if tv.Kind() != reflect.Struct {
		return nil, ErrNotAStruct
	}
	tt := tv.Type()
	newFields := make([]reflect.StructField, 0, len(flds))
	for _, i := range flds {
		fld, ok := tt.FieldByName(i)
		if !ok {
			return nil, ErrFieldNotFound
		}
		newFields = append(newFields, fld)
	}
	tr := reflect.New(reflect.StructOf(newFields))
	if notPointer {
		tr = tr.Elem()
	}
	return tr.Interface(), nil
}

func HasField(t interface{}, name string) bool {
	if t == nil {
		return false
	}
	tv := reflect.ValueOf(t)
	if tv.Kind() == reflect.Ptr {
		if tv.IsNil() {
			return false
		}
		tv = tv.Elem()
	}
	if tv.Kind() != reflect.Struct {
		return false
	}
	_, ok := tv.Type().FieldByName(name)
	return ok
}

func isType(t reflect.Type, _type interface{}) bool {
	return t.String() == reflect.TypeOf(_type).String()
}

func IsType(t interface{}, _type interface{}) bool {
	return isType(reflect.TypeOf(t), _type)
}

func FieldIsType(t interface{}, name string, _type interface{}) bool {
	if t == nil {
		return false
	}
	tv := reflect.ValueOf(t)
	if tv.Kind() == reflect.Ptr {
		if tv.IsNil() {
			return false
		}
		tv = tv.Elem()
	}
	if tv.Kind() != reflect.Struct {
		return false
	}
	fld, ok := tv.Type().FieldByName(name)
	if !ok {
		return false
	}
	return isType(fld.Type, _type)
}
