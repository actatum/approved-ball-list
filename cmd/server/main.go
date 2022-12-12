// Package main provides the entrypoint to the service when its run as a server.
package main

import (
	"fmt"
	"os"

	"github.com/actatum/approved-ball-list/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
