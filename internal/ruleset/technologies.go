package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const TechnologiesSchemaVersion = 2

type TechnologiesFile struct {
	SchemaVersion int                    `json:"schema_version"`
	Ruleset       string                 `json:"ruleset"`
	Sources       []Source               `json:"sources"`
	NewGameStart  NewGameTechnologyStart `json:"new_game_start"`
	Fields        []TechnologyField      `json:"technology_fields"`
	Technologies  []Technology           `json:"technologies"`
}

type NewGameTechnologyStart struct {
	AlwaysKnownTechFieldID  int             `json:"always_known_tech_field_id"`
	StagedKnownTechFieldIDs []int           `json:"staged_known_tech_field_ids"`
	Verification            string          `json:"verification"`
	Source                  FieldProvenance `json:"source"`
}

type TechnologyField struct {
	FieldID      int             `json:"field_id"`
	PreviousID   int             `json:"previous_id"`
	NextID       int             `json:"next_id"`
	ResearchCost int             `json:"research_cost_rp"`
	AIGroup      int             `json:"ai_group"`
	Source       FieldProvenance `json:"source"`
}

type Technology struct {
	ID                       string          `json:"id"`
	Order                    int             `json:"order"`
	TechnologyID             int             `json:"technology_id"`
	TechFieldID              int             `json:"tech_field_id"`
	TechFieldSource          FieldProvenance `json:"tech_field_source"`
	StrategicCombatAvailable bool            `json:"strategic_combat_available"`
	StrategicCombatSource    FieldProvenance `json:"strategic_combat_source"`
	NameKey                  string          `json:"name_key"`
	NameSource               FieldProvenance `json:"name_source"`
}

func LoadTechnologies(path string) (*TechnologiesFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file TechnologiesFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *TechnologiesFile) Validate() error {
	if f.SchemaVersion != TechnologiesSchemaVersion {
		return fmt.Errorf("unsupported technologies schema version %d", f.SchemaVersion)
	}
	if f.Ruleset == "" {
		return fmt.Errorf("ruleset is required")
	}
	if f.NewGameStart.AlwaysKnownTechFieldID != 0 || len(f.NewGameStart.StagedKnownTechFieldIDs) != 6 || f.NewGameStart.Verification == "" || f.NewGameStart.Source.SourceID == "" {
		return fmt.Errorf("new-game technology start metadata is incomplete")
	}
	wantStart := []int{29, 55, 22, 57, 28, 23}
	for i, want := range wantStart {
		if f.NewGameStart.StagedKnownTechFieldIDs[i] != want {
			return fmt.Errorf("new-game staged tech field[%d]=%d want=%d", i, f.NewGameStart.StagedKnownTechFieldIDs[i], want)
		}
	}
	if len(f.Fields) != 82 {
		return fmt.Errorf("expected 82 technology fields, got %d", len(f.Fields))
	}
	for i, field := range f.Fields {
		if field.FieldID != i+1 || field.PreviousID < 0 || field.PreviousID > 82 || field.NextID < 0 || field.NextID > 82 || field.ResearchCost <= 0 || field.Source.SourceID == "" {
			return fmt.Errorf("technology field %d is invalid", i+1)
		}
	}
	if len(f.Technologies) != 203 {
		return fmt.Errorf("expected 203 technologies, got %d", len(f.Technologies))
	}
	ids := make(map[string]struct{}, len(f.Technologies))
	keys := make(map[string]struct{}, len(f.Technologies))
	for order, tech := range f.Technologies {
		if tech.ID == "" || tech.NameKey == "" || tech.NameSource.SourceID == "" || tech.TechFieldSource.SourceID == "" || tech.StrategicCombatSource.SourceID == "" {
			return fmt.Errorf("technology id, name_key and source fields are required")
		}
		if tech.TechFieldID < -1 || tech.TechFieldID > 82 {
			return fmt.Errorf("technology %q tech_field_id=%d outside [-1,82]", tech.ID, tech.TechFieldID)
		}
		if tech.Order != order || tech.TechnologyID != order+1 {
			return fmt.Errorf("technology %q order/id mismatch: order=%d technology_id=%d expected=%d/%d", tech.ID, tech.Order, tech.TechnologyID, order, order+1)
		}
		if _, exists := ids[tech.ID]; exists {
			return fmt.Errorf("duplicate technology id %q", tech.ID)
		}
		ids[tech.ID] = struct{}{}
		if _, exists := keys[tech.NameKey]; exists {
			return fmt.Errorf("duplicate technology name key %q", tech.NameKey)
		}
		keys[tech.NameKey] = struct{}{}
	}
	return nil
}
