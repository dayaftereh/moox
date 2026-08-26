package moo2data

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"moox/internal/i18n"
	"moox/internal/lbx"
	"moox/internal/ruleset"
	"moox/internal/textscan"
)

const (
	technologyNamesSourceID = "moo2-1.31-techname-block0-technologies"
	technologyCount         = 203
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

	allRuns := textscan.ASCII(block, 3)
	techRuns, err := concreteTechnologyRuns(allRuns)
	if err != nil {
		return nil, err
	}

	out := &ruleset.TechnologiesFile{
		SchemaVersion: ruleset.TechnologiesSchemaVersion,
		Ruleset:       "moo2-1.31",
		Sources: []ruleset.Source{{
			ID:          technologyNamesSourceID,
			Type:        "original-observed",
			Description: "The 203-entry concrete technology-name section in Master of Orion II 1.31 TECHNAME.LBX block 0, bounded by No Tech and the following Biology section",
			Archive:     "TECHNAME.LBX",
			Block:       &blockIndex,
			SHA256:      fileHash,
			BlockSHA256: hex.EncodeToString(blockSum[:]),
		}},
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
		out.Technologies = append(out.Technologies, ruleset.Technology{
			ID:           id,
			Order:        order,
			TechnologyID: order + 1,
			NameKey:      nameKey,
			NameSource:   ruleset.FieldProvenance{SourceID: technologyNamesSourceID, Offset: &offset},
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
