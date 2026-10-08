//go:build go1.10
// +build go1.10

package sqlmock

import (
	"context"
	"database/sql/driver"
)

// OpenConnector binds db.Driver() to its mock, so it can be registered under another name.
func (d *mockDriver) OpenConnector(dsn string) (driver.Connector, error) {
	return &mockConnector{dsn: dsn, drv: d}, nil
}

type mockConnector struct {
	dsn string
	drv *mockDriver
}

func (c *mockConnector) Connect(context.Context) (driver.Conn, error) { return c.drv.Open(c.dsn) }

func (c *mockConnector) Driver() driver.Driver { return c }

// Open opens this mock database for an empty dsn, any other by its dsn.
func (c *mockConnector) Open(dsn string) (driver.Conn, error) {
	if dsn == "" {
		dsn = c.dsn
	}
	return c.drv.Open(dsn)
}
