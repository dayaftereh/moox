package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const TechnologiesSchemaVersion = 4

type TechnologiesFile struct {
	SchemaVersion int                    `json:"schema_version"`
	Ruleset       string                 `json:"ruleset"`
	Sources       []Source               `json:"sources"`
	NewGameStart  NewGameTechnologyStart `json:"new_game_start"`
	HyperAdvanced HyperAdvancedResearch  `json:"hyper_advanced"`
	AIResearch    TechnologyAIResearch   `json:"ai_research"`
	Fields        []TechnologyField      `json:"technology_fields"`
	Technologies  []Technology           `json:"technologies"`
}

type NewGameTechnologyStart struct {
	AlwaysKnownTechFieldID  int             `json:"always_known_tech_field_id"`
	StagedKnownTechFieldIDs []int           `json:"staged_known_tech_field_ids"`
	Verification            string          `json:"verification"`
	Source                  FieldProvenance `json:"source"`
}

type HyperAdvancedResearch struct {
	TechFieldIDs    []int           `json:"tech_field_ids"`
	CostIncrementRP int             `json:"cost_increment_rp"`
	Verification    string          `json:"verification"`
	Source          FieldProvenance `json:"source"`
}

type TechnologyAIResearch struct {
	TechnologyClasses []TechnologyAIClass `json:"technology_classes"`
	FieldGroupValues  []int               `json:"field_group_values"`
	Verification      string              `json:"verification"`
	Source            FieldProvenance     `json:"source"`
}

type TechnologyAIClass struct {
	ClassID              int  `json:"class_id"`
	BaseWeight           int  `json:"base_weight"`
	CompetitionSensitive bool `json:"competition_sensitive"`
}

type TechnologyField struct {
	FieldID         int             `json:"field_id"`
	PreviousID      int             `json:"previous_id"`
	NextID          int             `json:"next_id"`
	ResearchCost    int             `json:"research_cost_rp"`
	AIGroup         int             `json:"ai_group"`
	CategoryID      string          `json:"category_id"`
	CategoryOrder   int             `json:"category_order"`
	CategoryNameKey string          `json:"category_name_key"`
	Source          FieldProvenance `json:"source"`
}

type Technology struct {
	ID                       string          `json:"id"`
	Order                    int             `json:"order"`
	TechnologyID             int             `json:"technology_id"`
	TechFieldID              int             `json:"tech_field_id"`
	AIClass                  int             `json:"ai_class"`
	AIClassSource            FieldProvenance `json:"ai_class_source"`
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
	if len(f.HyperAdvanced.TechFieldIDs) != 8 || f.HyperAdvanced.CostIncrementRP != 10000 || f.HyperAdvanced.Verification == "" || f.HyperAdvanced.Source.SourceID == "" {
		return fmt.Errorf("hyper-advanced research metadata is incomplete")
	}
	for i, fieldID := range f.HyperAdvanced.TechFieldIDs {
		if fieldID != 75+i {
			return fmt.Errorf("hyper-advanced tech field[%d]=%d want=%d", i, fieldID, 75+i)
		}
	}
	if len(f.AIResearch.TechnologyClasses) != 41 || len(f.AIResearch.FieldGroupValues) != 23 || f.AIResearch.Verification == "" || f.AIResearch.Source.SourceID == "" {
		return fmt.Errorf("technology AI research metadata is incomplete")
	}
	for i, class := range f.AIResearch.TechnologyClasses {
		if class.ClassID != i || class.BaseWeight <= 0 {
			return fmt.Errorf("technology AI class[%d] is invalid: %+v", i, class)
		}
	}
	lastFieldGroupValue := -1
	for i, value := range f.AIResearch.FieldGroupValues {
		if value < 0 || value < lastFieldGroupValue {
			return fmt.Errorf("technology AI field-group value[%d]=%d is invalid", i, value)
		}
		lastFieldGroupValue = value
	}
	if len(f.Fields) != 82 {
		return fmt.Errorf("expected 82 technology fields, got %d", len(f.Fields))
	}
	for i, field := range f.Fields {
		categoryInvalid := field.FieldID != 74 && (field.CategoryID == "" || field.CategoryOrder < 0 || field.CategoryOrder > 7 || field.CategoryNameKey == "")
		if field.FieldID != i+1 || field.PreviousID < 0 || field.PreviousID > 82 || field.NextID < 0 || field.NextID > 82 || field.ResearchCost <= 0 || categoryInvalid || field.Source.SourceID == "" {
			return fmt.Errorf("technology field %d is invalid", i+1)
		}
	}
	if len(f.Technologies) != 203 {
		return fmt.Errorf("expected 203 technologies, got %d", len(f.Technologies))
	}
	ids := make(map[string]struct{}, len(f.Technologies))
	keys := make(map[string]struct{}, len(f.Technologies))
	for order, tech := range f.Technologies {
		if tech.ID == "" || tech.NameKey == "" || tech.NameSource.SourceID == "" || tech.TechFieldSource.SourceID == "" || tech.AIClassSource.SourceID == "" || tech.StrategicCombatSource.SourceID == "" {
			return fmt.Errorf("technology id, name_key and source fields are required")
		}
		if tech.TechFieldID < -1 || tech.TechFieldID > 82 {
			return fmt.Errorf("technology %q tech_field_id=%d outside [-1,82]", tech.ID, tech.TechFieldID)
		}
		if tech.AIClass < 0 || tech.AIClass >= len(f.AIResearch.TechnologyClasses) {
			return fmt.Errorf("technology %q ai_class=%d outside [0,%d]", tech.ID, tech.AIClass, len(f.AIResearch.TechnologyClasses)-1)
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
