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

func TestCombatShipPaletteCarrierMapping(t *testing.T) {
	tests := map[int]int{
		0: 44, 43: 44,
		45: 89, 88: 89,
		90: 134, 133: 134,
		135: 179, 178: 179,
		180: 224, 223: 224,
		225: 269, 268: 269,
		270: 314, 313: 314,
		315: 359, 358: 359,
	}
	for block, want := range tests {
		if got := combatShipPaletteCarrier(block); got != want {
			t.Fatalf("block %d carrier=%d, want %d", block, got, want)
		}
	}
	for _, block := range []int{44, 89, 134, 179, 224, 269, 314, 359, 360} {
		if got := combatShipPaletteCarrier(block); got != -1 {
			t.Fatalf("block %d carrier=%d, want unresolved/carrier", block, got)
		}
	}
}

func TestCombatSFXPaletteMapping(t *testing.T) {
	tests := map[int]int{
		2: 4, 7: 4,
		8: 1,
		9: 2, 13: 2,
		14: 4, 15: 4,
		16: 2, 39: 2,
		41: 2, 42: 2,
		43: 4, 46: 4,
		48: 4, 51: 4,
		52: 2, 67: 2,
		69: 1, 78: 1,
	}
	for block, want := range tests {
		if got := combatSFXPaletteBlock(block); got != want {
			t.Fatalf("block %d palette=%d, want %d", block, got, want)
		}
	}
	for _, block := range []int{0, 1, 40, 47, 68, 79} {
		if got := combatSFXPaletteBlock(block); got != -1 {
			t.Fatalf("block %d palette=%d, want unresolved", block, got)
		}
	}
}

func TestBeamsPaletteMapping(t *testing.T) {
	tests := []struct {
		block   int
		palette int
		carrier bool
	}{
		{1, 3, false}, {16, 3, false},
		{17, 1, false}, {32, 1, false},
		{33, 3, false}, {48, 3, false},
		{49, 1, false}, {64, 1, false},
		{65, 2, false}, {66, 2, false},
		{67, 4, false}, {68, 4, true}, {69, 4, false},
		{70, 4, true}, {87, 4, true},
		{88, 4, false}, {108, 4, false},
		{109, 4, true}, {129, 4, true},
		{131, 4, true}, {152, 4, true},
	}
	for _, tt := range tests {
		palette, carrier := beamsPaletteRule(tt.block)
		if palette != tt.palette || carrier != tt.carrier {
			t.Fatalf("block %d rule=(%d,%v), want (%d,%v)", tt.block, palette, carrier, tt.palette, tt.carrier)
		}
	}
	for _, block := range []int{0, 130, 153} {
		palette, carrier := beamsPaletteRule(block)
		if palette != -1 || carrier {
			t.Fatalf("block %d rule=(%d,%v), want unresolved", block, palette, carrier)
		}
	}
}

func TestBuffer0PaletteMapping(t *testing.T) {
	for _, block := range []int{1, 12, 15, 91, 112, 121, 132, 136, 142, 287} {
		if !buffer0UsesFonts1(block) {
			t.Fatalf("block %d should use FONTS#1", block)
		}
	}
	for _, block := range []int{0, 13, 14, 92, 111, 122, 131, 137, 141, 288} {
		if buffer0UsesFonts1(block) {
			t.Fatalf("block %d should remain unresolved", block)
		}
	}
}

func TestOfficerPaletteMapping(t *testing.T) {
	tests := map[int]int{0: 1, 209: 1, 210: 2, 276: 2, 277: 4, 343: 4}
	for block, want := range tests {
		if got := officerPaletteBlock(block); got != want {
			t.Fatalf("block %d palette=%d, want %d", block, got, want)
		}
	}
	for _, block := range []int{-1, 344} {
		if got := officerPaletteBlock(block); got != -1 {
			t.Fatalf("block %d palette=%d, want unresolved", block, got)
		}
	}
}

func TestFleetPaletteMapping(t *testing.T) {
	tests := []struct {
		block   int
		palette int
		carrier bool
	}{
		{0, 1, false}, {44, 1, false},
		{45, 1, true}, {81, 1, true},
		{83, 1, true}, {110, 1, true},
	}
	for _, tt := range tests {
		palette, carrier := fleetPaletteRule(tt.block)
		if palette != tt.palette || carrier != tt.carrier {
			t.Fatalf("block %d rule=(%d,%v), want (%d,%v)", tt.block, palette, carrier, tt.palette, tt.carrier)
		}
	}
	for _, block := range []int{82, 111, 112} {
		palette, carrier := fleetPaletteRule(block)
		if palette != -1 || carrier {
			t.Fatalf("block %d rule=(%d,%v), want unresolved", block, palette, carrier)
		}
	}
}

func TestMonsterPaletteMapping(t *testing.T) {
	tests := []struct {
		block   int
		palette int
		carrier int
	}{
		{7, 1, 14}, {8, 1, -1}, {9, 1, 14}, {12, 1, -1},
		{20, 1, -1}, {21, 1, -1}, {24, 1, -1}, {25, 1, 13},
	}
	for _, tt := range tests {
		palette, carrier, ok := monsterPaletteRule(tt.block)
		if !ok || palette != tt.palette || carrier != tt.carrier {
			t.Fatalf("block %d rule=(%d,%d,%v), want (%d,%d,true)", tt.block, palette, carrier, ok, tt.palette, tt.carrier)
		}
	}
	for _, block := range []int{0, 1, 2, 3, 4, 5, 6, 10, 11, 13, 14, 15, 16, 17, 18, 19, 22, 23, 26} {
		_, _, ok := monsterPaletteRule(block)
		if ok {
			t.Fatalf("block %d should remain self/carrier/unresolved", block)
		}
	}
}

func TestMultiGamePaletteCarrierMapping(t *testing.T) {
	tests := map[int]int{
		1: 0, 39: 0,
		40: 42,
		41: 0,
		43: 42, 45: 42,
		46: 0, 149: 0,
		150: 42, 253: 42,
		254: 0, 260: 0,
	}
	for block, want := range tests {
		if got := multiGamePaletteCarrier(block); got != want {
			t.Fatalf("block %d carrier=%d, want %d", block, got, want)
		}
	}
	for _, block := range []int{0, 42, 261} {
		if got := multiGamePaletteCarrier(block); got != -1 {
			t.Fatalf("block %d carrier=%d, want unresolved/self", block, got)
		}
	}
}

func TestWorkshopArchiveWidePaletteRules(t *testing.T) {
	tests := map[string]int{
		"CMBTFGTR.LBX": 4,
		"COLGCBT.LBX":  2,
		"COLROADS.LBX": 2,
		"COLVEGGI.LBX": 2,
		"STARBG.LBX":   1,
	}
	for archive, want := range tests {
		rule, ok := confirmedExternalPaletteRules[archive]
		if !ok || rule.Archive != "FONTS.LBX" || rule.PaletteBlock != want || !rule.ExternalOnly {
			t.Fatalf("archive %s rule=%+v ok=%v, want FONTS#%d external-only", archive, rule, ok, want)
		}
	}
}

func TestCombatPlanetPaletteCarrierMapping(t *testing.T) {
	tests := map[int]int{
		0: 5, 5: 5,
		6: 11, 10: 11, 11: 5,
		12: 17, 16: 17, 17: 5,
		18: 23, 22: 23, 23: 5,
		24: 29, 28: 29, 29: 5,
		30: 35, 34: 35, 35: 5,
		36: 41, 40: 41, 41: 5,
		42: 47, 46: 47, 47: 5,
		48: 53, 52: 53, 53: 5,
		54: 59, 58: 59, 59: 5,
		61: 62, 62: 5,
	}
	for block, want := range tests {
		if got := combatPlanetPaletteCarrier(block); got != want {
			t.Fatalf("block %d carrier=%d, want %d", block, got, want)
		}
	}
	for _, block := range []int{60, 63} {
		if got := combatPlanetPaletteCarrier(block); got != -1 {
			t.Fatalf("block %d carrier=%d, want unresolved", block, got)
		}
	}
}
