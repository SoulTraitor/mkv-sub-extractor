// Package extract provides MKV subtitle packet extraction functionality.
//
// pipeline.go contains the public integration API that orchestrates the full
// extraction pipeline: MKV demuxing, packet extraction, codec-specific parsing,
// output file naming, and ASS file writing.
package extract

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	matroska "github.com/luispater/matroska-go"

	"mkv-sub-extractor/pkg/assout"
	"mkv-sub-extractor/pkg/mkvinfo"
	"mkv-sub-extractor/pkg/output"
	"mkv-sub-extractor/pkg/subtitle"
)

// ExtractTrackToASS extracts a single subtitle track from an MKV file and writes
// it as an ASS output file. This is the primary public API for Phase 3 (CLI).
//
// For ASS/SSA tracks, the original styles from CodecPrivate are preserved (passthrough mode).
// For SRT tracks, a default ASS header with Microsoft YaHei font is generated.
//
// If outputDir is empty, the output file is written to the same directory as the MKV file.
// Returns the path to the created ASS file.
func ExtractTrackToASS(mkvPath string, track mkvinfo.SubtitleTrack, outputDir string) (string, error) {
	existingPaths := make(map[string]bool)
	return extractTrackToASS(mkvPath, track, outputDir, existingPaths)
}

// ExtractTracksToASS extracts multiple subtitle tracks from an MKV file, writing
// each as a separate ASS output file. The existingPaths map is shared across all
// tracks to handle naming collisions correctly.
//
// Returns a slice of output file paths in the same order as the input tracks.
func ExtractTracksToASS(mkvPath string, tracks []mkvinfo.SubtitleTrack, outputDir string) ([]string, error) {
	if len(tracks) == 0 {
		return nil, nil
	}

	existingPaths := make(map[string]bool)
	var outputPaths []string

	for _, track := range tracks {
		outPath, err := extractTrackToASS(mkvPath, track, outputDir, existingPaths)
		if err != nil {
			return outputPaths, fmt.Errorf("track %d (%s): %w", track.Number, track.FormatType, err)
		}
		outputPaths = append(outputPaths, outPath)
	}

	return outputPaths, nil
}

// extractTrackToASS is the internal implementation shared by both single and batch extraction.
// It accepts a shared existingPaths map for collision-aware output naming.
func extractTrackToASS(mkvPath string, track mkvinfo.SubtitleTrack, outputDir string, existingPaths map[string]bool) (string, error) {
	// 1. Open MKV file and create demuxer
	file, err := os.Open(mkvPath)
	if err != nil {
		return "", fmt.Errorf("open MKV file: %w", err)
	}
	defer file.Close()

	demuxer, err := matroska.NewDemuxer(file)
	if err != nil {
		return "", fmt.Errorf("create demuxer: %w", err)
	}
	defer demuxer.Close()

	// 2. Get track info for CodecPrivate
	var codecPrivate []byte
	numTracks, err := demuxer.GetNumTracks()
	if err != nil {
		return "", fmt.Errorf("get track count: %w", err)
	}

	found := false
	for i := uint(0); i < numTracks; i++ {
		info, err := demuxer.GetTrackInfo(i)
		if err != nil {
			continue
		}
		if info.Number == track.Number {
			codecPrivate = info.CodecPrivate
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("track number %d not found in MKV file", track.Number)
	}

	// 3. Extract raw packets
	packets, warning, err := ExtractSubtitlePackets(demuxer, track.Number)
	if err != nil {
		return "", fmt.Errorf("extract packets: %w", err)
	}
	_ = warning // Caller can check via logging if needed

	// 4. Gap-fill is already applied by ExtractSubtitlePackets

	// 5. Convert raw packets to SubtitleEvents based on codec
	events, err := packetsToEvents(packets, track.CodecID)
	if err != nil {
		return "", fmt.Errorf("convert packets to events: %w", err)
	}

	// 6. Determine output path
	if outputDir == "" {
		outputDir = filepath.Dir(mkvPath)
	}
	videoForNaming := filepath.Join(outputDir, filepath.Base(mkvPath))
	outputPath := output.GenerateOutputPath(videoForNaming, track, existingPaths)

	// 7. Write ASS output
	outFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("create output file: %w", err)
	}
	defer outFile.Close()

	switch {
	case track.CodecID == "S_TEXT/ASS" || track.CodecID == "S_TEXT/SSA":
		err = assout.WriteASSPassthrough(outFile, codecPrivate, track.CodecID, events)
	case track.CodecID == "S_TEXT/UTF8":
		err = assout.WriteSRTAsASS(outFile, events)
	default:
		return "", fmt.Errorf("unsupported codec ID: %s", track.CodecID)
	}
	if err != nil {
		// Clean up partial output file on write error
		outFile.Close()
		os.Remove(outputPath)
		return "", fmt.Errorf("write ASS output: %w", err)
	}

	return outputPath, nil
}

// packetsToEvents converts raw subtitle packets to SubtitleEvents based on codec type.
func packetsToEvents(packets []RawSubtitlePacket, codecID string) ([]subtitle.SubtitleEvent, error) {
	events := make([]subtitle.SubtitleEvent, 0, len(packets))

	switch {
	case codecID == "S_TEXT/ASS" || codecID == "S_TEXT/SSA":
		for i, pkt := range packets {
			readOrder, layer, remaining, err := subtitle.ParseASSBlockData(pkt.Data)
			if err != nil {
				return nil, fmt.Errorf("packet %d: parse ASS block data: %w", i, err)
			}

			// Split remaining into at most 7 parts: Style, Name, MarginL, MarginR, MarginV, Effect, Text
			// remaining = "Style,Name,MarginL,MarginR,MarginV,Effect,Text"
			parts := strings.SplitN(remaining, ",", 7)
			if len(parts) < 7 {
				return nil, fmt.Errorf("packet %d: expected 7 fields in remaining %q, got %d", i, remaining, len(parts))
			}

			events = append(events, subtitle.SubtitleEvent{
				Start:     pkt.StartTime,
				End:       pkt.EndTime,
				Layer:     layer,
				Style:     parts[0],
				Name:      parts[1],
				MarginL:   parts[2],
				MarginR:   parts[3],
				MarginV:   parts[4],
				Effect:    parts[5],
				Text:      parts[6],
				ReadOrder: readOrder,
			})
		}

	case codecID == "S_TEXT/UTF8":
		for _, pkt := range packets {
			events = append(events, subtitle.SubtitleEvent{
				Start:   pkt.StartTime,
				End:     pkt.EndTime,
				Layer:   0,
				Style:   "Default",
				Name:    "",
				MarginL: "0",
				MarginR: "0",
				MarginV: "0",
				Effect:  "",
				Text:    string(pkt.Data),
			})
		}

	default:
		return nil, fmt.Errorf("unsupported codec ID: %s", codecID)
	}

	return events, nil
}
