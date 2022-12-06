package sqlmock

import (
	"fmt"
)

func ExampleQueryMatcher() {
	// configure to use case sensitive SQL query matcher
	// instead of default regular expression matcher
	db, mock, err := New(QueryMatcherOption(QueryMatcherEqual))
	if err != nil {
		fmt.Println("failed to open sqlmock database:", err)
	}
	defer db.Close()

	rows := NewRows([]string{"id", "title"}).
		AddRow(1, "one").
		AddRow(2, "two")

	mock.ExpectSql(nil, "SELECT * FROM users").WillReturnRows(rows)

	rs, err := db.Query("SELECT * FROM users")
	if err != nil {
		fmt.Println("failed to match expected query")
		return
	}
	defer rs.Close()

	for rs.Next() {
		var id int
		var title string
		rs.Scan(&id, &title)
		fmt.Println("scanned id:", id, "and title:", title)
	}

	if rs.Err() != nil {
		fmt.Println("got rows error:", rs.Err())
	}
	// Output: scanned id: 1 and title: one
	// scanned id: 2 and title: two
}
