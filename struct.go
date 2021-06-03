package valkyrie

import (
	"encoding/json"
	"errors"
	"reflect"
)

var (
	NotPointer = errors.New("must pass a pointer, not a value")
	NilPointer = errors.New("nil pointer passed")
)

func CloneStruct(src, dest interface{}) error {
	d := reflect.ValueOf(dest)
	if d.Kind() != reflect.Ptr {
		return NotPointer
	}
	if d.IsNil() {
		return NilPointer
	}
	s := reflect.ValueOf(src)
	if s.Kind() != reflect.Ptr {
		return NotPointer
	}
	if s.IsNil() {
		return NilPointer
	}

	b, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}
