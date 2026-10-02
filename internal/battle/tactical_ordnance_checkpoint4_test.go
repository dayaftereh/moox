package battle

import (
	"reflect"
	"testing"

	"moox/internal/core"
)

func checkpoint4ProjectileTargetSpec() Spec {
	spec := baselineTacticalBattleSpec()
	spec.Tactical.Rules.Ordnance = checkpoint1OrdnanceRules()
	spec.Tactical.Ships[0].Weapons = []TacticalWeaponSpec{
		{Slot: 0, WeaponID: "laser_cannon", WeaponFamily: "beam", Count: 1, MountType: core.ShipWeaponMountStandard, FiringArcID: core.ShipWeaponFiringArc360, MinDamage: 1, MaxDamage: 4, TacticalSupport: "beam"},
		{Slot: 1, WeaponID: "laser_cannon", WeaponFamily: "beam", Count: 1, MountType: core.ShipWeaponMountPointDefense, FiringArcID: core.ShipWeaponFiringArc360, MinDamage: 1, MaxDamage: 4, TacticalSupport: "beam"},
		{Slot: 2, WeaponID: "laser_cannon", WeaponFamily: "beam", Count: 1, MountType: core.ShipWeaponMountHeavy, FiringArcID: core.ShipWeaponFiringArc360, MinDamage: 1, MaxDamage: 4, TacticalSupport: "beam"},
	}
	return spec
}

func checkpoint4EnemyProjectile() TacticalProjectileState {
	return TacticalProjectileState{
		ProjectileID:      1,
		SourceSeatID:      2,
		SourceShipID:      200,
		WeaponSlot:        0,
		TargetShipID:      100,
		WeaponID:          "nuclear_missile",
		WeaponFamily:      "missile",
		X:                 11,
		Y:                 10,
		Facing:            8,
		Speed:             12,
		MovementRemaining: 0,
		Multiplicity:      1,
		UnitDurability:    4,
		WarheadStrength:   8,
	}
}

func checkpoint4InjectProjectile(t *testing.T, s *Session, projectile TacticalProjectileState) {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tactical.ordnance.Projectiles = []TacticalProjectileState{projectile}
	s.tactical.ordnance.NextProjectileID = projectile.ProjectileID + 1
}

func checkpoint4TargetHasProjectile(targets []TacticalFireTarget, projectileID uint64) bool {
	for _, target := range targets {
		if target.TargetProjectileID == projectileID {
			return true
		}
	}
	return false
}

func TestSlice177Gate3Checkpoint4ProjectileInterceptionRemainderArithmetic(t *testing.T) {
	r := tacticalRuntime{}
	projectile := checkpoint4EnemyProjectile()
	projectile.Multiplicity = 3
	projectile.InterceptDamageRemainder = 2

	first, err := applyTacticalProjectileInterception(&r, &projectile, 7, "beam_manual", 100, 0, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if first.DestroyedUnits != 2 || first.MultiplicityBefore != 3 || first.MultiplicityAfter != 1 ||
		first.RemainderBefore != 2 || first.RemainderAfter != 1 || first.Destroyed {
		t.Fatalf("first interception=%+v projectile=%+v", first, projectile)
	}
	if projectile.Multiplicity != 1 || projectile.InterceptDamageRemainder != 1 {
		t.Fatalf("first projectile state=%+v", projectile)
	}

	second, err := applyTacticalProjectileInterception(&r, &projectile, 3, "beam_manual", 100, 0, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if second.DestroyedUnits != 1 || second.MultiplicityBefore != 1 || second.MultiplicityAfter != 0 ||
		second.RemainderBefore != 1 || second.RemainderAfter != 0 || !second.Destroyed {
		t.Fatalf("second interception=%+v projectile=%+v", second, projectile)
	}
	events := checkpoint3Events(r.events, "projectile_interception_resolved")
	if len(events) != 2 {
		t.Fatalf("interception events=%+v", events)
	}
	data := checkpoint2EventData(t, events[0])
	if data["interceptor_source_kind"] != "beam_manual" || data["interceptor_ship_id"] != float64(100) ||
		data["interceptor_weapon_slot"] != float64(0) || data["damage"] != float64(7) ||
		data["destroyed_units"] != float64(2) || data["multiplicity_before"] != float64(3) ||
		data["multiplicity_after"] != float64(1) || data["remainder_before"] != float64(2) ||
		data["remainder_after"] != float64(1) {
		t.Fatalf("interception event=%+v", data)
	}
}

func TestSlice177Gate3Checkpoint4ManualBeamTargetProjectionAllowsStandardAndPDExcludesHeavy(t *testing.T) {
	spec := checkpoint4ProjectileTargetSpec()
	r, err := newTacticalRuntime(*spec.Tactical)
	if err != nil {
		t.Fatal(err)
	}
	r.ordnance.Projectiles = []TacticalProjectileState{checkpoint4EnemyProjectile()}
	r.ordnance.NextProjectileID = 2

	actions := tacticalWeaponActions(*spec.Tactical, &r)
	if len(actions) != 3 {
		t.Fatalf("Beam actions=%+v", actions)
	}
	for _, slot := range []int{0, 1} {
		var got *TacticalWeaponActionAvailability
		for i := range actions {
			if actions[i].WeaponSlot == slot {
				got = &actions[i]
				break
			}
		}
		if got == nil || !got.Available || !checkpoint4TargetHasProjectile(got.Targets, 1) {
			t.Fatalf("slot %d projectile target action=%+v", slot, got)
		}
	}
	var heavy TacticalWeaponActionAvailability
	for _, action := range actions {
		if action.WeaponSlot == 2 {
			heavy = action
		}
	}
	if !heavy.Available {
		t.Fatalf("Heavy ship-target Beam unexpectedly unavailable=%+v", heavy)
	}
	if checkpoint4TargetHasProjectile(heavy.Targets, 1) {
		t.Fatalf("Heavy Beam exposed projectile target=%+v", heavy)
	}

	legal := legalFireActions(actions)
	for _, action := range legal {
		if action.WeaponSlot == 2 && checkpoint4TargetHasProjectile(action.Targets, 1) {
			t.Fatalf("Heavy legal-fire action exposed projectile=%+v", action)
		}
	}
}

func TestSlice177Gate3Checkpoint4ManualBeamDestroysProjectileAndConsumesReadiness(t *testing.T) {
	spec := checkpoint4ProjectileTargetSpec()
	spec.Tactical.Ships[0].Weapons = spec.Tactical.Ships[0].Weapons[:1]
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	checkpoint4InjectProjectile(t, s, checkpoint4EnemyProjectile())

	before := s.View()
	if len(before.Tactical.LegalFireActions) != 1 || !checkpoint4TargetHasProjectile(before.Tactical.LegalFireActions[0].Targets, 1) {
		t.Fatalf("manual projectile action missing=%+v", before.Tactical.LegalFireActions)
	}

	cmd, err := NewFireBeamCommand(1, FireBeamPayload{ShipID: 100, TargetProjectileID: 1, WeaponSlot: 0})
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := s.PrepareCommand(1, cmd)
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Result() != nil {
		t.Fatalf("projectile interception unexpectedly ended battle=%+v", prepared.Result())
	}
	if err := s.CommitPreparedCommand(prepared); err != nil {
		t.Fatal(err)
	}
	after := s.View()
	if len(after.Tactical.Ordnance.Projectiles) != 0 {
		t.Fatalf("destroyed projectile survived=%+v", after.Tactical.Ordnance.Projectiles)
	}
	state := tacticalStateForTest(t, after.Tactical.State.Ships, 100)
	if state.Weapons[0].Ready {
		t.Fatal("manual projectile Beam did not consume readiness")
	}

	beamEvents := checkpoint3Events(after.Tactical.Events, "beam_fired")
	if len(beamEvents) != 1 {
		t.Fatalf("manual projectile beam events=%+v", beamEvents)
	}
	beamData := checkpoint2EventData(t, beamEvents[0])
	if beamData["target_ship_id"] != float64(0) || beamData["target_projectile_id"] != float64(1) ||
		beamData["hit"] != true || beamData["damage"] != float64(4) {
		t.Fatalf("manual projectile beam=%+v", beamData)
	}
	intercepts := checkpoint3Events(after.Tactical.Events, "projectile_interception_resolved")
	if len(intercepts) != 1 {
		t.Fatalf("manual interception events=%+v", intercepts)
	}
	intercept := checkpoint2EventData(t, intercepts[0])
	if intercept["interceptor_source_kind"] != "beam_manual" || intercept["destroyed_units"] != float64(1) ||
		intercept["multiplicity_after"] != float64(0) {
		t.Fatalf("manual interception=%+v", intercept)
	}
	destroyed := checkpoint3Events(after.Tactical.Events, "projectile_destroyed")
	if len(destroyed) != 1 || checkpoint2EventData(t, destroyed[0])["reason"] != "intercepted" {
		t.Fatalf("manual projectile destruction=%+v", destroyed)
	}
}

func TestSlice177Gate3Checkpoint4ManualHeavyProjectileTargetRejectedWithoutMutation(t *testing.T) {
	spec := checkpoint4ProjectileTargetSpec()
	spec.Tactical.Ships[0].Weapons = []TacticalWeaponSpec{spec.Tactical.Ships[0].Weapons[2]}
	spec.Tactical.Ships[0].Weapons[0].Slot = 0
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	checkpoint4InjectProjectile(t, s, checkpoint4EnemyProjectile())
	before := s.View()

	cmd, err := NewFireBeamCommand(1, FireBeamPayload{ShipID: 100, TargetProjectileID: 1, WeaponSlot: 0})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.PrepareCommand(1, cmd); err == nil {
		t.Fatal("Heavy Beam unexpectedly accepted projectile target")
	}
	after := s.View()
	if !reflect.DeepEqual(before.Tactical.State, after.Tactical.State) ||
		!reflect.DeepEqual(before.Tactical.Ordnance, after.Tactical.Ordnance) ||
		!reflect.DeepEqual(before.Tactical.Events, after.Tactical.Events) {
		t.Fatalf("rejected Heavy projectile fire mutated state")
	}
}

func TestSlice177Gate3Checkpoint4BeamPayloadRequiresExactlyOneTargetKind(t *testing.T) {
	spec := checkpoint4ProjectileTargetSpec()
	spec.Tactical.Ships[0].Weapons = spec.Tactical.Ships[0].Weapons[:1]
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	checkpoint4InjectProjectile(t, s, checkpoint4EnemyProjectile())
	before := s.View()

	for _, payload := range []FireBeamPayload{
		{ShipID: 100, WeaponSlot: 0},
		{ShipID: 100, TargetShipID: 200, TargetProjectileID: 1, WeaponSlot: 0},
	} {
		cmd, err := NewFireBeamCommand(1, payload)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := s.PrepareCommand(1, cmd); err == nil {
			t.Fatalf("ambiguous/empty Beam target unexpectedly accepted=%+v", payload)
		}
	}
	after := s.View()
	if !reflect.DeepEqual(before.Tactical.State, after.Tactical.State) ||
		!reflect.DeepEqual(before.Tactical.Ordnance, after.Tactical.Ordnance) ||
		!reflect.DeepEqual(before.Tactical.Events, after.Tactical.Events) {
		t.Fatalf("rejected Beam target payload mutated state")
	}
}

func TestSlice177Gate3Checkpoint4ReactivePDKillsMIRVCarrierBeforeSplitAndPreservesReadiness(t *testing.T) {
	spec := checkpoint1OrdnanceSpec()
	missile := &spec.Tactical.Ships[0].Weapons[0]
	missile.Count = 1
	missile.ModifierIDs = []core.ShipWeaponModifierID{core.ShipWeaponModifierMIRV}
	spec.Tactical.Ships[1].BeamOffense = 25
	spec.Tactical.Ships[1].Weapons = []TacticalWeaponSpec{{
		Slot: 0, WeaponID: "phasor", WeaponFamily: "beam", Count: 1,
		MountType: core.ShipWeaponMountPointDefense, FiringArcID: core.ShipWeaponFiringArc360,
		MinDamage: 5, MaxDamage: 20, TacticalSupport: "beam",
	}}

	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	s.mu.Lock()
	defender := s.tactical.shipState(200)
	defender.Weapons[0].Ready = false
	beforeRNG := s.tactical.state.RNGState
	s.mu.Unlock()

	fire, err := NewFireOrdnanceCommand(1, FireOrdnancePayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	if err != nil {
		t.Fatal(err)
	}
	checkpoint2Commit(t, s, 1, fire)
	view := s.View()
	if len(view.Tactical.Ordnance.Projectiles) != 0 {
		t.Fatalf("intercepted MIRV carrier survived=%+v", view.Tactical.Ordnance.Projectiles)
	}
	defenderState := tacticalStateForTest(t, view.Tactical.State.Ships, 200)
	if defenderState.Weapons[0].Ready {
		t.Fatal("reactive PD changed false normal readiness")
	}
	if view.Tactical.State.RNGState == beforeRNG {
		t.Fatal("reactive PD did not consume frozen Beam RNG")
	}
	if len(checkpoint3Events(view.Tactical.Events, "projectile_mirv_split")) != 0 {
		t.Fatalf("MIRV split occurred after carrier interception events=%+v", view.Tactical.Events)
	}
	if len(checkpoint3Events(view.Tactical.Events, "projectile_impact_resolved")) != 0 {
		t.Fatalf("impact occurred after carrier interception")
	}

	beamEvents := checkpoint3Events(view.Tactical.Events, "beam_fired")
	if len(beamEvents) != 1 {
		t.Fatalf("reactive Beam events=%+v", beamEvents)
	}
	for _, event := range beamEvents {
		data := checkpoint2EventData(t, event)
		if data["reactive_point_defense"] != true || data["target_ship_id"] != float64(0) {
			t.Fatalf("reactive Beam identity=%+v", data)
		}
	}
	intercepts := checkpoint3Events(view.Tactical.Events, "projectile_interception_resolved")
	if len(intercepts) != 1 {
		t.Fatalf("reactive interception events=%+v", intercepts)
	}
	first := checkpoint2EventData(t, intercepts[0])
	damage, ok := first["damage"].(float64)
	if !ok || damage < 4 || first["interceptor_source_kind"] != "beam_reactive" ||
		first["destroyed_units"] != float64(1) || first["multiplicity_after"] != float64(0) ||
		first["remainder_after"] != float64(0) {
		t.Fatalf("reactive interception=%+v", first)
	}
	destroyed := checkpoint3Events(view.Tactical.Events, "projectile_destroyed")
	if len(destroyed) != 1 || checkpoint2EventData(t, destroyed[0])["reason"] != "intercepted" {
		t.Fatalf("reactive projectile destruction=%+v", destroyed)
	}
}

func TestSlice177Gate3Checkpoint4ReactivePDSlotsAreAscendingAndStandardIsExcluded(t *testing.T) {
	spec := checkpoint1OrdnanceSpec()
	spec.Tactical.Ships[1].Weapons = []TacticalWeaponSpec{
		{Slot: 5, WeaponID: "laser_cannon", WeaponFamily: "beam", Count: 1, MountType: core.ShipWeaponMountPointDefense, FiringArcID: core.ShipWeaponFiringArc360, MinDamage: 1, MaxDamage: 4, TacticalSupport: "beam"},
		{Slot: 1, WeaponID: "laser_cannon", WeaponFamily: "beam", Count: 1, MountType: core.ShipWeaponMountStandard, FiringArcID: core.ShipWeaponFiringArc360, MinDamage: 1, MaxDamage: 4, TacticalSupport: "beam"},
		{Slot: 2, WeaponID: "laser_cannon", WeaponFamily: "beam", Count: 1, MountType: core.ShipWeaponMountPointDefense, FiringArcID: core.ShipWeaponFiringArc360, MinDamage: 1, MaxDamage: 4, TacticalSupport: "beam"},
	}
	r, err := newTacticalRuntime(*spec.Tactical)
	if err != nil {
		t.Fatal(err)
	}
	projectile := checkpoint4EnemyProjectile()
	projectile.SourceShipID = 100
	projectile.SourceSeatID = 1
	projectile.TargetShipID = 200
	projectile.X = r.shipState(200).X
	projectile.Y = r.shipState(200).Y
	if got := tacticalReactivePointDefenseSlots(*spec.Tactical, &r, &spec.Tactical.Ships[1], projectile); !reflect.DeepEqual(got, []int{2, 5}) {
		t.Fatalf("reactive PD slots=%v want [2 5]", got)
	}
}
