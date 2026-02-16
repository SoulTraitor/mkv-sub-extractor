// Package extract provides MKV subtitle packet extraction functionality.
//
// It reads raw subtitle packets from a matroska-go Demuxer, filters by track
// number, and applies gap-fill logic for missing end times.
package extract

import (
	"fmt"
	"io"
	"sort"

	matroska "github.com/luispater/matroska-go"
)

// RawSubtitlePacket holds raw packet data from matroska-go before codec-specific parsing.
type RawSubtitlePacket struct {
	StartTime uint64 // nanoseconds
	EndTime   uint64 // nanoseconds (may be 0 if BlockDuration missing)
	Data      []byte // raw block data
}

// defaultEndTimePadding is the fallback duration (5 seconds in nanoseconds) applied
// to the last packet when its EndTime is missing.
const defaultEndTimePadding uint64 = 5_000_000_000

// timestampSanityThreshold is the threshold above which a StartTime value is likely
// IEEE 754 bit-encoded rather than a real nanosecond value. 10^16 ns ~ 115 days,
// which is generous for any real subtitle.
const timestampSanityThreshold uint64 = 10_000_000_000_000_000

// ExtractSubtitlePackets reads all packets from the demuxer, filters by trackNumber,
// collects them into RawSubtitlePackets, and applies gap-fill for missing EndTime values.
//
// Returns an error wrapping the underlying demuxer read error if one occurs (other than io.EOF).
// A warning is logged (via the returned warning string) if timestamp values appear to be
// IEEE 754 bit-encoded rather than real nanosecond values.
func ExtractSubtitlePackets(demuxer *matroska.Demuxer, trackNumber uint8) ([]RawSubtitlePacket, string, error) {
	var packets []RawSubtitlePacket
	var warning string

	for {
		pkt, err := demuxer.ReadPacket()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("read packet error: %w", err)
		}

		if pkt.Track != trackNumber {
			continue
		}

		raw := RawSubtitlePacket{
			StartTime: pkt.StartTime,
			EndTime:   pkt.EndTime,
			Data:      pkt.Data,
		}

		// Sanity check on first collected packet
		if len(packets) == 0 && raw.StartTime > timestampSanityThreshold {
			warning = fmt.Sprintf(
				"WARNING: first subtitle packet StartTime=%d exceeds sanity threshold (%d), "+
					"timestamps may be IEEE 754 bit-encoded rather than nanoseconds",
				raw.StartTime, timestampSanityThreshold,
			)
		}

		packets = append(packets, raw)
	}

	ApplyGapFill(packets)

	return packets, warning, nil
}

// ApplyGapFill fills in missing EndTime values using a gap-fill strategy:
//   - Sort packets by StartTime
//   - If a packet's EndTime <= StartTime, set EndTime to the next packet's StartTime
//   - For the last packet with missing EndTime, set EndTime = StartTime + 5 seconds
//   - Empty slices are handled safely (no-op)
func ApplyGapFill(packets []RawSubtitlePacket) {
	if len(packets) == 0 {
		return
	}

	// Sort by StartTime for correct gap-fill ordering
	sort.Slice(packets, func(i, j int) bool {
		return packets[i].StartTime < packets[j].StartTime
	})

	for i := range packets {
		if packets[i].EndTime <= packets[i].StartTime {
			if i+1 < len(packets) {
				packets[i].EndTime = packets[i+1].StartTime
			} else {
				packets[i].EndTime = packets[i].StartTime + defaultEndTimePadding
			}
		}
	}
}
