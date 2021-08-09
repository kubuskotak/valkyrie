package valkyrie

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type structCopy struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

func TestCloneStruct(t *testing.T) {
	s := &structCopy{
		Name:    "name",
		Address: "address",
	}
	a := &structCopy{}
	err := CloneStruct(s, a)
	assert.NoError(t, err)
	assert.Equal(t, s, a)
	assert.Equal(t, s.Name, a.Name)
	assert.Equal(t, "name", a.Name)
}

func TestCloneStructNil(t *testing.T) {
	var (
		a          *structCopy = nil
		sliceCycle             = []interface{}{nil}
	)

	cc := &structCopy{
		Name:    "name",
		Address: "address",
	}
	er := &json.UnmarshalTypeError{
		Value:  "array",
		Type:   reflect.TypeOf(structCopy{}),
		Offset: 1,
		Struct: "",
		Field:  "",
	}
	tt := []struct {
		src  interface{}
		dest interface{}
		want error
	}{
		{cc, a, NilPointer},
		{cc, structCopy{}, NotPointer},
		{a, cc, NilPointer},
		{structCopy{}, cc, NotPointer},
		{&sliceCycle, cc, er},
	}

	for _, c := range tt {
		t.Run(fmt.Sprintf("src: %v - dest: %v", c.src, c.dest), func(t *testing.T) {
			err := CloneStruct(c.src, c.dest)
			t.Log(err)
			assert.Error(t, err)
			assert.EqualError(t, c.want, err.Error())
		})
	}
}
