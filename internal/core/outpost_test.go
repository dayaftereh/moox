package core

import (
	"bytes"
	"testing"
)

func TestOutpostStateRoundTripsDeterministicallyInSchema23(t *testing.T) {
	state := NewSmallFixture(1710)
	target := &state.Galaxy.Systems[1].Planets[0]
	outpostID := state.NewID()
	state.Outposts = append(state.Outposts, Outpost{ID: outpostID, EmpireID: state.Empires[0].ID, PlanetID: target.ID})
	target.OutpostID = outpostID

	if StateSchemaVersion != 23 || state.SchemaVersion != 23 {
		t.Fatalf("schema=%d constant=%d want=22", state.SchemaVersion, StateSchemaVersion)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("schema-18 Outpost state invalid: %v", err)
	}
	encoded, err := MarshalState(state)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := UnmarshalState(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.Outposts) != 1 || decoded.Outposts[0].ID != outpostID || decoded.Outposts[0].EmpireID != state.Empires[0].ID || decoded.Outposts[0].PlanetID != target.ID {
		t.Fatalf("decoded Outposts=%+v", decoded.Outposts)
	}
	if decoded.Galaxy.Systems[1].Planets[0].OutpostID != outpostID {
		t.Fatalf("decoded Planet outpost link=%d want=%d", decoded.Galaxy.Systems[1].Planets[0].OutpostID, outpostID)
	}
	reencoded, err := MarshalState(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, reencoded) {
		t.Fatalf("schema-18 Outpost bytes changed across round trip\nfirst=%s\nsecond=%s", encoded, reencoded)
	}
}

func TestValidateRejectsInvalidOutpostPlanetLinks(t *testing.T) {
	t.Run("colony and outpost share planet", func(t *testing.T) {
		state := NewSmallFixture(1711)
		home := &state.Galaxy.Systems[0].Planets[0]
		outpostID := state.NewID()
		state.Outposts = append(state.Outposts, Outpost{ID: outpostID, EmpireID: state.Empires[0].ID, PlanetID: home.ID})
		home.OutpostID = outpostID
		if err := state.Validate(); err == nil {
			t.Fatal("state with Colony and Outpost on one Planet unexpectedly validated")
		}
	})

	t.Run("dangling planet outpost id", func(t *testing.T) {
		state := NewSmallFixture(1712)
		state.Galaxy.Systems[1].Planets[0].OutpostID = 999999
		if err := state.Validate(); err == nil {
			t.Fatal("Planet with dangling outpost_id unexpectedly validated")
		}
	})

	t.Run("missing reciprocal planet link", func(t *testing.T) {
		state := NewSmallFixture(1713)
		target := &state.Galaxy.Systems[1].Planets[0]
		outpostID := state.NewID()
		state.Outposts = append(state.Outposts, Outpost{ID: outpostID, EmpireID: state.Empires[0].ID, PlanetID: target.ID})
		if err := state.Validate(); err == nil {
			t.Fatal("Outpost without reciprocal Planet.OutpostID unexpectedly validated")
		}
	})
}
