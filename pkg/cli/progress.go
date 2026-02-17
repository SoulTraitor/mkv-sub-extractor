package cli

import (
	"mkv-sub-extractor/pkg/extract"
	"mkv-sub-extractor/pkg/mkvinfo"

	"github.com/schollz/progressbar/v3"
)

// TrackResult holds the outcome of extracting a single subtitle track.
// Used for continue-on-failure extraction where each track is attempted
// independently and all results are reported at the end.
type TrackResult struct {
	Track      mkvinfo.SubtitleTrack
	OutputPath string
	Error      error
}

// ExtractWithProgress extracts multiple subtitle tracks from an MKV file,
// showing a progress bar unless quiet mode is enabled. Each track is extracted
// independently — if one fails, extraction continues for remaining tracks.
//
// Returns a TrackResult for every track in the same order as the input slice.
func ExtractWithProgress(mkvPath string, tracks []mkvinfo.SubtitleTrack, outputDir string, quiet bool) []TrackResult {
	results := make([]TrackResult, len(tracks))

	var bar *progressbar.ProgressBar
	if !quiet {
		bar = progressbar.Default(int64(len(tracks)), "Extracting")
	}

	for i, track := range tracks {
		outputPath, err := extract.ExtractTrackToASS(mkvPath, track, outputDir)
		results[i] = TrackResult{
			Track:      track,
			OutputPath: outputPath,
			Error:      err,
		}

		if bar != nil {
			_ = bar.Add(1)
		}
	}

	if bar != nil {
		_ = bar.Finish()
	}

	return results
}
