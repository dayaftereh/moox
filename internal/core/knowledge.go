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
