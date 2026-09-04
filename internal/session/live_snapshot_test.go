package session

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"moox/internal/battle"
	"moox/internal/game"
	"moox/internal/protocol"
)

func liveSnapshotRules(t *testing.T) *game.EconomyRules {
	t.Helper()
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	return rules
}

func emptyLiveBatch(gameID string, seatID protocol.SeatID, turn, revision uint64) protocol.CommandBatch {
	return protocol.CommandBatch{
		SchemaVersion: protocol.CommandSchemaVersion,
		GameID:        gameID,
		SeatID:        seatID,
		Turn:          turn,
		BaseRevision:  revision,
		Commands:      []protocol.Command{},
	}
}

func TestLiveSnapshotPlanningPartialSubmissionRoundTrip(t *testing.T) {
	state, seats := twoSeatFixture(t)
	s, err := NewGameSession("live-planning", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.PublishDraftTelemetry(1, "cursor", "draft", map[string]any{"x": 7}); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(emptyLiveBatch("live-planning", 2, 1, 1)); err != nil {
		t.Fatal(err)
	}
	rules := liveSnapshotRules(t)
	encoded, err := s.MarshalLiveSnapshot(rules)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := UnmarshalLiveSnapshot(encoded, rules, nil)
	if err != nil {
		t.Fatal(err)
	}
	reencoded, err := restored.MarshalLiveSnapshot(rules)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, reencoded) {
		t.Fatalf("planning live snapshot changed on roundtrip\nbefore=%s\nafter =%s", encoded, reencoded)
	}
	player, err := restored.PlayerView(2)
	if err != nil {
		t.Fatal(err)
	}
	if !player.Seat.Submitted || player.Seat.Seat.Controller != ControllerBuiltinAI || player.OwnSubmission == nil {
		t.Fatalf("restored seat/submission=%+v submission=%+v", player.Seat, player.OwnSubmission)
	}

	batch := emptyLiveBatch("live-planning", 1, 1, 1)
	if err := s.SubmitTurn(batch); err != nil {
		t.Fatal(err)
	}
	if err := restored.SubmitTurn(batch); err != nil {
		t.Fatal(err)
	}
	left, _ := s.ObserverView()
	right, _ := restored.ObserverView()
	leftBytes, _ := json.Marshal(left)
	rightBytes, _ := json.Marshal(right)
	if !bytes.Equal(leftBytes, rightBytes) {
		t.Fatalf("continued planning diverged\nleft =%s\nright=%s", leftBytes, rightBytes)
	}
}

func TestLiveSnapshotStrictCompatibilityAndBoundaryRejection(t *testing.T) {
	state, seats := twoSeatFixture(t)
	s, err := NewGameSession("live-invalid", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	rules := liveSnapshotRules(t)
	encoded, err := s.MarshalLiveSnapshot(rules)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := UnmarshalLiveSnapshot(append(append([]byte(nil), encoded...), []byte(" {}")...), rules, nil); err == nil {
		t.Fatal("trailing JSON document unexpectedly restored")
	}

	other := *rules
	other.RulesetSHA256 = "deadbeef"
	if _, err := UnmarshalLiveSnapshot(encoded, &other, nil); !errors.Is(err, ErrLiveSnapshotRulesMismatch) {
		t.Fatalf("rules mismatch error=%v", err)
	}

	externalSeats := append([]Seat(nil), seats...)
	externalSeats[0].Controller = ControllerExternalAI
	external, err := NewGameSession("live-external", state, externalSeats)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := external.MarshalLiveSnapshot(rules); err == nil {
		t.Fatal("ExternalAI live save unexpectedly allowed")
	}

	if err := s.SubmitTurn(emptyLiveBatch("live-invalid", 1, 1, 1)); err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitTurn(emptyLiveBatch("live-invalid", 2, 1, 1)); err != nil {
		t.Fatal(err)
	}
	if s.Status().Phase != PhaseStrategicResolution {
		t.Fatalf("phase=%q", s.Status().Phase)
	}
	if _, err := s.MarshalLiveSnapshot(rules); err == nil {
		t.Fatal("strategic_resolution live save unexpectedly allowed")
	}
}

func TestLiveSnapshotInvasionDecisionRoundTripAndContinuation(t *testing.T) {
	s, resolver, _, _, colonyID, _ := newInvasionSession(t, 0xB210, "live-invasion", 1)
	submitEmptyInvasionTurn(t, s)
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	if s.Status().Phase != PhaseInvasionDecisions {
		t.Fatalf("phase=%q", s.Status().Phase)
	}
	encoded, err := s.MarshalLiveSnapshot(resolver.Rules)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := UnmarshalLiveSnapshot(encoded, resolver.Rules, resolver)
	if err != nil {
		t.Fatal(err)
	}
	reencoded, err := restored.MarshalLiveSnapshot(resolver.Rules)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, reencoded) {
		t.Fatal("invasion live snapshot changed on roundtrip")
	}
	command, err := game.NewDeclineInvasionCommand(1, game.DeclineInvasionPayload{ColonyID: colonyID})
	if err != nil {
		t.Fatal(err)
	}
	revision := s.Status().Revision
	if err := s.ResolveInvasionCommand(1, revision, command); err != nil {
		t.Fatal(err)
	}
	if err := restored.ResolveInvasionCommand(1, revision, command); err != nil {
		t.Fatal(err)
	}
	if err := s.CompleteTurn(); err != nil {
		t.Fatal(err)
	}
	if err := restored.CompleteTurn(); err != nil {
		t.Fatal(err)
	}
	left, err := s.MarshalLiveSnapshot(resolver.Rules)
	if err != nil {
		t.Fatal(err)
	}
	right, err := restored.MarshalLiveSnapshot(resolver.Rules)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(left, right) {
		t.Fatal("invasion save/load continuation diverged")
	}
}

func TestLiveSnapshotInteractivePostResolutionRoundTripAndContinuation(t *testing.T) {
	state, seats := twoSeatFixture(t)
	rules := loadSessionColonyBaseRules(t)
	rules.MineralIndustryPerWorker["abundant"] = 600
	state.Empires[0].KnownTechnologyIDs = []int{game.ColonyBaseTechnologyID}
	sourceID := state.Colonies[0].ID
	targetID := addSessionSameSystemPlanet(state)
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession("live-colony-base", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	queue, err := game.NewQueueBuildingCommand(1, game.QueueBuildingPayload{ColonyID: sourceID, BuildingID: game.ColonyBaseBuildingID})
	if err != nil {
		t.Fatal(err)
	}
	afterBuild := submitColonyShipSessionTurn(t, s, resolver, []protocol.Command{queue})
	if len(afterBuild.ColonyBaseResolutions) != 1 || s.Status().Phase != PhasePostResolution {
		t.Fatalf("post-resolution Colony Base boundary=%+v phase=%q", afterBuild.ColonyBaseResolutions, s.Status().Phase)
	}
	encoded, err := s.MarshalLiveSnapshot(rules)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := UnmarshalLiveSnapshot(encoded, rules, nil)
	if err != nil {
		t.Fatal(err)
	}
	reencoded, err := restored.MarshalLiveSnapshot(rules)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, reencoded) {
		t.Fatal("post-resolution live snapshot changed on roundtrip")
	}
	colonize, err := game.NewColonizeWithBaseCommand(1, game.ColonizeWithBasePayload{SourceColonyID: sourceID, PlanetID: targetID})
	if err != nil {
		t.Fatal(err)
	}
	revision := s.Status().Revision
	if err := s.ResolveColonyBaseCommand(1, revision, colonize, resolver); err != nil {
		t.Fatal(err)
	}
	if err := restored.ResolveColonyBaseCommand(1, revision, colonize, resolver); err != nil {
		t.Fatal(err)
	}
	if err := s.CompleteTurn(); err != nil {
		t.Fatal(err)
	}
	if err := restored.CompleteTurn(); err != nil {
		t.Fatal(err)
	}
	left, err := s.MarshalLiveSnapshot(rules)
	if err != nil {
		t.Fatal(err)
	}
	right, err := restored.MarshalLiveSnapshot(rules)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(left, right) {
		t.Fatal("post-resolution save/load continuation diverged")
	}
}

func TestLiveSnapshotActiveTacticalSessionRoundTripAndContinuation(t *testing.T) {
	state, seats, fixture := makeTacticalStrategicFixture(t, 1)
	resolver := loadTacticalEconomyResolver(t)
	s, err := NewGameSession("live-tactical-session", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	submitStubTurn(t, s, "live-tactical-session")
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	started, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if started.Phase != PhaseEncounters || len(started.Battles) != 1 || started.Battles[0].Spec.Tactical == nil {
		t.Fatalf("tactical boundary=%+v", started.Battles)
	}
	child := started.Battles[0]
	fire1, err := battle.NewFireBeamCommand(1, battle.FireBeamPayload{ShipID: fixture.attackerShipIDs[0], TargetShipID: fixture.defenderShipIDs[0], WeaponSlot: 0})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SubmitBattleCommand(child.Spec.ID, child.Spec.Attacker.SeatID, fire1); err != nil {
		t.Fatal(err)
	}
	encoded, err := s.MarshalLiveSnapshot(resolver.Rules)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := UnmarshalLiveSnapshot(encoded, resolver.Rules, resolver)
	if err != nil {
		t.Fatal(err)
	}
	reencoded, err := restored.MarshalLiveSnapshot(resolver.Rules)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, reencoded) {
		t.Fatal("active tactical GameSession snapshot changed on roundtrip")
	}

	continueTactical := func(target *GameSession) {
		endAttacker, _ := battle.NewEndActivationCommand(2, battle.EndActivationPayload{ShipID: fixture.attackerShipIDs[0]})
		if err := target.SubmitBattleCommand(child.Spec.ID, child.Spec.Attacker.SeatID, endAttacker); err != nil {
			t.Fatal(err)
		}
		endDefender, _ := battle.NewEndActivationCommand(3, battle.EndActivationPayload{ShipID: fixture.defenderShipIDs[0]})
		if err := target.SubmitBattleCommand(child.Spec.ID, child.Spec.Defender.SeatID, endDefender); err != nil {
			t.Fatal(err)
		}
		fire2, _ := battle.NewFireBeamCommand(4, battle.FireBeamPayload{ShipID: fixture.attackerShipIDs[0], TargetShipID: fixture.defenderShipIDs[0], WeaponSlot: 0})
		if err := target.SubmitBattleCommand(child.Spec.ID, child.Spec.Attacker.SeatID, fire2); err != nil {
			t.Fatal(err)
		}
	}
	continueTactical(s)
	continueTactical(restored)
	leftObserver, _ := s.ObserverView()
	rightObserver, _ := restored.ObserverView()
	leftJSON, _ := json.Marshal(leftObserver)
	rightJSON, _ := json.Marshal(rightObserver)
	if !bytes.Equal(leftJSON, rightJSON) {
		t.Fatalf("tactical Session continuation diverged\nleft=%s\nright=%s", leftJSON, rightJSON)
	}
	if leftObserver.Phase != PhasePostResolution {
		t.Fatalf("continued tactical phase=%q", leftObserver.Phase)
	}
	if err := s.CompleteTurn(); err != nil {
		t.Fatal(err)
	}
	if err := restored.CompleteTurn(); err != nil {
		t.Fatal(err)
	}
	left, err := s.MarshalLiveSnapshot(resolver.Rules)
	if err != nil {
		t.Fatal(err)
	}
	right, err := restored.MarshalLiveSnapshot(resolver.Rules)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(left, right) {
		t.Fatal("tactical Session save/load continuation diverged after next Planning boundary")
	}
}
