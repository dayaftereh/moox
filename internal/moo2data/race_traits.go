package moo2data

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"moox/internal/i18n"
	"moox/internal/lbx"
	"moox/internal/ruleset"
	"moox/internal/textscan"
)

const (
	raceStuffSourceID      = "moo2-1.31-racestuf-block0"
	raceStuffCostsSourceID = "moo2-1.31-racestuf-block6"
	raceWikiSourceID       = "strategywiki-moo2-race-design-options"
)

type raceOptionSpec struct {
	ID        string
	Label     string
	PickCost  int
	Scope     string
	Value     *float64
	ValueKind string
	Ability   string
	MutexWith []string
	Effects   string
}

type raceGroupSpec struct {
	ID              string
	Label           string
	Scope           string
	SelectionMode   string
	Required        bool
	DefaultOptionID string
	Options         []raceOptionSpec
}

// DecodeRaceTraits turns the English Race Design labels from the original 1.31
// RACESTUF.LBX into stable MOOX rule IDs. Pick costs are cross-referenced against
// the standard-game table documented by StrategyWiki and are deliberately
// marked separately from values directly observed in the original archive.
type RaceTraitsBundle struct {
	Rules   *ruleset.RaceTraitsFile
	English *i18n.File
}

func DecodeRaceTraits(installationRoot string) (*ruleset.RaceTraitsFile, error) {
	bundle, err := DecodeRaceTraitsBundle(installationRoot)
	if err != nil {
		return nil, err
	}
	return bundle.Rules, nil
}

func DecodeRaceTraitsBundle(installationRoot string) (*RaceTraitsBundle, error) {
	archivePath := filepath.Join(installationRoot, "RACESTUF.LBX")
	archive, err := lbx.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("open RACESTUF.LBX: %w", err)
	}
	if len(archive.Entries) <= 6 {
		return nil, fmt.Errorf("RACESTUF.LBX does not contain required blocks 0 and 6")
	}

	block, err := archive.ReadEntry(0)
	if err != nil {
		return nil, fmt.Errorf("read RACESTUF.LBX block 0: %w", err)
	}
	observed := textscan.ASCII(block, 2)
	specs := raceTraitSpecs()
	expectedStrings := 0
	expectedCosts := 0
	for _, group := range specs {
		expectedStrings += 1 + len(group.Options)
		expectedCosts += len(group.Options)
	}
	if len(observed) != expectedStrings {
		return nil, fmt.Errorf("RACESTUF.LBX block 0 yielded %d strings, expected %d", len(observed), expectedStrings)
	}

	costBlock, err := archive.ReadEntry(6)
	if err != nil {
		return nil, fmt.Errorf("read RACESTUF.LBX block 6: %w", err)
	}
	if len(costBlock) != expectedCosts+4 {
		return nil, fmt.Errorf("RACESTUF.LBX block 6 has %d bytes, expected %d pick costs plus 4 trailing bytes", len(costBlock), expectedCosts)
	}
	for i, b := range costBlock[expectedCosts:] {
		if b != 0 {
			return nil, fmt.Errorf("RACESTUF.LBX block 6 trailing byte %d is 0x%02X, expected zero", expectedCosts+i, b)
		}
	}

	fileHash, err := sha256File(archivePath)
	if err != nil {
		return nil, err
	}
	blockSum := sha256.Sum256(block)
	costBlockSum := sha256.Sum256(costBlock)
	blockIndex := 0
	costBlockIndex := 6
	out := &ruleset.RaceTraitsFile{
		SchemaVersion: ruleset.RaceTraitsSchemaVersion,
		Ruleset:       "moo2-1.31",
		PickBudget: ruleset.PickBudget{
			StartingPicks:    10,
			MaxNegativePicks: 10,
		},
		Sources: []ruleset.Source{
			{
				ID:          raceStuffSourceID,
				Type:        "original-observed",
				Description: "English Race Design labels observed in the local official Master of Orion II 1.31 data set",
				Archive:     "RACESTUF.LBX",
				Block:       &blockIndex,
				SHA256:      fileHash,
				BlockSHA256: hex.EncodeToString(blockSum[:]),
			},
			{
				ID:          raceStuffCostsSourceID,
				Type:        "original-observed",
				Description: "Race Design pick costs observed as signed bytes in the local official Master of Orion II 1.31 RACESTUF.LBX data set",
				Archive:     "RACESTUF.LBX",
				Block:       &costBlockIndex,
				SHA256:      fileHash,
				BlockSHA256: hex.EncodeToString(costBlockSum[:]),
			},
			{
				ID:           raceWikiSourceID,
				Type:         "secondary-reference",
				Description:  "Standard-game pick budget, trait scopes and behavioral cross-check",
				URL:          "https://strategywiki.org/wiki/Master_of_Orion_II:_Battle_at_Antares/Race_design_options",
				AccessedDate: "2026-08-26",
			},
		},
	}

	cursor := 0
	costCursor := 0
	for _, groupSpec := range specs {
		groupLabel := observed[cursor]
		if groupLabel.Value != groupSpec.Label {
			return nil, fmt.Errorf("RACESTUF group label at string %d is %q, expected %q", cursor, groupLabel.Value, groupSpec.Label)
		}
		cursor++
		groupOffset := groupLabel.Offset
		group := ruleset.RaceTraitGroup{
			ID:              groupSpec.ID,
			NameKey:         groupNameKey(groupSpec.ID),
			Scope:           groupSpec.Scope,
			SelectionMode:   groupSpec.SelectionMode,
			Required:        groupSpec.Required,
			DefaultOptionID: groupSpec.DefaultOptionID,
			NameSource:      ruleset.FieldProvenance{SourceID: raceStuffSourceID, Offset: &groupOffset},
			Options:         make([]ruleset.RaceTraitOption, 0, len(groupSpec.Options)),
		}
		for _, optionSpec := range groupSpec.Options {
			item := observed[cursor]
			if item.Value != optionSpec.Label {
				return nil, fmt.Errorf("RACESTUF option label at string %d is %q, expected %q", cursor, item.Value, optionSpec.Label)
			}
			cursor++
			offset := item.Offset
			costOffset := costCursor
			pickCost := int(int8(costBlock[costCursor]))
			if pickCost != optionSpec.PickCost {
				return nil, fmt.Errorf("RACESTUF pick cost %d for %s does not match expected research mapping %d", pickCost, optionSpec.ID, optionSpec.PickCost)
			}
			costCursor++
			group.Options = append(group.Options, ruleset.RaceTraitOption{
				ID:             optionSpec.ID,
				NameKey:        optionNameKey(optionSpec.ID),
				PickCost:       pickCost,
				PickCostSource: ruleset.FieldProvenance{SourceID: raceStuffCostsSourceID, Offset: &costOffset},
				Scope:          optionSpec.Scope,
				Value:          optionSpec.Value,
				ValueKind:      optionSpec.ValueKind,
				Ability:        optionSpec.Ability,
				MutexWith:      optionSpec.MutexWith,
				NameSource:     ruleset.FieldProvenance{SourceID: raceStuffSourceID, Offset: &offset},
				Verification:   ruleset.VerificationInfo{Name: "original-observed", PickCost: "original-observed", Effects: optionSpec.Effects},
			})
		}
		out.Groups = append(out.Groups, group)
	}

	if err := out.Validate(); err != nil {
		return nil, err
	}

	language := &i18n.File{
		SchemaVersion: i18n.SchemaVersion,
		Locale:        "en",
		Sources: []i18n.Source{{
			ID:          raceStuffSourceID,
			Type:        "original-observed",
			Description: "English Race Design names observed in Master of Orion II 1.31 RACESTUF.LBX",
			Archive:     "RACESTUF.LBX",
			Block:       &blockIndex,
			SHA256:      fileHash,
			BlockSHA256: hex.EncodeToString(blockSum[:]),
		}},
		Strings: make(map[string]string, expectedStrings),
	}
	cursor = 0
	for _, groupSpec := range specs {
		language.Strings[groupNameKey(groupSpec.ID)] = observed[cursor].Value
		cursor++
		for _, optionSpec := range groupSpec.Options {
			language.Strings[optionNameKey(optionSpec.ID)] = observed[cursor].Value
			cursor++
		}
	}
	if err := language.Validate(); err != nil {
		return nil, err
	}
	return &RaceTraitsBundle{Rules: out, English: language}, nil
}

func groupNameKey(id string) string  { return "race_traits.group." + id + ".name" }
func optionNameKey(id string) string { return "race_traits.option." + id + ".name" }

func raceTraitSpecs() []raceGroupSpec {
	return []raceGroupSpec{
		{
			ID: "population_growth", Label: "Population", Scope: "genetic", SelectionMode: "single",
			Options: []raceOptionSpec{
				numeric("population_growth_minus_50", "-50% Growth", -4, 0.5, "population_growth_multiplier"),
				numeric("population_growth_plus_50", "+50% Growth", 3, 1.5, "population_growth_multiplier"),
				numeric("population_growth_plus_100", "+100% Growth", 6, 2.0, "population_growth_multiplier"),
			},
		},
		{
			ID: "farming", Label: "Farming", Scope: "genetic", SelectionMode: "single",
			Options: []raceOptionSpec{
				withMutex(numeric("farming_minus_half", "-1/2 Food", -3, -0.5, "food_per_farmer_delta"), "lithovore"),
				withMutex(numeric("farming_plus_1", "+1 Food", 4, 1, "food_per_farmer_delta"), "lithovore"),
				withMutex(numeric("farming_plus_2", "+2 Food", 7, 2, "food_per_farmer_delta"), "lithovore"),
			},
		},
		{
			ID: "industry", Label: "Industry", Scope: "genetic", SelectionMode: "single",
			Options: []raceOptionSpec{
				numeric("industry_minus_1", "-1 Production", -3, -1, "production_per_worker_delta"),
				numeric("industry_plus_1", "+1 Production", 3, 1, "production_per_worker_delta"),
				numeric("industry_plus_2", "+2 Production", 6, 2, "production_per_worker_delta"),
			},
		},
		{
			ID: "science", Label: "Science", Scope: "genetic", SelectionMode: "single",
			Options: []raceOptionSpec{
				numeric("science_minus_1", "-1 Research", -3, -1, "research_per_scientist_delta"),
				numeric("science_plus_1", "+1 Research", 3, 1, "research_per_scientist_delta"),
				numeric("science_plus_2", "+2 Research", 6, 2, "research_per_scientist_delta"),
			},
		},
		{
			ID: "money", Label: "Money", Scope: "empire", SelectionMode: "single",
			Options: []raceOptionSpec{
				numeric("money_minus_0_5", "-0.5 BC", -4, -0.5, "tax_bc_per_population_delta"),
				numeric("money_plus_0_5", "+0.5 BC", 5, 0.5, "tax_bc_per_population_delta"),
				numeric("money_plus_1_0", "+1.0 BC", 8, 1, "tax_bc_per_population_delta"),
			},
		},
		{
			ID: "ship_defense", Label: "Ship Defense", Scope: "empire", SelectionMode: "single",
			Options: []raceOptionSpec{
				numeric("ship_defense_minus_20", "-20", -2, -20, "ship_defense_bonus"),
				numeric("ship_defense_plus_25", "+25", 3, 25, "ship_defense_bonus"),
				numeric("ship_defense_plus_50", "+50", 7, 50, "ship_defense_bonus"),
			},
		},
		{
			ID: "ship_attack", Label: "Ship Attack", Scope: "empire", SelectionMode: "single",
			Options: []raceOptionSpec{
				numeric("ship_attack_minus_20", "-20", -2, -20, "ship_attack_bonus"),
				numeric("ship_attack_plus_20", "+20", 2, 20, "ship_attack_bonus"),
				numeric("ship_attack_plus_50", "+50", 4, 50, "ship_attack_bonus"),
			},
		},
		{
			ID: "ground_combat", Label: "Ground Combat", Scope: "empire", SelectionMode: "single",
			Options: []raceOptionSpec{
				numeric("ground_combat_minus_10", "-10", -2, -10, "ground_combat_bonus"),
				numeric("ground_combat_plus_10", "+10", 2, 10, "ground_combat_bonus"),
				numeric("ground_combat_plus_20", "+20", 4, 20, "ground_combat_bonus"),
			},
		},
		{
			ID: "spying", Label: "Spying", Scope: "empire", SelectionMode: "single",
			Options: []raceOptionSpec{
				numeric("spying_minus_10", "-10", -3, -10, "spying_bonus"),
				numeric("spying_plus_10", "+10", 3, 10, "spying_bonus"),
				numeric("spying_plus_20", "+20", 6, 20, "spying_bonus"),
			},
		},
		{
			ID: "government", Label: "Governments", Scope: "empire", SelectionMode: "single", Required: true, DefaultOptionID: "government_dictatorship",
			Options: []raceOptionSpec{
				ability("government_feudal", "Feudal", -4, "empire", "government_feudal"),
				ability("government_dictatorship", "Dictatorship", 0, "empire", "government_dictatorship"),
				ability("government_democracy", "Democracy", 7, "empire", "government_democracy"),
				ability("government_unification", "Unification", 6, "empire", "government_unification"),
			},
		},
		{
			ID: "special_abilities", Label: "Special Abilities", Scope: "mixed", SelectionMode: "multi",
			Options: []raceOptionSpec{
				withMutex(ability("low_g_world", "Low-G World", -5, "genetic_homeworld", "low_g_world"), "high_g_world"),
				withMutex(ability("high_g_world", "High-G World", 6, "genetic_homeworld", "high_g_world"), "low_g_world"),
				ability("aquatic", "Aquatic", 5, "genetic_homeworld", "aquatic"),
				ability("subterranean", "Subterranean", 6, "genetic", "subterranean"),
				ability("large_home_world", "Large Home World", 1, "homeworld", "large_home_world"),
				withMutex(ability("rich_home_world", "Rich Home World", 2, "homeworld", "rich_home_world"), "poor_home_world"),
				withMutex(ability("poor_home_world", "Poor Home World", -1, "homeworld", "poor_home_world"), "rich_home_world"),
				ability("artifacts_world", "Artifacts World", 3, "homeworld", "artifacts_world"),
				withMutex(ability("cybernetic", "Cybernetic", 4, "genetic", "cybernetic"), "lithovore"),
				withMutex(ability("lithovore", "Lithovore", 10, "genetic", "lithovore"), "cybernetic", "farming_minus_half", "farming_plus_1", "farming_plus_2"),
				withMutex(ability("repulsive", "Repulsive", -6, "empire", "repulsive"), "charismatic"),
				withMutex(ability("charismatic", "Charismatic", 3, "empire", "charismatic"), "repulsive"),
				withMutex(ability("uncreative", "Uncreative", -4, "empire", "uncreative"), "creative"),
				withMutex(ability("creative", "Creative", 8, "empire", "creative"), "uncreative"),
				ability("tolerant", "Tolerant", 10, "genetic", "tolerant"),
				ability("fantastic_traders", "Fantastic Traders", 4, "empire", "fantastic_traders"),
				ability("telepathic", "Telepathic", 6, "empire", "telepathic"),
				ability("lucky", "Lucky", 3, "empire", "lucky"),
				ability("omniscient", "Omniscient", 3, "empire", "omniscient"),
				ability("stealthy_ships", "Stealthy Ships", 4, "empire", "stealthy_ships"),
				ability("trans_dimensional", "Trans Dimensional", 5, "empire", "trans_dimensional"),
				ability("warlord", "Warlord", 4, "empire", "warlord"),
			},
		},
	}
}

func numeric(id, label string, cost int, value float64, kind string) raceOptionSpec {
	return raceOptionSpec{
		ID: id, Label: label, PickCost: cost, Value: floatPtr(value), ValueKind: kind,
		Effects: "original-label-derived; exact formula integration pending subsystem parity tests",
	}
}

func ability(id, label string, cost int, scope, abilityID string) raceOptionSpec {
	return raceOptionSpec{
		ID: id, Label: label, PickCost: cost, Scope: scope, Ability: abilityID,
		Effects: "ability-identified; exact behavioral implementation pending parity tests",
	}
}

func withMutex(spec raceOptionSpec, ids ...string) raceOptionSpec {
	spec.MutexWith = append(spec.MutexWith, ids...)
	return spec
}

func floatPtr(v float64) *float64 { return &v }

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
