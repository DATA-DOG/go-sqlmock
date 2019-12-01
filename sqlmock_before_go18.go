// +build !go1.8

package sqlmock

import "log"

func (c *sqlmock) ExpectPing() *ExpectedPing {
	log.Println("ExpectPing has no effect on Go 1.7 or below")
	return &ExpectedPing{}
}
