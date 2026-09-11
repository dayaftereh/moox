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
	StrategicFleetSpecialNone           StrategicFleetSpecialKind = ""
	StrategicFleetSpecialColonyShip     StrategicFleetSpecialKind = "colony_ship"
	StrategicFleetSpecialOutpostShip    StrategicFleetSpecialKind = "outpost_ship"
	StrategicFleetSpecialTroopTransport StrategicFleetSpecialKind = "troop_transport"
)

type StrategicFleet struct {
	ID                  ID                        `json:"id"`
	EmpireID            ID                        `json:"empire_id"`
	Role                StrategicFleetRole        `json:"role"`
	SpecialKind         StrategicFleetSpecialKind `json:"special_kind,omitempty"`
	AtSystemID          ID                        `json:"at_system_id,omitempty"`
	SourceSystemID      ID                        `json:"source_system_id,omitempty"`
	DestinationSystemID ID                        `json:"destination_system_id,omitempty"`
	RemainingTurns      int                       `json:"remaining_turns,omitempty"`
	TransitTurnsTotal   int                       `json:"transit_turns_total,omitempty"`
	FTLSpeed            int                       `json:"ftl_speed,omitempty"`
	// Projection-only movement metadata. Authoritative GameState keeps these empty; player views fill an effective copy.
	WarpDriveID              string `json:"warp_drive_id,omitempty"`
	FuelCellID               string `json:"fuel_cell_id,omitempty"`
	FuelRangeParsecs         int    `json:"fuel_range_parsecs,omitempty"`
	RouteDistanceParsecs     int    `json:"route_distance_parsecs,omitempty"`
	RemainingDistanceParsecs int    `json:"remaining_distance_parsecs,omitempty"`
	ShipIDs                  []ID   `json:"ship_ids,omitempty"`
}

type DiplomaticStance string

const (
	DiplomaticStanceNeutral DiplomaticStance = "neutral"
	DiplomaticStancePeace   DiplomaticStance = "peace"
	DiplomaticStanceWar     DiplomaticStance = "war"
)

type DiplomaticRelation struct {
	FromEmpireID ID               `json:"from_empire_id"`
	ToEmpireID   ID               `json:"to_empire_id"`
	Stance       DiplomaticStance `json:"stance"`
}
type DiplomaticPeaceOffer struct {
	FromEmpireID ID `json:"from_empire_id"`
	ToEmpireID   ID `json:"to_empire_id"`
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
func (s *GameState) MayAttackEmpire(fromEmpireID, toEmpireID ID) bool {
	return s != nil && s.DiplomaticStanceBetween(fromEmpireID, toEmpireID) == DiplomaticStanceWar
}
func (s *GameState) HasDiplomaticPeaceOffer(fromEmpireID, toEmpireID ID) bool {
	if s == nil || fromEmpireID == 0 || toEmpireID == 0 || fromEmpireID == toEmpireID {
		return false
	}
	index := sort.Search(len(s.DiplomaticPeaceOffers), func(i int) bool {
		offer := s.DiplomaticPeaceOffers[i]
		return offer.FromEmpireID > fromEmpireID || (offer.FromEmpireID == fromEmpireID && offer.ToEmpireID >= toEmpireID)
	})
	return index < len(s.DiplomaticPeaceOffers) &&
		s.DiplomaticPeaceOffers[index].FromEmpireID == fromEmpireID &&
		s.DiplomaticPeaceOffers[index].ToEmpireID == toEmpireID
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
		if fleet.WarpDriveID != "" || fleet.FuelCellID != "" || fleet.FuelRangeParsecs != 0 || fleet.RouteDistanceParsecs != 0 || fleet.RemainingDistanceParsecs != 0 {
			return fmt.Errorf("strategic_fleet[%d] projection-only movement metadata must remain empty in authoritative state", i)
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
			} else if fleet.SourceSystemID != 0 || fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 || fleet.TransitTurnsTotal != 0 || fleet.FTLSpeed != 0 {
				return fmt.Errorf("strategic_fleet[%d] ordinary non-combat fleet cannot carry semantic transit state", i)
			}
		case StrategicFleetSpecialColonyShip, StrategicFleetSpecialOutpostShip, StrategicFleetSpecialTroopTransport:
			if fleet.Role != StrategicFleetRoleCivilian {
				return fmt.Errorf("strategic_fleet[%d] fixed special ship %q must be civilian", i, fleet.SpecialKind)
			}
			if fleet.FTLSpeed < 2 {
				return fmt.Errorf("strategic_fleet[%d] fixed special ship %q ftl_speed must be at least 2", i, fleet.SpecialKind)
			}
		default:
			return fmt.Errorf("strategic_fleet[%d] special_kind %q is invalid", i, fleet.SpecialKind)
		}
		if fleet.RemainingTurns < 0 || fleet.TransitTurnsTotal < 0 || fleet.FTLSpeed < 0 {
			return fmt.Errorf("strategic_fleet[%d] transit values must be non-negative", i)
		}
		if fleet.AtSystemID != 0 {
			if _, ok := systemIDs[fleet.AtSystemID]; !ok {
				return fmt.Errorf("strategic_fleet[%d] references unknown star system %d", i, fleet.AtSystemID)
			}
			if fleet.SourceSystemID != 0 || fleet.DestinationSystemID != 0 || fleet.RemainingTurns != 0 || fleet.TransitTurnsTotal != 0 {
				return fmt.Errorf("strategic_fleet[%d] stationary fleet cannot also be in transit", i)
			}
		} else if fleet.DestinationSystemID != 0 {
			if _, ok := systemIDs[fleet.DestinationSystemID]; !ok {
				return fmt.Errorf("strategic_fleet[%d] references unknown destination star system %d", i, fleet.DestinationSystemID)
			}
			if fleet.SourceSystemID != 0 {
				if _, ok := systemIDs[fleet.SourceSystemID]; !ok {
					return fmt.Errorf("strategic_fleet[%d] references unknown source star system %d", i, fleet.SourceSystemID)
				}
				if fleet.SourceSystemID == fleet.DestinationSystemID {
					return fmt.Errorf("strategic_fleet[%d] transit source and destination must differ", i)
				}
			}
			if (fleet.SourceSystemID == 0) != (fleet.TransitTurnsTotal == 0) {
				return fmt.Errorf("strategic_fleet[%d] transit route metadata must include both source_system_id and transit_turns_total", i)
			}
			if fleet.TransitTurnsTotal != 0 && fleet.TransitTurnsTotal < fleet.RemainingTurns {
				return fmt.Errorf("strategic_fleet[%d] transit_turns_total %d cannot be less than remaining_turns %d", i, fleet.TransitTurnsTotal, fleet.RemainingTurns)
			}
			fixedSpecial := fleet.SpecialKind == StrategicFleetSpecialColonyShip || fleet.SpecialKind == StrategicFleetSpecialOutpostShip || fleet.SpecialKind == StrategicFleetSpecialTroopTransport
			combatTransit := fleet.SpecialKind == StrategicFleetSpecialNone && fleet.Role == StrategicFleetRoleCombat
			if !fixedSpecial && !combatTransit {
				return fmt.Errorf("strategic_fleet[%d] fleet kind/role cannot use semantic transit in schema %d", i, StateSchemaVersion)
			}
			if fleet.RemainingTurns <= 0 {
				return fmt.Errorf("strategic_fleet[%d] in transit requires positive remaining_turns", i)
			}
		} else if fleet.SourceSystemID != 0 || fleet.TransitTurnsTotal != 0 {
			return fmt.Errorf("strategic_fleet[%d] fleet without destination cannot retain transit route metadata", i)
		} else if fleet.SpecialKind == StrategicFleetSpecialColonyShip || fleet.SpecialKind == StrategicFleetSpecialOutpostShip || fleet.SpecialKind == StrategicFleetSpecialTroopTransport {
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
		case DiplomaticStancePeace, DiplomaticStanceWar:
		default:
			return fmt.Errorf("diplomatic_relation[%d] stance %q is invalid; neutral relations must be omitted", i, relation.Stance)
		}
		if i > 0 && (relation.FromEmpireID < previous.FromEmpireID || (relation.FromEmpireID == previous.FromEmpireID && relation.ToEmpireID <= previous.ToEmpireID)) {
			return fmt.Errorf("diplomatic relations must be strictly ascending by from_empire_id,to_empire_id")
		}
		previous = relation
	}
	for i, relation := range state.DiplomaticRelations {
		if reverse := state.DiplomaticStanceBetween(relation.ToEmpireID, relation.FromEmpireID); reverse != relation.Stance {
			return fmt.Errorf("diplomatic_relation[%d] stance %q is not reciprocal (reverse=%q)", i, relation.Stance, reverse)
		}
	}
	var previousOffer DiplomaticPeaceOffer
	seenOfferPairs := make(map[[2]ID]struct{}, len(state.DiplomaticPeaceOffers))
	for i, offer := range state.DiplomaticPeaceOffers {
		if offer.FromEmpireID == 0 || offer.ToEmpireID == 0 {
			return fmt.Errorf("diplomatic_peace_offer[%d] empire ids must be non-zero", i)
		}
		if offer.FromEmpireID == offer.ToEmpireID {
			return fmt.Errorf("diplomatic_peace_offer[%d] cannot target the same empire", i)
		}
		if _, ok := empireIDs[offer.FromEmpireID]; !ok {
			return fmt.Errorf("diplomatic_peace_offer[%d] references unknown source empire %d", i, offer.FromEmpireID)
		}
		if _, ok := empireIDs[offer.ToEmpireID]; !ok {
			return fmt.Errorf("diplomatic_peace_offer[%d] references unknown target empire %d", i, offer.ToEmpireID)
		}
		if i > 0 && (offer.FromEmpireID < previousOffer.FromEmpireID || (offer.FromEmpireID == previousOffer.FromEmpireID && offer.ToEmpireID <= previousOffer.ToEmpireID)) {
			return fmt.Errorf("diplomatic peace offers must be strictly ascending by from_empire_id,to_empire_id")
		}
		if state.DiplomaticStanceBetween(offer.FromEmpireID, offer.ToEmpireID) != DiplomaticStanceWar {
			return fmt.Errorf("diplomatic_peace_offer[%d] requires war between empires %d and %d", i, offer.FromEmpireID, offer.ToEmpireID)
		}
		lo, hi := offer.FromEmpireID, offer.ToEmpireID
		if hi < lo {
			lo, hi = hi, lo
		}
		pair := [2]ID{lo, hi}
		if _, exists := seenOfferPairs[pair]; exists {
			return fmt.Errorf("diplomatic peace offers may contain at most one direction per empire pair %d/%d", lo, hi)
		}
		seenOfferPairs[pair] = struct{}{}
		previousOffer = offer
	}
	return nil
}
