package mkvinfo

import "time"

// FileInfo holds basic file metadata for display.
type FileInfo struct {
	FileName      string
	FileSize      int64         // bytes
	Duration      time.Duration // from SegmentInfo.Duration nanoseconds
	SubtitleCount int           // total subtitle tracks including image-based
	TextSubCount  int           // extractable text subtitle tracks only
}

// SubtitleTrack holds metadata for a single subtitle track.
type SubtitleTrack struct {
	Number        uint8  // Matroska TrackNumber -- the real track identifier, NOT the array index
	Index         int    // display index, 1-based sequential for our listing
	Language      string // raw ISO 639-2 code from Matroska, e.g. "eng", "jpn"
	LanguageName  string // resolved human-readable name, e.g. "English", "中文"
	FormatType    string // human-readable: "SRT", "ASS", "SSA", "PGS", etc.
	CodecID       string // raw Matroska codec ID, e.g. "S_TEXT/UTF8"
	Name          string // track name from metadata, may be empty
	IsDefault     bool   // FlagDefault
	IsForced      bool   // FlagForced
	IsText        bool   // true for S_TEXT/* codecs
	IsExtractable bool   // true if text-based and extractable
}

// MKVInfo bundles FileInfo and subtitle tracks returned from the main parsing function.
type MKVInfo struct {
	Info   FileInfo
	Tracks []SubtitleTrack
}
