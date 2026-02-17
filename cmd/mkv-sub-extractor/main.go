package main

import (
	"os"

	"mkv-sub-extractor/pkg/cli"
)

func main() {
	os.Exit(cli.Run())
}
