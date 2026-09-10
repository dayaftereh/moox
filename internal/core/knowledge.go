package core

import "sort"

// HasVisitedSystem reports whether this empire has permanently visited a star
// system. VisitedSystemIDs is kept strictly ascending for deterministic state.
func (e *Empire) HasVisitedSystem(systemID ID) bool {
	if e == nil || systemID == 0 {
		return false
	}
	i := sort.Search(len(e.VisitedSystemIDs), func(i int) bool { return e.VisitedSystemIDs[i] >= systemID })
	return i < len(e.VisitedSystemIDs) && e.VisitedSystemIDs[i] == systemID
}

// MarkSystemVisited records permanent system knowledge and returns true only
// when the system was newly added.
func (e *Empire) MarkSystemVisited(systemID ID) bool {
	if e == nil || systemID == 0 {
		return false
	}
	i := sort.Search(len(e.VisitedSystemIDs), func(i int) bool { return e.VisitedSystemIDs[i] >= systemID })
	if i < len(e.VisitedSystemIDs) && e.VisitedSystemIDs[i] == systemID {
		return false
	}
	e.VisitedSystemIDs = append(e.VisitedSystemIDs, 0)
	copy(e.VisitedSystemIDs[i+1:], e.VisitedSystemIDs[i:])
	e.VisitedSystemIDs[i] = systemID
	return true
}

// HasKnownEmpire reports whether this empire has established first contact
// with another empire.
func (e *Empire) HasKnownEmpire(empireID ID) bool {
	if e == nil || empireID == 0 || empireID == e.ID {
		return false
	}
	i := sort.Search(len(e.KnownEmpireIDs), func(i int) bool { return e.KnownEmpireIDs[i] >= empireID })
	return i < len(e.KnownEmpireIDs) && e.KnownEmpireIDs[i] == empireID
}

// MarkEmpireKnown records durable first-contact knowledge.
func (e *Empire) MarkEmpireKnown(empireID ID) bool {
	if e == nil || empireID == 0 || empireID == e.ID {
		return false
	}
	i := sort.Search(len(e.KnownEmpireIDs), func(i int) bool { return e.KnownEmpireIDs[i] >= empireID })
	if i < len(e.KnownEmpireIDs) && e.KnownEmpireIDs[i] == empireID {
		return false
	}
	e.KnownEmpireIDs = append(e.KnownEmpireIDs, 0)
	copy(e.KnownEmpireIDs[i+1:], e.KnownEmpireIDs[i:])
	e.KnownEmpireIDs[i] = empireID
	return true
}

// EmpiresHaveContact is the player-knowledge authority for whether two empire
// identities may be exposed to one another. Explicit persisted diplomacy is
// accepted as compatibility evidence for saves created before KnownEmpireIDs.
func (s *GameState) EmpiresHaveContact(a, b ID) bool {
	if s == nil || a == 0 || b == 0 || a == b {
		return false
	}
	var ea, eb *Empire
	for i := range s.Empires {
		switch s.Empires[i].ID {
		case a:
			ea = &s.Empires[i]
		case b:
			eb = &s.Empires[i]
		}
	}
	if ea == nil || eb == nil {
		return false
	}
	if ea.HasKnownEmpire(b) || eb.HasKnownEmpire(a) {
		return true
	}
	for _, relation := range s.DiplomaticRelations {
		if (relation.FromEmpireID == a && relation.ToEmpireID == b) || (relation.FromEmpireID == b && relation.ToEmpireID == a) {
			return true
		}
	}
	for _, offer := range s.DiplomaticPeaceOffers {
		if (offer.FromEmpireID == a && offer.ToEmpireID == b) || (offer.FromEmpireID == b && offer.ToEmpireID == a) {
			return true
		}
	}
	return false
}

// MarkEmpiresKnown persists a symmetric first-contact pair.
func (s *GameState) MarkEmpiresKnown(a, b ID) bool {
	if s == nil || a == 0 || b == 0 || a == b {
		return false
	}
	var ea, eb *Empire
	for i := range s.Empires {
		switch s.Empires[i].ID {
		case a:
			ea = &s.Empires[i]
		case b:
			eb = &s.Empires[i]
		}
	}
	if ea == nil || eb == nil {
		return false
	}
	changedA := ea.MarkEmpireKnown(b)
	changedB := eb.MarkEmpireKnown(a)
	return changedA || changedB
}
