[![Build Status](https://travis-ci.org/DATA-DOG/go-sqlmock.png)](https://travis-ci.org/DATA-DOG/go-sqlmock)
[![GoDoc](https://godoc.org/github.com/DATA-DOG/go-sqlmock?status.png)](https://godoc.org/github.com/DATA-DOG/go-sqlmock)

# Sql driver mock for Golang

This is a **mock** driver as **database/sql/driver** which is very flexible and pragmatic to
manage and mock expected queries. All the expectations should be met and all queries and actions
triggered should be mocked in order to pass a test.

## Install

    go get github.com/DATA-DOG/go-sqlmock

## Use it with pleasure

An example of some database interaction which you may want to test:

``` go
package main

import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "github.com/kisielk/sqlstruct"
    "fmt"
    "log"
)

const ORDER_PENDING = 0
const ORDER_CANCELLED = 1

type User struct {
    Id int `sql:"id"`
    Username string `sql:"username"`
    Balance float64 `sql:"balance"`
}

type Order struct {
    Id int `sql:"id"`
    Value float64 `sql:"value"`
    ReservedFee float64 `sql:"reserved_fee"`
    Status int `sql:"status"`
}

func cancelOrder(id int, db *sql.DB) (err error) {
    tx, err := db.Begin()
    if err != nil {
        return
    }

    var order Order
    var user User
    sql := fmt.Sprintf(`
SELECT %s, %s
FROM orders AS o
INNER JOIN users AS u ON o.buyer_id = u.id
WHERE o.id = ?
FOR UPDATE`,
    sqlstruct.ColumnsAliased(order, "o"),
    sqlstruct.ColumnsAliased(user, "u"))

    // fetch order to cancel
    rows, err := tx.Query(sql, id)
    if err != nil {
        tx.Rollback()
        return
    }

    defer rows.Close()
    // no rows, nothing to do
    if !rows.Next() {
        tx.Rollback()
        return
    }

    // read order
    err = sqlstruct.ScanAliased(&order, rows, "o")
    if err != nil {
        tx.Rollback()
        return
    }

    // ensure order status
    if order.Status != ORDER_PENDING {
        tx.Rollback()
        return
    }

    // read user
    err = sqlstruct.ScanAliased(&user, rows, "u")
    if err != nil {
        tx.Rollback()
        return
    }
    rows.Close() // manually close before other prepared statements

    // refund order value
    sql = "UPDATE users SET balance = balance + ? WHERE id = ?"
    refundStmt, err := tx.Prepare(sql)
    if err != nil {
        tx.Rollback()
        return
    }
    defer refundStmt.Close()
    _, err = refundStmt.Exec(order.Value + order.ReservedFee, user.Id)
    if err != nil {
        tx.Rollback()
        return
    }

    // update order status
    order.Status = ORDER_CANCELLED
    sql = "UPDATE orders SET status = ?, updated = NOW() WHERE id = ?"
    orderUpdStmt, err := tx.Prepare(sql)
    if err != nil {
        tx.Rollback()
        return
    }
    defer orderUpdStmt.Close()
    _, err = orderUpdStmt.Exec(order.Status, order.Id)
    if err != nil {
        tx.Rollback()
        return
    }
    return tx.Commit()
}

func main() {
    db, err := sql.Open("mysql", "root:nimda@/test")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    err = cancelOrder(1, db)
    if err != nil {
        log.Fatal(err)
    }
}
```

And the clean nice test:

``` go
package main

import (
    "database/sql"
    "github.com/DATA-DOG/go-sqlmock"
    "testing"
    "fmt"
)

// will test that order with a different status, cannot be cancelled
func TestShouldNotCancelOrderWithNonPendingStatus(t *testing.T) {
    // open database stub
    db, err := sqlmock.New()
    if err != nil {
        t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
    }

    // columns are prefixed with "o" since we used sqlstruct to generate them
    columns := []string{"o_id", "o_status"}
    // expect transaction begin
    sqlmock.ExpectBegin()
    // expect query to fetch order and user, match it with regexp
    sqlmock.ExpectQuery("SELECT (.+) FROM orders AS o INNER JOIN users AS u (.+) FOR UPDATE").
        WithArgs(1).
        WillReturnRows(sqlmock.NewRows(columns).FromCSVString("1,1"))
    // expect transaction rollback, since order status is "cancelled"
    sqlmock.ExpectRollback()

    // run the cancel order function
    err = cancelOrder(1, db)
    if err != nil {
        t.Errorf("Expected no error, but got %s instead", err)
    }
    // db.Close() ensures that all expectations have been met
    if err = db.Close(); err != nil {
        t.Errorf("Error '%s' was not expected while closing the database", err)
    }
}

// will test order cancellation
func TestShouldRefundUserWhenOrderIsCancelled(t *testing.T) {
    // open database stub
    db, err := sqlmock.New()
    if err != nil {
        t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
    }

    // columns are prefixed with "o" since we used sqlstruct to generate them
    columns := []string{"o_id", "o_status", "o_value", "o_reserved_fee", "u_id", "u_balance"}
    // expect transaction begin
    sqlmock.ExpectBegin()
    // expect query to fetch order and user, match it with regexp
    sqlmock.ExpectQuery("SELECT (.+) FROM orders AS o INNER JOIN users AS u (.+) FOR UPDATE").
        WithArgs(1).
        WillReturnRows(sqlmock.NewRows(columns).AddRow(1, 0, 25.75, 3.25, 2, 10.00))
    // expect user balance update
    sqlmock.ExpectExec("UPDATE users SET balance").
        WithArgs(25.75 + 3.25, 2). // refund amount, user id
        WillReturnResult(sqlmock.NewResult(0, 1)) // no insert id, 1 affected row
    // expect order status update
    sqlmock.ExpectExec("UPDATE orders SET status").
        WithArgs(ORDER_CANCELLED, 1). // status, id
        WillReturnResult(sqlmock.NewResult(0, 1)) // no insert id, 1 affected row
    // expect a transaction commit
    sqlmock.ExpectCommit()

    // run the cancel order function
    err = cancelOrder(1, db)
    if err != nil {
        t.Errorf("Expected no error, but got %s instead", err)
    }
    // db.Close() ensures that all expectations have been met
    if err = db.Close(); err != nil {
        t.Errorf("Error '%s' was not expected while closing the database", err)
    }
}

// will test order cancellation
func TestShouldRollbackOnError(t *testing.T) {
    // open database stub
    db, err := sqlmock.New()
    if err != nil {
        t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
    }

    // expect transaction begin
    sqlmock.ExpectBegin()
    // expect query to fetch order and user, match it with regexp
    sqlmock.ExpectQuery("SELECT (.+) FROM orders AS o INNER JOIN users AS u (.+) FOR UPDATE").
        WithArgs(1).
        WillReturnError(fmt.Errorf("Some error"))
    // should rollback since error was returned from query execution
    sqlmock.ExpectRollback()

    // run the cancel order function
    err = cancelOrder(1, db)
    // error should return back
    if err == nil {
        t.Error("Expected error, but got none")
    }
    // db.Close() ensures that all expectations have been met
    if err = db.Close(); err != nil {
        t.Errorf("Error '%s' was not expected while closing the database", err)
    }
}
```

## Expectations

All **Expect** methods return a **Mock** interface which allow you to describe
expectations in more details: return an error, expect specific arguments, return rows and so on.
**NOTE:** that if you call **WithArgs** on a non query based expectation, it will panic

A **Mock** interface:

``` go
type Mock interface {
	WithArgs(...driver.Value) Mock
	WillReturnError(error) Mock
	WillReturnRows(driver.Rows) Mock
	WillReturnResult(driver.Result) Mock
}
```

As an example we can expect a transaction commit and simulate an error for it:

``` go
sqlmock.ExpectCommit().WillReturnError(fmt.Errorf("Deadlock occured"))
```

In same fashion, we can expect queries to match arguments. If there are any, it must be matched.
Instead of result we can return error..

``` go
sqlmock.ExpectQuery("SELECT (.*) FROM orders").
	WithArgs("string value").
	WillReturnRows(sqlmock.NewRows([]string{"col"}).AddRow("val"))
```

**NOTE:** it matches a regular expression. Some regex special characters must be escaped if you want to match them.
For example if we want to match a subselect:

``` go
sqlmock.ExpectQuery("SELECT (.*) FROM orders WHERE id IN \\(SELECT id FROM finished WHERE status = 1\\)").
	WithArgs("string value").
	WillReturnRows(sqlmock.NewRows([]string{"col"}).AddRow("val"))
```

**WithArgs** expectation, compares values based on their type, for usual values like **string, float, int**
it matches the actual value. Types like **time** are compared only by type. Other types might require different ways
to compare them correctly, this may be improved.

You can build rows either from CSV string or from interface values:

**Rows** interface, which satisfies sql driver.Rows:

``` go
type Rows interface {
	AddRow(...driver.Value) Rows
	FromCSVString(s string) Rows
	Next([]driver.Value) error
	Columns() []string
	Close() error
}
```

Example for to build rows:

``` go
rs := sqlmock.NewRows([]string{"column1", "column2"}).
	FromCSVString("one,1\ntwo,2").
	AddRow("three", 3)
```

**Prepare** will ignore other expectations if ExpectPrepare not set. When set, can expect normal result or simulate an error:

``` go
rs := sqlmock.ExpectPrepare().
    WillReturnError(fmt.Errorf("Query prepare failed"))
```

## Run tests

    go test

## Documentation

Visit [godoc](http://godoc.org/github.com/DATA-DOG/go-sqlmock)
See **.travis.yml** for supported **go** versions

## Changes

- **2014-08-16** instead of **panic** during reflect type mismatch when comparing query arguments - now return error
- **2014-08-14** added **sqlmock.NewErrorResult** which gives an option to return driver.Result with errors for
interface methods, see [issue](https://github.com/DATA-DOG/go-sqlmock/issues/5)
- **2014-05-29** allow to match arguments in more sophisticated ways, by providing an **sqlmock.Argument** interface
- **2014-04-21** introduce **sqlmock.New()** to open a mock database connection for tests. This method
calls sql.DB.Ping to ensure that connection is open, see [issue](https://github.com/DATA-DOG/go-sqlmock/issues/4).
This way on Close it will surely assert if all expectations are met, even if database was not triggered at all.
The old way is still available, but it is advisable to call db.Ping manually before asserting with db.Close.
- **2014-02-14** RowsFromCSVString is now a part of Rows interface named as FromCSVString.
It has changed to allow more ways to construct rows and to easily extend this API in future.
See [issue 1](https://github.com/DATA-DOG/go-sqlmock/issues/1)
**RowsFromCSVString** is deprecated and will be removed in future

## Contributions

Feel free to open a pull request. Note, if you wish to contribute an extension to public (exported methods or types) -
please open an issue before, to discuss whether these changes can be accepted. All backward incompatible changes are
and will be treated cautiously

## License

The [three clause BSD license](http://en.wikipedia.org/wiki/BSD_licenses)

