package game

import (
	"reflect"
	"strings"
	"testing"

	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/protocol"
)

func TestSlice174SpecialRulesCatalogAndFighterRows(t *testing.T) {
	rules := loadColonyShipRules(t)
	if rules.ShipSpecials == nil || len(rules.ShipSpecials.Specials) != 38 {
		t.Fatalf("special rules=%v count=%d", rules.ShipSpecials != nil, len(rules.ShipSpecials.Specials))
	}
	battlePods, ok := rules.ShipSpecials.ByID("battle_pods")
	if !ok || battlePods.OriginalID != 4 || battlePods.Miniaturizes || battlePods.SupportClass != 1 || battlePods.TacticalSupport != "strategic_only" {
		t.Fatalf("Battle Pods rule=%+v ok=%v", battlePods, ok)
	}
	if got := battlePods.BaseSpaceByHull; len(got) != 6 || got[5] != -600 {
		t.Fatalf("Battle Pods Space=%v", got)
	}
	if _, ok := rules.ShipSpecials.ByID("regeneration"); ok {
		t.Fatal("Regeneration must not be selectable")
	}

	state := core.NewSmallFixture(0x174001)
	applyMilitaryReferenceProfile(t, rules, state, ReferenceTriangleProfileAllTech)
	catalog, err := rules.MilitaryDesignerCatalog(&state.Empires[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Specials) != 38 {
		t.Fatalf("special catalog=%d want=38", len(catalog.Specials))
	}
	var battlePodsChoice *MilitaryDesignerSpecialChoice
	for i := range catalog.Specials {
		if catalog.Specials[i].ID == "battle_pods" {
			battlePodsChoice = &catalog.Specials[i]
			break
		}
	}
	if battlePodsChoice == nil || battlePodsChoice.Description == "" {
		t.Fatalf("Battle Pods catalog description missing: %+v", battlePodsChoice)
	}
	if !catalog.AutomaticSystems.MegafluxersKnown || catalog.AutomaticSystems.HullSpacePercent != 125 {
		t.Fatalf("Megafluxers projection=%+v", catalog.AutomaticSystems)
	}

	fighters := map[string]bool{
		"fighter_bays":       false,
		"bomber_bays":        false,
		"heavy_fighter_bays": false,
	}
	for _, weapon := range catalog.Weapons {
		if _, ok := fighters[weapon.ID]; !ok {
			continue
		}
		fighters[weapon.ID] = true
		if weapon.Kind != "fighter" || weapon.CraftPerUnit != 4 || len(weapon.ModifierChoices) != 0 || len(weapon.AmmoChoices) != 0 || len(weapon.FiringArcChoices) != 0 || weapon.FixedFiringArcID != "" {
			t.Fatalf("fighter catalog %q=%+v", weapon.ID, weapon)
		}
	}
	for id, found := range fighters {
		if !found {
			t.Fatalf("fighter catalog missing %q", id)
		}
	}
}

func TestSlice174CapacityRangeAndSpecialLegality(t *testing.T) {
	rules := loadColonyShipRules(t)
	state := core.NewSmallFixture(0x174002)
	applyMilitaryReferenceProfile(t, rules, state, ReferenceTriangleProfileAllTech)
	empire := &state.Empires[0]

	specials := []core.ShipSpecialMount{
		{Slot: 0, SpecialID: "battle_pods"},
		{Slot: 1, SpecialID: "extended_fuel_tanks"},
	}
	spec, err := rules.compileMilitaryDesignSpec(empire, "doom_star", hullPictureForTest(t, rules, "doom_star"), militaryDesignSelections{}, nil, specials)
	if err != nil {
		t.Fatal(err)
	}
	if spec.HullSpace != 1500 {
		t.Fatalf("Doom Star HullSpace=%d want=1500", spec.HullSpace)
	}
	if spec.FuelRangeParsecs != 383 {
		t.Fatalf("Extended Fuel Tanks range=%d want=383", spec.FuelRangeParsecs)
	}
	if len(spec.Specials) != 2 || spec.Specials[0].SpecialID != "battle_pods" || spec.Specials[1].SpecialID != "extended_fuel_tanks" {
		t.Fatalf("special snapshot=%+v", spec.Specials)
	}

	_, quotes, specialSpace, _, err := rules.quoteShipSpecials(empire, 5, spec.ShieldID, []core.ShipSpecialMount{{Slot: 0, SpecialID: "battle_pods"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(quotes) != 1 || quotes[0].MiniaturizationLevel != 0 || quotes[0].UnitSpace != -600 || quotes[0].UnitCostPP != 2500 || specialSpace != -600 {
		t.Fatalf("Battle Pods quote=%+v totalSpace=%d", quotes, specialSpace)
	}

	for _, tc := range []struct {
		name     string
		shieldID string
		specials []core.ShipSpecialMount
		reason   string
	}{
		{
			name:     "inertial exclusion",
			shieldID: spec.ShieldID,
			specials: []core.ShipSpecialMount{{Slot: 0, SpecialID: "inertial_stabilizer"}, {Slot: 1, SpecialID: "inertial_nullifier"}},
			reason:   MilitaryDesignReasonSpecialExclusionConflict,
		},
		{
			name:     "shield prerequisite",
			shieldID: "",
			specials: []core.ShipSpecialMount{{Slot: 0, SpecialID: "hard_shields"}},
			reason:   MilitaryDesignReasonSpecialRequiresShield,
		},
		{
			name:     "regeneration locked",
			shieldID: spec.ShieldID,
			specials: []core.ShipSpecialMount{{Slot: 0, SpecialID: "regeneration"}},
			reason:   MilitaryDesignReasonSpecialNotSupported,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, _, err := rules.quoteShipSpecials(empire, 5, tc.shieldID, tc.specials)
			if err == nil || militaryDesignReasonFromError(err) != tc.reason {
				t.Fatalf("err=%v reason=%q want=%q", err, militaryDesignReasonFromError(err), tc.reason)
			}
		})
	}
}

func TestSlice174FighterQuoteUsesBayCountAndNoOrdnanceControls(t *testing.T) {
	rules := loadColonyShipRules(t)
	state := core.NewSmallFixture(0x174003)
	applyMilitaryReferenceProfile(t, rules, state, ReferenceTriangleProfileAllTech)
	empire := &state.Empires[0]

	quote, err := rules.quoteWeapon(empire, core.ShipWeaponMount{
		Slot: 0, WeaponID: "fighter_bays", Count: 2, MountType: core.ShipWeaponMountStandard,
	})
	if err != nil {
		t.Fatal(err)
	}
	if quote.Weapon.Kind != "fighter" || quote.Weapon.CraftPerUnit != 4 || quote.RowSpace != quote.UnitSpace*2 || quote.RowCostPP != quote.UnitCostPP*2 {
		t.Fatalf("fighter quote=%+v", quote)
	}

	_, err = rules.quoteWeapon(empire, core.ShipWeaponMount{
		Slot: 0, WeaponID: "fighter_bays", Count: 1, MountType: core.ShipWeaponMountStandard, FiringArcID: core.ShipWeaponFiringArc360,
	})
	if err == nil || militaryDesignReasonFromError(err) != MilitaryDesignReasonWeaponFiringArcNotAllowed {
		t.Fatalf("fighter firing arc err=%v", err)
	}
}

func TestSlice174TacticalSpecialMetadataAndInertialTurning(t *testing.T) {
	rules := loadColonyShipRules(t)
	ship := tacticalMetadataShip(41, 1, "nuclear_drive", true)
	ship.Spec.Specials = []core.ShipSpecialMount{
		{Slot: 0, SpecialID: "battle_pods"},
		{Slot: 1, SpecialID: "inertial_stabilizer"},
		{Slot: 2, SpecialID: "scout_lab"},
	}
	got := baselineTacticalShip(ship, protocol.SeatID(1), 10, 12, 0, 10, rules.TacticalCombat, rules.ShipSpecials)
	if got.TurningMode != battle.TacticalTurningInertialStabilizer {
		t.Fatalf("turning=%q want=%q", got.TurningMode, battle.TacticalTurningInertialStabilizer)
	}
	if len(got.Specials) != 3 {
		t.Fatalf("tactical specials=%+v", got.Specials)
	}
	if got.Specials[0].TacticalSupport != "strategic_only" || got.Specials[1].TacticalSupport != "turning_mode" || got.Specials[2].TacticalSupport != "metadata_only" {
		t.Fatalf("support mapping=%+v", got.Specials)
	}

	ship.Spec.Specials = []core.ShipSpecialMount{{Slot: 0, SpecialID: "inertial_nullifier"}}
	got = baselineTacticalShip(ship, protocol.SeatID(1), 10, 12, 0, 10, rules.TacticalCombat, rules.ShipSpecials)
	if got.TurningMode != battle.TacticalTurningInertialNullifier {
		t.Fatalf("nullifier turning=%q", got.TurningMode)
	}
}

func TestSlice174SavePayloadCanonicalSpecials(t *testing.T) {
	payload := SaveMilitaryDesignPayload{
		Name: "Special Test", HullID: "frigate", StrategicPictureID: 0,
		Specials: []core.ShipSpecialMount{
			{Slot: 3, SpecialID: "scout_lab"},
			{Slot: 1, SpecialID: "battle_pods"},
		},
	}
	if err := validateSaveMilitaryDesignPayload(payload); err != nil {
		t.Fatalf("valid Special payload: %v", err)
	}

	payload.Specials = append(payload.Specials, core.ShipSpecialMount{Slot: 4, SpecialID: "battle_pods"})
	if err := validateSaveMilitaryDesignPayload(payload); err == nil || !strings.Contains(err.Error(), MilitaryDesignReasonSpecialDuplicate) {
		t.Fatalf("duplicate Special err=%v", err)
	}
}

func TestSlice174ConstructionPreservesSpecialAndFighterSnapshot(t *testing.T) {
	rules := loadColonyShipRules(t)
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	state := core.NewSmallFixture(0x174004)
	applyMilitaryReferenceProfile(t, rules, state, ReferenceTriangleProfileAllTech)
	empire := &state.Empires[0]
	colony := &state.Colonies[0]

	save, err := NewSaveMilitaryDesignCommand(1, SaveMilitaryDesignPayload{
		Name:               "Special Carrier",
		HullID:             "frigate",
		StrategicPictureID: hullPictureForTest(t, rules, "frigate"),
		Specials: []core.ShipSpecialMount{
			{Slot: 0, SpecialID: "battle_pods"},
			{Slot: 1, SpecialID: "inertial_stabilizer"},
		},
		Weapons: []core.ShipWeaponMount{{
			Slot: 0, WeaponID: "fighter_bays", Count: 1, MountType: core.ShipWeaponMountStandard,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.saveMilitaryDesign(state, empire.ID, 1, save); err != nil {
		t.Fatal(err)
	}
	design := state.ShipDesigns[len(state.ShipDesigns)-1]
	if len(design.Spec.Specials) != 2 || len(design.Spec.Weapons) != 1 {
		t.Fatalf("saved snapshot=%+v", design.Spec)
	}

	queue, err := NewQueueMilitaryShipCommand(2, QueueMilitaryShipPayload{ColonyID: colony.ID, ShipDesignID: design.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.queueMilitaryShip(state, empire.ID, 1, queue); err != nil {
		t.Fatal(err)
	}
	if colony.Construction == nil || colony.Construction.MilitaryShipSpec == nil {
		t.Fatalf("queued construction=%+v", colony.Construction)
	}
	queued := cloneShipDesignSpec(*colony.Construction.MilitaryShipSpec)
	if !reflect.DeepEqual(queued.Specials, design.Spec.Specials) || !reflect.DeepEqual(queued.Weapons, design.Spec.Weapons) {
		t.Fatalf("queued equipment specials=%+v weapons=%+v want specials=%+v weapons=%+v", queued.Specials, queued.Weapons, design.Spec.Specials, design.Spec.Weapons)
	}

	cost, err := resolver.constructionProjectCostPP(state, colony, colony.Construction)
	if err != nil {
		t.Fatal(err)
	}
	colony.PopulationDynamics.ProductionAvailable = float64(cost)
	if _, err := resolver.advanceConstruction(state); err != nil {
		t.Fatal(err)
	}
	if len(state.Ships) != 1 {
		t.Fatalf("built ships=%d want=1", len(state.Ships))
	}
	built := state.Ships[0]
	if !reflect.DeepEqual(built.Spec.Specials, queued.Specials) || !reflect.DeepEqual(built.Spec.Weapons, queued.Weapons) {
		t.Fatalf("built equipment specials=%+v weapons=%+v want specials=%+v weapons=%+v", built.Spec.Specials, built.Spec.Weapons, queued.Specials, queued.Weapons)
	}
}
