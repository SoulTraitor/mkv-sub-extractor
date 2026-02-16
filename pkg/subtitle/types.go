package subtitle

// SubtitleEvent represents a single subtitle dialogue line extracted from an MKV file.
// Used by both ASS passthrough and SRT-to-ASS conversion paths.
type SubtitleEvent struct {
	Start     uint64 // nanoseconds (from Packet.StartTime)
	End       uint64 // nanoseconds (from Packet.EndTime)
	Layer     int    // ASS layer number (from block data or 0 for SRT)
	Style     string // ASS style name (from block data or "Default" for SRT)
	Name      string // speaker name (from block data or empty)
	MarginL   string // margin override or "0"
	MarginR   string // margin override or "0"
	MarginV   string // margin override or "0"
	Effect    string // ASS effect or empty
	Text      string // dialogue text (ASS override tags or converted SRT text)
	ReadOrder int    // ASS block ReadOrder (for stable sort; 0 for SRT)
}
