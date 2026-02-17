package assout

import (
	"bytes"
	"strings"
	"testing"

	"mkv-sub-extractor/pkg/subtitle"
)

// simpleV4PlusHeader is a minimal ASS header with Script Info and V4+ Styles,
// but without an [Events] section.
const simpleV4PlusHeader = `[Script Info]
Title: Test
ScriptType: v4.00+

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Default,Arial,20,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,0,0,0,0,100,100,0,0,1,2,1,2,10,10,10,1
`

func makeEvents(starts ...uint64) []subtitle.SubtitleEvent {
	var events []subtitle.SubtitleEvent
	for i, s := range starts {
		events = append(events, subtitle.SubtitleEvent{
			Start:     s,
			End:       s + 1_000_000_000,
			Layer:     0,
			Style:     "Default",
			Name:      "",
			MarginL:   "0",
			MarginR:   "0",
			MarginV:   "0",
			Effect:    "",
			Text:      "Line " + string(rune('A'+i)),
			ReadOrder: i,
		})
	}
	return events
}

func TestWriteASSPassthrough_SimpleHeaderNoEvents(t *testing.T) {
	var buf bytes.Buffer
	events := makeEvents(1_000_000_000, 3_000_000_000)

	err := WriteASSPassthrough(&buf, []byte(simpleV4PlusHeader), "S_TEXT/ASS", events)
	if err != nil {
		t.Fatalf("WriteASSPassthrough returned error: %v", err)
	}

	output := buf.String()

	// Should contain [Script Info]
	if !strings.Contains(output, "[Script Info]") {
		t.Error("Output missing [Script Info]")
	}

	// Should contain [V4+ Styles]
	if !strings.Contains(output, "[V4+ Styles]") {
		t.Error("Output missing [V4+ Styles]")
	}

	// Should contain [Events] section (appended)
	if !strings.Contains(output, "[Events]") {
		t.Error("Output missing [Events]")
	}

	// Should contain Format line
	if !strings.Contains(output, "Format: Layer, Start, End") {
		t.Error("Output missing Format line")
	}

	// Should contain 2 Dialogue lines
	dialogueCount := strings.Count(output, "Dialogue:")
	if dialogueCount != 2 {
		t.Errorf("Expected 2 Dialogue lines, got %d", dialogueCount)
	}
}

func TestWriteASSPassthrough_HeaderWithEventsAndFormat(t *testing.T) {
	// CodecPrivate already contains [Events] and Format line
	header := simpleV4PlusHeader + `
[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
`
	var buf bytes.Buffer
	events := makeEvents(1_000_000_000)

	err := WriteASSPassthrough(&buf, []byte(header), "S_TEXT/ASS", events)
	if err != nil {
		t.Fatalf("WriteASSPassthrough returned error: %v", err)
	}

	output := buf.String()

	// Should NOT have duplicate [Events] section
	eventsCount := strings.Count(output, "[Events]")
	if eventsCount != 1 {
		t.Errorf("Expected 1 [Events] section, got %d", eventsCount)
	}

	// Should NOT have duplicate Format line
	formatCount := strings.Count(output, "Format: Layer, Start, End")
	if formatCount != 1 {
		t.Errorf("Expected 1 Format line for events, got %d", formatCount)
	}
}

func TestWriteASSPassthrough_HeaderWithEventsNoFormat(t *testing.T) {
	// CodecPrivate has [Events] but no Format line
	header := simpleV4PlusHeader + `
[Events]
`
	var buf bytes.Buffer
	events := makeEvents(1_000_000_000)

	err := WriteASSPassthrough(&buf, []byte(header), "S_TEXT/ASS", events)
	if err != nil {
		t.Fatalf("WriteASSPassthrough returned error: %v", err)
	}

	output := buf.String()

	// Should have exactly 1 [Events] section
	eventsCount := strings.Count(output, "[Events]")
	if eventsCount != 1 {
		t.Errorf("Expected 1 [Events] section, got %d", eventsCount)
	}

	// Should have Format line added
	if !strings.Contains(output, "Format: Layer, Start, End") {
		t.Error("Output missing Format line when [Events] had none")
	}
}

func TestWriteASSPassthrough_EventsSortedByTimestamp(t *testing.T) {
	var buf bytes.Buffer
	// Create events in reverse order
	events := []subtitle.SubtitleEvent{
		{Start: 5_000_000_000, End: 6_000_000_000, Layer: 0, Style: "Default", Name: "", MarginL: "0", MarginR: "0", MarginV: "0", Effect: "", Text: "Third", ReadOrder: 2},
		{Start: 1_000_000_000, End: 2_000_000_000, Layer: 0, Style: "Default", Name: "", MarginL: "0", MarginR: "0", MarginV: "0", Effect: "", Text: "First", ReadOrder: 0},
		{Start: 3_000_000_000, End: 4_000_000_000, Layer: 0, Style: "Default", Name: "", MarginL: "0", MarginR: "0", MarginV: "0", Effect: "", Text: "Second", ReadOrder: 1},
	}

	err := WriteASSPassthrough(&buf, []byte(simpleV4PlusHeader), "S_TEXT/ASS", events)
	if err != nil {
		t.Fatalf("WriteASSPassthrough returned error: %v", err)
	}

	output := buf.String()

	// Find Dialogue lines in order
	lines := strings.Split(output, "\n")
	var dialogues []string
	for _, line := range lines {
		if strings.Contains(line, "Dialogue:") {
			dialogues = append(dialogues, line)
		}
	}

	if len(dialogues) != 3 {
		t.Fatalf("Expected 3 Dialogue lines, got %d", len(dialogues))
	}

	if !strings.Contains(dialogues[0], "First") {
		t.Errorf("First dialogue should be 'First', got: %s", dialogues[0])
	}
	if !strings.Contains(dialogues[1], "Second") {
		t.Errorf("Second dialogue should be 'Second', got: %s", dialogues[1])
	}
	if !strings.Contains(dialogues[2], "Third") {
		t.Errorf("Third dialogue should be 'Third', got: %s", dialogues[2])
	}
}

func TestWriteASSPassthrough_SSACodecTriggersConversion(t *testing.T) {
	ssaHeader := `[Script Info]
Title: Test
ScriptType: v4.00

[V4 Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, TertiaryColour, BackColour, Bold, Italic, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, AlphaLevel, Encoding
Style: Default,Arial,20,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,-1,0,1,2,1,2,10,10,10,0,1
`

	var buf bytes.Buffer
	events := makeEvents(1_000_000_000)

	err := WriteASSPassthrough(&buf, []byte(ssaHeader), "S_TEXT/SSA", events)
	if err != nil {
		t.Fatalf("WriteASSPassthrough returned error: %v", err)
	}

	output := buf.String()

	// Should have V4+ Styles (converted from V4)
	if !strings.Contains(output, "[V4+ Styles]") {
		t.Error("SSA header not converted to V4+ Styles")
	}
	if strings.Contains(output, "[V4 Styles]") {
		t.Error("V4 Styles still present after SSA conversion")
	}

	// Should have v4.00+
	if !strings.Contains(output, "v4.00+") {
		t.Error("ScriptType not converted to v4.00+")
	}
}

func TestWriteASSPassthrough_ASSCodecWithV4StylesConverts(t *testing.T) {
	// Some MKV files have CodecID "S_TEXT/ASS" but CodecPrivate with [V4 Styles].
	// The writer must convert these to V4+ regardless of codec ID.
	ssaHeader := `[Script Info]
Title: Test
ScriptType: v4.00

[V4 Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, TertiaryColour, BackColour, Bold, Italic, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, AlphaLevel, Encoding
Style: Font1OutlineShadow,Arial,20,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,-1,0,1,2,1,2,10,10,10,0,1
`

	var buf bytes.Buffer
	events := []subtitle.SubtitleEvent{{
		Start: 1_000_000_000, End: 2_000_000_000,
		Style: "Font1OutlineShadow", Text: "Hello",
		MarginL: "0", MarginR: "0", MarginV: "0",
	}}

	err := WriteASSPassthrough(&buf, []byte(ssaHeader), "S_TEXT/ASS", events)
	if err != nil {
		t.Fatalf("WriteASSPassthrough returned error: %v", err)
	}

	output := buf.String()

	// Should have V4+ Styles (converted from V4 despite ASS codec ID)
	if !strings.Contains(output, "[V4+ Styles]") {
		t.Error("V4 Styles not converted to V4+ for S_TEXT/ASS codec")
	}
	if strings.Contains(output, "[V4 Styles]") {
		t.Error("V4 Styles still present -- conversion not applied for S_TEXT/ASS codec")
	}

	// Style name should be preserved in both header and dialogue
	if !strings.Contains(output, "Style: Font1OutlineShadow,") {
		t.Error("Style definition not preserved after conversion")
	}
	if !strings.Contains(output, "Dialogue: 0,0:00:01.00,0:00:02.00,Font1OutlineShadow,") {
		t.Error("Dialogue line missing or style name corrupted")
	}
}

func TestWriteASSPassthrough_ASSCodecPreservesVerbatim(t *testing.T) {
	var buf bytes.Buffer
	events := makeEvents(1_000_000_000)

	err := WriteASSPassthrough(&buf, []byte(simpleV4PlusHeader), "S_TEXT/ASS", events)
	if err != nil {
		t.Fatalf("WriteASSPassthrough returned error: %v", err)
	}

	output := buf.String()

	// Should contain the original title unchanged
	if !strings.Contains(output, "Title: Test") {
		t.Error("ASS header not preserved verbatim -- Title missing")
	}

	// Should contain the original style unchanged
	if !strings.Contains(output, "Style: Default,Arial,20") {
		t.Error("ASS header not preserved verbatim -- Style changed")
	}
}

func TestWriteASSPassthrough_EmptyEvents(t *testing.T) {
	var buf bytes.Buffer

	err := WriteASSPassthrough(&buf, []byte(simpleV4PlusHeader), "S_TEXT/ASS", nil)
	if err != nil {
		t.Fatalf("WriteASSPassthrough returned error: %v", err)
	}

	output := buf.String()

	// Should have header and [Events] but no Dialogue lines
	if !strings.Contains(output, "[Script Info]") {
		t.Error("Missing [Script Info]")
	}
	if !strings.Contains(output, "[Events]") {
		t.Error("Missing [Events]")
	}

	dialogueCount := strings.Count(output, "Dialogue:")
	if dialogueCount != 0 {
		t.Errorf("Expected 0 Dialogue lines for empty events, got %d", dialogueCount)
	}
}

func TestWriteASSPassthrough_CRLFLineEndings(t *testing.T) {
	var buf bytes.Buffer
	events := makeEvents(1_000_000_000)

	err := WriteASSPassthrough(&buf, []byte(simpleV4PlusHeader), "S_TEXT/ASS", events)
	if err != nil {
		t.Fatalf("WriteASSPassthrough returned error: %v", err)
	}

	output := buf.String()

	// All lines should end with \r\n
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			// Last empty line from trailing newline split is OK
			continue
		}
		if !strings.HasSuffix(line, "\r") {
			t.Errorf("Line %d does not have CRLF ending: %q", i+1, line)
		}
	}
}

func TestWriteASSPassthrough_ReadOrderSecondarySort(t *testing.T) {
	var buf bytes.Buffer
	// Events with same start time but different ReadOrder
	events := []subtitle.SubtitleEvent{
		{Start: 1_000_000_000, End: 2_000_000_000, Layer: 0, Style: "Default", Name: "", MarginL: "0", MarginR: "0", MarginV: "0", Effect: "", Text: "ReadOrder2", ReadOrder: 2},
		{Start: 1_000_000_000, End: 2_000_000_000, Layer: 0, Style: "Default", Name: "", MarginL: "0", MarginR: "0", MarginV: "0", Effect: "", Text: "ReadOrder0", ReadOrder: 0},
		{Start: 1_000_000_000, End: 2_000_000_000, Layer: 0, Style: "Default", Name: "", MarginL: "0", MarginR: "0", MarginV: "0", Effect: "", Text: "ReadOrder1", ReadOrder: 1},
	}

	err := WriteASSPassthrough(&buf, []byte(simpleV4PlusHeader), "S_TEXT/ASS", events)
	if err != nil {
		t.Fatalf("WriteASSPassthrough returned error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(output, "\n")
	var dialogues []string
	for _, line := range lines {
		if strings.Contains(line, "Dialogue:") {
			dialogues = append(dialogues, line)
		}
	}

	if len(dialogues) != 3 {
		t.Fatalf("Expected 3 Dialogue lines, got %d", len(dialogues))
	}

	if !strings.Contains(dialogues[0], "ReadOrder0") {
		t.Errorf("First dialogue should be ReadOrder0, got: %s", dialogues[0])
	}
	if !strings.Contains(dialogues[1], "ReadOrder1") {
		t.Errorf("Second dialogue should be ReadOrder1, got: %s", dialogues[1])
	}
	if !strings.Contains(dialogues[2], "ReadOrder2") {
		t.Errorf("Third dialogue should be ReadOrder2, got: %s", dialogues[2])
	}
}

func TestWriteASSPassthrough_DoesNotMutateInputSlice(t *testing.T) {
	events := []subtitle.SubtitleEvent{
		{Start: 3_000_000_000, End: 4_000_000_000, Layer: 0, Style: "Default", Name: "", MarginL: "0", MarginR: "0", MarginV: "0", Effect: "", Text: "B", ReadOrder: 1},
		{Start: 1_000_000_000, End: 2_000_000_000, Layer: 0, Style: "Default", Name: "", MarginL: "0", MarginR: "0", MarginV: "0", Effect: "", Text: "A", ReadOrder: 0},
	}

	var buf bytes.Buffer
	err := WriteASSPassthrough(&buf, []byte(simpleV4PlusHeader), "S_TEXT/ASS", events)
	if err != nil {
		t.Fatalf("WriteASSPassthrough returned error: %v", err)
	}

	// Original slice should be unchanged
	if events[0].Text != "B" {
		t.Errorf("First event mutated: Text = %q, want B", events[0].Text)
	}
	if events[1].Text != "A" {
		t.Errorf("Second event mutated: Text = %q, want A", events[1].Text)
	}
}

func TestWriteASSPassthrough_DialogueFieldsPreserved(t *testing.T) {
	var buf bytes.Buffer
	events := []subtitle.SubtitleEvent{
		{
			Start:     1_000_000_000,
			End:       2_000_000_000,
			Layer:     3,
			Style:     "TopStyle",
			Name:      "Speaker",
			MarginL:   "20",
			MarginR:   "30",
			MarginV:   "40",
			Effect:    "Fade",
			Text:      "Hello, world!",
			ReadOrder: 0,
		},
	}

	err := WriteASSPassthrough(&buf, []byte(simpleV4PlusHeader), "S_TEXT/ASS", events)
	if err != nil {
		t.Fatalf("WriteASSPassthrough returned error: %v", err)
	}

	output := buf.String()

	// Check dialogue line has correct fields
	expectedParts := []string{
		"Dialogue: 3,",          // Layer=3
		"0:00:01.00,0:00:02.00", // timestamps
		"TopStyle",              // Style
		"Speaker",               // Name
		"20,30,40",              // Margins
		"Fade",                  // Effect
		"Hello, world!",         // Text (with comma preserved)
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Output missing %q", part)
		}
	}
}

func TestWriteASSPassthrough_CaseInsensitiveEventsDetection(t *testing.T) {
	// Some files might have [events] in different case
	header := simpleV4PlusHeader + "\n[events]\nFormat: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n"

	var buf bytes.Buffer
	err := WriteASSPassthrough(&buf, []byte(header), "S_TEXT/ASS", nil)
	if err != nil {
		t.Fatalf("WriteASSPassthrough returned error: %v", err)
	}

	output := buf.String()

	// Should NOT add another [Events] section (case-insensitive detection)
	eventsCount := 0
	for _, line := range strings.Split(strings.ToLower(output), "\n") {
		if strings.Contains(line, "[events]") {
			eventsCount++
		}
	}
	if eventsCount != 1 {
		t.Errorf("Expected 1 [events] section (case-insensitive), got %d", eventsCount)
	}
}
