// +build go1.9

package sqlmock

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
)

type CustomConverter struct{}

func (s CustomConverter) ConvertValue(v interface{}) (driver.Value, error) {
	switch v.(type) {
	case string:
		return v.(string), nil
	case []string:
		return v.([]string), nil
	case int:
		return v.(int), nil
	default:
		return nil, errors.New(fmt.Sprintf("cannot convert %T with value %v", v, v))
	}
}

func TestCustomValueConverterExec(t *testing.T) {
	db, mock, _ := New(ValueConverterOption(CustomConverter{}))
	expectedQuery := "INSERT INTO tags \\(name,email,age,hobbies\\) VALUES \\(\\?,\\?,\\?,\\?\\)"
	query := "INSERT INTO tags (name,email,age,hobbies) VALUES (?,?,?,?)"
	name := "John"
	email := "j@jj.j"
	age := 12
	hobbies := []string{"soccer", "netflix"}
	mock.ExpectBegin()
	mock.ExpectPrepare(expectedQuery)
	mock.ExpectExec(expectedQuery).WithArgs(name, email, age, hobbies).WillReturnResult(NewResult(1, 1))
	mock.ExpectCommit()

	ctx := context.Background()
	tx, e := db.BeginTx(ctx, nil)
	if e != nil {
		t.Error(e)
		return
	}
	stmt, e := db.PrepareContext(ctx, query)
	if e != nil {
		t.Error(e)
		return
	}
	_, e = stmt.Exec(name, email, age, hobbies)
	if e != nil {
		t.Error(e)
		return
	}
	tx.Commit()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
