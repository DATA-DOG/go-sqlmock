package sqlmock

import "strings"

// strip out new lines and trim spaces
func stripQuery(q string) (s string) {
	s = strings.Replace(q, "\n", " ", -1)
	s = strings.Replace(s, "\r", "", -1)
	s = strings.TrimSpace(s)
	return
}
