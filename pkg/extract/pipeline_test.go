package extract

import (
	"os"
	"path/filepath"
	"testing"

	"mkv-sub-extractor/pkg/mkvinfo"
	"mkv-sub-extractor/pkg/subtitle"
)

func TestExtractTrackToASS_Signature(t *testing.T) {
	// Verify the function signature compiles correctly.
	// ExtractTrackToASS(mkvPath string, track mkvinfo.SubtitleTrack, outputDir string) (string, error)
	var fn func(string, mkvinfo.SubtitleTrack, string) (string, error)
	fn = ExtractTrackToASS
	_ = fn
}

func TestExtractTracksToASS_Signature(t *testing.T) {
	// Verify the batch function signature compiles correctly.
	// ExtractTracksToASS(mkvPath string, tracks []mkvinfo.SubtitleTrack, outputDir string) ([]string, error)
	var fn func(string, []mkvinfo.SubtitleTrack, string) ([]string, error)
	fn = ExtractTracksToASS
	_ = fn
}

func TestExtractTrackToASS_NonExistentFile(t *testing.T) {
	track := mkvinfo.SubtitleTrack{
		Number:  3,
		CodecID: "S_TEXT/ASS",
	}

	_, err := ExtractTrackToASS("/nonexistent/path/video.mkv", track, "")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}

func TestExtractTrackToASS_InvalidMKVFile(t *testing.T) {
	// Create a temporary file with invalid MKV content
	tmpDir := t.TempDir()
	fakeMKV := filepath.Join(tmpDir, "notreal.mkv")
	if err := os.WriteFile(fakeMKV, []byte("this is not an MKV file"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	track := mkvinfo.SubtitleTrack{
		Number:  3,
		CodecID: "S_TEXT/ASS",
	}

	_, err := ExtractTrackToASS(fakeMKV, track, "")
	if err == nil {
		t.Fatal("expected error for invalid MKV file, got nil")
	}
}

func TestExtractTracksToASS_EmptyTracks(t *testing.T) {
	// Empty tracks list should return nil, nil
	paths, err := ExtractTracksToASS("/nonexistent/path/video.mkv", nil, "")
	if err != nil {
		t.Fatalf("expected nil error for empty tracks, got: %v", err)
	}
	if paths != nil {
		t.Fatalf("expected nil paths for empty tracks, got: %v", paths)
	}
}

func TestExtractTracksToASS_EmptySlice(t *testing.T) {
	paths, err := ExtractTracksToASS("/nonexistent/path/video.mkv", []mkvinfo.SubtitleTrack{}, "")
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got: %v", err)
	}
	if paths != nil {
		t.Fatalf("expected nil paths for empty slice, got: %v", paths)
	}
}

func TestPacketsToEvents_ASSCodec(t *testing.T) {
	packets := []RawSubtitlePacket{
		{
			StartTime: 1_000_000_000,
			EndTime:   2_000_000_000,
			Data:      []byte("0,1,Default,,0,0,0,,Hello world"),
		},
		{
			StartTime: 3_000_000_000,
			EndTime:   4_000_000_000,
			Data:      []byte("1,0,Sign,,10,20,30,Banner,{\\pos(320,50)}Sign text"),
		},
	}

	events, err := packetsToEvents(packets, "S_TEXT/ASS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	// Check first event
	ev := events[0]
	if ev.Start != 1_000_000_000 {
		t.Errorf("event[0].Start = %d, want 1000000000", ev.Start)
	}
	if ev.End != 2_000_000_000 {
		t.Errorf("event[0].End = %d, want 2000000000", ev.End)
	}
	if ev.Layer != 1 {
		t.Errorf("event[0].Layer = %d, want 1", ev.Layer)
	}
	if ev.ReadOrder != 0 {
		t.Errorf("event[0].ReadOrder = %d, want 0", ev.ReadOrder)
	}
	if ev.Style != "Default" {
		t.Errorf("event[0].Style = %q, want %q", ev.Style, "Default")
	}
	if ev.Text != "Hello world" {
		t.Errorf("event[0].Text = %q, want %q", ev.Text, "Hello world")
	}

	// Check second event
	ev2 := events[1]
	if ev2.Layer != 0 {
		t.Errorf("event[1].Layer = %d, want 0", ev2.Layer)
	}
	if ev2.ReadOrder != 1 {
		t.Errorf("event[1].ReadOrder = %d, want 1", ev2.ReadOrder)
	}
	if ev2.Style != "Sign" {
		t.Errorf("event[1].Style = %q, want %q", ev2.Style, "Sign")
	}
	if ev2.MarginL != "10" {
		t.Errorf("event[1].MarginL = %q, want %q", ev2.MarginL, "10")
	}
	if ev2.MarginR != "20" {
		t.Errorf("event[1].MarginR = %q, want %q", ev2.MarginR, "20")
	}
	if ev2.MarginV != "30" {
		t.Errorf("event[1].MarginV = %q, want %q", ev2.MarginV, "30")
	}
	if ev2.Effect != "Banner" {
		t.Errorf("event[1].Effect = %q, want %q", ev2.Effect, "Banner")
	}
	if ev2.Text != "{\\pos(320,50)}Sign text" {
		t.Errorf("event[1].Text = %q, want %q", ev2.Text, "{\\pos(320,50)}Sign text")
	}
}

func TestPacketsToEvents_ASSCodec_TextWithCommas(t *testing.T) {
	// Text field can contain commas
	packets := []RawSubtitlePacket{
		{
			StartTime: 1_000_000_000,
			EndTime:   2_000_000_000,
			Data:      []byte("0,0,Default,,0,0,0,,Hello, world, this has, commas"),
		},
	}

	events, err := packetsToEvents(packets, "S_TEXT/ASS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if events[0].Text != "Hello, world, this has, commas" {
		t.Errorf("Text = %q, want %q", events[0].Text, "Hello, world, this has, commas")
	}
}

func TestPacketsToEvents_SSACodec(t *testing.T) {
	// SSA codec uses the same parsing path as ASS
	packets := []RawSubtitlePacket{
		{
			StartTime: 1_000_000_000,
			EndTime:   2_000_000_000,
			Data:      []byte("0,0,Default,,0,0,0,,SSA text"),
		},
	}

	events, err := packetsToEvents(packets, "S_TEXT/SSA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Text != "SSA text" {
		t.Errorf("Text = %q, want %q", events[0].Text, "SSA text")
	}
}

func TestPacketsToEvents_SRTCodec(t *testing.T) {
	packets := []RawSubtitlePacket{
		{
			StartTime: 1_000_000_000,
			EndTime:   2_000_000_000,
			Data:      []byte("Hello world"),
		},
		{
			StartTime: 3_000_000_000,
			EndTime:   4_000_000_000,
			Data:      []byte("Second line"),
		},
	}

	events, err := packetsToEvents(packets, "S_TEXT/UTF8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	ev := events[0]
	if ev.Layer != 0 {
		t.Errorf("SRT event Layer = %d, want 0", ev.Layer)
	}
	if ev.Style != "Default" {
		t.Errorf("SRT event Style = %q, want %q", ev.Style, "Default")
	}
	if ev.MarginL != "0" {
		t.Errorf("SRT event MarginL = %q, want %q", ev.MarginL, "0")
	}
	if ev.Text != "Hello world" {
		t.Errorf("SRT event Text = %q, want %q", ev.Text, "Hello world")
	}
}

func TestPacketsToEvents_UnsupportedCodec(t *testing.T) {
	packets := []RawSubtitlePacket{
		{StartTime: 0, EndTime: 1, Data: []byte("data")},
	}

	_, err := packetsToEvents(packets, "S_HDMV/PGS")
	if err == nil {
		t.Fatal("expected error for unsupported codec, got nil")
	}
}

func TestPacketsToEvents_EmptyPackets(t *testing.T) {
	events, err := packetsToEvents(nil, "S_TEXT/UTF8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected 0 events, got %d", len(events))
	}
}

func TestPacketsToEvents_InvalidASSData(t *testing.T) {
	packets := []RawSubtitlePacket{
		{
			StartTime: 1_000_000_000,
			EndTime:   2_000_000_000,
			Data:      []byte("not enough fields"),
		},
	}

	_, err := packetsToEvents(packets, "S_TEXT/ASS")
	if err == nil {
		t.Fatal("expected error for invalid ASS block data, got nil")
	}
}

// Ensure SubtitleEvent is the correct type from the subtitle package
func TestPacketsToEvents_ReturnsSubtitleEvents(t *testing.T) {
	packets := []RawSubtitlePacket{
		{StartTime: 0, EndTime: 1, Data: []byte("text")},
	}
	events, err := packetsToEvents(packets, "S_TEXT/UTF8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Type assertion at compile time
	var _ []subtitle.SubtitleEvent = events
}
