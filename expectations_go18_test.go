//go:build go1.8
// +build go1.8

package sqlmock

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestQueryExpectationArgComparison(t *testing.T) {
	e := &queryBasedExpectation{converter: driver.DefaultParameterConverter, noArgs: true}
	against := []driver.NamedValue{{Value: int64(5), Ordinal: 1}}
	if err := e.argsMatches(against); err == nil {
		t.Error("arguments should not match, since argument was passed, but noArgs was set")
	}

	e.noArgs = false
	if err := e.argsMatches(against); err != nil {
		t.Error("arguments should match, since argument was passed, but no expected args or noArgs was set")
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
	e := &queryBasedExpectation{converter: driver.DefaultParameterConverter, noArgs: true}
	against := []driver.NamedValue{{Value: int64(5), Name: "id"}}
	if err := e.argsMatches(against); err == nil {
		t.Error("arguments should not match, since argument was passed, but noArgs was set")
	}

	e.noArgs = false
	if err := e.argsMatches(against); err != nil {
		t.Error("arguments should match, since argument was passed, but no expected args or noArgs was set")
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
	values := []driver.NamedValue{
		{Ordinal: 1, Name: "test", Value: "test"},
	}
	for _, tc := range []struct {
		name  string
		value driver.Value
		want  string
	}{
		{name: "string", value: "test", want: "test"},
		{name: "error", value: errors.New("test error"), want: "test error"},
		{name: "integer", value: 42, want: "42"},
		{name: "percent", value: "100% mismatch", want: "100% mismatch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("attemptArgMatch panicked: %v", r)
				}
			}()
			e := &queryBasedExpectation{converter: new(panicConverter), args: []driver.Value{tc.value}}
			if err := e.attemptArgMatch(values); err == nil || err.Error() != tc.want {
				t.Errorf("got error %v, want %q", err, tc.want)
			}
		})
	}
}

type int64Argument int64

func (a int64Argument) Match(v driver.Value) bool {
	return v.(int64) == int64(a)
}

func TestUnorderedArgumentsRecoverTypeMismatch(t *testing.T) {
	for _, query := range []bool{false, true} {
		t.Run(fmt.Sprint("query=", query), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("matching a later expectation panicked: %v", r)
				}
			}()
			db, mock, err := New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			mock.MatchExpectationsInOrder(false)
			for _, arg := range []driver.Value{int64Argument(1), "one"} {
				if query {
					mock.ExpectQuery("SELECT").WithArgs(arg).WillReturnRows(NewRows([]string{"value"}).AddRow("matched"))
				} else {
					mock.ExpectExec("UPDATE").WithArgs(arg).WillReturnResult(NewResult(0, 1))
				}
			}
			for _, arg := range []interface{}{"one", int64(1)} {
				if query {
					var result string
					if err := db.QueryRow("SELECT", arg).Scan(&result); err != nil || result != "matched" {
						t.Fatalf("query result %q, error %v", result, err)
					}
				} else if _, err := db.Exec("UPDATE", arg); err != nil {
					t.Fatal(err)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
