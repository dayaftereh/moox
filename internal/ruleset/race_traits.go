package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const RaceTraitsSchemaVersion = 1

type RaceTraitsFile struct {
	SchemaVersion int              `json:"schema_version"`
	Ruleset       string           `json:"ruleset"`
	PickBudget    PickBudget       `json:"pick_budget"`
	Sources       []Source         `json:"sources"`
	Groups        []RaceTraitGroup `json:"groups"`
}

type PickBudget struct {
	StartingPicks    int `json:"starting_picks"`
	MaxNegativePicks int `json:"max_negative_picks"`
}

type Source struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Description  string `json:"description,omitempty"`
	Path         string `json:"path,omitempty"`
	Archive      string `json:"archive,omitempty"`
	Block        *int   `json:"block,omitempty"`
	SHA256       string `json:"sha256,omitempty"`
	BlockSHA256  string `json:"block_sha256,omitempty"`
	URL          string `json:"url,omitempty"`
	AccessedDate string `json:"accessed_date,omitempty"`
}

type RaceTraitGroup struct {
	ID              string            `json:"id"`
	Label           string            `json:"label"`
	Scope           string            `json:"scope"`
	SelectionMode   string            `json:"selection_mode"`
	Required        bool              `json:"required"`
	DefaultOptionID string            `json:"default_option_id,omitempty"`
	Source          FieldProvenance   `json:"source"`
	Options         []RaceTraitOption `json:"options"`
}

type RaceTraitOption struct {
	ID           string           `json:"id"`
	Label        string           `json:"label"`
	PickCost     int              `json:"pick_cost"`
	Scope        string           `json:"scope,omitempty"`
	Value        *float64         `json:"value,omitempty"`
	ValueKind    string           `json:"value_kind,omitempty"`
	Ability      string           `json:"ability,omitempty"`
	MutexWith    []string         `json:"mutex_with,omitempty"`
	Source       FieldProvenance  `json:"source"`
	Verification VerificationInfo `json:"verification"`
}

type FieldProvenance struct {
	SourceID string `json:"source_id"`
	Offset   *int   `json:"offset,omitempty"`
}

type VerificationInfo struct {
	Label    string `json:"label"`
	PickCost string `json:"pick_cost"`
	Effects  string `json:"effects"`
}

func LoadRaceTraits(path string) (*RaceTraitsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file RaceTraitsFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	if err := file.Validate(); err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *RaceTraitsFile) Validate() error {
	if f.SchemaVersion != RaceTraitsSchemaVersion {
		return fmt.Errorf("unsupported race traits schema version %d", f.SchemaVersion)
	}
	if f.Ruleset == "" {
		return fmt.Errorf("ruleset is required")
	}
	if f.PickBudget.StartingPicks < 0 || f.PickBudget.MaxNegativePicks < 0 {
		return fmt.Errorf("pick budget values must be non-negative magnitudes")
	}

	groupIDs := make(map[string]struct{}, len(f.Groups))
	optionIDs := make(map[string]struct{})
	for _, group := range f.Groups {
		if group.ID == "" || group.Label == "" {
			return fmt.Errorf("trait group id and label are required")
		}
		if _, exists := groupIDs[group.ID]; exists {
			return fmt.Errorf("duplicate trait group id %q", group.ID)
		}
		groupIDs[group.ID] = struct{}{}
		if group.SelectionMode != "single" && group.SelectionMode != "multi" {
			return fmt.Errorf("trait group %q has invalid selection mode %q", group.ID, group.SelectionMode)
		}
		if group.SelectionMode == "multi" && group.DefaultOptionID != "" {
			return fmt.Errorf("multi-select group %q cannot have a default option", group.ID)
		}
		if len(group.Options) == 0 {
			return fmt.Errorf("trait group %q has no options", group.ID)
		}
		foundDefault := group.DefaultOptionID == ""
		for _, option := range group.Options {
			if option.ID == "" || option.Label == "" {
				return fmt.Errorf("trait option id and label are required in group %q", group.ID)
			}
			if _, exists := optionIDs[option.ID]; exists {
				return fmt.Errorf("duplicate trait option id %q", option.ID)
			}
			optionIDs[option.ID] = struct{}{}
			if option.ID == group.DefaultOptionID {
				foundDefault = true
			}
		}
		if !foundDefault {
			return fmt.Errorf("group %q default option %q does not exist", group.ID, group.DefaultOptionID)
		}
	}

	for _, group := range f.Groups {
		for _, option := range group.Options {
			for _, other := range option.MutexWith {
				if _, ok := optionIDs[other]; !ok {
					return fmt.Errorf("option %q references unknown mutex option %q", option.ID, other)
				}
			}
		}
	}
	return nil
}
