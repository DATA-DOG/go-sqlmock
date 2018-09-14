package sqlmock

import "database/sql/driver"

// ValueConverterOption allows to create a sqlmock connection
// with a custom ValueConverter to support drivers with special data types.
func ValueConverterOption(converter driver.ValueConverter) func(*sqlmock) error {
	return func(s *sqlmock) error {
		s.converter = converter
		return nil
	}
}
