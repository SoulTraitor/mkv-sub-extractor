package mkvinfo

// codecInfo holds human-readable format type and text/image classification for a codec.
type codecInfo struct {
	FormatType string // Human-readable name (e.g. "SRT", "ASS")
	IsText     bool   // true = text subtitle, false = image/bitmap subtitle
}

// codecMap maps Matroska codec ID strings to their classification.
// Source: https://www.matroska.org/technical/subtitles.html
var codecMap = map[string]codecInfo{
	"S_TEXT/UTF8":    {FormatType: "SRT", IsText: true},
	"S_TEXT/SSA":     {FormatType: "SSA", IsText: true},
	"S_TEXT/ASS":     {FormatType: "ASS", IsText: true},
	"S_TEXT/WEBVTT":  {FormatType: "WebVTT", IsText: true},
	"S_TEXT/USF":     {FormatType: "USF", IsText: true},
	"S_HDMV/PGS":    {FormatType: "PGS", IsText: false},
	"S_VOBSUB":      {FormatType: "VobSub", IsText: false},
	"S_DVBSUB":      {FormatType: "DVB", IsText: false},
	"S_HDMV/TEXTST": {FormatType: "HDMV Text", IsText: true},
	"S_ARIBSUB":     {FormatType: "ARIB", IsText: true},
}

// ClassifyCodec returns the human-readable format type and text/image classification
// for a given Matroska codec ID. Unknown codec IDs return the raw ID as format type
// and isText=false. Never panics.
func ClassifyCodec(codecID string) (formatType string, isText bool) {
	if info, ok := codecMap[codecID]; ok {
		return info.FormatType, info.IsText
	}
	return codecID, false
}
