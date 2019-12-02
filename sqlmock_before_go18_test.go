// +build !go1.8

package sqlmock

import (
	"fmt"
	"testing"
	"time"
)

func TestSqlmockExpectPingHasNoEffect(t *testing.T) {
	db, mock, err := New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	e := mock.ExpectPing()

	// Methods on the expectation can be called
	e.WillDelayFor(time.Hour).WillReturnError(fmt.Errorf("an error"))

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expected no error to be returned, but got '%s'", err)
	}
}
