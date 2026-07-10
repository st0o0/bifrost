package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		// Wired to the real healthcheck in Phase 2.
		os.Exit(0)
	}
	fmt.Fprintln(os.Stdout, "bifrost")
}
