package extract

import (
	"testing"
)

func TestRawSubtitlePacketType(t *testing.T) {
	// Verify the type is accessible and fields work correctly
	pkt := RawSubtitlePacket{
		StartTime: 1_000_000_000,
		EndTime:   2_000_000_000,
		Data:      []byte("hello"),
	}
	if pkt.StartTime != 1_000_000_000 {
		t.Errorf("StartTime = %d, want 1000000000", pkt.StartTime)
	}
	if pkt.EndTime != 2_000_000_000 {
		t.Errorf("EndTime = %d, want 2000000000", pkt.EndTime)
	}
	if string(pkt.Data) != "hello" {
		t.Errorf("Data = %q, want %q", string(pkt.Data), "hello")
	}
}

func TestApplyGapFill_AllValid(t *testing.T) {
	// All packets have valid EndTime -> no changes
	packets := []RawSubtitlePacket{
		{StartTime: 1_000_000_000, EndTime: 2_000_000_000, Data: []byte("a")},
		{StartTime: 3_000_000_000, EndTime: 4_000_000_000, Data: []byte("b")},
		{StartTime: 5_000_000_000, EndTime: 6_000_000_000, Data: []byte("c")},
	}
	ApplyGapFill(packets)

	expected := []uint64{2_000_000_000, 4_000_000_000, 6_000_000_000}
	for i, pkt := range packets {
		if pkt.EndTime != expected[i] {
			t.Errorf("packet[%d].EndTime = %d, want %d", i, pkt.EndTime, expected[i])
		}
	}
}

func TestApplyGapFill_EndTimeZero(t *testing.T) {
	// Packet with EndTime=0 -> gets next packet's StartTime
	packets := []RawSubtitlePacket{
		{StartTime: 1_000_000_000, EndTime: 0, Data: []byte("a")},
		{StartTime: 3_000_000_000, EndTime: 4_000_000_000, Data: []byte("b")},
	}
	ApplyGapFill(packets)

	if packets[0].EndTime != 3_000_000_000 {
		t.Errorf("packet[0].EndTime = %d, want 3000000000", packets[0].EndTime)
	}
	if packets[1].EndTime != 4_000_000_000 {
		t.Errorf("packet[1].EndTime = %d, want 4000000000", packets[1].EndTime)
	}
}

func TestApplyGapFill_EndTimeEqualsStartTime(t *testing.T) {
	// Packet with EndTime==StartTime -> gets next packet's StartTime
	packets := []RawSubtitlePacket{
		{StartTime: 1_000_000_000, EndTime: 1_000_000_000, Data: []byte("a")},
		{StartTime: 3_000_000_000, EndTime: 5_000_000_000, Data: []byte("b")},
	}
	ApplyGapFill(packets)

	if packets[0].EndTime != 3_000_000_000 {
		t.Errorf("packet[0].EndTime = %d, want 3000000000", packets[0].EndTime)
	}
}

func TestApplyGapFill_LastPacketMissing(t *testing.T) {
	// Last packet with EndTime=0 -> gets StartTime + 5s
	packets := []RawSubtitlePacket{
		{StartTime: 1_000_000_000, EndTime: 2_000_000_000, Data: []byte("a")},
		{StartTime: 3_000_000_000, EndTime: 0, Data: []byte("b")},
	}
	ApplyGapFill(packets)

	expected := uint64(3_000_000_000 + 5_000_000_000)
	if packets[1].EndTime != expected {
		t.Errorf("packet[1].EndTime = %d, want %d", packets[1].EndTime, expected)
	}
}

func TestApplyGapFill_SinglePacketMissing(t *testing.T) {
	// Single packet with EndTime=0 -> gets StartTime + 5s
	packets := []RawSubtitlePacket{
		{StartTime: 10_000_000_000, EndTime: 0, Data: []byte("a")},
	}
	ApplyGapFill(packets)

	expected := uint64(10_000_000_000 + 5_000_000_000)
	if packets[0].EndTime != expected {
		t.Errorf("packet[0].EndTime = %d, want %d", packets[0].EndTime, expected)
	}
}

func TestApplyGapFill_Empty(t *testing.T) {
	// Empty slice -> no panic
	packets := []RawSubtitlePacket{}
	ApplyGapFill(packets) // Should not panic
}

func TestApplyGapFill_NilSlice(t *testing.T) {
	// Nil slice -> no panic
	ApplyGapFill(nil) // Should not panic
}

func TestApplyGapFill_SortsBeforeFilling(t *testing.T) {
	// Packets arrive out of order; gap-fill should sort first
	packets := []RawSubtitlePacket{
		{StartTime: 5_000_000_000, EndTime: 0, Data: []byte("c")},
		{StartTime: 1_000_000_000, EndTime: 0, Data: []byte("a")},
		{StartTime: 3_000_000_000, EndTime: 0, Data: []byte("b")},
	}
	ApplyGapFill(packets)

	// After sort: [1s, 3s, 5s]
	// Gap-fill: 1s->3s, 3s->5s, 5s->5s+5s=10s
	if packets[0].StartTime != 1_000_000_000 {
		t.Errorf("packet[0].StartTime = %d, want 1000000000", packets[0].StartTime)
	}
	if packets[0].EndTime != 3_000_000_000 {
		t.Errorf("packet[0].EndTime = %d, want 3000000000", packets[0].EndTime)
	}
	if packets[1].StartTime != 3_000_000_000 {
		t.Errorf("packet[1].StartTime = %d, want 3000000000", packets[1].StartTime)
	}
	if packets[1].EndTime != 5_000_000_000 {
		t.Errorf("packet[1].EndTime = %d, want 5000000000", packets[1].EndTime)
	}
	if packets[2].StartTime != 5_000_000_000 {
		t.Errorf("packet[2].StartTime = %d, want 5000000000", packets[2].StartTime)
	}
	expected := uint64(5_000_000_000 + 5_000_000_000)
	if packets[2].EndTime != expected {
		t.Errorf("packet[2].EndTime = %d, want %d", packets[2].EndTime, expected)
	}
}

func TestApplyGapFill_MultipleMissingMiddle(t *testing.T) {
	// Multiple packets in a row with missing EndTime
	packets := []RawSubtitlePacket{
		{StartTime: 1_000_000_000, EndTime: 0, Data: []byte("a")},
		{StartTime: 2_000_000_000, EndTime: 0, Data: []byte("b")},
		{StartTime: 3_000_000_000, EndTime: 4_000_000_000, Data: []byte("c")},
	}
	ApplyGapFill(packets)

	if packets[0].EndTime != 2_000_000_000 {
		t.Errorf("packet[0].EndTime = %d, want 2000000000", packets[0].EndTime)
	}
	if packets[1].EndTime != 3_000_000_000 {
		t.Errorf("packet[1].EndTime = %d, want 3000000000", packets[1].EndTime)
	}
	if packets[2].EndTime != 4_000_000_000 {
		t.Errorf("packet[2].EndTime = %d, want 4000000000", packets[2].EndTime)
	}
}
