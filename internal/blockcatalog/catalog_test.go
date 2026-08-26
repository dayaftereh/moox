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

	if got := classify("ANY.LBX", 0, nil).Class; got != "empty" {
		t.Fatalf("empty class=%q", got)
	}
}
