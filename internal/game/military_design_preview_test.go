package game

import (
	"testing"

	"moox/internal/core"
)

func TestSlice172MilitaryDesignPreviewUsesAuthoritativeBeamQuote(t *testing.T) {
	rules := loadColonyShipRules(t)
	state := core.NewSmallFixture(0x17210)
	if err := rules.ApplyReferenceTriangleTechnologyProfile(state, ReferenceTriangleProfileAllTech); err != nil {
		t.Fatal(err)
	}
	noShield := ""
	preview := rules.PreviewMilitaryDesign(&state.Empires[0], MilitaryDesignDraft{
		HullID: "battleship", StrategicPictureID: 24, ShieldID: &noShield,
		Weapons: []core.ShipWeaponMount{{
			Slot: 0, WeaponID: "phasor", Count: 3, MountType: core.ShipWeaponMountPointDefense,
			ModifierIDs: []core.ShipWeaponModifierID{
				core.ShipWeaponModifierAutofire,
				core.ShipWeaponModifierShieldPiercing,
			},
		}},
	})
	if !preview.SaveAvailable || preview.RejectionReason != "" || preview.Spec == nil {
		t.Fatalf("preview rejected: %+v", preview)
	}
	if len(preview.Weapons) != 1 {
		t.Fatalf("weapon rows=%d want=1", len(preview.Weapons))
	}
	row := preview.Weapons[0]
	if row.MountType != core.ShipWeaponMountPointDefense || row.MiniaturizationLevel != 4 || row.UnitSpace != 5 || row.UnitCostPP != 3 || row.RowSpace != 15 || row.RowCostPP != 9 {
		t.Fatalf("weapon preview=%+v", row)
	}
	if len(row.ModifierIDs) != 2 || row.ModifierIDs[0] != core.ShipWeaponModifierShieldPiercing || row.ModifierIDs[1] != core.ShipWeaponModifierAutofire {
		t.Fatalf("canonical modifiers=%v", row.ModifierIDs)
	}
	if preview.Spec.SpaceUsed != preview.SpaceUsed || preview.Spec.BaseDesignCostPP != preview.BaseDesignCostPP || preview.Spec.ProductionCostPP != preview.ProductionCostPP {
		t.Fatalf("projection/spec totals diverged: preview=%+v spec=%+v", preview, *preview.Spec)
	}
}

func TestSlice172MilitaryDesignPreviewReturnsStableWeaponRejection(t *testing.T) {
	rules := loadColonyShipRules(t)
	state := core.NewSmallFixture(0x17211)
	if err := rules.ApplyReferenceTriangleTechnologyProfile(state, ReferenceTriangleProfileAllTech); err != nil {
		t.Fatal(err)
	}
	preview := rules.PreviewMilitaryDesign(&state.Empires[0], MilitaryDesignDraft{
		HullID: "battleship", StrategicPictureID: 24,
		Weapons: []core.ShipWeaponMount{{
			Slot: 0, WeaponID: "mass_driver", Count: 1,
			ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierContinuous},
		}},
	})
	if preview.SaveAvailable || preview.RejectionReason != MilitaryDesignReasonWeaponModifierNotAllowed {
		t.Fatalf("preview rejection=%+v", preview)
	}
	if preview.HullSpace <= 0 || preview.RejectionMessage == "" {
		t.Fatalf("preview lacks usable rejection context: %+v", preview)
	}
}
