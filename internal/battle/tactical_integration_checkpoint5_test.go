package battle

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestSlice175Checkpoint5ObserverPlayerAndPersistenceBoundary(t *testing.T) {
	spec := checkpoint4MixedWeaponSpec()
	spec.Tactical.Ships[1].ShieldID = "class_iii_shield"
	spec.Tactical.Ships[1].ShieldStrength = 3
	spec.Tactical.Ships[1].ShieldArcMax = 15

	session, err := NewSession(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := session.Start(); err != nil {
		t.Fatal(err)
	}

	observer := session.View()
	if observer.Tactical == nil || len(observer.Tactical.WeaponActions) != 2 {
		t.Fatalf("observer weapon actions=%+v", observer.Tactical)
	}
	if observer.Tactical.WeaponActions[0].WeaponFamily != "beam" || !observer.Tactical.WeaponActions[0].Available || observer.Tactical.WeaponActions[0].CommandKind != CommandFireBeam {
		t.Fatalf("observer supported action=%+v", observer.Tactical.WeaponActions[0])
	}
	if observer.Tactical.WeaponActions[1].WeaponFamily != "missile" || observer.Tactical.WeaponActions[1].Available || observer.Tactical.WeaponActions[1].UnavailableReason != TacticalWeaponUnavailableConsumer {
		t.Fatalf("observer unsupported action=%+v", observer.Tactical.WeaponActions[1])
	}
	defender := checkpoint3ShipView(t, observer, 200)
	if defender.ShieldArcMax != 15 || defender.ShieldStrength != 3 || defender.ShieldArcCurrent != [4]int{15, 15, 15, 15} {
		t.Fatalf("observer shield foundation=%+v", defender)
	}

	activePlayer, err := session.PlayerView(1)
	if err != nil {
		t.Fatal(err)
	}
	if activePlayer.Tactical == nil || len(activePlayer.Tactical.WeaponActions) != 2 {
		t.Fatalf("active player weapon actions=%+v", activePlayer.Tactical)
	}
	nonActivePlayer, err := session.PlayerView(2)
	if err != nil {
		t.Fatal(err)
	}
	if nonActivePlayer.Tactical == nil || len(nonActivePlayer.Tactical.WeaponActions) != 0 {
		t.Fatalf("non-active player leaked weapon command authority=%+v", nonActivePlayer.Tactical)
	}

	encoded, err := json.Marshal(session.LiveSnapshot())
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

	restoredObserver := restored.View()
	if restoredObserver.Tactical == nil {
		t.Fatal("restored Tactical view missing")
	}
	if !reflect.DeepEqual(observer.Tactical.WeaponActions, restoredObserver.Tactical.WeaponActions) {
		t.Fatalf("weapon actions changed across JSON persistence boundary:\nbefore=%+v\nafter=%+v", observer.Tactical.WeaponActions, restoredObserver.Tactical.WeaponActions)
	}
	if !reflect.DeepEqual(observer.Tactical.State.Ships, restoredObserver.Tactical.State.Ships) {
		t.Fatalf("Tactical state changed across JSON persistence boundary:\nbefore=%+v\nafter=%+v", observer.Tactical.State.Ships, restoredObserver.Tactical.State.Ships)
	}
	if restoredObserver.Spec.Tactical == nil || restoredObserver.Spec.Tactical.Ships[0].Weapons[1].WeaponFamily != "missile" {
		t.Fatalf("weapon-family foundation missing after persistence: %+v", restoredObserver.Spec.Tactical)
	}

	restoredObserver.Tactical.WeaponActions[0].Available = false
	restoredObserver.Tactical.WeaponActions[0].Targets[0].TargetShipID = 999
	fresh := restored.View()
	if fresh.Tactical == nil || !fresh.Tactical.WeaponActions[0].Available || fresh.Tactical.WeaponActions[0].Targets[0].TargetShipID != 200 {
		t.Fatalf("mutating restored observer projection aliased authority: %+v", fresh.Tactical)
	}
}
