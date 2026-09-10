package session

import (
	"testing"

	"moox/internal/core"
	"moox/internal/game"
)

func TestDiplomacyImmediateLifecycleRevisionEventsAndProjection(t *testing.T) {
	state, seats := twoSeatFixture(t)
	first, second := state.Empires[0].ID, state.Empires[1].ID
	state.MarkEmpiresKnown(first, second)

	s, err := NewGameSession("diplomacy", state, seats)
	if err != nil {
		t.Fatal(err)
	}

	start := s.Status()
	declare, _ := game.NewDeclareWarCommand(1, second)
	if err := s.ResolveDiplomacyCommand(1, start.Revision, declare); err != nil {
		t.Fatal(err)
	}
	warStatus := s.Status()
	if warStatus.Revision != start.Revision+1 {
		t.Fatalf("war revision=%d want=%d", warStatus.Revision, start.Revision+1)
	}
	p1, _ := s.PlayerView(1)
	p2, _ := s.PlayerView(2)
	if len(p1.Diplomacy) != 1 || p1.Diplomacy[0].OtherEmpireID != second || p1.Diplomacy[0].Stance != core.DiplomaticStanceWar {
		t.Fatalf("seat1 diplomacy=%+v", p1.Diplomacy)
	}
	if len(p2.Diplomacy) != 1 || p2.Diplomacy[0].OtherEmpireID != first || p2.Diplomacy[0].Stance != core.DiplomaticStanceWar {
		t.Fatalf("seat2 diplomacy=%+v", p2.Diplomacy)
	}

	offer, _ := game.NewOfferPeaceCommand(1, second)
	if err := s.ResolveDiplomacyCommand(1, start.Revision, offer); err == nil {
		t.Fatal("expected stale immediate revision to fail")
	}
	if err := s.ResolveDiplomacyCommand(1, warStatus.Revision, offer); err != nil {
		t.Fatal(err)
	}
	offerStatus := s.Status()
	p1, _ = s.PlayerView(1)
	p2, _ = s.PlayerView(2)
	if !p1.Diplomacy[0].OutgoingPeaceOffer || p1.Diplomacy[0].IncomingPeaceOffer {
		t.Fatalf("seat1 offer projection=%+v", p1.Diplomacy[0])
	}
	if !p2.Diplomacy[0].IncomingPeaceOffer || p2.Diplomacy[0].OutgoingPeaceOffer {
		t.Fatalf("seat2 offer projection=%+v", p2.Diplomacy[0])
	}

	accept, _ := game.NewAcceptPeaceCommand(1, first)
	if err := s.ResolveDiplomacyCommand(2, offerStatus.Revision, accept); err != nil {
		t.Fatal(err)
	}
	p1, _ = s.PlayerView(1)
	p2, _ = s.PlayerView(2)
	if p1.Diplomacy[0].Stance != core.DiplomaticStancePeace || p2.Diplomacy[0].Stance != core.DiplomaticStancePeace {
		t.Fatalf("peace projection p1=%+v p2=%+v", p1.Diplomacy, p2.Diplomacy)
	}
	if p1.Diplomacy[0].IncomingPeaceOffer || p1.Diplomacy[0].OutgoingPeaceOffer || p2.Diplomacy[0].IncomingPeaceOffer || p2.Diplomacy[0].OutgoingPeaceOffer {
		t.Fatal("accepted peace offer was not cleared from projections")
	}

	observer, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, event := range observer.Events {
		if len(event.Kind) >= 10 && event.Kind[:10] == "diplomacy." {
			got = append(got, event.Kind)
		}
	}
	want := []string{"diplomacy.war_declared", "diplomacy.peace_offered", "diplomacy.peace_accepted"}
	if len(got) != len(want) {
		t.Fatalf("diplomacy events=%v want=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("diplomacy events=%v want=%v", got, want)
		}
	}
}

func TestDiplomacyImmediateClosesAtFirstTurnSubmission(t *testing.T) {
	state, seats := twoSeatFixture(t)
	s, err := NewGameSession("diplomacy-submit", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	status := s.Status()
	if err := s.SubmitTurn(makeBatch(t, "diplomacy-submit", 2, status.Turn, status.Revision, "test.wait")); err != nil {
		t.Fatal(err)
	}
	declare, _ := game.NewDeclareWarCommand(1, state.Empires[1].ID)
	if err := s.ResolveDiplomacyCommand(1, status.Revision, declare); err == nil {
		t.Fatal("expected diplomacy after first turn submission to fail")
	}
	if got := s.Status().Revision; got != status.Revision {
		t.Fatalf("rejected diplomacy changed revision: %d -> %d", status.Revision, got)
	}
	p1, _ := s.PlayerView(1)
	if p1.Diplomacy[0].Stance != core.DiplomaticStanceNeutral {
		t.Fatalf("rejected diplomacy changed stance: %+v", p1.Diplomacy)
	}
}

func TestDiplomacyImmediateRequiresSequenceOne(t *testing.T) {
	state, seats := twoSeatFixture(t)
	s, err := NewGameSession("diplomacy-sequence", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	declare, _ := game.NewDeclareWarCommand(2, state.Empires[1].ID)
	if err := s.ResolveDiplomacyCommand(1, s.Status().Revision, declare); err == nil {
		t.Fatal("expected non-one immediate command sequence to fail")
	}
}

func TestDiplomacyImmediateRejectedAfterBattleMaterializes(t *testing.T) {
	state, seats, fixture := makeTacticalStrategicFixture(t, 1)
	resolver := loadTacticalEconomyResolver(t)
	s, err := NewGameSession("diplomacy-battle", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	submitStubTurn(t, s, "diplomacy-battle")
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	before, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if before.Phase != PhaseEncounters || len(before.Battles) != 1 {
		t.Fatalf("battle did not materialize: phase=%s battles=%d", before.Phase, len(before.Battles))
	}
	offer, _ := game.NewOfferPeaceCommand(1, fixture.defenderEmpireID)
	if err := s.ResolveDiplomacyCommand(1, before.Revision, offer); err == nil {
		t.Fatal("expected diplomacy during materialized encounter to fail")
	}
	after, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	if after.Revision != before.Revision || after.Phase != PhaseEncounters || len(after.Battles) != 1 || len(after.State.DiplomaticPeaceOffers) != 0 {
		t.Fatalf("rejected encounter diplomacy mutated session: before=%+v after=%+v", before, after)
	}
}
