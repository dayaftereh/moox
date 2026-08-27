package moo2data

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"path/filepath"

	"moox/internal/i18n"
	"moox/internal/lbx"
	"moox/internal/moo2exe"
	"moox/internal/ruleset"
)

const (
	planetClassNamesSourceID        = "moo2-1.31-estrings-planet-class-names"
	planetClassStringMapSourceID    = "moo2-1.31-orion2-planet-class-string-map"
	planetSizeGenerationSourceID    = "moo2-1.31-orion2-planet-size-generation"
	planetMineralExtractionSourceID = "moo2-1.31-orion2-mineral-extraction"
	planetFoodPerFarmerSourceID     = "moo2-1.31-orion2-food-per-farmer"

	planetClassStringsCodeSHA256    = "576b1282664e03a2c86b994aa43bcd1002041e39f4f6b973cf5a5082eb3962b2"
	planetClimateStringsCodeSHA256  = "6014154c11d3064f71b0d2449bec190689ff7f819f5fc1c2ed6efc17b88e7364"
	generateSizeCodeSHA256          = "fd62f903bd030531115d3e68e282473d9c91f8859c18a48a4822e5e9263ef13b"
	getMineralsExtractedCodeSHA256  = "22afc8196d5ec3202691a4da50cf896a4595b241b14ddc9188144fd877324572"
	generateFoodPerFarmerCodeSHA256 = "c7eb7c6d8427eb9451723ec5ac50b06385a1ad6f728d431b8856217ed826f00b"
)

type PlanetClassesBundle struct {
	Rules   *ruleset.PlanetClassesFile
	English *i18n.File
}

type planetClassEvidence struct {
	ExecutableSHA256  string
	SizeThresholds    []int
	MineralExtraction []int
	FoodPerFarmer     []int
}

type planetClassSpec struct {
	ID           string
	Name         string
	EStringIndex int
}

type packedEString struct {
	Value  string
	Offset int
}

var planetSizeSpecs = []planetClassSpec{
	{ID: "tiny", Name: "Tiny", EStringIndex: 0x2AB},
	{ID: "small", Name: "Small", EStringIndex: 0x1E0},
	{ID: "medium", Name: "Medium", EStringIndex: 0x173},
	{ID: "large", Name: "Large", EStringIndex: 0x168},
	{ID: "huge", Name: "Huge", EStringIndex: 0x143},
}

var mineralClassSpecs = []planetClassSpec{
	{ID: "ultra_poor", Name: "Ultra Poor", EStringIndex: 0x2AC},
	{ID: "poor", Name: "Poor", EStringIndex: 0x1A7},
	{ID: "abundant", Name: "Abundant", EStringIndex: 0x2AD},
	{ID: "rich", Name: "Rich", EStringIndex: 0x1C2},
	{ID: "ultra_rich", Name: "Ultra Rich", EStringIndex: 0x2AE},
}

var gravityClassSpecs = []planetClassSpec{
	{ID: "low_g", Name: "Low G", EStringIndex: 0x2AF},
	{ID: "normal_g", Name: "Normal G", EStringIndex: 0x2B0},
	{ID: "heavy_g", Name: "Heavy G", EStringIndex: 0x2B1},
}

var planetClimateSpecs = []planetClassSpec{
	{ID: "toxic", Name: "Toxic", EStringIndex: 0x21B},
	{ID: "radiated", Name: "Radiated", EStringIndex: 0x2CF},
	{ID: "barren", Name: "Barren", EStringIndex: 0x2D0},
	{ID: "desert", Name: "Desert", EStringIndex: 0x2D1},
	{ID: "tundra", Name: "Tundra", EStringIndex: 0x2D2},
	{ID: "ocean", Name: "Ocean", EStringIndex: 0x18F},
	{ID: "swamp", Name: "Swamp", EStringIndex: 0x1F5},
	{ID: "arid", Name: "Arid", EStringIndex: 0x0B8},
	{ID: "terran", Name: "Terran", EStringIndex: 0x2D3},
	{ID: "gaia", Name: "Gaia", EStringIndex: 0x12F},
}

func DecodePlanetClasses(installationRoot string) (*PlanetClassesBundle, error) {
	evidence, err := verifyOriginalPlanetClassEvidence(installationRoot)
	if err != nil {
		return nil, err
	}
	return decodePlanetClassesWithEvidence(installationRoot, evidence)
}

func decodePlanetClassesWithEvidence(installationRoot string, evidence planetClassEvidence) (*PlanetClassesBundle, error) {
	if len(evidence.SizeThresholds) != 5 || len(evidence.MineralExtraction) != 5 || len(evidence.FoodPerFarmer) != 10 {
		return nil, fmt.Errorf("planet evidence counts size/mineral/food=%d/%d/%d, expected 5/5/10", len(evidence.SizeThresholds), len(evidence.MineralExtraction), len(evidence.FoodPerFarmer))
	}

	estringsPath := filepath.Join(installationRoot, "ESTRINGS.LBX")
	archive, err := lbx.Open(estringsPath)
	if err != nil {
		return nil, fmt.Errorf("open ESTRINGS.LBX: %w", err)
	}
	if len(archive.Entries) != 1 {
		return nil, fmt.Errorf("ESTRINGS.LBX has %d blocks, expected 1", len(archive.Entries))
	}
	block, err := archive.ReadEntry(0)
	if err != nil {
		return nil, fmt.Errorf("read ESTRINGS.LBX block 0: %w", err)
	}
	strings, err := parseEStringsBlock(block)
	if err != nil {
		return nil, fmt.Errorf("parse ESTRINGS.LBX block 0: %w", err)
	}
	if len(strings) < 0x32C {
		return nil, fmt.Errorf("ESTRINGS.LBX block 0 has %d packed strings, expected at least 812", len(strings))
	}
	fileHash, err := sha256File(estringsPath)
	if err != nil {
		return nil, err
	}
	blockSum := sha256.Sum256(block)
	blockIndex := 0

	out := &ruleset.PlanetClassesFile{
		SchemaVersion: ruleset.PlanetClassesSchemaVersion,
		Ruleset:       "moo2-1.31",
		Sources: []ruleset.Source{
			{
				ID:          planetClassNamesSourceID,
				Type:        "original-observed",
				Description: "Canonical planet size, mineral, gravity and climate names resolved from the packed English ESTRINGS.LBX block 0 string indices used by the original runtime arrays",
				Archive:     "ESTRINGS.LBX",
				Block:       &blockIndex,
				SHA256:      fileHash,
				BlockSHA256: hex.EncodeToString(blockSum[:]),
			},
			{
				ID:          planetClassStringMapSourceID,
				Type:        "original-observed",
				Description: "Orion2.exe 1.31 Load_H_And_Pro_File_Strings_ assigns explicit E_Strings_ indices to _planet_size_string, _mineral_class_string, _planet_gravity_string and _planet_climate_string; assignment code ranges are SHA-256 verified",
				Archive:     "Orion2.exe",
				SHA256:      evidence.ExecutableSHA256,
			},
			{
				ID:          planetSizeGenerationSourceID,
				Type:        "original-observed",
				Description: "Orion2.exe 1.31 Generate_Size_ rolls Random_(10) and selects the first size whose _planet_size_table threshold is >= the roll; object 2 offset 0x57F7 begins 1,3,7,9,10",
				Archive:     "Orion2.exe",
				SHA256:      evidence.ExecutableSHA256,
			},
			{
				ID:          planetMineralExtractionSourceID,
				Type:        "original-observed",
				Description: "Orion2.exe 1.31 Get_Minerals_Extracted_ returns mineral_class+1 for classes 0..4; object 2 _minerals_extracted_table at 0x5812 also contains 1,2,3,4,5",
				Archive:     "Orion2.exe",
				SHA256:      evidence.ExecutableSHA256,
			},
			{
				ID:          planetFoodPerFarmerSourceID,
				Type:        "original-observed",
				Description: "Orion2.exe 1.31 Generate_Food_Per_Farmer_ indexes _food_per_farmer_table at object 2 offset 0x581C by climate; values are 0,0,0,1,1,2,2,1,2,3",
				Archive:     "Orion2.exe",
				SHA256:      evidence.ExecutableSHA256,
			},
		},
	}
	english := &i18n.File{
		SchemaVersion: i18n.SchemaVersion,
		Locale:        "en",
		Sources: []i18n.Source{{
			ID:          planetClassNamesSourceID,
			Type:        "original-observed",
			Description: "Planet class names resolved from ESTRINGS.LBX block 0 through original runtime string indices",
			Archive:     "ESTRINGS.LBX",
			Block:       &blockIndex,
			SHA256:      fileHash,
			BlockSHA256: hex.EncodeToString(blockSum[:]),
		}},
		Strings: make(map[string]string, 23),
	}

	for index, spec := range planetSizeSpecs {
		entry, err := checkedPlanetEString(strings, spec)
		if err != nil {
			return nil, err
		}
		key := "planet_size." + spec.ID + ".name"
		offset := 0x57F7 + index
		out.Sizes = append(out.Sizes, ruleset.PlanetSize{
			ID:                           spec.ID,
			Index:                        index,
			NameKey:                      key,
			NameSource:                   ruleset.FieldProvenance{SourceID: planetClassNamesSourceID, Offset: intPtr(entry.Offset)},
			GenerationRollUpperThreshold: evidence.SizeThresholds[index],
			GenerationSource:             ruleset.FieldProvenance{SourceID: planetSizeGenerationSourceID, Offset: &offset},
		})
		english.Strings[key] = spec.Name
	}
	for index, spec := range mineralClassSpecs {
		entry, err := checkedPlanetEString(strings, spec)
		if err != nil {
			return nil, err
		}
		key := "mineral_class." + spec.ID + ".name"
		offset := 0x5812 + index
		out.MineralClasses = append(out.MineralClasses, ruleset.MineralClass{
			ID:               spec.ID,
			Index:            index,
			NameKey:          key,
			NameSource:       ruleset.FieldProvenance{SourceID: planetClassNamesSourceID, Offset: intPtr(entry.Offset)},
			BaseExtraction:   evidence.MineralExtraction[index],
			ExtractionSource: ruleset.FieldProvenance{SourceID: planetMineralExtractionSourceID, Offset: &offset},
		})
		english.Strings[key] = spec.Name
	}
	for index, spec := range gravityClassSpecs {
		entry, err := checkedPlanetEString(strings, spec)
		if err != nil {
			return nil, err
		}
		key := "gravity_class." + spec.ID + ".name"
		out.GravityClasses = append(out.GravityClasses, ruleset.GravityClass{
			ID:         spec.ID,
			Index:      index,
			NameKey:    key,
			NameSource: ruleset.FieldProvenance{SourceID: planetClassNamesSourceID, Offset: intPtr(entry.Offset)},
		})
		english.Strings[key] = spec.Name
	}
	for index, spec := range planetClimateSpecs {
		entry, err := checkedPlanetEString(strings, spec)
		if err != nil {
			return nil, err
		}
		key := "planet_climate." + spec.ID + ".name"
		offset := 0x581C + index
		out.Climates = append(out.Climates, ruleset.PlanetClimate{
			ID:                spec.ID,
			Index:             index,
			NameKey:           key,
			NameSource:        ruleset.FieldProvenance{SourceID: planetClassNamesSourceID, Offset: intPtr(entry.Offset)},
			BaseFoodPerFarmer: evidence.FoodPerFarmer[index],
			FoodSource:        ruleset.FieldProvenance{SourceID: planetFoodPerFarmerSourceID, Offset: &offset},
		})
		english.Strings[key] = spec.Name
	}

	if err := out.Validate(); err != nil {
		return nil, err
	}
	if err := english.Validate(); err != nil {
		return nil, err
	}
	return &PlanetClassesBundle{Rules: out, English: english}, nil
}

func verifyOriginalPlanetClassEvidence(installationRoot string) (planetClassEvidence, error) {
	var evidence planetClassEvidence
	exe, err := moo2exe.Open(filepath.Join(installationRoot, "Orion2.exe"))
	if err != nil {
		return evidence, fmt.Errorf("open Orion2.exe for planet classes: %w", err)
	}
	checks := []struct {
		Name   string
		Offset int
		Size   int
		SHA256 string
	}{
		{Name: "planet size/mineral/gravity string assignments", Offset: 0xBE51D, Size: 0xC3, SHA256: planetClassStringsCodeSHA256},
		{Name: "planet climate string assignments", Offset: 0xBE793, Size: 0x96, SHA256: planetClimateStringsCodeSHA256},
		{Name: "Generate_Size_", Offset: 0x7C7D1, Size: 0x36, SHA256: generateSizeCodeSHA256},
		{Name: "Get_Minerals_Extracted_", Offset: 0x7CD40, Size: 0x15, SHA256: getMineralsExtractedCodeSHA256},
		{Name: "Generate_Food_Per_Farmer_", Offset: 0x7BFA3, Size: 0x3D, SHA256: generateFoodPerFarmerCodeSHA256},
	}
	for _, check := range checks {
		code, err := exe.ReadObject(1, check.Offset, check.Size)
		if err != nil {
			return evidence, fmt.Errorf("read Orion2.exe %s: %w", check.Name, err)
		}
		sum := sha256.Sum256(code)
		if got := hex.EncodeToString(sum[:]); got != check.SHA256 {
			return evidence, fmt.Errorf("Orion2.exe %s hash=%s, expected %s", check.Name, got, check.SHA256)
		}
	}

	readBytes := func(offset, size int) ([]byte, error) {
		data, err := exe.ReadObject(2, offset, size)
		if err != nil {
			return nil, fmt.Errorf("read Orion2.exe object 2 offset 0x%X: %w", offset, err)
		}
		return data, nil
	}
	sizeThresholds, err := readBytes(0x57F7, 5)
	if err != nil {
		return evidence, err
	}
	minerals, err := readBytes(0x5812, 5)
	if err != nil {
		return evidence, err
	}
	food, err := readBytes(0x581C, 10)
	if err != nil {
		return evidence, err
	}
	evidence.ExecutableSHA256 = exe.SHA256()
	evidence.SizeThresholds = bytesToInts(sizeThresholds)
	evidence.MineralExtraction = bytesToInts(minerals)
	evidence.FoodPerFarmer = bytesToInts(food)
	return evidence, nil
}

func parseEStringsBlock(block []byte) ([]packedEString, error) {
	if len(block) < 4 {
		return nil, fmt.Errorf("block is only %d bytes", len(block))
	}
	if kind := binary.LittleEndian.Uint16(block[0:2]); kind != 1 {
		return nil, fmt.Errorf("packed string type=%d, expected 1", kind)
	}
	payloadSize := int(binary.LittleEndian.Uint16(block[2:4]))
	if payloadSize != len(block)-4 {
		return nil, fmt.Errorf("packed string payload size=%d, actual=%d", payloadSize, len(block)-4)
	}
	var out []packedEString
	for pos := 4; pos < len(block); {
		end := bytes.IndexByte(block[pos:], 0)
		if end < 0 {
			return nil, fmt.Errorf("unterminated packed string at block offset 0x%X", pos)
		}
		out = append(out, packedEString{Value: string(block[pos : pos+end]), Offset: pos})
		pos += end + 1
	}
	return out, nil
}

func checkedPlanetEString(strings []packedEString, spec planetClassSpec) (packedEString, error) {
	if spec.EStringIndex < 0 || spec.EStringIndex >= len(strings) {
		return packedEString{}, fmt.Errorf("planet class %s E string index 0x%X outside %d strings", spec.ID, spec.EStringIndex, len(strings))
	}
	entry := strings[spec.EStringIndex]
	if entry.Value != spec.Name {
		return packedEString{}, fmt.Errorf("planet class %s E string 0x%X=%q, expected %q", spec.ID, spec.EStringIndex, entry.Value, spec.Name)
	}
	return entry, nil
}

func bytesToInts(data []byte) []int {
	out := make([]int, len(data))
	for index, value := range data {
		out[index] = int(value)
	}
	return out
}

func intPtr(value int) *int { return &value }
