package sqlmock

import (
	"database/sql"
	"strings"
	"testing"
)

func TestDatabaseSQLReturnsCorrectInstances(t *testing.T) {
	mockOne, err := NewMockConn("one")
	if err != nil {
		t.Errorf("got error on init conn: %v", err)
		return
	}
	mockTwo, err := NewMockConn("two")
	if err != nil {
		t.Errorf("got error on init conn: %v", err)
		return
	}
	_, err = NewMockConn("one")
	if err == nil {
		t.Error("expect error on requesting same id, got nil instead")
		return
	}
	if !strings.Contains(err.Error(), "already a connection") {
		t.Errorf("expect error message hinting at already existing connection. got %s", err.Error())
		return
	}

	mockOne.ExpectQuery("SELECT one")
	mockTwo.ExpectQuery("SELECT two")

	dbOne, err := sql.Open("mock", "id=one")
	if err != nil {
		t.Errorf("expect db one to be returned, got error instead: %v", err)
		return
	}
	err = dbOne.Ping()
	if err != nil {
		t.Errorf("error on ping db one: %v", err)
		return
	}
	dbTwo, err := sql.Open("mock", "id=two")
	if err != nil {
		t.Errorf("expect db two to be returned, got error instead: %v", err)
		return
	}
	err = dbTwo.Ping()
	if err != nil {
		t.Errorf("error on ping db two: %v", err)
	}
	d, err := sql.Open("mock", "id=three")
	if err != nil {
		t.Errorf("error on open nonexistent mock: %v", err)
	}
	err = d.Ping()
	if err == nil {
		t.Errorf("expect nonexistent mock request to return error, got nil instead with %+v.", d)
		return
	}
	if !strings.Contains(err.Error(), "no connection with ID") {
		t.Errorf("expect error message hinting at wrong id. got message: %s", err.Error())
		return
	}

	for id, db := range map[string]*sql.DB{"one": dbOne, "two": dbTwo} {
		go func(id string, db *sql.DB) {
			_, err := db.Query("SELECT " + id)
			if err != nil {
				t.Errorf("error on query: %v", err)
			}
			err = db.Close()
			if err != nil {
				t.Errorf("error on db close (id %s): %v", id, err)
			}
		}(id, db)
	}
}
