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
	if len(catalog.Weapons) != 13 {
		t.Fatalf("weapon choices=%d want 13", len(catalog.Weapons))
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
	if laser.Key != "frigate:laser_cannon" || laser.Spec.ProductionCostPP != 30 || laser.Spec.SpaceUsed != 10 || laser.CommandPointCost != 1 || !reflect.DeepEqual(laser.Weapons, []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1, MountType: core.ShipWeaponMountStandard}}) {
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
	if len(catalog.Weapons) != 13 {
		t.Fatalf("locked Beam catalog size=%d want 13", len(catalog.Weapons))
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
