package main

import (
	"os"

	"github.com/Flpvoigt/Golang_project/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
