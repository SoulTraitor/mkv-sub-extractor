package assout

import (
	"strings"
	"testing"
)

// Minimal SSA V4 header for testing
const minimalSSAHeader = `[Script Info]
Title: Test
ScriptType: v4.00
PlayResX: 384
PlayResY: 288

[V4 Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, TertiaryColour, BackColour, Bold, Italic, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, AlphaLevel, Encoding
Style: Default,Arial,20,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,-1,0,1,2,1,2,10,10,10,0,1

[Events]
Format: Marked, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
Dialogue: Marked=0,0:00:01.00,0:00:05.00,Default,,0,0,0,,Hello world`

func TestConvertSSAHeaderToASS_FullConversion(t *testing.T) {
	result := ConvertSSAHeaderToASS(minimalSSAHeader)

	// Verify ScriptType updated
	if !strings.Contains(result, "ScriptType: v4.00+") {
		t.Error("ScriptType not updated to v4.00+")
	}

	// Verify section header updated
	if !strings.Contains(result, "[V4+ Styles]") {
		t.Error("[V4 Styles] not updated to [V4+ Styles]")
	}
	if strings.Contains(result, "[V4 Styles]") {
		t.Error("[V4 Styles] still present after conversion")
	}

	// Verify Format line updated to ASS fields
	if !strings.Contains(result, "OutlineColour") {
		t.Error("Format line missing OutlineColour")
	}
	if strings.Contains(result, "TertiaryColour") {
		t.Error("Format line still contains TertiaryColour")
	}
	if strings.Contains(result, "AlphaLevel") {
		t.Error("Format line still contains AlphaLevel")
	}
	if !strings.Contains(result, "Underline") {
		t.Error("Format line missing Underline")
	}
	if !strings.Contains(result, "StrikeOut") {
		t.Error("Format line missing StrikeOut")
	}

	// Verify Style line has 23 fields
	for _, line := range strings.Split(result, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "Style:") {
			afterStyle := strings.TrimPrefix(strings.TrimSpace(line), "Style:")
			fields := strings.Split(strings.TrimSpace(afterStyle), ",")
			if len(fields) != 23 {
				t.Errorf("Style line has %d fields, want 23; line: %s", len(fields), line)
			}
		}
	}

	// Verify Events Format line has Layer instead of Marked
	for _, line := range strings.Split(result, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Format:") && strings.Contains(trimmed, "Start") && strings.Contains(trimmed, "End") {
			if strings.Contains(trimmed, "Marked") {
				t.Error("Events Format line still contains Marked")
			}
			if !strings.Contains(trimmed, "Layer") {
				t.Error("Events Format line missing Layer")
			}
		}
	}

	// Verify Dialogue line has Layer value, not Marked=N
	for _, line := range strings.Split(result, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "Dialogue:") {
			if strings.Contains(line, "Marked=") {
				t.Errorf("Dialogue line still has Marked=: %s", line)
			}
		}
	}
}

func TestConvertSSAHeaderToASS_ScriptTypeReplacement(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "v4.00 to v4.00+",
			input: "[V4 Styles]\nFormat: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, TertiaryColour, BackColour, Bold, Italic, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, AlphaLevel, Encoding\n\n[Script Info]\nScriptType: v4.00",
			want:  "v4.00+",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertSSAHeaderToASS(tt.input)
			if !strings.Contains(result, "ScriptType: "+tt.want) {
				t.Errorf("ScriptType not converted; got:\n%s", result)
			}
		})
	}
}

func TestConvertSSAHeaderToASS_SectionHeaderReplacement(t *testing.T) {
	input := "[V4 Styles]\nFormat: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, TertiaryColour, BackColour, Bold, Italic, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, AlphaLevel, Encoding"
	result := ConvertSSAHeaderToASS(input)

	if !strings.Contains(result, "[V4+ Styles]") {
		t.Error("[V4 Styles] not replaced with [V4+ Styles]")
	}
	if strings.Contains(result, "[V4 Styles]") {
		t.Error("[V4 Styles] still present")
	}
}

func TestConvertSSAHeaderToASS_StyleFieldRemapping(t *testing.T) {
	input := `[V4 Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, TertiaryColour, BackColour, Bold, Italic, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, AlphaLevel, Encoding
Style: Default,Arial,20,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,-1,0,1,2,1,2,10,10,10,0,1`

	result := ConvertSSAHeaderToASS(input)

	// Find the Style line
	var styleLine string
	for _, line := range strings.Split(result, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "Style:") {
			styleLine = line
			break
		}
	}

	if styleLine == "" {
		t.Fatal("No Style line found in output")
	}

	afterStyle := strings.TrimPrefix(strings.TrimSpace(styleLine), "Style:")
	fields := strings.Split(strings.TrimSpace(afterStyle), ",")

	if len(fields) != 23 {
		t.Fatalf("Expected 23 fields, got %d: %v", len(fields), fields)
	}

	// Check specific field values:
	// fields[0]=Name=Default, [1]=Fontname=Arial, [2]=Fontsize=20
	// fields[3]=PrimaryColour, [4]=SecondaryColour
	// fields[5]=OutlineColour (was TertiaryColour=&H00000000)
	if strings.TrimSpace(fields[5]) != "&H00000000" {
		t.Errorf("OutlineColour = %q, want &H00000000 (from TertiaryColour)", strings.TrimSpace(fields[5]))
	}

	// fields[9]=Underline=0, [10]=StrikeOut=0
	if strings.TrimSpace(fields[9]) != "0" {
		t.Errorf("Underline = %q, want 0", strings.TrimSpace(fields[9]))
	}
	if strings.TrimSpace(fields[10]) != "0" {
		t.Errorf("StrikeOut = %q, want 0", strings.TrimSpace(fields[10]))
	}

	// fields[11]=ScaleX=100, [12]=ScaleY=100, [13]=Spacing=0, [14]=Angle=0
	if strings.TrimSpace(fields[11]) != "100" {
		t.Errorf("ScaleX = %q, want 100", strings.TrimSpace(fields[11]))
	}
	if strings.TrimSpace(fields[12]) != "100" {
		t.Errorf("ScaleY = %q, want 100", strings.TrimSpace(fields[12]))
	}
	if strings.TrimSpace(fields[13]) != "0" {
		t.Errorf("Spacing = %q, want 0", strings.TrimSpace(fields[13]))
	}
	if strings.TrimSpace(fields[14]) != "0" {
		t.Errorf("Angle = %q, want 0", strings.TrimSpace(fields[14]))
	}

	// fields[22]=Encoding=1 (AlphaLevel removed)
	if strings.TrimSpace(fields[22]) != "1" {
		t.Errorf("Encoding = %q, want 1", strings.TrimSpace(fields[22]))
	}
}

func TestConvertSSAHeaderToASS_AlignmentConversion(t *testing.T) {
	tests := []struct {
		name     string
		ssaAlign string
		assAlign string
	}{
		{"bottom-left", "1", "1"},
		{"bottom-center", "2", "2"},
		{"bottom-right", "3", "3"},
		{"top-left", "5", "7"},
		{"top-center", "6", "8"},
		{"top-right", "7", "9"},
		{"middle-left", "9", "4"},
		{"middle-center", "10", "5"},
		{"middle-right", "11", "6"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := `[V4 Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, TertiaryColour, BackColour, Bold, Italic, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, AlphaLevel, Encoding
Style: Default,Arial,20,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,-1,0,1,2,1,` + tt.ssaAlign + `,10,10,10,0,1`

			result := ConvertSSAHeaderToASS(input)

			// Find style line and check alignment field (index 18 in ASS)
			for _, line := range strings.Split(result, "\n") {
				if strings.HasPrefix(strings.TrimSpace(line), "Style:") {
					afterStyle := strings.TrimPrefix(strings.TrimSpace(line), "Style:")
					fields := strings.Split(strings.TrimSpace(afterStyle), ",")
					if len(fields) != 23 {
						t.Fatalf("Expected 23 fields, got %d", len(fields))
					}
					// ASS Alignment is at index 18
					gotAlign := strings.TrimSpace(fields[18])
					if gotAlign != tt.assAlign {
						t.Errorf("Alignment = %q, want %q (SSA %s -> ASS %s)", gotAlign, tt.assAlign, tt.ssaAlign, tt.assAlign)
					}
				}
			}
		})
	}
}

func TestConvertSSAHeaderToASS_MarkedToLayer(t *testing.T) {
	input := `[V4 Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, TertiaryColour, BackColour, Bold, Italic, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, AlphaLevel, Encoding
Style: Default,Arial,20,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,-1,0,1,2,1,2,10,10,10,0,1

[Events]
Format: Marked, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
Dialogue: Marked=0,0:00:01.00,0:00:05.00,Default,,0,0,0,,Hello`

	result := ConvertSSAHeaderToASS(input)

	// Check Format line
	for _, line := range strings.Split(result, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Format:") && strings.Contains(trimmed, "Start") {
			if strings.Contains(trimmed, "Marked") {
				t.Error("Events Format still contains 'Marked'")
			}
			if !strings.Contains(trimmed, "Layer") {
				t.Error("Events Format missing 'Layer'")
			}
		}
	}

	// Check Dialogue line
	for _, line := range strings.Split(result, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Dialogue:") {
			if strings.Contains(trimmed, "Marked=") {
				t.Errorf("Dialogue still contains 'Marked=': %s", trimmed)
			}
			// Should start with "Dialogue: 0,"
			afterDialogue := strings.TrimPrefix(trimmed, "Dialogue:")
			afterDialogue = strings.TrimSpace(afterDialogue)
			if !strings.HasPrefix(afterDialogue, "0,") {
				t.Errorf("Dialogue layer value wrong, got: %s", afterDialogue)
			}
		}
	}
}

func TestConvertSSAHeaderToASS_AlreadyV4Plus(t *testing.T) {
	input := `[Script Info]
Title: Test
ScriptType: v4.00+

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Default,Arial,20,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,0,0,0,0,100,100,0,0,1,2,1,2,10,10,10,1`

	result := ConvertSSAHeaderToASS(input)

	if result != input {
		t.Error("Already V4+ header should be returned unchanged")
		t.Logf("Input:\n%s", input)
		t.Logf("Result:\n%s", result)
	}
}

func TestConvertSSAHeaderToASS_MultipleStyles(t *testing.T) {
	input := `[V4 Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, TertiaryColour, BackColour, Bold, Italic, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, AlphaLevel, Encoding
Style: Default,Arial,20,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,-1,0,1,2,1,2,10,10,10,0,1
Style: Top,Arial,18,&H00FFFF00,&H000000FF,&H00000000,&H80000000,0,0,1,1,0,6,10,10,10,0,1`

	result := ConvertSSAHeaderToASS(input)

	styleCount := 0
	for _, line := range strings.Split(result, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "Style:") {
			styleCount++
			afterStyle := strings.TrimPrefix(strings.TrimSpace(line), "Style:")
			fields := strings.Split(strings.TrimSpace(afterStyle), ",")
			if len(fields) != 23 {
				t.Errorf("Style line has %d fields, want 23: %s", len(fields), line)
			}
		}
	}

	if styleCount != 2 {
		t.Errorf("Expected 2 Style lines, got %d", styleCount)
	}

	// Second style has SSA alignment 6 (top-center) -> ASS alignment 8
	for _, line := range strings.Split(result, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "Style: Top") {
			afterStyle := strings.TrimPrefix(strings.TrimSpace(line), "Style:")
			fields := strings.Split(strings.TrimSpace(afterStyle), ",")
			if len(fields) == 23 {
				align := strings.TrimSpace(fields[18])
				if align != "8" {
					t.Errorf("Top style alignment = %q, want 8 (SSA 6 -> ASS 8)", align)
				}
			}
		}
	}
}

func TestConvertSSAHeaderToASS_DialogueMarkedValues(t *testing.T) {
	// Test with different Marked values
	input := `[V4 Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, TertiaryColour, BackColour, Bold, Italic, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, AlphaLevel, Encoding
Style: Default,Arial,20,&H00FFFFFF,&H000000FF,&H00000000,&H80000000,-1,0,1,2,1,2,10,10,10,0,1

[Events]
Format: Marked, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
Dialogue: Marked=0,0:00:01.00,0:00:05.00,Default,,0,0,0,,Hello
Dialogue: Marked=1,0:00:06.00,0:00:10.00,Default,,0,0,0,,World`

	result := ConvertSSAHeaderToASS(input)

	dialogues := []string{}
	for _, line := range strings.Split(result, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "Dialogue:") {
			dialogues = append(dialogues, strings.TrimSpace(line))
		}
	}

	if len(dialogues) != 2 {
		t.Fatalf("Expected 2 Dialogue lines, got %d", len(dialogues))
	}

	// First dialogue: Marked=0 -> 0
	if !strings.HasPrefix(dialogues[0], "Dialogue: 0,") {
		t.Errorf("First dialogue expected 'Dialogue: 0,...', got: %s", dialogues[0])
	}

	// Second dialogue: Marked=1 -> 1
	if !strings.HasPrefix(dialogues[1], "Dialogue: 1,") {
		t.Errorf("Second dialogue expected 'Dialogue: 1,...', got: %s", dialogues[1])
	}
}
