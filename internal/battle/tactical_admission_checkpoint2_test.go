package battle

import (
	"strings"
	"testing"

	"moox/internal/core"
)

func TestSlice175Checkpoint2BattleValidatorAcceptsBroadDerivedCombatants(t *testing.T) {
	spec := baselineTacticalBattleSpec()
	spec.Tactical.Ships[0] = TacticalShipSpec{
		ShipID: 100, EmpireID: 1, SeatID: 1, X: 10, Y: 10,
		HullID: "doom_star", WarpDriveID: "interphased_drive", ComputerID: "moleculartronic_computer",
		ArmorID: "xentronium_armor", ShieldID: "class_x_shield", FuelCellID: "thorium_fuel_cells",
		MinimumCombatSpeed: 13, CurrentCombatSpeed: 23, BeamOffense: 125, BeamDefense: 115,
		ArmorMax: 1500, StructureMax: 150, ShieldStrength: 10, ShieldArcMax: 350,
		Weapons: []TacticalWeaponSpec{{Slot: 0, WeaponID: "fighter_bays", Count: 1, TacticalSupport: "metadata_only"}},
	}
	spec.Attacker.ShipIDs = []core.ID{100}
	if _, err := NewSession(spec); err != nil {
		t.Fatalf("broad derived combatant rejected: %v", err)
	}
}

func TestSlice175Checkpoint2BattleValidatorRejectsUnknownAndImpossibleDerivedValues(t *testing.T) {
	tests := []struct {
		name  string
		apply func(*TacticalShipSpec)
		want  string
	}{
		{name: "hull", apply: func(s *TacticalShipSpec) { s.HullID = "unknown_hull" }, want: "unknown hull"},
		{name: "drive", apply: func(s *TacticalShipSpec) { s.WarpDriveID = "unknown_drive" }, want: "unknown drive"},
		{name: "computer", apply: func(s *TacticalShipSpec) { s.ComputerID = "unknown_computer" }, want: "unknown computer"},
		{name: "armor", apply: func(s *TacticalShipSpec) { s.ArmorID = "unknown_armor" }, want: "unknown armor"},
		{name: "shield", apply: func(s *TacticalShipSpec) { s.ShieldID = "unknown_shield" }, want: "unknown shield"},
		{name: "fuel", apply: func(s *TacticalShipSpec) { s.FuelCellID = "unknown_fuel" }, want: "unknown fuel"},
		{name: "minimum speed", apply: func(s *TacticalShipSpec) { s.MinimumCombatSpeed++ }, want: "minimum combat speed"},
		{name: "current speed", apply: func(s *TacticalShipSpec) { s.CurrentCombatSpeed = s.MinimumCombatSpeed - 1 }, want: "current combat speed"},
		{name: "offense", apply: func(s *TacticalShipSpec) { s.BeamOffense++ }, want: "Beam Offense"},
		{name: "defense", apply: func(s *TacticalShipSpec) { s.BeamDefense++ }, want: "Beam Defense"},
		{name: "armor max", apply: func(s *TacticalShipSpec) { s.ArmorMax++ }, want: "durability"},
		{name: "structure max", apply: func(s *TacticalShipSpec) { s.StructureMax++ }, want: "durability"},
		{name: "shield strength", apply: func(s *TacticalShipSpec) { s.ShieldStrength = 1 }, want: "shield strength/arc"},
		{name: "shield arc", apply: func(s *TacticalShipSpec) { s.ShieldArcMax = 5 }, want: "shield strength/arc"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec := baselineTacticalBattleSpec()
			tc.apply(&spec.Tactical.Ships[0])
			_, err := NewSession(spec)
			if err == nil {
				t.Fatal("expected invalid combatant to reject")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%q want substring %q", err, tc.want)
			}
		})
	}
}

func TestSlice175Checkpoint2BattleValidatorRejectsMalformedSpecialSnapshots(t *testing.T) {
	tests := []struct {
		name     string
		specials []TacticalSpecialSpec
		want     string
	}{
		{
			name: "duplicate special",
			specials: []TacticalSpecialSpec{
				{Slot: 0, SpecialID: "battle_pods", TacticalSupport: "strategic_only"},
				{Slot: 1, SpecialID: "battle_pods", TacticalSupport: "strategic_only"},
			},
			want: "duplicate special",
		},
		{
			name:     "invalid support",
			specials: []TacticalSpecialSpec{{Slot: 0, SpecialID: "battle_pods", TacticalSupport: "future_magic"}},
			want:     "invalid support mode",
		},
		{
			name: "descending slots",
			specials: []TacticalSpecialSpec{
				{Slot: 1, SpecialID: "battle_pods", TacticalSupport: "strategic_only"},
				{Slot: 0, SpecialID: "inertial_stabilizer", TacticalSupport: "turning_mode"},
			},
			want: "special slots",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec := baselineTacticalBattleSpec()
			spec.Tactical.Ships[0].Specials = tc.specials
			_, err := NewSession(spec)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err=%v want substring %q", err, tc.want)
			}
		})
	}
}
