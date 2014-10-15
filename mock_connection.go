package sqlmock

import (
	"fmt"
	"regexp"
	"sync"
)

var (
	mutex           sync.RWMutex
	mockConnections = make(map[string]*MockConn)
)

// MockConn is a mocked connection for database/sql
//
// It is a (database/sql/driver).Conn, so it can be used just like a normal connection
//
// Apply the expectations to the MockConn
type MockConn struct {
	*conn
}

// NewMockConn creates a new mock connection instance
//
// Instances are identified using an ID string. You can pass the expected ID as a DSN
// like so:
//
//   db, err := sql.Open("mock", "id=instance_id")
//
// IDs are unique, creating a new mock conn when the ID is already present will
// end in an error.
func NewMockConn(ID string) (*MockConn, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := mockConnections[ID]; ok {
		return nil, fmt.Errorf("there is already a connection with the ID %s", ID)
	}
	c := &MockConn{conn: &conn{}}
	mockConnections[ID] = c
	return c, nil
}

func mockConn(ID string) (*MockConn, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	if conn, ok := mockConnections[ID]; ok {
		return conn, nil
	} else {
		return nil, fmt.Errorf("no connection with ID %s", ID)
	}
}

// ExpectBegin expects transaction to be started
func (mock *MockConn) ExpectBegin() Mock {
	e := &expectedBegin{}
	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

// ExpectCommit expects transaction to be commited
func (mock *MockConn) ExpectCommit() Mock {
	e := &expectedCommit{}
	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

// ExpectRollback expects transaction to be rolled back
func (mock *MockConn) ExpectRollback() Mock {
	e := &expectedRollback{}
	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

// ExpectPrepare expects Query to be prepared
func (mock *MockConn) ExpectPrepare() Mock {
	e := &expectedPrepare{}
	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

// ExpectExec expects database Exec to be triggered, which will match
// the given query string as a regular expression
func (mock *MockConn) ExpectExec(sqlRegexStr string) Mock {
	e := &expectedExec{}
	e.sqlRegex = regexp.MustCompile(sqlRegexStr)
	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

// ExpectQuery database Query to be triggered, which will match
// the given query string as a regular expression
func (mock *MockConn) ExpectQuery(sqlRegexStr string) Mock {
	e := &expectedQuery{}
	e.sqlRegex = regexp.MustCompile(sqlRegexStr)

	mock.conn.expectations = append(mock.conn.expectations, e)
	mock.conn.active = e
	return mock.conn
}

// Close a mock database driver connection. It should
// be always called to ensure that all expectations
// were met successfully. Returns error if there is any
func (mock *MockConn) Close() (err error) {
	for _, e := range mock.conn.expectations {
		if !e.fulfilled() {
			err = fmt.Errorf("there is a remaining expectation %T which was not matched yet", e)
			break
		}
	}
	mock.conn.expectations = []expectation{}
	mock.conn.active = nil
	return err
}
