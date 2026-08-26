package lbx

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

const Magic uint16 = 0xFEAD

var ErrNotLBX = errors.New("not a SimTex LBX container")

// Kind describes the outer file format detected from the file header.
type Kind string

const (
	KindLBX     Kind = "lbx"
	KindSmacker Kind = "smacker"
	KindUnknown Kind = "unknown"
)

// Entry describes one block stored in an LBX archive.
type Entry struct {
	Index  int   `json:"index"`
	Offset int64 `json:"offset"`
	Size   int64 `json:"size"`
}

// File contains parsed LBX metadata. Data blocks are read lazily from Path.
type File struct {
	Path       string  `json:"path"`
	Size       int64   `json:"size"`
	EntryCount uint16  `json:"entry_count"`
	Reserved   uint32  `json:"reserved"`
	DataOffset uint32  `json:"data_offset"`
	Entries    []Entry `json:"entries"`
}

// Detect identifies formats that are relevant to the MOO2 data set.
func Detect(path string) (Kind, error) {
	f, err := os.Open(path)
	if err != nil {
		return KindUnknown, err
	}
	defer f.Close()

	var head [4]byte
	n, err := f.Read(head[:])
	if err != nil && err != io.EOF {
		return KindUnknown, err
	}
	if n < len(head) {
		return KindUnknown, nil
	}
	if string(head[:]) == "SMK2" || string(head[:]) == "SMK4" {
		return KindSmacker, nil
	}
	if binary.LittleEndian.Uint16(head[2:4]) == Magic {
		return KindLBX, nil
	}
	return KindUnknown, nil
}

// Open parses a SimTex LBX container without loading its payloads into memory.
func Open(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if stat.Size() < 12 {
		return nil, fmt.Errorf("%w: file too small", ErrNotLBX)
	}

	var header [8]byte
	if _, err := io.ReadFull(f, header[:]); err != nil {
		return nil, err
	}

	count := binary.LittleEndian.Uint16(header[0:2])
	magic := binary.LittleEndian.Uint16(header[2:4])
	reserved := binary.LittleEndian.Uint32(header[4:8])
	if magic != Magic || count == 0 {
		return nil, fmt.Errorf("%w: magic=0x%04X entries=%d", ErrNotLBX, magic, count)
	}

	offsetCount := int(count) + 1
	tableBytes := int64(offsetCount * 4)
	tableEnd := int64(8) + tableBytes
	if tableEnd > stat.Size() {
		return nil, fmt.Errorf("invalid LBX offset table: need %d bytes, file has %d", tableEnd, stat.Size())
	}

	rawOffsets := make([]byte, tableBytes)
	if _, err := io.ReadFull(f, rawOffsets); err != nil {
		return nil, err
	}
	offsets := make([]uint32, offsetCount)
	for i := range offsets {
		offsets[i] = binary.LittleEndian.Uint32(rawOffsets[i*4 : i*4+4])
	}

	if int64(offsets[0]) < tableEnd {
		return nil, fmt.Errorf("invalid LBX first data offset 0x%X before table end 0x%X", offsets[0], tableEnd)
	}

	entries := make([]Entry, count)
	for i := 0; i < int(count); i++ {
		start := int64(offsets[i])
		end := int64(offsets[i+1])
		if start < 0 || end < start || end > stat.Size() {
			return nil, fmt.Errorf("invalid LBX entry %d range [%d,%d) for file size %d", i, start, end, stat.Size())
		}
		entries[i] = Entry{Index: i, Offset: start, Size: end - start}
	}

	return &File{
		Path:       path,
		Size:       stat.Size(),
		EntryCount: count,
		Reserved:   reserved,
		DataOffset: offsets[0],
		Entries:    entries,
	}, nil
}

// CopyEntry streams one LBX block to w.
func (f *File) CopyEntry(index int, w io.Writer) error {
	if index < 0 || index >= len(f.Entries) {
		return fmt.Errorf("entry index %d out of range [0,%d)", index, len(f.Entries))
	}

	src, err := os.Open(f.Path)
	if err != nil {
		return err
	}
	defer src.Close()

	entry := f.Entries[index]
	if _, err := src.Seek(entry.Offset, io.SeekStart); err != nil {
		return err
	}
	_, err = io.CopyN(w, src, entry.Size)
	return err
}

// ReadEntry reads one LBX block into memory.
func (f *File) ReadEntry(index int) ([]byte, error) {
	if index < 0 || index >= len(f.Entries) {
		return nil, fmt.Errorf("entry index %d out of range [0,%d)", index, len(f.Entries))
	}
	entry := f.Entries[index]
	data := make([]byte, entry.Size)

	src, err := os.Open(f.Path)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	if _, err := src.ReadAt(data, entry.Offset); err != nil {
		return nil, err
	}
	return data, nil
}

// EntrySHA256 returns a stable content hash for one block.
func (f *File) EntrySHA256(index int) (string, error) {
	h := sha256.New()
	if err := f.CopyEntry(index, h); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
