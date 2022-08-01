// +build go1.8

package sqlmock

import (
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"
)

func TestQueryExpectationArgComparison(t *testing.T) {
	e := &queryBasedExpectation{converter: driver.DefaultParameterConverter}
	against := []driver.NamedValue{{Value: int64(5), Ordinal: 1}}
	if err := e.argsMatches(against); err == nil {
		t.Errorf("arguments should not match, since no expectation was set, but argument was passed")
	}

	e.args = []driver.Value{5, "str"}

	against = []driver.NamedValue{{Value: int64(5), Ordinal: 1}}
	if err := e.argsMatches(against); err == nil {
		t.Error("arguments should not match, since the size is not the same")
	}

	against = []driver.NamedValue{
		{Value: int64(3), Ordinal: 1},
		{Value: "str", Ordinal: 2},
	}
	if err := e.argsMatches(against); err == nil {
		t.Error("arguments should not match, since the first argument (int value) is different")
	}

	against = []driver.NamedValue{
		{Value: int64(5), Ordinal: 1},
		{Value: "st", Ordinal: 2},
	}
	if err := e.argsMatches(against); err == nil {
		t.Error("arguments should not match, since the second argument (string value) is different")
	}

	against = []driver.NamedValue{
		{Value: int64(5), Ordinal: 1},
		{Value: "str", Ordinal: 2},
	}
	if err := e.argsMatches(against); err != nil {
		t.Errorf("arguments should match, but it did not: %s", err)
	}

	const longForm = "Jan 2, 2006 at 3:04pm (MST)"
	tm, _ := time.Parse(longForm, "Feb 3, 2013 at 7:54pm (PST)")
	e.args = []driver.Value{5, tm}

	against = []driver.NamedValue{
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
	against := []driver.NamedValue{
		{Value: true, Ordinal: 1},
	}
	if err := e.argsMatches(against); err != nil {
		t.Error("arguments should match, since arguments are the same")
	}

	e = &queryBasedExpectation{args: []driver.Value{false}, converter: driver.DefaultParameterConverter}
	against = []driver.NamedValue{
		{Value: false, Ordinal: 1},
	}
	if err := e.argsMatches(against); err != nil {
		t.Error("arguments should match, since argument are the same")
	}

	e = &queryBasedExpectation{args: []driver.Value{true}, converter: driver.DefaultParameterConverter}
	against = []driver.NamedValue{
		{Value: false, Ordinal: 1},
	}
	if err := e.argsMatches(against); err == nil {
		t.Error("arguments should not match, since argument is different")
	}

	e = &queryBasedExpectation{args: []driver.Value{false}, converter: driver.DefaultParameterConverter}
	against = []driver.NamedValue{
		{Value: true, Ordinal: 1},
	}
	if err := e.argsMatches(against); err == nil {
		t.Error("arguments should not match, since argument is different")
	}
}

func TestQueryExpectationNamedArgComparison(t *testing.T) {
	e := &queryBasedExpectation{converter: driver.DefaultParameterConverter}
	against := []driver.NamedValue{{Value: int64(5), Name: "id"}}
	if err := e.argsMatches(against); err == nil {
		t.Errorf("arguments should not match, since no expectation was set, but argument was passed")
	}

	e.args = []driver.Value{
		sql.Named("id", 5),
		sql.Named("s", "str"),
	}

	if err := e.argsMatches(against); err == nil {
		t.Error("arguments should not match, since the size is not the same")
	}

	against = []driver.NamedValue{
		{Value: int64(5), Name: "id"},
		{Value: "str", Name: "s"},
	}

	if err := e.argsMatches(against); err != nil {
		t.Errorf("arguments should have matched, but it did not: %v", err)
	}

	against = []driver.NamedValue{
		{Value: int64(5), Name: "id"},
		{Value: "str", Name: "username"},
	}

	if err := e.argsMatches(against); err == nil {
		t.Error("arguments matched, but it should have not due to Name")
	}

	e.args = []driver.Value{int64(5), "str"}

	against = []driver.NamedValue{
		{Value: int64(5), Ordinal: 0},
		{Value: "str", Ordinal: 1},
	}

	if err := e.argsMatches(against); err == nil {
		t.Error("arguments matched, but it should have not due to wrong Ordinal position")
	}

	against = []driver.NamedValue{
		{Value: int64(5), Ordinal: 1},
		{Value: "str", Ordinal: 2},
	}

	if err := e.argsMatches(against); err != nil {
		t.Errorf("arguments should have matched, but it did not: %v", err)
	}
}

type panicConverter struct {
}

func (s panicConverter) ConvertValue(v interface{}) (driver.Value, error) {
	panic(v)
}

func Test_queryBasedExpectation_attemptArgMatch(t *testing.T) {
	e := &queryBasedExpectation{converter: new(panicConverter), args: []driver.Value{"test"}}
	values := []driver.NamedValue{
		{Ordinal: 1, Name: "test", Value: "test"},
	}
	if err := e.attemptArgMatch(values); err == nil {
		t.Errorf("error expected")
	}
}
