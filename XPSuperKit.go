package XPSuperKit

import (
	"io"
	"reflect"
)

func XPFloat() *XPFloatImpl {
	return &(XPFloatImpl{})
}

func XPString() *XPStringImpl {
	return &(XPStringImpl{})
}

func XPNumber() *XPNumberImpl {
	return &(XPNumberImpl{})
}

func XPJson(args ...interface{}) (*XPJsonImpl, error) {
	switch len(args) {
	case 0:
		return NewJson(), nil
	case 1:
		t := reflect.TypeOf(args[0])

		switch t.Kind() {
		case reflect.String:
			return NewJsonFromString(reflect.ValueOf(args[0]).String())
		case reflect.Map, reflect.Struct:
			return NewJsonFromInterface(args[0]), nil
		default:
			return nil, ErrorN("XPJson received invalid value type on args[0]")
		}
	default:
		return nil, ErrorF("XPJson received too many arguments %d", len(args))
	}
}

func XPJsonFromBytes(b []byte) (*XPJsonImpl, error) {
	return NewJsonFromBytes(b)
}

func XPJsonFromReader(reader io.Reader) (*XPJsonImpl, error) {
	return NewJsonFromReader(reader)
}

func XPFilePath(args ...string) (*XPFilePathImpl, error) {
	switch len(args) {
	case 0:
		return NewFilePathFromCurrentPath()
	case 1:
		return NewFilePath(args[0])
	default:
		return nil, ErrorF("XPFilePath received too many arguments %d", len(args))
	}
}

func XPHttp() *XPHttpImpl {
	return NewHttp()
}

func XPIdGenerator(workerId int64) (*XPIdGeneratorImpl, error) {
	return NewIdGenerator(workerId)
}

func XPJwt() *XPJwtImpl {
	return &(XPJwtImpl{})
}

func XPStruct(s interface{}) *XPStructImpl {
	return NewXPStruct(s)
}

func XPConfig() *XPConfigImpl {
	return NewXPConfig(nil)
}

func XPWorker() *XPWorkerImpl {
	return NewXPWorker()
}

func XPAsync() *XPAsyncImpl {
	return NewAsync()
}

func XPQueue(capacity uint32) *XPQueueImpl {
	return NewXPQueue(capacity)
}

func XPMemoryCache(count int, capacity int) *XPMemoryCacheImpl {
	return NewMemoryCache(count, capacity)
}

func XPIP() *XPIPImpl {
	return &(XPIPImpl{})
}