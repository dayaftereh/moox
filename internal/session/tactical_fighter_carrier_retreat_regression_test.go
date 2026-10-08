package session

import (
	"path/filepath"
	"testing"

	"moox/internal/battle"
	"moox/internal/game"
	"moox/internal/protocol"
)

func TestCombatReviewContinuesAfterFighterCarrierLoss(t *testing.T) {
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := rules.NewReferenceTriangleCombatReviewGame(1)
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	seats := []Seat{
		{ID: generated.Players[0].SeatID, EmpireID: generated.Players[0].EmpireID, Name: generated.Players[0].Name, Controller: ControllerLocalHuman},
		{ID: generated.Players[1].SeatID, EmpireID: generated.Players[1].EmpireID, Name: generated.Players[1].Name, Controller: ControllerBuiltinAI},
	}
	s, err := NewGameSession("debug-retreat", generated.State, seats)
	if err != nil {
		t.Fatal(err)
	}
	submitStubTurn(t, s, "debug-retreat")
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	obs, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if len(obs.Battles) != 1 || obs.Battles[0].Tactical == nil {
		t.Fatalf("battle=%+v", obs.Battles)
	}
	battleID := obs.Battles[0].Spec.ID
	human := seats[0].ID
	aiSeat := seats[1].ID
	targetID := obs.Battles[0].Spec.Defender.ShipIDs[0]

	build := func(sequence uint32, action battle.TacticalWeaponActionAvailability) (protocol.Command, bool) {
		for _, target := range action.Targets {
			if int64(target.TargetShipID) != int64(targetID) {
				continue
			}
			switch action.CommandKind {
			case battle.CommandFireBeam:
				cmd, e := battle.NewFireBeamCommand(sequence, battle.FireBeamPayload{ShipID: action.ShipID, TargetShipID: target.TargetShipID, WeaponSlot: action.WeaponSlot})
				if e != nil {
					t.Fatal(e)
				}
				return cmd, true
			case battle.CommandFireOrdnance:
				cmd, e := battle.NewFireOrdnanceCommand(sequence, battle.FireOrdnancePayload{ShipID: action.ShipID, TargetShipID: target.TargetShipID, WeaponSlot: action.WeaponSlot})
				if e != nil {
					t.Fatal(e)
				}
				return cmd, true
			case battle.CommandLaunchFighters:
				cmd, e := battle.NewLaunchFightersCommand(sequence, battle.LaunchFightersPayload{ShipID: action.ShipID, TargetShipID: target.TargetShipID, WeaponSlot: action.WeaponSlot})
				if e != nil {
					t.Fatal(e)
				}
				return cmd, true
			}
		}
		return protocol.Command{}, false
	}

	for steps := 0; steps < 100; steps++ {
		views, err := s.PlayerBattleViews(human)
		if err != nil {
			t.Fatal(err)
		}
		if len(views) != 1 || views[0].Tactical == nil {
			t.Fatalf("human views=%+v", views)
		}
		v := views[0]
		activeID := v.Tactical.State.ActiveShipID
		activeSeat := protocol.SeatID(0)
		for _, ship := range v.Tactical.Ships {
			if ship.ShipID == activeID {
				activeSeat = ship.SeatID
				break
			}
		}
		if activeSeat != human {
			break
		}
		fired := false
		for _, action := range v.Tactical.WeaponActions {
			if !action.Available {
				continue
			}
			if cmd, ok := build(v.Tactical.State.NextCommandSequence, action); ok {
				if err := s.SubmitBattleCommand(battleID, human, cmd); err != nil {
					t.Fatal(err)
				}
				fired = true
				break
			}
		}
		if fired {
			continue
		}
		if !v.Tactical.CanEndActivation {
			t.Fatalf("human active %d cannot end", activeID)
		}
		cmd, err := battle.NewEndActivationCommand(v.Tactical.State.NextCommandSequence, battle.EndActivationPayload{ShipID: activeID})
		if err != nil {
			t.Fatal(err)
		}
		if err := s.SubmitBattleCommand(battleID, human, cmd); err != nil {
			t.Fatal(err)
		}
	}

	handedBack := false
	for step := 0; step < 200; step++ {
		decisions := s.battleDecisionCatalogLocked(aiSeat)
		if len(decisions) != 1 || len(decisions[0].Actions) == 0 {
			t.Fatalf("AI lost all legal actions at step %d while its ships survive", step)
		}
		cmd := decisions[0].Actions[0]
		if cmd.Kind == battle.CommandRetreat {
			t.Fatalf("AI chose forced retreat at step %d while its ships survive", step)
		}
		if err := s.SubmitBattleCommand(battleID, aiSeat, cmd); err != nil {
			t.Fatalf("AI step %d: %v: %v", step, cmd.Kind, err)
		}
		observer, err := s.ObserverView()
		if err != nil {
			t.Fatal(err)
		}
		if len(observer.Battles) != 1 || observer.Battles[0].Phase != battle.PhaseActive || observer.Battles[0].Tactical == nil {
			t.Fatalf("AI unexpectedly ended tactical battle at step %d, command %s", step, cmd.Kind)
		}
		view := observer.Battles[0].Tactical
		activeSeat := protocol.SeatID(0)
		for _, ship := range view.Ships {
			if ship.ShipID == view.State.ActiveShipID {
				activeSeat = ship.SeatID
				break
			}
		}
		if activeSeat == human {
			handedBack = true
			break
		}
	}
	if !handedBack {
		t.Fatal("AI did not return a regular active Tactical turn to Human")
	}
}
