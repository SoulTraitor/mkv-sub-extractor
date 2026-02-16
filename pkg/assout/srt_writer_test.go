package assout

import (
	"bytes"
	"strings"
	"testing"

	"mkv-sub-extractor/pkg/subtitle"
)

func TestWriteSRTAsASS_TwoSimpleEvents(t *testing.T) {
	events := []subtitle.SubtitleEvent{
		{
			Start: 1_000_000_000, // 0:00:01.00
			End:   3_000_000_000, // 0:00:03.00
			Text:  "Hello, world!",
		},
		{
			Start: 5_000_000_000, // 0:00:05.00
			End:   7_000_000_000, // 0:00:07.00
			Text:  "Second line",
		},
	}

	var buf bytes.Buffer
	err := WriteSRTAsASS(&buf, events)
	if err != nil {
		t.Fatalf("WriteSRTAsASS returned error: %v", err)
	}

	output := buf.String()

	// Must contain all sections
	if !strings.Contains(output, "[Script Info]") {
		t.Error("output missing [Script Info] section")
	}
	if !strings.Contains(output, "[V4+ Styles]") {
		t.Error("output missing [V4+ Styles] section")
	}
	if !strings.Contains(output, "[Events]") {
		t.Error("output missing [Events] section")
	}

	// Must contain exactly 2 Dialogue lines
	count := strings.Count(output, "Dialogue:")
	if count != 2 {
		t.Errorf("expected 2 Dialogue lines, got %d", count)
	}

	if !strings.Contains(output, "Dialogue: 0,0:00:01.00,0:00:03.00,Default,,0,0,0,,Hello, world!") {
		t.Error("first dialogue line not found or incorrect")
	}
	if !strings.Contains(output, "Dialogue: 0,0:00:05.00,0:00:07.00,Default,,0,0,0,,Second line") {
		t.Error("second dialogue line not found or incorrect")
	}
}

func TestWriteSRTAsASS_EventsSortedByTimestamp(t *testing.T) {
	// Input events in reverse order
	events := []subtitle.SubtitleEvent{
		{
			Start: 10_000_000_000, // 0:00:10.00
			End:   12_000_000_000,
			Text:  "Second",
		},
		{
			Start: 1_000_000_000, // 0:00:01.00
			End:   3_000_000_000,
			Text:  "First",
		},
	}

	var buf bytes.Buffer
	err := WriteSRTAsASS(&buf, events)
	if err != nil {
		t.Fatalf("WriteSRTAsASS returned error: %v", err)
	}

	output := buf.String()

	// Search for Dialogue lines specifically to avoid matching header text
	// ("Second" appears in "SecondaryColour" in the header)
	firstIdx := strings.Index(output, "Dialogue: 0,0:00:01.00,")
	secondIdx := strings.Index(output, "Dialogue: 0,0:00:10.00,")
	if firstIdx == -1 || secondIdx == -1 {
		t.Fatal("dialogue lines not found in output")
	}
	if firstIdx >= secondIdx {
		t.Error("events not sorted by timestamp: 1s event should appear before 10s event")
	}
}

func TestWriteSRTAsASS_SRTTagConversion(t *testing.T) {
	events := []subtitle.SubtitleEvent{
		{
			Start: 0,
			End:   1_000_000_000,
			Text:  "<i>italic text</i>",
		},
	}

	var buf bytes.Buffer
	err := WriteSRTAsASS(&buf, events)
	if err != nil {
		t.Fatalf("WriteSRTAsASS returned error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, `{\i1}italic text{\i0}`) {
		t.Errorf("SRT <i> tags not converted to ASS override tags in output:\n%s", output)
	}
}

func TestWriteSRTAsASS_MultiLineNewlineConversion(t *testing.T) {
	events := []subtitle.SubtitleEvent{
		{
			Start: 0,
			End:   1_000_000_000,
			Text:  "Line one\nLine two",
		},
	}

	var buf bytes.Buffer
	err := WriteSRTAsASS(&buf, events)
	if err != nil {
		t.Fatalf("WriteSRTAsASS returned error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, `Line one\NLine two`) {
		t.Errorf("newline not converted to \\N in output:\n%s", output)
	}
	// Should not contain literal \n in dialogue text
	// Find the Dialogue line and check
	for _, line := range strings.Split(output, "\r\n") {
		if strings.HasPrefix(line, "Dialogue:") {
			if strings.Contains(line, "\n") {
				t.Error("dialogue line contains literal newline instead of \\N")
			}
		}
	}
}

func TestWriteSRTAsASS_EmptyEventsList(t *testing.T) {
	var buf bytes.Buffer
	err := WriteSRTAsASS(&buf, []subtitle.SubtitleEvent{})
	if err != nil {
		t.Fatalf("WriteSRTAsASS returned error: %v", err)
	}

	output := buf.String()

	// Must still have header, events section, and format line
	if !strings.Contains(output, "[Script Info]") {
		t.Error("output missing [Script Info] section")
	}
	if !strings.Contains(output, "[Events]") {
		t.Error("output missing [Events] section")
	}
	if !strings.Contains(output, "Format: Layer, Start, End,") {
		t.Error("output missing events Format line")
	}

	// No Dialogue lines
	if strings.Contains(output, "Dialogue:") {
		t.Error("empty events should produce no Dialogue lines")
	}
}

func TestWriteSRTAsASS_StartsWithScriptInfoAndContainsMicrosoftYaHei(t *testing.T) {
	var buf bytes.Buffer
	err := WriteSRTAsASS(&buf, []subtitle.SubtitleEvent{})
	if err != nil {
		t.Fatalf("WriteSRTAsASS returned error: %v", err)
	}

	output := buf.String()

	if !strings.HasPrefix(output, "[Script Info]") {
		t.Error("output does not start with [Script Info]")
	}
	if !strings.Contains(output, "Microsoft YaHei") {
		t.Error("output does not contain Microsoft YaHei font")
	}
}

func TestWriteSRTAsASS_CRLFLineEndings(t *testing.T) {
	events := []subtitle.SubtitleEvent{
		{
			Start: 0,
			End:   1_000_000_000,
			Text:  "Test",
		},
	}

	var buf bytes.Buffer
	err := WriteSRTAsASS(&buf, events)
	if err != nil {
		t.Fatalf("WriteSRTAsASS returned error: %v", err)
	}

	output := buf.String()

	// Every line should end with \r\n
	// Split on \r\n and check that no individual segment contains a bare \n
	// (except possibly the last empty segment after the final \r\n)
	segments := strings.Split(output, "\r\n")
	for i, seg := range segments {
		if i == len(segments)-1 {
			// Last segment can be empty (trailing \r\n)
			continue
		}
		if strings.Contains(seg, "\n") {
			t.Errorf("segment %d contains bare \\n (not CRLF): %q", i, seg)
		}
	}
}

func TestWriteSRTAsASS_DialogueUsesDefaultStyleAndLayer0(t *testing.T) {
	events := []subtitle.SubtitleEvent{
		{
			Start: 0,
			End:   1_000_000_000,
			Text:  "Test dialogue",
		},
	}

	var buf bytes.Buffer
	err := WriteSRTAsASS(&buf, events)
	if err != nil {
		t.Fatalf("WriteSRTAsASS returned error: %v", err)
	}

	output := buf.String()

	// Dialogue line should start with layer 0 and use Default style
	expected := "Dialogue: 0,0:00:00.00,0:00:01.00,Default,,0,0,0,,Test dialogue\r\n"
	if !strings.Contains(output, expected) {
		t.Errorf("expected dialogue line not found:\nwant: %q\ngot output:\n%s", expected, output)
	}
}

func TestWriteSRTAsASS_DoesNotMutateInputSlice(t *testing.T) {
	events := []subtitle.SubtitleEvent{
		{Start: 10_000_000_000, End: 12_000_000_000, Text: "Second"},
		{Start: 1_000_000_000, End: 3_000_000_000, Text: "First"},
	}

	// Record original order
	origFirst := events[0].Text

	var buf bytes.Buffer
	_ = WriteSRTAsASS(&buf, events)

	// Input slice should not be reordered
	if events[0].Text != origFirst {
		t.Error("WriteSRTAsASS mutated the input events slice")
	}
}
