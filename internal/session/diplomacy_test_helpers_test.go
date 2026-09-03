package session

import "moox/internal/core"

func reciprocalWarRelations(a, b core.ID) []core.DiplomaticRelation {
	if b < a {
		a, b = b, a
	}
	return []core.DiplomaticRelation{{FromEmpireID: a, ToEmpireID: b, Stance: core.DiplomaticStanceWar}, {FromEmpireID: b, ToEmpireID: a, Stance: core.DiplomaticStanceWar}}
}
