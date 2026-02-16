package output

import (
	"path/filepath"
	"testing"

	"mkv-sub-extractor/pkg/mkvinfo"
)

func TestGenerateOutputPath_Basic(t *testing.T) {
	existing := make(map[string]bool)
	track := mkvinfo.SubtitleTrack{Language: "eng"}

	result := GenerateOutputPath("movie.mkv", track, existing)
	expected := "movie.eng.ass"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestGenerateOutputPath_EmptyLanguage(t *testing.T) {
	existing := make(map[string]bool)
	track := mkvinfo.SubtitleTrack{Language: ""}

	result := GenerateOutputPath("movie.mkv", track, existing)
	expected := "movie.und.ass"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestGenerateOutputPath_UndLanguage(t *testing.T) {
	existing := make(map[string]bool)
	track := mkvinfo.SubtitleTrack{Language: "und"}

	result := GenerateOutputPath("movie.mkv", track, existing)
	expected := "movie.und.ass"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestGenerateOutputPath_WithDirectory(t *testing.T) {
	existing := make(map[string]bool)
	track := mkvinfo.SubtitleTrack{Language: "chi"}

	result := GenerateOutputPath(filepath.Join("path", "to", "movie.mkv"), track, existing)
	expected := filepath.Join("path", "to", "movie.chi.ass")
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestGenerateOutputPath_CollisionWithTrackName(t *testing.T) {
	existing := make(map[string]bool)

	track1 := mkvinfo.SubtitleTrack{Language: "eng"}
	result1 := GenerateOutputPath("movie.mkv", track1, existing)
	if result1 != "movie.eng.ass" {
		t.Errorf("first track: got %q, want %q", result1, "movie.eng.ass")
	}

	track2 := mkvinfo.SubtitleTrack{Language: "eng", Name: "Commentary"}
	result2 := GenerateOutputPath("movie.mkv", track2, existing)
	expected2 := "movie.eng.Commentary.ass"
	if result2 != expected2 {
		t.Errorf("second track: got %q, want %q", result2, expected2)
	}
}

func TestGenerateOutputPath_CollisionFallbackSequence(t *testing.T) {
	existing := make(map[string]bool)

	track := mkvinfo.SubtitleTrack{Language: "eng"}

	result1 := GenerateOutputPath("movie.mkv", track, existing)
	if result1 != "movie.eng.ass" {
		t.Errorf("first: got %q, want %q", result1, "movie.eng.ass")
	}

	// Second track with no name falls through to sequence numbers
	result2 := GenerateOutputPath("movie.mkv", track, existing)
	if result2 != "movie.eng.2.ass" {
		t.Errorf("second: got %q, want %q", result2, "movie.eng.2.ass")
	}
}

func TestGenerateOutputPath_TripleCollision(t *testing.T) {
	existing := make(map[string]bool)
	track := mkvinfo.SubtitleTrack{Language: "eng"}

	result1 := GenerateOutputPath("movie.mkv", track, existing)
	if result1 != "movie.eng.ass" {
		t.Errorf("first: got %q, want %q", result1, "movie.eng.ass")
	}

	result2 := GenerateOutputPath("movie.mkv", track, existing)
	if result2 != "movie.eng.2.ass" {
		t.Errorf("second: got %q, want %q", result2, "movie.eng.2.ass")
	}

	result3 := GenerateOutputPath("movie.mkv", track, existing)
	if result3 != "movie.eng.3.ass" {
		t.Errorf("third: got %q, want %q", result3, "movie.eng.3.ass")
	}
}

func TestGenerateOutputPath_MarksPathInExisting(t *testing.T) {
	existing := make(map[string]bool)
	track := mkvinfo.SubtitleTrack{Language: "jpn"}

	result := GenerateOutputPath("movie.mkv", track, existing)
	if !existing[result] {
		t.Error("returned path was not marked in existingPaths map")
	}
}

func TestSanitizeFileName_Basic(t *testing.T) {
	result := sanitizeFileName("Commentary")
	if result != "Commentary" {
		t.Errorf("got %q, want %q", result, "Commentary")
	}
}

func TestSanitizeFileName_SpecialCharacters(t *testing.T) {
	result := sanitizeFileName("Track/Name:With<Special>Chars")
	// / : < > replaced with _, collapsed
	if result != "Track_Name_With_Special_Chars" {
		t.Errorf("got %q, want %q", result, "Track_Name_With_Special_Chars")
	}
}

func TestSanitizeFileName_LengthLimit(t *testing.T) {
	long := "This is a very long track name that exceeds the fifty character limit for filenames"
	result := sanitizeFileName(long)
	if len(result) > 50 {
		t.Errorf("sanitized name length %d exceeds 50: %q", len(result), result)
	}
}

func TestSanitizeFileName_EmptyAfterSanitize(t *testing.T) {
	result := sanitizeFileName("///:::")
	if result != "track" {
		t.Errorf("got %q, want %q", result, "track")
	}
}

func TestSanitizeFileName_LeadingTrailingUnderscore(t *testing.T) {
	result := sanitizeFileName("__test__")
	if result != "test" {
		t.Errorf("got %q, want %q", result, "test")
	}
}

func TestSanitizeFileName_MultipleConsecutiveSpecialChars(t *testing.T) {
	result := sanitizeFileName("a!!!b")
	if result != "a_b" {
		t.Errorf("got %q, want %q", result, "a_b")
	}
}

func TestSanitizeFileName_AllowedChars(t *testing.T) {
	// Hyphens, dots, spaces, digits should be preserved
	result := sanitizeFileName("Track-1 v2.0")
	if result != "Track-1 v2.0" {
		t.Errorf("got %q, want %q", result, "Track-1 v2.0")
	}
}
