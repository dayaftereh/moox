package battle

import (
	"bytes"
	"encoding/json"
	"testing"

	"moox/internal/core"
)

func TestSlice178Gate3Checkpoint6VectorKAchillesInternalWeights(t *testing.T) {
	state := TacticalInternalState{
		Weapons:                []TacticalInternalWeaponState{{Slot: 0, Current: 10}, {Slot: 1, Current: 4}},
		Specials:               []TacticalInternalSpecialState{{Slot: 0, Current: 12}},
		ShieldGeneratorCurrent: 20,
	}
	got := tacticalBeamInternalWeights(state, true)
	if got[2] != 30 || got[3] != 12 || got[18] != 60 || got[10] != 12 {
		t.Fatalf("Achilles weights weaponA/weaponB/shield/other=%d/%d/%d/%d want=30/12/60/12", got[2], got[3], got[18], got[10])
	}
	normal := tacticalBeamInternalWeights(state, false)
	if normal[2] != 10 || normal[3] != 4 || normal[18] != 20 || normal[10] != 12 {
		t.Fatalf("normal weights weaponA/weaponB/shield/other=%d/%d/%d/%d want=10/4/20/12", normal[2], normal[3], normal[18], normal[10])
	}
}

func TestSlice178Gate3Checkpoint6InternalThresholdOriginalFormula(t *testing.T) {
	target := TacticalShipSpec{StructureMax: 40}
	state := TacticalShipState{StructureDamage: 10, Internal: TacticalInternalState{EngineCurrent: 20, SecondaryCurrent: 2, Weapons: []TacticalInternalWeaponState{{Slot: 0, Current: 8}}, Specials: []TacticalInternalSpecialState{{Slot: 0, Current: 4}}, ShieldGeneratorCurrent: 6}}
	// Original total is Engine/2 + Secondary + Weapons + Specials + Shield = 10+2+8+4+6 = 30.
	if got := tacticalBeamInternalTotal(state.Internal); got != 30 {
		t.Fatalf("internal total=%d want=30", got)
	}
	// remaining Structure is 30 -> floor(30*100/(30+30)) = 50.
	if got := tacticalBeamInternalThreshold(target, state, false, false); got != 50 {
		t.Fatalf("normal threshold=%d want=50", got)
	}
	if got := tacticalBeamInternalThreshold(target, state, true, false); got != 100 {
		t.Fatalf("Achilles threshold=%d want=100", got)
	}
	if got := tacticalBeamInternalThreshold(target, state, false, true); got != 25 {
		t.Fatalf("monster/Antaran threshold=%d want=25", got)
	}
}

func TestSlice178Gate3Checkpoint6PersistenceDetachmentRoundTrip(t *testing.T) {
	spec := slice178Checkpoint3Spec("fighter_bays", 1)
	spec.Tactical.Rules.Specials = slice178Checkpoint2SpecialRules()
	spec.Tactical.Ships[0].Specials = []TacticalSpecialSpec{
		{Slot: 0, SpecialID: "time_warp_facilitator", TacticalSupport: "tactical_passive"},
		{Slot: 1, SpecialID: "cloaking_device", TacticalSupport: "tactical_passive"},
	}
	spec.Tactical.Ships[0].Internal = &TacticalInternalSpec{
		EngineMax:    2,
		SecondaryMax: 1,
		Weapons: []TacticalInternalWeaponSpec{
			{Slot: 0, Max: 1, UnitSize: 1},
			{Slot: 1, Max: 1, UnitSize: 1},
		},
		Specials: []TacticalInternalSpecialSpec{{Slot: 0, Max: 1}, {Slot: 1, Max: 1}},
	}
	if err := validateSpec(spec); err != nil {
		t.Fatalf("combined CP6 spec invalid: %v", err)
	}

	session, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Start(); err != nil {
		t.Fatal(err)
	}
	launch, err := NewLaunchFightersCommand(1, LaunchFightersPayload{ShipID: 100, WeaponSlot: 1, TargetShipID: 200})
	if err != nil {
		t.Fatal(err)
	}
	submitPrepared(t, session, 1, launch)
	if len(session.tactical.fighters.Groups) != 1 {
		t.Fatalf("fighter groups after launch=%+v", session.tactical.fighters.Groups)
	}

	captured := session.tactical.shipState(200)
	captured.OwnerEmpireID = 1
	captured.ControllerSeatID = 1
	captured.PostCaptureState = true
	captured.ActivationComplete = true
	if err := reconcileTacticalCaptureMovingActors(*spec.Tactical, session.tactical, 200, 1, 1, 1); err != nil {
		t.Fatal(err)
	}
	if len(session.tactical.fighters.Groups) != 1 || session.tactical.fighters.Groups[0].Phase != TacticalFighterPhaseReturning {
		t.Fatalf("capture did not preserve a valid returning Fighter group: %+v", session.tactical.fighters.Groups)
	}

	source := session.tactical.shipState(100)
	source.ActivationComplete = false
	session.tactical.state.PassKind = TacticalPassKindTimeWarp
	session.tactical.state.TimeWarpSeatID = 1
	session.tactical.state.TimeWarpQueue = []core.ID{100}
	session.tactical.state.ActiveShipID = 100

	before := session.LiveSnapshot()
	if err := validateLiveSnapshot(before); err != nil {
		t.Fatalf("combined CP6 snapshot rejected before transport: %v", err)
	}
	encoded, err := json.Marshal(before)
	if err != nil {
		t.Fatal(err)
	}
	var transported LiveSnapshot
	if err := json.Unmarshal(encoded, &transported); err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreLiveSnapshot(transported)
	if err != nil {
		t.Fatal(err)
	}
	afterJSON, err := json.Marshal(restored.LiveSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, afterJSON) {
		t.Fatalf("CP6 save/load/reconnect roundtrip changed authoritative snapshot\nbefore=%s\nafter =%s", encoded, afterJSON)
	}

	got := restored.LiveSnapshot()
	if got.Tactical.State.PassKind != TacticalPassKindTimeWarp || got.Tactical.State.TimeWarpSeatID != 1 || len(got.Tactical.State.TimeWarpQueue) != 1 || got.Tactical.State.TimeWarpQueue[0] != 100 {
		t.Fatalf("restored Time-Warp state=%+v", got.Tactical.State)
	}
	if len(got.Tactical.Fighters.Groups) != 1 || got.Tactical.Fighters.Groups[0].Phase != TacticalFighterPhaseReturning {
		t.Fatalf("restored Fighter state=%+v", got.Tactical.Fighters)
	}
	var restoredCaptured *TacticalShipState
	for i := range got.Tactical.State.Ships {
		if got.Tactical.State.Ships[i].ShipID == 200 {
			restoredCaptured = &got.Tactical.State.Ships[i]
			break
		}
	}
	if restoredCaptured == nil || restoredCaptured.OwnerEmpireID != 1 || restoredCaptured.ControllerSeatID != 1 || !restoredCaptured.PostCaptureState {
		t.Fatalf("restored mutable authority=%+v", restoredCaptured)
	}

	player, err := restored.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	observer := restored.View()
	if player.Tactical == nil || observer.Tactical == nil || len(player.Tactical.Fighters.Groups) != 1 || len(observer.Tactical.Fighters.Groups) != 1 {
		t.Fatalf("missing CP6 projection after restore player=%+v observer=%+v", player.Tactical, observer.Tactical)
	}
	var playerSource *TacticalShipView
	for i := range player.Tactical.Ships {
		if player.Tactical.Ships[i].ShipID == 100 {
			playerSource = &player.Tactical.Ships[i]
			break
		}
	}
	if playerSource == nil || len(playerSource.Specials) != 2 || !playerSource.Specials[0].Functional || !playerSource.Specials[1].Functional {
		t.Fatalf("server-projected Special functionality=%+v", playerSource)
	}

	player.Tactical.Fighters.Groups[0].Multiplicity = 999
	playerSource.Specials[0].Functional = false
	playerSource.Internal.EngineCurrent = 999
	observer.Tactical.Fighters.Groups[0].Multiplicity = 777
	observer.Tactical.State.TimeWarpQueue[0] = 200
	freshPlayer, err := restored.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	freshObserver := restored.View()
	if freshPlayer.Tactical.Fighters.Groups[0].Multiplicity == 999 || freshObserver.Tactical.Fighters.Groups[0].Multiplicity == 777 {
		t.Fatal("Fighter projection aliases authoritative runtime")
	}
	if freshObserver.Tactical.State.TimeWarpQueue[0] != 100 {
		t.Fatalf("observer Time-Warp queue aliases authority: %+v", freshObserver.Tactical.State.TimeWarpQueue)
	}
	for _, ship := range freshPlayer.Tactical.Ships {
		if ship.ShipID == 100 {
			if len(ship.Specials) != 2 || !ship.Specials[0].Functional || ship.Internal.EngineCurrent == 999 {
				t.Fatalf("player Special/internal projection aliases authority: %+v", ship)
			}
		}
	}
	sequence := freshObserver.Tactical.State.NextCommandSequence
	endOriginal, err := NewEndActivationCommand(sequence, EndActivationPayload{ShipID: 100})
	if err != nil {
		t.Fatal(err)
	}
	endRestored, err := NewEndActivationCommand(sequence, EndActivationPayload{ShipID: 100})
	if err != nil {
		t.Fatal(err)
	}
	submitPrepared(t, session, 1, endOriginal)
	submitPrepared(t, restored, 1, endRestored)
	left, err := json.Marshal(session.LiveSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	right, err := json.Marshal(restored.LiveSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(left, right) {
		t.Fatalf("CP6 replay/continuation diverged after restore\noriginal=%s\nrestored=%s", left, right)
	}
}

func TestSlice178Gate3Checkpoint6VectorLLegacyCompatibilityNoHiddenState(t *testing.T) {
	original := newStartedTacticalSession(t)
	snapshot := original.LiveSnapshot()
	snapshot.SchemaVersion = legacyLiveSnapshotSchemaVersion
	snapshot.Tactical.NextMovingActorOrder = 0
	snapshot.Tactical.Legacy17_8Disabled = false

	restored, err := RestoreLiveSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	got := restored.LiveSnapshot()
	if got.SchemaVersion != LiveSnapshotSchemaVersion || got.Tactical == nil || !got.Tactical.Legacy17_8Disabled {
		t.Fatalf("legacy compatibility header=%+v", got.Tactical)
	}
	if got.Tactical.Fighters != nil && (len(got.Tactical.Fighters.BayStates) != 0 || len(got.Tactical.Fighters.Groups) != 0) {
		t.Fatalf("legacy migration synthesized Fighter state: %+v", got.Tactical.Fighters)
	}
	for _, ship := range got.Tactical.State.Ships {
		if ship.Internal.EngineCurrent != 0 || ship.Internal.SecondaryCurrent != 0 || ship.Internal.ShieldGeneratorCurrent != 0 || len(ship.Internal.Weapons) != 0 || len(ship.Internal.Specials) != 0 || ship.CloakState != TacticalCloakStateNone || ship.EnergyAbsorberStored != 0 {
			t.Fatalf("legacy migration synthesized Slice-17.8 hidden state: %+v", ship)
		}
	}
	reexported := restored.LiveSnapshot()
	if reexported.SchemaVersion != LiveSnapshotSchemaVersion || reexported.Tactical == nil || !reexported.Tactical.Legacy17_8Disabled {
		t.Fatalf("legacy compatibility state was not preserved on re-save: %+v", reexported.Tactical)
	}
}

func TestSlice178Gate3Checkpoint6AchillesWiresIntoBeamDamage(t *testing.T) {
	spec := baselineTacticalBattleSpec()
	spec.Tactical.Rules.Specials = slice178Checkpoint2SpecialRules()
	spec.Tactical.Ships[0].Specials = []TacticalSpecialSpec{{Slot: 0, SpecialID: "achilles_targeting_unit", TacticalSupport: "tactical_passive"}}
	spec.Tactical.Ships[0].Internal = &TacticalInternalSpec{
		EngineMax: 2, SecondaryMax: 1,
		Weapons:  []TacticalInternalWeaponSpec{{Slot: 0, Max: 1, UnitSize: 1}},
		Specials: []TacticalInternalSpecialSpec{{Slot: 0, Max: 1}},
	}
	spec.Tactical.Ships[1].Internal = &TacticalInternalSpec{EngineMax: 2, SecondaryMax: 1}
	spec.Tactical.Ships[1].X = spec.Tactical.Ships[0].X + 2

	r, err := newTacticalRuntime(*spec.Tactical)
	if err != nil {
		t.Fatal(err)
	}
	target := r.shipState(spec.Tactical.Ships[1].ShipID)
	if target == nil {
		t.Fatal("target runtime state missing")
	}
	armorBefore := target.ArmorCurrent
	internalBefore := tacticalBeamInternalTotal(target.Internal)
	if armorBefore <= 0 || internalBefore <= 0 {
		t.Fatalf("Achilles fixture armor/internal=%d/%d", armorBefore, internalBefore)
	}
	if _, err := prepareFireBeam(spec, &r, spec.Tactical.Ships[0].SeatID, r.state.NextCommandSequence, FireBeamPayload{
		ShipID: spec.Tactical.Ships[0].ShipID, TargetShipID: spec.Tactical.Ships[1].ShipID, WeaponSlot: 0,
	}); err != nil {
		t.Fatal(err)
	}
	target = r.shipState(spec.Tactical.Ships[1].ShipID)
	if target.ArmorCurrent != armorBefore {
		t.Fatalf("Achilles Beam consumed armor=%d want unchanged %d", target.ArmorCurrent, armorBefore)
	}
	if got := tacticalBeamInternalTotal(target.Internal); got >= internalBefore {
		t.Fatalf("Achilles Beam internal total=%d want below %d", got, internalBefore)
	}
}
