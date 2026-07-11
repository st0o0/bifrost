//go:build !linux

package main

import "os"

func main() {
	if len(os.Args) > 1 {
		if handled, code := runCmd(os.Args[1], os.Stdin, os.Stdout); handled {
			os.Exit(code)
		}
	}
	os.Exit(0)
}
