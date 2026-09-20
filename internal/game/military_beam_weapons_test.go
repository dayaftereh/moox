package game

import (
	"strings"
	"testing"

	"moox/internal/core"
)

func TestSlice172BeamQuoteMountsAndRefinement(t *testing.T) {
	rules := loadColonyShipRules(t)
	state := core.NewSmallFixture(0x17201)
	empire := &state.Empires[0]
	addBaselineMilitaryTechnologies(empire)
	addColonyShipTestTechnology(empire, 100)

	for _, tc := range []struct {
		name      string
		mountType core.ShipWeaponMountType
		wantSpace int
		wantCost  int
	}{
		{name: "standard", mountType: core.ShipWeaponMountStandard, wantSpace: 10, wantCost: 5},
		{name: "heavy", mountType: core.ShipWeaponMountHeavy, wantSpace: 20, wantCost: 10},
		{name: "point defense", mountType: core.ShipWeaponMountPointDefense, wantSpace: 5, wantCost: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			quote, err := rules.quoteBeamWeapon(empire, core.ShipWeaponMount{
				Slot: 0, WeaponID: "laser_cannon", Count: 1, MountType: tc.mountType,
			})
			if err != nil {
				t.Fatal(err)
			}
			if quote.UnitSpace != tc.wantSpace || quote.UnitCostPP != tc.wantCost || quote.RowSpace != tc.wantSpace || quote.RowCostPP != tc.wantCost {
				t.Fatalf("quote=%+v want space/cost=%d/%d", quote, tc.wantSpace, tc.wantCost)
			}
		})
	}

	_, err := rules.quoteBeamWeapon(empire, core.ShipWeaponMount{
		Slot: 0, WeaponID: "laser_cannon", Count: 1,
		ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierArmorPiercing},
	})
	if err == nil || !strings.Contains(err.Error(), MilitaryDesignReasonWeaponModifierRefinementRequired) {
		t.Fatalf("M0 armor piercing error=%v", err)
	}
}

func TestSlice172AntaranField74MiniaturizationIsTerminal(t *testing.T) {
	rules := loadColonyShipRules(t)
	state := core.NewSmallFixture(0x17204)
	if err := rules.ApplyReferenceTriangleTechnologyProfile(state, ReferenceTriangleProfileAllTech); err != nil {
		t.Fatal(err)
	}
	empire := &state.Empires[0]
	for _, technologyID := range []int{123, 47} {
		level, err := rules.shipComponentMiniaturizationLevel(empire, technologyID)
		if err != nil {
			t.Fatalf("technology %d miniaturization: %v", technologyID, err)
		}
		if level != 0 {
			t.Fatalf("special field-74 technology %d miniaturization=%d want=0", technologyID, level)
		}
	}
	particle, err := rules.quoteBeamWeapon(empire, core.ShipWeaponMount{
		Slot: 0, WeaponID: "particle_beam", Count: 1, MountType: core.ShipWeaponMountStandard,
	})
	if err != nil {
		t.Fatal(err)
	}
	if particle.MiniaturizationLevel != 0 || particle.UnitSpace != 15 || particle.UnitCostPP != 35 {
		t.Fatalf("Particle Beam field-74 quote=%+v", particle)
	}
}
func TestSlice172BeamQuoteAllTechModifierArithmeticAndCanonicalization(t *testing.T) {
	rules := loadColonyShipRules(t)
	state := core.NewSmallFixture(0x17202)
	if err := rules.ApplyReferenceTriangleTechnologyProfile(state, ReferenceTriangleProfileAllTech); err != nil {
		t.Fatal(err)
	}
	empire := &state.Empires[0]

	quote, err := rules.quoteBeamWeapon(empire, core.ShipWeaponMount{
		Slot: 3, WeaponID: "phasor", Count: 3, MountType: core.ShipWeaponMountPointDefense,
		ModifierIDs: []core.ShipWeaponModifierID{
			core.ShipWeaponModifierAutofire,
			core.ShipWeaponModifierShieldPiercing,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if quote.Mount.MountType != core.ShipWeaponMountPointDefense {
		t.Fatalf("mount type=%q", quote.Mount.MountType)
	}
	wantModifiers := []core.ShipWeaponModifierID{
		core.ShipWeaponModifierShieldPiercing,
		core.ShipWeaponModifierAutofire,
	}
	if len(quote.Mount.ModifierIDs) != len(wantModifiers) || quote.Mount.ModifierIDs[0] != wantModifiers[0] || quote.Mount.ModifierIDs[1] != wantModifiers[1] {
		t.Fatalf("canonical modifiers=%v want=%v", quote.Mount.ModifierIDs, wantModifiers)
	}
	if quote.ModifierMask != 0xC4 {
		t.Fatalf("modifier mask=%#x want=0xC4", quote.ModifierMask)
	}
	// Phasor base 10/10; PD -> 5/5; Shield Piercing + Autofire => +100% -> 10/10;
	// All-Tech gives Phasor mini level 4: Beam Space 50%% and cost 30%% -> unit 5/3, then count 3.
	if quote.MiniaturizationLevel != 4 || quote.UnitSpace != 5 || quote.UnitCostPP != 3 || quote.RowSpace != 15 || quote.RowCostPP != 9 {
		t.Fatalf("all-tech phasor quote=%+v", quote)
	}
}

func TestSlice172BeamQuoteRejectsExecutableIncompatibleModifier(t *testing.T) {
	rules := loadColonyShipRules(t)
	state := core.NewSmallFixture(0x17203)
	if err := rules.ApplyReferenceTriangleTechnologyProfile(state, ReferenceTriangleProfileAllTech); err != nil {
		t.Fatal(err)
	}
	_, err := rules.quoteBeamWeapon(&state.Empires[0], core.ShipWeaponMount{
		Slot: 0, WeaponID: "mass_driver", Count: 1,
		ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierContinuous},
	})
	if err == nil || !strings.Contains(err.Error(), MilitaryDesignReasonWeaponModifierNotAllowed) {
		t.Fatalf("Mass Driver Continuous error=%v", err)
	}
}

func TestSlice172CanonicalWeaponMountsSortSlotsAndRejectDuplicate(t *testing.T) {
	got, err := canonicalWeaponMounts([]core.ShipWeaponMount{
		{Slot: 5, WeaponID: "laser_cannon", Count: 1},
		{Slot: 1, WeaponID: "mass_driver", Count: 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Slot != 1 || got[1].Slot != 5 || got[0].MountType != core.ShipWeaponMountStandard || got[1].MountType != core.ShipWeaponMountStandard {
		t.Fatalf("canonical mounts=%+v", got)
	}
	_, err = canonicalWeaponMounts([]core.ShipWeaponMount{
		{Slot: 0, WeaponID: "laser_cannon", Count: 1},
		{Slot: 0, WeaponID: "mass_driver", Count: 1},
	})
	if err == nil || !strings.Contains(err.Error(), MilitaryDesignReasonWeaponSlotDuplicate) {
		t.Fatalf("duplicate slot error=%v", err)
	}
}
