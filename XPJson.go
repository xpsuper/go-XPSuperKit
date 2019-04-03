package XPSuperKit

import (
	"bytes"
	"io"
	"log"
	"reflect"
	"strconv"
	"encoding/json"
)

type XPJsonImpl struct {
	data interface{}
}

func (j *XPJsonImpl) marshalJSON() ([]byte, error) {
	return json.Marshal(&j.data)
}

func (j *XPJsonImpl) unmarshalJSON(p []byte) error {
	dec := json.NewDecoder(bytes.NewBuffer(p))
	dec.UseNumber()
	return dec.Decode(&j.data)
}

func (j *XPJsonImpl) float64Value() (float64, error) {
	switch j.data.(type) {
	case json.Number:
		return j.data.(json.Number).Float64()
	case float32, float64:
		return reflect.ValueOf(j.data).Float(), nil
	case int, int8, int16, int32, int64:
		return float64(reflect.ValueOf(j.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return float64(reflect.ValueOf(j.data).Uint()), nil
	}

	return 0, ErrorN("invalid value type")
}

func (j *XPJsonImpl) intValue() (int, error) {
	switch j.data.(type) {
	case json.Number:
		i, err := j.data.(json.Number).Int64()
		return int(i), err
	case float32, float64:
		return int(reflect.ValueOf(j.data).Float()), nil
	case int, int8, int16, int32, int64:
		return int(reflect.ValueOf(j.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return int(reflect.ValueOf(j.data).Uint()), nil
	}

	return 0, ErrorN("invalid value type")
}

func (j *XPJsonImpl) int64Value() (int64, error) {
	switch j.data.(type) {
	case json.Number:
		return j.data.(json.Number).Int64()
	case float32, float64:
		return int64(reflect.ValueOf(j.data).Float()), nil
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(j.data).Int(), nil
	case uint, uint8, uint16, uint32, uint64:
		return int64(reflect.ValueOf(j.data).Uint()), nil
	}
	return 0, ErrorN("invalid value type")
}

func (j *XPJsonImpl) uint64Value() (uint64, error) {
	switch j.data.(type) {
	case json.Number:
		return strconv.ParseUint(j.data.(json.Number).String(), 10, 64)
	case float32, float64:
		return uint64(reflect.ValueOf(j.data).Float()), nil
	case int, int8, int16, int32, int64:
		return uint64(reflect.ValueOf(j.data).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(j.data).Uint(), nil
	}

	return 0, ErrorN("invalid value type")
}

func NewJson() *XPJsonImpl {
	return &XPJsonImpl{
		data: make(map[string]interface{}),
	}
}

func NewJsonFromBytes(body []byte) (*XPJsonImpl, error) {
	j := new(XPJsonImpl)
	err := j.unmarshalJSON(body)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func NewJsonFromString(str string) (*XPJsonImpl, error) {
	b := []byte(str)
	j := new(XPJsonImpl)
	err := j.unmarshalJSON(b)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func NewJsonFromInterface(src interface{}) *XPJsonImpl {
	j := new(XPJsonImpl)
	j.data = src
	return j
}

func NewJsonFromReader(reader io.Reader) (*XPJsonImpl, error) {
	j := new(XPJsonImpl)
	dec := json.NewDecoder(reader)
	dec.UseNumber()
	err := dec.Decode(&j.data)

	return j, err
}

func (j *XPJsonImpl) Interface() interface{} {
	return j.data
}

func (j *XPJsonImpl) Encode() ([]byte, error) {
	return j.marshalJSON()
}

func (j *XPJsonImpl) EncodePretty() ([]byte, error) {
	return json.MarshalIndent(&j.data, "", "  ")
}

func (j *XPJsonImpl) ToString() (string, error) {
	b, err := j.marshalJSON()
	return string(b), err
}

func (j *XPJsonImpl) ToStringPretty() (string, error) {
	b, err := json.MarshalIndent(&j.data, "", "  ")
	return string(b), err
}

func (j *XPJsonImpl) Set(key string, value interface{}) {
	m, err := j.GetMap()
	if err != nil {
		return
	}
	m[key] = value
}

func (j *XPJsonImpl) SetPath(branch []string, value interface{}) {
	if len(branch) == 0 {
		j.data = value
		return
	}

	if _, ok := (j.data).(map[string]interface{}); !ok {
		j.data = make(map[string]interface{})
	}
	curr := j.data.(map[string]interface{})

	for i := 0; i < len(branch)-1; i++ {
		b := branch[i]

		if _, ok := curr[b]; !ok {
			n := make(map[string]interface{})
			curr[b] = n
			curr = n
			continue
		}

		if _, ok := curr[b].(map[string]interface{}); !ok {
			n := make(map[string]interface{})
			curr[b] = n
		}

		curr = curr[b].(map[string]interface{})
	}

	curr[branch[len(branch)-1]] = value
}

func (j *XPJsonImpl) TryGet(key string) (*XPJsonImpl, bool) {
	m, err := j.GetMap()
	if err == nil {
		if value, ok := m[key]; ok {
			return &XPJsonImpl{value}, true
		}
	}
	return nil, false
}

func (j *XPJsonImpl) Get(key string) *XPJsonImpl {
	m, err := j.GetMap()
	if err == nil {
		if value, ok := m[key]; ok {
			return &XPJsonImpl{value}
		}
	}
	return &XPJsonImpl{nil}
}

func (j *XPJsonImpl) GetPath(branch ...string) *XPJsonImpl {
	jin := j
	for _, p := range branch {
		jin = jin.Get(p)
	}
	return jin
}

func (j *XPJsonImpl) IndexOf(index int) *XPJsonImpl {
	a, err := j.GetArray()
	if err == nil {
		if len(a) > index {
			return &XPJsonImpl{a[index]}
		}
	}
	return &XPJsonImpl{nil}
}

func (j *XPJsonImpl) Delete(key string) {
	m, err := j.GetMap()
	if err != nil {
		return
	}
	delete(m, key)
}

func (j *XPJsonImpl) GetMap() (map[string]interface{}, error) {
	if m, ok := (j.data).(map[string]interface{}); ok {
		return m, nil
	}
	return nil, ErrorN("type assertion to map[string]interface{} failed")
}

func (j *XPJsonImpl) GetArray() ([]interface{}, error) {
	if a, ok := (j.data).([]interface{}); ok {
		return a, nil
	}
	return nil, ErrorN("type assertion to []interface{} failed")
}

func (j *XPJsonImpl) GetBool() (bool, error) {
	if s, ok := (j.data).(bool); ok {
		return s, nil
	}
	return false, ErrorN("type assertion to bool failed")
}

func (j *XPJsonImpl) GetString() (string, error) {
	if s, ok := (j.data).(string); ok {
		return s, nil
	}
	return "", ErrorN("type assertion to string failed")
}

func (j *XPJsonImpl) GetBytes() ([]byte, error) {
	if s, ok := (j.data).(string); ok {
		return []byte(s), nil
	}
	return nil, ErrorN("type assertion to []byte failed")
}

func (j *XPJsonImpl) GetStringArray() ([]string, error) {
	arr, err := j.GetArray()
	if err != nil {
		return nil, err
	}
	retArr := make([]string, 0, len(arr))
	for _, a := range arr {
		if a == nil {
			retArr = append(retArr, "")
			continue
		}
		s, ok := a.(string)
		if !ok {
			return nil, ErrorN("type assertion to []string failed")
		}
		retArr = append(retArr, s)
	}
	return retArr, nil
}

func (j *XPJsonImpl) ArrayValue(args ...[]interface{}) []interface{} {
	var def []interface{}

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("ArrayValue() received too many arguments %d", len(args))
	}

	a, err := j.GetArray()
	if err == nil {
		return a
	}

	return def
}

func (j *XPJsonImpl) MapValue(args ...map[string]interface{}) map[string]interface{} {
	var def map[string]interface{}

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MapValue() received too many arguments %d", len(args))
	}

	a, err := j.GetMap()
	if err == nil {
		return a
	}

	return def
}

func (j *XPJsonImpl) StringValue(args ...string) string {
	var def string

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("StringValue() received too many arguments %d", len(args))
	}

	s, err := j.GetString()
	if err == nil {
		return s
	}

	return def
}

func (j *XPJsonImpl) StringArrayValue(args ...[]string) []string {
	var def []string

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("StringArrayValue() received too many arguments %d", len(args))
	}

	a, err := j.GetStringArray()
	if err == nil {
		return a
	}

	return def
}

func (j *XPJsonImpl) IntValue(args ...int) int {
	var def int

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("IntValue() received too many arguments %d", len(args))
	}

	i, err := j.intValue()
	if err == nil {
		return i
	}

	return def
}

func (j *XPJsonImpl) Float64Value(args ...float64) float64 {
	var def float64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("Float64Value() received too many arguments %d", len(args))
	}

	f, err := j.float64Value()
	if err == nil {
		return f
	}

	return def
}

func (j *XPJsonImpl) BoolValue(args ...bool) bool {
	var def bool

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("BoolValue() received too many arguments %d", len(args))
	}

	b, err := j.GetBool()
	if err == nil {
		return b
	}

	return def
}

func (j *XPJsonImpl) Int64Value(args ...int64) int64 {
	var def int64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("Int64Value() received too many arguments %d", len(args))
	}

	i, err := j.int64Value()
	if err == nil {
		return i
	}

	return def
}

func (j *XPJsonImpl) UInt64Value(args ...uint64) uint64 {
	var def uint64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("UInt64Value() received too many arguments %d", len(args))
	}

	i, err := j.uint64Value()
	if err == nil {
		return i
	}

	return def
}