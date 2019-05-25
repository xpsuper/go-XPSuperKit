package XPSuperKit

import (
	"fmt"
	"reflect"
	"strings"
	"errors"
)

var (
	XPStructDefaultTagName = "structs"
)

type XPStructImpl struct {
	raw     interface{}
	value   reflect.Value
	TagName string
}

type XPStructField struct {
	value      reflect.Value
	field      reflect.StructField
	defaultTag string
}

func NewXPStruct(s interface{}) *XPStructImpl {
	return &XPStructImpl{
		raw:     s,
		value:   strctVal(s),
		TagName: XPStructDefaultTagName,
	}
}

func IsStruct(s interface{}) bool {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// uninitialized zero value of a struct
	if v.Kind() == reflect.Invalid {
		return false
	}

	return v.Kind() == reflect.Struct
}

func (s *XPStructImpl) Map() map[string]interface{} {
	out := make(map[string]interface{})
	s.FillMap(out)
	return out
}

func (s *XPStructImpl) FillMap(out map[string]interface{}) {
	if out == nil {
		return
	}

	fields := s.structFields()

	for _, field := range fields {
		name := field.Name
		val := s.value.FieldByName(name)
		isSubStruct := false
		var finalVal interface{}

		tagName, tagOpts := parseTag(field.Tag.Get(s.TagName))
		if tagName != "" {
			name = tagName
		}

		// if the value is a zero value and the field is marked as omitempty do
		// not include
		if tagOpts.Has("omitempty") {
			zero := reflect.Zero(val.Type()).Interface()
			current := val.Interface()

			if reflect.DeepEqual(current, zero) {
				continue
			}
		}

		if !tagOpts.Has("omitnested") {
			finalVal = s.nested(val)

			v := reflect.ValueOf(val.Interface())
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}

			switch v.Kind() {
			case reflect.Map, reflect.Struct:
				isSubStruct = true
			}
		} else {
			finalVal = val.Interface()
		}

		if tagOpts.Has("string") {
			s, ok := val.Interface().(fmt.Stringer)
			if ok {
				out[name] = s.String()
			}
			continue
		}

		if isSubStruct && (tagOpts.Has("flatten")) {
			for k := range finalVal.(map[string]interface{}) {
				out[k] = finalVal.(map[string]interface{})[k]
			}
		} else {
			out[name] = finalVal
		}
	}
}

func (s *XPStructImpl) Values() []interface{} {
	fields := s.structFields()

	var t []interface{}

	for _, field := range fields {
		val := s.value.FieldByName(field.Name)

		_, tagOpts := parseTag(field.Tag.Get(s.TagName))

		// if the value is a zero value and the field is marked as omitempty do
		// not include
		if tagOpts.Has("omitempty") {
			zero := reflect.Zero(val.Type()).Interface()
			current := val.Interface()

			if reflect.DeepEqual(current, zero) {
				continue
			}
		}

		if tagOpts.Has("string") {
			s, ok := val.Interface().(fmt.Stringer)
			if ok {
				t = append(t, s.String())
			}
			continue
		}

		if IsStruct(val.Interface()) && !tagOpts.Has("omitnested") {
			// look out for embedded structs, and convert them to a
			// []interface{} to be added to the final values slice
			for _, embeddedVal := range NewXPStruct(val.Interface()).Values() {
				t = append(t, embeddedVal)
			}
		} else {
			t = append(t, val.Interface())
		}
	}

	return t
}

func (s *XPStructImpl) Fields() []*XPStructField {
	return getFields(s.value, s.TagName)
}

func (s *XPStructImpl) Name() string {
	return s.value.Type().Name()
}

func (s *XPStructImpl) Names() []string {
	fields := getFields(s.value, s.TagName)

	names := make([]string, len(fields))

	for i, field := range fields {
		names[i] = field.Name()
	}

	return names
}

func (s *XPStructImpl) Field(name string) *XPStructField {
	f, ok := s.FieldOk(name)
	if !ok {
		panic("field not found")
	}

	return f
}

func (s *XPStructImpl) FieldOk(name string) (*XPStructField, bool) {
	t := s.value.Type()

	field, ok := t.FieldByName(name)
	if !ok {
		return nil, false
	}

	return &XPStructField{
		field:      field,
		value:      s.value.FieldByName(name),
		defaultTag: s.TagName,
	}, true
}

func (s *XPStructImpl) IsZero() bool {
	fields := s.structFields()

	for _, field := range fields {
		val := s.value.FieldByName(field.Name)

		_, tagOpts := parseTag(field.Tag.Get(s.TagName))

		if IsStruct(val.Interface()) && !tagOpts.Has("omitnested") {
			ok :=  NewXPStruct(val.Interface()).IsZero()
			if !ok {
				return false
			}

			continue
		}

		// zero value of the given field, such as "" for string, 0 for int
		zero := reflect.Zero(val.Type()).Interface()

		//  current value of the given field
		current := val.Interface()

		if !reflect.DeepEqual(current, zero) {
			return false
		}
	}

	return true
}

func (s *XPStructImpl) HasZero() bool {
	fields := s.structFields()

	for _, field := range fields {
		val := s.value.FieldByName(field.Name)

		_, tagOpts := parseTag(field.Tag.Get(s.TagName))

		if IsStruct(val.Interface()) && !tagOpts.Has("omitnested") {
			ok := NewXPStruct(val.Interface()).HasZero()
			if ok {
				return true
			}

			continue
		}

		// zero value of the given field, such as "" for string, 0 for int
		zero := reflect.Zero(val.Type()).Interface()

		//  current value of the given field
		current := val.Interface()

		if reflect.DeepEqual(current, zero) {
			return true
		}
	}

	return false
}

func (s *XPStructImpl) structFields() []reflect.StructField {
	t := s.value.Type()

	var f []reflect.StructField

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// we can't access the value of unexported fields
		if field.PkgPath != "" {
			continue
		}

		// don't check if it's omitted
		if tag := field.Tag.Get(s.TagName); tag == "-" {
			continue
		}

		f = append(f, field)
	}

	return f
}

func (s *XPStructImpl) nested(val reflect.Value) interface{} {
	var finalVal interface{}

	v := reflect.ValueOf(val.Interface())
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		n := NewXPStruct(val.Interface())
		n.TagName = s.TagName
		m := n.Map()

		// do not add the converted value if there are no exported fields, ie:
		// time.Time
		if len(m) == 0 {
			finalVal = val.Interface()
		} else {
			finalVal = m
		}
	case reflect.Map:
		// get the element type of the map
		mapElem := val.Type()
		switch val.Type().Kind() {
		case reflect.Ptr, reflect.Array, reflect.Map,
			reflect.Slice, reflect.Chan:
			mapElem = val.Type().Elem()
			if mapElem.Kind() == reflect.Ptr {
				mapElem = mapElem.Elem()
			}
		}

		// only iterate over struct types, ie: map[string]StructType,
		// map[string][]StructType,
		if mapElem.Kind() == reflect.Struct ||
			(mapElem.Kind() == reflect.Slice &&
				mapElem.Elem().Kind() == reflect.Struct) {
			m := make(map[string]interface{}, val.Len())
			for _, k := range val.MapKeys() {
				m[k.String()] = s.nested(val.MapIndex(k))
			}
			finalVal = m
			break
		}

		// TODO(arslan): should this be optional?
		finalVal = val.Interface()
	case reflect.Slice, reflect.Array:
		if val.Type().Kind() == reflect.Interface {
			finalVal = val.Interface()
			break
		}

		// TODO(arslan): should this be optional?
		// do not iterate of non struct types, just pass the value. Ie: []int,
		// []string, co... We only iterate further if it's a struct.
		// i.e []foo or []*foo
		if val.Type().Elem().Kind() != reflect.Struct &&
			!(val.Type().Elem().Kind() == reflect.Ptr &&
				val.Type().Elem().Elem().Kind() == reflect.Struct) {
			finalVal = val.Interface()
			break
		}

		slices := make([]interface{}, val.Len(), val.Len())
		for x := 0; x < val.Len(); x++ {
			slices[x] = s.nested(val.Index(x))
		}
		finalVal = slices
	default:
		finalVal = val.Interface()
	}

	return finalVal
}

func getFields(v reflect.Value, tagName string) []*XPStructField {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()

	var fields []*XPStructField

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if tag := field.Tag.Get(tagName); tag == "-" {
			continue
		}

		f := &XPStructField{
			field: field,
			value: v.FieldByName(field.Name),
		}

		fields = append(fields, f)

	}

	return fields
}

func strctVal(s interface{}) reflect.Value {
	v := reflect.ValueOf(s)

	// if pointer get the underlying element≤
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		panic("not struct")
	}

	return v
}

var (
	errNotExported = errors.New("field is not exported")
	errNotSettable = errors.New("field is not settable")
)

func (f *XPStructField) Tag(key string) string {
	return f.field.Tag.Get(key)
}

func (f *XPStructField) Value() interface{} {
	return f.value.Interface()
}

func (f *XPStructField) IsEmbedded() bool {
	return f.field.Anonymous
}

func (f *XPStructField) IsExported() bool {
	return f.field.PkgPath == ""
}

func (f *XPStructField) IsZero() bool {
	zero := reflect.Zero(f.value.Type()).Interface()
	current := f.Value()

	return reflect.DeepEqual(current, zero)
}

func (f *XPStructField) Name() string {
	return f.field.Name
}

func (f *XPStructField) Kind() reflect.Kind {
	return f.value.Kind()
}

func (f *XPStructField) Set(val interface{}) error {
	// we can't set unexported fields, so be sure this field is exported
	if !f.IsExported() {
		return errNotExported
	}

	// do we get here? not sure...
	if !f.value.CanSet() {
		return errNotSettable
	}

	given := reflect.ValueOf(val)

	if f.value.Kind() != given.Kind() {
		return fmt.Errorf("wrong kind. got: %s want: %s", given.Kind(), f.value.Kind())
	}

	f.value.Set(given)
	return nil
}

func (f *XPStructField) Zero() error {
	zero := reflect.Zero(f.value.Type()).Interface()
	return f.Set(zero)
}

func (f *XPStructField) Fields() []*XPStructField {
	return getFields(f.value, f.defaultTag)
}


func (f *XPStructField) Field(name string) *XPStructField {
	field, ok := f.FieldOk(name)
	if !ok {
		panic("field not found")
	}

	return field
}

func (f *XPStructField) FieldOk(name string) (*XPStructField, bool) {
	value := &f.value
	// value must be settable so we need to make sure it holds the address of the
	// variable and not a copy, so we can pass the pointer to strctVal instead of a
	// copy (which is not assigned to any variable, hence not settable).
	// see "https://blog.golang.org/laws-of-reflection#TOC_8."
	if f.value.Kind() != reflect.Ptr {
		a := f.value.Addr()
		value = &a
	}
	v := strctVal(value.Interface())
	t := v.Type()

	field, ok := t.FieldByName(name)
	if !ok {
		return nil, false
	}

	return &XPStructField{
		field: field,
		value: v.FieldByName(name),
	}, true
}

type tagOptions []string

func (t tagOptions) Has(opt string) bool {
	for _, tagOpt := range t {
		if tagOpt == opt {
			return true
		}
	}

	return false
}

func parseTag(tag string) (string, tagOptions) {
	// tag is one of followings:
	// ""
	// "name"
	// "name,opt"
	// "name,opt,opt2"
	// ",opt"

	res := strings.Split(tag, ",")
	return res[0], res[1:]
}

func (f *XPStructField) Copy(src, dst interface{}) error {
	srcV, err := srcFilter(src)
	if err != nil {
		return err
	}
	dstV, err := dstFilter(dst)
	if err != nil {
		return err
	}
	srcKeys := make(map[string]bool)
	for i := 0; i < srcV.NumField(); i++ {
		srcKeys[srcV.Type().Field(i).Name] = true
	}
	for i := 0; i < dstV.Elem().NumField(); i++ {
		fName := dstV.Elem().Type().Field(i).Name
		if _, ok := srcKeys[fName]; ok {
			v := srcV.FieldByName(dstV.Elem().Type().Field(i).Name)
			if v.CanInterface() {
				dstV.Elem().Field(i).Set(v)
			}
		}
	}

	return nil
}

func srcFilter(src interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(src)
	if v.Type().Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return reflect.Zero(v.Type()), errors.New("src type error: not a struct or a pointer to struct")
	}
	return v, nil
}

func dstFilter(src interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(src)
	if v.Type().Kind() != reflect.Ptr {
		return reflect.Zero(v.Type()), errors.New("src type error: not a pointer to struct")
	}
	if v.Elem().Kind() != reflect.Struct {
		return reflect.Zero(v.Type()), errors.New("src type error: not point to struct")
	}
	return v, nil
}