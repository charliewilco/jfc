package main

import (
	"os"

	"github.com/charliewilco/jfc/internal/jfc"
)

func main() {
	os.Exit(jfc.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, os.Getwd))
}
