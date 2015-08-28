package sqlmock

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"
	"time"
)

type matcher struct {
}

func (m matcher) Match(driver.Value) bool {
	return true
}

func TestQueryExpectationArgComparison(t *testing.T) {
	e := &queryBasedExpectation{}
	against := []driver.Value{5}
	if !e.argsMatches(against) {
		t.Error("arguments should match, since the no expectation was set")
	}

	e.args = []driver.Value{5, "str"}

	against = []driver.Value{5}
	if e.argsMatches(against) {
		t.Error("arguments should not match, since the size is not the same")
	}

	against = []driver.Value{3, "str"}
	if e.argsMatches(against) {
		t.Error("arguments should not match, since the first argument (int value) is different")
	}

	against = []driver.Value{5, "st"}
	if e.argsMatches(against) {
		t.Error("arguments should not match, since the second argument (string value) is different")
	}

	against = []driver.Value{5, "str"}
	if !e.argsMatches(against) {
		t.Error("arguments should match, but it did not")
	}

	e.args = []driver.Value{5, time.Now()}

	const longForm = "Jan 2, 2006 at 3:04pm (MST)"
	tm, _ := time.Parse(longForm, "Feb 3, 2013 at 7:54pm (PST)")

	against = []driver.Value{5, tm}
	if !e.argsMatches(against) {
		t.Error("arguments should match (time will be compared only by type), but it did not")
	}

	against = []driver.Value{5, matcher{}}
	if !e.argsMatches(against) {
		t.Error("arguments should match, but it did not")
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

func ExampleExpectExec() {
	db, mock, _ := New()
	result := NewErrorResult(fmt.Errorf("some error"))
	mock.ExpectExec("^INSERT (.+)").WillReturnResult(result)
	res, _ := db.Exec("INSERT something")
	_, err := res.LastInsertId()
	fmt.Println(err)
	// Output: some error
}
