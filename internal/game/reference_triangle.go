package game

import (
	"fmt"

	"moox/internal/core"
)

const ReferenceTriangleScenarioID = "triangle-2pc-v1"
const ReferenceTriangleSeed uint64 = 0x800A

var referenceTrianglePlayers = []NewGamePlayerSpec{
	{SeatID: 1, EmpireName: "Human", RaceID: "human"},
	{SeatID: 2, EmpireName: "Darlok", RaceID: "darlok"},
	{SeatID: 3, EmpireName: "Psilon", RaceID: "psilon"},
}

// NewReferenceTriangleGame builds the durable three-player movement/contact QA
// scenario. Its three home systems form a near-equilateral integer-coordinate
// triangle whose authoritative strategic distance is exactly 2 pc on every edge.
func (r *EconomyRules) NewReferenceTriangleGame(seed uint64) (NewGameResult, error) {
	if r == nil || r.NewGameGalaxy == nil {
		return NewGameResult{}, fmt.Errorf("new game galaxy rules are unavailable")
	}

	state := core.NewGameState(seed)
	rng := state.RNG()
	state.Galaxy.ID = state.NewID()

	players := make([]NewGamePlayerResult, len(referenceTrianglePlayers))
	for i, player := range referenceTrianglePlayers {
		if _, ok := r.RaceModifiers[player.RaceID]; !ok {
			return NewGameResult{}, fmt.Errorf("reference triangle ruleset has no race %q", player.RaceID)
		}
		empire := core.Empire{
			ID:              state.NewID(),
			Name:            player.EmpireName,
			RaceID:          player.RaceID,
			PlayerColorSlot: i + 1,
			Freighters:      r.NewGameGalaxy.Start.Freighters,
			Treasury:        core.EmpireTreasuryState{BalanceBC: r.NewGameGalaxy.Start.TreasuryBC},
		}
		state.Empires = append(state.Empires, empire)
		players[i] = NewGamePlayerResult{SeatID: player.SeatID, EmpireID: empire.ID, RaceID: empire.RaceID, Name: empire.Name}
	}

	if err := r.InitializeNewGameTechnologies(state, NewGameTechnologyStateOptions{
		Level:           NewGameTechnologyAverage,
		StrategicCombat: false,
		NewGameRNG:      rng,
	}); err != nil {
		return NewGameResult{}, fmt.Errorf("initialize reference triangle technologies: %w", err)
	}

	systems := r.referenceTriangleSystemPlans(state)
	if err := r.materializeNewGameSystems(state, systems); err != nil {
		return NewGameResult{}, err
	}
	homeIndexes := []int{0, 1, 2}
	if err := r.initializeNewGameStartingAssets(state, homeIndexes, NewGameTechnologyAverage); err != nil {
		return NewGameResult{}, err
	}
	if err := materializeNewGameBodies(state, systems); err != nil {
		return NewGameResult{}, err
	}
	if err := r.finalizeNewGameEconomy(state); err != nil {
		return NewGameResult{}, err
	}

	state.AddEvent("game_started", "deterministic 2pc triangle reference game created")
	if err := state.Validate(); err != nil {
		return NewGameResult{}, fmt.Errorf("validate reference triangle game: %w", err)
	}
	state.CommitRNG(rng)
	return NewGameResult{State: state, Players: players}, nil
}

func (r *EconomyRules) referenceTriangleSystemPlans(state *core.GameState) []newGameSystemPlan {
	type star struct {
		name string
		x    int
		y    int
	}
	// Integer map coordinates cannot express a mathematically exact equilateral
	// triangle. These points are visually equilateral while every edge remains in
	// the authoritative 2 pc distance bucket (ceil(pixelDistance/30) == 2).
	stars := []star{
		{name: "Human Home", x: 430, y: 330},
		{name: "Darlok Home", x: 489, y: 330},
		{name: "Psilon Home", x: 459, y: 381},
	}
	plans := make([]newGameSystemPlan, len(stars))
	for i, star := range stars {
		plans[i] = newGameSystemPlan{
			ID:       state.NewID(),
			Name:     star.name,
			X:        star.x,
			Y:        star.y,
			Spectral: 2,
			Planets: []newGamePlanetPlan{
				{Orbit: 0, SizeID: r.NewGameGalaxy.Homeworld.SizeID, MineralID: r.NewGameGalaxy.Homeworld.MineralID, GravityID: r.NewGameGalaxy.Homeworld.GravityID, ClimateID: r.NewGameGalaxy.Homeworld.ClimateID},
				{Orbit: 2, SizeID: "small", MineralID: "poor", GravityID: "normal_g", ClimateID: "desert"},
				{Orbit: 4, SizeID: "large", MineralID: "abundant", GravityID: "normal_g", ClimateID: "barren"},
			},
			Bodies: []newGameBodyPlan{
				{Orbit: 0, Kind: core.OrbitalBodyPlanet},
				{Orbit: 2, Kind: core.OrbitalBodyPlanet},
				{Orbit: 4, Kind: core.OrbitalBodyPlanet},
			},
		}
	}
	return plans
}
