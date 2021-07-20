package main

import (
	"os"

	"github.com/EgorKurito/TokenCoin/cli"
)

func main() {
	defer os.Exit(0)
	cmd := cli.CommandLine{}
	cmd.Run()
}
