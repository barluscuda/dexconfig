package dexconfig

import (
	"encoding"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	durationType    = reflect.TypeOf(time.Duration(0))
	timeType        = reflect.TypeOf(time.Time{})
	textUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

// FieldError describes a failure to load a single struct field.
type FieldError struct {
	Field string
	Key   string
	Err   error
}

func (e *FieldError) Error() string {
	if e.Key != "" {
		return fmt.Sprintf("field %q (env %q): %v", e.Field, e.Key, e.Err)
	}
	return fmt.Sprintf("field %q: %v", e.Field, e.Err)
}

func (e *FieldError) Unwrap() error { return e.Err }

type tagInfo struct {
	key      string
	def      string
	hasDef   bool
	required bool
	skip     bool
}

func defaultLookup(key string) (string, bool) { return os.LookupEnv(key) }

func loadStruct(v reflect.Value, o *options) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !fieldType.IsExported() {
			continue
		}

		if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct &&
			!implementsTextUnmarshaler(field.Type()) && field.Type().Elem() != timeType {
			if field.IsNil() {
				if !field.CanSet() {
					continue
				}
				field.Set(reflect.New(field.Type().Elem()))
			}
			if err := loadStruct(field.Elem(), o); err != nil {
				return err
			}
			continue
		}

		if field.Kind() == reflect.Struct &&
			field.Type() != durationType &&
			field.Type() != timeType &&
			!implementsTextUnmarshaler(reflect.PointerTo(field.Type())) {
			if err := loadStruct(field, o); err != nil {
				return err
			}
			continue
		}

		if !field.CanSet() {
			continue
		}

		raw := fieldType.Tag.Get(o.tagName)
		if raw == "" {
			continue
		}

		info := parseTag(raw)
		if info.skip {
			continue
		}

		key := info.key
		if o.prefix != "" && key != "" {
			key = o.prefix + "_" + key
		}

		val, found := o.lookup(key)
		if !found || val == "" {
			if info.required {
				return &FieldError{Field: fieldType.Name, Key: key, Err: errors.New("required environment variable not set")}
			}
			if info.hasDef {
				val = info.def
			} else {
				continue
			}
		}

		if err := setValue(field, val, o.separator); err != nil {
			return &FieldError{Field: fieldType.Name, Key: key, Err: err}
		}
	}
	return nil
}

func parseTag(tag string) tagInfo {
	info := tagInfo{}

	parts := strings.SplitN(tag, ":", 2)
	info.key = strings.TrimSpace(parts[0])

	if info.key == "-" {
		info.skip = true
		return info
	}

	if len(parts) > 1 {
		rest := parts[1]
		if idx := strings.Index(rest, ";required"); idx >= 0 {
			info.required = true
			rest = rest[:idx] + rest[idx+len(";required"):]
		} else if rest == "required" {
			info.required = true
			rest = ""
		}
		if rest != "" || strings.Contains(parts[1], ":") {
			info.def = rest
			info.hasDef = true
		}
	}
	return info
}

func implementsTextUnmarshaler(t reflect.Type) bool {
	return t.Implements(textUnmarshaler)
}

func setValue(field reflect.Value, val string, sep string) error {
	if field.CanAddr() {
		if u, ok := field.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return u.UnmarshalText([]byte(val))
		}
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(val)
		return nil

	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		field.SetBool(b)
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == durationType {
			d, err := time.ParseDuration(val)
			if err != nil {
				return err
			}
			field.SetInt(int64(d))
			return nil
		}
		i, err := strconv.ParseInt(val, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetInt(i)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(val, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetUint(u)
		return nil

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetFloat(f)
		return nil

	case reflect.Slice:
		return setSlice(field, val, sep)

	case reflect.Map:
		return setMap(field, val, sep)


	case reflect.Ptr:
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setValue(field.Elem(), val, sep)
	}

	return fmt.Errorf("unsupported type: %s", field.Type().String())
}

func setSlice(field reflect.Value, val string, sep string) error {
	if val == "" {
		field.Set(reflect.MakeSlice(field.Type(), 0, 0))
		return nil
	}
	parts := strings.Split(val, sep)
	slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))
	for i, p := range parts {
		if err := setValue(slice.Index(i), strings.TrimSpace(p), sep); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	field.Set(slice)
	return nil
}

func setMap(field reflect.Value, val string, sep string) error {
	m := reflect.MakeMap(field.Type())
	if val == "" {
		field.Set(m)
		return nil
	}
	keyType := field.Type().Key()
	elemType := field.Type().Elem()

	for _, pair := range strings.Split(val, sep) {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) != 2 {
			return fmt.Errorf("invalid map entry %q (expected key:value)", pair)
		}
		k := reflect.New(keyType).Elem()
		if err := setValue(k, strings.TrimSpace(kv[0]), sep); err != nil {
			return fmt.Errorf("map key %q: %w", kv[0], err)
		}
		e := reflect.New(elemType).Elem()
		if err := setValue(e, strings.TrimSpace(kv[1]), sep); err != nil {
			return fmt.Errorf("map value for %q: %w", kv[0], err)
		}
		m.SetMapIndex(k, e)
	}
	field.Set(m)
	return nil
}
