package main

import (
	"os"
	"shortk/internal/cli"
)

func main() {
	cli.Run(os.Args[1:])
}
