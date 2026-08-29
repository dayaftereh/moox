package core

import (
	"fmt"
	"sort"
)

type StrategicFleetRole string

const (
	StrategicFleetRoleCombat   StrategicFleetRole = "combat"
	StrategicFleetRoleCivilian StrategicFleetRole = "civilian"
)

type StrategicFleet struct {
	ID         ID                 `json:"id"`
	EmpireID   ID                 `json:"empire_id"`
	Role       StrategicFleetRole `json:"role"`
	AtSystemID ID                 `json:"at_system_id,omitempty"`
}

type DiplomaticStance string

const (
	DiplomaticStanceNeutral DiplomaticStance = "neutral"
	DiplomaticStanceHostile DiplomaticStance = "hostile"
)

type DiplomaticRelation struct {
	FromEmpireID ID               `json:"from_empire_id"`
	ToEmpireID   ID               `json:"to_empire_id"`
	Stance       DiplomaticStance `json:"stance"`
}

func (s *GameState) DiplomaticStanceBetween(fromEmpireID, toEmpireID ID) DiplomaticStance {
	if s == nil || fromEmpireID == 0 || toEmpireID == 0 || fromEmpireID == toEmpireID {
		return DiplomaticStanceNeutral
	}
	index := sort.Search(len(s.DiplomaticRelations), func(i int) bool {
		relation := s.DiplomaticRelations[i]
		return relation.FromEmpireID > fromEmpireID || (relation.FromEmpireID == fromEmpireID && relation.ToEmpireID >= toEmpireID)
	})
	if index < len(s.DiplomaticRelations) {
		relation := s.DiplomaticRelations[index]
		if relation.FromEmpireID == fromEmpireID && relation.ToEmpireID == toEmpireID {
			return relation.Stance
		}
	}
	return DiplomaticStanceNeutral
}

func validateStrategicState(
	state *GameState,
	empireIDs map[ID]struct{},
	systemIDs map[ID]struct{},
	checkID func(ID, string) error,
) error {
	lastFleetID := ID(0)
	for i, fleet := range state.StrategicFleets {
		if i > 0 && fleet.ID <= lastFleetID {
			return fmt.Errorf("strategic fleets must be strictly ascending by id")
		}
		if err := checkID(fleet.ID, fmt.Sprintf("strategic_fleet[%d]", i)); err != nil {
			return err
		}
		if _, ok := empireIDs[fleet.EmpireID]; !ok {
			return fmt.Errorf("strategic_fleet[%d] references unknown empire %d", i, fleet.EmpireID)
		}
		switch fleet.Role {
		case StrategicFleetRoleCombat, StrategicFleetRoleCivilian:
		default:
			return fmt.Errorf("strategic_fleet[%d] role %q is invalid", i, fleet.Role)
		}
		if fleet.AtSystemID != 0 {
			if _, ok := systemIDs[fleet.AtSystemID]; !ok {
				return fmt.Errorf("strategic_fleet[%d] references unknown star system %d", i, fleet.AtSystemID)
			}
		}
		lastFleetID = fleet.ID
	}

	var previous DiplomaticRelation
	for i, relation := range state.DiplomaticRelations {
		if relation.FromEmpireID == 0 || relation.ToEmpireID == 0 {
			return fmt.Errorf("diplomatic_relation[%d] empire ids must be non-zero", i)
		}
		if relation.FromEmpireID == relation.ToEmpireID {
			return fmt.Errorf("diplomatic_relation[%d] cannot target the same empire", i)
		}
		if _, ok := empireIDs[relation.FromEmpireID]; !ok {
			return fmt.Errorf("diplomatic_relation[%d] references unknown source empire %d", i, relation.FromEmpireID)
		}
		if _, ok := empireIDs[relation.ToEmpireID]; !ok {
			return fmt.Errorf("diplomatic_relation[%d] references unknown target empire %d", i, relation.ToEmpireID)
		}
		switch relation.Stance {
		case DiplomaticStanceNeutral, DiplomaticStanceHostile:
		default:
			return fmt.Errorf("diplomatic_relation[%d] stance %q is invalid", i, relation.Stance)
		}
		if i > 0 {
			if relation.FromEmpireID < previous.FromEmpireID || (relation.FromEmpireID == previous.FromEmpireID && relation.ToEmpireID <= previous.ToEmpireID) {
				return fmt.Errorf("diplomatic relations must be strictly ascending by from_empire_id,to_empire_id")
			}
		}
		previous = relation
	}
	return nil
}
