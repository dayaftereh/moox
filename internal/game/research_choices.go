package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
)

type ResearchCategory struct {
	ID              string `json:"id"`
	Order           int    `json:"order"`
	NameKey         string `json:"name_key"`
	RootTechFieldID int    `json:"root_tech_field_id"`
}
type ResearchTechnologyEffect struct {
	Kind             string  `json:"kind"`
	ID               string  `json:"id,omitempty"`
	ProductionCostPP float64 `json:"production_cost_pp,omitempty"`
	MaintenanceBC    int     `json:"maintenance_bc,omitempty"`
	FTLSpeed         int     `json:"ftl_speed,omitempty"`
	RangeParsecs     int     `json:"range_parsecs,omitempty"`
	PopulationBonus  float64 `json:"population_bonus,omitempty"`
	GrowthBonus      float64 `json:"growth_bonus,omitempty"`
}

type ResearchTechnologyInfo struct {
	TechnologyID      int                        `json:"technology_id"`
	TechnologyKey     string                     `json:"technology_key"`
	TechnologyNameKey string                     `json:"technology_name_key"`
	Description       string                     `json:"description,omitempty"`
	Effects           []ResearchTechnologyEffect `json:"effects,omitempty"`
}

type ResearchChoice struct {
	CategoryID          string                     `json:"category_id"`
	CategoryOrder       int                        `json:"category_order"`
	CategoryNameKey     string                     `json:"category_name_key"`
	TechFieldID         int                        `json:"tech_field_id"`
	PreviousTechFieldID int                        `json:"previous_tech_field_id"`
	NextTechFieldID     int                        `json:"next_tech_field_id"`
	BaseCostRP          float64                    `json:"base_cost_rp"`
	SelectionMode       core.ResearchSelectionMode `json:"selection_mode"`
	TechnologyIDs       []int                      `json:"technology_ids"`
	TechnologyKeys      []string                   `json:"technology_keys"`
	TechnologyNameKeys  []string                   `json:"technology_name_keys"`
	TechnologyInfo      []ResearchTechnologyInfo   `json:"technology_info,omitempty"`
	CompletedLevels     int                        `json:"completed_levels,omitempty"`
	ResearchLevel       int                        `json:"research_level,omitempty"`
}

// AvailableResearchChoices returns the server-authoritative research frontier.
// A field is selectable when it is not already completed and its predecessor
// is known (or it is a root field). The server also projects the race-specific
// application policy so Human UI, built-in AI and remote agents receive the
// same legal actions without inventing Technology ownership.
func (r *EconomyRules) AvailableResearchChoices(state *core.GameState, empireID core.ID) ([]ResearchChoice, error) {
	if r == nil {
		return nil, fmt.Errorf("economy rules must not be nil")
	}
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return nil, fmt.Errorf("unknown empire %d", empireID)
	}
	modifiers, ok := r.RaceModifiers[empire.RaceID]
	if !ok {
		return nil, fmt.Errorf("empire %d has unknown race %q", empireID, empire.RaceID)
	}

	knownFields := make(map[int]struct{}, len(empire.KnownTechnologyFieldIDs))
	for _, fieldID := range empire.KnownTechnologyFieldIDs {
		knownFields[fieldID] = struct{}{}
	}
	knownTech := make(map[int]struct{}, len(empire.KnownTechnologyIDs))
	for _, technologyID := range empire.KnownTechnologyIDs {
		knownTech[technologyID] = struct{}{}
	}

	fieldIDs := make([]int, 0, len(r.TechnologyFieldCostsRP))
	for fieldID := range r.TechnologyFieldCostsRP {
		fieldIDs = append(fieldIDs, fieldID)
	}
	sort.Ints(fieldIDs)

	choices := make([]ResearchChoice, 0)
	for _, fieldID := range fieldIDs {
		hyperAdvanced := r.isHyperAdvancedField(fieldID)
		if _, known := knownFields[fieldID]; known && !hyperAdvanced {
			continue
		}
		previousID := r.TechnologyFieldPreviousID[fieldID]
		if previousID != 0 {
			if _, known := knownFields[previousID]; !known {
				continue
			}
		}

		categoryID, ok := r.TechnologyFieldCategoryID[fieldID]
		if !ok || categoryID == "" {
			return nil, fmt.Errorf("researchable field %d has no normalized category", fieldID)
		}
		categoryNameKey := r.TechnologyFieldCategoryNameKey[fieldID]
		categoryOrder := r.TechnologyFieldCategoryOrder[fieldID]
		if hyperAdvanced {
			costRP, err := r.researchFieldCostRP(empire, fieldID)
			if err != nil {
				return nil, err
			}
			completed, _ := hyperAdvancedCompletedLevels(empire, fieldID)
			choices = append(choices, ResearchChoice{
				CategoryID:          categoryID,
				CategoryOrder:       categoryOrder,
				CategoryNameKey:     categoryNameKey,
				TechFieldID:         fieldID,
				PreviousTechFieldID: previousID,
				NextTechFieldID:     r.TechnologyFieldNextID[fieldID],
				BaseCostRP:          costRP,
				SelectionMode:       core.ResearchSelectionRepeatField,
				CompletedLevels:     completed,
				ResearchLevel:       completed + 1,
			})
			continue
		}

		allIDs := r.TechnologyIDsByField[fieldID]
		technologyIDs := make([]int, 0, len(allIDs))
		for _, technologyID := range allIDs {
			if _, known := knownTech[technologyID]; known {
				continue
			}
			technologyIDs = append(technologyIDs, technologyID)
		}
		// Hyper-advanced placeholders currently have no concrete Technology IDs
		// in the normalized table. Do not offer an unmaterializable choice.
		if len(technologyIDs) == 0 {
			continue
		}

		mode := r.researchSelectionMode(modifiers, fieldID)
		if mode == core.ResearchSelectionFixedOne {
			fixedID, found := fixedResearchTechnology(empire.UncreativeResearchChoices, fieldID)
			if !found {
				// A later external-acquisition repair can legitimately leave an
				// Uncreative field with no eligible application. The original leaves
				// that field open but unselectable; mirror that legal-action surface.
				continue
			}
			if _, known := knownTech[fixedID]; known {
				return nil, fmt.Errorf("uncreative empire %d already knows fixed technology %d for incomplete field %d; external-acquisition semantics are not modeled yet", empireID, fixedID, fieldID)
			}
			if !containsInt(technologyIDs, fixedID) {
				return nil, fmt.Errorf("uncreative empire %d fixed technology %d is not a legal application of field %d", empireID, fixedID, fieldID)
			}
			technologyIDs = []int{fixedID}
		}

		technologyKeys := make([]string, len(technologyIDs))
		technologyNameKeys := make([]string, len(technologyIDs))
		technologyInfo := make([]ResearchTechnologyInfo, len(technologyIDs))
		for i, technologyID := range technologyIDs {
			key, ok := r.TechnologyKeyByID[technologyID]
			if !ok || key == "" {
				return nil, fmt.Errorf("technology %d has no stable internal key", technologyID)
			}
			nameKey, ok := r.TechnologyNameKeyByID[technologyID]
			if !ok || nameKey == "" {
				return nil, fmt.Errorf("technology %d has no name key", technologyID)
			}
			technologyKeys[i] = key
			technologyNameKeys[i] = nameKey
			description := r.TechnologyDescriptionByID[technologyID]
			if description == "" {
				return nil, fmt.Errorf("technology %d has no original description", technologyID)
			}
			technologyInfo[i] = ResearchTechnologyInfo{
				TechnologyID:      technologyID,
				TechnologyKey:     key,
				TechnologyNameKey: nameKey,
				Description:       description,
				Effects:           r.researchTechnologyEffects(technologyID),
			}
		}
		choices = append(choices, ResearchChoice{
			CategoryID:          categoryID,
			CategoryOrder:       categoryOrder,
			CategoryNameKey:     categoryNameKey,
			TechFieldID:         fieldID,
			PreviousTechFieldID: previousID,
			NextTechFieldID:     r.TechnologyFieldNextID[fieldID],
			BaseCostRP:          r.TechnologyFieldCostsRP[fieldID],
			SelectionMode:       mode,
			TechnologyIDs:       technologyIDs,
			TechnologyKeys:      technologyKeys,
			TechnologyNameKeys:  technologyNameKeys,
			TechnologyInfo:      technologyInfo,
		})
	}
	return choices, nil
}

func (r *EconomyRules) researchTechnologyEffects(technologyID int) []ResearchTechnologyEffect {
	if r == nil || technologyID == 0 {
		return nil
	}
	effects := make([]ResearchTechnologyEffect, 0)
	buildingIDs := make([]string, 0)
	for buildingID, definition := range r.BuildingDefinitions {
		if definition.TechnologyID == technologyID {
			buildingIDs = append(buildingIDs, buildingID)
		}
	}
	sort.Strings(buildingIDs)
	for _, buildingID := range buildingIDs {
		definition := r.BuildingDefinitions[buildingID]
		effects = append(effects, ResearchTechnologyEffect{
			Kind:             "building_unlock",
			ID:               buildingID,
			ProductionCostPP: definition.ProductionCostPP,
			MaintenanceBC:    definition.MaintenanceBC,
		})
	}
	projectIDs := make([]string, 0)
	for projectID, definition := range r.PlanetaryTransformations {
		if definition.TechnologyID == technologyID {
			projectIDs = append(projectIDs, projectID)
		}
	}
	sort.Strings(projectIDs)
	for _, projectID := range projectIDs {
		definition := r.PlanetaryTransformations[projectID]
		effects = append(effects, ResearchTechnologyEffect{Kind: "planetary_project_unlock", ID: projectID, ProductionCostPP: definition.ProductionCostPP})
	}
	for _, definition := range r.ShipDrives {
		if definition.TechnologyID == technologyID {
			effects = append(effects, ResearchTechnologyEffect{Kind: "ship_drive_unlock", ID: definition.ID, FTLSpeed: definition.FTLSpeed})
		}
	}
	for _, definition := range r.ShipComputers {
		if definition.TechnologyID == technologyID {
			effects = append(effects, ResearchTechnologyEffect{Kind: "ship_computer_unlock", ID: definition.ID})
		}
	}
	for _, definition := range r.ShipArmors {
		if definition.TechnologyID == technologyID {
			effects = append(effects, ResearchTechnologyEffect{Kind: "ship_armor_unlock", ID: definition.ID})
		}
	}
	for _, definition := range r.ShipShields {
		if definition.TechnologyID == technologyID {
			effects = append(effects, ResearchTechnologyEffect{Kind: "ship_shield_unlock", ID: definition.ID})
		}
	}
	for _, definition := range r.ShipFuelCells {
		if definition.TechnologyID == technologyID {
			effects = append(effects, ResearchTechnologyEffect{Kind: "ship_fuel_cell_unlock", ID: definition.ID, RangeParsecs: definition.RangeParsecs})
		}
	}
	if bonus, ok := r.PopulationGrowthTechnologyBonusByID[technologyID]; ok && bonus != 0 {
		effects = append(effects, ResearchTechnologyEffect{Kind: "population_growth_bonus", GrowthBonus: bonus})
	}
	if technologyID == r.AdvancedCityPlanningTechnologyID && r.AdvancedCityPlanningCapacityBonus != 0 {
		effects = append(effects, ResearchTechnologyEffect{Kind: "population_capacity_bonus", PopulationBonus: r.AdvancedCityPlanningCapacityBonus})
	}
	return effects
}
func (r *EconomyRules) AvailableResearchCategories() []ResearchCategory {
	if r == nil || len(r.ResearchCategories) == 0 {
		return nil
	}
	return append([]ResearchCategory(nil), r.ResearchCategories...)
}
func (r *EconomyRules) researchSelectionMode(modifiers RaceEconomyModifiers, fieldID int) core.ResearchSelectionMode {
	if r.isHyperAdvancedField(fieldID) {
		return core.ResearchSelectionRepeatField
	}
	if _, general := r.GeneralResearchFieldIDs[fieldID]; general {
		return core.ResearchSelectionAll
	}
	if modifiers.Creative {
		return core.ResearchSelectionAll
	}
	if modifiers.Uncreative {
		return core.ResearchSelectionFixedOne
	}
	return core.ResearchSelectionChooseOne
}

func fixedResearchTechnology(choices []core.FixedResearchChoice, fieldID int) (int, bool) {
	index := sort.Search(len(choices), func(i int) bool { return choices[i].TechFieldID >= fieldID })
	if index >= len(choices) || choices[index].TechFieldID != fieldID {
		return 0, false
	}
	return choices[index].TechnologyID, true
}

func researchChoiceByField(choices []ResearchChoice, fieldID int) (ResearchChoice, bool) {
	index := sort.Search(len(choices), func(i int) bool { return choices[i].TechFieldID >= fieldID })
	if index >= len(choices) || choices[index].TechFieldID != fieldID {
		return ResearchChoice{}, false
	}
	return choices[index], true
}
