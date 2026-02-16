package mkvinfo

import (
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// ResolveLanguageName converts an ISO 639-2 language code to a human-readable
// display name, preferring native-script names (e.g. "中文" for Chinese).
//
// Fallback chain: native-script name -> English name -> raw code -> "Unknown".
// Empty string and "und" (undetermined) return "Unknown".
func ResolveLanguageName(code string) string {
	if code == "" || code == "und" {
		return "Unknown"
	}

	base, err := language.ParseBase(code)
	if err != nil {
		return code // Return raw code if unparseable
	}

	tag := language.Make(base.String())

	// Try native-script name first (e.g. "English", "中文", "日本語")
	if name := display.Self.Name(tag); name != "" {
		return name
	}

	// Fallback to English name
	if name := display.English.Languages().Name(tag); name != "" {
		return name
	}

	return code
}
