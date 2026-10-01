package battle

import (
	"encoding/json"
	"testing"

	"moox/internal/core"
)

type checkpoint3DamageEvent struct {
	ApplicationIndex     int `json:"application_index"`
	ApplicationCount     int `json:"application_count"`
	ShieldArc            int `json:"shield_arc"`
	ShieldAbsorbed       int `json:"shield_absorbed"`
	LowerLayerOverflow   int `json:"lower_layer_overflow"`
	ExtraStructureDamage int `json:"extra_structure_damage"`
	MarineBefore         int `json:"marine_before"`
	MarineAfter          int `json:"marine_after"`
	MarineKilled         int `json:"marine_killed"`
}

func checkpoint3DoomStarTarget(withShield bool) TacticalShipSpec {
	target := TacticalShipSpec{
		ShipID: 200, EmpireID: 2, SeatID: 2, X: 11, Y: 10,
		HullID: "doom_star", WarpDriveID: "nuclear_drive", ComputerID: "electronic_computer",
		ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells",
		MinimumCombatSpeed: 3, CurrentCombatSpeed: 3, BeamOffense: 25, BeamDefense: 15,
		ArmorMax: 150, StructureMax: 150, MarineMax: 50,
	}
	if withShield {
		target.ShieldID = "class_i_shield"
		target.ShieldStrength = 1
		target.ShieldArcMax = 35
	}
	return target
}

func checkpoint3FireWeapon(t *testing.T, weapon TacticalWeaponSpec, withShield bool) View {
	t.Helper()
	spec := baselineTacticalBattleSpec()
	weapon.Slot = 0
	weapon.Count = 1
	weapon.WeaponFamily = "beam"
	weapon.TacticalSupport = "beam"
	weapon.FiringArcID = core.ShipWeaponFiringArc360
	if weapon.MountType == "" {
		weapon.MountType = core.ShipWeaponMountStandard
	}
	spec.Tactical.Ships[0].Weapons[0] = weapon
	spec.Tactical.Ships[1] = checkpoint3DoomStarTarget(withShield)

	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	submitPrepared(t, s, 1, mustFireCommand(t, 1, 100, 200, 0))
	return s.View()
}

func checkpoint3DamageEvents(t *testing.T, view View) []checkpoint3DamageEvent {
	t.Helper()
	var out []checkpoint3DamageEvent
	for _, event := range view.Tactical.Events {
		if event.Kind != "battle_damage_applied" || event.CommandSequence != 1 {
			continue
		}
		var decoded checkpoint3DamageEvent
		if err := json.Unmarshal(event.Data, &decoded); err != nil {
			t.Fatal(err)
		}
		out = append(out, decoded)
	}
	return out
}

func TestSlice176Checkpoint3ParticleBeamPiercesShield(t *testing.T) {
	view := checkpoint3FireWeapon(t, TacticalWeaponSpec{
		WeaponID: "particle_beam", MinDamage: 10, MaxDamage: 30,
		InnateEffectIDs: []string{"natural_shield_piercing"},
	}, true)

	target := tacticalStateForTest(t, view.Tactical.State.Ships, 200)
	if target.ShieldArcCurrent != [4]int{35, 35, 35, 35} {
		t.Fatalf("Particle Shield Piercing changed shield pools: %v", target.ShieldArcCurrent)
	}
	if target.ArmorCurrent != 120 {
		t.Fatalf("Particle armor=%d want 120", target.ArmorCurrent)
	}
	events := checkpoint3DamageEvents(t, view)
	if len(events) != 1 || events[0].ShieldAbsorbed != 0 || events[0].LowerLayerOverflow != 30 {
		t.Fatalf("Particle damage event=%+v", events)
	}
}

func TestSlice176Checkpoint3ArmorPiercingBypassesArmorFirst(t *testing.T) {
	view := checkpoint3FireWeapon(t, TacticalWeaponSpec{
		WeaponID: "laser_cannon", MinDamage: 1, MaxDamage: 4,
		ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierArmorPiercing},
	}, false)

	target := tacticalStateForTest(t, view.Tactical.State.Ships, 200)
	if target.ArmorCurrent != 150 || target.StructureDamage != 4 {
		t.Fatalf("Armor Piercing target armor/structure=%d/%d want 150/4", target.ArmorCurrent, target.StructureDamage)
	}
}

func TestSlice176Checkpoint3GravitonAddsPostLayerStructureDamage(t *testing.T) {
	view := checkpoint3FireWeapon(t, TacticalWeaponSpec{
		WeaponID: "graviton_beam", MinDamage: 3, MaxDamage: 15,
		InnateEffectIDs: []string{"extra_structure_damage"},
	}, false)

	target := tacticalStateForTest(t, view.Tactical.State.Ships, 200)
	if target.ArmorCurrent != 135 || target.StructureDamage != 15 {
		t.Fatalf("Graviton target armor/structure=%d/%d want 135/15", target.ArmorCurrent, target.StructureDamage)
	}
	events := checkpoint3DamageEvents(t, view)
	if len(events) != 1 || events[0].ExtraStructureDamage != 15 {
		t.Fatalf("Graviton damage event=%+v", events)
	}
}

func TestSlice176Checkpoint3NeutronKillsMarinesFromPostShieldDamage(t *testing.T) {
	view := checkpoint3FireWeapon(t, TacticalWeaponSpec{
		WeaponID: "neutron_blaster", MinDamage: 3, MaxDamage: 12,
		InnateEffectIDs: []string{"marine_kill"},
	}, false)

	target := tacticalStateForTest(t, view.Tactical.State.Ships, 200)
	if target.MarineCurrent != 48 {
		t.Fatalf("Neutron Marines=%d want 48", target.MarineCurrent)
	}
	events := checkpoint3DamageEvents(t, view)
	if len(events) != 1 || events[0].MarineBefore != 50 || events[0].MarineAfter != 48 || events[0].MarineKilled != 2 {
		t.Fatalf("Neutron marine event=%+v", events)
	}
}

func TestSlice176Checkpoint3PlasmaEnvelopingAppliesAllFourFacings(t *testing.T) {
	view := checkpoint3FireWeapon(t, TacticalWeaponSpec{
		WeaponID: "plasma_cannon", MinDamage: 6, MaxDamage: 30,
		InnateEffectIDs: []string{"all_four_shields", "double_range_penalty"},
	}, false)

	target := tacticalStateForTest(t, view.Tactical.State.Ships, 200)
	if target.ArmorCurrent != 30 {
		t.Fatalf("Plasma Enveloping armor=%d want 30", target.ArmorCurrent)
	}
	events := checkpoint3DamageEvents(t, view)
	if len(events) != 4 {
		t.Fatalf("Plasma damage event count=%d want 4", len(events))
	}
	for i, event := range events {
		if event.ApplicationIndex != i+1 || event.ApplicationCount != 4 || event.ShieldArc != i {
			t.Fatalf("Plasma event[%d]=%+v", i, event)
		}
	}
}

func TestSlice176Checkpoint3DeathRayKillsMarinesFromPostShieldDamage(t *testing.T) {
	view := checkpoint3FireWeapon(t, TacticalWeaponSpec{
		WeaponID: "death_ray", MinDamage: 50, MaxDamage: 100,
		InnateEffectIDs: []string{"marine_kill"},
	}, false)

	target := tacticalStateForTest(t, view.Tactical.State.Ships, 200)
	if target.MarineCurrent != 30 {
		t.Fatalf("Death Ray Marines=%d want 30", target.MarineCurrent)
	}
	events := checkpoint3DamageEvents(t, view)
	if len(events) != 1 || events[0].MarineBefore != 50 || events[0].MarineAfter != 30 || events[0].MarineKilled != 20 {
		t.Fatalf("Death Ray marine event=%+v", events)
	}
}

type checkpoint3BeamEvent struct {
	RangeIndex      int `json:"range_index"`
	Damage          int `json:"damage"`
	DamagePercent   int `json:"damage_percent"`
	ScaledMinDamage int `json:"scaled_min_damage"`
	ScaledMaxDamage int `json:"scaled_max_damage"`
}

func checkpoint3FireWeaponAtTargetX(t *testing.T, weapon TacticalWeaponSpec, withShield bool, targetX int) View {
	t.Helper()
	spec := baselineTacticalBattleSpec()
	weapon.Slot = 0
	weapon.Count = 1
	weapon.WeaponFamily = "beam"
	weapon.TacticalSupport = "beam"
	weapon.FiringArcID = core.ShipWeaponFiringArc360
	if weapon.MountType == "" {
		weapon.MountType = core.ShipWeaponMountStandard
	}
	spec.Tactical.Ships[0].Weapons[0] = weapon
	target := checkpoint3DoomStarTarget(withShield)
	target.X = targetX
	spec.Tactical.Ships[1] = target

	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	submitPrepared(t, s, 1, mustFireCommand(t, 1, 100, 200, 0))
	return s.View()
}

func checkpoint3BeamEvents(t *testing.T, view View) []checkpoint3BeamEvent {
	t.Helper()
	var out []checkpoint3BeamEvent
	for _, event := range view.Tactical.Events {
		if event.Kind != "beam_fired" || event.CommandSequence != 1 {
			continue
		}
		var decoded checkpoint3BeamEvent
		if err := json.Unmarshal(event.Data, &decoded); err != nil {
			t.Fatal(err)
		}
		out = append(out, decoded)
	}
	return out
}

func TestSlice176Checkpoint3HeavyMountUsesTransformedRangeAndDamagePercent(t *testing.T) {
	view := checkpoint3FireWeaponAtTargetX(t, TacticalWeaponSpec{
		WeaponID: "particle_beam", MinDamage: 10, MaxDamage: 30,
		MountType:       core.ShipWeaponMountHeavy,
		InnateEffectIDs: []string{"natural_shield_piercing"},
	}, false, 22)

	events := checkpoint3BeamEvents(t, view)
	if len(events) != 1 {
		t.Fatalf("Heavy Beam event count=%d want 1", len(events))
	}
	got := events[0]
	if got.RangeIndex != 2 || got.DamagePercent != 140 || got.ScaledMinDamage != 14 || got.ScaledMaxDamage != 42 {
		t.Fatalf("Heavy Beam event=%+v want range=2 percent=140 scaled=14-42", got)
	}
}

func TestSlice176Checkpoint3PointDefenseMountUsesShipTargetProfile(t *testing.T) {
	view := checkpoint3FireWeaponAtTargetX(t, TacticalWeaponSpec{
		WeaponID: "laser_cannon", MinDamage: 1, MaxDamage: 4,
		MountType: core.ShipWeaponMountPointDefense,
	}, false, 16)

	events := checkpoint3BeamEvents(t, view)
	if len(events) != 1 {
		t.Fatalf("PD Beam event count=%d want 1", len(events))
	}
	got := events[0]
	if got.RangeIndex != 4 || got.DamagePercent != 20 || got.ScaledMinDamage != 1 || got.ScaledMaxDamage != 1 || got.Damage != 1 {
		t.Fatalf("PD Beam event=%+v want range=4 percent=20 scaled=1-1 damage=1", got)
	}
}

func TestSlice176Checkpoint3SelectedNoRangeDissipationSuppressesDamagePenalty(t *testing.T) {
	view := checkpoint3FireWeaponAtTargetX(t, TacticalWeaponSpec{
		WeaponID: "laser_cannon", MinDamage: 1, MaxDamage: 4,
		ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierNoRangeDissipation},
	}, false, 22)

	events := checkpoint3BeamEvents(t, view)
	if len(events) != 1 {
		t.Fatalf("No Range Dissipation event count=%d want 1", len(events))
	}
	got := events[0]
	if got.RangeIndex != 4 || got.DamagePercent != 100 || got.ScaledMinDamage != 1 || got.ScaledMaxDamage != 4 {
		t.Fatalf("No Range Dissipation event=%+v want range=4 percent=100 scaled=1-4", got)
	}
}

func TestSlice176Checkpoint3SelectedShieldPiercingBypassesShield(t *testing.T) {
	view := checkpoint3FireWeaponAtTargetX(t, TacticalWeaponSpec{
		WeaponID: "phasor", MinDamage: 5, MaxDamage: 20,
		ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierShieldPiercing},
	}, true, 11)

	target := tacticalStateForTest(t, view.Tactical.State.Ships, 200)
	if target.ShieldArcCurrent != [4]int{35, 35, 35, 35} {
		t.Fatalf("selected Shield Piercing changed shield pools: %v", target.ShieldArcCurrent)
	}
	if target.ArmorCurrent != 130 {
		t.Fatalf("selected Shield Piercing armor=%d want 130", target.ArmorCurrent)
	}
}

type checkpoint3BeamResolutionEvent struct {
	RangeIndex       int `json:"range_index"`
	RawHitRoll       int `json:"raw_hit_roll"`
	EffectiveHitRoll int `json:"effective_hit_roll"`
	Threshold        int `json:"threshold"`
	Damage           int `json:"damage"`
	DamagePercent    int `json:"damage_percent"`
}

func checkpoint3BeamResolutionEvents(t *testing.T, view View) []checkpoint3BeamResolutionEvent {
	t.Helper()
	var out []checkpoint3BeamResolutionEvent
	for _, event := range view.Tactical.Events {
		if event.Kind != "beam_fired" || event.CommandSequence != 1 {
			continue
		}
		var decoded checkpoint3BeamResolutionEvent
		if err := json.Unmarshal(event.Data, &decoded); err != nil {
			t.Fatal(err)
		}
		out = append(out, decoded)
	}
	return out
}

func checkpoint3FireWeaponAtTargetXWithRawAtMost(t *testing.T, weapon TacticalWeaponSpec, targetX, maxRaw int) View {
	t.Helper()
	spec := baselineTacticalBattleSpec()
	weapon.Slot = 0
	weapon.Count = 1
	weapon.WeaponFamily = "beam"
	weapon.TacticalSupport = "beam"
	weapon.FiringArcID = core.ShipWeaponFiringArc360
	if weapon.MountType == "" {
		weapon.MountType = core.ShipWeaponMountStandard
	}
	spec.Tactical.Ships[0].Weapons[0] = weapon
	target := checkpoint3DoomStarTarget(false)
	target.X = targetX
	spec.Tactical.Ships[1] = target

	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}

	s.mu.Lock()
	chosen := uint32(0)
	for seed := uint32(1); seed < 100000; seed++ {
		probe := cloneTacticalRuntime(*s.tactical)
		probe.state.RNGState = seed
		raw, err := probe.random(100, s.spec.Tactical.Rules)
		if err == nil && raw <= maxRaw {
			chosen = seed
			break
		}
	}
	if chosen == 0 {
		s.mu.Unlock()
		t.Fatalf("failed to find Tactical RNG seed with next raw <= %d", maxRaw)
	}
	s.tactical.state.RNGState = chosen
	s.mu.Unlock()

	submitPrepared(t, s, 1, mustFireCommand(t, 1, 100, 200, 0))
	return s.View()
}

func TestSlice176Checkpoint3IonPulseInnateArmorBypassUsesNeutralMonsterHook(t *testing.T) {
	view := checkpoint3FireWeaponAtTargetX(t, TacticalWeaponSpec{
		WeaponID: "ion_pulse_cannon", MinDamage: 2, MaxDamage: 10,
		InnateEffectIDs: []string{"armor_structure_bypass", "monster_antaran_immunity_exception"},
	}, false, 11)

	target := tacticalStateForTest(t, view.Tactical.State.Ships, 200)
	if target.ArmorCurrent != 150 || target.StructureDamage != 10 {
		t.Fatalf("Ion Pulse target armor/structure=%d/%d want 150/10", target.ArmorCurrent, target.StructureDamage)
	}
}

func TestSlice176Checkpoint3FusionDoublesSignedRangeToHit(t *testing.T) {
	view := checkpoint3FireWeaponAtTargetX(t, TacticalWeaponSpec{
		WeaponID: "fusion_beam", MinDamage: 2, MaxDamage: 6,
	}, false, 22)

	events := checkpoint3BeamResolutionEvents(t, view)
	if len(events) != 1 {
		t.Fatalf("Fusion Beam event count=%d want 1", len(events))
	}
	if events[0].RangeIndex != 4 || events[0].Threshold != 95 {
		t.Fatalf("Fusion Beam event=%+v want range=4 threshold=95", events[0])
	}
}

func TestSlice176Checkpoint3MaulerHeavyConsumesLowRollThenAlwaysHits(t *testing.T) {
	view := checkpoint3FireWeaponAtTargetXWithRawAtMost(t, TacticalWeaponSpec{
		WeaponID: "mauler_device", MinDamage: 100, MaxDamage: 100,
		MountType:       core.ShipWeaponMountHeavy,
		InnateEffectIDs: []string{"always_hit", "double_range_dissipation"},
	}, 22, 10)

	events := checkpoint3BeamResolutionEvents(t, view)
	if len(events) != 1 {
		t.Fatalf("Mauler event count=%d want 1", len(events))
	}
	got := events[0]
	if got.RawHitRoll > 10 || got.EffectiveHitRoll != 100 || got.Threshold != 60 ||
		got.RangeIndex != 2 || got.DamagePercent != 140 || got.Damage != 140 {
		t.Fatalf("Mauler event=%+v want low raw, effective=100 threshold=60 range=2 percent=140 damage=140", got)
	}
}
