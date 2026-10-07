package battle

import (
	"encoding/json"
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func checkpoint2ActionBySlot(t *testing.T, view View, slot int) TacticalWeaponActionAvailability {
	t.Helper()
	if view.Tactical == nil {
		t.Fatal("tactical view missing")
	}
	for _, action := range view.Tactical.WeaponActions {
		if action.WeaponSlot == slot {
			return action
		}
	}
	t.Fatalf("weapon action for slot %d missing: %+v", slot, view.Tactical.WeaponActions)
	return TacticalWeaponActionAvailability{}
}

func checkpoint2Commit(t *testing.T, s *Session, seat protocol.SeatID, command protocol.Command) {
	t.Helper()
	prepared, err := s.PrepareCommand(seat, command)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.CommitPreparedCommand(prepared); err != nil {
		t.Fatal(err)
	}
}

func checkpoint2EndActive(t *testing.T, s *Session, spec Spec) {
	t.Helper()
	view := s.View()
	if view.Tactical == nil {
		t.Fatal("tactical view missing")
	}
	activeID := view.Tactical.State.ActiveShipID
	active := tacticalShipSpec(*spec.Tactical, activeID)
	if active == nil {
		t.Fatalf("active ship %d missing", activeID)
	}
	command, err := NewEndActivationCommand(view.Tactical.State.NextCommandSequence, EndActivationPayload{ShipID: activeID})
	if err != nil {
		t.Fatal(err)
	}
	checkpoint2Commit(t, s, active.SeatID, command)
}

func checkpoint2EventData(t *testing.T, event TacticalEvent) map[string]any {
	t.Helper()
	var data map[string]any
	if err := json.Unmarshal(event.Data, &data); err != nil {
		t.Fatal(err)
	}
	return data
}

func TestTacticalProjectileDestinationFollowsDirectDiagonalTrajectory(t *testing.T) {
	tests := []struct {
		name                           string
		fromX, fromY, targetX, targetY int
		maxFields                      int
		wantX, wantY, wantMoved        int
	}{
		{name: "fighter-speed-20", fromX: 25, fromY: 9, targetX: 54, targetY: 12, maxFields: 20, wantX: 44, wantY: 11, wantMoved: 20},
		{name: "missile-speed-16", fromX: 25, fromY: 9, targetX: 55, targetY: 13, maxFields: 16, wantX: 40, wantY: 11, wantMoved: 16},
		{name: "reverse-diagonal", fromX: 55, fromY: 13, targetX: 25, targetY: 9, maxFields: 16, wantX: 40, wantY: 11, wantMoved: 16},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x, y, moved := tacticalProjectileDestination(tc.fromX, tc.fromY, tc.targetX, tc.targetY, tc.maxFields)
			if x != tc.wantX || y != tc.wantY || moved != tc.wantMoved {
				t.Fatalf("destination=(%d,%d) moved=%d want (%d,%d) moved=%d", x, y, moved, tc.wantX, tc.wantY, tc.wantMoved)
			}
		})
	}
}

func TestSlice177Gate3Checkpoint2LaunchesAndMovesStableOrdnanceActors(t *testing.T) {
	spec := checkpoint1OrdnanceSpec()
	spec.Tactical.Ships[1].X = 100

	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}

	initial := s.View()
	missileAction := checkpoint2ActionBySlot(t, initial, 0)
	torpedoAction := checkpoint2ActionBySlot(t, initial, 1)
	if !missileAction.Available || missileAction.CommandKind != CommandFireOrdnance || len(missileAction.Targets) != 1 || missileAction.Targets[0].TargetShipID != 200 {
		t.Fatalf("Missile action=%+v", missileAction)
	}
	if !torpedoAction.Available || torpedoAction.CommandKind != CommandFireOrdnance || len(torpedoAction.Targets) != 1 || torpedoAction.Targets[0].TargetShipID != 200 {
		t.Fatalf("Torpedo action=%+v", torpedoAction)
	}

	missileCommand, err := NewFireOrdnanceCommand(initial.Tactical.State.NextCommandSequence, FireOrdnancePayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	if err != nil {
		t.Fatal(err)
	}
	checkpoint2Commit(t, s, 1, missileCommand)

	afterMissile := s.View()
	if afterMissile.Tactical.Ordnance.NextProjectileID != 2 || len(afterMissile.Tactical.Ordnance.Projectiles) != 1 {
		t.Fatalf("Missile launch ordnance state=%+v", afterMissile.Tactical.Ordnance)
	}
	if got := afterMissile.Tactical.Ordnance.WeaponStates[0]; got.AmmoRemaining != 4 || got.CooldownRounds != 0 {
		t.Fatalf("Missile live rack state=%+v", got)
	}
	missile := afterMissile.Tactical.Ordnance.Projectiles[0]
	if missile.ProjectileID != 1 || missile.SourceShipID != 100 || missile.TargetShipID != 200 || missile.WeaponID != "nuclear_missile" || missile.Multiplicity != 2 || missile.Speed != 12 || missile.UnitDurability != 4 || missile.WarheadStrength != 8 {
		t.Fatalf("Missile projectile=%+v", missile)
	}
	if missile.X != 22 || missile.Y != 10 || missile.DistanceTraveled != 240 || missile.MovementRemaining != 0 {
		t.Fatalf("Missile launch quantum=%+v", missile)
	}
	missileAction = checkpoint2ActionBySlot(t, afterMissile, 0)
	if missileAction.Available || missileAction.UnavailableReason != TacticalWeaponUnavailableNotReady {
		t.Fatalf("Missile post-launch action=%+v", missileAction)
	}
	if len(afterMissile.Tactical.Events) < 3 || afterMissile.Tactical.Events[len(afterMissile.Tactical.Events)-2].Kind != "ordnance_launched" || afterMissile.Tactical.Events[len(afterMissile.Tactical.Events)-1].Kind != "projectile_moved" {
		t.Fatalf("Missile event tail=%+v", afterMissile.Tactical.Events)
	}
	launchData := checkpoint2EventData(t, afterMissile.Tactical.Events[len(afterMissile.Tactical.Events)-2])
	if launchData["projectile_id"] != float64(1) || launchData["ammo_remaining"] != float64(4) || launchData["beam_interceptor_dcv"] != float64(50) {
		t.Fatalf("Missile launch event=%+v", launchData)
	}

	torpedoCommand, err := NewFireOrdnanceCommand(afterMissile.Tactical.State.NextCommandSequence, FireOrdnancePayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 1})
	if err != nil {
		t.Fatal(err)
	}
	checkpoint2Commit(t, s, 1, torpedoCommand)

	afterTorpedo := s.View()
	if afterTorpedo.Tactical.Ordnance.NextProjectileID != 3 || len(afterTorpedo.Tactical.Ordnance.Projectiles) != 2 {
		t.Fatalf("Torpedo launch ordnance state=%+v", afterTorpedo.Tactical.Ordnance)
	}
	if got := afterTorpedo.Tactical.Ordnance.WeaponStates[1]; got.CooldownRounds != 2 {
		t.Fatalf("Torpedo cooldown=%+v", got)
	}
	torpedo := afterTorpedo.Tactical.Ordnance.Projectiles[1]
	if torpedo.ProjectileID != 2 || torpedo.WeaponID != "anti_matter_torpedoes" || torpedo.Multiplicity != 3 || torpedo.Speed != 20 || torpedo.UnitDurability != 5 || torpedo.WarheadStrength != 25 {
		t.Fatalf("Torpedo projectile=%+v", torpedo)
	}
	if torpedo.X != 30 || torpedo.Y != 10 || torpedo.DistanceTraveled != 400 || torpedo.MovementRemaining != 0 {
		t.Fatalf("Torpedo launch quantum=%+v", torpedo)
	}
	torpedoAction = checkpoint2ActionBySlot(t, afterTorpedo, 1)
	if torpedoAction.Available || torpedoAction.UnavailableReason != TacticalWeaponUnavailableCooldown {
		t.Fatalf("Torpedo post-launch action=%+v", torpedoAction)
	}
}

func TestSlice177Gate3Checkpoint2RejectedLaunchIsAtomic(t *testing.T) {
	spec := checkpoint1OrdnanceSpec()
	spec.Tactical.Ships[1].X = 100
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	first, err := NewFireOrdnanceCommand(1, FireOrdnancePayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	if err != nil {
		t.Fatal(err)
	}
	checkpoint2Commit(t, s, 1, first)

	before := s.LiveSnapshot()
	second, err := NewFireOrdnanceCommand(before.Tactical.State.NextCommandSequence, FireOrdnancePayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.PrepareCommand(1, second); err == nil {
		t.Fatal("same-activation Missile relaunch unexpectedly prepared")
	}
	after := s.LiveSnapshot()
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("rejected ordnance command mutated authority\nbefore=%+v\nafter=%+v", before.Tactical.Ordnance, after.Tactical.Ordnance)
	}
}

func TestSlice177Gate3Checkpoint2TorpedoCooldownAndRoundMovement(t *testing.T) {
	spec := checkpoint1OrdnanceSpec()
	spec.Tactical.Ships[1].X = 100
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}

	fire, err := NewFireOrdnanceCommand(1, FireOrdnancePayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 1})
	if err != nil {
		t.Fatal(err)
	}
	checkpoint2Commit(t, s, 1, fire)
	if got := s.View().Tactical.Ordnance.Projectiles[0]; got.X != 30 || got.DistanceTraveled != 400 {
		t.Fatalf("launch quantum=%+v", got)
	}

	checkpoint2EndActive(t, s, spec)
	checkpoint2EndActive(t, s, spec)
	round2 := s.View()
	if round2.Tactical.State.Round != 2 {
		t.Fatalf("round=%d want=2", round2.Tactical.State.Round)
	}
	if got := round2.Tactical.Ordnance.WeaponStates[1]; got.CooldownRounds != 1 {
		t.Fatalf("round2 Torpedo cooldown=%+v", got)
	}
	if got := round2.Tactical.Ordnance.Projectiles[0]; got.X != 50 || got.DistanceTraveled != 800 {
		t.Fatalf("round2 projectile quantum=%+v", got)
	}
	if action := checkpoint2ActionBySlot(t, round2, 1); action.Available || action.UnavailableReason != TacticalWeaponUnavailableCooldown {
		t.Fatalf("round2 Torpedo action=%+v", action)
	}
	events := round2.Tactical.Events
	if len(events) < 2 || events[len(events)-2].Kind != "projectile_moved" || events[len(events)-1].Kind != "round_started" {
		t.Fatalf("round2 transition event tail=%+v", events)
	}

	checkpoint2EndActive(t, s, spec)
	checkpoint2EndActive(t, s, spec)
	round3 := s.View()
	if round3.Tactical.State.Round != 3 {
		t.Fatalf("round=%d want=3", round3.Tactical.State.Round)
	}
	if got := round3.Tactical.Ordnance.WeaponStates[1]; got.CooldownRounds != 0 {
		t.Fatalf("round3 Torpedo cooldown=%+v", got)
	}
	if got := round3.Tactical.Ordnance.Projectiles[0]; got.X != 70 || got.DistanceTraveled != 1200 {
		t.Fatalf("round3 projectile quantum=%+v", got)
	}
	if action := checkpoint2ActionBySlot(t, round3, 1); !action.Available || action.CommandKind != CommandFireOrdnance {
		t.Fatalf("round3 Torpedo action=%+v", action)
	}
}

func TestSlice177Gate3Checkpoint2RetargetsAscendingAndDestroysWithoutTarget(t *testing.T) {
	spec := checkpoint1OrdnanceSpec()
	spec.Defender.ShipIDs = []core.ID{200, 202, 201}
	template := spec.Tactical.Ships[1]
	template.ShipID = 202
	template.X = 100
	spec.Tactical.Ships = append(spec.Tactical.Ships, template)
	template.ShipID = 201
	template.X = 80
	spec.Tactical.Ships = append(spec.Tactical.Ships, template)

	r, err := newTacticalRuntime(*spec.Tactical)
	if err != nil {
		t.Fatal(err)
	}
	r.ordnance.Projectiles = []TacticalProjectileState{{
		ProjectileID: 1, SourceSeatID: 1, SourceShipID: 100, WeaponSlot: 0,
		TargetShipID: 200, WeaponID: "nuclear_missile", WeaponFamily: "missile",
		X: 10, Y: 10, Facing: 0, Speed: 12, Multiplicity: 2, UnitDurability: 4, WarheadStrength: 8,
	}}
	r.ordnance.NextProjectileID = 2
	r.shipState(200).Destroyed = true
	if err := processTacticalProjectiles(*spec.Tactical, &r, 0, 0); err != nil {
		t.Fatal(err)
	}
	if len(r.ordnance.Projectiles) != 1 || r.ordnance.Projectiles[0].TargetShipID != 201 {
		t.Fatalf("retargeted projectile=%+v", r.ordnance.Projectiles)
	}
	if len(r.events) < 3 || r.events[len(r.events)-2].Kind != "projectile_retargeted" || r.events[len(r.events)-1].Kind != "projectile_moved" {
		t.Fatalf("retarget event tail=%+v", r.events)
	}
	retargetData := checkpoint2EventData(t, r.events[len(r.events)-2])
	if retargetData["previous_target_ship_id"] != float64(200) || retargetData["new_target_ship_id"] != float64(201) {
		t.Fatalf("retarget event=%+v", retargetData)
	}

	r.shipState(201).Destroyed = true
	r.shipState(202).Destroyed = true
	r.ordnance.Projectiles[0].TargetShipID = 200
	if err := processTacticalProjectiles(*spec.Tactical, &r, 0, 0); err != nil {
		t.Fatal(err)
	}
	if len(r.ordnance.Projectiles) != 0 {
		t.Fatalf("no-target projectile survived=%+v", r.ordnance.Projectiles)
	}
	last := r.events[len(r.events)-1]
	if last.Kind != "projectile_destroyed" {
		t.Fatalf("last event=%+v", last)
	}
	destroyedData := checkpoint2EventData(t, last)
	if destroyedData["projectile_id"] != float64(1) || destroyedData["reason"] != "no_target" {
		t.Fatalf("projectile destroyed event=%+v", destroyedData)
	}
}

func TestSlice177Gate3Checkpoint2EmissionsGuidanceRemainsUnavailable(t *testing.T) {
	spec := checkpoint1OrdnanceSpec()
	spec.Tactical.Ships[0].Weapons[0].ModifierIDs = []core.ShipWeaponModifierID{core.ShipWeaponModifierEmissionsGuidance}
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	view := s.View()
	action := checkpoint2ActionBySlot(t, view, 0)
	if action.Available || action.CommandKind != CommandFireOrdnance || action.UnavailableReason != TacticalWeaponUnavailableOrdnanceEffect {
		t.Fatalf("Emissions Guidance action=%+v", action)
	}
	before := s.LiveSnapshot()
	command, err := NewFireOrdnanceCommand(view.Tactical.State.NextCommandSequence, FireOrdnancePayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.PrepareCommand(1, command); err == nil {
		t.Fatal("Emissions Guidance launch unexpectedly prepared")
	}
	if !reflect.DeepEqual(before, s.LiveSnapshot()) {
		t.Fatal("rejected Emissions Guidance launch mutated BattleSession")
	}
}

func TestSlice177Gate3Checkpoint2MissileAmmoExhaustionProjectsExplicitReason(t *testing.T) {
	spec := checkpoint1OrdnanceSpec()
	spec.Tactical.Ships[0].Weapons[0].Ammo = 1
	spec.Tactical.Ships[1].X = 100
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	view := s.View()
	command, err := NewFireOrdnanceCommand(view.Tactical.State.NextCommandSequence, FireOrdnancePayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	if err != nil {
		t.Fatal(err)
	}
	checkpoint2Commit(t, s, 1, command)
	after := s.View()
	if got := after.Tactical.Ordnance.WeaponStates[0].AmmoRemaining; got != 0 {
		t.Fatalf("Missile ammo remaining=%d want=0", got)
	}
	action := checkpoint2ActionBySlot(t, after, 0)
	if action.Available || action.CommandKind != CommandFireOrdnance || action.UnavailableReason != TacticalWeaponUnavailableAmmoExhausted {
		t.Fatalf("exhausted Missile action=%+v", action)
	}
}
