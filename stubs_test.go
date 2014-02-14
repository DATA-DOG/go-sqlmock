package sqlmock

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"
)

type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

type NullInt struct {
	Integer int
	Valid   bool
}

// Satisfy sql.Scanner interface
func (ni *NullInt) Scan(value interface{}) (err error) {
	if value == nil {
		ni.Integer, ni.Valid = 0, false
		return
	}

	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		ni.Integer, ni.Valid = v.(int), true
		return
	case []byte:
		ni.Integer, err = strconv.Atoi(string(v))
		ni.Valid = (err == nil)
		return
	case string:
		ni.Integer, err = strconv.Atoi(v)
		ni.Valid = (err == nil)
		return
	}

	ni.Valid = false
	return fmt.Errorf("Can't convert %T to integer", value)
}

// Satisfy sql.Valuer interface.
func (ni NullInt) Value() (driver.Value, error) {
	if !ni.Valid {
		return nil, nil
	}
	return ni.Integer, nil
}

// Satisfy sql.Scanner interface
func (nt *NullTime) Scan(value interface{}) (err error) {
	if value == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return
	}

	switch v := value.(type) {
	case time.Time:
		nt.Time, nt.Valid = v, true
		return
	}

	nt.Valid = false
	return fmt.Errorf("Can't convert %T to time.Time", value)
}

// Satisfy sql.Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}
