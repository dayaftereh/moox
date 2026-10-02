package session

import (
	"bytes"
	"path/filepath"
	"reflect"
	"testing"

	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func TestSlice177Gate3Checkpoint5SelectorSaveBuildBattleSessionOrdnanceIntegration(t *testing.T) {
	state, seats := twoSeatFixture(t)
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	rules.MineralIndustryPerWorker["abundant"] = 600
	if err := rules.ApplyReferenceTriangleTechnologyProfile(state, game.ReferenceTriangleProfileAllTech); err != nil {
		t.Fatal(err)
	}

	attackerEmpireID := state.Empires[0].ID
	defenderEmpireID := state.Empires[1].ID
	defenderSpec := core.ShipDesignSpec{
		HullID: "frigate", StrategicPictureID: 0, WarpDriveID: "nuclear_drive", FTLSpeed: 2,
		ComputerID: "electronic_computer", ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells", FuelRangeParsecs: 4,
		HullBaseCostPP: 20, HullSpace: 25, SpaceUsed: 10, BaseDesignCostPP: 25, ProductionCostPP: 25,
	}
	defenderDesignID := state.NewID()
	state.ShipDesigns = append(state.ShipDesigns, core.ShipDesign{
		ID: defenderDesignID, EmpireID: defenderEmpireID, Revision: 1, Name: "Checkpoint 5 Ordnance Target", Spec: defenderSpec,
	})
	defenderShipID := state.NewID()
	state.Ships = append(state.Ships, core.Ship{
		ID: defenderShipID, EmpireID: defenderEmpireID, SourceDesignID: defenderDesignID, SourceDesignRevision: 1,
		Name: "Checkpoint 5 Ordnance Target", Spec: defenderSpec,
	})
	defenderFleetID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
		ID: defenderFleetID, EmpireID: defenderEmpireID, Role: core.StrategicFleetRoleCombat,
		AtSystemID: state.Galaxy.Systems[0].ID, ShipIDs: []core.ID{defenderShipID},
	})
	state.DiplomaticRelations = reciprocalWarRelations(attackerEmpireID, defenderEmpireID)
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}

	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession("slice177-cp5-selector-build-ordnance", state, seats)
	if err != nil {
		t.Fatal(err)
	}

	wantMount := core.ShipWeaponMount{
		Slot: 0, WeaponID: "nuclear_missile", Count: 1, Ammo: 5,
		MountType: core.ShipWeaponMountStandard, FiringArcID: core.ShipWeaponFiringArc360,
	}
	save, err := game.NewSaveMilitaryDesignCommand(1, game.SaveMilitaryDesignPayload{
		Name: "CP5 Nuclear Rack", HullID: game.SupportedMilitaryHullID, StrategicPictureID: 2,
		Weapons: []core.ShipWeaponMount{wantMount},
	})
	if err != nil {
		t.Fatal(err)
	}
	first := submitColonyShipSessionTurn(t, s, resolver, []protocol.Command{save})
	var design *core.ShipDesign
	for i := range first.State.ShipDesigns {
		if first.State.ShipDesigns[i].EmpireID == attackerEmpireID && first.State.ShipDesigns[i].Name == "CP5 Nuclear Rack" {
			design = &first.State.ShipDesigns[i]
			break
		}
	}
	if design == nil || len(design.Spec.Weapons) != 1 {
		t.Fatalf("saved ordnance design=%+v", design)
	}
	if got := design.Spec.Weapons[0]; !reflect.DeepEqual(got, wantMount) {
		t.Fatalf("saved ordnance mount=%+v want=%+v", got, wantMount)
	}
	designID := design.ID
	completeColonyShipSessionTurn(t, s)

	queue, err := game.NewQueueMilitaryShipCommand(1, game.QueueMilitaryShipPayload{
		ColonyID: state.Colonies[0].ID, ShipDesignID: designID,
	})
	if err != nil {
		t.Fatal(err)
	}
	before, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion, GameID: before.GameID, SeatID: 1,
		Turn: before.Turn, BaseRevision: before.Revision, Commands: []protocol.Command{queue},
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion, GameID: before.GameID, SeatID: 2,
		Turn: before.Turn, BaseRevision: before.Revision,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	encounter, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if encounter.Phase != PhaseEncounters || len(encounter.Battles) != 1 {
		t.Fatalf("post-build encounter phase=%q battles=%d", encounter.Phase, len(encounter.Battles))
	}

	var built *core.Ship
	for i := range encounter.State.Ships {
		if encounter.State.Ships[i].EmpireID == attackerEmpireID && encounter.State.Ships[i].SourceDesignID == designID {
			built = &encounter.State.Ships[i]
			break
		}
	}
	if built == nil || len(built.Spec.Weapons) != 1 || !reflect.DeepEqual(built.Spec.Weapons[0], wantMount) {
		t.Fatalf("built ordnance ship=%+v", built)
	}

	child := encounter.Battles[0]
	if child.Spec.Tactical == nil || child.Tactical == nil {
		t.Fatalf("BattleSession Tactical projection missing: %+v", child)
	}
	var tacticalShip *battle.TacticalShipSpec
	for i := range child.Spec.Tactical.Ships {
		if child.Spec.Tactical.Ships[i].ShipID == built.ID {
			tacticalShip = &child.Spec.Tactical.Ships[i]
			break
		}
	}
	if tacticalShip == nil || len(tacticalShip.Weapons) != 1 {
		t.Fatalf("built Ship Tactical spec=%+v", tacticalShip)
	}
	tacticalWeapon := tacticalShip.Weapons[0]
	if tacticalWeapon.WeaponID != "nuclear_missile" || tacticalWeapon.WeaponFamily != "missile" ||
		tacticalWeapon.Count != 1 || tacticalWeapon.Ammo != 5 ||
		tacticalWeapon.FiringArcID != core.ShipWeaponFiringArc360 ||
		tacticalWeapon.MinDamage != 8 || tacticalWeapon.MaxDamage != 8 ||
		tacticalWeapon.TacticalSupport != "ordnance" {
		t.Fatalf("Tactical ordnance projection=%+v", tacticalWeapon)
	}

	var action *battle.TacticalWeaponActionAvailability
	for i := range child.Tactical.WeaponActions {
		if child.Tactical.WeaponActions[i].WeaponSlot == 0 {
			action = &child.Tactical.WeaponActions[i]
			break
		}
	}
	if action == nil || !action.Available || action.CommandKind != battle.CommandFireOrdnance ||
		len(action.Targets) != 1 || action.Targets[0].TargetShipID != defenderShipID {
		t.Fatalf("Tactical ordnance action=%+v", action)
	}
	decision, err := s.DecisionView(child.Spec.Attacker.SeatID, resolver)
	if err != nil {
		t.Fatal(err)
	}
	if len(decision.Decisions.Battles) != 1 || len(decision.Decisions.Battles[0].Actions) == 0 ||
		decision.Decisions.Battles[0].Actions[0].Kind != battle.CommandFireOrdnance {
		t.Fatalf("Tactical AI ordnance catalog=%+v", decision.Decisions.Battles)
	}
	if len(child.Tactical.Ordnance.WeaponStates) != 1 ||
		child.Tactical.Ordnance.WeaponStates[0].ShipID != built.ID ||
		child.Tactical.Ordnance.WeaponStates[0].AmmoRemaining != 5 {
		t.Fatalf("initial live ordnance state=%+v", child.Tactical.Ordnance)
	}

	encoded, err := s.MarshalLiveSnapshot(rules)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := UnmarshalLiveSnapshot(encoded, rules, resolver)
	if err != nil {
		t.Fatal(err)
	}
	reencoded, err := restored.MarshalLiveSnapshot(rules)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, reencoded) {
		t.Fatal("ordnance GameSession live snapshot changed on active-Encounter roundtrip")
	}
	restoredObserver, err := restored.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(encounter.Battles, restoredObserver.Battles) {
		t.Fatal("ordnance BattleSession projection changed after active-Encounter restore")
	}

	fire, err := battle.NewFireOrdnanceCommand(1, battle.FireOrdnancePayload{
		ShipID: built.ID, TargetShipID: defenderShipID, WeaponSlot: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []*GameSession{s, restored} {
		if err := target.SubmitBattleCommand(child.Spec.ID, child.Spec.Attacker.SeatID, fire); err != nil {
			t.Fatal(err)
		}
	}
	afterFire, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	restoredAfterFire, err := restored.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(afterFire.Battles, restoredAfterFire.Battles) || afterFire.Phase != restoredAfterFire.Phase {
		t.Fatal("ordnance continuation diverged after GameSession restore")
	}
	if len(afterFire.Battles) != 1 || afterFire.Battles[0].Tactical == nil {
		t.Fatalf("post-launch Tactical battle missing")
	}
	launched := false
	for _, event := range afterFire.Battles[0].Tactical.Events {
		if event.Kind == "ordnance_launched" && event.CommandSequence == 1 {
			launched = true
			break
		}
	}
	if !launched {
		t.Fatalf("Selector->Save->Build->BattleSession ordnance did not launch: events=%+v", afterFire.Battles[0].Tactical.Events)
	}
	if len(afterFire.Battles[0].Tactical.Ordnance.WeaponStates) != 1 ||
		afterFire.Battles[0].Tactical.Ordnance.WeaponStates[0].AmmoRemaining != 4 {
		t.Fatalf("post-launch ammo state=%+v", afterFire.Battles[0].Tactical.Ordnance.WeaponStates)
	}
}
