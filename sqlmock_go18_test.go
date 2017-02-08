// +build go1.8

package sqlmock

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestContextExecCancel(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM users").
		WillDelayFor(time.Second).
		WillReturnResult(NewResult(1, 1))

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Millisecond * 10)
		cancel()
	}()

	_, err = db.ExecContext(ctx, "DELETE FROM users")
	if err == nil {
		t.Error("error was expected, but there was none")
	}

	if err != ErrCancelled {
		t.Errorf("was expecting cancel error, but got: %v", err)
	}

	_, err = db.ExecContext(ctx, "DELETE FROM users")
	if err != context.Canceled {
		t.Error("error was expected since context was already done, but there was none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestPreparedStatementContextExecCancel(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectPrepare("DELETE FROM users").
		ExpectExec().
		WillDelayFor(time.Second).
		WillReturnResult(NewResult(1, 1))

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Millisecond * 10)
		cancel()
	}()

	stmt, err := db.Prepare("DELETE FROM users")
	if err != nil {
		t.Errorf("error was not expected, but got: %v", err)
	}

	_, err = stmt.ExecContext(ctx)
	if err == nil {
		t.Error("error was expected, but there was none")
	}

	if err != ErrCancelled {
		t.Errorf("was expecting cancel error, but got: %v", err)
	}

	_, err = stmt.ExecContext(ctx)
	if err != context.Canceled {
		t.Error("error was expected since context was already done, but there was none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestContextExecWithNamedArg(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM users").
		WithArgs(sql.Named("id", 5)).
		WillDelayFor(time.Second).
		WillReturnResult(NewResult(1, 1))

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Millisecond * 10)
		cancel()
	}()

	_, err = db.ExecContext(ctx, "DELETE FROM users WHERE id = :id", sql.Named("id", 5))
	if err == nil {
		t.Error("error was expected, but there was none")
	}

	if err != ErrCancelled {
		t.Errorf("was expecting cancel error, but got: %v", err)
	}

	_, err = db.ExecContext(ctx, "DELETE FROM users WHERE id = :id", sql.Named("id", 5))
	if err != context.Canceled {
		t.Error("error was expected since context was already done, but there was none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestContextExec(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("DELETE FROM users").
		WillReturnResult(NewResult(1, 1))

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Millisecond * 10)
		cancel()
	}()

	res, err := db.ExecContext(ctx, "DELETE FROM users")
	if err != nil {
		t.Errorf("error was not expected, but got: %v", err)
	}

	affected, err := res.RowsAffected()
	if affected != 1 {
		t.Errorf("expected affected rows 1, but got %v", affected)
	}

	if err != nil {
		t.Errorf("error was not expected, but got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestContextQueryCancel(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rs := NewRows([]string{"id", "title"}).AddRow(5, "hello world")

	mock.ExpectQuery("SELECT (.+) FROM articles WHERE id = ?").
		WithArgs(5).
		WillDelayFor(time.Second).
		WillReturnRows(rs)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Millisecond * 10)
		cancel()
	}()

	_, err = db.QueryContext(ctx, "SELECT id, title FROM articles WHERE id = ?", 5)
	if err == nil {
		t.Error("error was expected, but there was none")
	}

	if err != ErrCancelled {
		t.Errorf("was expecting cancel error, but got: %v", err)
	}

	_, err = db.QueryContext(ctx, "SELECT id, title FROM articles WHERE id = ?", 5)
	if err != context.Canceled {
		t.Error("error was expected since context was already done, but there was none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestPreparedStatementContextQueryCancel(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rs := NewRows([]string{"id", "title"}).AddRow(5, "hello world")

	mock.ExpectPrepare("SELECT (.+) FROM articles WHERE id = ?").
		ExpectQuery().
		WithArgs(5).
		WillDelayFor(time.Second).
		WillReturnRows(rs)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Millisecond * 10)
		cancel()
	}()

	stmt, err := db.Prepare("SELECT id, title FROM articles WHERE id = ?")
	if err != nil {
		t.Errorf("error was not expected, but got: %v", err)
	}

	_, err = stmt.QueryContext(ctx, 5)
	if err == nil {
		t.Error("error was expected, but there was none")
	}

	if err != ErrCancelled {
		t.Errorf("was expecting cancel error, but got: %v", err)
	}

	_, err = stmt.QueryContext(ctx, 5)
	if err != context.Canceled {
		t.Error("error was expected since context was already done, but there was none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestContextQuery(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rs := NewRows([]string{"id", "title"}).AddRow(5, "hello world")

	mock.ExpectQuery("SELECT (.+) FROM articles WHERE id =").
		WithArgs(sql.Named("id", 5)).
		WillDelayFor(time.Millisecond * 3).
		WillReturnRows(rs)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Millisecond * 10)
		cancel()
	}()

	rows, err := db.QueryContext(ctx, "SELECT id, title FROM articles WHERE id = :id", sql.Named("id", 5))
	if err != nil {
		t.Errorf("error was not expected, but got: %v", err)
	}

	if !rows.Next() {
		t.Error("expected one row, but there was none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestContextBeginCancel(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin().WillDelayFor(time.Second)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Millisecond * 10)
		cancel()
	}()

	_, err = db.BeginTx(ctx, nil)
	if err == nil {
		t.Error("error was expected, but there was none")
	}

	if err != ErrCancelled {
		t.Errorf("was expecting cancel error, but got: %v", err)
	}

	_, err = db.BeginTx(ctx, nil)
	if err != context.Canceled {
		t.Error("error was expected since context was already done, but there was none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestContextBegin(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin().WillDelayFor(time.Millisecond * 3)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Millisecond * 10)
		cancel()
	}()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Errorf("error was not expected, but got: %v", err)
	}

	if tx == nil {
		t.Error("expected tx, but there was nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestContextPrepareCancel(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectPrepare("SELECT").WillDelayFor(time.Second)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Millisecond * 10)
		cancel()
	}()

	_, err = db.PrepareContext(ctx, "SELECT")
	if err == nil {
		t.Error("error was expected, but there was none")
	}

	if err != ErrCancelled {
		t.Errorf("was expecting cancel error, but got: %v", err)
	}

	_, err = db.PrepareContext(ctx, "SELECT")
	if err != context.Canceled {
		t.Error("error was expected since context was already done, but there was none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestContextPrepare(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectPrepare("SELECT").WillDelayFor(time.Millisecond * 3)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Millisecond * 10)
		cancel()
	}()

	stmt, err := db.PrepareContext(ctx, "SELECT")
	if err != nil {
		t.Errorf("error was not expected, but got: %v", err)
	}

	if stmt == nil {
		t.Error("expected stmt, but there was nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}
