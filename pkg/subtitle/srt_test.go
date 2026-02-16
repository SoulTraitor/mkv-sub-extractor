package subtitle

import "testing"

func TestConvertSRTTagsToASS(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{"no tags", "Hello world", "Hello world"},
		{"bold", "<b>bold</b>", "{\\b1}bold{\\b0}"},
		{"italic", "<i>italic</i>", "{\\i1}italic{\\i0}"},
		{"underline", "<u>underline</u>", "{\\u1}underline{\\u0}"},
		{"nested bold italic", "<b><i>bold italic</i></b>", "{\\b1}{\\i1}bold italic{\\i0}{\\b0}"},
		{
			"font color red RGB",
			`<font color="#FF0000">red</font>`,
			"{\\c&H0000FF&}red{\\c}",
		},
		{
			"font color green RGB",
			`<font color="#00FF00">green</font>`,
			"{\\c&H00FF00&}green{\\c}",
		},
		{
			"font color blue RGB",
			`<font color="#0000FF">blue</font>`,
			"{\\c&HFF0000&}blue{\\c}",
		},
		{
			"font color single quotes",
			"<font color='#AABBCC'>test</font>",
			"{\\c&HCCBBAA&}test{\\c}",
		},
		{"uppercase tags", "<B>UPPER</B>", "{\\b1}UPPER{\\b0}"},
		{"unclosed bold", "<b>unclosed", "{\\b1}unclosed"},
		{"unknown tags stripped", "<unknown>text</unknown>", "text"},
		{"mixed content", "plain <i>mixed</i> plain", "plain {\\i1}mixed{\\i0} plain"},
		{
			"named color red",
			`<font color="red">text</font>`,
			"{\\c&H0000FF&}text{\\c}",
		},
		{
			"named color blue",
			`<font color="blue">text</font>`,
			"{\\c&HFF0000&}text{\\c}",
		},
		{
			"named color green",
			`<font color="green">text</font>`,
			"{\\c&H008000&}text{\\c}",
		},
		{
			"named color yellow",
			`<font color="yellow">text</font>`,
			"{\\c&H00FFFF&}text{\\c}",
		},
		{
			"named color white",
			`<font color="white">text</font>`,
			"{\\c&HFFFFFF&}text{\\c}",
		},
		{
			"named color cyan",
			`<font color="cyan">text</font>`,
			"{\\c&HFFFF00&}text{\\c}",
		},
		{
			"named color magenta",
			`<font color="magenta">text</font>`,
			"{\\c&HFF00FF&}text{\\c}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertSRTTagsToASS(tt.text)
			if got != tt.want {
				t.Errorf("ConvertSRTTagsToASS(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}
