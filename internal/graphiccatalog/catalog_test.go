package graphiccatalog

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildCatalogAndExport(t *testing.T) {
	source := t.TempDir()
	out := filepath.Join(t.TempDir(), "out")

	internal := syntheticGraphic(true)
	external := syntheticGraphic(false)
	archive := buildLBX([][]byte{internal, external, []byte("not-a-graphic")})
	if err := os.WriteFile(filepath.Join(source, "TEST.LBX"), archive, 0o644); err != nil {
		t.Fatal(err)
	}

	manifest, err := Build(source, out, Options{ExportPNG: true})
	if err != nil {
		t.Fatal(err)
	}
	if manifest.LBXArchives != 1 {
		t.Fatalf("LBXArchives=%d want=1", manifest.LBXArchives)
	}
	if manifest.GraphicBlocks != 2 || manifest.InternalPalette != 1 || manifest.ExternalPalette != 1 {
		t.Fatalf("unexpected graphics counts: %+v", manifest)
	}
	if manifest.FramesTotal != 2 || manifest.FramesExported != 1 || manifest.FramesComplete != 1 || manifest.FramesFailed != 0 {
		t.Fatalf("unexpected frame counts: %+v", manifest)
	}
	pngPath := filepath.Join(out, "test", "block_0000_frame_000.png")
	if _, err := os.Stat(pngPath); err != nil {
		t.Fatalf("expected PNG %s: %v", pngPath, err)
	}
	if _, err := os.Stat(filepath.Join(out, "manifest.json")); err != nil {
		t.Fatalf("manifest missing: %v", err)
	}
	if got := manifest.Archives[0].Graphics[1].Frames[0].Status; got != "external_palette_pending" {
		t.Fatalf("external frame status=%q", got)
	}
}

func syntheticGraphic(internalPalette bool) []byte {
	frame := []byte{1, 0, 0, 0, 2, 0, 0, 0, 128, 129, 0, 0, 0xE8, 0x03}
	headerEnd := 20
	paletteBytes := 0
	flags := uint16(0)
	if internalPalette {
		flags = 0x1000
		paletteBytes = 12
	}
	frameOffset := headerEnd + paletteBytes
	data := make([]byte, frameOffset+len(frame))
	binary.LittleEndian.PutUint16(data[0:2], 2)
	binary.LittleEndian.PutUint16(data[2:4], 1)
	binary.LittleEndian.PutUint16(data[4:6], 0)
	binary.LittleEndian.PutUint16(data[6:8], 1)
	binary.LittleEndian.PutUint16(data[8:10], 0)
	binary.LittleEndian.PutUint16(data[10:12], flags)
	binary.LittleEndian.PutUint32(data[12:16], uint32(frameOffset))
	binary.LittleEndian.PutUint32(data[16:20], uint32(len(data)))
	if internalPalette {
		binary.LittleEndian.PutUint16(data[20:22], 128)
		binary.LittleEndian.PutUint16(data[22:24], 2)
		copy(data[24:28], []byte{0, 63, 0, 0})
		copy(data[28:32], []byte{0, 0, 63, 0})
	}
	copy(data[frameOffset:], frame)
	return data
}

func buildLBX(blocks [][]byte) []byte {
	count := len(blocks)
	headerEnd := 8 + (count+1)*4
	offset := headerEnd
	offsets := make([]int, count+1)
	for i, block := range blocks {
		offsets[i] = offset
		offset += len(block)
	}
	offsets[count] = offset
	data := make([]byte, offset)
	binary.LittleEndian.PutUint16(data[0:2], uint16(count))
	binary.LittleEndian.PutUint16(data[2:4], 0xFEAD)
	for i, value := range offsets {
		binary.LittleEndian.PutUint32(data[8+i*4:12+i*4], uint32(value))
	}
	cursor := headerEnd
	for _, block := range blocks {
		copy(data[cursor:], block)
		cursor += len(block)
	}
	return data
}

func TestBuildResolvesConfirmedBLDG0Palette(t *testing.T) {
	source := t.TempDir()
	out := filepath.Join(t.TempDir(), "out")

	palette := make([]byte, 256*4)
	for i := 0; i < 256; i++ {
		palette[i*4] = 1
	}
	palette[128*4+1] = 63
	palette[129*4+2] = 63
	fonts := buildLBX([][]byte{[]byte("font"), []byte("unused"), palette})
	if err := os.WriteFile(filepath.Join(source, "FONTS.LBX"), fonts, 0o644); err != nil {
		t.Fatal(err)
	}
	bldg := buildLBX([][]byte{syntheticGraphic(false)})
	if err := os.WriteFile(filepath.Join(source, "BLDG0.LBX"), bldg, 0o644); err != nil {
		t.Fatal(err)
	}

	manifest, err := Build(source, out, Options{ExportPNG: true})
	if err != nil {
		t.Fatal(err)
	}
	archive := findArchiveRecord(t, manifest, "BLDG0.LBX")
	gr := archive.Graphics[0]
	if gr.PaletteContextStatus != "resolved" || gr.PaletteContextSource != "FONTS.LBX#2" || gr.PaletteContextConfidence != "confirmed" {
		t.Fatalf("palette context=%+v", gr)
	}
	if len(gr.Frames) != 1 || gr.Frames[0].Status != "complete" {
		t.Fatalf("frames=%+v", gr.Frames)
	}
	if manifest.PaletteContextsResolved < 1 {
		t.Fatalf("resolved contexts=%d", manifest.PaletteContextsResolved)
	}
}

func TestBuildResolvesConfirmedCouncilPaletteCarrier(t *testing.T) {
	source := t.TempDir()
	out := filepath.Join(t.TempDir(), "out")
	council := buildLBX([][]byte{syntheticFullPaletteGraphic(), syntheticGraphic(false)})
	if err := os.WriteFile(filepath.Join(source, "COUNCIL.LBX"), council, 0o644); err != nil {
		t.Fatal(err)
	}

	manifest, err := Build(source, out, Options{ExportPNG: true})
	if err != nil {
		t.Fatal(err)
	}
	archive := findArchiveRecord(t, manifest, "COUNCIL.LBX")
	if len(archive.Graphics) != 2 {
		t.Fatalf("graphics=%d", len(archive.Graphics))
	}
	gr := archive.Graphics[1]
	if gr.PaletteContextStatus != "resolved" || gr.PaletteContextSource != "COUNCIL.LBX#0" {
		t.Fatalf("palette context=%+v", gr)
	}
	if len(gr.Frames) != 1 || gr.Frames[0].Status != "complete" {
		t.Fatalf("frames=%+v", gr.Frames)
	}
}

func syntheticFullPaletteGraphic() []byte {
	frame := []byte{1, 0, 0, 0, 2, 0, 0, 0, 128, 129, 0, 0, 0xE8, 0x03}
	const headerEnd = 20
	const paletteBytes = 4 + 256*4
	frameOffset := headerEnd + paletteBytes
	data := make([]byte, frameOffset+len(frame))
	binary.LittleEndian.PutUint16(data[0:2], 2)
	binary.LittleEndian.PutUint16(data[2:4], 1)
	binary.LittleEndian.PutUint16(data[6:8], 1)
	binary.LittleEndian.PutUint16(data[10:12], 0x1000)
	binary.LittleEndian.PutUint32(data[12:16], uint32(frameOffset))
	binary.LittleEndian.PutUint32(data[16:20], uint32(len(data)))
	binary.LittleEndian.PutUint16(data[20:22], 0)
	binary.LittleEndian.PutUint16(data[22:24], 256)
	for i := 0; i < 256; i++ {
		off := 24 + i*4
		data[off] = 1
	}
	data[24+128*4+1] = 63
	data[24+129*4+2] = 63
	copy(data[frameOffset:], frame)
	return data
}

func findArchiveRecord(t *testing.T, manifest *Manifest, path string) ArchiveRecord {
	t.Helper()
	for _, archive := range manifest.Archives {
		if archive.Path == path {
			return archive
		}
	}
	t.Fatalf("archive %s not found", path)
	return ArchiveRecord{}
}

func TestShipPaletteCarrierMapping(t *testing.T) {
	tests := map[int]int{
		0: 49, 48: 49,
		50: 99, 98: 99,
		100: 149, 148: 149,
		150: 199, 198: 199,
		200: 249, 248: 249,
		250: 299, 298: 299,
		300: 349, 348: 349,
		350: 399, 398: 399,
		400: 413, 404: 413,
		407: 419,
		408: 414, 412: 414, 420: 414, 424: 414,
		409: 416, 421: 416,
		410: 418, 422: 418,
		411: 415, 423: 415,
	}
	for block, want := range tests {
		if got := shipPaletteCarrier(block); got != want {
			t.Fatalf("block %d carrier=%d, want %d", block, got, want)
		}
	}
	for _, block := range []int{49, 99, 149, 199, 249, 299, 349, 399, 405, 406, 413, 414, 415, 416, 417, 418, 419, 425, 448} {
		if got := shipPaletteCarrier(block); got != -1 {
			t.Fatalf("block %d carrier=%d, want unresolved", block, got)
		}
	}
}
