package moo2gfx

import (
	"encoding/binary"
	"testing"
)

func TestParseAndDecodeInternalPaletteGraphic(t *testing.T) {
	// 2x1, one frame, internal palette at index 128 with two colors.
	// Frame: indicator=1, y=0, run=2, dx=0, pixels 128/129, end-of-frame.
	frame := []byte{1, 0, 0, 0, 2, 0, 0, 0, 128, 129, 0, 0, 0xE8, 0x03}
	const headerEnd = 20
	const paletteBytes = 12 // shift/count + 2 colors
	frameOffset := headerEnd + paletteBytes
	data := make([]byte, frameOffset+len(frame))
	binary.LittleEndian.PutUint16(data[0:2], 2)
	binary.LittleEndian.PutUint16(data[2:4], 1)
	binary.LittleEndian.PutUint16(data[4:6], 0)
	binary.LittleEndian.PutUint16(data[6:8], 1)
	binary.LittleEndian.PutUint16(data[8:10], 0)
	binary.LittleEndian.PutUint16(data[10:12], FlagInternalPalette)
	binary.LittleEndian.PutUint32(data[12:16], uint32(frameOffset))
	binary.LittleEndian.PutUint32(data[16:20], uint32(len(data)))
	binary.LittleEndian.PutUint16(data[20:22], 128)
	binary.LittleEndian.PutUint16(data[22:24], 2)
	copy(data[24:28], []byte{0, 63, 0, 0})
	copy(data[28:32], []byte{0, 0, 63, 0})
	copy(data[frameOffset:], frame)

	g, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	img, info, err := g.DecodeFrame(0)
	if err != nil {
		t.Fatal(err)
	}
	if info.DrawnPixels != 2 || info.MissingPalettePixels != 0 {
		t.Fatalf("decode info=%+v", info)
	}
	p0 := img.NRGBAAt(0, 0)
	p1 := img.NRGBAAt(1, 0)
	if p0.R != 252 || p0.G != 0 || p1.G != 252 || p1.R != 0 {
		t.Fatalf("pixels=%+v %+v", p0, p1)
	}
}
