package sqlmock

import (
	"errors"
	"testing"
)

// +build go1.6

func TestExpectedPreparedStatemtCloseError(t *testing.T) {
	conn, mock, err := New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database:", err)
	}

	mock.ExpectBegin()
	want := errors.New("STMT ERROR")
	mock.ExpectPrepare("SELECT").WillReturnCloseError(want)

	txn, err := conn.Begin()
	if err != nil {
		t.Fatalf("unexpected error while opening transaction:", err)
	}

	stmt, err := txn.Prepare("SELECT")
	if err != nil {
		t.Fatalf("unexpected error while preparing a statement:", err)
	}

	if err := stmt.Close(); err != want {
		t.Fatalf("Got = %v, want = %v", err, want)
	}
}
