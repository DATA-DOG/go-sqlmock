package sqlmock

import (
	"database/sql/driver"
	"reflect"
)

// Argument interface allows to match
// any argument in specific way when used with
// ExpectedQuery and ExpectedExec expectations.
type Argument interface {
	Match(driver.Value) bool
}

// AnyArg will return an Argument which can
// match any kind of arguments.
//
// Useful for time.Time or similar kinds of arguments.
func AnyArg() Argument {
	return anyArgument{}
}

type anyArgument struct{}

func (a anyArgument) Match(_ driver.Value) bool {
	return true
}

// NotEmptyArg will return an Argument which can
// match any kind of non zero arguments.
//
// Logic by type:
// type     |answer | condition
// -----------------------
// int64     true    v != 0
// float64   true    v != 0.0
// bool      true    any values
// []byte    true    len(v) != 0
// string    true    v != ""
// time.Time true    non zero value
// nil       false
func NotEmptyArg() Argument {
	return notEmptyArgument{}
}

type notEmptyArgument struct{}

func (a notEmptyArgument) Match(v driver.Value) bool {
	if v == nil {
		return false
	}

	switch v.(type) {
	case bool:
		return true
	default:
		return v != reflect.Zero(reflect.TypeOf(v)).Interface()
	}
}
