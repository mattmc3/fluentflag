# fluentflag

> A fluent CLI flag builder for Go

**fluentflag** provides a fluent, type-safe API for defining and registering command-line flags in Go. It wraps the standard library's `flag` package, making it easier to declare flags, set defaults, add short aliases, and handle slices - all with a clean, chainable syntax.

**NOTE:** _Requires Go 1.18+_

## Why

Go's standard library `flag` package is simple and reliable, but it lacks some features:

-   Poor support for `-s` (short) and `--long` flags
-   Lackluster support for accumulating values into a slice
-   No use of Go generics for type safety and convenience

Many third-party argument parsers exist, but most are large and complex. **fluentflag** provides a few awesome enhancements to make `flag` more usable - support for short aliases, slice support for accumulating values, a better usage printer, and generics - in a single, lightweight file you can simply drop into your project or use as a module.

## Features

-   Type-safe flag registration (bool, string, int, int64, float64, uint, uint64)
-   Fluent API for chaining options
-   Short flag aliases (e.g. `-n` for `--name`)
-   Slice flag support (accumulate multiple values)
-   Works seamlessly with the Go standard library's `flag` package

## Install

```sh
go get github.com/mattmc3/fluentflag
```

## Example

```go
package main

import (
    "fmt"
    "github.com/mattmc3/fluentflag"
)

type MyOpts struct {
    Help          bool
    Name          string
    MinArgs       int
    MaxArgs       int
    IgnoreUnknown bool
    StopNonOpt    bool
    Exclusive     *[]int
    Version       bool
}

func main() {
    opts := MyOpts{}
    builder := fluentflag.NewFlagBuilder()
    builder.StringFlag("name", "Command name for error messages").
        Alias('n').
        Default("foo").
        Build(&opts.Name)
    builder.BoolFlag("help", "Show this help message").
        Alias('h').
        Build(&opts.Help)
    builder.IntFlag("min-args", "Minimum number of non-option arguments").
        Alias('N').
        Default(-1).
        Build(&opts.MinArgs)
    builder.IntFlag("max-args", "Maximum number of non-option arguments").
        Alias('X').
        Default(-1).
        Build(&opts.MaxArgs)
    builder.BoolFlag("ignore-unknown", "Ignore unknown options").
        Alias('i').
        Build(&opts.IgnoreUnknown)
    builder.BoolFlag("stop-nonopt", "Stop scanning at first non-option").
        Alias('s').
        Build(&opts.StopNonOpt)
    builder.BoolFlag("version", "Print version number").
        Alias('v').
        Build(&opts.Version)
    opts.Exclusive = builder.IntFlag("exclusive", "Comma-separated mutually exclusive options").
        Alias('x').
        BuildSlice()

    // Parse flags as usual
    flag.Parse()

    fmt.Printf("%+v\n", opts)
}
```

## Example CLI Usage

```sh
# Show help
argparser --help

# Set the name and version
argparser --name=bar --version

# Use short flags
argparser -n bar -v

# Set min and max args
argparser --min-args=2 --max-args=5

# Use slice flag multiple times
argparser --exclusive=1 --exclusive=2 --exclusive=3

# Mix short and long flags
argparser -n bar -X 10 --ignore-unknown --exclusive=7 --exclusive=8
```

## API

-   `NewFlagBuilder() *FlagBuilder`
    Create a new flag builder.
-   `StringFlag(name, usage string) *FluentFlag[string]`
    Create a new string flag.
-   `BoolFlag(name, usage string) *FluentFlag[bool]`
    Create a new boolean flag.
-   `IntFlag(name, usage string) *FluentFlag[int]`
    Create a new integer flag.
-   `.Alias(rune)`
    Set a short flag alias (e.g. `-n` for `--name`).
-   `.Default(value T)`
    Set a default value for the flag.
-   `.Build(ptr *T)`
    Register the flag and bind it to the provided variable.
-   `.BuildVar() *T`
    Register the flag and return a pointer to the storage variable.
-   `.BuildSlice() *[]T`
    Register a flag that accumulates values into a slice.
