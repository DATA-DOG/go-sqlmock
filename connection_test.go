package sqlmock

import (
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"
)

func TestExecNoExpectations(t *testing.T) {
	c := &conn{
		expectations: []expectation{
			&expectedExec{
				queryBasedExpectation: queryBasedExpectation{
					commonExpectation: commonExpectation{
						triggered: true,
						err:       errors.New("WillReturnError"),
					},
					sqlRegex: regexp.MustCompile(regexp.QuoteMeta("otherquery")),
					args:     []driver.Value{456},
				},
			},
		},
	}
	res, err := c.Exec("query", []driver.Value{123})
	if res != nil {
		t.Error("Result should be nil")
	}
	if err == nil {
		t.Error("error should not be nil")
	}
	pattern := regexp.MustCompile(regexp.QuoteMeta("all expectations were already fulfilled, call to exec"))
	if !pattern.MatchString(err.Error()) {
		t.Errorf("error should match expected error message (actual: %s)", err.Error())
	}
}

func TestExecExpectationMismatch(t *testing.T) {
	c := &conn{
		expectations: []expectation{
			&expectedQuery{
				queryBasedExpectation: queryBasedExpectation{
					commonExpectation: commonExpectation{
						err: errors.New("WillReturnError"),
					},
					sqlRegex: regexp.MustCompile(regexp.QuoteMeta("otherquery")),
					args:     []driver.Value{456},
				},
			},
		},
	}
	res, err := c.Exec("query", []driver.Value{123})
	if res != nil {
		t.Error("Result should be nil")
	}
	if err == nil {
		t.Error("error should not be nil")
	}
	pattern := regexp.MustCompile(regexp.QuoteMeta("was not expected, next expectation is"))
	if !pattern.MatchString(err.Error()) {
		t.Errorf("error should match expected error message (actual: %s)", err.Error())
	}
}

func TestExecQueryMismatch(t *testing.T) {
	c := &conn{
		expectations: []expectation{
			&expectedExec{
				queryBasedExpectation: queryBasedExpectation{
					commonExpectation: commonExpectation{
						err: errors.New("WillReturnError"),
					},
					sqlRegex: regexp.MustCompile(regexp.QuoteMeta("otherquery")),
					args:     []driver.Value{456},
				},
			},
		},
	}
	res, err := c.Exec("query", []driver.Value{123})
	if res != nil {
		t.Error("Result should be nil")
	}
	if err == nil {
		t.Error("error should not be nil")
	}
	pattern := regexp.MustCompile(regexp.QuoteMeta("does not match regex"))
	if !pattern.MatchString(err.Error()) {
		t.Errorf("error should match expected error message (actual: %s)", err.Error())
	}
}

func TestExecArgsMismatch(t *testing.T) {
	c := &conn{
		expectations: []expectation{
			&expectedExec{
				queryBasedExpectation: queryBasedExpectation{
					commonExpectation: commonExpectation{
						err: errors.New("WillReturnError"),
					},
					sqlRegex: regexp.MustCompile(regexp.QuoteMeta("query")),
					args:     []driver.Value{456},
				},
			},
		},
	}
	res, err := c.Exec("query", []driver.Value{123})
	if res != nil {
		t.Error("Result should be nil")
	}
	if err == nil {
		t.Error("error should not be nil")
	}
	pattern := regexp.MustCompile(regexp.QuoteMeta("does not match expected"))
	if !pattern.MatchString(err.Error()) {
		t.Errorf("error should match expected error message (actual: %s)", err.Error())
	}
}

func TestExecWillReturnError(t *testing.T) {
	c := &conn{
		expectations: []expectation{
			&expectedExec{
				queryBasedExpectation: queryBasedExpectation{
					commonExpectation: commonExpectation{
						err: errors.New("WillReturnError"),
					},
					sqlRegex: regexp.MustCompile(regexp.QuoteMeta("query")),
				},
			},
		},
	}
	res, err := c.Exec("query", []driver.Value{123})
	if res != nil {
		t.Error("Result should be nil")
	}
	if err == nil {
		t.Error("error should not be nil")
	}
	if err.Error() != "WillReturnError" {
		t.Errorf("error should match expected error message (actual: %s)", err.Error())
	}
}

func TestExecMissingResult(t *testing.T) {
	c := &conn{
		expectations: []expectation{
			&expectedExec{
				queryBasedExpectation: queryBasedExpectation{
					commonExpectation: commonExpectation{},
					sqlRegex:          regexp.MustCompile(regexp.QuoteMeta("query")),
					args:              []driver.Value{123},
				},
			},
		},
	}
	res, err := c.Exec("query", []driver.Value{123})
	if res != nil {
		t.Error("Result should be nil")
	}
	if err == nil {
		t.Error("error should not be nil")
	}
	pattern := regexp.MustCompile(regexp.QuoteMeta("must return a database/sql/driver.result, but it was not set for expectation"))
	if !pattern.MatchString(err.Error()) {
		t.Errorf("error should match expected error message (actual: %s)", err.Error())
	}
}

func TestExec(t *testing.T) {
	expectedResult := driver.Result(&result{})
	c := &conn{
		expectations: []expectation{
			&expectedExec{
				queryBasedExpectation: queryBasedExpectation{
					commonExpectation: commonExpectation{},
					sqlRegex:          regexp.MustCompile(regexp.QuoteMeta("query")),
					args:              []driver.Value{123},
				},
				result: expectedResult,
			},
		},
	}
	res, err := c.Exec("query", []driver.Value{123})
	if res == nil {
		t.Error("Result should not be nil")
	}
	if res != expectedResult {
		t.Errorf("Result should match expected Result (actual %+v)", res)
	}
	if err != nil {
		t.Errorf("error should be nil (actual %s)", err.Error())
	}
}

func TestQueryNoExpectations(t *testing.T) {
	c := &conn{
		expectations: []expectation{
			&expectedQuery{
				queryBasedExpectation: queryBasedExpectation{
					commonExpectation: commonExpectation{
						triggered: true,
						err:       errors.New("WillReturnError"),
					},
					sqlRegex: regexp.MustCompile(regexp.QuoteMeta("otherquery")),
					args:     []driver.Value{456},
				},
			},
		},
	}
	res, err := c.Query("query", []driver.Value{123})
	if res != nil {
		t.Error("Rows should be nil")
	}
	if err == nil {
		t.Error("error should not be nil")
	}
	pattern := regexp.MustCompile(regexp.QuoteMeta("all expectations were already fulfilled, call to query"))
	if !pattern.MatchString(err.Error()) {
		t.Errorf("error should match expected error message (actual: %s)", err.Error())
	}
}

func TestQueryExpectationMismatch(t *testing.T) {
	c := &conn{
		expectations: []expectation{
			&expectedExec{
				queryBasedExpectation: queryBasedExpectation{
					commonExpectation: commonExpectation{
						err: errors.New("WillReturnError"),
					},
					sqlRegex: regexp.MustCompile(regexp.QuoteMeta("otherquery")),
					args:     []driver.Value{456},
				},
			},
		},
	}
	res, err := c.Query("query", []driver.Value{123})
	if res != nil {
		t.Error("Rows should be nil")
	}
	if err == nil {
		t.Error("error should not be nil")
	}
	pattern := regexp.MustCompile(regexp.QuoteMeta("was not expected, next expectation is"))
	if !pattern.MatchString(err.Error()) {
		t.Errorf("error should match expected error message (actual: %s)", err.Error())
	}
}

func TestQueryQueryMismatch(t *testing.T) {
	c := &conn{
		expectations: []expectation{
			&expectedQuery{
				queryBasedExpectation: queryBasedExpectation{
					commonExpectation: commonExpectation{
						err: errors.New("WillReturnError"),
					},
					sqlRegex: regexp.MustCompile(regexp.QuoteMeta("otherquery")),
					args:     []driver.Value{456},
				},
			},
		},
	}
	res, err := c.Query("query", []driver.Value{123})
	if res != nil {
		t.Error("Rows should be nil")
	}
	if err == nil {
		t.Error("error should not be nil")
	}
	pattern := regexp.MustCompile(regexp.QuoteMeta("does not match regex"))
	if !pattern.MatchString(err.Error()) {
		t.Errorf("error should match expected error message (actual: %s)", err.Error())
	}
}

func TestQueryArgsMismatch(t *testing.T) {
	c := &conn{
		expectations: []expectation{
			&expectedQuery{
				queryBasedExpectation: queryBasedExpectation{
					commonExpectation: commonExpectation{
						err: errors.New("WillReturnError"),
					},
					sqlRegex: regexp.MustCompile(regexp.QuoteMeta("query")),
					args:     []driver.Value{456},
				},
			},
		},
	}
	res, err := c.Query("query", []driver.Value{123})
	if res != nil {
		t.Error("Rows should be nil")
	}
	if err == nil {
		t.Error("error should not be nil")
	}
	pattern := regexp.MustCompile(regexp.QuoteMeta("does not match expected"))
	if !pattern.MatchString(err.Error()) {
		t.Errorf("error should match expected error message (actual: %s)", err.Error())
	}
}

func TestQueryWillReturnError(t *testing.T) {
	c := &conn{
		expectations: []expectation{
			&expectedQuery{
				queryBasedExpectation: queryBasedExpectation{
					commonExpectation: commonExpectation{
						err: errors.New("WillReturnError"),
					},
					sqlRegex: regexp.MustCompile(regexp.QuoteMeta("query")),
				},
			},
		},
	}
	res, err := c.Query("query", []driver.Value{123})
	if res != nil {
		t.Error("Rows should be nil")
	}
	if err == nil {
		t.Error("error should not be nil")
	}
	if err.Error() != "WillReturnError" {
		t.Errorf("error should match expected error message (actual: %s)", err.Error())
	}
}

func TestQueryMissingRows(t *testing.T) {
	c := &conn{
		expectations: []expectation{
			&expectedQuery{
				queryBasedExpectation: queryBasedExpectation{
					commonExpectation: commonExpectation{},
					sqlRegex:          regexp.MustCompile(regexp.QuoteMeta("query")),
					args:              []driver.Value{123},
				},
			},
		},
	}
	res, err := c.Query("query", []driver.Value{123})
	if res != nil {
		t.Error("Rows should be nil")
	}
	if err == nil {
		t.Error("error should not be nil")
	}
	pattern := regexp.MustCompile(regexp.QuoteMeta("must return a database/sql/driver.rows, but it was not set for expectation"))
	if !pattern.MatchString(err.Error()) {
		t.Errorf("error should match expected error message (actual: %s)", err.Error())
	}
}

func TestQuery(t *testing.T) {
	expectedRows := driver.Rows(&rows{})
	c := &conn{
		expectations: []expectation{
			&expectedQuery{
				queryBasedExpectation: queryBasedExpectation{
					commonExpectation: commonExpectation{},
					sqlRegex:          regexp.MustCompile(regexp.QuoteMeta("query")),
					args:              []driver.Value{123},
				},
				rows: expectedRows,
			},
		},
	}
	rows, err := c.Query("query", []driver.Value{123})
	if rows == nil {
		t.Error("Rows should not be nil")
	}
	if rows != expectedRows {
		t.Errorf("Rows should match expected Rows (actual %+v)", rows)
	}
	if err != nil {
		t.Errorf("error should be nil (actual %s)", err.Error())
	}
}
