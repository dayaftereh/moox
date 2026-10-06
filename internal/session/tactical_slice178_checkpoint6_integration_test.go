package session

import (
	"path/filepath"
	"reflect"
	"testing"

	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/protocol"
)

func TestSlice178Gate3Checkpoint6SelectorSaveBuildBattleSessionIntegration(t *testing.T) {
	state, seats := twoSeatFixture(t)
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	rules.MineralIndustryPerWorker["abundant"] = 1200
	if err := rules.ApplyReferenceTriangleTechnologyProfile(state, game.ReferenceTriangleProfileAllTech); err != nil {
		t.Fatal(err)
	}
	state.Colonies[0].Buildings = append(state.Colonies[0].Buildings, "star_base")

	attackerEmpireID := state.Empires[0].ID
	defenderEmpireID := state.Empires[1].ID
	defenderSpec := core.ShipDesignSpec{
		HullID: "frigate", StrategicPictureID: 0, WarpDriveID: "nuclear_drive", FTLSpeed: 2,
		ComputerID: "electronic_computer", ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells", FuelRangeParsecs: 4,
		HullBaseCostPP: 20, HullSpace: 25, SpaceUsed: 10, BaseDesignCostPP: 25, ProductionCostPP: 25,
	}
	defenderDesignID := state.NewID()
	state.ShipDesigns = append(state.ShipDesigns, core.ShipDesign{ID: defenderDesignID, EmpireID: defenderEmpireID, Revision: 1, Name: "CP6 Target", Spec: defenderSpec})
	defenderShipID := state.NewID()
	state.Ships = append(state.Ships, core.Ship{ID: defenderShipID, EmpireID: defenderEmpireID, SourceDesignID: defenderDesignID, SourceDesignRevision: 1, Name: "CP6 Target", Spec: defenderSpec})
	defenderFleetID := state.NewID()
	state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{ID: defenderFleetID, EmpireID: defenderEmpireID, Role: core.StrategicFleetRoleCombat, AtSystemID: state.Galaxy.Systems[0].ID, ShipIDs: []core.ID{defenderShipID}})
	state.DiplomaticRelations = reciprocalWarRelations(attackerEmpireID, defenderEmpireID)
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}

	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession("slice178-cp6-selector-build-tactical", state, seats)
	if err != nil {
		t.Fatal(err)
	}

	wantWeapon := core.ShipWeaponMount{Slot: 0, WeaponID: "fighter_bays", Count: 1, MountType: core.ShipWeaponMountStandard}
	wantSpecials := []core.ShipSpecialMount{
		{Slot: 0, SpecialID: "achilles_targeting_unit"},
		{Slot: 1, SpecialID: "time_warp_facilitator"},
		{Slot: 2, SpecialID: "transporters"},
	}
	save, err := game.NewSaveMilitaryDesignCommand(1, game.SaveMilitaryDesignPayload{
		Name: "CP6 Carrier", HullID: "cruiser", StrategicPictureID: 16,
		Weapons: []core.ShipWeaponMount{wantWeapon}, Specials: wantSpecials,
	})
	if err != nil {
		t.Fatal(err)
	}
	first := submitColonyShipSessionTurn(t, s, resolver, []protocol.Command{save})
	var design *core.ShipDesign
	for i := range first.State.ShipDesigns {
		if first.State.ShipDesigns[i].EmpireID == attackerEmpireID && first.State.ShipDesigns[i].Name == "CP6 Carrier" {
			design = &first.State.ShipDesigns[i]
			break
		}
	}
	if design == nil || !reflect.DeepEqual(design.Spec.Weapons, []core.ShipWeaponMount{wantWeapon}) || !reflect.DeepEqual(design.Spec.Specials, wantSpecials) {
		t.Fatalf("saved Slice-17.8 carrier design=%+v", design)
	}
	designID := design.ID
	completeColonyShipSessionTurn(t, s)

	queue, err := game.NewQueueMilitaryShipCommand(1, game.QueueMilitaryShipPayload{ColonyID: state.Colonies[0].ID, ShipDesignID: designID})
	if err != nil {
		t.Fatal(err)
	}
	before, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: before.GameID, SeatID: 1, Turn: before.Turn, BaseRevision: before.Revision, Commands: []protocol.Command{queue}}); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: before.GameID, SeatID: 2, Turn: before.Turn, BaseRevision: before.Revision}); err != nil {
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
	if built == nil || !reflect.DeepEqual(built.Spec.Weapons, design.Spec.Weapons) || !reflect.DeepEqual(built.Spec.Specials, design.Spec.Specials) {
		t.Fatalf("built Slice-17.8 carrier=%+v", built)
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
	if tacticalShip == nil || len(tacticalShip.Weapons) != 1 || tacticalShip.Weapons[0].WeaponFamily != "fighter" || tacticalShip.Weapons[0].TacticalSupport != "fighter" || tacticalShip.Weapons[0].CraftPerUnit != 4 {
		t.Fatalf("Tactical Fighter projection=%+v", tacticalShip)
	}
	if !reflect.DeepEqual([]string{tacticalShip.Specials[0].SpecialID, tacticalShip.Specials[1].SpecialID, tacticalShip.Specials[2].SpecialID}, []string{"achilles_targeting_unit", "time_warp_facilitator", "transporters"}) {
		t.Fatalf("Tactical Special projection=%+v", tacticalShip.Specials)
	}
	if tacticalShip.Internal == nil || len(tacticalShip.Internal.Weapons) != 1 || len(tacticalShip.Internal.Specials) != 3 {
		t.Fatalf("Tactical Internal projection=%+v", tacticalShip.Internal)
	}
	if len(child.Spec.Tactical.FighterDoctrines) == 0 || child.Spec.Tactical.FighterDoctrines[0].SeatID != 1 || child.Spec.Tactical.FighterDoctrines[0].BeamPayload == nil {
		t.Fatalf("Tactical Fighter doctrine=%+v", child.Spec.Tactical.FighterDoctrines)
	}

	var fighterAction *battle.TacticalWeaponActionAvailability
	for i := range child.Tactical.WeaponActions {
		if child.Tactical.WeaponActions[i].ShipID == built.ID && child.Tactical.WeaponActions[i].WeaponSlot == 0 {
			fighterAction = &child.Tactical.WeaponActions[i]
			break
		}
	}
	if child.Tactical.State.ActiveShipID != built.ID || fighterAction == nil || !fighterAction.Available || fighterAction.CommandKind != battle.CommandLaunchFighters || len(fighterAction.Targets) == 0 {
		t.Fatalf("Tactical Fighter action active=%d action=%+v", child.Tactical.State.ActiveShipID, fighterAction)
	}

	if len(s.battles) != 1 {
		t.Fatalf("child BattleSessions=%d want=1", len(s.battles))
	}
	player, err := s.battles[0].PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	var playerShip *battle.TacticalShipView
	for i := range player.Tactical.Ships {
		if player.Tactical.Ships[i].ShipID == built.ID {
			playerShip = &player.Tactical.Ships[i]
			break
		}
	}
	if playerShip == nil || len(playerShip.Specials) != 3 || !playerShip.Specials[0].Functional || !playerShip.Specials[1].Functional || !playerShip.Specials[2].Functional {
		t.Fatalf("server-projected carrier Specials=%+v", playerShip)
	}

	target := fighterAction.Targets[0]
	launch, err := battle.NewLaunchFightersCommand(child.Tactical.State.NextCommandSequence, battle.LaunchFightersPayload{ShipID: built.ID, WeaponSlot: 0, TargetShipID: target.TargetShipID, TargetProjectileID: target.TargetProjectileID, TargetFighterGroupID: target.TargetFighterGroupID})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitBattleCommand(child.Spec.ID, child.Spec.Attacker.SeatID, launch); err != nil {
		t.Fatal(err)
	}
	afterLaunch, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if len(afterLaunch.Battles) != 1 || afterLaunch.Battles[0].Tactical == nil || len(afterLaunch.Battles[0].Tactical.Fighters.Groups) != 1 {
		t.Fatalf("Selector->Save->Build->BattleSession Fighter launch missing: %+v", afterLaunch.Battles)
	}
	foundLaunch := false
	for _, event := range afterLaunch.Battles[0].Tactical.Events {
		if event.Kind == "fighter_group_launched" {
			foundLaunch = true
			break
		}
	}
	if !foundLaunch {
		t.Fatalf("Selector->Save->Build->BattleSession Fighter event missing: %+v", afterLaunch.Battles[0].Tactical.Events)
	}
}
