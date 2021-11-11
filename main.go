package main

import (
	"os"

	"git.sr.ht/~lofi/lib"
	"github.com/1o-fyi/lofi/cmd"
)

func main() {
	os.Stdout.Write(<-lib.EncodeHex([]byte("1o.fyi")))

	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
