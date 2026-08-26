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

func TestDecodeNoCompressionFrame(t *testing.T) {
	const headerEnd = 20
	const paletteBytes = 12
	frameOffset := headerEnd + paletteBytes
	frame := []byte{0, 1, 1, 0}
	data := make([]byte, frameOffset+len(frame))
	binary.LittleEndian.PutUint16(data[0:2], 2)
	binary.LittleEndian.PutUint16(data[2:4], 2)
	binary.LittleEndian.PutUint16(data[6:8], 1)
	binary.LittleEndian.PutUint16(data[10:12], FlagInternalPalette|FlagNoCompression)
	binary.LittleEndian.PutUint32(data[12:16], uint32(frameOffset))
	binary.LittleEndian.PutUint32(data[16:20], uint32(len(data)))
	binary.LittleEndian.PutUint16(data[20:22], 0)
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
	if info.DrawnPixels != 4 || info.MissingPalettePixels != 0 {
		t.Fatalf("decode info=%+v", info)
	}
	if got := img.NRGBAAt(0, 0); got.R != 252 || got.G != 0 {
		t.Fatalf("pixel 0,0=%+v", got)
	}
	if got := img.NRGBAAt(1, 0); got.G != 252 || got.R != 0 {
		t.Fatalf("pixel 1,0=%+v", got)
	}
}

func TestDisplayDecoderCompositesJunctionFrames(t *testing.T) {
	frame0 := []byte{1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0xE8, 0x03}
	frame1 := []byte{1, 0, 0, 0, 1, 0, 1, 0, 1, 0, 0, 0, 0xE8, 0x03}
	const headerEnd = 24
	const paletteBytes = 12
	frame0Offset := headerEnd + paletteBytes
	frame1Offset := frame0Offset + len(frame0)
	data := make([]byte, frame1Offset+len(frame1))
	binary.LittleEndian.PutUint16(data[0:2], 2)
	binary.LittleEndian.PutUint16(data[2:4], 1)
	binary.LittleEndian.PutUint16(data[6:8], 2)
	binary.LittleEndian.PutUint16(data[10:12], FlagInternalPalette|FlagJunction)
	binary.LittleEndian.PutUint32(data[12:16], uint32(frame0Offset))
	binary.LittleEndian.PutUint32(data[16:20], uint32(frame1Offset))
	binary.LittleEndian.PutUint32(data[20:24], uint32(len(data)))
	binary.LittleEndian.PutUint16(data[24:26], 0)
	binary.LittleEndian.PutUint16(data[26:28], 2)
	copy(data[28:32], []byte{0, 63, 0, 0})
	copy(data[32:36], []byte{0, 0, 63, 0})
	copy(data[frame0Offset:], frame0)
	copy(data[frame1Offset:], frame1)

	g, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	decoder := g.NewDisplayDecoder()
	first, _, err := decoder.DecodeNext()
	if err != nil {
		t.Fatal(err)
	}
	if got := first.NRGBAAt(0, 0); got.R != 252 || got.G != 0 {
		t.Fatalf("first pixel=%+v", got)
	}
	second, _, err := decoder.DecodeNext()
	if err != nil {
		t.Fatal(err)
	}
	if got := second.NRGBAAt(0, 0); got.R != 252 || got.G != 0 {
		t.Fatalf("composite retained pixel=%+v", got)
	}
	if got := second.NRGBAAt(1, 0); got.G != 252 || got.R != 0 {
		t.Fatalf("composite updated pixel=%+v", got)
	}
}

func TestDecodeNoCompressionFrameAlignmentPadding(t *testing.T) {
	const headerEnd = 20
	const paletteBytes = 16
	frameOffset := headerEnd + paletteBytes
	frame := []byte{0, 1, 2, 0} // three pixels plus one zero alignment byte
	data := make([]byte, frameOffset+len(frame))
	binary.LittleEndian.PutUint16(data[0:2], 3)
	binary.LittleEndian.PutUint16(data[2:4], 1)
	binary.LittleEndian.PutUint16(data[6:8], 1)
	binary.LittleEndian.PutUint16(data[10:12], FlagInternalPalette|FlagNoCompression)
	binary.LittleEndian.PutUint32(data[12:16], uint32(frameOffset))
	binary.LittleEndian.PutUint32(data[16:20], uint32(len(data)))
	binary.LittleEndian.PutUint16(data[20:22], 0)
	binary.LittleEndian.PutUint16(data[22:24], 3)
	copy(data[24:28], []byte{0, 63, 0, 0})
	copy(data[28:32], []byte{0, 0, 63, 0})
	copy(data[32:36], []byte{0, 0, 0, 63})
	copy(data[frameOffset:], frame)

	g, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	img, info, err := g.DecodeFrame(0)
	if err != nil {
		t.Fatal(err)
	}
	if info.DrawnPixels != 3 || info.MissingPalettePixels != 0 {
		t.Fatalf("decode info=%+v", info)
	}
	if got := img.NRGBAAt(2, 0); got.B != 252 {
		t.Fatalf("pixel 2,0=%+v", got)
	}
}

func TestDecodeNoCompressionFrameRejectsNonZeroPadding(t *testing.T) {
	const headerEnd = 20
	const paletteBytes = 8
	frameOffset := headerEnd + paletteBytes
	data := make([]byte, frameOffset+4)
	binary.LittleEndian.PutUint16(data[0:2], 3)
	binary.LittleEndian.PutUint16(data[2:4], 1)
	binary.LittleEndian.PutUint16(data[6:8], 1)
	binary.LittleEndian.PutUint16(data[10:12], FlagInternalPalette|FlagNoCompression)
	binary.LittleEndian.PutUint32(data[12:16], uint32(frameOffset))
	binary.LittleEndian.PutUint32(data[16:20], uint32(len(data)))
	binary.LittleEndian.PutUint16(data[20:22], 0)
	binary.LittleEndian.PutUint16(data[22:24], 1)
	copy(data[24:28], []byte{0, 63, 0, 0})
	copy(data[frameOffset:], []byte{0, 0, 0, 1})

	g, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := g.DecodeFrame(0); err == nil {
		t.Fatal("expected non-zero alignment padding to be rejected")
	}
}
