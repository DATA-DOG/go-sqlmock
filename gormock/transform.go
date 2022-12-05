package e2e

import (
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/pubgo/sqlmock"
	"gorm.io/gorm/schema"
)

func parseColumn(table interface{}) []string {
	column := make([]string, 0)
	s, err := schema.Parse(table, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		return column
	}
	for _, v := range s.Fields {
		if len(v.DBName) != 0 {
			column = append(column, v.DBName)
		}
	}
	return column
}

func parseValue(table interface{}, column []string) []driver.Value {
	row := make([]driver.Value, 0, len(column))
	s, err := schema.Parse(table, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		return row
	}

	var reflectValue = reflect.ValueOf(table)
	for _, col := range column {
		fv, _ := s.FieldsByDBName[col].ValueOf(context.Background(), reflectValue)
		row = append(row, fv)
	}
	return row
}

func parseField(table interface{}) map[string]driver.Value {
	row := make(map[string]driver.Value)
	s, err := schema.Parse(table, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		return row
	}

	column := parseColumn(table)
	var reflectValue = reflect.ValueOf(table)
	for _, col := range column {
		fv, zero := s.FieldsByDBName[col].ValueOf(context.Background(), reflectValue)
		if zero {
			continue
		}

		row[col] = fv
	}
	return row
}

func parseWhere(table interface{}) (string, []driver.Value) {
	column := parseColumn(table)
	row := make([]driver.Value, 0, len(column))
	s, err := schema.Parse(table, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		return "", row
	}

	var sql = ""
	var reflectValue = reflect.ValueOf(table)
	for _, col := range column {
		fv, zero := s.FieldsByDBName[col].ValueOf(context.Background(), reflectValue)
		if zero {
			continue
		}

		row = append(row, fv)
		if sql != "" {
			sql += " AND "
		}
		sql += fmt.Sprintf("%s *", col)
	}
	return sql, row
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
		columns = parseColumn(vv.Index(0).Interface())
		for i := 0; i < vv.Len(); i++ {
			values = append(values, vv.Index(i).Interface())
		}
	} else {
		columns = parseColumn(dst)
		values = append(values, dst)
	}

	rows := sqlmock.NewRows(columns)
	for i := range values {
		rows.AddRow(parseValue(values[i], columns)...)
	}
	return rows
}

func insertSql(tableName string, sql string) string {
	return "INSERT INTO" + fmt.Sprintf(` "%s" *%s VALUES *`, tableName, strings.ReplaceAll(sql, " ", ""))
}

func deleteSql(tableName string, where string) string {
	if where == "" {
		return "DELETE FROM" + fmt.Sprintf(` "%s"*`, tableName)
	}
	return "DELETE FROM" + fmt.Sprintf(` "%s" WHERE %s*`, tableName, where)
}

func updateSql(tableName string, where string) string {
	if where == "" {
		return "UPDATE" + fmt.Sprintf(` "%s" SET`, tableName)
	}
	return "UPDATE" + fmt.Sprintf(` "%s" SET * WHERE %s*`, tableName, where)
}

func selectSql(tableName string, where string) string {
	if where == "" {
		return "SELECT * FROM" + fmt.Sprintf(` "%s"*`, tableName)
	}
	return "SELECT * FROM" + fmt.Sprintf(` "%s" WHERE %s*`, tableName, where)
}

func parseVal(val interface{}) []driver.Value {
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
