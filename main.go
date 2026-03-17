package main

import (
	"os"

	"github.com/aleks/switchnix/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
