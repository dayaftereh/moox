package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const ShipSpecialsSchemaVersion = 1

type ShipSpecialsFile struct {
	SchemaVersion int                 `json:"schema_version"`
	Ruleset       string              `json:"ruleset"`
	Megafluxers   ShipMegafluxersRule `json:"megafluxers"`
	Specials      []ShipSpecialRule   `json:"specials"`
}

type ShipMegafluxersRule struct {
	TechnologyID     int `json:"technology_id"`
	HullSpacePercent int `json:"hull_space_percent"`
}

type ShipSpecialRule struct {
	ID                   string `json:"id"`
	OriginalID           int    `json:"original_id"`
	TechnologyID         int    `json:"technology_id"`
	BaseSpaceByHull      []int  `json:"base_space_by_hull"`
	BaseCostByHullPP     []int  `json:"base_cost_by_hull_pp"`
	ExcludedOriginalIDs  []int  `json:"excluded_original_ids,omitempty"`
	RequiresShield       bool   `json:"requires_shield,omitempty"`
	MiniaturizationClass int    `json:"miniaturization_class"`
	Miniaturizes         bool   `json:"miniaturizes"`
	SupportClass         int    `json:"support_class"`
	TacticalSupport      string `json:"tactical_support"`
}

func LoadShipSpecials(path string) (*ShipSpecialsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file ShipSpecialsFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *ShipSpecialsFile) Validate() error {
	if f == nil {
		return fmt.Errorf("ship specials rules are nil")
	}
	if f.SchemaVersion != ShipSpecialsSchemaVersion {
		return fmt.Errorf("unsupported ship specials schema version %d", f.SchemaVersion)
	}
	if f.Ruleset == "" {
		return fmt.Errorf("ship specials ruleset is required")
	}
	if f.Megafluxers.TechnologyID <= 0 || f.Megafluxers.HullSpacePercent <= 100 {
		return fmt.Errorf("invalid Megafluxers rule")
	}
	if len(f.Specials) != 38 {
		return fmt.Errorf("expected exactly 38 selectable ship specials, got %d", len(f.Specials))
	}
	seenIDs := make(map[string]struct{}, len(f.Specials))
	seenOriginal := make(map[int]struct{}, len(f.Specials))
	for i, special := range f.Specials {
		if special.ID == "" || special.OriginalID < 1 || special.OriginalID > 38 || special.TechnologyID <= 0 {
			return fmt.Errorf("invalid ship special[%d] identity", i)
		}
		if _, ok := seenIDs[special.ID]; ok {
			return fmt.Errorf("duplicate ship special id %q", special.ID)
		}
		if _, ok := seenOriginal[special.OriginalID]; ok {
			return fmt.Errorf("duplicate ship special original id %d", special.OriginalID)
		}
		seenIDs[special.ID] = struct{}{}
		seenOriginal[special.OriginalID] = struct{}{}
		if len(special.BaseSpaceByHull) != 6 || len(special.BaseCostByHullPP) != 6 {
			return fmt.Errorf("ship special %q requires six hull Space/Cost values", special.ID)
		}
		for hull := 0; hull < 6; hull++ {
			if special.BaseCostByHullPP[hull] <= 0 {
				return fmt.Errorf("ship special %q has invalid hull %d PP", special.ID, hull)
			}
			if special.ID == "battle_pods" {
				if special.BaseSpaceByHull[hull] >= 0 {
					return fmt.Errorf("Battle Pods hull %d Space must be negative", hull)
				}
			} else if special.BaseSpaceByHull[hull] <= 0 {
				return fmt.Errorf("ship special %q has invalid hull %d Space", special.ID, hull)
			}
		}
		if special.MiniaturizationClass != 1 {
			return fmt.Errorf("ship special %q miniaturization class=%d want=1", special.ID, special.MiniaturizationClass)
		}
		if special.ID == "battle_pods" {
			if special.Miniaturizes {
				return fmt.Errorf("Battle Pods must not miniaturize")
			}
		} else if !special.Miniaturizes {
			return fmt.Errorf("ship special %q must miniaturize", special.ID)
		}
		if special.SupportClass < 1 || special.SupportClass > 3 || special.TacticalSupport == "" {
			return fmt.Errorf("ship special %q has invalid support classification", special.ID)
		}
		selfExcluded := false
		for _, excluded := range special.ExcludedOriginalIDs {
			if excluded < 1 || excluded > 38 {
				return fmt.Errorf("ship special %q exclusion %d is outside 1..38", special.ID, excluded)
			}
			if excluded == special.OriginalID {
				selfExcluded = true
			}
		}
		if !selfExcluded {
			return fmt.Errorf("ship special %q must self-exclude", special.ID)
		}
		expectedSupport, ok := expectedShipSpecialTacticalSupport(special.ID)
		if !ok {
			return fmt.Errorf("ship special %q has no frozen tactical support classification", special.ID)
		}
		if special.TacticalSupport != expectedSupport {
			return fmt.Errorf("ship special %q tactical support=%q want=%q", special.ID, special.TacticalSupport, expectedSupport)
		}
	}
	return nil
}

func expectedShipSpecialTacticalSupport(id string) (string, bool) {
	switch id {
	case "battle_pods", "extended_fuel_tanks":
		return "strategic_only", true
	case "inertial_nullifier", "inertial_stabilizer":
		return "turning_mode", true
	case "phase_shifter", "sub_space_teleporter":
		return "tactical_active", true
	case "achilles_targeting_unit", "augmented_engines", "automated_repair_unit", "battle_scanner",
		"cloaking_device", "damper_field", "displacement_device", "ecm_jammer", "energy_absorber",
		"fast_missile_racks", "hard_shields", "heavy_armor", "high_energy_focus", "hyper_x_capacitors",
		"lightning_field", "multi_phased_shields", "multi_wave_ecm_jammer", "phasing_cloak",
		"quantum_detonator", "rangemaster_unit", "reflection_field", "reinforced_hull",
		"security_stations", "shield_capacitors", "structural_analyzer", "time_warp_facilitator",
		"transporters", "troop_pods", "warp_dissipator", "wide_area_jammer":
		return "tactical_passive", true
	case "scout_lab", "stealth_field":
		return "metadata_only", true
	default:
		return "", false
	}
}

func (f *ShipSpecialsFile) ByID(id string) (ShipSpecialRule, bool) {
	if f == nil {
		return ShipSpecialRule{}, false
	}
	for _, special := range f.Specials {
		if special.ID == id {
			return special, true
		}
	}
	return ShipSpecialRule{}, false
}

func (f *ShipSpecialsFile) ByOriginalID(id int) (ShipSpecialRule, bool) {
	if f == nil {
		return ShipSpecialRule{}, false
	}
	for _, special := range f.Specials {
		if special.OriginalID == id {
			return special, true
		}
	}
	return ShipSpecialRule{}, false
}
