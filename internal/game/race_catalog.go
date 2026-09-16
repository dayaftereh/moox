package game

import (
	"fmt"
	"sort"

	"moox/internal/ruleset"
)

type PresetRaceAvailability string

const (
	PresetRaceAvailabilitySupported PresetRaceAvailability = "supported"
	PresetRaceAvailabilityPlanned   PresetRaceAvailability = "planned"

	DefaultPlayerRaceID = "human"
	FixedOpponentRaceID = "darlok"
)

type PresetRaceProfile struct {
	ID                 string                 `json:"id"`
	Order              int                    `json:"order"`
	NameKey            string                 `json:"name_key"`
	PlayerAvailability PresetRaceAvailability `json:"player_availability"`
	TraitIDs           []string               `json:"trait_ids"`
	CardFactTraitIDs   []string               `json:"card_fact_trait_ids"`
}

type PresetRaceCatalog struct {
	DefaultPlayerRaceID string              `json:"default_player_race_id"`
	FixedOpponentRaceID string              `json:"fixed_opponent_race_id"`
	Profiles            []PresetRaceProfile `json:"profiles"`
}

var frozenPresetRaceOrder = []string{
	"alkari", "bulrathi", "darlok", "elerian", "gnolam", "human", "klackon", "meklar", "mrrshan", "psilon", "sakkra", "silicoid", "trilarian",
}

var frozenRaceCardFacts = map[string][]string{
	"alkari":    {"ship_defense_plus_50", "artifacts_world", "government_dictatorship"},
	"bulrathi":  {"high_g_world", "ship_attack_plus_20", "ground_combat_plus_10"},
	"darlok":    {"spying_plus_20", "stealthy_ships", "government_dictatorship"},
	"elerian":   {"omniscient", "telepathic", "government_feudal"},
	"gnolam":    {"fantastic_traders", "money_plus_1_0", "lucky"},
	"human":     {"government_democracy"},
	"klackon":   {"government_unification", "farming_plus_1", "industry_plus_1", "uncreative"},
	"meklar":    {"cybernetic", "industry_plus_2", "government_dictatorship"},
	"mrrshan":   {"ship_attack_plus_50", "warlord", "rich_home_world"},
	"psilon":    {"science_plus_2", "creative", "low_g_world"},
	"sakkra":    {"population_growth_plus_100", "subterranean", "farming_plus_1"},
	"silicoid":  {"lithovore", "tolerant", "population_growth_minus_50"},
	"trilarian": {"aquatic", "trans_dimensional", "government_dictatorship"},
}

func IsSupportedPlayerRaceID(id string) bool {
	return id == "human" || id == "klackon"
}

func PresetRaceCatalogFromRules(rules *EconomyRules) (PresetRaceCatalog, error) {
	if rules == nil || len(rules.PresetRaces) == 0 {
		return PresetRaceCatalog{}, fmt.Errorf("preset race rules are unavailable")
	}

	races := append([]ruleset.Race(nil), rules.PresetRaces...)
	sort.SliceStable(races, func(i, j int) bool { return races[i].Order < races[j].Order })
	if len(races) != len(frozenPresetRaceOrder) {
		return PresetRaceCatalog{}, fmt.Errorf("preset race catalog has %d races, want %d", len(races), len(frozenPresetRaceOrder))
	}

	catalog := PresetRaceCatalog{
		DefaultPlayerRaceID: DefaultPlayerRaceID,
		FixedOpponentRaceID: FixedOpponentRaceID,
		Profiles:            make([]PresetRaceProfile, 0, len(races)),
	}
	for i, race := range races {
		if race.ID != frozenPresetRaceOrder[i] || race.Order != i {
			return PresetRaceCatalog{}, fmt.Errorf("preset race[%d]=%q order=%d, want %q order=%d", i, race.ID, race.Order, frozenPresetRaceOrder[i], i)
		}
		traitIDs := make([]string, 0, len(race.TraitSelections))
		selected := make(map[string]struct{}, len(race.TraitSelections))
		for _, selection := range race.TraitSelections {
			traitIDs = append(traitIDs, selection.TraitID)
			selected[selection.TraitID] = struct{}{}
		}
		facts, ok := frozenRaceCardFacts[race.ID]
		if !ok {
			return PresetRaceCatalog{}, fmt.Errorf("preset race %q has no frozen card facts", race.ID)
		}
		for _, factID := range facts {
			if _, ok := selected[factID]; !ok {
				return PresetRaceCatalog{}, fmt.Errorf("preset race %q card fact %q is not a selected trait", race.ID, factID)
			}
		}
		availability := PresetRaceAvailabilityPlanned
		if IsSupportedPlayerRaceID(race.ID) {
			availability = PresetRaceAvailabilitySupported
		}
		catalog.Profiles = append(catalog.Profiles, PresetRaceProfile{
			ID:                 race.ID,
			Order:              race.Order,
			NameKey:            race.NameKey,
			PlayerAvailability: availability,
			TraitIDs:           traitIDs,
			CardFactTraitIDs:   append([]string(nil), facts...),
		})
	}
	return catalog, nil
}
