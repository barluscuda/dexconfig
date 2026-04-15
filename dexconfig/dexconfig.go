package dexconfig

import (
	"errors"
	"reflect"
)

func LoadConfig(c interface{}) error {
	val := reflect.ValueOf(c)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return errors.New("config must be a non-nil pointer")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return errors.New("config must point to a struct")
	}

	return loadStruct(val)
}