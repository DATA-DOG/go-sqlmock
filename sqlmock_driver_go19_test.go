// +build go1.9

package sqlmock

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
)

func TestStatementTX(t *testing.T) {
	t.Parallel()
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	prep := mock.ExpectPrepare("SELECT")
	mock.ExpectBegin()

	prep.ExpectQuery().WithArgs(1).WillReturnError(errors.New("fast fail"))

	stmt, err := db.Prepare("SELECT title, body FROM articles WHERE id = ?")
	if err != nil {
		t.Fatalf("unexpected error on prepare: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("unexpected error on begin: %v", err)
	}

	// upgrade connection for statement
	txStmt := tx.Stmt(stmt)
	_, err = txStmt.Query(1)
	if err == nil || err.Error() != "fast fail" {
		t.Fatalf("unexpected result: %v", err)
	}
}

func Test_sqlmock_CheckNamedValue(t *testing.T) {
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	tests := []struct {
		name    string
		arg     *driver.NamedValue
		wantErr bool
	}{
		{
			arg:     &driver.NamedValue{Name: "test", Value: "test"},
			wantErr: false,
		},
		{
			arg:     &driver.NamedValue{Name: "test", Value: sql.Out{}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := mock.(*sqlmock).CheckNamedValue(tt.arg); (err != nil) != tt.wantErr {
				t.Errorf("CheckNamedValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
