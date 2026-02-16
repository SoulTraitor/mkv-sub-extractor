package subtitle

import "testing"

func TestParseASSBlockData(t *testing.T) {
	tests := []struct {
		name          string
		data          []byte
		wantReadOrder int
		wantLayer     int
		wantRemaining string
		wantErr       bool
	}{
		{
			name:          "basic dialogue",
			data:          []byte("0,0,Default,,0,0,0,,Hello world"),
			wantReadOrder: 0,
			wantLayer:     0,
			wantRemaining: "Default,,0,0,0,,Hello world",
		},
		{
			name:          "comma in text",
			data:          []byte("1,0,Default,,0,0,0,,Hello, world!"),
			wantReadOrder: 1,
			wantLayer:     0,
			wantRemaining: "Default,,0,0,0,,Hello, world!",
		},
		{
			name:          "multiple commas in text",
			data:          []byte("42,1,Italic,,0,0,0,,Text with, many, commas"),
			wantReadOrder: 42,
			wantLayer:     1,
			wantRemaining: "Italic,,0,0,0,,Text with, many, commas",
		},
		{
			name:          "empty text field",
			data:          []byte("0,0,Default,,0,0,0,,"),
			wantReadOrder: 0,
			wantLayer:     0,
			wantRemaining: "Default,,0,0,0,,",
		},
		{
			name:    "not enough fields",
			data:    []byte("bad"),
			wantErr: true,
		},
		{
			name:    "non-numeric ReadOrder",
			data:    []byte("abc,0,Default,,0,0,0,,text"),
			wantErr: true,
		},
		{
			name:    "empty data",
			data:    []byte(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readOrder, layer, remaining, err := ParseASSBlockData(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseASSBlockData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if readOrder != tt.wantReadOrder {
				t.Errorf("readOrder = %d, want %d", readOrder, tt.wantReadOrder)
			}
			if layer != tt.wantLayer {
				t.Errorf("layer = %d, want %d", layer, tt.wantLayer)
			}
			if remaining != tt.wantRemaining {
				t.Errorf("remaining = %q, want %q", remaining, tt.wantRemaining)
			}
		})
	}
}
