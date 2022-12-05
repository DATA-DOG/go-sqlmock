/*
Package sqlmock is a mock library implementing sql driver. Which has one and only
purpose - to simulate any sql driver behavior in tests, without needing a real
database connection. It helps to maintain correct **TDD** workflow.

It does not require any modifications to your source code in order to test
and mock database operations. Supports concurrency and multiple database mocking.

The driver allows to mock any sql driver method behavior.
*/
package sqlmock

import (
	"database/sql/driver"
)

// Common interface serves to create expectations
// for any kind of database action in order to mock
// and test real database behavior.
type Common interface {
	// ExpectClose queues an expectation for this database
	// action to be triggered. the *ExpectedClose allows
	// to mock database response
	ExpectClose() *ExpectedClose

	// ExpectationsWereMet checks whether all queued expectations
	// were met in order. If any of them was not met - an error is returned.
	ExpectationsWereMet() error

	// ExpectPrepare expects Prepare() to be called with expectedSQL query.
	// the *ExpectedPrepare allows to mock database response.
	// Note that you may expect Query() or Exec() on the *ExpectedPrepare
	// statement to prevent repeating expectedSQL
	ExpectPrepare(expectedSQL string) *ExpectedPrepare

	// ExpectBegin expects *sql.DB.Begin to be called.
	// the *ExpectedBegin allows to mock database response
	ExpectBegin() *ExpectedBegin

	// ExpectCommit expects *sql.Tx.Commit to be called.
	// the *ExpectedCommit allows to mock database response
	ExpectCommit() *ExpectedCommit

	// ExpectRollback expects *sql.Tx.Rollback to be called.
	// the *ExpectedRollback allows to mock database response
	ExpectRollback() *ExpectedRollback

	// ExpectPing expected *sql.DB.Ping to be called.
	// the *ExpectedPing allows to mock database response
	//
	// Ping support only exists in the SQL library in Go 1.8 and above.
	// ExpectPing in Go <=1.7 will return an ExpectedPing but not register
	// any expectations.
	//
	// You must enable pings using MonitorPingsOption for this to register
	// any expectations.
	ExpectPing() *ExpectedPing

	// MatchExpectationsInOrder gives an option whether to match all
	// expectations in the order they were set or not.
	//
	// By default it is set to - true. But if you use goroutines
	// to parallelize your query executation, that option may
	// be handy.
	//
	// This option may be turned on anytime during tests. As soon
	// as it is switched to false, expectations will be matched
	// in any order. Or otherwise if switched to true, any unmatched
	// expectations will be expected in order
	MatchExpectationsInOrder(bool)

	// NewRows allows Rows to be created from a
	// sql driver.Value slice or from the CSV string and
	// to be used as sql driver.Rows.
	NewRows(columns []string) *Rows

	ExpectSql(expectedOpt Argument, expectedSQL string) *ExpectedSql
}

type sqlmock struct {
	ordered      bool
	dsn          string
	opened       int
	drv          *mockDriver
	converter    driver.ValueConverter
	queryMatcher QueryMatcher
	monitorPings bool

	expected []expectation
}
