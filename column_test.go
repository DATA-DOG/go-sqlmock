package sqlmock

import (
	"reflect"
	"testing"
	"time"
)

func TestColumn(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2020-06-20T22:08:41Z")
	column1 := NewColumn("test").OfType("VARCHAR", "").Nullable(true).WithLength(100)
	column2 := NewColumn("number").OfType("DECIMAL", float64(0.0)).Nullable(false).WithPrecisionAndScale(10, 4)
	column3 := NewColumn("when").OfType("TIMESTAMP", now)

	if column1.ScanType().Kind() != reflect.String {
		t.Errorf("string scanType mismatch: %v", column1.ScanType())
	}
	if column2.ScanType().Kind() != reflect.Float64 {
		t.Errorf("float scanType mismatch: %v", column2.ScanType())
	}
	if column3.ScanType() != reflect.TypeOf(time.Time{}) {
		t.Errorf("time scanType mismatch: %v", column3.ScanType())
	}

	nullable, ok := column1.IsNullable()
	if !nullable || !ok {
		t.Errorf("'test' column should be nullable")
	}
	nullable, ok = column2.IsNullable()
	if nullable || !ok {
		t.Errorf("'number' column should not be nullable")
	}
	nullable, ok = column3.IsNullable()
	if ok {
		t.Errorf("'when' column nullability should be unknown")
	}

	length, ok := column1.Length()
	if length != 100 || !ok {
		t.Errorf("'test' column wrong length")
	}
	length, ok = column2.Length()
	if ok {
		t.Errorf("'number' column is not of variable length type")
	}
	length, ok = column3.Length()
	if ok {
		t.Errorf("'when' column is not of variable length type")
	}

	_, _, ok = column1.PrecisionScale()
	if ok {
		t.Errorf("'test' column not applicable")
	}
	precision, scale, ok := column2.PrecisionScale()
	if precision != 10 || scale != 4 || !ok {
		t.Errorf("'number' column not applicable")
	}
	_, _, ok = column3.PrecisionScale()
	if ok {
		t.Errorf("'when' column not applicable")
	}
}
