package game

import (
	"reflect"
	"testing"

	"moox/internal/core"
)

func TestReferenceTriangleCombatReviewGameCreatesImmediateFalconEncounter(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	generated, err := rules.NewReferenceTriangleCombatReviewGame(ReferenceTriangleSeed)
	if err != nil {
		t.Fatal(err)
	}
	state := generated.State
	if len(state.Galaxy.Systems) != 4 {
		t.Fatalf("systems=%d want 4", len(state.Galaxy.Systems))
	}
	var fightSystemID core.ID
	for _, system := range state.Galaxy.Systems {
		if system.Name == "Fight System" {
			fightSystemID = system.ID
			break
		}
	}
	if fightSystemID == 0 {
		t.Fatal("Fight System missing")
	}
	humanID := generated.Players[0].EmpireID
	aiID := generated.Players[1].EmpireID
	if !state.MayAttackEmpire(humanID, aiID) || !state.MayAttackEmpire(aiID, humanID) {
		t.Fatal("combat review empires are not mutually hostile")
	}
	wantSpec := referenceCombatReviewFalconSpec()
	for _, empireID := range []core.ID{humanID, aiID} {
		var reviewFleet *core.StrategicFleet
		for i := range state.StrategicFleets {
			fleet := &state.StrategicFleets[i]
			if fleet.EmpireID == empireID && fleet.Role == core.StrategicFleetRoleCombat && fleet.AtSystemID == fightSystemID {
				reviewFleet = fleet
				break
			}
		}
		if reviewFleet == nil || len(reviewFleet.ShipIDs) != 3 {
			t.Fatalf("empire %d review fleet=%+v want three ships", empireID, reviewFleet)
		}
		for _, shipID := range reviewFleet.ShipIDs {
			ship := shipByID(state, shipID)
			if ship == nil || !reflect.DeepEqual(ship.Spec, wantSpec) {
				t.Fatalf("empire %d ship %d does not match Falcon review spec: %+v", empireID, shipID, ship)
			}
		}
	}
	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	_, encounters, err := resolver.prepareEncounterBoundary(encounterTestContext(state), state)
	if err != nil {
		t.Fatal(err)
	}
	if len(encounters) != 1 {
		t.Fatalf("encounters=%d want 1: %+v", len(encounters), encounters)
	}
	encounter := encounters[0]
	if encounter.SystemID != fightSystemID || len(encounter.Attacker.ShipIDs) != 3 || len(encounter.Defender.ShipIDs) != 3 {
		t.Fatalf("encounter=%+v want 3v3 at Fight System", encounter)
	}
	if encounter.Tactical == nil || encounter.TacticalUnsupportedReason != "" {
		t.Fatalf("tactical=%+v unsupported=%q", encounter.Tactical, encounter.TacticalUnsupportedReason)
	}
}
