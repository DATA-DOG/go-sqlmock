package sqlmock

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
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

func ExampleExpectedExec() {
	db, mock, _ := New()
	result := NewErrorResult(fmt.Errorf("some error"))
	mock.ExpectExec("^INSERT (.+)").WillReturnResult(result)
	res, _ := db.Exec("INSERT something")
	_, err := res.LastInsertId()
	fmt.Println(err)
	// Output: some error
}

func TestBuildQuery(t *testing.T) {
	db, mock, _ := New()
	query := `
		SELECT
			name,
			email,
			address,
			anotherfield
		FROM user
		where
			name    = 'John'
			and
			address = 'Jakarta'

	`

	mock.ExpectQuery(query)
	mock.ExpectExec(query)
	mock.ExpectPrepare(query)

	db.QueryRow(query)
	db.Exec(query)
	db.Prepare(query)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestCustomValueConverterQueryScan(t *testing.T) {
	db, mock, _ := New(ValueConverterOption(CustomConverter{}))
	query := `
		SELECT
			name,
			email,
			address,
			anotherfield
		FROM user
		where
			name    = 'John'
			and
			address = 'Jakarta'

	`
	expectedStringValue := "ValueOne"
	expectedIntValue := 2
	expectedArrayValue := []string{"Three", "Four"}
	mock.ExpectQuery(query).WillReturnRows(mock.NewRows([]string{"One", "Two", "Three"}).AddRow(expectedStringValue, expectedIntValue, []string{"Three", "Four"}))
	row := db.QueryRow(query)
	var stringValue string
	var intValue int
	var arrayValue []string
	if e := row.Scan(&stringValue, &intValue, &arrayValue); e != nil {
		t.Error(e)
	}
	if stringValue != expectedStringValue {
		t.Errorf("Expectation %s does not met: %s", expectedStringValue, stringValue)
	}
	if intValue != expectedIntValue {
		t.Errorf("Expectation %d does not met: %d", expectedIntValue, intValue)
	}
	if !reflect.DeepEqual(expectedArrayValue, arrayValue) {
		t.Errorf("Expectation %v does not met: %v", expectedArrayValue, arrayValue)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestQueryWithNoArgsAndWithArgsPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Error("Expected panic for using WithArgs and ExpectNoArgs together")
	}()
	mock := &sqlmock{}
	mock.ExpectQuery("SELECT (.+) FROM user").WithArgs("John").WithoutArgs()
}

func TestExecWithNoArgsAndWithArgsPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			return
		}
		t.Error("Expected panic for using WithArgs and ExpectNoArgs together")
	}()
	mock := &sqlmock{}
	mock.ExpectExec("^INSERT INTO user").WithArgs("John").WithoutArgs()
}


func TestQueryWillReturnsNil(t *testing.T) {
	t.Parallel()

	db, mock, err := New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	defer func() {
		if err := recover(); err != nil {
			t.Error(err)
		}
	}()

	mock.ExpectQuery("SELECT (.+) FROM users WHERE (.+)").WithArgs("test").WillReturnRows(nil)
	query := "SELECT name, email FROM users WHERE name = ?"
	_, err = mock.(*sqlmock).Query(query, []driver.Value{"test"})
	if err != nil {
		t.Error(err)
	}
}
