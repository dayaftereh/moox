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

type StrategicFleetSpecialKind string

const (
	StrategicFleetSpecialNone        StrategicFleetSpecialKind = ""
	StrategicFleetSpecialColonyShip  StrategicFleetSpecialKind = "colony_ship"
	StrategicFleetSpecialOutpostShip StrategicFleetSpecialKind = "outpost_ship"
)

type StrategicFleet struct {
	ID                  ID                        `json:"id"`
	EmpireID            ID                        `json:"empire_id"`
	Role                StrategicFleetRole        `json:"role"`
	SpecialKind         StrategicFleetSpecialKind `json:"special_kind,omitempty"`
	AtSystemID          ID                        `json:"at_system_id,omitempty"`
	DestinationSystemID ID                        `json:"destination_system_id,omitempty"`
	RemainingTurns      int                       `json:"remaining_turns,omitempty"`
	FTLSpeed            int                       `json:"ftl_speed,omitempty"`
	ShipIDs             []ID                      `json:"ship_ids,omitempty"`
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
	ships map[ID]Ship,
	checkID func(ID, string) error,
) error {
	lastFleetID := ID(0)
	assignedShips := make(map[ID]ID, len(ships))
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
		if fleet.Role != StrategicFleetRoleCombat && len(fleet.ShipIDs) != 0 {
			return fmt.Errorf("strategic_fleet[%d] non-combat fleet cannot contain concrete military ships", i)
		}
		lastShipID := ID(0)
		for shipIndex, shipID := range fleet.ShipIDs {
			if shipIndex > 0 && shipID <= lastShipID {
				return fmt.Errorf("strategic_fleet[%d] ship ids must be strictly ascending", i)
			}
			ship, ok := ships[shipID]
			if !ok {
				return fmt.Errorf("strategic_fleet[%d] references unknown ship %d", i, shipID)
			}
			if ship.EmpireID != fleet.EmpireID {
				return fmt.Errorf("strategic_fleet[%d] ship %d owner %d differs from fleet owner %d", i, shipID, ship.EmpireID, fleet.EmpireID)
			}
			if previousFleet, exists := assignedShips[shipID]; exists {
				return fmt.Errorf("ship %d is assigned to both fleet %d and fleet %d", shipID, previousFleet, fleet.ID)
			}
			assignedShips[shipID] = fleet.ID
			lastShipID = shipID
		}
		switch fleet.SpecialKind {
		case StrategicFleetSpecialNone:
			if fleet.Role == StrategicFleetRoleCombat {
				if len(fleet.ShipIDs) == 0 {
					return fmt.Errorf("strategic_fleet[%d] combat fleet must contain at least one concrete ship", i)
				}
				if fleet.FTLSpeed != 0 {
					return fmt.Errorf("strategic_fleet[%d] ordinary combat fleet ftl_speed must remain zero; movement speed is derived", i)
				}
			} else if fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 || fleet.FTLSpeed != 0 {
				return fmt.Errorf("strategic_fleet[%d] ordinary non-combat fleet cannot carry semantic transit state", i)
			}
		case StrategicFleetSpecialColonyShip, StrategicFleetSpecialOutpostShip:
			if fleet.Role != StrategicFleetRoleCivilian {
				return fmt.Errorf("strategic_fleet[%d] fixed special ship %q must be civilian", i, fleet.SpecialKind)
			}
			if fleet.FTLSpeed < 2 {
				return fmt.Errorf("strategic_fleet[%d] fixed special ship %q ftl_speed must be at least 2", i, fleet.SpecialKind)
			}
		default:
			return fmt.Errorf("strategic_fleet[%d] special_kind %q is invalid", i, fleet.SpecialKind)
		}
		if fleet.RemainingTurns < 0 || fleet.FTLSpeed < 0 {
			return fmt.Errorf("strategic_fleet[%d] transit values must be non-negative", i)
		}
		if fleet.AtSystemID != 0 {
			if _, ok := systemIDs[fleet.AtSystemID]; !ok {
				return fmt.Errorf("strategic_fleet[%d] references unknown star system %d", i, fleet.AtSystemID)
			}
			if fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 {
				return fmt.Errorf("strategic_fleet[%d] stationary fleet cannot also be in transit", i)
			}
		} else if fleet.DestinationSystemID != 0 {
			if _, ok := systemIDs[fleet.DestinationSystemID]; !ok {
				return fmt.Errorf("strategic_fleet[%d] references unknown destination star system %d", i, fleet.DestinationSystemID)
			}
			fixedSpecial := fleet.SpecialKind == StrategicFleetSpecialColonyShip || fleet.SpecialKind == StrategicFleetSpecialOutpostShip
			combatTransit := fleet.SpecialKind == StrategicFleetSpecialNone && fleet.Role == StrategicFleetRoleCombat
			if !fixedSpecial && !combatTransit {
				return fmt.Errorf("strategic_fleet[%d] fleet kind/role cannot use semantic transit in schema %d", i, StateSchemaVersion)
			}
			if fleet.RemainingTurns <= 0 {
				return fmt.Errorf("strategic_fleet[%d] in transit requires positive remaining_turns", i)
			}
		} else if fleet.SpecialKind == StrategicFleetSpecialColonyShip || fleet.SpecialKind == StrategicFleetSpecialOutpostShip {
			return fmt.Errorf("strategic_fleet[%d] fixed special ship %q requires a current or destination system", i, fleet.SpecialKind)
		} else if fleet.Role == StrategicFleetRoleCombat && fleet.SpecialKind == StrategicFleetSpecialNone {
			return fmt.Errorf("strategic_fleet[%d] combat fleet requires a current or destination system", i)
		} else if fleet.RemainingTurns != 0 {
			return fmt.Errorf("strategic_fleet[%d] locationless fleet cannot have remaining_turns", i)
		}
		lastFleetID = fleet.ID
	}
	for shipID := range ships {
		if _, assigned := assignedShips[shipID]; !assigned {
			return fmt.Errorf("ship %d is not assigned to a strategic combat fleet", shipID)
		}
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
