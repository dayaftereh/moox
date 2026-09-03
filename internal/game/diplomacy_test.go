package game

import (
	"fmt"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
	"reflect"
)

func diplomacyTestState(t *testing.T) (*core.GameState, core.ID, core.ID) {
	t.Helper()
	state := core.NewSmallFixture(0xD1A10)
	first := state.Empires[0].ID
	second := state.NewID()
	state.Empires = append(state.Empires, core.Empire{ID: second, Name: "Second", RaceID: "darlok"})
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	return state, first, second
}

func TestDiplomacyWarOfferAcceptLifecycle(t *testing.T) {
	state, first, second := diplomacyTestState(t)
	declare, _ := NewDeclareWarCommand(1, second)
	events, err := ResolveDiplomacyCommand(state, first, protocol.SeatID(1), declare)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != "diplomacy.war_declared" {
		t.Fatalf("declare events=%+v", events)
	}
	if state.DiplomaticStanceBetween(first, second) != core.DiplomaticStanceWar || state.DiplomaticStanceBetween(second, first) != core.DiplomaticStanceWar {
		t.Fatalf("war not reciprocal: %+v", state.DiplomaticRelations)
	}
	if !state.MayAttackEmpire(first, second) || !state.MayAttackEmpire(second, first) {
		t.Fatal("war must authorize both normal attack directions")
	}
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}

	offer, _ := NewOfferPeaceCommand(1, second)
	events, err = ResolveDiplomacyCommand(state, first, 1, offer)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != "diplomacy.peace_offered" || !state.HasDiplomaticPeaceOffer(first, second) {
		t.Fatalf("offer state/events=%+v %+v", state.DiplomaticPeaceOffers, events)
	}
	reciprocalOffer, _ := NewOfferPeaceCommand(1, first)
	if _, err := ResolveDiplomacyCommand(state, second, 2, reciprocalOffer); err == nil {
		t.Fatal("expected reciprocal offer to be rejected in favor of accept")
	}

	accept, _ := NewAcceptPeaceCommand(1, first)
	events, err = ResolveDiplomacyCommand(state, second, 2, accept)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Kind != "diplomacy.peace_accepted" {
		t.Fatalf("accept events=%+v", events)
	}
	if state.DiplomaticStanceBetween(first, second) != core.DiplomaticStancePeace || state.DiplomaticStanceBetween(second, first) != core.DiplomaticStancePeace {
		t.Fatalf("peace not reciprocal: %+v", state.DiplomaticRelations)
	}
	if len(state.DiplomaticPeaceOffers) != 0 {
		t.Fatalf("offers not cleared: %+v", state.DiplomaticPeaceOffers)
	}
	if state.MayAttackEmpire(first, second) || state.MayAttackEmpire(second, first) {
		t.Fatal("peace must not authorize normal attack")
	}
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestDiplomacyRejectsInvalidPayloadAndTransitions(t *testing.T) {
	state, first, second := diplomacyTestState(t)
	bad := protocol.Command{SchemaVersion: protocol.CommandSchemaVersion, Sequence: 1, Kind: CommandDeclareWar, Payload: []byte(`{"target_empire_id":` + fmtID(second) + `,"extra":true}`)}
	if _, err := ResolveDiplomacyCommand(state, first, 1, bad); err == nil {
		t.Fatal("expected unknown JSON field to fail")
	}
	offer, _ := NewOfferPeaceCommand(1, second)
	if _, err := ResolveDiplomacyCommand(state, first, 1, offer); err == nil {
		t.Fatal("expected peace offer while neutral to fail")
	}
	accept, _ := NewAcceptPeaceCommand(1, second)
	if _, err := ResolveDiplomacyCommand(state, first, 1, accept); err == nil {
		t.Fatal("expected peace accept without war/offer to fail")
	}
}

func mustDiplomacyCommand(t *testing.T, command protocol.Command, err error) protocol.Command {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	return command
}
func fmtID(id core.ID) string { return fmt.Sprintf("%d", id) }

func TestDiplomacyLifecycleDeterministicReplay(t *testing.T) {
	run := func() ([]byte, []DomainEvent) {
		state, first, second := diplomacyTestState(t)
		declare, err := NewDeclareWarCommand(1, second)
		if err != nil {
			t.Fatal(err)
		}
		offer, err := NewOfferPeaceCommand(1, second)
		if err != nil {
			t.Fatal(err)
		}
		accept, err := NewAcceptPeaceCommand(1, first)
		if err != nil {
			t.Fatal(err)
		}
		commands := []struct {
			actor core.ID
			seat  protocol.SeatID
			cmd   protocol.Command
		}{
			{actor: first, seat: 1, cmd: declare},
			{actor: first, seat: 1, cmd: offer},
			{actor: second, seat: 2, cmd: accept},
		}
		var events []DomainEvent
		for _, step := range commands {
			resolved, err := ResolveDiplomacyCommand(state, step.actor, step.seat, step.cmd)
			if err != nil {
				t.Fatal(err)
			}
			events = append(events, resolved...)
		}
		encoded, err := core.MarshalState(state)
		if err != nil {
			t.Fatal(err)
		}
		return encoded, events
	}
	stateA, eventsA := run()
	stateB, eventsB := run()
	if string(stateA) != string(stateB) {
		t.Fatalf("replayed diplomacy state diverged:\nA=%s\nB=%s", stateA, stateB)
	}
	if !reflect.DeepEqual(eventsA, eventsB) {
		t.Fatalf("replayed diplomacy events diverged:\nA=%+v\nB=%+v", eventsA, eventsB)
	}
}
