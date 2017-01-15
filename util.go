package sqlmock

import (
	"regexp"
	"strings"
)

var re = regexp.MustCompile("\\s+")

// strip out new lines and trim spaces
func stripQuery(q string) (s string) {
	return strings.TrimSpace(re.ReplaceAllString(q, " "))
}

// mimicking how sql.DB build their queries
func buildQuery(q string)string{
	q = strings.TrimSpace(q)
	lines := strings.Split(q,"\n")
	var newQuery string
	for _,l := range lines{
		newQuery = newQuery +" " +strings.TrimSpace(l)
	}
	return strings.TrimSpace(newQuery)
}