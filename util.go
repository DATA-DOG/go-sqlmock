package sqlmock

import (
	"database/sql/driver"
	"encoding/json"
	"log"
)

func jsonify(val interface{}) string {
	var data, err = json.Marshal(val)
	if err != nil {
		log.Fatalln(err)
	}
	return string(data)
}

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

func convValue(args []driver.NamedValue) []driver.Value {
	values := make([]driver.Value, len(args))
	for i := range args {
		values[i] = args[i].Value
	}
	return values
}

func rawBytes(col driver.Value) (_ []byte, ok bool) {
	val, ok := col.([]byte)
	if !ok || len(val) == 0 {
		return nil, false
	}
	// Copy the bytes from the mocked row into a shared raw buffer, which we'll replace the content of later
	// This allows scanning into sql.RawBytes to correctly become invalid on subsequent calls to Next(), Scan() or Close()
	b := make([]byte, len(val))
	copy(b, val)
	return b, true
}
