package sqlmock

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"
	"time"
)

func TestQueryExpectationArgComparison(t *testing.T) {
	e := &queryBasedExpectation{}
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

	e = &queryBasedExpectation{args: []driver.Value{true}}
	against := []namedValue{
		{Value: true, Ordinal: 1},
	}
	if err := e.argsMatches(against); err != nil {
		t.Error("arguments should match, since arguments are the same")
	}

	e = &queryBasedExpectation{args: []driver.Value{false}}
	against = []namedValue{
		{Value: false, Ordinal: 1},
	}
	if err := e.argsMatches(against); err != nil {
		t.Error("arguments should match, since argument are the same")
	}

	e = &queryBasedExpectation{args: []driver.Value{true}}
	against = []namedValue{
		{Value: false, Ordinal: 1},
	}
	if err := e.argsMatches(against); err == nil {
		t.Error("arguments should not match, since argument is different")
	}

	e = &queryBasedExpectation{args: []driver.Value{false}}
	against = []namedValue{
		{Value: true, Ordinal: 1},
	}
	if err := e.argsMatches(against); err == nil {
		t.Error("arguments should not match, since argument is different")
	}
}

func TestQueryExpectationSqlMatch(t *testing.T) {
	e := &ExpectedExec{}

	e.sqlRegex = regexp.MustCompile("SELECT x FROM")
	if !e.queryMatches("SELECT x FROM someting") {
		t.Errorf("Sql must have matched the query")
	}

	e.sqlRegex = regexp.MustCompile("SELECT COUNT\\(x\\) FROM")
	if !e.queryMatches("SELECT COUNT(x) FROM someting") {
		t.Errorf("Sql must have matched the query")
	}
}

func ExampleExpectedExec() {
	db, mock, _ := New()
	result := NewErrorResult(fmt.Errorf("some error"))
	mock.ExpectExec("^INSERT (.+)").WillReturnResult(result)
	res, _ := db.Exec("INSERT something")
	_, err := res.LastInsertId()
	fmt.Println(err)
	// Output: some error
}

func TestBuildQuery(t *testing.T) {
	db, mock, _ := New()
	query := `
		SELECT
			name,
			email,
			address,
			anotherfield
		FROM user
		where
			name    = 'John'
			and
			address = 'Jakarta'

	`

	mock.ExpectQuery(query)
	mock.ExpectExec(query)
	mock.ExpectPrepare(query)

	db.QueryRow(query)
	db.Exec(query)
	db.Prepare(query)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
