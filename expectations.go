package sqlmock

import (
	"database/sql/driver"
	"reflect"
	"regexp"
)

// Argument interface allows to match
// any argument in specific way
type Argument interface {
	Match(driver.Value) bool
}

// an expectation interface
type expectation interface {
	fulfilled() bool
	setError(err error)
}

// common expectation struct
// satisfies the expectation interface
type commonExpectation struct {
	triggered bool
	err       error
}

func (e *commonExpectation) fulfilled() bool {
	return e.triggered
}

func (e *commonExpectation) setError(err error) {
	e.err = err
}

// query based expectation
// adds a query matching logic
type queryBasedExpectation struct {
	commonExpectation
	sqlRegex *regexp.Regexp
	args     []driver.Value
}

func (e *queryBasedExpectation) queryMatches(sql string) bool {
	return e.sqlRegex.MatchString(sql)
}

func (e *queryBasedExpectation) argsMatches(args []driver.Value) bool {
	if nil == e.args {
		return true
	}
	if len(args) != len(e.args) {
		return false
	}
	for k, v := range args {
		matcher, ok := e.args[k].(Argument)
		if ok {
			if !matcher.Match(v) {
				return false
			}
			continue
		}
		vi := reflect.ValueOf(v)
		ai := reflect.ValueOf(e.args[k])
		switch vi.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if vi.Int() != ai.Int() {
				return false
			}
		case reflect.Float32, reflect.Float64:
			if vi.Float() != ai.Float() {
				return false
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if vi.Uint() != ai.Uint() {
				return false
			}
		case reflect.String:
			if vi.String() != ai.String() {
				return false
			}
		default:
			// compare types like time.Time based on type only
			if vi.Kind() != ai.Kind() {
				return false
			}
		}
	}
	return true
}

// begin transaction
type expectedBegin struct {
	commonExpectation
}

// tx commit
type expectedCommit struct {
	commonExpectation
}

// tx rollback
type expectedRollback struct {
	commonExpectation
}

// query expectation
type expectedQuery struct {
	queryBasedExpectation

	rows driver.Rows
}

// exec query expectation
type expectedExec struct {
	queryBasedExpectation

	result driver.Result
}

// Prepare expectation
type expectedPrepare struct {
	commonExpectation

	statement driver.Stmt
}
