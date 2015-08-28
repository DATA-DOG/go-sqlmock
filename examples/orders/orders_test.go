package main

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// will test that order with a different status, cannot be cancelled
func TestShouldNotCancelOrderWithNonPendingStatus(t *testing.T) {
	// open database stub
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// columns are prefixed with "o" since we used sqlstruct to generate them
	columns := []string{"o_id", "o_status"}
	// expect transaction begin
	mock.ExpectBegin()
	// expect query to fetch order and user, match it with regexp
	mock.ExpectQuery("SELECT (.+) FROM orders AS o INNER JOIN users AS u (.+) FOR UPDATE").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).FromCSVString("1,1"))
	// expect transaction rollback, since order status is "cancelled"
	mock.ExpectRollback()

	// run the cancel order function
	err = cancelOrder(1, db)
	if err != nil {
		t.Errorf("Expected no error, but got %s instead", err)
	}
	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

// will test order cancellation
func TestShouldRefundUserWhenOrderIsCancelled(t *testing.T) {
	// open database stub
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// columns are prefixed with "o" since we used sqlstruct to generate them
	columns := []string{"o_id", "o_status", "o_value", "o_reserved_fee", "u_id", "u_balance"}
	// expect transaction begin
	mock.ExpectBegin()
	// expect query to fetch order and user, match it with regexp
	mock.ExpectQuery("SELECT (.+) FROM orders AS o INNER JOIN users AS u (.+) FOR UPDATE").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(1, 0, 25.75, 3.25, 2, 10.00))
	// expect user balance update
	mock.ExpectPrepare("UPDATE users SET balance").ExpectExec().
		WithArgs(25.75+3.25, 2).                  // refund amount, user id
		WillReturnResult(sqlmock.NewResult(0, 1)) // no insert id, 1 affected row
	// expect order status update
	mock.ExpectPrepare("UPDATE orders SET status").ExpectExec().
		WithArgs(ORDER_CANCELLED, 1).             // status, id
		WillReturnResult(sqlmock.NewResult(0, 1)) // no insert id, 1 affected row
	// expect a transaction commit
	mock.ExpectCommit()

	// run the cancel order function
	err = cancelOrder(1, db)
	if err != nil {
		t.Errorf("Expected no error, but got %s instead", err)
	}
	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

// will test order cancellation
func TestShouldRollbackOnError(t *testing.T) {
	// open database stub
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// expect transaction begin
	mock.ExpectBegin()
	// expect query to fetch order and user, match it with regexp
	mock.ExpectQuery("SELECT (.+) FROM orders AS o INNER JOIN users AS u (.+) FOR UPDATE").
		WithArgs(1).
		WillReturnError(fmt.Errorf("Some error"))
	// should rollback since error was returned from query execution
	mock.ExpectRollback()

	// run the cancel order function
	err = cancelOrder(1, db)
	// error should return back
	if err == nil {
		t.Error("Expected error, but got none")
	}
	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}
