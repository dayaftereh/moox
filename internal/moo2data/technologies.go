package moo2data

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"moox/internal/i18n"
	"moox/internal/lbx"
	"moox/internal/ruleset"
	"moox/internal/textscan"
)

const (
	technologyNamesSourceID      = "moo2-1.31-techname-block0-technologies"
	technologyTableSourceID      = "moo2-1.31-orion2-technology-table"
	technologyFieldsSourceID     = "moo2-1.31-orion2-technology-fields"
	newGameFieldsSourceID        = "moo2-1.31-orion2-new-game-techfields"
	hyperAdvancedSourceID        = "moo2-1.31-orion2-hyper-advanced-research"
	technologyAIResearchSourceID = "moo2-1.31-orion2-technology-ai-research"
	technologyCount              = 203
	technologyTableOffset        = 0x1FC720
	technologyRecordSize         = 13
	technologyFieldOffset        = 0x1FBFB5
	technologyFieldSize          = 23
	technologyFieldCount         = 82
	newGameFieldsOffset          = 0x1FF7B0
	technologyAIClassTableOffset = 0x1FB82A
	technologyAIClassCount       = 41
	technologyAIFieldGroupOffset = 0x201CA0
	technologyAIFieldGroupCount  = 23
)

type TechnologiesBundle struct {
	Rules   *ruleset.TechnologiesFile
	English *i18n.File
}

func DecodeTechnologies(installationRoot string) (*TechnologiesBundle, error) {
	path := filepath.Join(installationRoot, "TECHNAME.LBX")
	archive, err := lbx.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open TECHNAME.LBX: %w", err)
	}
	block, err := archive.ReadEntry(0)
	if err != nil {
		return nil, fmt.Errorf("read TECHNAME.LBX block 0: %w", err)
	}
	fileHash, err := sha256File(path)
	if err != nil {
		return nil, err
	}
	blockSum := sha256.Sum256(block)
	blockIndex := 0

	exePath := filepath.Join(installationRoot, "Orion2.exe")
	exeData, err := os.ReadFile(exePath)
	if err != nil {
		return nil, fmt.Errorf("read Orion2.exe: %w", err)
	}
	exeHash, err := sha256File(exePath)
	if err != nil {
		return nil, err
	}
	technologyFields, stagedStartFields, err := decodeOriginalTechnologyTables(exeData)
	if err != nil {
		return nil, err
	}
	aiResearch, err := decodeOriginalTechnologyAIResearch(exeData)
	if err != nil {
		return nil, err
	}

	allRuns := textscan.ASCII(block, 3)
	techRuns, err := concreteTechnologyRuns(allRuns)
	if err != nil {
		return nil, err
	}

	out := &ruleset.TechnologiesFile{
		SchemaVersion: ruleset.TechnologiesSchemaVersion,
		Ruleset:       "moo2-1.31",
		Sources: []ruleset.Source{
			{
				ID:          technologyNamesSourceID,
				Type:        "original-observed",
				Description: "The 203-entry concrete technology-name section in Master of Orion II 1.31 TECHNAME.LBX block 0, bounded by No Tech and the following Biology section",
				Archive:     "TECHNAME.LBX",
				Block:       &blockIndex,
				SHA256:      fileHash,
				BlockSHA256: hex.EncodeToString(blockSum[:]),
			},
			{
				ID:          technologyTableSourceID,
				Type:        "original-observed",
				Description: "Orion2.exe 1.31 packed concrete technology table: 203 records x 13 bytes at file offset 0x1FC720; uint16 +0 is technology field id and byte +5 is the strategic-combat availability flag",
				Archive:     "Orion2.exe",
				SHA256:      exeHash,
			},
			{
				ID:          technologyFieldsSourceID,
				Type:        "original-observed",
				Description: "Orion2.exe 1.31 technology-field table: fields 1..82, 23-byte records at file offset 0x1FBFB5; previous/next field ids, RP cost and AI group",
				Archive:     "Orion2.exe",
				SHA256:      exeHash,
			},
			{
				ID:          newGameFieldsSourceID,
				Type:        "original-observed",
				Description: "Orion2.exe 1.31 six-entry uint16 staged new-game tech-field list at file offset 0x1FF7B0: 29,55,22,57,28,23; field-0 always-known behavior is cross-checked against classic reverse-engineering documentation",
				Archive:     "Orion2.exe",
				SHA256:      exeHash,
			},
			{
				ID:          technologyAIResearchSourceID,
				Type:        "original-observed",
				Description: "Orion2.exe 1.31 technology AI/research metadata: technology record byte +3 is the AI class; class base-weight/competition pairs are at file offset 0x1FB82A and the 23 field-group progression values are at file offset 0x201CA0",
				Archive:     "Orion2.exe",
				SHA256:      exeHash,
			},
			{
				ID:          hyperAdvancedSourceID,
				Type:        "original-observed",
				Description: "Orion2.exe 1.31 Player_Research_Cost_ at object-1 relative offset 0xD1E96 / VA 0xE1E96 uses fields 75..82 counters at player+0x21C..0x223 and adds counter*10000 RP; Give_Player_Field_ increments the selected counter on Hyper-Advanced breakthrough",
				Archive:     "Orion2.exe",
				SHA256:      exeHash,
			},
		},
		HyperAdvanced: ruleset.HyperAdvancedResearch{
			TechFieldIDs:    []int{75, 76, 77, 78, 79, 80, 81, 82},
			CostIncrementRP: 10000,
			Verification:    "original-exe-player-research-cost-hyper-counter-runtime",
			Source:          ruleset.FieldProvenance{SourceID: hyperAdvancedSourceID, Offset: intPtr(0xD1E96)},
		},
		NewGameStart: ruleset.NewGameTechnologyStart{
			AlwaysKnownTechFieldID:  0,
			StagedKnownTechFieldIDs: stagedStartFields,
			Verification:            "original-exe-staged-list-plus-classic-field0-invariant",
			Source:                  ruleset.FieldProvenance{SourceID: newGameFieldsSourceID, Offset: intPtr(newGameFieldsOffset)},
		},
		AIResearch: aiResearch,
		Fields:     technologyFields,
	}
	english := &i18n.File{
		SchemaVersion: i18n.SchemaVersion,
		Locale:        "en",
		Sources: []i18n.Source{{
			ID:          technologyNamesSourceID,
			Type:        "original-observed",
			Description: "Concrete technology names from TECHNAME.LBX block 0",
			Archive:     "TECHNAME.LBX",
			Block:       &blockIndex,
			SHA256:      fileHash,
			BlockSHA256: hex.EncodeToString(blockSum[:]),
		}},
		Strings: make(map[string]string, technologyCount),
	}

	for order, run := range techRuns {
		id := stableTechnologyID(run.Value)
		if id == "" {
			return nil, fmt.Errorf("technology %d name %q produced empty stable id", order+1, run.Value)
		}
		nameKey := "technology." + id + ".name"
		offset := run.Offset
		fieldOffset := technologyTableOffset + order*technologyRecordSize
		out.Technologies = append(out.Technologies, ruleset.Technology{
			ID:                       id,
			Order:                    order,
			TechnologyID:             order + 1,
			TechFieldID:              decodeTechnologyFieldID(exeData[fieldOffset]),
			AIClass:                  int(exeData[fieldOffset+3]),
			AIClassSource:            ruleset.FieldProvenance{SourceID: technologyAIResearchSourceID, Offset: intPtr(fieldOffset + 3)},
			TechFieldSource:          ruleset.FieldProvenance{SourceID: technologyTableSourceID, Offset: intPtr(fieldOffset)},
			StrategicCombatAvailable: exeData[fieldOffset+5] != 0,
			StrategicCombatSource:    ruleset.FieldProvenance{SourceID: technologyTableSourceID, Offset: intPtr(fieldOffset + 5)},
			NameKey:                  nameKey,
			NameSource:               ruleset.FieldProvenance{SourceID: technologyNamesSourceID, Offset: &offset},
		})
		english.Strings[nameKey] = run.Value
	}

	if err := out.Validate(); err != nil {
		return nil, err
	}
	if err := english.Validate(); err != nil {
		return nil, err
	}
	return &TechnologiesBundle{Rules: out, English: english}, nil
}

func decodeOriginalTechnologyAIResearch(exeData []byte) (ruleset.TechnologyAIResearch, error) {
	classEnd := technologyAIClassTableOffset + technologyAIClassCount*2
	fieldGroupEnd := technologyAIFieldGroupOffset + technologyAIFieldGroupCount*4
	if len(exeData) < classEnd || len(exeData) < fieldGroupEnd {
		return ruleset.TechnologyAIResearch{}, fmt.Errorf("Orion2.exe is too short for technology AI research tables")
	}
	classes := make([]ruleset.TechnologyAIClass, technologyAIClassCount)
	for classID := 0; classID < technologyAIClassCount; classID++ {
		offset := technologyAIClassTableOffset + classID*2
		classes[classID] = ruleset.TechnologyAIClass{
			ClassID:              classID,
			BaseWeight:           int(exeData[offset]),
			CompetitionSensitive: exeData[offset+1] == 0,
		}
	}
	fieldGroups := make([]int, technologyAIFieldGroupCount)
	for i := range fieldGroups {
		offset := technologyAIFieldGroupOffset + i*4
		fieldGroups[i] = int(binary.LittleEndian.Uint32(exeData[offset : offset+4]))
	}
	return ruleset.TechnologyAIResearch{
		TechnologyClasses: classes,
		FieldGroupValues:  fieldGroups,
		Verification:      "original-exe-calc-tech-value-and-choose-tech-application",
		Source:            ruleset.FieldProvenance{SourceID: technologyAIResearchSourceID, Offset: intPtr(technologyAIClassTableOffset)},
	}, nil
}

func decodeOriginalTechnologyTables(exeData []byte) ([]ruleset.TechnologyField, []int, error) {
	techEnd := technologyTableOffset + technologyCount*technologyRecordSize
	fieldEnd := technologyFieldOffset + technologyFieldCount*technologyFieldSize
	startEnd := newGameFieldsOffset + 6*2
	if len(exeData) < techEnd || len(exeData) < fieldEnd || len(exeData) < startEnd {
		return nil, nil, fmt.Errorf("Orion2.exe is too short for normalized technology tables")
	}

	fields := make([]ruleset.TechnologyField, 0, technologyFieldCount)
	for i := 0; i < technologyFieldCount; i++ {
		offset := technologyFieldOffset + i*technologyFieldSize
		record := exeData[offset : offset+technologyFieldSize]
		fieldID := i + 1
		sequenceMarker := int(binary.LittleEndian.Uint16(record[21:23]))
		if sequenceMarker != fieldID+1 && fieldID != technologyFieldCount {
			return nil, nil, fmt.Errorf("technology field record %d carries sequence marker %d", fieldID, sequenceMarker)
		}
		fields = append(fields, ruleset.TechnologyField{
			FieldID:      fieldID,
			PreviousID:   int(binary.LittleEndian.Uint16(record[0:2])),
			NextID:       int(binary.LittleEndian.Uint16(record[2:4])),
			ResearchCost: int(binary.LittleEndian.Uint32(record[12:16])),
			AIGroup:      int(record[16]),
			Source:       ruleset.FieldProvenance{SourceID: technologyFieldsSourceID, Offset: intPtr(offset)},
		})
	}
	applyTechnologyFieldCategories(fields)

	staged := make([]int, 6)
	for i := range staged {
		staged[i] = int(binary.LittleEndian.Uint16(exeData[newGameFieldsOffset+i*2 : newGameFieldsOffset+i*2+2]))
	}
	want := []int{29, 55, 22, 57, 28, 23}
	for i := range want {
		if staged[i] != want[i] {
			return nil, nil, fmt.Errorf("new-game staged tech field[%d]=%d want=%d", i, staged[i], want[i])
		}
	}
	return fields, staged, nil
}

type technologyResearchCategory struct {
	id      string
	order   int
	nameKey string
	fields  []int
}

var technologyResearchCategories = []technologyResearchCategory{
	{id: "construction", order: 0, nameKey: "research.category.construction", fields: []int{29, 4, 3, 21, 20, 62, 63, 19, 8, 11, 67, 42, 58, 78}},
	{id: "chemistry", order: 1, nameKey: "research.category.chemistry", fields: []int{22, 9, 2, 47, 53, 50, 48, 80}},
	{id: "computers", order: 2, nameKey: "research.category.computers", fields: []int{28, 56, 15, 60, 14, 25, 24, 33, 49, 81}},
	{id: "physics", order: 3, nameKey: "research.category.physics", fields: []int{57, 31, 66, 54, 16, 65, 52, 59, 51, 39, 69, 77}},
	{id: "power", order: 4, nameKey: "research.category.power", fields: []int{55, 23, 5, 41, 13, 46, 37, 38, 40, 76}},
	{id: "sociology", order: 5, nameKey: "research.category.sociology", fields: []int{10, 73, 43, 12, 6, 32, 82}},
	{id: "biology", order: 6, nameKey: "research.category.biology", fields: []int{18, 1, 34, 35, 44, 30, 17, 70, 75}},
	{id: "force_fields", order: 7, nameKey: "research.category.force_fields", fields: []int{7, 36, 45, 27, 72, 64, 26, 61, 71, 68, 79}},
}

// applyTechnologyFieldCategories adds the stable eight player-facing research
// chains used by the strategic HMI. The mapping is deliberately keyed by the
// original TechField IDs rather than inferred from a synthetic previous/next
// graph, so the raw-data decoder remains deterministic in extractor fixtures.
// Field 74 is the original non-research sentinel and intentionally has no
// player-facing category.
func applyTechnologyFieldCategories(fields []ruleset.TechnologyField) {
	for _, category := range technologyResearchCategories {
		for _, fieldID := range category.fields {
			if fieldID <= 0 || fieldID > len(fields) {
				continue
			}
			field := &fields[fieldID-1]
			field.CategoryID = category.id
			field.CategoryOrder = category.order
			field.CategoryNameKey = category.nameKey
		}
	}
}
func concreteTechnologyRuns(runs []textscan.String) ([]textscan.String, error) {
	start := -1
	for i, run := range runs {
		if run.Value == "Achilles Targeting Unit" {
			start = i
			break
		}
	}
	if start < 1 || runs[start-1].Value != "No Tech" {
		return nil, fmt.Errorf("TECHNAME block 0 technology section start not found after No Tech")
	}
	end := start + technologyCount - 1
	if end >= len(runs) || runs[end].Value != "Zortrium Armor" {
		return nil, fmt.Errorf("TECHNAME block 0 technology section end mismatch")
	}
	if end+1 >= len(runs) || runs[end+1].Value != "Biology" {
		return nil, fmt.Errorf("TECHNAME block 0 expected Biology after technology section")
	}
	return runs[start : end+1], nil
}

func stableTechnologyID(name string) string {
	var b strings.Builder
	underscore := false
	for _, r := range strings.ToLower(name) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
			underscore = false
			continue
		}
		if b.Len() > 0 && !underscore {
			b.WriteByte('_')
			underscore = true
		}
	}
	return strings.Trim(b.String(), "_")
}

func decodeTechnologyFieldID(value byte) int {
	if value == 0xFF {
		return -1
	}
	return int(value)
}
