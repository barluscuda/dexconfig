// Package dexconfig loads configuration from environment variables into a
// struct using `env` struct tags. It supports nested structs, pointer-to-struct
// fields, default values, required fields, slices, maps, custom unmarshalers,
// and time.Duration.
package dexconfig

import (
	"errors"
	"fmt"
	"reflect"
)

// Option configures LoadConfig behavior.
type Option func(*options)

type options struct {
	prefix    string
	lookup    LookupFunc
	tagName   string
	separator string
}

// LookupFunc resolves an environment variable name to its value and a bool
// indicating whether the variable was set. It mirrors os.LookupEnv.
type LookupFunc func(key string) (string, bool)

// WithPrefix sets a prefix that is prepended (with an underscore) to every
// environment variable key resolved during loading.
func WithPrefix(prefix string) Option {
	return func(o *options) { o.prefix = prefix }
}

// WithLookup overrides the function used to resolve environment variables.
// The default is os.LookupEnv.
func WithLookup(fn LookupFunc) Option {
	return func(o *options) {
		if fn != nil {
			o.lookup = fn
		}
	}
}

// WithTagName overrides the struct tag name used to read configuration
// directives. The default is "env".
func WithTagName(name string) Option {
	return func(o *options) {
		if name != "" {
			o.tagName = name
		}
	}
}

// WithSeparator overrides the separator used when parsing slice and map
// values. The default is ",".
func WithSeparator(sep string) Option {
	return func(o *options) {
		if sep != "" {
			o.separator = sep
		}
	}
}

// LoadConfig populates the struct pointed to by c from environment variables.
// c must be a non-nil pointer to a struct.
func LoadConfig(c any, opts ...Option) error {
	if c == nil {
		return errors.New("dexconfig: config must be a non-nil pointer")
	}

	val := reflect.ValueOf(c)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return errors.New("dexconfig: config must be a non-nil pointer")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return errors.New("dexconfig: config must point to a struct")
	}

	o := options{
		lookup:    defaultLookup,
		tagName:   "env",
		separator: ",",
	}
	for _, opt := range opts {
		opt(&o)
	}

	if err := loadStruct(val, &o); err != nil {
		return fmt.Errorf("dexconfig: %w", err)
	}
	return nil
}
