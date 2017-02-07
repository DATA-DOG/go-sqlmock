// +build go1.8

package sqlmock

import (
	"database/sql"
	"database/sql/driver"
	"testing"
)

func TestQueryExpectationNamedArgComparison(t *testing.T) {
	e := &queryBasedExpectation{}
	against := []namedValue{{Value: int64(5), Name: "id"}}
	if err := e.argsMatches(against); err != nil {
		t.Errorf("arguments should match, since the no expectation was set, but got err: %s", err)
	}

	e.args = []driver.Value{
		sql.Named("id", 5),
		sql.Named("s", "str"),
	}

	if err := e.argsMatches(against); err == nil {
		t.Error("arguments should not match, since the size is not the same")
	}

	against = []namedValue{
		{Value: int64(5), Name: "id"},
		{Value: "str", Name: "s"},
	}

	if err := e.argsMatches(against); err != nil {
		t.Errorf("arguments should have matched, but it did not: %v", err)
	}

	against = []namedValue{
		{Value: int64(5), Name: "id"},
		{Value: "str", Name: "username"},
	}

	if err := e.argsMatches(against); err == nil {
		t.Error("arguments matched, but it should have not due to Name")
	}

	e.args = []driver.Value{int64(5), "str"}

	against = []namedValue{
		{Value: int64(5), Ordinal: 0},
		{Value: "str", Ordinal: 1},
	}

	if err := e.argsMatches(against); err == nil {
		t.Error("arguments matched, but it should have not due to wrong Ordinal position")
	}

	against = []namedValue{
		{Value: int64(5), Ordinal: 1},
		{Value: "str", Ordinal: 2},
	}

	if err := e.argsMatches(against); err != nil {
		t.Errorf("arguments should have matched, but it did not: %v", err)
	}
}
