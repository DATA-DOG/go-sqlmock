package sqlmock

import (
	"fmt"
	"testing"
)

func ExampleQueryMatcher() {
	// configure to use case sensitive SQL query matcher
	// instead of default regular expression matcher
	db, mock, err := New(QueryMatcherOption(QueryMatcherEqual))
	if err != nil {
		fmt.Println("failed to open sqlmock database:", err)
	}
	defer db.Close()

	rows := NewRows([]string{"id", "title"}).
		AddRow(1, "one").
		AddRow(2, "two")

	mock.ExpectQuery("SELECT * FROM users").WillReturnRows(rows)

	rs, err := db.Query("SELECT * FROM users")
	if err != nil {
		fmt.Println("failed to match expected query")
		return
	}
	defer rs.Close()

	for rs.Next() {
		var id int
		var title string
		rs.Scan(&id, &title)
		fmt.Println("scanned id:", id, "and title:", title)
	}

	if rs.Err() != nil {
		fmt.Println("got rows error:", rs.Err())
	}
	// Output: scanned id: 1 and title: one
	// scanned id: 2 and title: two
}

func TestQueryStringStripping(t *testing.T) {
	assert := func(actual, expected string) {
		if res := stripQuery(actual); res != expected {
			t.Errorf("Expected '%s' to be '%s', but got '%s'", actual, expected, res)
		}
	}

	assert(" SELECT 1", "SELECT 1")
	assert("SELECT   1 FROM   d", "SELECT 1 FROM d")
	assert(`
    SELECT c
    FROM D
`, "SELECT c FROM D")
	assert("UPDATE  (.+) SET  ", "UPDATE (.+) SET")
}

func TestQueryMatcherRegexp(t *testing.T) {
	type testCase struct {
		expected string
		actual   string
		err      error
	}

	cases := []testCase{
		{"?\\l", "SEL", fmt.Errorf("error parsing regexp: missing argument to repetition operator: `?`")},
		{"SELECT (.+) FROM users", "SELECT name, email FROM users WHERE id = ?", nil},
		{"Select (.+) FROM users", "SELECT name, email FROM users WHERE id = ?", fmt.Errorf(`could not match actual sql: "SELECT name, email FROM users WHERE id = ?" with expected regexp "Select (.+) FROM users"`)},
		{"SELECT (.+) FROM\nusers", "SELECT name, email\n FROM users\n WHERE id = ?", nil},
	}

	for i, c := range cases {
		err := QueryMatcherRegexp.Match(c.expected, c.actual)
		if err == nil && c.err != nil {
			t.Errorf(`got no error, but expected "%v" at %d case`, c.err, i)
			continue
		}
		if err != nil && c.err == nil {
			t.Errorf(`got unexpected error "%v" at %d case`, err, i)
			continue
		}
		if err == nil {
			continue
		}
		if err.Error() != c.err.Error() {
			t.Errorf(`expected error "%v", but got "%v" at %d case`, c.err, err, i)
		}
	}
}

func TestQueryMatcherEqual(t *testing.T) {
	type testCase struct {
		expected string
		actual   string
		err      error
	}

	cases := []testCase{
		{"SELECT name, email FROM users WHERE id = ?", "SELECT name, email\n FROM users\n WHERE id = ?", nil},
		{"SELECT", "Select", fmt.Errorf(`actual sql: "Select" does not equal to expected "SELECT"`)},
		{"SELECT from users", "SELECT from table", fmt.Errorf(`actual sql: "SELECT from table" does not equal to expected "SELECT from users"`)},
	}

	for i, c := range cases {
		err := QueryMatcherEqual.Match(c.expected, c.actual)
		if err == nil && c.err != nil {
			t.Errorf(`got no error, but expected "%v" at %d case`, c.err, i)
			continue
		}
		if err != nil && c.err == nil {
			t.Errorf(`got unexpected error "%v" at %d case`, err, i)
			continue
		}
		if err == nil {
			continue
		}
		if err.Error() != c.err.Error() {
			t.Errorf(`expected error "%v", but got "%v" at %d case`, c.err, err, i)
		}
	}
}
