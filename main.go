package main

import (
	"os"

	"git.sr.ht/~johns/lofi/cmd"
)

func main() {
	cmd.Execute()
}

func out(b []byte) {
	os.Stdout.Write([]byte("\n"))
	os.Stdout.Write(b)
}
