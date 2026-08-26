package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const RacesSchemaVersion = 1

type RacesFile struct {
	SchemaVersion int      `json:"schema_version"`
	Ruleset       string   `json:"ruleset"`
	Sources       []Source `json:"sources"`
	Races         []Race   `json:"races"`
}

type Race struct {
	ID               string               `json:"id"`
	Order            int                  `json:"order"`
	NameKey          string               `json:"name_key"`
	NameSource       FieldProvenance      `json:"name_source"`
	TraitSelections  []RaceTraitSelection `json:"trait_selections"`
	DerivedPickTotal int                  `json:"derived_pick_total"`
	PortraitAssetKey string               `json:"portrait_asset_key"`
	IconAssetKey     string               `json:"icon_asset_key"`
	Source           RaceSource           `json:"source"`
}

type RaceTraitSelection struct {
	TraitID      string `json:"trait_id"`
	Verification string `json:"verification"`
}

type RaceSource struct {
	SourceID     string `json:"source_id"`
	RecordIndex  int    `json:"record_index"`
	RecordOffset int    `json:"record_offset"`
	RecordSize   int    `json:"record_size"`
	RecordSHA256 string `json:"record_sha256"`
}

func LoadRaces(path string) (*RacesFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file RacesFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *RacesFile) ValidateAgainstTraits(traits *RaceTraitsFile) error {
	if f.SchemaVersion != RacesSchemaVersion {
		return fmt.Errorf("unsupported races schema version %d", f.SchemaVersion)
	}
	if f.Ruleset == "" {
		return fmt.Errorf("ruleset is required")
	}
	if traits == nil {
		return fmt.Errorf("race traits are required for validation")
	}

	type traitInfo struct {
		Cost          int
		GroupID       string
		SelectionMode string
	}
	traitByID := make(map[string]traitInfo)
	for _, group := range traits.Groups {
		for _, option := range group.Options {
			traitByID[option.ID] = traitInfo{Cost: option.PickCost, GroupID: group.ID, SelectionMode: group.SelectionMode}
		}
	}

	ids := make(map[string]struct{}, len(f.Races))
	orders := make(map[int]struct{}, len(f.Races))
	nameKeys := make(map[string]struct{}, len(f.Races))
	for _, race := range f.Races {
		if race.ID == "" || race.NameKey == "" {
			return fmt.Errorf("race id and name_key are required")
		}
		if _, exists := ids[race.ID]; exists {
			return fmt.Errorf("duplicate race id %q", race.ID)
		}
		ids[race.ID] = struct{}{}
		if _, exists := orders[race.Order]; exists {
			return fmt.Errorf("duplicate race order %d", race.Order)
		}
		orders[race.Order] = struct{}{}
		if _, exists := nameKeys[race.NameKey]; exists {
			return fmt.Errorf("duplicate race name key %q", race.NameKey)
		}
		nameKeys[race.NameKey] = struct{}{}
		if race.PortraitAssetKey == "" || race.IconAssetKey == "" {
			return fmt.Errorf("race %q requires semantic portrait/icon asset keys", race.ID)
		}
		if len(race.TraitSelections) == 0 {
			return fmt.Errorf("race %q has no trait selections", race.ID)
		}

		seenTraits := make(map[string]struct{})
		singleGroups := make(map[string]string)
		pickTotal := 0
		for _, selected := range race.TraitSelections {
			if selected.Verification == "" {
				return fmt.Errorf("race %q trait %q has no verification level", race.ID, selected.TraitID)
			}
			if _, exists := seenTraits[selected.TraitID]; exists {
				return fmt.Errorf("race %q repeats trait %q", race.ID, selected.TraitID)
			}
			seenTraits[selected.TraitID] = struct{}{}
			info, ok := traitByID[selected.TraitID]
			if !ok {
				return fmt.Errorf("race %q references unknown trait %q", race.ID, selected.TraitID)
			}
			if info.SelectionMode == "single" {
				if previous, exists := singleGroups[info.GroupID]; exists {
					return fmt.Errorf("race %q selects both %q and %q from single-select group %q", race.ID, previous, selected.TraitID, info.GroupID)
				}
				singleGroups[info.GroupID] = selected.TraitID
			}
			pickTotal += info.Cost
		}
		if pickTotal != race.DerivedPickTotal {
			return fmt.Errorf("race %q derived pick total=%d, calculated=%d", race.ID, race.DerivedPickTotal, pickTotal)
		}
	}
	return nil
}
