package subtitle

import (
	"fmt"
	"regexp"
	"strings"
)

// Package-level compiled regexes for SRT tag conversion (compiled once, not per-call).
var (
	// Opening tags (case insensitive)
	reBoldOpen      = regexp.MustCompile(`(?i)<b>`)
	reBoldClose     = regexp.MustCompile(`(?i)</b>`)
	reItalicOpen    = regexp.MustCompile(`(?i)<i>`)
	reItalicClose   = regexp.MustCompile(`(?i)</i>`)
	reUnderlineOpen = regexp.MustCompile(`(?i)<u>`)
	reUnderlineClose = regexp.MustCompile(`(?i)</u>`)

	// Font color with hex value: <font color="#RRGGBB"> or <font color='#RRGGBB'>
	reFontColorHex = regexp.MustCompile(`(?i)<font\s+color=["']#([0-9a-fA-F]{6})["']>`)

	// Font color with named color: <font color="red"> or <font color='red'>
	reFontColorNamed = regexp.MustCompile(`(?i)<font\s+color=["']([a-zA-Z]+)["']>`)

	// Closing font tag
	reFontClose = regexp.MustCompile(`(?i)</font>`)

	// Strip remaining unrecognized HTML tags
	reStripTags = regexp.MustCompile(`<[^>]*>`)
)

// namedColors maps common SRT named colors to their RGB hex values.
var namedColors = map[string]string{
	"red":     "FF0000",
	"blue":    "0000FF",
	"green":   "008000",
	"yellow":  "FFFF00",
	"white":   "FFFFFF",
	"cyan":    "00FFFF",
	"magenta": "FF00FF",
}

// rgbToBGR reverses a 6-character RGB hex string to BGR order for ASS format.
func rgbToBGR(rgb string) string {
	return rgb[4:6] + rgb[2:4] + rgb[0:2]
}

// ConvertSRTTagsToASS converts SRT HTML-style tags to ASS override tags.
//
// Supported conversions:
//   - <b>...</b>  -> {\b1}...{\b0}
//   - <i>...</i>  -> {\i1}...{\i0}
//   - <u>...</u>  -> {\u1}...{\u0}
//   - <font color="#RRGGBB">...</font> -> {\c&HBBGGRR&}...{\c}
//   - <font color="name">...</font>    -> {\c&HBBGGRR&}...{\c} (named colors)
//
// Unknown HTML tags are stripped. Tags are case insensitive.
// Unclosed tags are left as-is (ASS override tags reset per dialogue line).
func ConvertSRTTagsToASS(text string) string {
	// Bold
	text = reBoldOpen.ReplaceAllString(text, "{\\b1}")
	text = reBoldClose.ReplaceAllString(text, "{\\b0}")

	// Italic
	text = reItalicOpen.ReplaceAllString(text, "{\\i1}")
	text = reItalicClose.ReplaceAllString(text, "{\\i0}")

	// Underline
	text = reUnderlineOpen.ReplaceAllString(text, "{\\u1}")
	text = reUnderlineClose.ReplaceAllString(text, "{\\u0}")

	// Font color with hex value: RGB -> BGR
	text = reFontColorHex.ReplaceAllStringFunc(text, func(match string) string {
		submatch := reFontColorHex.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return ""
		}
		bgr := rgbToBGR(strings.ToUpper(submatch[1]))
		return fmt.Sprintf("{\\c&H%s&}", bgr)
	})

	// Font color with named color
	text = reFontColorNamed.ReplaceAllStringFunc(text, func(match string) string {
		submatch := reFontColorNamed.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return ""
		}
		name := strings.ToLower(submatch[1])
		rgb, ok := namedColors[name]
		if !ok {
			return "" // Unknown color name: strip the tag
		}
		bgr := rgbToBGR(rgb)
		return fmt.Sprintf("{\\c&H%s&}", bgr)
	})

	// Close font tag -> reset color
	text = reFontClose.ReplaceAllString(text, "{\\c}")

	// Strip any remaining unrecognized HTML tags
	text = reStripTags.ReplaceAllString(text, "")

	return text
}
