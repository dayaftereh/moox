package lbx

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenAndReadEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.lbx")

	const dataOffset = uint32(0x20)
	payloadA := []byte("alpha")
	payloadB := []byte{1, 2, 3, 4}
	buf := make([]byte, int(dataOffset)+len(payloadA)+len(payloadB))
	binary.LittleEndian.PutUint16(buf[0:2], 2)
	binary.LittleEndian.PutUint16(buf[2:4], Magic)
	binary.LittleEndian.PutUint32(buf[4:8], 0)
	binary.LittleEndian.PutUint32(buf[8:12], dataOffset)
	binary.LittleEndian.PutUint32(buf[12:16], dataOffset+uint32(len(payloadA)))
	binary.LittleEndian.PutUint32(buf[16:20], dataOffset+uint32(len(payloadA)+len(payloadB)))
	copy(buf[dataOffset:], payloadA)
	copy(buf[int(dataOffset)+len(payloadA):], payloadB)
	if err := os.WriteFile(path, buf, 0o644); err != nil {
		t.Fatal(err)
	}

	parsed, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.EntryCount != 2 || len(parsed.Entries) != 2 {
		t.Fatalf("unexpected entry count: %+v", parsed)
	}
	if parsed.DataOffset != dataOffset {
		t.Fatalf("data offset=%d want=%d", parsed.DataOffset, dataOffset)
	}

	gotA, err := parsed.ReadEntry(0)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotA, payloadA) {
		t.Fatalf("entry 0=%q want=%q", gotA, payloadA)
	}
	gotB, err := parsed.ReadEntry(1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotB, payloadB) {
		t.Fatalf("entry 1=%v want=%v", gotB, payloadB)
	}
}

func TestRejectsWrongMagic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.lbx")
	data := make([]byte, 32)
	binary.LittleEndian.PutUint16(data[0:2], 1)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Open(path)
	if !errors.Is(err, ErrNotLBX) {
		t.Fatalf("error=%v, want ErrNotLBX", err)
	}
}

func TestDetectSmacker(t *testing.T) {
	path := filepath.Join(t.TempDir(), "movie.lbx")
	if err := os.WriteFile(path, []byte("SMK2payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	kind, err := Detect(path)
	if err != nil {
		t.Fatal(err)
	}
	if kind != KindSmacker {
		t.Fatalf("kind=%s want=%s", kind, KindSmacker)
	}
}
