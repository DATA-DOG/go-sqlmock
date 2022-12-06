package sqlmock

import "database/sql/driver"

// Matcher interface allows to match
// any argument in specific way when used with Expected expectations.
type Matcher interface {
	Match(driver.Value) bool
}

type MatchFunc func(driver.Value) bool

func (a MatchFunc) Match(v driver.Value) bool { return a(v) }

// Any will return an Matcher which can
// match any kind of arguments.
//
// Useful for time.Time or similar kinds of arguments.
func Any() Matcher {
	return MatchFunc(func(value driver.Value) bool { return true })
}

func Exec() Matcher {
	return MatchFunc(func(value driver.Value) bool { return value == "exec" })
}

func Query() Matcher {
	return MatchFunc(func(value driver.Value) bool { return value == "query" })
}
