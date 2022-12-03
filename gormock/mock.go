package e2e

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/pubgo/sqlmock"
	"github.com/tidwall/match"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type dbMock struct {
	tb   TestingTB
	mock sqlmock.Sqlmock
	db   *gorm.DB

	query     bool
	delete    bool
	update    bool
	create    bool
	tx        bool
	prepare   bool
	fields    map[string]driver.Value
	column    []string
	tableName string
}

func (m *dbMock) Mock() sqlmock.Sqlmock { return m.mock }
func (m *dbMock) DB() *gorm.DB          { return m.db }

func (m *dbMock) createExpect(model schema.Tabler) *dbMock {
	if model == nil {
		m.tb.Fatalf("model is nil")
		return m
	}

	return &dbMock{
		mock:      m.mock,
		db:        m.db,
		tb:        m.tb,
		tableName: model.TableName(),
		column:    parseColumn(model),
		fields:    parseField(model),
	}
}

func (m *dbMock) do(err error, ret driver.Result, rows *sqlmock.Rows) {
	if m.tx {
		m.mock.ExpectBegin()
	}

	var sql = ""
	var args []driver.Value
	for _, name := range m.column {
		var val, ok = m.fields[name]
		if ok {
			sql += fmt.Sprintf("%s *", name)
			args = append(args, parseVal(val)...)
		}
	}

	if m.prepare {
		// TODO prepare
	}

	var e interface{}
	if m.query {
		e = m.mock.ExpectQuery(selectSql(m.tableName, sql))
	}

	if m.create {
		e = m.mock.ExpectQuery(insertSql(m.tableName, sql))
	}

	if m.update {
		e = m.mock.ExpectExec(updateSql(m.tableName, sql))
	}

	if m.delete {
		e = m.mock.ExpectExec(deleteSql(m.tableName, sql))
	}

	switch e := e.(type) {
	case *sqlmock.ExpectedQuery:
		e.WithArgs(args...)
		if err != nil {
			e.WillReturnError(err)
		}

		if rows != nil {
			e.WillReturnRows(rows)
		}
	case *sqlmock.ExpectedExec:
		e.WithArgs(args...)
		if err != nil {
			e.WillReturnError(err)
		}

		if ret != nil {
			e.WillReturnResult(ret)
		}
	}

	if m.tx {
		if err == nil {
			m.mock.ExpectCommit()
		} else {
			m.mock.ExpectRollback()
		}
	}
}

func (m *dbMock) ExpectTx() *dbMock {
	m.tx = true
	return m
}

func (m *dbMock) ExpectPrepare() *dbMock {
	m.prepare = true
	return m
}

func (m *dbMock) ExpectField(name string, value interface{}) *dbMock {
	m.fields[name] = value
	return m
}

func (m *dbMock) ExpectErr(err error) {
	m.do(err, nil, nil)
}

func (m *dbMock) ExpectResult(lastInsertID int64, rowsAffected int64) {
	m.do(nil, sqlmock.NewResult(lastInsertID, rowsAffected), nil)
}

func (m *dbMock) ExpectReturn(returns interface{}) {
	m.do(nil, nil, ModelToRows(returns))
}

func (m *dbMock) Insert(model schema.Tabler) *dbMock {
	var mm = m.createExpect(model)
	mm.create = true
	return mm
}

func (m *dbMock) Delete(model schema.Tabler) *dbMock {
	var mm = m.createExpect(model)
	mm.delete = true
	return mm
}

func (m *dbMock) Update(model schema.Tabler) *dbMock {
	var mm = m.createExpect(model)
	mm.update = true
	return mm
}

func (m *dbMock) Find(model schema.Tabler) *dbMock {
	var mm = m.createExpect(model)
	mm.query = true
	return mm
}

func NewMockDB(tb TestingTB) *dbMock {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		expectedSQL = strings.ReplaceAll(expectedSQL, "**", "*")

		tb.Logf("\n expectedSQL => %s \n actualSQL   => %s \n matchSQL    => %v",
			expectedSQL, actualSQL, match.Match(actualSQL, expectedSQL))

		if match.Match(actualSQL, expectedSQL) {
			return nil
		}

		return fmt.Errorf(`could not match actual sql: "%s" with expected regexp "%s"`, actualSQL, expectedSQL)
	})))

	if err != nil {
		tb.Fatalf("%v", err)
		return nil
	}

	tb.Cleanup(func() {
		err := mock.ExpectationsWereMet()
		if err != nil {
			tb.Fatalf("%v", err)
		}
	})

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  "sqlmock_db_0",
		DriverName:           "postgres",
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		tb.Fatalf("%v", err)
		return nil
	}

	return &dbMock{db: gormDB, mock: mock, tb: tb}
}
