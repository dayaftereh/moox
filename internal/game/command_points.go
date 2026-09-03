package game

import (
	"fmt"

	"moox/internal/core"
)

const (
	commandStationStarBase      = "star_base"
	commandStationBattlestation = "battlestation"
	commandStationStarFortress  = "star_fortress"
)

func commandStationTier(buildingID string) int {
	switch buildingID {
	case commandStationStarBase:
		return 1
	case commandStationBattlestation:
		return 2
	case commandStationStarFortress:
		return 3
	default:
		return 0
	}
}

func colonyCommandStation(colony *core.Colony) (string, int, error) {
	if colony == nil {
		return "", 0, fmt.Errorf("colony must not be nil")
	}
	var stationID string
	var tier int
	for _, buildingID := range colony.Buildings {
		candidateTier := commandStationTier(buildingID)
		if candidateTier == 0 {
			continue
		}
		if tier != 0 {
			return "", 0, fmt.Errorf("colony %d owns multiple command station tiers %q and %q", colony.ID, stationID, buildingID)
		}
		stationID = buildingID
		tier = candidateTier
	}
	return stationID, tier, nil
}

func applyCompletedCommandStation(colony *core.Colony, buildingID string) error {
	if colony == nil {
		return fmt.Errorf("colony must not be nil")
	}
	newTier := commandStationTier(buildingID)
	if newTier == 0 {
		colony.Buildings = append(colony.Buildings, buildingID)
		return nil
	}
	_, existingTier, err := colonyCommandStation(colony)
	if err != nil {
		return err
	}
	if existingTier >= newTier {
		return fmt.Errorf("colony %d cannot replace command station tier %d with tier %d", colony.ID, existingTier, newTier)
	}
	buildings := make([]string, 0, len(colony.Buildings)+1)
	for _, owned := range colony.Buildings {
		if commandStationTier(owned) != 0 {
			continue
		}
		buildings = append(buildings, owned)
	}
	colony.Buildings = append(buildings, buildingID)
	return nil
}

func (r *EconomyRules) deriveEmpireCommandPoints(state *core.GameState, empire *core.Empire) (core.EmpireCommandPoints, error) {
	if r == nil {
		return core.EmpireCommandPoints{}, fmt.Errorf("economy rules must not be nil")
	}
	if state == nil || empire == nil {
		return core.EmpireCommandPoints{}, fmt.Errorf("state and empire must not be nil")
	}

	used := 0
	for index := range state.Ships {
		ship := &state.Ships[index]
		if ship.EmpireID != empire.ID {
			continue
		}
		hull, ok := r.ShipHulls[ship.Spec.HullID]
		if !ok {
			return core.EmpireCommandPoints{}, fmt.Errorf("empire %d ship %d references unknown hull %q", empire.ID, ship.ID, ship.Spec.HullID)
		}
		if hull.SizeIndex < 0 {
			return core.EmpireCommandPoints{}, fmt.Errorf("empire %d ship %d hull %q has invalid size index %d", empire.ID, ship.ID, ship.Spec.HullID, hull.SizeIndex)
		}
		used += hull.SizeIndex + 1
	}
	for index := range state.StrategicFleets {
		fleet := &state.StrategicFleets[index]
		if fleet.EmpireID != empire.ID {
			continue
		}
		switch fleet.SpecialKind {
		case core.StrategicFleetSpecialColonyShip, core.StrategicFleetSpecialOutpostShip, core.StrategicFleetSpecialTroopTransport:
			used += r.CommandPoints.FixedSpecialShipPoints
		}
	}

	capacity := r.CommandPoints.BaseCapacity
	stationCount := 0
	ownedColonies := 0
	for index := range state.Colonies {
		colony := &state.Colonies[index]
		if colony.EmpireID != empire.ID {
			continue
		}
		ownedColonies++
		stationID, _, err := colonyCommandStation(colony)
		if err != nil {
			return core.EmpireCommandPoints{}, err
		}
		if stationID == "" {
			continue
		}
		points, ok := r.CommandPoints.StationPoints[stationID]
		if !ok {
			return core.EmpireCommandPoints{}, fmt.Errorf("command station %q has no Command Point rule", stationID)
		}
		capacity += points
		stationCount++
	}

	communicationsPerStation := 0
	for technologyID, points := range r.CommandPoints.CommunicationsPointsByTechnologyID {
		if points > communicationsPerStation && empireKnowsTechnology(empire, technologyID) {
			communicationsPerStation = points
		}
	}
	capacity += communicationsPerStation * stationCount

	modifiers, ok := r.RaceModifiers[empire.RaceID]
	if !ok {
		return core.EmpireCommandPoints{}, fmt.Errorf("unknown race %q for empire %d", empire.RaceID, empire.ID)
	}
	if modifiers.Warlord {
		capacity += r.CommandPoints.WarlordPointsPerColony * ownedColonies
	}

	if modifiers.GovernmentTraitID == r.CommandPoints.ImperiumRequiredGovernmentTraitID && empireKnowsTechnology(empire, r.CommandPoints.ImperiumTechnologyID) {
		capacity += capacity * r.CommandPoints.ImperiumBonusNumerator / r.CommandPoints.ImperiumBonusDenominator
	}
	if capacity < 0 || used < 0 {
		return core.EmpireCommandPoints{}, fmt.Errorf("empire %d derived invalid Command Points capacity=%d used=%d", empire.ID, capacity, used)
	}
	return core.EmpireCommandPoints{Capacity: capacity, Used: used}, nil
}

func commandPointOverage(points core.EmpireCommandPoints) int {
	if points.Used <= points.Capacity {
		return 0
	}
	return points.Used - points.Capacity
}
