package main

import (
	"os"

	"github.com/1o-fyi/lofi/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
