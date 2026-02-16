package mkvinfo

import (
	"fmt"
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

	// Duration is uint64 nanoseconds; 0 means unknown.
	duration := time.Duration(segInfo.Duration)

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
