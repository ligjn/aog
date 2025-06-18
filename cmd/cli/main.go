package main

import (
	"os"

	cli "github.com/ligjn/aog/cmd/cli/core"
)

func main() {
	command := cli.NewCommand()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
