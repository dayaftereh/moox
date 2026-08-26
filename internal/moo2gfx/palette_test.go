package moo2gfx

import "testing"

func TestParseExternalPalette(t *testing.T) {
	data := make([]byte, ExternalPaletteBytes)
	data[0] = 1
	data[1] = 10
	data[2] = 15
	data[3] = 19
	data[1020] = 1
	data[1021] = 63
	data[1022] = 62
	data[1023] = 61

	p, err := ParseExternalPalette(data)
	if err != nil {
		t.Fatal(err)
	}
	first := p.Colors[0]
	if first.Marker != 1 || first.R != 40 || first.G != 60 || first.B != 76 {
		t.Fatalf("first=%+v", first)
	}
	last := p.Colors[255]
	if last.R != 252 || last.G != 248 || last.B != 244 {
		t.Fatalf("last=%+v", last)
	}
	if got := p.Swatch(2).Bounds().Dx(); got != 32 {
		t.Fatalf("swatch width=%d", got)
	}
}

func TestExternalPaletteApplyPreservesInternal(t *testing.T) {
	data := make([]byte, ExternalPaletteBytes)
	data[4+1] = 10
	p, err := ParseExternalPalette(data)
	if err != nil {
		t.Fatal(err)
	}
	g := &Graphic{}
	g.PaletteValid[1] = true
	g.Palette[1].R = 99
	p.Apply(g)
	if g.Palette[1].R != 99 {
		t.Fatalf("internal palette entry overwritten: %+v", g.Palette[1])
	}
	if !g.PaletteValid[2] {
		t.Fatal("external palette did not fill missing entry")
	}
}
