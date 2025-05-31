// fluentflag.go
// Copyright (c) 2025 mattmc3
// SPDX-License-Identifier: MIT
// Project home: https://github.com/mattmc3/fluentflag

package fluentflag

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
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
func (self *accumValues[T]) String() string {
	if self.target == nil {
		return "[]"
	}
	return fmt.Sprintf("%v", *self.target)
}

// Set appends a new value to the slice.
func (self *accumValues[T]) Set(val string) error {
	parsed, err := parse[T](val)
	if err != nil {
		return err
	}
	*self.target = append(*self.target, parsed)
	return nil
}

// Opt is a CLI option
type FluentFlag[T FlagType] struct {
	builder    *FlagBuilder
	name       string
	alias      rune
	defaultVal T
	usage      string
}

// Alias sets a short flag (eg: -f) alias for the standard long flag.
func (self *FluentFlag[T]) Alias(alias rune) *FluentFlag[T] {
	self.alias = alias
	return self
}

// Default sets the default value for the flag.
func (self *FluentFlag[T]) Default(defaultVal T) *FluentFlag[T] {
	self.defaultVal = defaultVal
	return self
}

// Build registers the flag with the standard library flag package using the provided pointer.
func (self *FluentFlag[T]) Build(ptr *T) {
	self.builder.flagsBuilt = append(self.builder.flagsBuilt, self)
	self.builder.building = nil
	switch any(self.defaultVal).(type) {
	case bool:
		self.builder.flagSet.BoolVar(any(ptr).(*bool), self.name, any(self.defaultVal).(bool), self.usage)
		if self.alias != 0 {
			self.builder.flagSet.BoolVar(any(ptr).(*bool), string(self.alias), any(self.defaultVal).(bool), "")
		}
	case int:
		self.builder.flagSet.IntVar(any(ptr).(*int), self.name, any(self.defaultVal).(int), self.usage)
		if self.alias != 0 {
			self.builder.flagSet.IntVar(any(ptr).(*int), string(self.alias), any(self.defaultVal).(int), "")
		}
	case int64:
		self.builder.flagSet.Int64Var(any(ptr).(*int64), self.name, any(self.defaultVal).(int64), self.usage)
		if self.alias != 0 {
			self.builder.flagSet.Int64Var(any(ptr).(*int64), string(self.alias), any(self.defaultVal).(int64), "")
		}
	case float64:
		self.builder.flagSet.Float64Var(any(ptr).(*float64), self.name, any(self.defaultVal).(float64), self.usage)
		if self.alias != 0 {
			self.builder.flagSet.Float64Var(any(ptr).(*float64), string(self.alias), any(self.defaultVal).(float64), "")
		}
	case string:
		self.builder.flagSet.StringVar(any(ptr).(*string), self.name, any(self.defaultVal).(string), self.usage)
		if self.alias != 0 {
			self.builder.flagSet.StringVar(any(ptr).(*string), string(self.alias), any(self.defaultVal).(string), "")
		}
	case uint:
		self.builder.flagSet.UintVar(any(ptr).(*uint), self.name, any(self.defaultVal).(uint), self.usage)
		if self.alias != 0 {
			self.builder.flagSet.UintVar(any(ptr).(*uint), string(self.alias), any(self.defaultVal).(uint), "")
		}
	case uint64:
		self.builder.flagSet.Uint64Var(any(ptr).(*uint64), self.name, any(self.defaultVal).(uint64), self.usage)
		if self.alias != 0 {
			self.builder.flagSet.Uint64Var(any(ptr).(*uint64), string(self.alias), any(self.defaultVal).(uint64), "")
		}
	default:
		panic("unsupported flag type")
	}
}

// BuildVar registers the flag and returns a pointer to the storage variable.
func (self *FluentFlag[T]) BuildVar() *T {
	var v T
	self.Build(&v)
	return &v
}

// BuildSlice registers a flag that accumulates values into a slice of T.
// Returns a pointer to the slice ([]T) that the user can use directly.
func (self *FluentFlag[T]) BuildSlice() *[]T {
	self.builder.flagsBuilt = append(self.builder.flagsBuilt, self)
	self.builder.building = nil
	slice := new([]T) // allocate on heap
	*slice = []T{}
	val := &accumValues[T]{target: slice}
	self.builder.flagSet.Var(val, self.name, self.usage)
	if self.alias != 0 {
		self.builder.flagSet.Var(val, string(self.alias), "")
	}
	return slice
}

// FluentFlag provides usage/help string for the option.
func (self *FluentFlag[T]) Usage() string {
	typeStr := fmt.Sprintf("%T", self.defaultVal)
	if dot := strings.LastIndex(typeStr, "."); dot != -1 {
		typeStr = typeStr[dot+1:]
	}
	if typeStr == "bool" {
		typeStr = ""
	} else {
		typeStr = " " + typeStr
	}

	def := ""
	var zero T
	switch val := any(self.defaultVal).(type) {
	case bool:
		if val {
			def = " (default true)"
		}
	case string:
		if val != "" {
			def = fmt.Sprintf(" (default %q)", val)
		}
	default:
		if self.defaultVal != zero {
			def = fmt.Sprintf(" (default %v)", val)
		}
	}

	names := ""
	if self.alias != 0 {
		names = fmt.Sprintf("-%c, --%s", self.alias, self.name)
	} else {
		names = fmt.Sprintf("    --%s", self.name)
	}
	line := fmt.Sprintf("%s%s", names, typeStr)
	const maxLen = 25
	if len(line) >= maxLen {
		return fmt.Sprintf("  %-*s\n  %-*s%s%s", maxLen, line, maxLen, "", self.usage, def)
	}
	return fmt.Sprintf("  %-*s%s%s", maxLen, line, self.usage, def)
}

// FlagBuilder provides a fluent API for building and registering command-line flags.
type FlagBuilder struct {
	flagSet    *flag.FlagSet
	flagsBuilt []any     // store built flags
	building   any       // store the currently building flag
	output     io.Writer // optional output writer for usage
}

// SetOutput sets the output writer for usage/help text.
func (b *FlagBuilder) SetOutput(w io.Writer) {
	b.output = w
}

// NewFlagBuilder creates a new FlagBuilder using flag.CommandLine.
func NewFlagBuilder() *FlagBuilder {
	return &FlagBuilder{flagSet: flag.CommandLine}
}

// NewFlagBuilderForSet creates a new FlagBuilder with a custom FlagSet.
func NewFlagBuilderWithSet(flagSet *flag.FlagSet) *FlagBuilder {
	if flagSet == nil {
		flagSet = flag.CommandLine
	}
	return &FlagBuilder{flagSet: flagSet}
}

// BoolFlag defines a boolean flag
func (self *FlagBuilder) BoolFlag(name, usage string) *FluentFlag[bool] {
	return newFlag[bool](self, name, usage)
}

// StringFlag defines a string flag
func (self *FlagBuilder) StringFlag(name, usage string) *FluentFlag[string] {
	return newFlag[string](self, name, usage)
}

// IntFlag defines an int flag
func (self *FlagBuilder) IntFlag(name, usage string) *FluentFlag[int] {
	return newFlag[int](self, name, usage)
}

// Int64Flag defines an int64 flag
func (self *FlagBuilder) Int64Flag(name, usage string) *FluentFlag[int64] {
	return newFlag[int64](self, name, usage)
}

// Float64Flag defines a float64 flag
func (self *FlagBuilder) Float64Flag(name, usage string) *FluentFlag[float64] {
	return newFlag[float64](self, name, usage)
}

// UintFlag defines a uint flag
func (self *FlagBuilder) UintFlag(name, usage string) *FluentFlag[uint] {
	return newFlag[uint](self, name, usage)
}

// Uint64Flag defines a uint64 flag
func (self *FlagBuilder) Uint64Flag(name, usage string) *FluentFlag[uint64] {
	return newFlag[uint64](self, name, usage)
}

// NewFlagBuilder creates a new FlagBuilder for the given flag name and usage description.
func newFlag[T FlagType](builder *FlagBuilder, name, usage string) *FluentFlag[T] {
	if builder.building != nil {
		panic("fluentflag: previous flag not built (call Build, BuildVar, or BuildSlice)")
	}
	flag := &FluentFlag[T]{
		builder: builder,
		name:    name,
		usage:   usage,
	}
	builder.building = flag
	return flag
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

// PrintUsage prints usage for all built flags.
func (b *FlagBuilder) PrintUsage() {
	w := b.output
	if w == nil {
		w = os.Stderr
	}
	for _, f := range b.flagsBuilt {
		if u, ok := f.(interface{ Usage() string }); ok {
			fmt.Fprintln(w, u.Usage())
		}
	}
}
