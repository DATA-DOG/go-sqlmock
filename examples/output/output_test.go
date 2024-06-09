package main

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNamedOutputArgs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	inOutInputValue := "abcInput"
	mock.ExpectExec("EXEC spWithNamedOutputParameters").
		WithArgs(
			sqlmock.NamedOutputArg("outArg", "123Output"),
			sqlmock.NamedInputOutputArg("inoutArg", &inOutInputValue, "abcOutput"),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// now we execute our method
	outArg := ""
	inoutArg := "abcInput"
	if err = execWithNamedOutputArgs(db, &outArg, &inoutArg); err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	if outArg != "123Output" {
		t.Errorf("unexpected outArg value")
	}

	if inoutArg != "abcOutput" {
		t.Errorf("unexpected inoutArg value")
	}
}

func TestTypedOutputArgs(t *testing.T) {
	rcArg := new(ReturnStatus) // here we will store the return code

	valueConverter := sqlmock.NewPassthroughValueConverter(rcArg) // we need this converter to bypass the default ValueConverter logic that alter original value's type
	db, mock, err := sqlmock.New(sqlmock.ValueConverterOption(valueConverter))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rcFromSp := ReturnStatus(123) // simulate the return code from the stored procedure
	mock.ExpectExec("EXEC spWithReturnCode").
		WithArgs(
			sqlmock.TypedOutputArg(&rcFromSp), // using this func we can provide the expected type and value
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// now we execute our method
	if err = execWithTypedOutputArgs(db, rcArg); err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	if *rcArg != 123 {
		t.Errorf("unexpected rcArg value")
	}
}
