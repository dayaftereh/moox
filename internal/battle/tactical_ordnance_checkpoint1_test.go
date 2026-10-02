package battle

import (
	"encoding/json"
	"reflect"
	"testing"

	"moox/internal/core"
)

func checkpoint1OrdnanceRules() *TacticalOrdnanceRulesSnapshot {
	return &TacticalOrdnanceRulesSnapshot{
		MovementUnitsPerField:     20,
		FastSpeedBonus:            4,
		TorpedoCooldownRounds:     2,
		MIRVMultiplier:            4,
		PlasmaDissipationPerField: 5,
		Weapons:                   expectedTacticalOrdnanceWeaponRules(),
	}
}

func checkpoint1OrdnanceSpec() Spec {
	spec := baselineTacticalBattleSpec()
	spec.Tactical.Rules.Ordnance = checkpoint1OrdnanceRules()
	spec.Tactical.Ships[0].Weapons = []TacticalWeaponSpec{
		{Slot: 0, WeaponID: "nuclear_missile", WeaponFamily: "missile", Count: 2, Ammo: 5, FiringArcID: core.ShipWeaponFiringArc360, MinDamage: 8, MaxDamage: 8, TacticalSupport: "ordnance"},
		{Slot: 1, WeaponID: "anti_matter_torpedoes", WeaponFamily: "torpedo", Count: 3, FiringArcID: core.ShipWeaponFiringArc360, MinDamage: 25, MaxDamage: 25, TacticalSupport: "ordnance"},
	}
	return spec
}

func TestSlice177Gate3Checkpoint1InitializesDetachedOrdnanceProjection(t *testing.T) {
	s, err := NewSession(checkpoint1OrdnanceSpec())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	view := s.View()
	if view.Tactical == nil {
		t.Fatal("tactical view missing")
	}
	ordnance := view.Tactical.Ordnance
	if ordnance.NextProjectileID != 1 || len(ordnance.Projectiles) != 0 || len(ordnance.WeaponStates) != 2 {
		t.Fatalf("initial ordnance state=%+v", ordnance)
	}
	if got := ordnance.WeaponStates[0]; got.ShipID != 100 || got.WeaponSlot != 0 || got.AmmoRemaining != 5 || got.CooldownRounds != 0 {
		t.Fatalf("Missile state=%+v", got)
	}
	if got := ordnance.WeaponStates[1]; got.ShipID != 100 || got.WeaponSlot != 1 || got.AmmoRemaining != 0 || got.CooldownRounds != 0 {
		t.Fatalf("Torpedo state=%+v", got)
	}
	if len(view.Tactical.WeaponActions) != 2 {
		t.Fatalf("weapon actions=%+v", view.Tactical.WeaponActions)
	}
	for _, action := range view.Tactical.WeaponActions {
		if action.Available || action.CommandKind != "" || action.UnavailableReason != TacticalWeaponUnavailableConsumer {
			t.Fatalf("checkpoint 1 unexpectedly enabled ordnance action: %+v", action)
		}
	}
	if view.Spec.Tactical == nil || view.Spec.Tactical.Rules.Ordnance == nil || len(view.Spec.Tactical.Rules.Ordnance.Weapons) != 7 {
		t.Fatalf("ordnance rule snapshot=%+v", view.Spec.Tactical)
	}

	view.Tactical.Ordnance.WeaponStates[0].AmmoRemaining = 0
	view.Spec.Tactical.Rules.Ordnance.Weapons[0].BaseSpeed = 99
	fresh := s.View()
	if fresh.Tactical.Ordnance.WeaponStates[0].AmmoRemaining != 5 || fresh.Spec.Tactical.Rules.Ordnance.Weapons[0].BaseSpeed != 12 {
		t.Fatalf("projection aliases authority: state=%+v rules=%+v", fresh.Tactical.Ordnance, fresh.Spec.Tactical.Rules.Ordnance)
	}

	player, err := s.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	if player.Tactical == nil || player.Tactical.State.RNGState != 0 || player.Spec.Tactical.InitialRNGState != 0 {
		t.Fatalf("player RNG redaction changed: %+v", player.Tactical)
	}
	if !reflect.DeepEqual(player.Tactical.Ordnance.WeaponStates, fresh.Tactical.Ordnance.WeaponStates) {
		t.Fatalf("player ordnance projection differs: player=%+v observer=%+v", player.Tactical.Ordnance, fresh.Tactical.Ordnance)
	}
}

func TestSlice177Gate3Checkpoint1LiveSnapshotRoundTripAndLegacyMigration(t *testing.T) {
	s, err := NewSession(checkpoint1OrdnanceSpec())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	snapshot := s.LiveSnapshot()
	if snapshot.Tactical == nil || snapshot.Tactical.Ordnance == nil {
		t.Fatalf("new snapshot missing explicit ordnance state: %+v", snapshot.Tactical)
	}
	encoded, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var decoded LiveSnapshot
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreLiveSnapshot(decoded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(snapshot, restored.LiveSnapshot()) {
		t.Fatalf("ordnance snapshot changed across round trip\nbefore=%+v\nafter=%+v", snapshot.Tactical.Ordnance, restored.LiveSnapshot().Tactical.Ordnance)
	}

	missing := decoded
	missing.Tactical.Ordnance = nil
	if _, err := RestoreLiveSnapshot(missing); err == nil {
		t.Fatal("new executable-ordnance snapshot without ordnance state unexpectedly restored")
	}

	legacySpec := checkpoint4MixedWeaponSpec()
	legacy, err := NewSession(legacySpec)
	if err != nil {
		t.Fatal(err)
	}
	if err := legacy.Start(); err != nil {
		t.Fatal(err)
	}
	legacySnapshot := legacy.LiveSnapshot()
	legacySnapshot.Tactical.Ordnance = nil
	migrated, err := RestoreLiveSnapshot(legacySnapshot)
	if err != nil {
		t.Fatalf("legacy metadata-only snapshot failed migration: %v", err)
	}
	migratedSnapshot := migrated.LiveSnapshot()
	if migratedSnapshot.Tactical == nil || migratedSnapshot.Tactical.Ordnance == nil {
		t.Fatal("legacy migration did not materialize explicit ordnance state")
	}
	got := migratedSnapshot.Tactical.Ordnance
	if got.NextProjectileID != 1 || len(got.Projectiles) != 0 || len(got.WeaponStates) != 1 || got.WeaponStates[0].AmmoRemaining != 5 {
		t.Fatalf("legacy migrated ordnance state=%+v", got)
	}
}
func TestSlice177Gate3Checkpoint1RejectsSupportFamilyMismatch(t *testing.T) {
	badOrdnance := TacticalWeaponSpec{WeaponID: "nuclear_bomb", WeaponFamily: "bomb", Count: 1, MinDamage: 3, MaxDamage: 12, TacticalSupport: "ordnance"}
	if err := validateTacticalWeaponSnapshot(badOrdnance); err == nil {
		t.Fatal("Bomb with ordnance support unexpectedly validated")
	}
	badBombardment := TacticalWeaponSpec{WeaponID: "nuclear_missile", WeaponFamily: "missile", Count: 1, MinDamage: 8, MaxDamage: 8, TacticalSupport: "bombardment_only"}
	if err := validateTacticalWeaponSnapshot(badBombardment); err == nil {
		t.Fatal("Missile with bombardment-only support unexpectedly validated")
	}
}
