package mkvinfo

import (
	"strings"
	"testing"
	"time"
)

func TestFormatTrackLine_TextSubtitleAllFields(t *testing.T) {
	track := SubtitleTrack{
		Index:        1,
		LanguageName: "English",
		FormatType:   "SRT",
		CodecID:      "S_TEXT/UTF8",
		Name:         "Commentary",
		IsDefault:    true,
		IsForced:     false,
		IsText:       true,
	}
	got := FormatTrackLine(track, true)
	want := `[1] English (SRT) "Commentary" S_TEXT/UTF8 [default]`
	if got != want {
		t.Errorf("FormatTrackLine() =\n  %q\nwant:\n  %q", got, want)
	}
}

func TestFormatTrackLine_ImageSubtitle(t *testing.T) {
	track := SubtitleTrack{
		Index:        4,
		LanguageName: "English",
		FormatType:   "PGS",
		CodecID:      "S_HDMV/PGS",
		IsDefault:    false,
		IsForced:     false,
		IsText:       false,
	}
	got := FormatTrackLine(track, false)
	if !strings.HasSuffix(got, "-- not extractable (image subtitle)") {
		t.Errorf("FormatTrackLine() for image track should end with image marker, got: %q", got)
	}
	want := `[4] English (PGS) S_HDMV/PGS -- not extractable (image subtitle)`
	if got != want {
		t.Errorf("FormatTrackLine() =\n  %q\nwant:\n  %q", got, want)
	}
}

func TestFormatTrackLine_NoTrackName(t *testing.T) {
	track := SubtitleTrack{
		Index:        2,
		LanguageName: "中文",
		FormatType:   "ASS",
		CodecID:      "S_TEXT/ASS",
		Name:         "", // No name
		IsDefault:    false,
		IsForced:     false,
		IsText:       true,
	}
	got := FormatTrackLine(track, false)
	// Should not contain any quotes
	if strings.Contains(got, `"`) {
		t.Errorf("FormatTrackLine() with empty name should not contain quotes, got: %q", got)
	}
	want := `[2] 中文 (ASS) S_TEXT/ASS`
	if got != want {
		t.Errorf("FormatTrackLine() =\n  %q\nwant:\n  %q", got, want)
	}
}

func TestFormatTrackLine_NoFlags(t *testing.T) {
	track := SubtitleTrack{
		Index:        3,
		LanguageName: "日本語",
		FormatType:   "ASS",
		CodecID:      "S_TEXT/ASS",
		Name:         "",
		IsDefault:    false,
		IsForced:     false,
		IsText:       true,
	}
	got := FormatTrackLine(track, true) // showDefault=true but IsDefault=false
	if strings.Contains(got, "[default]") || strings.Contains(got, "[forced]") {
		t.Errorf("FormatTrackLine() should have no flags, got: %q", got)
	}
	want := `[3] 日本語 (ASS) S_TEXT/ASS`
	if got != want {
		t.Errorf("FormatTrackLine() =\n  %q\nwant:\n  %q", got, want)
	}
}

func TestFormatTrackLine_ForcedFlag(t *testing.T) {
	track := SubtitleTrack{
		Index:        1,
		LanguageName: "English",
		FormatType:   "SRT",
		CodecID:      "S_TEXT/UTF8",
		IsDefault:    false,
		IsForced:     true,
		IsText:       true,
	}
	got := FormatTrackLine(track, false)
	want := `[1] English (SRT) S_TEXT/UTF8 [forced]`
	if got != want {
		t.Errorf("FormatTrackLine() =\n  %q\nwant:\n  %q", got, want)
	}
}

func TestFormatFileInfoHeader(t *testing.T) {
	info := FileInfo{
		FileName:      "movie.mkv",
		FileSize:      8800000000, // ~8.2 GB
		Duration:      2*time.Hour + 15*time.Minute + 30*time.Second,
		SubtitleCount: 4,
		TextSubCount:  3,
	}
	got := FormatFileInfoHeader(info)

	// Check each expected component
	if !strings.Contains(got, "File: movie.mkv") {
		t.Errorf("header should contain file name, got: %q", got)
	}
	if !strings.Contains(got, "Duration: 2:15:30") {
		t.Errorf("header should contain formatted duration, got: %q", got)
	}
	if !strings.Contains(got, "Subtitle tracks: 4") {
		t.Errorf("header should contain track count, got: %q", got)
	}
	if !strings.Contains(got, "Size:") {
		t.Errorf("header should contain size, got: %q", got)
	}
	// Verify the size is human-readable (should be "8.2 GB" for 8800000000 bytes)
	if !strings.Contains(got, "GB") {
		t.Errorf("header size should be in GB, got: %q", got)
	}
}

func TestFormatTrackListing_MixedTracks(t *testing.T) {
	info := MKVInfo{
		Info: FileInfo{
			FileName:      "test.mkv",
			FileSize:      1073741824, // 1 GB
			Duration:      90 * time.Minute,
			SubtitleCount: 3,
			TextSubCount:  2,
		},
		Tracks: []SubtitleTrack{
			{Index: 1, LanguageName: "English", FormatType: "SRT", CodecID: "S_TEXT/UTF8", IsDefault: true, IsText: true},
			{Index: 2, LanguageName: "中文", FormatType: "ASS", CodecID: "S_TEXT/ASS", Name: "简体中文字幕", IsDefault: false, IsText: true},
			{Index: 3, LanguageName: "English", FormatType: "PGS", CodecID: "S_HDMV/PGS", IsDefault: false, IsText: false},
		},
	}

	got := FormatTrackListing(info)

	// Should contain file header
	if !strings.Contains(got, "File: test.mkv") {
		t.Errorf("listing should contain file header, got:\n%s", got)
	}

	// Should contain track lines
	if !strings.Contains(got, "[1] English (SRT)") {
		t.Errorf("listing should contain track 1, got:\n%s", got)
	}
	if !strings.Contains(got, "[2] 中文 (ASS)") {
		t.Errorf("listing should contain track 2, got:\n%s", got)
	}
	if !strings.Contains(got, "not extractable (image subtitle)") {
		t.Errorf("listing should contain image marker for track 3, got:\n%s", got)
	}

	// Should show [default] since not all tracks are default (mixed: true, false, false)
	if !strings.Contains(got, "[default]") {
		t.Errorf("listing should show [default] for mixed default flags, got:\n%s", got)
	}

	// Should NOT contain the Chinese "no text subs" message since there ARE text subs
	if strings.Contains(got, "未找到可提取的文本字幕轨道") {
		t.Errorf("listing should NOT contain no-text-subs message when text subs exist, got:\n%s", got)
	}
}

func TestFormatTrackListing_OnlyImageTracks(t *testing.T) {
	info := MKVInfo{
		Info: FileInfo{
			FileName:      "movie.mkv",
			FileSize:      5000000000,
			Duration:      120 * time.Minute,
			SubtitleCount: 2,
			TextSubCount:  0,
		},
		Tracks: []SubtitleTrack{
			{Index: 1, LanguageName: "English", FormatType: "PGS", CodecID: "S_HDMV/PGS", IsDefault: true, IsText: false},
			{Index: 2, LanguageName: "中文", FormatType: "VobSub", CodecID: "S_VOBSUB", IsDefault: true, IsText: false},
		},
	}

	got := FormatTrackListing(info)

	// Should contain the Chinese message for no text subtitles
	if !strings.Contains(got, "未找到可提取的文本字幕轨道") {
		t.Errorf("listing with only image tracks should contain Chinese message, got:\n%s", got)
	}

	// Track lines should still be present
	if !strings.Contains(got, "[1] English (PGS)") {
		t.Errorf("listing should still show image track 1, got:\n%s", got)
	}
}

func TestFormatTrackListing_NoTracks(t *testing.T) {
	info := MKVInfo{
		Info: FileInfo{
			FileName:      "empty.mkv",
			FileSize:      100000,
			Duration:      5 * time.Minute,
			SubtitleCount: 0,
			TextSubCount:  0,
		},
		Tracks: nil,
	}

	got := FormatTrackListing(info)

	if !strings.Contains(got, "No subtitle tracks found in this file.") {
		t.Errorf("listing with no tracks should contain no-tracks message, got:\n%s", got)
	}
}

func TestShouldShowDefault_AllTrue(t *testing.T) {
	tracks := []SubtitleTrack{
		{IsDefault: true},
		{IsDefault: true},
		{IsDefault: true},
	}
	if ShouldShowDefault(tracks) {
		t.Error("ShouldShowDefault() should return false when all tracks are default")
	}
}

func TestShouldShowDefault_Mixed(t *testing.T) {
	tracks := []SubtitleTrack{
		{IsDefault: true},
		{IsDefault: false},
		{IsDefault: true},
	}
	if !ShouldShowDefault(tracks) {
		t.Error("ShouldShowDefault() should return true when tracks have mixed default flags")
	}
}

func TestShouldShowDefault_AllFalse(t *testing.T) {
	tracks := []SubtitleTrack{
		{IsDefault: false},
		{IsDefault: false},
	}
	if !ShouldShowDefault(tracks) {
		t.Error("ShouldShowDefault() should return true when not all tracks are default")
	}
}

func TestShouldShowDefault_Empty(t *testing.T) {
	if ShouldShowDefault(nil) {
		t.Error("ShouldShowDefault() should return false for empty track list")
	}
}

func TestShouldShowDefault_SingleDefault(t *testing.T) {
	tracks := []SubtitleTrack{
		{IsDefault: true},
	}
	if ShouldShowDefault(tracks) {
		t.Error("ShouldShowDefault() should return false when single track is default (all are default)")
	}
}

func TestShouldShowDefault_SingleNonDefault(t *testing.T) {
	tracks := []SubtitleTrack{
		{IsDefault: false},
	}
	if !ShouldShowDefault(tracks) {
		t.Error("ShouldShowDefault() should return true when single track is not default")
	}
}
