package dexconfig

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func loadStruct(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			if err := loadStruct(field.Elem()); err != nil {
				return err
			}
			continue
		}

		if field.Kind() == reflect.Struct && field.Type() != reflect.TypeOf(time.Duration(0)) {
			if err := loadStruct(field); err != nil {
				return err
			}
			continue
		}

		if !field.CanSet() {
			continue
		}

		tag := fieldType.Tag.Get("env")
		if tag == "" {
			continue
		}

		key, def := parseTag(tag)

		val := os.Getenv(key)
		if val == "" {
			val = def
		}

		if err := setValue(field, val); err != nil {
			return errors.New(fieldType.Name + ": " + err.Error())
		}
	}
	return nil
}

func parseTag(tag string) (key string, def string) {
	parts := strings.SplitN(tag, ";", 2)

	key = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		def = parts[1]
	}
	return
}

func setValue(field reflect.Value, val string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(val)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(val)
			if err != nil {
				return err
			}
			field.SetInt(int64(d))
			return nil
		}

		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(i)

	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		field.SetBool(b)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)

	default:
		return errors.New("unsupport type: " + field.Kind().String())
	}

	return nil
}
