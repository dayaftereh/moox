package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
	"moox/internal/protocol"
)

const (
	CommandDeclareWar  = "diplomacy.declare_war"
	CommandOfferPeace  = "diplomacy.offer_peace"
	CommandAcceptPeace = "diplomacy.accept_peace"
)

type DeclareWarPayload struct {
	TargetEmpireID core.ID `json:"target_empire_id"`
}
type OfferPeacePayload struct {
	TargetEmpireID core.ID `json:"target_empire_id"`
}
type AcceptPeacePayload struct {
	FromEmpireID core.ID `json:"from_empire_id"`
}

type WarDeclaredEvent struct {
	ActorEmpireID  core.ID               `json:"actor_empire_id"`
	TargetEmpireID core.ID               `json:"target_empire_id"`
	PreviousStance core.DiplomaticStance `json:"previous_stance"`
}
type PeaceOfferedEvent struct {
	FromEmpireID core.ID `json:"from_empire_id"`
	ToEmpireID   core.ID `json:"to_empire_id"`
}
type PeaceAcceptedEvent struct {
	AcceptingEmpireID core.ID               `json:"accepting_empire_id"`
	OfferingEmpireID  core.ID               `json:"offering_empire_id"`
	PreviousStance    core.DiplomaticStance `json:"previous_stance"`
	CurrentStance     core.DiplomaticStance `json:"current_stance"`
}

func IsDiplomacyCommand(kind string) bool {
	switch kind {
	case CommandDeclareWar, CommandOfferPeace, CommandAcceptPeace:
		return true
	default:
		return false
	}
}
func NewDeclareWarCommand(sequence uint32, target core.ID) (protocol.Command, error) {
	return protocol.NewCommand(sequence, CommandDeclareWar, DeclareWarPayload{TargetEmpireID: target})
}
func NewOfferPeaceCommand(sequence uint32, target core.ID) (protocol.Command, error) {
	return protocol.NewCommand(sequence, CommandOfferPeace, OfferPeacePayload{TargetEmpireID: target})
}
func NewAcceptPeaceCommand(sequence uint32, from core.ID) (protocol.Command, error) {
	return protocol.NewCommand(sequence, CommandAcceptPeace, AcceptPeacePayload{FromEmpireID: from})
}

func ResolveDiplomacyCommand(state *core.GameState, actor core.ID, seatID protocol.SeatID, command protocol.Command) ([]DomainEvent, error) {
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}
	if actor == 0 || empireByID(state, actor) == nil {
		return nil, fmt.Errorf("acting empire %d is unknown", actor)
	}
	switch command.Kind {
	case CommandDeclareWar:
		var p DeclareWarPayload
		if err := decodeStrictCommandPayload(command, CommandDeclareWar, &p); err != nil {
			return nil, err
		}
		if err := validateDiplomacyTarget(state, actor, p.TargetEmpireID); err != nil {
			return nil, err
		}
		previous := state.DiplomaticStanceBetween(actor, p.TargetEmpireID)
		if previous == core.DiplomaticStanceWar {
			return nil, fmt.Errorf("empires %d and %d are already at war", actor, p.TargetEmpireID)
		}
		setDiplomaticPair(state, actor, p.TargetEmpireID, core.DiplomaticStanceWar)
		removePeaceOffersForPair(state, actor, p.TargetEmpireID)
		event, err := NewDomainEvent("diplomacy.war_declared", seatID, command.Sequence, WarDeclaredEvent{ActorEmpireID: actor, TargetEmpireID: p.TargetEmpireID, PreviousStance: previous})
		if err != nil {
			return nil, err
		}
		return []DomainEvent{event}, nil
	case CommandOfferPeace:
		var p OfferPeacePayload
		if err := decodeStrictCommandPayload(command, CommandOfferPeace, &p); err != nil {
			return nil, err
		}
		if err := validateDiplomacyTarget(state, actor, p.TargetEmpireID); err != nil {
			return nil, err
		}
		if state.DiplomaticStanceBetween(actor, p.TargetEmpireID) != core.DiplomaticStanceWar {
			return nil, fmt.Errorf("peace may only be offered while at war")
		}
		if state.HasDiplomaticPeaceOffer(actor, p.TargetEmpireID) {
			return nil, fmt.Errorf("peace offer from empire %d to %d already exists", actor, p.TargetEmpireID)
		}
		if state.HasDiplomaticPeaceOffer(p.TargetEmpireID, actor) {
			return nil, fmt.Errorf("incoming peace offer from empire %d must be accepted instead", p.TargetEmpireID)
		}
		insertPeaceOffer(state, core.DiplomaticPeaceOffer{FromEmpireID: actor, ToEmpireID: p.TargetEmpireID})
		event, err := NewDomainEvent("diplomacy.peace_offered", seatID, command.Sequence, PeaceOfferedEvent{FromEmpireID: actor, ToEmpireID: p.TargetEmpireID})
		if err != nil {
			return nil, err
		}
		return []DomainEvent{event}, nil
	case CommandAcceptPeace:
		var p AcceptPeacePayload
		if err := decodeStrictCommandPayload(command, CommandAcceptPeace, &p); err != nil {
			return nil, err
		}
		if err := validateDiplomacyTarget(state, actor, p.FromEmpireID); err != nil {
			return nil, err
		}
		if state.DiplomaticStanceBetween(actor, p.FromEmpireID) != core.DiplomaticStanceWar {
			return nil, fmt.Errorf("peace may only be accepted while at war")
		}
		if !state.HasDiplomaticPeaceOffer(p.FromEmpireID, actor) {
			return nil, fmt.Errorf("no incoming peace offer from empire %d", p.FromEmpireID)
		}
		setDiplomaticPair(state, actor, p.FromEmpireID, core.DiplomaticStancePeace)
		removePeaceOffersForPair(state, actor, p.FromEmpireID)
		event, err := NewDomainEvent("diplomacy.peace_accepted", seatID, command.Sequence, PeaceAcceptedEvent{AcceptingEmpireID: actor, OfferingEmpireID: p.FromEmpireID, PreviousStance: core.DiplomaticStanceWar, CurrentStance: core.DiplomaticStancePeace})
		if err != nil {
			return nil, err
		}
		return []DomainEvent{event}, nil
	default:
		return nil, fmt.Errorf("unsupported diplomacy command kind %q", command.Kind)
	}
}

func validateDiplomacyTarget(state *core.GameState, actor, other core.ID) error {
	if other == 0 {
		return fmt.Errorf("target empire id must be non-zero")
	}
	if other == actor {
		return fmt.Errorf("empire %d cannot target itself diplomatically", actor)
	}
	if empireByID(state, other) == nil {
		return fmt.Errorf("target empire %d is unknown", other)
	}
	return nil
}
func setDiplomaticPair(state *core.GameState, a, b core.ID, stance core.DiplomaticStance) {
	upsertDiplomaticRelation(state, core.DiplomaticRelation{FromEmpireID: a, ToEmpireID: b, Stance: stance})
	upsertDiplomaticRelation(state, core.DiplomaticRelation{FromEmpireID: b, ToEmpireID: a, Stance: stance})
}
func upsertDiplomaticRelation(state *core.GameState, relation core.DiplomaticRelation) {
	i := sort.Search(len(state.DiplomaticRelations), func(i int) bool {
		r := state.DiplomaticRelations[i]
		return r.FromEmpireID > relation.FromEmpireID || (r.FromEmpireID == relation.FromEmpireID && r.ToEmpireID >= relation.ToEmpireID)
	})
	if i < len(state.DiplomaticRelations) && state.DiplomaticRelations[i].FromEmpireID == relation.FromEmpireID && state.DiplomaticRelations[i].ToEmpireID == relation.ToEmpireID {
		state.DiplomaticRelations[i] = relation
		return
	}
	state.DiplomaticRelations = append(state.DiplomaticRelations, core.DiplomaticRelation{})
	copy(state.DiplomaticRelations[i+1:], state.DiplomaticRelations[i:])
	state.DiplomaticRelations[i] = relation
}
func insertPeaceOffer(state *core.GameState, offer core.DiplomaticPeaceOffer) {
	i := sort.Search(len(state.DiplomaticPeaceOffers), func(i int) bool {
		o := state.DiplomaticPeaceOffers[i]
		return o.FromEmpireID > offer.FromEmpireID || (o.FromEmpireID == offer.FromEmpireID && o.ToEmpireID >= offer.ToEmpireID)
	})
	state.DiplomaticPeaceOffers = append(state.DiplomaticPeaceOffers, core.DiplomaticPeaceOffer{})
	copy(state.DiplomaticPeaceOffers[i+1:], state.DiplomaticPeaceOffers[i:])
	state.DiplomaticPeaceOffers[i] = offer
}
func removePeaceOffersForPair(state *core.GameState, a, b core.ID) {
	kept := state.DiplomaticPeaceOffers[:0]
	for _, o := range state.DiplomaticPeaceOffers {
		if (o.FromEmpireID == a && o.ToEmpireID == b) || (o.FromEmpireID == b && o.ToEmpireID == a) {
			continue
		}
		kept = append(kept, o)
	}
	state.DiplomaticPeaceOffers = kept
}
