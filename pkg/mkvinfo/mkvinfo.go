package mkvinfo

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	matroska "github.com/luispater/matroska-go"
)

// GetMKVInfo opens an MKV file and returns structured information about the file
// and all subtitle tracks it contains.
func GetMKVInfo(path string) (*MKVInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("cannot stat file: %w", err)
	}

	demuxer, err := matroska.NewDemuxer(file)
	if err != nil {
		return nil, fmt.Errorf("not a valid MKV file: %w", err)
	}
	defer demuxer.Close()

	segInfo, err := demuxer.GetFileInfo()
	if err != nil {
		return nil, fmt.Errorf("cannot read file info: %w", err)
	}

	// Matroska Duration is a float element, but matroska-go reads it via ReadUInt(),
	// returning the raw IEEE 754 bits as uint64. Reinterpret as float, then scale.
	// The element may be 4-byte (float32) or 8-byte (float64); uint64 values
	// fitting in 32 bits are float32 (any real float64 has upper bits set).
	var durationTicks float64
	if segInfo.Duration <= math.MaxUint32 {
		durationTicks = float64(math.Float32frombits(uint32(segInfo.Duration)))
	} else {
		durationTicks = math.Float64frombits(segInfo.Duration)
	}
	duration := time.Duration(durationTicks * float64(segInfo.TimecodeScale))

	numTracks, err := demuxer.GetNumTracks()
	if err != nil {
		return nil, fmt.Errorf("cannot get track count: %w", err)
	}

	var tracks []SubtitleTrack
	displayIndex := 1

	for i := uint(0); i < numTracks; i++ {
		info, err := demuxer.GetTrackInfo(i)
		if err != nil {
			continue // Skip tracks that can't be read
		}

		// Filter to subtitle tracks only (Type == 17)
		if info.Type != matroska.TypeSubtitle {
			continue
		}

		formatType, isText := ClassifyCodec(info.CodecID)

		track := SubtitleTrack{
			Number:        info.Number,
			Index:         displayIndex,
			Language:      info.Language,
			LanguageName:  ResolveLanguageName(info.Language),
			FormatType:    formatType,
			CodecID:       info.CodecID,
			Name:          info.Name,
			IsDefault:     info.Default,
			IsForced:      info.Forced,
			IsText:        isText,
			IsExtractable: isText, // Only text subtitles are extractable
		}

		tracks = append(tracks, track)
		displayIndex++
	}

	// Count totals
	subtitleCount := len(tracks)
	textSubCount := 0
	for _, t := range tracks {
		if t.IsText {
			textSubCount++
		}
	}

	result := &MKVInfo{
		Info: FileInfo{
			FileName:      filepath.Base(path),
			FileSize:      stat.Size(),
			Duration:      duration,
			SubtitleCount: subtitleCount,
			TextSubCount:  textSubCount,
		},
		Tracks: tracks,
	}

	return result, nil
}
