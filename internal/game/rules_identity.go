package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type RulesIdentity struct {
	ID     string `json:"id"`
	SHA256 string `json:"sha256"`
}

var authoritativeRulesFiles = []string{
	"buildings.json",
	"economy.json",
	"new_game_galaxy.json",
	"planet_classes.json",
	"race_traits.json",
	"races.json",
	"ship_hulls.json",
	"tactical_combat.json",
	"technologies.json",
}

func loadRulesIdentity(rulesetDir string) (RulesIdentity, error) {
	clean := filepath.Clean(rulesetDir)
	id := strings.TrimSpace(filepath.Base(clean))
	if id == "" || id == "." || id == string(filepath.Separator) {
		return RulesIdentity{}, fmt.Errorf("ruleset directory has no stable base identity")
	}
	names := append([]string(nil), authoritativeRulesFiles...)
	sort.Strings(names)
	var manifest bytes.Buffer
	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join(clean, name))
		if err != nil {
			return RulesIdentity{}, fmt.Errorf("fingerprint rules file %s: %w", name, err)
		}
		var compact bytes.Buffer
		if err := json.Compact(&compact, raw); err != nil {
			return RulesIdentity{}, fmt.Errorf("compact rules file %s: %w", name, err)
		}
		sum := sha256.Sum256(compact.Bytes())
		fmt.Fprintf(&manifest, "%s:%x\n", name, sum)
	}
	total := sha256.Sum256(manifest.Bytes())
	return RulesIdentity{ID: id, SHA256: fmt.Sprintf("%x", total)}, nil
}

func (r *EconomyRules) Identity() RulesIdentity {
	if r == nil {
		return RulesIdentity{}
	}
	return RulesIdentity{ID: r.RulesetID, SHA256: r.RulesetSHA256}
}
