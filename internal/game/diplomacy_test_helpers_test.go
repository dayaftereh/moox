package game

import (
	"sort"

	"moox/internal/core"
)

func reciprocalWarRelations(a, b core.ID) []core.DiplomaticRelation {
	return warRelations([2]core.ID{a, b})
}

func warRelations(pairs ...[2]core.ID) []core.DiplomaticRelation {
	relations := make([]core.DiplomaticRelation, 0, len(pairs)*2)
	for _, pair := range pairs {
		a, b := pair[0], pair[1]
		relations = append(relations,
			core.DiplomaticRelation{FromEmpireID: a, ToEmpireID: b, Stance: core.DiplomaticStanceWar},
			core.DiplomaticRelation{FromEmpireID: b, ToEmpireID: a, Stance: core.DiplomaticStanceWar},
		)
	}
	sort.Slice(relations, func(i, j int) bool {
		if relations[i].FromEmpireID != relations[j].FromEmpireID {
			return relations[i].FromEmpireID < relations[j].FromEmpireID
		}
		return relations[i].ToEmpireID < relations[j].ToEmpireID
	})
	return relations
}
