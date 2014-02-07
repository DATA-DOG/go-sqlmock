package sqlmock

import (
	"testing"
)

func TestQueryStringStripping(t *testing.T) {
	assert := func(actual, expected string) {
		if res := stripQuery(actual); res != expected {
			t.Errorf("Expected '%s' to be '%s', but got '%s'", actual, expected, res)
		}
	}

	assert(" SELECT 1", "SELECT 1")
	assert("SELECT   1 FROM   d", "SELECT 1 FROM d")
	assert(`
    SELECT c
    FROM D
`, "SELECT c FROM D")
	assert("UPDATE  (.+) SET  ", "UPDATE (.+) SET")
}
