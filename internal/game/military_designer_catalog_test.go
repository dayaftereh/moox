package game

import (
	"reflect"
	"testing"

	"moox/internal/core"
)

func designerWeaponChoiceByID(items []MilitaryDesignerWeaponChoice, id string) (MilitaryDesignerWeaponChoice, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return MilitaryDesignerWeaponChoice{}, false
}
func TestMilitaryDesignerCatalogProjectsOrderedAuthoritativeBaseline(t *testing.T) {
	rules := loadColonyShipRules(t)
	empire := core.NewSmallFixture(1815).Empires[0]
	addBaselineMilitaryTechnologies(&empire)
	addColonyShipTestTechnology(&empire, 100)

	catalog, err := rules.MilitaryDesignerCatalog(&empire)
	if err != nil {
		t.Fatal(err)
	}
	wantHulls := []string{"frigate", "destroyer", "cruiser", "battleship", "titan", "doom_star"}
	if len(catalog.Hulls) != len(wantHulls) {
		t.Fatalf("hulls=%d want=%d: %+v", len(catalog.Hulls), len(wantHulls), catalog.Hulls)
	}
	for i, wantID := range wantHulls {
		hull := catalog.Hulls[i]
		if hull.ID != wantID || hull.SizeIndex != i || hull.CommandPointCost != i+1 {
			t.Fatalf("hull[%d]=%+v want id=%s size=%d cp=%d", i, hull, wantID, i, i+1)
		}
	}
	if !catalog.Hulls[0].SaveAvailable || catalog.Hulls[0].LockReason != "" {
		t.Fatalf("Frigate availability=%+v", catalog.Hulls[0])
	}
	for i := 0; i <= 3; i++ {
		if !catalog.Hulls[i].SaveAvailable || catalog.Hulls[i].LockReason != "" {
			t.Fatalf("base hull availability[%d]=%+v", i, catalog.Hulls[i])
		}
	}
	if catalog.Hulls[4].RequiredTechnologyKey != "titan_construction" || catalog.Hulls[4].RequiredTechnologyID != 186 || catalog.Hulls[4].TechnologyKnown || catalog.Hulls[4].LockReason != ShipDesignerLockTechnologyRequired {
		t.Fatalf("Titan lock=%+v", catalog.Hulls[4])
	}
	if catalog.Hulls[5].RequiredTechnologyKey != "doom_star_construction" || catalog.Hulls[5].RequiredTechnologyID != 55 || catalog.Hulls[5].TechnologyKnown || catalog.Hulls[5].LockReason != ShipDesignerLockTechnologyRequired {
		t.Fatalf("Doom Star lock=%+v", catalog.Hulls[5])
	}
	if len(catalog.Weapons) != 29 {
		t.Fatalf("weapon choices=%d want 29", len(catalog.Weapons))
	}
	laserChoice, ok := designerWeaponChoiceByID(catalog.Weapons, "laser_cannon")
	if !ok || laserChoice.TechnologyID != 100 || !laserChoice.Available {
		t.Fatalf("Laser choice=%+v ok=%v", laserChoice, ok)
	}
	if laserChoice.MinDamage != 1 || laserChoice.MaxDamage != 4 || laserChoice.EffectiveStandardSpace != 10 || laserChoice.EffectiveStandardCostPP != 5 {
		t.Fatalf("Laser choice=%+v", laserChoice)
	}
	if len(laserChoice.MountChoices) != 3 || len(laserChoice.ModifierChoices) != 4 {
		t.Fatalf("Laser mount/modifier choices=%+v / %+v", laserChoice.MountChoices, laserChoice.ModifierChoices)
	}
	if laserChoice.DefaultFiringArcID != core.ShipWeaponFiringArcForward || len(laserChoice.FiringArcChoices) != 5 {
		t.Fatalf("Laser firing arcs=%+v default=%q", laserChoice.FiringArcChoices, laserChoice.DefaultFiringArcID)
	}
	if laserChoice.FiringArcChoices[1].ID != core.ShipWeaponFiringArcForwardExtended || laserChoice.FiringArcChoices[1].SpaceCostPercent != 25 ||
		laserChoice.FiringArcChoices[4].ID != core.ShipWeaponFiringArc360 || laserChoice.FiringArcChoices[4].ProductionCostPercent != 50 {
		t.Fatalf("Laser firing arc metadata=%+v", laserChoice.FiringArcChoices)
	}
	missileChoice, ok := designerWeaponChoiceByID(catalog.Weapons, "nuclear_missile")
	if !ok || missileChoice.Kind != "missile" || missileChoice.Available || missileChoice.DefaultAmmo != 5 || missileChoice.FixedFiringArcID != core.ShipWeaponFiringArc360 {
		t.Fatalf("Nuclear Missile catalog choice=%+v ok=%v", missileChoice, ok)
	}
	if !reflect.DeepEqual(missileChoice.AmmoChoices, []int{2, 5, 10, 15, 20}) || len(missileChoice.MountChoices) != 1 || len(missileChoice.ModifierChoices) != 5 {
		t.Fatalf("Nuclear Missile config choices=%+v mounts=%+v modifiers=%+v", missileChoice.AmmoChoices, missileChoice.MountChoices, missileChoice.ModifierChoices)
	}
	if missileChoice.EffectiveStandardSpace != 20 || missileChoice.EffectiveStandardCostPP != 20 || missileChoice.TacticalSupport != "metadata_only" {
		t.Fatalf("Nuclear Missile default quote metadata=%+v", missileChoice)
	}
	torpedoChoice, ok := designerWeaponChoiceByID(catalog.Weapons, "anti_matter_torpedoes")
	if !ok || torpedoChoice.DefaultAmmo != 2 || torpedoChoice.FixedFiringArcID != core.ShipWeaponFiringArc360 || len(torpedoChoice.AmmoChoices) != 0 {
		t.Fatalf("Anti-Matter Torpedo catalog choice=%+v ok=%v", torpedoChoice, ok)
	}
	bombChoice, ok := designerWeaponChoiceByID(catalog.Weapons, "nuclear_bomb")
	if !ok || bombChoice.DefaultAmmo != 10 || bombChoice.FixedFiringArcID != core.ShipWeaponFiringArc360 || len(bombChoice.ModifierChoices) != 0 {
		t.Fatalf("Nuclear Bomb catalog choice=%+v ok=%v", bombChoice, ok)
	}
	if catalog.ProductionCostNumerator != 1 || catalog.ProductionCostDenominator != 1 {
		t.Fatalf("production ratio=%d/%d want 1/1", catalog.ProductionCostNumerator, catalog.ProductionCostDenominator)
	}
	if len(catalog.Variants) != 8 {
		t.Fatalf("variants=%+v want unarmed + Laser for four base hulls", catalog.Variants)
	}
	if catalog.Variants[0].Key != "frigate:none" || catalog.Variants[0].Spec.ProductionCostPP != 25 || catalog.Variants[0].Spec.HullSpace != 25 || catalog.Variants[0].Spec.SpaceUsed != 0 || catalog.Variants[0].CommandPointCost != 1 {
		t.Fatalf("unarmed preview=%+v", catalog.Variants[0])
	}
	laser := catalog.Variants[1]
	if laser.Key != "frigate:laser_cannon" || laser.Spec.ProductionCostPP != 30 || laser.Spec.SpaceUsed != 10 || laser.CommandPointCost != 1 || !reflect.DeepEqual(laser.Weapons, []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1, MountType: core.ShipWeaponMountStandard, FiringArcID: core.ShipWeaponFiringArcForward}}) {
		t.Fatalf("Laser preview=%+v", laser)
	}
}

func TestMilitaryDesignerCatalogLocksUnknownLaserAndOmitsIllegalPreview(t *testing.T) {
	rules := loadColonyShipRules(t)
	empire := core.NewSmallFixture(1816).Empires[0]
	addBaselineMilitaryTechnologies(&empire)

	catalog, err := rules.MilitaryDesignerCatalog(&empire)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Weapons) != 29 {
		t.Fatalf("locked weapon catalog size=%d want 29", len(catalog.Weapons))
	}
	laserChoice, ok := designerWeaponChoiceByID(catalog.Weapons, "laser_cannon")
	if !ok || laserChoice.TechnologyKnown || laserChoice.Available || laserChoice.LockReason != ShipDesignerLockTechnologyRequired {
		t.Fatalf("locked Laser=%+v ok=%v", laserChoice, ok)
	}
	if len(catalog.Variants) != 4 || catalog.Variants[0].Key != "frigate:none" || catalog.Variants[3].Key != "battleship:none" {
		t.Fatalf("variants=%+v want four unarmed base-hull previews", catalog.Variants)
	}
}
func TestMilitaryDesignerCatalogProjectsFeudalProductionRatio(t *testing.T) {
	rules := loadColonyShipRules(t)
	empire := core.NewSmallFixture(1817).Empires[0]
	addBaselineMilitaryTechnologies(&empire)
	modifier := rules.RaceModifiers[empire.RaceID]
	modifier.GovernmentTraitID = "government_feudal"
	rules.RaceModifiers[empire.RaceID] = modifier

	catalog, err := rules.MilitaryDesignerCatalog(&empire)
	if err != nil {
		t.Fatal(err)
	}
	if catalog.ProductionCostNumerator != 2 || catalog.ProductionCostDenominator != 3 {
		t.Fatalf("production ratio=%d/%d want 2/3", catalog.ProductionCostNumerator, catalog.ProductionCostDenominator)
	}
	if catalog.Variants[0].Spec.BaseDesignCostPP != 25 || catalog.Variants[0].Spec.ProductionCostPP != 17 {
		t.Fatalf("Feudal base preview=%+v", catalog.Variants[0].Spec)
	}
}
