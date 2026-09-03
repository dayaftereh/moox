package core

import "testing"

func TestDiplomacySchema23ValidationAndRoundTrip(t *testing.T) {
	state := NewSmallFixture(0xD122)
	first := state.Empires[0].ID
	second := state.NewID()
	state.Empires = append(state.Empires, Empire{ID: second, Name: "Second", RaceID: "darlok"})

	state.DiplomaticRelations = reciprocalRelations(first, second, DiplomaticStanceWar)
	state.DiplomaticPeaceOffers = []DiplomaticPeaceOffer{{FromEmpireID: first, ToEmpireID: second}}
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	if !state.MayAttackEmpire(first, second) || !state.HasDiplomaticPeaceOffer(first, second) {
		t.Fatal("war/offer helpers do not reflect authoritative state")
	}
	encoded, err := MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.SchemaVersion != 23 || !loaded.HasDiplomaticPeaceOffer(first, second) || loaded.DiplomaticStanceBetween(second, first) != DiplomaticStanceWar {
		t.Fatalf("roundtrip lost diplomacy: %+v", loaded)
	}

	state.DiplomaticPeaceOffers = nil
	state.DiplomaticRelations = reciprocalRelations(first, second, DiplomaticStancePeace)
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	if state.MayAttackEmpire(first, second) {
		t.Fatal("peace must not authorize attack")
	}
}

func TestDiplomacySchema23RejectsNonCanonicalState(t *testing.T) {
	makeState := func() (*GameState, ID, ID) {
		state := NewSmallFixture(0xD123)
		first := state.Empires[0].ID
		second := state.NewID()
		state.Empires = append(state.Empires, Empire{ID: second, Name: "Second", RaceID: "darlok"})
		return state, first, second
	}
	t.Run("explicit neutral row", func(t *testing.T) {
		state, first, second := makeState()
		state.DiplomaticRelations = []DiplomaticRelation{{FromEmpireID: first, ToEmpireID: second, Stance: DiplomaticStanceNeutral}}
		if err := state.Validate(); err == nil {
			t.Fatal("expected explicit neutral relation to fail")
		}
	})
	t.Run("asymmetric war", func(t *testing.T) {
		state, first, second := makeState()
		state.DiplomaticRelations = []DiplomaticRelation{{FromEmpireID: first, ToEmpireID: second, Stance: DiplomaticStanceWar}}
		if err := state.Validate(); err == nil {
			t.Fatal("expected asymmetric war to fail")
		}
	})
	t.Run("peace offer outside war", func(t *testing.T) {
		state, first, second := makeState()
		state.DiplomaticRelations = reciprocalRelations(first, second, DiplomaticStancePeace)
		state.DiplomaticPeaceOffers = []DiplomaticPeaceOffer{{FromEmpireID: first, ToEmpireID: second}}
		if err := state.Validate(); err == nil {
			t.Fatal("expected peace offer outside war to fail")
		}
	})
	t.Run("reciprocal offers", func(t *testing.T) {
		state, first, second := makeState()
		state.DiplomaticRelations = reciprocalRelations(first, second, DiplomaticStanceWar)
		state.DiplomaticPeaceOffers = []DiplomaticPeaceOffer{{FromEmpireID: first, ToEmpireID: second}, {FromEmpireID: second, ToEmpireID: first}}
		if err := state.Validate(); err == nil {
			t.Fatal("expected reciprocal pending offers to fail")
		}
	})
}
