package game

import (
	"path/filepath"
	"testing"

	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/protocol"
	"moox/internal/ruleset"
)

func tacticalMetadataRules(t *testing.T) *ruleset.TacticalCombatFile {
	t.Helper()
	rules, err := LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	return rules.TacticalCombat
}

func tacticalMetadataShip(id, empireID core.ID, drive string, armed bool) core.Ship {
	spec := core.ShipDesignSpec{
		HullID: "frigate", WarpDriveID: drive, ComputerID: "electronic_computer",
		ArmorID: "titanium_armor", FuelCellID: "standard_fuel_cells", FuelRangeParsecs: 4,
	}
	if armed {
		spec.Weapons = []core.ShipWeaponMount{{Slot: 0, WeaponID: "laser_cannon", Count: 1}}
	}
	return core.Ship{ID: id, EmpireID: empireID, Name: "Tactical Test", Spec: spec}
}

func TestTacticalMetadataSupportsTwoByTwoWithOpenFieldDeployment(t *testing.T) {
	state := &core.GameState{Ships: []core.Ship{
		tacticalMetadataShip(101, 1, "fusion_drive", true),
		tacticalMetadataShip(102, 1, "fusion_drive", true),
		tacticalMetadataShip(201, 2, "nuclear_drive", false),
		tacticalMetadataShip(202, 2, "nuclear_drive", false),
	}}
	encounter := Encounter{
		SystemID:     7,
		Attacker:     EncounterSide{EmpireID: 1, SeatID: 1, CombatFleetIDs: []core.ID{11}, ShipIDs: []core.ID{101, 102}},
		Defender:     EncounterSide{EmpireID: 2, SeatID: 2, CombatFleetIDs: []core.ID{22}, ShipIDs: []core.ID{201, 202}},
		Participants: []protocol.SeatID{1, 2},
	}
	tactical, reason, err := tacticalMetadataForEncounter(state, encounter, tacticalMetadataRules(t))
	if err != nil {
		t.Fatal(err)
	}
	if reason != "" || tactical == nil {
		t.Fatalf("2v2 tactical metadata unsupported: tactical=%+v reason=%q", tactical, reason)
	}
	if len(tactical.Ships) != 4 {
		t.Fatalf("tactical ships=%d want 4", len(tactical.Ships))
	}
	want := map[core.ID]struct{ x, y, facing, speed int }{
		101: {10, 9, 0, 22},
		102: {10, 11, 0, 22},
		201: {14, 9, 8, 20},
		202: {14, 11, 8, 20},
	}
	occupied := map[[2]int]core.ID{}
	for _, ship := range tactical.Ships {
		expected, ok := want[ship.ShipID]
		if !ok {
			t.Fatalf("unexpected tactical ship %d", ship.ShipID)
		}
		if ship.X != expected.x || ship.Y != expected.y || ship.Facing != expected.facing || ship.CurrentCombatSpeed != expected.speed {
			t.Fatalf("ship %d deployment=(%d,%d) facing=%d speed=%d want (%d,%d) facing=%d speed=%d", ship.ShipID, ship.X, ship.Y, ship.Facing, ship.CurrentCombatSpeed, expected.x, expected.y, expected.facing, expected.speed)
		}
		if ship.TurningMode != "normal" {
			t.Fatalf("ship %d turning mode=%q want normal", ship.ShipID, ship.TurningMode)
		}
		key := [2]int{ship.X, ship.Y}
		if prior := occupied[key]; prior != 0 {
			t.Fatalf("ships %d and %d share deployment %v", prior, ship.ShipID, key)
		}
		occupied[key] = ship.ShipID
	}
}

func TestTacticalMetadataPreservesPersistentShipVisualGenome(t *testing.T) {
	attacker := tacticalMetadataShip(101, 1, "fusion_drive", true)
	attacker.SourceDesignID = 501
	attacker.SourceDesignRevision = 3
	attacker.Spec.StrategicPictureID = 17
	genome := testPersistentVisualGenome()
	genome.HullID = attacker.Spec.HullID
	attacker.SourceVisualRevision = 7
	attacker.VisualGenome = &genome
	state := &core.GameState{Ships: []core.Ship{
		attacker,
		tacticalMetadataShip(201, 2, "nuclear_drive", false),
	}}
	encounter := Encounter{
		SystemID:     7,
		Attacker:     EncounterSide{EmpireID: 1, SeatID: 1, CombatFleetIDs: []core.ID{11}, ShipIDs: []core.ID{101}},
		Defender:     EncounterSide{EmpireID: 2, SeatID: 2, CombatFleetIDs: []core.ID{22}, ShipIDs: []core.ID{201}},
		Participants: []protocol.SeatID{1, 2},
	}
	tactical, reason, err := tacticalMetadataForEncounter(state, encounter, tacticalMetadataRules(t))
	if err != nil {
		t.Fatal(err)
	}
	if reason != "" || tactical == nil {
		t.Fatalf("tactical metadata unsupported: tactical=%+v reason=%q", tactical, reason)
	}
	var got *battle.TacticalShipSpec
	for i := range tactical.Ships {
		if tactical.Ships[i].ShipID == attacker.ID {
			got = &tactical.Ships[i]
			break
		}
	}
	if got == nil {
		t.Fatal("attacker TacticalShipSpec missing")
	}
	if got.SourceDesignID != attacker.SourceDesignID || got.SourceDesignRevision != attacker.SourceDesignRevision || got.StrategicPictureID != attacker.Spec.StrategicPictureID {
		t.Fatalf("source design identity=(%d,%d,%d) want (%d,%d,%d)", got.SourceDesignID, got.SourceDesignRevision, got.StrategicPictureID, attacker.SourceDesignID, attacker.SourceDesignRevision, attacker.Spec.StrategicPictureID)
	}
	if got.SourceVisualRevision != attacker.SourceVisualRevision {
		t.Fatalf("source visual revision=%d want %d", got.SourceVisualRevision, attacker.SourceVisualRevision)
	}
	if got.VisualGenome == nil || got.VisualGenome.Seed != genome.Seed || got.VisualGenome.HullID != genome.HullID {
		t.Fatalf("persisted visual genome not preserved: %+v", got.VisualGenome)
	}
	if got.VisualGenome == attacker.VisualGenome {
		t.Fatal("TacticalShipSpec aliases strategic Ship visual genome")
	}
	if len(got.VisualGenome.Primitives) > 0 && len(attacker.VisualGenome.Primitives) > 0 && &got.VisualGenome.Primitives[0] == &attacker.VisualGenome.Primitives[0] {
		t.Fatal("TacticalShipSpec aliases strategic Ship visual primitive slice")
	}
}

func TestTacticalMetadataAllowsAllUnarmedBattle(t *testing.T) {
	state := &core.GameState{Ships: []core.Ship{
		tacticalMetadataShip(101, 1, "nuclear_drive", false),
		tacticalMetadataShip(201, 2, "nuclear_drive", false),
	}}
	encounter := Encounter{
		SystemID:     7,
		Attacker:     EncounterSide{EmpireID: 1, SeatID: 1, CombatFleetIDs: []core.ID{11}, ShipIDs: []core.ID{101}},
		Defender:     EncounterSide{EmpireID: 2, SeatID: 2, CombatFleetIDs: []core.ID{22}, ShipIDs: []core.ID{201}},
		Participants: []protocol.SeatID{1, 2},
	}
	tactical, reason, err := tacticalMetadataForEncounter(state, encounter, tacticalMetadataRules(t))
	if err != nil {
		t.Fatal(err)
	}
	if tactical == nil || reason != "" || len(tactical.Ships) != 2 {
		t.Fatalf("all-unarmed battle tactical=%+v reason=%q", tactical, reason)
	}
	for _, ship := range tactical.Ships {
		if len(ship.Weapons) != 0 {
			t.Fatalf("unarmed tactical ship %d unexpectedly has weapons %+v", ship.ShipID, ship.Weapons)
		}
	}
}

func TestTacticalMetadataAllowsThreeByTwoAndIgnoresCivilianAndColonyContext(t *testing.T) {
	state := &core.GameState{Ships: []core.Ship{
		tacticalMetadataShip(101, 1, "fusion_drive", true),
		tacticalMetadataShip(102, 1, "fusion_drive", false),
		tacticalMetadataShip(103, 1, "fusion_drive", false),
		tacticalMetadataShip(201, 2, "nuclear_drive", false),
		tacticalMetadataShip(202, 2, "nuclear_drive", false),
	}}
	encounter := Encounter{
		SystemID:          7,
		Attacker:          EncounterSide{EmpireID: 1, SeatID: 1, CombatFleetIDs: []core.ID{11}, CivilianFleetIDs: []core.ID{12}, ShipIDs: []core.ID{101, 102, 103}},
		Defender:          EncounterSide{EmpireID: 2, SeatID: 2, CombatFleetIDs: []core.ID{22}, CivilianFleetIDs: []core.ID{23}, ShipIDs: []core.ID{201, 202}},
		DefenderColonyIDs: []core.ID{31},
		Participants:      []protocol.SeatID{1, 2},
	}
	tactical, reason, err := tacticalMetadataForEncounter(state, encounter, tacticalMetadataRules(t))
	if err != nil {
		t.Fatal(err)
	}
	if tactical == nil || reason != "" || len(tactical.Ships) != 5 {
		t.Fatalf("3v2 battle tactical=%+v reason=%q", tactical, reason)
	}
	want := []core.ID{101, 102, 103, 201, 202}
	for i, ship := range tactical.Ships {
		if ship.ShipID != want[i] {
			t.Fatalf("tactical ship[%d]=%d want %d", i, ship.ShipID, want[i])
		}
	}
}

func TestTacticalMetadataPreservesWeaponSlotsAndCounts(t *testing.T) {
	attacker := tacticalMetadataShip(101, 1, "fusion_drive", true)
	attacker.Spec.Weapons = []core.ShipWeaponMount{
		{Slot: 0, WeaponID: "laser_cannon", Count: 2},
		{Slot: 1, WeaponID: "laser_cannon", Count: 1},
	}
	state := &core.GameState{Ships: []core.Ship{
		attacker,
		tacticalMetadataShip(201, 2, "nuclear_drive", false),
	}}
	encounter := Encounter{
		SystemID:     7,
		Attacker:     EncounterSide{EmpireID: 1, SeatID: 1, CombatFleetIDs: []core.ID{11}, ShipIDs: []core.ID{101}},
		Defender:     EncounterSide{EmpireID: 2, SeatID: 2, CombatFleetIDs: []core.ID{22}, ShipIDs: []core.ID{201}},
		Participants: []protocol.SeatID{1, 2},
	}
	tactical, reason, err := tacticalMetadataForEncounter(state, encounter, tacticalMetadataRules(t))
	if err != nil {
		t.Fatal(err)
	}
	if reason != "" || tactical == nil {
		t.Fatalf("multi-mount tactical metadata unsupported: tactical=%+v reason=%q", tactical, reason)
	}
	var projected *battle.TacticalShipSpec
	for i := range tactical.Ships {
		if tactical.Ships[i].ShipID == attacker.ID {
			projected = &tactical.Ships[i]
			break
		}
	}
	if projected == nil || len(projected.Weapons) != 2 {
		t.Fatalf("projected attacker weapons=%+v", projected)
	}
	if projected.Weapons[0].Slot != 0 || projected.Weapons[0].Count != 2 || projected.Weapons[1].Slot != 1 || projected.Weapons[1].Count != 1 {
		t.Fatalf("projected weapons=%+v", projected.Weapons)
	}
	for _, weapon := range projected.Weapons {
		if weapon.WeaponID != "laser_cannon" || weapon.MinDamage != 1 || weapon.MaxDamage != 4 {
			t.Fatalf("projected Laser baseline=%+v", weapon)
		}
	}
}

func TestSlice176TacticalConfiguredBeamSnapshotEnablesGenericFire(t *testing.T) {
	attacker := tacticalMetadataShip(101, 1, "fusion_drive", true)
	attacker.Spec.Weapons = []core.ShipWeaponMount{
		{Slot: 0, WeaponID: "laser_cannon", Count: 1, MountType: core.ShipWeaponMountStandard},
		{
			Slot: 1, WeaponID: "phasor", Count: 1, MountType: core.ShipWeaponMountHeavy,
			ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierShieldPiercing}, FiringArcID: core.ShipWeaponFiringArc360,
		},
	}
	defender := tacticalMetadataShip(201, 2, "nuclear_drive", false)
	state := &core.GameState{Ships: []core.Ship{attacker, defender}}
	encounter := Encounter{
		SystemID:     7,
		Attacker:     EncounterSide{EmpireID: 1, SeatID: 1, CombatFleetIDs: []core.ID{11}, ShipIDs: []core.ID{101}},
		Defender:     EncounterSide{EmpireID: 2, SeatID: 2, CombatFleetIDs: []core.ID{22}, ShipIDs: []core.ID{201}},
		Participants: []protocol.SeatID{1, 2},
	}
	tactical, reason, err := tacticalMetadataForEncounter(state, encounter, tacticalMetadataRules(t))
	if err != nil {
		t.Fatal(err)
	}
	if reason != "" || tactical == nil {
		t.Fatalf("configured Beam tactical metadata unsupported: tactical=%+v reason=%q", tactical, reason)
	}
	var projected *battle.TacticalShipSpec
	for i := range tactical.Ships {
		if tactical.Ships[i].ShipID == attacker.ID {
			projected = &tactical.Ships[i]
			break
		}
	}
	if projected == nil || len(projected.Weapons) != 2 {
		t.Fatalf("projected attacker weapons=%+v", projected)
	}
	phasor := projected.Weapons[1]
	if phasor.WeaponID != "phasor" || phasor.MountType != core.ShipWeaponMountHeavy || phasor.FiringArcID != core.ShipWeaponFiringArc360 || phasor.MinDamage != 5 || phasor.MaxDamage != 20 || phasor.TacticalSupport != "beam" {
		t.Fatalf("projected Phasor metadata=%+v", phasor)
	}
	if len(phasor.ModifierIDs) != 1 || phasor.ModifierIDs[0] != core.ShipWeaponModifierShieldPiercing {
		t.Fatalf("projected Phasor modifiers=%v", phasor.ModifierIDs)
	}

	session, err := battle.NewSession(battle.Spec{
		ID: 17202, GameID: "slice-17-2", StrategicTurn: 1, SystemID: encounter.SystemID,
		Attacker: battle.Side{
			EmpireID: encounter.Attacker.EmpireID, SeatID: encounter.Attacker.SeatID,
			CombatFleetIDs: append([]core.ID(nil), encounter.Attacker.CombatFleetIDs...),
			ShipIDs:        append([]core.ID(nil), encounter.Attacker.ShipIDs...),
		},
		Defender: battle.Side{
			EmpireID: encounter.Defender.EmpireID, SeatID: encounter.Defender.SeatID,
			CombatFleetIDs: append([]core.ID(nil), encounter.Defender.CombatFleetIDs...),
			ShipIDs:        append([]core.ID(nil), encounter.Defender.ShipIDs...),
		},
		Participants: append([]protocol.SeatID(nil), encounter.Participants...),
		Seed:         17202, Tactical: tactical,
	})
	if err != nil {
		t.Fatalf("BattleSession rejected lossless configured Beam metadata: %v", err)
	}
	if err := session.Start(); err != nil {
		t.Fatalf("BattleSession start rejected configured Beam metadata: %v", err)
	}
	view := session.View()
	if view.Tactical == nil {
		t.Fatal("BattleSession tactical view missing")
	}
	var tacticalAttacker *battle.TacticalShipView
	for i := range view.Tactical.Ships {
		if view.Tactical.Ships[i].ShipID == attacker.ID {
			tacticalAttacker = &view.Tactical.Ships[i]
			break
		}
	}
	if tacticalAttacker == nil || len(tacticalAttacker.Weapons) != 2 {
		t.Fatalf("BattleSession attacker weapons=%+v", tacticalAttacker)
	}
	viewPhasor := tacticalAttacker.Weapons[1]
	if viewPhasor.WeaponID != "phasor" || viewPhasor.MountType != core.ShipWeaponMountHeavy || viewPhasor.FiringArcID != core.ShipWeaponFiringArc360 || viewPhasor.TacticalSupport != "beam" || !viewPhasor.Ready {
		t.Fatalf("BattleSession Phasor view=%+v", viewPhasor)
	}
	if len(viewPhasor.ModifierIDs) != 1 || viewPhasor.ModifierIDs[0] != core.ShipWeaponModifierShieldPiercing {
		t.Fatalf("BattleSession Phasor modifiers=%v", viewPhasor.ModifierIDs)
	}
	foundLaser, foundPhasor := false, false
	for _, action := range view.Tactical.LegalFireActions {
		if action.WeaponID == "laser_cannon" {
			foundLaser = true
		}
		if action.WeaponSlot == 1 && action.WeaponID == "phasor" {
			foundPhasor = true
		}
	}
	if !foundLaser || !foundPhasor {
		t.Fatalf("generic Beam fire actions missing laser=%v phasor=%v actions=%+v", foundLaser, foundPhasor, view.Tactical.LegalFireActions)
	}

	view.Spec.Tactical.Ships[0].Weapons[1].ModifierIDs[0] = core.ShipWeaponModifierAutofire
	fresh := session.View()
	var freshPhasor *battle.TacticalWeaponSpec
	for i := range fresh.Spec.Tactical.Ships {
		if fresh.Spec.Tactical.Ships[i].ShipID == attacker.ID {
			freshPhasor = &fresh.Spec.Tactical.Ships[i].Weapons[1]
			break
		}
	}
	if freshPhasor == nil || len(freshPhasor.ModifierIDs) != 1 || freshPhasor.ModifierIDs[0] != core.ShipWeaponModifierShieldPiercing {
		t.Fatalf("BattleSession weapon snapshot clone aliased modifiers: %+v", freshPhasor)
	}
}

func TestSlice177Checkpoint2TacticalMetadataProjectsExecutableOrdnance(t *testing.T) {
	attacker := tacticalMetadataShip(101, 1, "fusion_drive", true)
	attacker.Spec.Weapons = []core.ShipWeaponMount{{
		Slot: 0, WeaponID: "nuclear_missile", Count: 2, MountType: core.ShipWeaponMountStandard,
		Ammo: 5, FiringArcID: core.ShipWeaponFiringArc360,
	}}
	defender := tacticalMetadataShip(201, 2, "nuclear_drive", false)
	state := &core.GameState{Ships: []core.Ship{attacker, defender}}
	encounter := Encounter{
		SystemID:     7,
		Attacker:     EncounterSide{EmpireID: 1, SeatID: 1, CombatFleetIDs: []core.ID{11}, ShipIDs: []core.ID{101}},
		Defender:     EncounterSide{EmpireID: 2, SeatID: 2, CombatFleetIDs: []core.ID{22}, ShipIDs: []core.ID{201}},
		Participants: []protocol.SeatID{1, 2},
	}
	tactical, reason, err := tacticalMetadataForEncounter(state, encounter, tacticalMetadataRules(t))
	if err != nil {
		t.Fatal(err)
	}
	if reason != "" || tactical == nil {
		t.Fatalf("ordnance tactical metadata unsupported: tactical=%+v reason=%q", tactical, reason)
	}
	var projected *battle.TacticalShipSpec
	for i := range tactical.Ships {
		if tactical.Ships[i].ShipID == attacker.ID {
			projected = &tactical.Ships[i]
			break
		}
	}
	if projected == nil || len(projected.Weapons) != 1 {
		t.Fatalf("projected ordnance weapons=%+v", projected)
	}
	weapon := projected.Weapons[0]
	if weapon.WeaponID != "nuclear_missile" || weapon.Ammo != 5 || weapon.FiringArcID != core.ShipWeaponFiringArc360 || weapon.TacticalSupport != "ordnance" {
		t.Fatalf("projected ordnance metadata=%+v", weapon)
	}

	session, err := battle.NewSession(battle.Spec{
		ID: 17341, GameID: "slice-17-3", StrategicTurn: 1, SystemID: encounter.SystemID,
		Attacker: battle.Side{
			EmpireID: encounter.Attacker.EmpireID, SeatID: encounter.Attacker.SeatID,
			CombatFleetIDs: append([]core.ID(nil), encounter.Attacker.CombatFleetIDs...),
			ShipIDs:        append([]core.ID(nil), encounter.Attacker.ShipIDs...),
		},
		Defender: battle.Side{
			EmpireID: encounter.Defender.EmpireID, SeatID: encounter.Defender.SeatID,
			CombatFleetIDs: append([]core.ID(nil), encounter.Defender.CombatFleetIDs...),
			ShipIDs:        append([]core.ID(nil), encounter.Defender.ShipIDs...),
		},
		Participants: append([]protocol.SeatID(nil), encounter.Participants...),
		Seed:         17341, Tactical: tactical,
	})
	if err != nil {
		t.Fatalf("BattleSession rejected ordnance metadata: %v", err)
	}
	if err := session.Start(); err != nil {
		t.Fatalf("BattleSession start rejected ordnance metadata: %v", err)
	}
	view := session.View()
	if view.Tactical == nil {
		t.Fatal("BattleSession tactical view missing")
	}
	var viewWeapon *battle.TacticalWeaponView
	for i := range view.Tactical.Ships {
		if view.Tactical.Ships[i].ShipID == attacker.ID && len(view.Tactical.Ships[i].Weapons) == 1 {
			viewWeapon = &view.Tactical.Ships[i].Weapons[0]
			break
		}
	}
	if viewWeapon == nil || viewWeapon.WeaponID != "nuclear_missile" || viewWeapon.Ammo != 5 || viewWeapon.FiringArcID != core.ShipWeaponFiringArc360 || viewWeapon.TacticalSupport != "ordnance" || !viewWeapon.Ready {
		t.Fatalf("BattleSession ordnance view=%+v", viewWeapon)
	}
	for _, action := range view.Tactical.LegalFireActions {
		if action.WeaponID == "nuclear_missile" || action.WeaponSlot == 0 {
			t.Fatalf("ordnance leaked into legacy Beam fire-action surface: %+v", action)
		}
	}
}
