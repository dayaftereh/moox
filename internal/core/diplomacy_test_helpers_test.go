package core

func reciprocalRelations(a, b ID, stance DiplomaticStance) []DiplomaticRelation {
	if b < a {
		a, b = b, a
	}
	return []DiplomaticRelation{{FromEmpireID: a, ToEmpireID: b, Stance: stance}, {FromEmpireID: b, ToEmpireID: a, Stance: stance}}
}
