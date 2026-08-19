package main

import (
	"os"

	"github.com/haritsAchmad/devkit-playground/internal/cli"
)

var version = "dev"

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, version))
}
