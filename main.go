package main

import (
	"os"

	"jfc/internal/jfc"
)

func main() {
	os.Exit(jfc.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, os.Getwd))
}
