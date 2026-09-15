package core

import "testing"

func TestDifficultyIDLegacyEmptyMeansNormalWithoutSchemaBump(t *testing.T) {
	state := NewSmallFixture(0x1612)
	state.DifficultyID = ""
	if state.SchemaVersion != 23 || StateSchemaVersion != 23 {
		t.Fatalf("difficulty must not require schema bump: state=%d const=%d", state.SchemaVersion, StateSchemaVersion)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("legacy empty difficulty rejected: %v", err)
	}
	if got := state.EffectiveDifficultyID(); got != DifficultyNormal {
		t.Fatalf("effective difficulty=%q want %q", got, DifficultyNormal)
	}

	state.DifficultyID = DifficultyImpossible
	if err := state.Validate(); err != nil {
		t.Fatalf("explicit supported difficulty rejected: %v", err)
	}
	if got := state.EffectiveDifficultyID(); got != DifficultyImpossible {
		t.Fatalf("effective explicit difficulty=%q", got)
	}

	state.DifficultyID = DifficultyID("nightmare")
	if err := state.Validate(); err == nil {
		t.Fatal("unsupported difficulty_id was accepted")
	}
}
