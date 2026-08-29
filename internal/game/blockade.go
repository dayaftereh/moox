package game

import (
	"fmt"
	"sort"

	"moox/internal/core"
)

func recomputeSystemBlockades(state *core.GameState) error {
	if state == nil {
		return fmt.Errorf("game state must not be nil")
	}

	colonyOwnerByID := make(map[core.ID]core.ID, len(state.Colonies))
	for i := range state.Colonies {
		colony := state.Colonies[i]
		colonyOwnerByID[colony.ID] = colony.EmpireID
	}

	for si := range state.Galaxy.Systems {
		system := &state.Galaxy.Systems[si]
		targetOwners := make(map[core.ID]struct{})
		for _, planet := range system.Planets {
			if planet.ColonyID == 0 {
				continue
			}
			ownerID, ok := colonyOwnerByID[planet.ColonyID]
			if !ok {
				return fmt.Errorf("system %d planet %d references unknown colony %d", system.ID, planet.ID, planet.ColonyID)
			}
			targetOwners[ownerID] = struct{}{}
		}

		blockedOwners := make(map[core.ID]struct{})
		for _, fleet := range state.StrategicFleets {
			if fleet.Role != core.StrategicFleetRoleCombat || fleet.AtSystemID != system.ID {
				continue
			}
			for targetEmpireID := range targetOwners {
				if targetEmpireID == fleet.EmpireID {
					continue
				}
				if state.DiplomaticStanceBetween(fleet.EmpireID, targetEmpireID) == core.DiplomaticStanceHostile {
					blockedOwners[targetEmpireID] = struct{}{}
				}
			}
		}

		if len(blockedOwners) == 0 {
			system.BlockadedEmpireIDs = nil
			continue
		}
		blockaded := make([]core.ID, 0, len(blockedOwners))
		for empireID := range blockedOwners {
			blockaded = append(blockaded, empireID)
		}
		sort.Slice(blockaded, func(i, j int) bool { return blockaded[i] < blockaded[j] })
		system.BlockadedEmpireIDs = blockaded
	}
	return nil
}
