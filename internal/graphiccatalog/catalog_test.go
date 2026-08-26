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
