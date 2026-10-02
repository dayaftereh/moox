package battle

import (
	"testing"

	"moox/internal/core"
)

func checkpoint3Events(events []TacticalEvent, kind string) []TacticalEvent {
	out := make([]TacticalEvent, 0)
	for _, event := range events {
		if event.Kind == kind {
			out = append(out, event)
		}
	}
	return out
}

func checkpoint3ProjectileForWeapon(t *testing.T, spec TacticalSpec, weapon TacticalWeaponSpec, target TacticalShipState) TacticalProjectileState {
	t.Helper()
	configured, err := deriveConfiguredTacticalProjectile(spec, weapon)
	if err != nil {
		t.Fatal(err)
	}
	return TacticalProjectileState{
		ProjectileID:      1,
		SourceSeatID:      spec.Ships[0].SeatID,
		SourceShipID:      spec.Ships[0].ShipID,
		WeaponSlot:        weapon.Slot,
		TargetShipID:      target.ShipID,
		WeaponID:          weapon.WeaponID,
		WeaponFamily:      tacticalWeaponFamily(weapon),
		ModifierIDs:       append([]core.ShipWeaponModifierID(nil), weapon.ModifierIDs...),
		X:                 target.X,
		Y:                 target.Y,
		Facing:            0,
		Speed:             configured.Speed,
		MovementRemaining: 0,
		Multiplicity:      weapon.Count,
		UnitDurability:    configured.UnitDurability,
		WarheadStrength:   configured.WarheadStrength,
	}
}

func TestSlice177Gate3Checkpoint3ConfiguredProjectileReferenceVectors(t *testing.T) {
	spec := checkpoint1OrdnanceSpec()

	zeon := spec.Tactical.Ships[0].Weapons[0]
	zeon.WeaponID = "zeon_missile"
	zeon.MinDamage = 30
	zeon.MaxDamage = 30
	zeon.ModifierIDs = []core.ShipWeaponModifierID{
		core.ShipWeaponModifierECCM,
		core.ShipWeaponModifierArmored,
		core.ShipWeaponModifierFast,
	}
	configured, err := deriveConfiguredTacticalProjectile(*spec.Tactical, zeon)
	if err != nil {
		t.Fatal(err)
	}
	if configured.Speed != 16 || configured.UnitDurability != 32 || configured.BeamInterceptorDCV != 130 || configured.WarheadStrength != 30 || !configured.ECCM {
		t.Fatalf("Zeon Fast/Armored/ECCM=%+v", configured)
	}

	plasma := spec.Tactical.Ships[0].Weapons[1]
	plasma.WeaponID = "plasma_torpedoes"
	plasma.MinDamage = 120
	plasma.MaxDamage = 120
	plasma.ModifierIDs = []core.ShipWeaponModifierID{
		core.ShipWeaponModifierNoRangeDissipation,
		core.ShipWeaponModifierEnveloping,
		core.ShipWeaponModifierECCM,
		core.ShipWeaponModifierOverloaded,
	}
	configured, err = deriveConfiguredTacticalProjectile(*spec.Tactical, plasma)
	if err != nil {
		t.Fatal(err)
	}
	if configured.Speed != 24 || configured.UnitDurability != 5 || configured.BeamInterceptorDCV != 120 || configured.WarheadStrength != 180 ||
		!configured.NoRangeDissipation || !configured.Enveloping || !configured.ECCM {
		t.Fatalf("Plasma configured=%+v", configured)
	}
}

func TestSlice177Gate3Checkpoint3ECCMImpactDefenseAndRNGOrder(t *testing.T) {
	spec := checkpoint1OrdnanceSpec()
	weapon := &spec.Tactical.Ships[0].Weapons[0]
	weapon.WeaponID = "zeon_missile"
	weapon.MinDamage = 30
	weapon.MaxDamage = 30
	weapon.ModifierIDs = []core.ShipWeaponModifierID{
		core.ShipWeaponModifierECCM,
		core.ShipWeaponModifierArmored,
		core.ShipWeaponModifierFast,
	}
	spec.Tactical.Ships[0].MissileScannerReduction = 20
	spec.Tactical.Ships[1].MissileDefense = 80
	spec.Tactical.Ships[1].ArmorMax = 1000
	spec.Tactical.Ships[1].StructureMax = 1000
	spec.Tactical.Ships[1].ShieldStrength = 0
	spec.Tactical.Ships[1].ShieldArcMax = 0

	r, err := newTacticalRuntime(*spec.Tactical)
	if err != nil {
		t.Fatal(err)
	}
	target := *r.shipState(200)
	projectile := checkpoint3ProjectileForWeapon(t, *spec.Tactical, *weapon, target)
	r.ordnance.Projectiles = []TacticalProjectileState{projectile}
	r.ordnance.NextProjectileID = 2

	if err := processTacticalProjectiles(*spec.Tactical, &r, 0, 0); err != nil {
		t.Fatal(err)
	}
	if len(r.ordnance.Projectiles) != 0 {
		t.Fatalf("impact-complete projectile survived=%+v", r.ordnance.Projectiles)
	}
	impacts := checkpoint3Events(r.events, "projectile_impact_resolved")
	if len(impacts) != 2 {
		t.Fatalf("impact events=%+v", impacts)
	}
	first := checkpoint2EventData(t, impacts[0])
	second := checkpoint2EventData(t, impacts[1])
	if first["effective_missile_defense"] != float64(30) || first["raw_hit_roll"] != float64(100) || first["hit"] != true || first["damage_applied"] != float64(30) {
		t.Fatalf("first ECCM impact=%+v", first)
	}
	if second["effective_missile_defense"] != float64(30) || second["raw_hit_roll"] != float64(19) || second["hit"] != false || second["damage_applied"] != float64(0) {
		t.Fatalf("second ECCM impact=%+v", second)
	}
	if r.state.RNGState != 814028291 {
		t.Fatalf("RNG state=%d want=814028291", r.state.RNGState)
	}
	if got := r.shipState(200).ArmorCurrent; got != 970 {
		t.Fatalf("target armor=%d want=970", got)
	}
	damageEvents := checkpoint3Events(r.events, "battle_damage_applied")
	if len(damageEvents) != 1 {
		t.Fatalf("battle damage events=%+v", damageEvents)
	}
	damage := checkpoint2EventData(t, damageEvents[0])
	if damage["projectile_id"] != float64(1) || damage["impact_index"] != float64(1) || damage["weapon_id"] != "zeon_missile" {
		t.Fatalf("projectile damage event=%+v", damage)
	}
	destroyed := checkpoint3Events(r.events, "projectile_destroyed")
	if len(destroyed) != 1 || checkpoint2EventData(t, destroyed[0])["reason"] != "impact_complete" {
		t.Fatalf("projectile destruction=%+v", destroyed)
	}
}

func TestSlice177Gate3Checkpoint3MIRVSplitPrecedesExpandedImpactAttempts(t *testing.T) {
	spec := checkpoint1OrdnanceSpec()
	weapon := &spec.Tactical.Ships[0].Weapons[0]
	weapon.ModifierIDs = []core.ShipWeaponModifierID{core.ShipWeaponModifierMIRV}
	spec.Tactical.Ships[1].MissileDefense = 100
	spec.Tactical.Ships[1].ArmorMax = 1000
	spec.Tactical.Ships[1].StructureMax = 1000

	r, err := newTacticalRuntime(*spec.Tactical)
	if err != nil {
		t.Fatal(err)
	}
	projectile := checkpoint3ProjectileForWeapon(t, *spec.Tactical, *weapon, *r.shipState(200))
	r.ordnance.Projectiles = []TacticalProjectileState{projectile}
	r.ordnance.NextProjectileID = 2

	if err := processTacticalProjectiles(*spec.Tactical, &r, 0, 0); err != nil {
		t.Fatal(err)
	}
	splits := checkpoint3Events(r.events, "projectile_mirv_split")
	if len(splits) != 1 {
		t.Fatalf("MIRV split events=%+v", splits)
	}
	split := checkpoint2EventData(t, splits[0])
	if split["multiplicity_before"] != float64(2) || split["multiplicity_after"] != float64(8) {
		t.Fatalf("MIRV split=%+v", split)
	}
	impacts := checkpoint3Events(r.events, "projectile_impact_resolved")
	if len(impacts) != 8 {
		t.Fatalf("MIRV impact count=%d want=8", len(impacts))
	}
	if splits[0].Sequence >= impacts[0].Sequence {
		t.Fatalf("MIRV split sequence=%d first impact=%d", splits[0].Sequence, impacts[0].Sequence)
	}
	for i, event := range impacts {
		data := checkpoint2EventData(t, event)
		if data["impact_index"] != float64(i+1) || data["impact_count"] != float64(8) {
			t.Fatalf("MIRV impact[%d]=%+v", i, data)
		}
	}
}

func TestSlice177Gate3Checkpoint3PlasmaDissipationAndNRDOverload(t *testing.T) {
	t.Run("dissipated before RNG", func(t *testing.T) {
		spec := checkpoint1OrdnanceSpec()
		weapon := &spec.Tactical.Ships[0].Weapons[1]
		weapon.WeaponID = "plasma_torpedoes"
		weapon.Count = 1
		weapon.MinDamage = 120
		weapon.MaxDamage = 120
		weapon.ModifierIDs = nil
		spec.Tactical.Ships[1].ArmorMax = 1000
		spec.Tactical.Ships[1].StructureMax = 1000

		r, err := newTacticalRuntime(*spec.Tactical)
		if err != nil {
			t.Fatal(err)
		}
		projectile := checkpoint3ProjectileForWeapon(t, *spec.Tactical, *weapon, *r.shipState(200))
		projectile.DistanceTraveled = 24 * spec.Tactical.Rules.Ordnance.MovementUnitsPerField
		r.ordnance.Projectiles = []TacticalProjectileState{projectile}
		r.ordnance.NextProjectileID = 2
		beforeRNG := r.state.RNGState

		if err := processTacticalProjectiles(*spec.Tactical, &r, 0, 0); err != nil {
			t.Fatal(err)
		}
		if r.state.RNGState != beforeRNG || len(checkpoint3Events(r.events, "projectile_impact_resolved")) != 0 {
			t.Fatalf("dissipated Plasma consumed impact RNG/events")
		}
		destroyed := checkpoint3Events(r.events, "projectile_destroyed")
		if len(destroyed) != 1 || checkpoint2EventData(t, destroyed[0])["reason"] != "dissipated" {
			t.Fatalf("Plasma dissipated destruction=%+v", destroyed)
		}
	})

	t.Run("NRD preserves overloaded warhead", func(t *testing.T) {
		spec := checkpoint1OrdnanceSpec()
		weapon := &spec.Tactical.Ships[0].Weapons[1]
		weapon.WeaponID = "plasma_torpedoes"
		weapon.Count = 1
		weapon.MinDamage = 120
		weapon.MaxDamage = 120
		weapon.ModifierIDs = []core.ShipWeaponModifierID{
			core.ShipWeaponModifierNoRangeDissipation,
			core.ShipWeaponModifierOverloaded,
		}
		spec.Tactical.Ships[1].ArmorMax = 1000
		spec.Tactical.Ships[1].StructureMax = 1000
		spec.Tactical.Ships[1].MissileDefense = 0

		r, err := newTacticalRuntime(*spec.Tactical)
		if err != nil {
			t.Fatal(err)
		}
		projectile := checkpoint3ProjectileForWeapon(t, *spec.Tactical, *weapon, *r.shipState(200))
		projectile.DistanceTraveled = 24 * spec.Tactical.Rules.Ordnance.MovementUnitsPerField
		r.ordnance.Projectiles = []TacticalProjectileState{projectile}
		r.ordnance.NextProjectileID = 2

		if err := processTacticalProjectiles(*spec.Tactical, &r, 0, 0); err != nil {
			t.Fatal(err)
		}
		impacts := checkpoint3Events(r.events, "projectile_impact_resolved")
		if len(impacts) != 1 {
			t.Fatalf("NRD Plasma impacts=%+v", impacts)
		}
		data := checkpoint2EventData(t, impacts[0])
		if data["warhead_strength"] != float64(180) || data["raw_hit_roll"] != float64(100) || data["hit"] != true || data["damage_applied"] != float64(180) {
			t.Fatalf("NRD/Overloaded Plasma impact=%+v", data)
		}
		if got := r.shipState(200).ArmorCurrent; got != 820 {
			t.Fatalf("NRD/Overloaded Plasma armor=%d want=820", got)
		}
	})
}

func TestSlice177Gate3Checkpoint3EnvelopingUsesStableFourArcApplications(t *testing.T) {
	spec := checkpoint1OrdnanceSpec()
	weapon := &spec.Tactical.Ships[0].Weapons[1]
	weapon.Count = 1
	weapon.ModifierIDs = []core.ShipWeaponModifierID{core.ShipWeaponModifierEnveloping}
	spec.Tactical.Ships[1].MissileDefense = 0
	spec.Tactical.Ships[1].ArmorMax = 200
	spec.Tactical.Ships[1].StructureMax = 200
	spec.Tactical.Ships[1].ShieldStrength = 0
	spec.Tactical.Ships[1].ShieldArcMax = 10

	r, err := newTacticalRuntime(*spec.Tactical)
	if err != nil {
		t.Fatal(err)
	}
	targetState := r.shipState(200)
	targetState.ShieldArcCurrent = [4]int{10, 10, 10, 10}
	projectile := checkpoint3ProjectileForWeapon(t, *spec.Tactical, *weapon, *targetState)
	r.ordnance.Projectiles = []TacticalProjectileState{projectile}
	r.ordnance.NextProjectileID = 2

	if err := processTacticalProjectiles(*spec.Tactical, &r, 0, 0); err != nil {
		t.Fatal(err)
	}
	impacts := checkpoint3Events(r.events, "projectile_impact_resolved")
	if len(impacts) != 1 {
		t.Fatalf("Enveloping impacts=%+v", impacts)
	}
	impact := checkpoint2EventData(t, impacts[0])
	if impact["damage_applied"] != float64(100) || impact["warhead_strength"] != float64(25) {
		t.Fatalf("Enveloping impact=%+v", impact)
	}
	damageEvents := checkpoint3Events(r.events, "battle_damage_applied")
	if len(damageEvents) != 4 {
		t.Fatalf("Enveloping damage events=%+v", damageEvents)
	}
	for i, event := range damageEvents {
		data := checkpoint2EventData(t, event)
		if data["shield_arc"] != float64(i) || data["application_index"] != float64(i+1) || data["application_count"] != float64(4) {
			t.Fatalf("Enveloping damage[%d]=%+v", i, data)
		}
	}
	if got := r.shipState(200).ShieldArcCurrent; got != [4]int{0, 0, 0, 0} {
		t.Fatalf("Enveloping shield arcs=%v", got)
	}
	if got := r.shipState(200).ArmorCurrent; got != 140 {
		t.Fatalf("Enveloping armor=%d want=140", got)
	}
}

func TestSlice177Gate3Checkpoint3ImmediateImpactCompletesBattleSession(t *testing.T) {
	spec := checkpoint1OrdnanceSpec()
	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	command, err := NewFireOrdnanceCommand(1, FireOrdnancePayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := s.PrepareCommand(1, command)
	if err != nil {
		t.Fatal(err)
	}
	result := prepared.Result()
	if result == nil || result.WinnerSeat != 1 || result.Outcome != TacticalOutcomeVictory || len(result.DestroyedShipIDs) != 1 || result.DestroyedShipIDs[0] != 200 {
		t.Fatalf("immediate impact result=%+v", result)
	}
	if err := s.CommitPreparedCommand(prepared); err != nil {
		t.Fatal(err)
	}
	view := s.View()
	if view.Phase != PhaseCompleted || view.Result == nil || view.Result.WinnerSeat != 1 {
		t.Fatalf("completed impact battle=%+v", view.Result)
	}
	if len(view.Tactical.Ordnance.Projectiles) != 0 {
		t.Fatalf("impact-complete projectile persisted=%+v", view.Tactical.Ordnance.Projectiles)
	}
	if len(checkpoint3Events(view.Tactical.Events, "ship_destroyed")) != 1 || len(checkpoint3Events(view.Tactical.Events, "winner_determined")) != 1 {
		t.Fatalf("terminal ordnance events=%+v", view.Tactical.Events)
	}
}

func TestSlice177Gate3Checkpoint3RoundTransitionImpactPropagatesTerminalResult(t *testing.T) {
	spec := checkpoint1OrdnanceSpec()
	spec.Tactical.Ships[1].X = 23

	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	fire, _ := NewFireOrdnanceCommand(1, FireOrdnancePayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	checkpoint2Commit(t, s, 1, fire)
	if got := s.View().Tactical.Ordnance.Projectiles; len(got) != 1 || got[0].X != 22 {
		t.Fatalf("pre-round projectile=%+v", got)
	}

	endA, _ := NewEndActivationCommand(2, EndActivationPayload{ShipID: 100})
	checkpoint2Commit(t, s, 1, endA)
	endD, _ := NewEndActivationCommand(3, EndActivationPayload{ShipID: 200})
	prepared, err := s.PrepareCommand(2, endD)
	if err != nil {
		t.Fatal(err)
	}
	result := prepared.Result()
	if result == nil || result.WinnerSeat != 1 || result.Outcome != TacticalOutcomeVictory {
		t.Fatalf("round-transition impact result=%+v", result)
	}
	if err := s.CommitPreparedCommand(prepared); err != nil {
		t.Fatal(err)
	}
	if view := s.View(); view.Phase != PhaseCompleted || view.Result == nil || view.Result.WinnerSeat != 1 {
		t.Fatalf("round-transition completed battle=%+v", view.Result)
	}
}

func TestSlice177Gate3Checkpoint3LaunchUsesConfiguredProjectileModifiers(t *testing.T) {
	spec := checkpoint1OrdnanceSpec()
	spec.Tactical.Ships[1].X = 100

	missile := &spec.Tactical.Ships[0].Weapons[0]
	missile.WeaponID = "zeon_missile"
	missile.MinDamage = 30
	missile.MaxDamage = 30
	missile.ModifierIDs = []core.ShipWeaponModifierID{
		core.ShipWeaponModifierECCM,
		core.ShipWeaponModifierArmored,
		core.ShipWeaponModifierFast,
	}

	torpedo := &spec.Tactical.Ships[0].Weapons[1]
	torpedo.WeaponID = "plasma_torpedoes"
	torpedo.Count = 1
	torpedo.MinDamage = 120
	torpedo.MaxDamage = 120
	torpedo.ModifierIDs = []core.ShipWeaponModifierID{
		core.ShipWeaponModifierNoRangeDissipation,
		core.ShipWeaponModifierOverloaded,
	}

	s, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}

	fireMissile, err := NewFireOrdnanceCommand(1, FireOrdnancePayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 0})
	if err != nil {
		t.Fatal(err)
	}
	checkpoint2Commit(t, s, 1, fireMissile)
	afterMissile := s.View()
	if len(afterMissile.Tactical.Ordnance.Projectiles) != 1 {
		t.Fatalf("Zeon projectile state=%+v", afterMissile.Tactical.Ordnance.Projectiles)
	}
	gotMissile := afterMissile.Tactical.Ordnance.Projectiles[0]
	if gotMissile.Speed != 16 || gotMissile.UnitDurability != 32 || gotMissile.WarheadStrength != 30 ||
		gotMissile.X != 26 || gotMissile.DistanceTraveled != 320 {
		t.Fatalf("configured Zeon launch=%+v", gotMissile)
	}
	launches := checkpoint3Events(afterMissile.Tactical.Events, "ordnance_launched")
	if len(launches) != 1 {
		t.Fatalf("Zeon launch events=%+v", launches)
	}
	launch := checkpoint2EventData(t, launches[0])
	if launch["speed"] != float64(16) || launch["unit_durability"] != float64(32) ||
		launch["beam_interceptor_dcv"] != float64(130) || launch["warhead_strength"] != float64(30) {
		t.Fatalf("configured Zeon launch event=%+v", launch)
	}

	fireTorpedo, err := NewFireOrdnanceCommand(afterMissile.Tactical.State.NextCommandSequence, FireOrdnancePayload{ShipID: 100, TargetShipID: 200, WeaponSlot: 1})
	if err != nil {
		t.Fatal(err)
	}
	checkpoint2Commit(t, s, 1, fireTorpedo)
	afterTorpedo := s.View()
	if len(afterTorpedo.Tactical.Ordnance.Projectiles) != 2 {
		t.Fatalf("configured projectile state=%+v", afterTorpedo.Tactical.Ordnance.Projectiles)
	}
	gotTorpedo := afterTorpedo.Tactical.Ordnance.Projectiles[1]
	if gotTorpedo.Speed != 24 || gotTorpedo.UnitDurability != 5 || gotTorpedo.WarheadStrength != 180 ||
		gotTorpedo.X != 34 || gotTorpedo.DistanceTraveled != 480 {
		t.Fatalf("configured Plasma launch=%+v", gotTorpedo)
	}
	launches = checkpoint3Events(afterTorpedo.Tactical.Events, "ordnance_launched")
	if len(launches) != 2 {
		t.Fatalf("configured launch events=%+v", launches)
	}
	launch = checkpoint2EventData(t, launches[1])
	if launch["speed"] != float64(24) || launch["unit_durability"] != float64(5) ||
		launch["beam_interceptor_dcv"] != float64(120) || launch["warhead_strength"] != float64(180) {
		t.Fatalf("configured Plasma launch event=%+v", launch)
	}
}
