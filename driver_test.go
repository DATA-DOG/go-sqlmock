package sqlmock

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
)

type void struct{}

func (void) Print(...interface{}) {}

type converter struct{}

func (c *converter) ConvertValue(v interface{}) (driver.Value, error) {
	return nil, errors.New("converter disabled")
}

func ExampleNew() {
	db, mock, err := New()
	if err != nil {
		fmt.Println("expected no error, but got:", err)
		return
	}
	defer db.Close()
	// now we can expect operations performed on db
	mock.ExpectBegin().WillReturnError(fmt.Errorf("an error will occur on db.Begin() call"))
}

func TestShouldOpenConnectionIssue15(t *testing.T) {
	db, mock, err := New()
	if err != nil {
		t.Errorf("expected no error, but got: %s", err)
	}
	if len(pool.conns) != 1 {
		t.Errorf("expected 1 connection in pool, but there is: %d", len(pool.conns))
	}

	smock, _ := mock.(*sqlmock)
	if smock.opened != 1 {
		t.Errorf("expected 1 connection on mock to be opened, but there is: %d", smock.opened)
	}

	// defer so the rows gets closed first
	defer func() {
		if smock.opened != 0 {
			t.Errorf("expected no connections on mock to be opened, but there is: %d", smock.opened)
		}
	}()

	mock.ExpectQuery("SELECT").WillReturnRows(NewRows([]string{"one", "two"}).AddRow("val1", "val2"))
	rows, err := db.Query("SELECT")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	defer rows.Close()

	mock.ExpectExec("UPDATE").WillReturnResult(NewResult(1, 1))
	if _, err = db.Exec("UPDATE"); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// now there should be two connections open
	if smock.opened != 2 {
		t.Errorf("expected 2 connection on mock to be opened, but there is: %d", smock.opened)
	}

	mock.ExpectClose()
	if err = db.Close(); err != nil {
		t.Errorf("expected no error on close, but got: %s", err)
	}

	// one is still reserved for rows
	if smock.opened != 1 {
		t.Errorf("expected 1 connection on mock to be still reserved for rows, but there is: %d", smock.opened)
	}
}

func TestTwoOpenConnectionsOnTheSameDSN(t *testing.T) {
	db, mock, err := New()
	if err != nil {
		t.Errorf("expected no error, but got: %s", err)
	}
	db2, mock2, err := New()
	if err != nil {
		t.Errorf("expected no error, but got: %s", err)
	}
	if len(pool.conns) != 2 {
		t.Errorf("expected 2 connection in pool, but there is: %d", len(pool.conns))
	}

	if db == db2 {
		t.Errorf("expected not the same database instance, but it is the same")
	}
	if mock == mock2 {
		t.Errorf("expected not the same mock instance, but it is the same")
	}
}

func TestWithOptions(t *testing.T) {
	c := &converter{}
	_, mock, err := New(ValueConverterOption(c))
	if err != nil {
		t.Errorf("expected no error, but got: %s", err)
	}
	smock, _ := mock.(*sqlmock)
	if smock.converter.(*converter) != c {
		t.Errorf("expected a custom converter to be set")
	}
}

func TestWrongDSN(t *testing.T) {
	t.Parallel()
	db, _, _ := New()
	defer db.Close()
	if _, err := db.Driver().Open("wrong_dsn"); err == nil {
		t.Error("expected error on Open")
	}
}

func TestNewDSN(t *testing.T) {
	if _, _, err := NewWithDSN("sqlmock_db_99"); err != nil {
		t.Errorf("expected no error on NewWithDSN, but got: %s", err)
	}
}

func TestDuplicateNewDSN(t *testing.T) {
	if _, _, err := NewWithDSN("sqlmock_db_1"); err == nil {
		t.Error("expected error on NewWithDSN")
	}
}
