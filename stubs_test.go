package sqlmock

import (
	"database/sql/driver"
	"errors"
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
	case int:
		ni.Integer, ni.Valid = v, true
		return
	case int8:
		ni.Integer, ni.Valid = int(v), true
		return
	case int16:
		ni.Integer, ni.Valid = int(v), true
		return
	case int32:
		ni.Integer, ni.Valid = int(v), true
		return
	case int64:
		const maxUint = ^uint(0)
		const minUint = 0
		const maxInt = int(maxUint >> 1)
		const minInt = -maxInt - 1

		if v > int64(maxInt) || v < int64(minInt) {
			return errors.New("value out of int range")
		}
		ni.Integer, ni.Valid = int(v), true
		return
	case []byte:
		n, err := strconv.Atoi(string(v))
		if err != nil {
			return err
		}
		ni.Integer, ni.Valid = n, true
		return nil
	case string:
		n, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		ni.Integer, ni.Valid = n, true
		return nil
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
