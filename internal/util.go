package internal

import (
	"io"
	"log"
)

func HandleClose(closer io.Closer) {
	if closer == nil {
		return
	}

	var err = closer.Close()
	if err != nil {
		log.Printf("%v", err)
	}
}
