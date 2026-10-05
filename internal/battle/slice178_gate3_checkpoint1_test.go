package battle

import "testing"

func TestSlice178Gate3Checkpoint1RequiresDoctrineForFighterCarrier(t *testing.T) {
	spec := TacticalSpec{Ships: []TacticalShipSpec{{
		ShipID: 1, EmpireID: 10, SeatID: 1,
		Weapons: []TacticalWeaponSpec{{Slot: 0, WeaponID: "fighter_bays", WeaponFamily: "fighter", Count: 1, TacticalSupport: "fighter", CraftPerUnit: 4}},
	}}}
	if err := validateTacticalFighterDoctrines(spec); err == nil {
		t.Fatal("Fighter carrier without doctrine unexpectedly validated")
	}
	spec.FighterDoctrines = []TacticalFighterDoctrineSpec{{
		SeatID: 1, EmpireID: 10, BestWarpDriveIndex: 1, ArmorDurabilityMultiple: 1,
	}}
	if err := validateTacticalFighterDoctrines(spec); err != nil {
		t.Fatalf("valid Fighter doctrine rejected: %v", err)
	}
}

func TestSlice178Gate3Checkpoint1CloneDetachesNewSnapshotSlices(t *testing.T) {
	spec := TacticalSpec{
		Rules: TacticalRulesSnapshot{
			Fighter:  &TacticalFighterRulesSnapshot{ArmorDurabilityMultiples: []int{1, 2, 4, 6, 8, 10}},
			Boarding: &TacticalBoardingRulesSnapshot{CrewQualityBonuses: []int{0, 5, 10, 15}},
			Specials: &TacticalSpecialRulesSnapshot{SelfTeleportMaxRange: 20},
		},
		FighterDoctrines: []TacticalFighterDoctrineSpec{{
			SeatID: 1, EmpireID: 10, BestWarpDriveIndex: 6, ArmorDurabilityMultiple: 10,
			BeamPayload: &TacticalFighterPayloadSpec{WeaponID: "phasor", MinDamage: 5, MaxDamage: 20, InnateEffectIDs: []string{"shield_piercing"}},
		}},
		Ships: []TacticalShipSpec{{
			ShipID: 1,
			Internal: &TacticalInternalSpec{
				Weapons:  []TacticalInternalWeaponSpec{{Slot: 0, Max: 4, UnitSize: 1}},
				Specials: []TacticalInternalSpecialSpec{{Slot: 0, Max: 2}},
			},
		}},
	}
	cloned := CloneTacticalSpec(spec)
	cloned.Rules.Fighter.ArmorDurabilityMultiples[0] = 99
	cloned.Rules.Boarding.CrewQualityBonuses[0] = 99
	cloned.FighterDoctrines[0].BeamPayload.InnateEffectIDs[0] = "changed"
	cloned.Ships[0].Internal.Weapons[0].Max = 99
	cloned.Ships[0].Internal.Specials[0].Max = 99

	if spec.Rules.Fighter.ArmorDurabilityMultiples[0] != 1 ||
		spec.Rules.Boarding.CrewQualityBonuses[0] != 0 ||
		spec.FighterDoctrines[0].BeamPayload.InnateEffectIDs[0] != "shield_piercing" ||
		spec.Ships[0].Internal.Weapons[0].Max != 4 ||
		spec.Ships[0].Internal.Specials[0].Max != 2 {
		t.Fatalf("CloneTacticalSpec aliases Slice-17.8 snapshot state: %+v", spec)
	}
}
