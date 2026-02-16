package mkvinfo

import "testing"

func TestResolveLanguageName(t *testing.T) {
	tests := []struct {
		code     string
		wantName string
	}{
		{"eng", "English"},
		{"und", "Unknown"},
		{"", "Unknown"},
		{"zzz", "zzz"}, // Invalid code returns raw
	}

	for _, tt := range tests {
		t.Run("code_"+tt.code, func(t *testing.T) {
			got := ResolveLanguageName(tt.code)
			if got != tt.wantName {
				t.Errorf("ResolveLanguageName(%q) = %q, want %q", tt.code, got, tt.wantName)
			}
		})
	}
}

func TestResolveLanguageName_CJK(t *testing.T) {
	// Chinese (bibliographic code "chi" -> BCP 47 "zh")
	chiName := ResolveLanguageName("chi")
	if chiName == "" || chiName == "chi" || chiName == "Unknown" {
		t.Errorf("ResolveLanguageName(\"chi\") = %q, want a resolved Chinese name", chiName)
	}
	t.Logf("chi -> %q", chiName)

	// Japanese (bibliographic code "jpn" -> BCP 47 "ja")
	jpnName := ResolveLanguageName("jpn")
	if jpnName == "" || jpnName == "jpn" || jpnName == "Unknown" {
		t.Errorf("ResolveLanguageName(\"jpn\") = %q, want a resolved Japanese name", jpnName)
	}
	t.Logf("jpn -> %q", jpnName)
}

func TestResolveLanguageName_BibliographicCodes(t *testing.T) {
	// French: bibliographic "fre" (vs terminology "fra")
	freName := ResolveLanguageName("fre")
	if freName == "" || freName == "fre" || freName == "Unknown" {
		t.Errorf("ResolveLanguageName(\"fre\") = %q, want a resolved French name", freName)
	}
	t.Logf("fre -> %q", freName)

	// German: bibliographic "ger" (vs terminology "deu")
	gerName := ResolveLanguageName("ger")
	if gerName == "" || gerName == "ger" || gerName == "Unknown" {
		t.Errorf("ResolveLanguageName(\"ger\") = %q, want a resolved German name", gerName)
	}
	t.Logf("ger -> %q", gerName)
}
