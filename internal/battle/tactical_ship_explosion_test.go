package battle

import (
	"encoding/json"
	"reflect"
	"testing"

	"moox/internal/core"
)

func explosionEventData(t *testing.T, event TacticalEvent) map[string]any {
	t.Helper()
	var data map[string]any
	if err := json.Unmarshal(event.Data, &data); err != nil {
		t.Fatal(err)
	}
	return data
}

func explosionEvents(events []TacticalEvent, kind string) []TacticalEvent {
	out := make([]TacticalEvent, 0)
	for _, event := range events {
		if event.Kind == kind {
			out = append(out, event)
		}
	}
	return out
}

func TestSlice178ReviewHullExplosionProfilesMatchOriginalObject2(t *testing.T) {
	want := map[string]struct {
		size  int
		value int
	}{
		"frigate": {0, 2}, "destroyer": {1, 5}, "cruiser": {2, 10},
		"battleship": {3, 15}, "titan": {4, 20}, "doom_star": {5, 40},
	}
	for hull, expected := range want {
		size, value, ok := tacticalHullExplosionProfile(hull)
		if !ok || size != expected.size || value != expected.value {
			t.Fatalf("%s profile=%d/%d/%v want=%d/%d/true", hull, size, value, ok, expected.size, expected.value)
		}
	}
}

func TestSlice178ReviewOrdinaryShipDestructionIsVisualOnly(t *testing.T) {
	spec := baselineTacticalBattleSpec()
	r, err := newTacticalRuntime(*spec.Tactical)
	if err != nil {
		t.Fatal(err)
	}
	source := spec.Tactical.Ships[0]
	target := spec.Tactical.Ships[1]
	sourceState := r.shipState(source.ShipID)
	targetState := r.shipState(target.ShipID)
	sourceState.Destroyed = true
	sourceState.StructureDamage = source.StructureMax
	beforeTarget := *targetState
	beforeEvents := len(r.events)

	if err := tacticalResolveShipDestruction(*spec.Tactical, &r, source.ShipID, source.SeatID, 1, "test"); err != nil {
		t.Fatal(err)
	}
	got := r.events[beforeEvents:]
	if len(got) != 1 || got[0].Kind != "ship_destroyed" {
		t.Fatalf("ordinary destruction events=%+v", got)
	}
	data := explosionEventData(t, got[0])
	if data["explosion_visual"] != true || data["quantum_detonator"] != nil {
		t.Fatalf("ordinary destruction data=%+v", data)
	}
	if !reflect.DeepEqual(*targetState, beforeTarget) {
		t.Fatalf("ordinary ship explosion mutated nearby target: before=%+v after=%+v", beforeTarget, *targetState)
	}
}

func TestSlice178ReviewQuantumDeathExplosionUsesOriginalPowerRadiusAndFriendlyFire(t *testing.T) {
	spec := baselineTacticalBattleSpec()
	spec.Tactical.Rules.Specials = slice178Checkpoint2SpecialRules()
	spec.Tactical.Ships[0].HullID = "frigate"
	spec.Tactical.Ships[0].Specials = []TacticalSpecialSpec{{Slot: 0, SpecialID: "quantum_detonator", TacticalSupport: "tactical_passive"}}
	spec.Tactical.Ships[0].X, spec.Tactical.Ships[0].Y = 10, 10
	spec.Tactical.Ships[1].HullID = "frigate"
	spec.Tactical.Ships[1].X, spec.Tactical.Ships[1].Y = 16, 10
	spec.Tactical.Ships[1].ShieldID = ""
	spec.Tactical.Ships[1].ShieldStrength = 0
	spec.Tactical.Ships[1].ShieldArcMax = 0
	spec.Tactical.Ships[1].ArmorMax = 0
	spec.Tactical.Ships[1].StructureMax = 10000

	friendly := spec.Tactical.Ships[1]
	friendly.ShipID = core.ID(300)
	friendly.EmpireID = spec.Tactical.Ships[0].EmpireID
	friendly.SeatID = spec.Tactical.Ships[0].SeatID
	friendly.X, friendly.Y = 10, 11
	spec.Tactical.Ships = append(spec.Tactical.Ships, friendly)

	far := spec.Tactical.Ships[1]
	far.ShipID = core.ID(301)
	far.X, far.Y = 17, 10
	spec.Tactical.Ships = append(spec.Tactical.Ships, far)

	r, err := newTacticalRuntime(*spec.Tactical)
	if err != nil {
		t.Fatal(err)
	}
	sourceState := r.shipState(spec.Tactical.Ships[0].ShipID)
	sourceState.Destroyed = true
	sourceState.StructureDamage = spec.Tactical.Ships[0].StructureMax
	targetBefore := r.shipState(spec.Tactical.Ships[1].ShipID).StructureDamage
	friendlyBefore := r.shipState(300).StructureDamage
	farBefore := r.shipState(301).StructureDamage

	if err := tacticalResolveShipDestruction(*spec.Tactical, &r, spec.Tactical.Ships[0].ShipID, 1, 1, "test_quantum"); err != nil {
		t.Fatal(err)
	}
	destroyed := explosionEvents(r.events, "ship_destroyed")
	if len(destroyed) != 1 {
		t.Fatalf("ship_destroyed count=%d events=%+v", len(destroyed), r.events)
	}
	data := explosionEventData(t, destroyed[0])
	if data["quantum_detonator"] != true || data["explosion_power"] != float64(30) ||
		data["explosion_radius_pixels"] != float64(135) || data["explosion_pixels_per_field"] != float64(20) {
		t.Fatalf("Quantum destruction metadata=%+v", data)
	}
	damageEvents := explosionEvents(r.events, "battle_damage_applied")
	if len(damageEvents) != 2 {
		t.Fatalf("Quantum affected damage events=%d want=2 events=%+v", len(damageEvents), r.events)
	}
	if r.shipState(spec.Tactical.Ships[1].ShipID).StructureDamage <= targetBefore {
		t.Fatal("enemy in 6-field radius took no Quantum damage")
	}
	if r.shipState(300).StructureDamage <= friendlyBefore {
		t.Fatal("friendly ship in Quantum radius took no damage")
	}
	if r.shipState(301).StructureDamage != farBefore {
		t.Fatal("ship 7 fields away took Quantum damage")
	}
	resolved := explosionEvents(r.events, "ship_explosion_resolved")
	if len(resolved) != 1 {
		t.Fatalf("ship_explosion_resolved count=%d", len(resolved))
	}
}

func TestSlice178ReviewQuantumRawBlastUsesSquaredDistanceFalloff(t *testing.T) {
	if got := tacticalQuantumExplosionRawBlast(300, 0); got != 300 {
		t.Fatalf("center blast=%d want=300", got)
	}
	radiusSquared := tacticalQuantumExplosionRadiusPixels * tacticalQuantumExplosionRadiusPixels
	if got := tacticalQuantumExplosionRawBlast(300, radiusSquared); got != 1 {
		t.Fatalf("edge blast=%d want minimum 1", got)
	}
}
