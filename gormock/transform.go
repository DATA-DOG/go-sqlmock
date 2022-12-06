package e2e

import (
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"
	"sync"

	"github.com/pubgo/sqlmock"
	"gorm.io/gorm/schema"
)

func parseColumn(table interface{}) []*schema.Field {
	column := make([]*schema.Field, 0)
	s, err := schema.Parse(table, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		return column
	}
	for _, v := range s.Fields {
		if len(v.DBName) != 0 {
			column = append(column, v)
		}
	}
	return column
}

func parseValue(table interface{}) []driver.Value {
	var row []driver.Value
	s, err := schema.Parse(table, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		return row
	}

	var reflectValue = reflect.ValueOf(table)
	for _, col := range parseColumn(table) {
		fv, _ := s.FieldsByDBName[col.DBName].ValueOf(context.Background(), reflectValue)
		row = append(row, fv)
	}
	return row
}

func ModelToRows(dst interface{}) *sqlmock.Rows {
	if dst == nil {
		return sqlmock.NewRows(nil)
	}

	var columns []string
	var vv = reflect.ValueOf(dst)
	if vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}

	var values []interface{}
	if vv.Kind() == reflect.Array || vv.Kind() == reflect.Slice {
		for _, f := range parseColumn(vv.Index(0).Interface()) {
			columns = append(columns, f.DBName)
		}
		for i := 0; i < vv.Len(); i++ {
			values = append(values, vv.Index(i).Interface())
		}
	} else {
		for _, f := range parseColumn(dst) {
			columns = append(columns, f.DBName)
		}
		values = append(values, dst)
	}

	rows := sqlmock.NewRows(columns)
	for i := range values {
		rows.AddRow(parseValue(values[i])...)
	}
	return rows
}

func insertSql(tableName string) string {
	return "INSERT INTO" + fmt.Sprintf(` "%s" *`, tableName)
}

func deleteSql(tableName string, where string) string {
	if where == "" {
		return "DELETE FROM" + fmt.Sprintf(` "%s"*`, tableName)
	}
	return "DELETE FROM" + fmt.Sprintf(` "%s" WHERE %s*`, tableName, where)
}

func updateSql(tableName string, where string) string {
	if where == "" {
		return "UPDATE" + fmt.Sprintf(` "%s" SET*`, tableName)
	}
	return "UPDATE" + fmt.Sprintf(` "%s" SET * WHERE %s*`, tableName, where)
}

func selectSql(tableName string, where string) string {
	if where == "" {
		return "SELECT * FROM" + fmt.Sprintf(` "%s"*`, tableName)
	}
	return "SELECT * FROM" + fmt.Sprintf(` "%s" WHERE %s*`, tableName, where)
}

func parseVal(val interface{}) []driver.Value { //nolint
	var values []driver.Value
	if val == nil {
		values = append(values, nil)
		return values
	}

	var vv = reflect.ValueOf(val)
	for vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}

	if vv.Kind() == reflect.Array || vv.Kind() == reflect.Slice {
		for i := 0; i < vv.Len(); i++ {
			values = append(values, vv.Index(i).Interface())
		}
	} else {
		values = append(values, val)
	}
	return values
}
