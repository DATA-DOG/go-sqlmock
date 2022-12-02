package sqlmock

import (
	"database/sql/driver"
)

func convNameValue(args []driver.Value) []driver.NamedValue {
	namedArgs := make([]driver.NamedValue, len(args))
	for i := range args {
		v := args[i]
		switch v := v.(type) {
		case driver.NamedValue:
			namedArgs[i] = v
		case *driver.NamedValue:
			namedArgs[i] = *v
		default:
			namedArgs[i] = driver.NamedValue{Ordinal: i + 1, Value: v}
		}
	}
	return namedArgs
}
