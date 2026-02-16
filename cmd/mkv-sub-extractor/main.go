package main

import (
	"fmt"
	"os"

	"mkv-sub-extractor/pkg/extract"
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

	// Temporary: extract the first extractable text subtitle track for pipeline verification
	for _, track := range result.Tracks {
		if !track.IsExtractable {
			continue
		}
		fmt.Printf("\nExtracting track %d (%s, %s)...\n", track.Index, track.FormatType, track.LanguageName)
		outputPath, err := extract.ExtractTrackToASS(path, track, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Extraction error: %v\n", err)
		} else {
			fmt.Printf("  -> %s\n", outputPath)
		}
		break // Only extract the first track for verification
	}
}
