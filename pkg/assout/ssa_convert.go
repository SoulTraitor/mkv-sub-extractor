package assout

import (
	"fmt"
	"strconv"
	"strings"
)

// ssaAlignmentToASS maps SSA alignment values to ASS numpad-style alignment values.
//
// SSA:  1-3 = bottom (L/C/R), 5-7 = top (L/C/R), 9-11 = middle (L/C/R)
// ASS:  1-3 = bottom (L/C/R), 4-6 = middle (L/C/R), 7-9 = top (L/C/R)
var ssaAlignmentToASS = map[int]int{
	1:  1,  // bottom-left
	2:  2,  // bottom-center
	3:  3,  // bottom-right
	5:  7,  // top-left
	6:  8,  // top-center
	7:  9,  // top-right
	9:  4,  // middle-left
	10: 5,  // middle-center
	11: 6,  // middle-right
}

// ConvertSSAHeaderToASS converts an SSA V4 header to ASS V4+ format.
//
// The conversion includes:
//  1. ScriptType: v4.00 -> v4.00+ (exact match only, won't match v4.00+)
//  2. [V4 Styles] -> [V4+ Styles] (exact match only)
//  3. Style Format line: 18 SSA fields -> 23 ASS fields with correct mapping
//  4. Style lines: field values remapped (TertiaryColour->OutlineColour, +6 new fields, -AlphaLevel)
//  5. Alignment conversion (SSA numbering -> ASS numpad numbering)
//  6. Events section: Marked -> Layer in Format and Dialogue lines
//
// If the header does not contain "[V4 Styles]" (already V4+ or unknown format),
// it is returned unchanged.
func ConvertSSAHeaderToASS(header string) string {
	// If header doesn't contain [V4 Styles], it's already V4+ or not SSA -- return unchanged
	if !strings.Contains(header, "[V4 Styles]") {
		return header
	}

	lines := strings.Split(header, "\n")
	var result []string
	inStylesSection := false
	inEventsSection := false
	formatProcessed := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 1. Replace ScriptType: v4.00 with v4.00+ (exact match -- don't match v4.00+)
		if strings.HasPrefix(trimmed, "ScriptType:") {
			val := strings.TrimSpace(strings.TrimPrefix(trimmed, "ScriptType:"))
			if val == "v4.00" {
				line = strings.Replace(line, "v4.00", "v4.00+", 1)
			}
			result = append(result, line)
			continue
		}

		// 2. Replace [V4 Styles] with [V4+ Styles]
		if trimmed == "[V4 Styles]" {
			line = strings.Replace(line, "[V4 Styles]", "[V4+ Styles]", 1)
			inStylesSection = true
			inEventsSection = false
			formatProcessed = false
			result = append(result, line)
			continue
		}

		// Track section transitions
		if strings.HasPrefix(trimmed, "[") && trimmed != "[V4 Styles]" {
			if trimmed == "[Events]" {
				inEventsSection = true
			} else {
				inEventsSection = false
			}
			inStylesSection = false
			formatProcessed = false
			result = append(result, line)
			continue
		}

		// 3. Handle Format line in [V4 Styles] section
		if inStylesSection && strings.HasPrefix(trimmed, "Format:") && !formatProcessed {
			result = append(result, convertStyleFormatLine(line))
			formatProcessed = true
			continue
		}

		// 4. Handle Style lines in [V4 Styles] section
		if inStylesSection && strings.HasPrefix(trimmed, "Style:") {
			result = append(result, convertStyleLine(line))
			continue
		}

		// 6. Handle Events section: Marked -> Layer
		if inEventsSection && strings.HasPrefix(trimmed, "Format:") {
			line = convertEventsFormatLine(line)
			result = append(result, line)
			continue
		}
		if inEventsSection && strings.HasPrefix(trimmed, "Dialogue:") {
			line = convertDialogueLine(line)
			result = append(result, line)
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// convertStyleFormatLine replaces the SSA V4 Style Format line with the ASS V4+ equivalent.
func convertStyleFormatLine(line string) string {
	// Preserve any leading whitespace from the original line
	prefix := ""
	for _, ch := range line {
		if ch == ' ' || ch == '\t' {
			prefix += string(ch)
		} else {
			break
		}
	}
	return prefix + "Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, " +
		"OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, " +
		"ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, " +
		"Alignment, MarginL, MarginR, MarginV, Encoding"
}

// convertStyleLine converts an SSA V4 Style line to ASS V4+ format.
//
// SSA fields (18): Name, Fontname, Fontsize, PrimaryColour, SecondaryColour,
//
//	TertiaryColour, BackColour, Bold, Italic, BorderStyle, Outline, Shadow,
//	Alignment, MarginL, MarginR, MarginV, AlphaLevel, Encoding
//
// ASS fields (23): Name, Fontname, Fontsize, PrimaryColour, SecondaryColour,
//
//	OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut,
//	ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow,
//	Alignment, MarginL, MarginR, MarginV, Encoding
func convertStyleLine(line string) string {
	trimmed := strings.TrimSpace(line)
	afterStyle := strings.TrimPrefix(trimmed, "Style:")
	afterStyle = strings.TrimSpace(afterStyle)

	fields := strings.Split(afterStyle, ",")
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}

	if len(fields) < 18 {
		// Not enough fields for SSA V4 format, return unchanged
		return line
	}

	// SSA field indices (0-based):
	// 0=Name, 1=Fontname, 2=Fontsize, 3=PrimaryColour, 4=SecondaryColour,
	// 5=TertiaryColour, 6=BackColour, 7=Bold, 8=Italic, 9=BorderStyle,
	// 10=Outline, 11=Shadow, 12=Alignment, 13=MarginL, 14=MarginR,
	// 15=MarginV, 16=AlphaLevel, 17=Encoding

	name := fields[0]
	fontname := fields[1]
	fontsize := fields[2]
	primaryColour := fields[3]
	secondaryColour := fields[4]
	outlineColour := fields[5] // TertiaryColour -> OutlineColour
	backColour := fields[6]
	bold := fields[7]
	italic := fields[8]
	borderStyle := fields[9]
	outline := fields[10]
	shadow := fields[11]
	alignment := fields[12]
	marginL := fields[13]
	marginR := fields[14]
	marginV := fields[15]
	// fields[16] = AlphaLevel (dropped)
	encoding := fields[17]

	// 5. Convert SSA alignment to ASS alignment
	if alignInt, err := strconv.Atoi(alignment); err == nil {
		if assAlign, ok := ssaAlignmentToASS[alignInt]; ok {
			alignment = strconv.Itoa(assAlign)
		}
	}

	// Preserve leading whitespace
	prefix := ""
	for _, ch := range line {
		if ch == ' ' || ch == '\t' {
			prefix += string(ch)
		} else {
			break
		}
	}

	// Build ASS V4+ style line with 23 fields
	return fmt.Sprintf("%sStyle: %s,%s,%s,%s,%s,%s,%s,%s,%s,0,0,100,100,0,0,%s,%s,%s,%s,%s,%s,%s,%s",
		prefix,
		name, fontname, fontsize,
		primaryColour, secondaryColour, outlineColour, backColour,
		bold, italic,
		// Underline=0, StrikeOut=0, ScaleX=100, ScaleY=100, Spacing=0, Angle=0
		borderStyle, outline, shadow,
		alignment, marginL, marginR, marginV, encoding,
	)
}

// convertEventsFormatLine replaces "Marked" with "Layer" in the Events Format line.
func convertEventsFormatLine(line string) string {
	return strings.Replace(line, "Marked", "Layer", 1)
}

// convertDialogueLine replaces "Marked=N" with just "N" in Dialogue lines.
func convertDialogueLine(line string) string {
	trimmed := strings.TrimSpace(line)
	afterDialogue := strings.TrimPrefix(trimmed, "Dialogue:")
	afterDialogue = strings.TrimSpace(afterDialogue)

	// Check if the first field starts with "Marked="
	if strings.HasPrefix(afterDialogue, "Marked=") {
		// Extract the value after "Marked="
		rest := strings.TrimPrefix(afterDialogue, "Marked=")
		// The value continues until the next comma
		commaIdx := strings.Index(rest, ",")
		if commaIdx >= 0 {
			value := rest[:commaIdx]
			remainder := rest[commaIdx:]

			// Preserve leading whitespace
			prefix := ""
			for _, ch := range line {
				if ch == ' ' || ch == '\t' {
					prefix += string(ch)
				} else {
					break
				}
			}
			return prefix + "Dialogue: " + value + remainder
		}
	}

	return line
}
