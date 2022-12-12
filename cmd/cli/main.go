// Package main is the entrypoint for the CLI tool.
package main

import (
	"os"

	"github.com/actatum/approved-ball-list/internal/app"
)

func main() {
	app.CLI(os.Args[1:])
}
