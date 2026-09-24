package game

import (
	"strings"
	"testing"

	"moox/internal/core"
)

func TestSlice173OrdnanceQuoteCanonicalAmmoArcAndBaseArithmetic(t *testing.T) {
	rules := loadColonyShipRules(t)
	state := core.NewSmallFixture(0x17301)
	empire := &state.Empires[0]
	addBaselineMilitaryTechnologies(empire)
	addColonyShipTestTechnology(empire, 121, 12, 119)

	cases := []struct {
		name      string
		mount     core.ShipWeaponMount
		wantAmmo  int
		wantSpace int
		wantCost  int
	}{
		{
			name:      "missile rack five",
			mount:     core.ShipWeaponMount{Slot: 0, WeaponID: "nuclear_missile", Count: 1, Ammo: 5},
			wantAmmo:  5,
			wantSpace: 20,
			wantCost:  20,
		},
		{
			name:      "torpedo fixed ammo",
			mount:     core.ShipWeaponMount{Slot: 0, WeaponID: "anti_matter_torpedoes", Count: 1},
			wantAmmo:  2,
			wantSpace: 20,
			wantCost:  15,
		},
		{
			name:      "bomb fixed ammo",
			mount:     core.ShipWeaponMount{Slot: 0, WeaponID: "nuclear_bomb", Count: 1},
			wantAmmo:  10,
			wantSpace: 5,
			wantCost:  3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			quote, err := rules.quoteWeapon(empire, tc.mount)
			if err != nil {
				t.Fatal(err)
			}
			if quote.Mount.Ammo != tc.wantAmmo {
				t.Fatalf("ammo=%d want=%d", quote.Mount.Ammo, tc.wantAmmo)
			}
			if quote.Mount.FiringArcID != core.ShipWeaponFiringArc360 {
				t.Fatalf("arc=%q want=%q", quote.Mount.FiringArcID, core.ShipWeaponFiringArc360)
			}
			if quote.Mount.MountType != core.ShipWeaponMountStandard {
				t.Fatalf("mount=%q want standard", quote.Mount.MountType)
			}
			if quote.UnitSpace != tc.wantSpace || quote.UnitCostPP != tc.wantCost {
				t.Fatalf("quote=%+v want unit space/cost=%d/%d", quote, tc.wantSpace, tc.wantCost)
			}
		})
	}
}

func TestSlice173OrdnanceQuoteRejectsFrozenInvalidConfigurations(t *testing.T) {
	rules := loadColonyShipRules(t)
	state := core.NewSmallFixture(0x17302)
	empire := &state.Empires[0]
	addBaselineMilitaryTechnologies(empire)
	addColonyShipTestTechnology(empire, 121, 12, 119)

	cases := []struct {
		name   string
		mount  core.ShipWeaponMount
		reason string
	}{
		{
			name:   "missile ammo required",
			mount:  core.ShipWeaponMount{Slot: 0, WeaponID: "nuclear_missile", Count: 1},
			reason: MilitaryDesignReasonWeaponAmmoRequired,
		},
		{
			name:   "missile ammo invalid",
			mount:  core.ShipWeaponMount{Slot: 0, WeaponID: "nuclear_missile", Count: 1, Ammo: 3},
			reason: MilitaryDesignReasonWeaponAmmoInvalid,
		},
		{
			name:   "torpedo ammo fixed",
			mount:  core.ShipWeaponMount{Slot: 0, WeaponID: "anti_matter_torpedoes", Count: 1, Ammo: 5},
			reason: MilitaryDesignReasonWeaponAmmoNotAllowed,
		},
		{
			name:   "ordnance arc fixed",
			mount:  core.ShipWeaponMount{Slot: 0, WeaponID: "nuclear_bomb", Count: 1, FiringArcID: core.ShipWeaponFiringArcID("forward")},
			reason: MilitaryDesignReasonWeaponFiringArcNotAllowed,
		},
		{
			name:   "ordnance mount standard only",
			mount:  core.ShipWeaponMount{Slot: 0, WeaponID: "anti_matter_torpedoes", Count: 1, MountType: core.ShipWeaponMountHeavy},
			reason: MilitaryDesignReasonWeaponMountNotAllowed,
		},
		{
			name: "missile mirv requires refinement",
			mount: core.ShipWeaponMount{
				Slot: 0, WeaponID: "nuclear_missile", Count: 1, Ammo: 5,
				ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierMIRV},
			},
			reason: MilitaryDesignReasonWeaponModifierRefinementRequired,
		},
		{
			name: "bomb modifiers forbidden",
			mount: core.ShipWeaponMount{
				Slot: 0, WeaponID: "nuclear_bomb", Count: 1,
				ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierECCM},
			},
			reason: MilitaryDesignReasonWeaponModifierNotAllowed,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := rules.quoteWeapon(empire, tc.mount)
			if err == nil || !strings.Contains(err.Error(), tc.reason) {
				t.Fatalf("error=%v want reason %q", err, tc.reason)
			}
		})
	}
}

func TestSlice173MissileEmissionsGuidanceArithmeticAtM0(t *testing.T) {
	rules := loadColonyShipRules(t)
	state := core.NewSmallFixture(0x17303)
	empire := &state.Empires[0]
	addBaselineMilitaryTechnologies(empire)
	addColonyShipTestTechnology(empire, 121)

	quote, err := rules.quoteWeapon(empire, core.ShipWeaponMount{
		Slot: 0, WeaponID: "nuclear_missile", Count: 2, Ammo: 5,
		ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierEmissionsGuidance},
	})
	if err != nil {
		t.Fatal(err)
	}
	if quote.MiniaturizationLevel != 0 {
		t.Fatalf("mini level=%d want=0", quote.MiniaturizationLevel)
	}
	if quote.UnitSpace != 80 || quote.UnitCostPP != 80 || quote.RowSpace != 160 || quote.RowCostPP != 160 {
		t.Fatalf("emissions-guidance quote=%+v", quote)
	}
}

func TestSlice173WeaponValidationReasonOrdering(t *testing.T) {
	rules := loadColonyShipRules(t)
	state := core.NewSmallFixture(0x17304)
	empire := &state.Empires[0]
	addBaselineMilitaryTechnologies(empire)

	assertReason := func(name string, mount core.ShipWeaponMount, want string) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			_, err := rules.quoteWeapon(empire, mount)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("error=%v want first reason %q", err, want)
			}
		})
	}

	assertReason("generic mount syntax before identity", core.ShipWeaponMount{
		Slot: 0, WeaponID: "unknown_weapon", Count: 1, MountType: core.ShipWeaponMountType("invalid"),
	}, MilitaryDesignReasonWeaponMountNotAllowed)
	assertReason("count before identity", core.ShipWeaponMount{
		Slot: 0, WeaponID: "unknown_weapon", Count: 0,
	}, MilitaryDesignReasonWeaponCountInvalid)
	assertReason("identity before modifier", core.ShipWeaponMount{
		Slot: 0, WeaponID: "unknown_weapon", Count: 1,
		ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierID("bogus")},
	}, MilitaryDesignReasonWeaponUnknown)
	assertReason("technology before family mount ammo arc modifier", core.ShipWeaponMount{
		Slot: 0, WeaponID: "nuclear_missile", Count: 1, MountType: core.ShipWeaponMountHeavy,
		FiringArcID: core.ShipWeaponFiringArcID("forward"),
		ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierID("bogus")},
	}, MilitaryDesignReasonWeaponTechnologyRequired)

	addColonyShipTestTechnology(empire, 121)

	assertReason("family mount before ammo", core.ShipWeaponMount{
		Slot: 0, WeaponID: "nuclear_missile", Count: 1, MountType: core.ShipWeaponMountHeavy,
		FiringArcID: core.ShipWeaponFiringArcID("forward"),
		ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierID("bogus")},
	}, MilitaryDesignReasonWeaponMountNotAllowed)
	assertReason("ammo before arc and modifier", core.ShipWeaponMount{
		Slot: 0, WeaponID: "nuclear_missile", Count: 1,
		FiringArcID: core.ShipWeaponFiringArcID("forward"),
		ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierID("bogus")},
	}, MilitaryDesignReasonWeaponAmmoRequired)
	assertReason("arc before modifier", core.ShipWeaponMount{
		Slot: 0, WeaponID: "nuclear_missile", Count: 1, Ammo: 5,
		FiringArcID: core.ShipWeaponFiringArcID("forward"),
		ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierID("bogus")},
	}, MilitaryDesignReasonWeaponFiringArcNotAllowed)
	assertReason("modifier after arc", core.ShipWeaponMount{
		Slot: 0, WeaponID: "nuclear_missile", Count: 1, Ammo: 5,
		FiringArcID: core.ShipWeaponFiringArc360,
		ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierID("bogus")},
	}, MilitaryDesignReasonWeaponModifierUnknown)

	_, err := canonicalWeaponMounts([]core.ShipWeaponMount{
		{Slot: 0, WeaponID: "unknown_a", Count: 1},
		{Slot: 0, WeaponID: "unknown_b", Count: 1},
	})
	if err == nil || !strings.Contains(err.Error(), MilitaryDesignReasonWeaponSlotDuplicate) {
		t.Fatalf("duplicate-slot generic validation error=%v", err)
	}
}
