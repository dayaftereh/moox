package moo2data

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"path/filepath"

	"moox/internal/i18n"
	"moox/internal/lbx"
	"moox/internal/ruleset"
	"moox/internal/textscan"
)

const (
	raceNamesSourceID = "moo2-1.31-estrings-block0-race-names"
	raceHelpSourceID  = "moo2-1.31-help-block0-race-records"
)

type RacesBundle struct {
	Rules   *ruleset.RacesFile
	English *i18n.File
}

type racePresetSpec struct {
	ID               string
	Name             string
	HelpName         string
	HelpRecord       int
	HelpRecordSHA256 string
	TraitIDs         []string
}

func DecodeRaces(installationRoot string, traits *ruleset.RaceTraitsFile) (*RacesBundle, error) {
	if traits == nil {
		return nil, fmt.Errorf("race traits are required")
	}
	if err := traits.Validate(); err != nil {
		return nil, fmt.Errorf("validate race traits: %w", err)
	}

	nameArchivePath := filepath.Join(installationRoot, "ESTRINGS.LBX")
	nameArchive, err := lbx.Open(nameArchivePath)
	if err != nil {
		return nil, fmt.Errorf("open ESTRINGS.LBX: %w", err)
	}
	nameBlock, err := nameArchive.ReadEntry(0)
	if err != nil {
		return nil, fmt.Errorf("read ESTRINGS.LBX block 0: %w", err)
	}
	nameFileHash, err := sha256File(nameArchivePath)
	if err != nil {
		return nil, err
	}
	nameBlockSum := sha256.Sum256(nameBlock)
	nameBlockIndex := 0

	helpArchivePath := filepath.Join(installationRoot, "HELP.LBX")
	helpArchive, err := lbx.Open(helpArchivePath)
	if err != nil {
		return nil, fmt.Errorf("open HELP.LBX: %w", err)
	}
	helpBlock, err := helpArchive.ReadEntry(0)
	if err != nil {
		return nil, fmt.Errorf("read HELP.LBX block 0: %w", err)
	}
	helpFileHash, err := sha256File(helpArchivePath)
	if err != nil {
		return nil, err
	}
	helpBlockSum := sha256.Sum256(helpBlock)
	helpBlockIndex := 0
	arrayCount, recordSize, err := parseFixedArrayHeader(helpBlock)
	if err != nil {
		return nil, fmt.Errorf("HELP.LBX block 0: %w", err)
	}

	traitCosts := make(map[string]int)
	for _, group := range traits.Groups {
		for _, option := range group.Options {
			traitCosts[option.ID] = option.PickCost
		}
	}

	specs := racePresetSpecs()
	out := &ruleset.RacesFile{
		SchemaVersion: ruleset.RacesSchemaVersion,
		Ruleset:       "moo2-1.31",
		Sources: []ruleset.Source{
			{
				ID:          raceNamesSourceID,
				Type:        "original-observed",
				Description: "Canonical singular race names observed in Master of Orion II 1.31 ESTRINGS.LBX block 0",
				Archive:     "ESTRINGS.LBX",
				Block:       &nameBlockIndex,
				SHA256:      nameFileHash,
				BlockSHA256: hex.EncodeToString(nameBlockSum[:]),
			},
			{
				ID:          raceHelpSourceID,
				Type:        "original-observed",
				Description: "Preset race help records in Master of Orion II 1.31 HELP.LBX block 0; trait mappings are derived from the behavior stated by those records",
				Archive:     "HELP.LBX",
				Block:       &helpBlockIndex,
				SHA256:      helpFileHash,
				BlockSHA256: hex.EncodeToString(helpBlockSum[:]),
			},
		},
	}

	english := &i18n.File{
		SchemaVersion: i18n.SchemaVersion,
		Locale:        "en",
		Sources: []i18n.Source{
			{
				ID:          raceNamesSourceID,
				Type:        "original-observed",
				Description: "Canonical singular race names observed in Master of Orion II 1.31 ESTRINGS.LBX block 0",
				Archive:     "ESTRINGS.LBX",
				Block:       &nameBlockIndex,
				SHA256:      nameFileHash,
				BlockSHA256: hex.EncodeToString(nameBlockSum[:]),
			},
		},
		Strings: make(map[string]string, len(specs)),
	}

	nameRuns := textscan.ASCII(nameBlock, 2)
	for order, spec := range specs {
		nameOffset, err := findExactStringOffset(nameRuns, spec.Name)
		if err != nil {
			return nil, fmt.Errorf("race %s canonical name: %w", spec.ID, err)
		}
		if spec.HelpRecord < 0 || spec.HelpRecord >= arrayCount {
			return nil, fmt.Errorf("race %s HELP record %d outside [0,%d)", spec.ID, spec.HelpRecord, arrayCount)
		}
		recordOffset := 4 + spec.HelpRecord*recordSize
		record := helpBlock[recordOffset : recordOffset+recordSize]
		recordSum := sha256.Sum256(record)
		recordHash := hex.EncodeToString(recordSum[:])
		if recordHash != spec.HelpRecordSHA256 {
			return nil, fmt.Errorf("race %s HELP record %d hash=%s, expected %s", spec.ID, spec.HelpRecord, recordHash, spec.HelpRecordSHA256)
		}
		runs := textscan.ASCII(record, 4)
		if len(runs) < 2 || runs[0].Value != spec.HelpName {
			return nil, fmt.Errorf("race %s HELP record %d does not contain expected heading %q", spec.ID, spec.HelpRecord, spec.HelpName)
		}

		selections := make([]ruleset.RaceTraitSelection, 0, len(spec.TraitIDs))
		pickTotal := 0
		for _, traitID := range spec.TraitIDs {
			cost, ok := traitCosts[traitID]
			if !ok {
				return nil, fmt.Errorf("race %s maps to unknown trait %q", spec.ID, traitID)
			}
			pickTotal += cost
			selections = append(selections, ruleset.RaceTraitSelection{TraitID: traitID, Verification: "help-description-derived"})
		}
		nameOffsetCopy := nameOffset
		nameKey := "race." + spec.ID + ".name"
		out.Races = append(out.Races, ruleset.Race{
			ID:               spec.ID,
			Order:            order,
			NameKey:          nameKey,
			NameSource:       ruleset.FieldProvenance{SourceID: raceNamesSourceID, Offset: &nameOffsetCopy},
			TraitSelections:  selections,
			DerivedPickTotal: pickTotal,
			PortraitAssetKey: "race." + spec.ID + ".portrait",
			IconAssetKey:     "race." + spec.ID + ".icon",
			Source: ruleset.RaceSource{
				SourceID:     raceHelpSourceID,
				RecordIndex:  spec.HelpRecord,
				RecordOffset: recordOffset,
				RecordSize:   recordSize,
				RecordSHA256: recordHash,
			},
		})
		english.Strings[nameKey] = spec.Name
	}

	if err := out.ValidateAgainstTraits(traits); err != nil {
		return nil, err
	}
	if err := english.Validate(); err != nil {
		return nil, err
	}
	return &RacesBundle{Rules: out, English: english}, nil
}

func parseFixedArrayHeader(data []byte) (count, recordSize int, err error) {
	if len(data) < 4 {
		return 0, 0, fmt.Errorf("fixed array too small")
	}
	count = int(binary.LittleEndian.Uint16(data[0:2]))
	recordSize = int(binary.LittleEndian.Uint16(data[2:4]))
	if count <= 0 || recordSize <= 0 || 4+count*recordSize != len(data) {
		return 0, 0, fmt.Errorf("invalid fixed array count=%d record_size=%d block_size=%d", count, recordSize, len(data))
	}
	return count, recordSize, nil
}

func findExactStringOffset(runs []textscan.String, value string) (int, error) {
	found := -1
	for _, run := range runs {
		if run.Value != value {
			continue
		}
		if found >= 0 {
			return 0, fmt.Errorf("string %q occurs more than once", value)
		}
		found = run.Offset
	}
	if found < 0 {
		return 0, fmt.Errorf("string %q not found", value)
	}
	return found, nil
}

func racePresetSpecs() []racePresetSpec {
	return []racePresetSpec{
		{ID: "alkari", Name: "Alkari", HelpName: "Alkari", HelpRecord: 630, HelpRecordSHA256: "bed6797fe5a0b50a0b391c0f7bd4a1116d98c2af4e40a6ee5bf416390848c048", TraitIDs: []string{"ship_defense_plus_50", "artifacts_world", "government_dictatorship"}},
		{ID: "bulrathi", Name: "Bulrathi", HelpName: "Bulrathi", HelpRecord: 631, HelpRecordSHA256: "88440426c91719ed003ebcff6e5b1fff2d682298c2c367f67cd9d072b79762a8", TraitIDs: []string{"high_g_world", "ship_attack_plus_20", "ground_combat_plus_10", "government_dictatorship"}},
		{ID: "darlok", Name: "Darlok", HelpName: "Darloks", HelpRecord: 632, HelpRecordSHA256: "5432e500826398aefbb0b389fa41503f6fba6baeaf9319b70f1d64ad9c398539", TraitIDs: []string{"spying_plus_20", "stealthy_ships", "government_dictatorship"}},
		{ID: "elerian", Name: "Elerian", HelpName: "Elerians", HelpRecord: 633, HelpRecordSHA256: "f98747e9b041863f182fc7e51360d61f887b20b19da1fb16a56eb3b2321c7d6c", TraitIDs: []string{"omniscient", "ship_attack_plus_20", "ship_defense_plus_25", "telepathic", "government_feudal"}},
		{ID: "gnolam", Name: "Gnolam", HelpName: "Gnolams", HelpRecord: 634, HelpRecordSHA256: "4985885bed65bb788d2ad1fdf83bb14fbacc3b37e8137d2fdd7d4d48ff1316f3", TraitIDs: []string{"fantastic_traders", "money_plus_1_0", "lucky", "low_g_world", "ground_combat_minus_10", "government_dictatorship"}},
		{ID: "human", Name: "Human", HelpName: "Humans", HelpRecord: 635, HelpRecordSHA256: "ecba58d8308ee39cfbe8d9d3e3593604e47df7d7181b801ea0c2eda7475da56a", TraitIDs: []string{"charismatic", "government_democracy"}},
		{ID: "klackon", Name: "Klackon", HelpName: "Klackons", HelpRecord: 636, HelpRecordSHA256: "4c605cf8b3b51b669692c0190bbf88c6360aa92b0437cbbbf0fbddffd90fcffd", TraitIDs: []string{"government_unification", "farming_plus_1", "industry_plus_1", "uncreative"}},
		{ID: "meklar", Name: "Meklar", HelpName: "Meklars", HelpRecord: 637, HelpRecordSHA256: "8c72a8db87c108a09e870855041ded0ac43dedb5bad3f393b13629da6903bf19", TraitIDs: []string{"cybernetic", "industry_plus_2", "government_dictatorship"}},
		{ID: "mrrshan", Name: "Mrrshan", HelpName: "Mrrshan", HelpRecord: 638, HelpRecordSHA256: "cbdbce1e1d8f3bf3b289e6155c03da22548825622f34dc63803cffcff6b7c587", TraitIDs: []string{"ship_attack_plus_50", "warlord", "rich_home_world", "government_dictatorship"}},
		{ID: "psilon", Name: "Psilon", HelpName: "Psilons", HelpRecord: 639, HelpRecordSHA256: "83ed040e0f108727157484d00bd4f4e2917e07d83bb939274236d83908cb0733", TraitIDs: []string{"science_plus_2", "creative", "low_g_world", "large_home_world", "government_dictatorship"}},
		{ID: "sakkra", Name: "Sakkra", HelpName: "Sakkra", HelpRecord: 640, HelpRecordSHA256: "145caca47f908c7df869316bffd9c60e3f20ae66c52f620c707242cb5979f148", TraitIDs: []string{"population_growth_plus_100", "subterranean", "farming_plus_1", "spying_minus_10", "large_home_world", "government_feudal"}},
		{ID: "silicoid", Name: "Silicoid", HelpName: "Silicoids", HelpRecord: 641, HelpRecordSHA256: "3dc5c69772e8aa4c095b24d8ae4082eef7836d6bccdd96baef4bde2c56ba9b5c", TraitIDs: []string{"lithovore", "tolerant", "population_growth_minus_50", "repulsive", "government_dictatorship"}},
		{ID: "trilarian", Name: "Trilarian", HelpName: "Trilarians", HelpRecord: 642, HelpRecordSHA256: "96402d02bb473434587e5ae13109bf25ce16af53c28e5ae61d8e5a1c052fbac5", TraitIDs: []string{"aquatic", "trans_dimensional", "government_dictatorship"}},
	}
}
