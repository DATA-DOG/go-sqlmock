package main

import (
	"database/sql"
	"fmt"
)

// this type emulate go-mssqldb's ReturnStatus type, used to get rc from SQL Server stored procedures
// https://github.com/microsoft/go-mssqldb/blob/main/mssql.go
type ReturnStatus int32

func execWithNamedOutputArgs(db *sql.DB, outputArg *string, inputOutputArg *string) (err error) {
	_, err = db.Exec("EXEC spWithNamedOutputParameters",
		sql.Named("outArg", sql.Out{Dest: outputArg}),
		sql.Named("inoutArg", sql.Out{In: true, Dest: inputOutputArg}),
	)
	if err != nil {
		return
	}
	return
}

func execWithTypedOutputArgs(db *sql.DB, rcArg *ReturnStatus) (err error) {
	if _, err = db.Exec("EXEC spWithReturnCode", rcArg); err != nil {
		return
	}
	return
}

func main() {
	// @NOTE: the real connection is not required for tests
	db, err := sql.Open("mssql", "myconnectionstring")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	outputArg := ""
	inputOutputArg := "abcInput"

	if err = execWithNamedOutputArgs(db, &outputArg, &inputOutputArg); err != nil {
		panic(err)
	}

	rcArg := new(ReturnStatus)
	if err = execWithTypedOutputArgs(db, rcArg); err != nil {
		panic(err)
	}

	if _, err = fmt.Printf("outputArg: %s, inputOutputArg: %s, rcArg: %d", outputArg, inputOutputArg, *rcArg); err != nil {
		panic(err)
	}
}
