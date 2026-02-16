package assout

import "testing"

func TestFormatASSTimestamp(t *testing.T) {
	tests := []struct {
		name string
		ns   uint64
		want string
	}{
		{"zero", 0, "0:00:00.00"},
		{"1 second", 1_000_000_000, "0:00:01.00"},
		{"1 minute", 60_000_000_000, "0:01:00.00"},
		{"1 hour", 3_600_000_000_000, "1:00:00.00"},
		{"fractional seconds 5.55s", 5_550_000_000, "0:00:05.55"},
		{"rounding up at 5ms boundary", 5_555_000_000, "0:00:05.56"},
		{"rounding down at 4ms", 5_554_000_000, "0:00:05.55"},
		{"all components 1:02:03.45", 3_723_450_000_000, "1:02:03.45"},
		{"10 hours multi-digit", 36_000_000_000_000, "10:00:00.00"},
		{"exactly 5ms rounds to 1cs", 5_000_000, "0:00:00.01"},
		{"just under 0.5cs rounds to 0", 4_999_999, "0:00:00.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatASSTimestamp(tt.ns)
			if got != tt.want {
				t.Errorf("FormatASSTimestamp(%d) = %q, want %q", tt.ns, got, tt.want)
			}
		})
	}
}
