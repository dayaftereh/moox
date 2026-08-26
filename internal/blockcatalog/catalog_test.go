package blockcatalog

import (
	"encoding/binary"
	"testing"
)

func TestClassifyEvidenceBasedTypes(t *testing.T) {
	fixed := make([]byte, 104)
	binary.LittleEndian.PutUint16(fixed[0:2], 1)
	binary.LittleEndian.PutUint16(fixed[2:4], 100)
	copy(fixed[4:], []byte("Type ORION2 to run the game!"))
	if got := classify("JIMTEXT.LBX", 0, fixed).Class; got != "fixed_record_v1" {
		t.Fatalf("fixed class=%q", got)
	}

	wave := []byte("RIFF\x00\x00\x00\x00WAVEfmt ")
	if got := classify("ANY.LBX", 1, wave).Class; got != "riff_wave" {
		t.Fatalf("wave class=%q", got)
	}

	palette := make([]byte, 12544)
	if got := classify("FONTS.LBX", 1, palette).Class; got != "external_palette" {
		t.Fatalf("palette class=%q", got)
	}

	array := make([]byte, 34)
	binary.LittleEndian.PutUint16(array[0:2], 3)
	binary.LittleEndian.PutUint16(array[2:4], 10)
	if got := classify("ANY.LBX", 2, array).Class; got != "fixed_array_v1" {
		t.Fatalf("array class=%q", got)
	}

	if got := classify("FONTS.LBX", 0, make([]byte, 100)).Class; got != "font_data" {
		t.Fatalf("font class=%q", got)
	}

	slots := make([]byte, 80)
	copy(slots[0:20], []byte("ALPHA"))
	copy(slots[20:40], []byte("BETA"))
	if got := classify("ANY.LBX", 3, slots).Class; got != "fixed_ascii_slots_candidate" {
		t.Fatalf("slot class=%q", got)
	}

	if got := classify("ANY.LBX", 4, []byte("cats rule dogs drool!!")).Class; got != "ascii_blob_candidate" {
		t.Fatalf("ascii blob class=%q", got)
	}

	signed := []byte{0xfc, 3, 6, 0xfd, 4, 7, 0xfd, 3, 6, 0xfd, 3, 6, 0xfc, 5, 8, 0xfe, 3, 7}
	if got := classify("ANY.LBX", 5, signed).Class; got != "signed_byte_table_candidate" {
		t.Fatalf("signed table class=%q", got)
	}

	if got := classify("ANY.LBX", 0, nil).Class; got != "empty" {
		t.Fatalf("empty class=%q", got)
	}
}
