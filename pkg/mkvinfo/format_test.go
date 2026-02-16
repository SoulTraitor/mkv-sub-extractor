package mkvinfo

import (
	"testing"
	"time"
)

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"zero", 0, "0 B"},
		{"small bytes", 500, "500 B"},
		{"exactly 1 KB", 1024, "1.0 KB"},
		{"exactly 1 GB", 1073741824, "1.0 GB"},
		{"about 1.2 GB", 1288490189, "1.2 GB"},
		{"negative", -1, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatFileSize(tt.bytes)
			if got != tt.want {
				t.Errorf("FormatFileSize(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"zero", 0, "Unknown"},
		{"negative", -1 * time.Second, "Unknown"},
		{"5 min 12 sec", 5*time.Minute + 12*time.Second, "0:05:12"},
		{"1 hr 42 min 30 sec", 1*time.Hour + 42*time.Minute + 30*time.Second, "1:42:30"},
		{"over 24 hours", 25 * time.Hour, "25:00:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}
