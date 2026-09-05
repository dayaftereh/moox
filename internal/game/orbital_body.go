package game

import "moox/internal/core"

type orbitalBodyTarget struct {
	System *core.StarSystem
	Body   *core.OrbitalBody
	Planet *core.Planet
}

func orbitalBodyTargetByID(state *core.GameState, bodyID core.ID) orbitalBodyTarget {
	if state == nil || bodyID == 0 {
		return orbitalBodyTarget{}
	}
	for si := range state.Galaxy.Systems {
		system := &state.Galaxy.Systems[si]
		for bi := range system.Bodies {
			body := &system.Bodies[bi]
			if body.ID != bodyID {
				continue
			}
			var planet *core.Planet
			if body.PlanetID != 0 {
				for pi := range system.Planets {
					if system.Planets[pi].ID == body.PlanetID {
						planet = &system.Planets[pi]
						break
					}
				}
			}
			return orbitalBodyTarget{System: system, Body: body, Planet: planet}
		}
		// Backward-compatible implicit Planet body for pre-15.2 fixtures/saves.
		for pi := range system.Planets {
			planet := &system.Planets[pi]
			if planet.ID == bodyID {
				return orbitalBodyTarget{System: system, Planet: planet}
			}
		}
	}
	return orbitalBodyTarget{}
}

func outpostTargetBodyID(outpost core.Outpost) core.ID {
	if outpost.BodyID != 0 {
		return outpost.BodyID
	}
	return outpost.PlanetID
}

func setOrbitalBodyOutpost(target orbitalBodyTarget, outpostID core.ID) {
	if target.Body != nil {
		target.Body.OutpostID = outpostID
	}
	if target.Planet != nil {
		target.Planet.OutpostID = outpostID
	}
}

func orbitalBodyOccupiedByOutpost(state *core.GameState, target orbitalBodyTarget) bool {
	if target.Body != nil && target.Body.OutpostID != 0 {
		return true
	}
	if target.Planet != nil && target.Planet.OutpostID != 0 {
		return true
	}
	bodyID := core.ID(0)
	if target.Body != nil {
		bodyID = target.Body.ID
	} else if target.Planet != nil {
		bodyID = target.Planet.ID
	}
	for _, outpost := range state.Outposts {
		if outpostTargetBodyID(outpost) == bodyID {
			return true
		}
	}
	return false
}
