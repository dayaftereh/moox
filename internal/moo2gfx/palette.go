package moo2gfx

import (
	"fmt"
	"image"
	"image/color"
)

const ExternalPaletteBytes = 256 * 4

type PaletteColor struct {
	Index  int  `json:"index"`
	Marker byte `json:"marker"`
	DACR   byte `json:"dac_r"`
	DACG   byte `json:"dac_g"`
	DACB   byte `json:"dac_b"`
	R      byte `json:"r"`
	G      byte `json:"g"`
	B      byte `json:"b"`
}

type ExternalPalette struct {
	Colors [256]PaletteColor
}

func ParseExternalPalette(data []byte) (*ExternalPalette, error) {
	if len(data) < ExternalPaletteBytes {
		return nil, fmt.Errorf("external palette needs at least %d bytes, got %d", ExternalPaletteBytes, len(data))
	}
	p := &ExternalPalette{}
	for i := 0; i < 256; i++ {
		off := i * 4
		p.Colors[i] = PaletteColor{
			Index:  i,
			Marker: data[off],
			DACR:   data[off+1],
			DACG:   data[off+2],
			DACB:   data[off+3],
			R:      scaleDAC(data[off+1]),
			G:      scaleDAC(data[off+2]),
			B:      scaleDAC(data[off+3]),
		}
	}
	return p, nil
}

func (p *ExternalPalette) Apply(g *Graphic) {
	for i, entry := range p.Colors {
		if g.PaletteValid[i] {
			continue
		}
		g.Palette[i] = color.NRGBA{R: entry.R, G: entry.G, B: entry.B, A: 255}
		g.PaletteValid[i] = true
	}
}

func (p *ExternalPalette) Swatch(cell int) *image.NRGBA {
	if cell < 1 {
		cell = 1
	}
	img := image.NewNRGBA(image.Rect(0, 0, 16*cell, 16*cell))
	for i, entry := range p.Colors {
		x0 := (i % 16) * cell
		y0 := (i / 16) * cell
		c := color.NRGBA{R: entry.R, G: entry.G, B: entry.B, A: 255}
		for y := y0; y < y0+cell; y++ {
			for x := x0; x < x0+cell; x++ {
				img.SetNRGBA(x, y, c)
			}
		}
	}
	return img
}
