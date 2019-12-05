// +build !go1.8

package sqlmock

import (
	"database/sql/driver"
	"testing"
	"time"
)

func TestQueryExpectationArgComparison(t *testing.T) {
	e := &queryBasedExpectation{converter: driver.DefaultParameterConverter}
	against := []namedValue{{Value: int64(5), Ordinal: 1}}
	if err := e.argsMatches(against); err != nil {
		t.Errorf("arguments should match, since the no expectation was set, but got err: %s", err)
	}

	e.args = []driver.Value{5, "str"}

	against = []namedValue{{Value: int64(5), Ordinal: 1}}
	if err := e.argsMatches(against); err == nil {
		t.Error("arguments should not match, since the size is not the same")
	}

	against = []namedValue{
		{Value: int64(3), Ordinal: 1},
		{Value: "str", Ordinal: 2},
	}
	if err := e.argsMatches(against); err == nil {
		t.Error("arguments should not match, since the first argument (int value) is different")
	}

	against = []namedValue{
		{Value: int64(5), Ordinal: 1},
		{Value: "st", Ordinal: 2},
	}
	if err := e.argsMatches(against); err == nil {
		t.Error("arguments should not match, since the second argument (string value) is different")
	}

	against = []namedValue{
		{Value: int64(5), Ordinal: 1},
		{Value: "str", Ordinal: 2},
	}
	if err := e.argsMatches(against); err != nil {
		t.Errorf("arguments should match, but it did not: %s", err)
	}

	const longForm = "Jan 2, 2006 at 3:04pm (MST)"
	tm, _ := time.Parse(longForm, "Feb 3, 2013 at 7:54pm (PST)")
	e.args = []driver.Value{5, tm}

	against = []namedValue{
		{Value: int64(5), Ordinal: 1},
		{Value: tm, Ordinal: 2},
	}
	if err := e.argsMatches(against); err != nil {
		t.Error("arguments should match, but it did not")
	}

	e.args = []driver.Value{5, AnyArg()}
	if err := e.argsMatches(against); err != nil {
		t.Errorf("arguments should match, but it did not: %s", err)
	}
}

func TestQueryExpectationArgComparisonBool(t *testing.T) {
	var e *queryBasedExpectation

	e = &queryBasedExpectation{args: []driver.Value{true}, converter: driver.DefaultParameterConverter}
	against := []namedValue{
		{Value: true, Ordinal: 1},
	}
	if err := e.argsMatches(against); err != nil {
		t.Error("arguments should match, since arguments are the same")
	}

	e = &queryBasedExpectation{args: []driver.Value{false}, converter: driver.DefaultParameterConverter}
	against = []namedValue{
		{Value: false, Ordinal: 1},
	}
	if err := e.argsMatches(against); err != nil {
		t.Error("arguments should match, since argument are the same")
	}

	e = &queryBasedExpectation{args: []driver.Value{true}, converter: driver.DefaultParameterConverter}
	against = []namedValue{
		{Value: false, Ordinal: 1},
	}
	if err := e.argsMatches(against); err == nil {
		t.Error("arguments should not match, since argument is different")
	}

	e = &queryBasedExpectation{args: []driver.Value{false}, converter: driver.DefaultParameterConverter}
	against = []namedValue{
		{Value: true, Ordinal: 1},
	}
	if err := e.argsMatches(against); err == nil {
		t.Error("arguments should not match, since argument is different")
	}
}

type panicConverter struct {
}

func (s panicConverter) ConvertValue(v interface{}) (driver.Value, error) {
	panic(v)
}

func Test_queryBasedExpectation_attemptArgMatch(t *testing.T) {
	e := &queryBasedExpectation{converter: new(panicConverter), args: []driver.Value{"test"}}
	values := []namedValue{
		{Ordinal: 1, Name: "test", Value: "test"},
	}
	if err := e.attemptArgMatch(values); err == nil {
		t.Errorf("error expected")
	}
}
