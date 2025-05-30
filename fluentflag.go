// fluentflag.go
// Copyright (c) 2025 mattmc3
// SPDX-License-Identifier: MIT
// Project home: https://github.com/mattmc3/fluentflag

package fluentflag

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
)

// FlagType is a type constraint for the basic flag data types supported by FlagBuilder.
type FlagType interface {
	~bool | ~string | ~int | ~int64 | ~float64 | ~uint | ~uint64
}

// accumValues implements flag.Value for accumulating values into a slice.
type accumValues[T FlagType] struct {
	target *[]T
}

// String returns the string representation of the accumulated slice.
func (s *accumValues[T]) String() string {
	if s.target == nil {
		return "[]"
	}
	return fmt.Sprintf("%v", *s.target)
}

// Set appends a new value to the slice.
func (s *accumValues[T]) Set(val string) error {
	parsed, err := parse[T](val)
	if err != nil {
		return err
	}
	*s.target = append(*s.target, parsed)
	return nil
}

// FlagBuilder provides a fluent API for building and registering command-line flags.
type FlagBuilder[T FlagType] struct {
	name       string
	alias      rune
	defaultVal T
	usage      string
}

// NewFlagBuilder creates a new FlagBuilder for the given flag name and usage description.
func NewFlagBuilder[T FlagType](name, usage string) *FlagBuilder[T] {
	return &FlagBuilder[T]{
		name:  name,
		usage: usage,
	}
}

// Alias sets a short flag (eg: -f) alias for the standard long flag.
func (b *FlagBuilder[T]) Alias(alias rune) *FlagBuilder[T] {
	b.alias = alias
	return b
}

// Default sets the default value for the flag.
func (b *FlagBuilder[T]) Default(defaultVal T) *FlagBuilder[T] {
	b.defaultVal = defaultVal
	return b
}

// Build registers the flag with the standard library flag package using the provided pointer.
func (b *FlagBuilder[T]) Build(ptr *T) {
	switch any(b.defaultVal).(type) {
	case bool:
		flag.BoolVar(any(ptr).(*bool), b.name, any(b.defaultVal).(bool), b.usage)
		if b.alias != 0 {
			flag.BoolVar(any(ptr).(*bool), string(b.alias), any(b.defaultVal).(bool), "")
		}
	case int:
		flag.IntVar(any(ptr).(*int), b.name, any(b.defaultVal).(int), b.usage)
		if b.alias != 0 {
			flag.IntVar(any(ptr).(*int), string(b.alias), any(b.defaultVal).(int), "")
		}
	case int64:
		flag.Int64Var(any(ptr).(*int64), b.name, any(b.defaultVal).(int64), b.usage)
		if b.alias != 0 {
			flag.Int64Var(any(ptr).(*int64), string(b.alias), any(b.defaultVal).(int64), "")
		}
	case float64:
		flag.Float64Var(any(ptr).(*float64), b.name, any(b.defaultVal).(float64), b.usage)
		if b.alias != 0 {
			flag.Float64Var(any(ptr).(*float64), string(b.alias), any(b.defaultVal).(float64), "")
		}
	case string:
		flag.StringVar(any(ptr).(*string), b.name, any(b.defaultVal).(string), b.usage)
		if b.alias != 0 {
			flag.StringVar(any(ptr).(*string), string(b.alias), any(b.defaultVal).(string), "")
		}
	case uint:
		flag.UintVar(any(ptr).(*uint), b.name, any(b.defaultVal).(uint), b.usage)
		if b.alias != 0 {
			flag.UintVar(any(ptr).(*uint), string(b.alias), any(b.defaultVal).(uint), "")
		}
	case uint64:
		flag.Uint64Var(any(ptr).(*uint64), b.name, any(b.defaultVal).(uint64), b.usage)
		if b.alias != 0 {
			flag.Uint64Var(any(ptr).(*uint64), string(b.alias), any(b.defaultVal).(uint64), "")
		}
	default:
		panic("unsupported flag type")
	}
}

// BuildVar registers the flag and returns a pointer to the storage variable.
func (b *FlagBuilder[T]) BuildVar() *T {
	var v T
	b.Build(&v)
	return &v
}

// BuildSlice registers a flag that accumulates values into a slice of T.
// Returns a pointer to the slice ([]T) that the user can use directly.
func (b *FlagBuilder[T]) BuildSlice() *[]T {
	slice := new([]T) // allocate on heap
	*slice = []T{}
	val := &accumValues[T]{target: slice}
	flag.Var(val, b.name, b.usage)
	if b.alias != 0 {
		flag.Var(val, string(b.alias), "")
	}
	return slice
}

// Parse turns a string into the data type for a flag
func parse[T FlagType](s string) (T, error) {
	var v T
	switch any(v).(type) {
	case bool:
		v, err := strconv.ParseBool(s)
		return any(v).(T), err
	case string:
		return any(s).(T), nil
	case int:
		v, err := strconv.Atoi(s)
		return any(v).(T), err
	case int64:
		v, err := strconv.ParseInt(s, 10, 64)
		return any(v).(T), err
	case float64:
		v, err := strconv.ParseFloat(s, 64)
		return any(v).(T), err
	case uint:
		v, err := strconv.ParseUint(s, 10, 0)
		return any(uint(v)).(T), err
	case uint64:
		v, err := strconv.ParseUint(s, 10, 64)
		return any(v).(T), err
	default:
		return v, errors.New("unsupported flag type")
	}
}
