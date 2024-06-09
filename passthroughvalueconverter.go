package sqlmock

import (
	"database/sql/driver"
	"fmt"
)

type PassthroughValueConverter struct {
	passthroughTypes []string
}

func NewPassthroughValueConverter(typeSamples ...interface{}) *PassthroughValueConverter {
	c := &PassthroughValueConverter{}

	for _, sampleValue := range typeSamples {
		c.passthroughTypes = append(c.passthroughTypes, fmt.Sprintf("%T", sampleValue))
	}

	return c
}

func (c *PassthroughValueConverter) ConvertValue(v interface{}) (driver.Value, error) {
	valueType := fmt.Sprintf("%T", v)
	for _, passthroughType := range c.passthroughTypes {
		if valueType == passthroughType {
			return v, nil
		}
	}

	return driver.DefaultParameterConverter.ConvertValue(v)
}
