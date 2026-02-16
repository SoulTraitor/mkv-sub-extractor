package subtitle

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseASSBlockData parses the comma-separated fields from a Matroska ASS/SSA
// subtitle block's Data payload.
//
// Block data format: ReadOrder,Layer,Style,Name,MarginL,MarginR,MarginV,Effect,Text
//
// The Text field (9th) may contain commas, so we split into at most 9 parts.
// Returns the ReadOrder, Layer, and the remaining fields (Style through Text)
// as a single string suitable for inclusion in an ASS Dialogue line.
func ParseASSBlockData(data []byte) (readOrder int, layer int, remaining string, err error) {
	s := string(data)
	if s == "" {
		return 0, 0, "", fmt.Errorf("empty block data")
	}

	// Split into at most 9 parts: ReadOrder, Layer, and 7 remaining fields (Style..Text)
	parts := strings.SplitN(s, ",", 9)
	if len(parts) < 9 {
		return 0, 0, "", fmt.Errorf("expected 9 fields, got %d", len(parts))
	}

	readOrder, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, "", fmt.Errorf("invalid ReadOrder %q: %w", parts[0], err)
	}

	layer, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, "", fmt.Errorf("invalid Layer %q: %w", parts[1], err)
	}

	// Rejoin fields 2-8 (Style through Text) for the Dialogue line
	remaining = strings.Join(parts[2:], ",")

	return readOrder, layer, remaining, nil
}
