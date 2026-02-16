package main

import (
	"fmt"
	"os"

	"mkv-sub-extractor/pkg/mkvinfo"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: mkv-sub-extractor <file.mkv>")
		os.Exit(1)
	}

	path := os.Args[1]

	result, err := mkvinfo.GetMKVInfo(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(mkvinfo.FormatTrackListing(*result))
}
