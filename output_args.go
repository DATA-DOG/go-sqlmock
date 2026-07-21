package sqlmock

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
)

type namedInOutValue struct {
	Name             string
	ExpectedInValue  interface{}
	ReturnedOutValue interface{}
	In               bool
}

// Match implements the Argument interface, allowing check if the given value matches the expected input value provided using NamedInputOutputArg function.
func (n namedInOutValue) Match(v driver.Value) bool {
	out, ok := v.(sql.Out)

	return ok && out.In == n.In && (!n.In || reflect.DeepEqual(out.Dest, n.ExpectedInValue))
}

// NamedInputArg can ben used to simulate an output value passed back from the database.
// returnedOutValue can be a value or a pointer to the value.
func NamedOutputArg(name string, returnedOutValue interface{}) interface{} {
	return namedInOutValue{
		Name:             name,
		ReturnedOutValue: returnedOutValue,
		In:               false,
	}
}

// NamedInputOutputArg can be used to both check if expected input value is provided and to simulate an output value passed back from the database.
// expectedInValue must be a pointer to the value, returnedOutValue can be a value or a pointer to the value.
func NamedInputOutputArg(name string, expectedInValue interface{}, returnedOutValue interface{}) interface{} {
	return namedInOutValue{
		Name:             name,
		ExpectedInValue:  expectedInValue,
		ReturnedOutValue: returnedOutValue,
		In:               true,
	}
}

type typedOutValue struct {
	TypeName         string
	ReturnedOutValue interface{}
}

// Match implements the Argument interface, allowing check if the given value matches the expected type provided using TypedOutputArg function.
func (n typedOutValue) Match(v driver.Value) bool {
	return n.TypeName == fmt.Sprintf("%T", v)
}

// TypeOutputArg can be used to simulate an output value passed back from the database, setting value based on the type.
// returnedOutValue must be a pointer to the value.
func TypedOutputArg(returnedOutValue interface{}) interface{} {
	return typedOutValue{
		TypeName:         fmt.Sprintf("%T", returnedOutValue),
		ReturnedOutValue: returnedOutValue,
	}
}

func setOutputValues(currentArgs []driver.NamedValue, expectedArgs []driver.Value) {
	for _, expectedArg := range expectedArgs {
		if outVal, ok := expectedArg.(namedInOutValue); ok {
			for _, currentArg := range currentArgs {
				if currentArg.Name == outVal.Name {
					if sqlOut, ok := currentArg.Value.(sql.Out); ok {
						reflect.ValueOf(sqlOut.Dest).Elem().Set(reflect.Indirect(reflect.ValueOf(outVal.ReturnedOutValue)))
					}

					break
				}
			}
		}

		if outVal, ok := expectedArg.(typedOutValue); ok {
			for _, currentArg := range currentArgs {
				if fmt.Sprintf("%T", currentArg.Value) == outVal.TypeName {
					reflect.ValueOf(currentArg.Value).Elem().Set(reflect.Indirect(reflect.ValueOf(outVal.ReturnedOutValue)))

					break
				}
			}
		}
	}
}
