package rawextract

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildExtractsLBXAndSmacker(t *testing.T) {
	source := t.TempDir()
	out := filepath.Join(t.TempDir(), "out")
	archive := buildLBX([][]byte{[]byte("alpha"), []byte{1, 2, 3}})
	if err := os.WriteFile(filepath.Join(source, "DATA.LBX"), archive, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "MOVIE.LBX"), []byte("SMK2movie-data"), 0o644); err != nil {
		t.Fatal(err)
	}

	manifest, err := Build(source, out, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if manifest.LBXArchives != 1 || manifest.SmackerFiles != 1 || manifest.BlockCount != 2 {
		t.Fatalf("unexpected manifest counts: %+v", manifest)
	}
	if manifest.PayloadBytes != 8 {
		t.Fatalf("payload bytes=%d want=8", manifest.PayloadBytes)
	}
	data, err := os.ReadFile(filepath.Join(out, "lbx", "data", "block_0000.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "alpha" {
		t.Fatalf("block data=%q", data)
	}
	smk, err := os.ReadFile(filepath.Join(out, "smacker", "MOVIE.smk"))
	if err != nil {
		t.Fatal(err)
	}
	if string(smk[:4]) != "SMK2" {
		t.Fatalf("smacker signature=%q", smk[:4])
	}
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
