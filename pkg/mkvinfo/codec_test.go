package mkvinfo

import "testing"

func TestClassifyCodec_KnownCodecs(t *testing.T) {
	tests := []struct {
		codecID    string
		wantFormat string
		wantIsText bool
	}{
		{"S_TEXT/UTF8", "SRT", true},
		{"S_TEXT/SSA", "SSA", true},
		{"S_TEXT/ASS", "ASS", true},
		{"S_TEXT/WEBVTT", "WebVTT", true},
		{"S_TEXT/USF", "USF", true},
		{"S_HDMV/PGS", "PGS", false},
		{"S_VOBSUB", "VobSub", false},
		{"S_DVBSUB", "DVB", false},
		{"S_HDMV/TEXTST", "HDMV Text", true},
		{"S_ARIBSUB", "ARIB", true},
	}

	for _, tt := range tests {
		t.Run(tt.codecID, func(t *testing.T) {
			format, isText := ClassifyCodec(tt.codecID)
			if format != tt.wantFormat {
				t.Errorf("ClassifyCodec(%q) format = %q, want %q", tt.codecID, format, tt.wantFormat)
			}
			if isText != tt.wantIsText {
				t.Errorf("ClassifyCodec(%q) isText = %v, want %v", tt.codecID, isText, tt.wantIsText)
			}
		})
	}
}

func TestClassifyCodec_UnknownCodec(t *testing.T) {
	format, isText := ClassifyCodec("S_UNKNOWN/FOO")
	if format != "S_UNKNOWN/FOO" {
		t.Errorf("ClassifyCodec unknown: format = %q, want %q", format, "S_UNKNOWN/FOO")
	}
	if isText != false {
		t.Errorf("ClassifyCodec unknown: isText = %v, want false", isText)
	}
}

func TestClassifyCodec_EmptyString(t *testing.T) {
	format, isText := ClassifyCodec("")
	if format != "" {
		t.Errorf("ClassifyCodec empty: format = %q, want %q", format, "")
	}
	if isText != false {
		t.Errorf("ClassifyCodec empty: isText = %v, want false", isText)
	}
}
