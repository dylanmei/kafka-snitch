package main

import (
	"fmt"
	"io"
)

const Version = "0.1"

var Revision string

func PrintVersion(w io.Writer) {
	rev := Revision
	if rev == "" {
		rev = "dev"
	}

	fmt.Fprintf(w, "kafka-snitch %s.%s\n", Version, rev)
}
