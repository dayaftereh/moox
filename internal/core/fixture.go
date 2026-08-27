package core

// NewSmallFixture builds the first deterministic headless fixture. The layout is
// intentionally explicit test scaffolding, not a claim about original MOO2
// galaxy-generation probabilities.
func NewSmallFixture(seed uint64) *GameState {
	s := NewGameState(seed)
	s.Galaxy.ID = s.NewID()

	systems := []struct {
		name    string
		x, y    int
		planets []Planet
	}{
		{name: "Alpha", x: 20, y: 35, planets: []Planet{{Name: "Alpha I", Orbit: 1, SizeID: "medium", MineralID: "abundant", GravityID: "normal_g", ClimateID: "terran"}}},
		{name: "Beta", x: 65, y: 42, planets: []Planet{{Name: "Beta I", Orbit: 1, SizeID: "small", MineralID: "rich", GravityID: "normal_g", ClimateID: "desert"}, {Name: "Beta II", Orbit: 2, SizeID: "large", MineralID: "poor", GravityID: "heavy_g", ClimateID: "barren"}}},
		{name: "Gamma", x: 44, y: 78, planets: []Planet{{Name: "Gamma I", Orbit: 1, SizeID: "huge", MineralID: "ultra_poor", GravityID: "low_g", ClimateID: "ocean"}}},
	}
	for _, spec := range systems {
		system := StarSystem{ID: s.NewID(), Name: spec.name, X: spec.x, Y: spec.y}
		for _, planetSpec := range spec.planets {
			planetSpec.ID = s.NewID()
			system.Planets = append(system.Planets, planetSpec)
		}
		s.Galaxy.Systems = append(s.Galaxy.Systems, system)
	}

	empire := Empire{ID: s.NewID(), Name: "Fixture Empire", RaceID: "human"}
	homeworld := &s.Galaxy.Systems[0].Planets[0]
	colony := Colony{ID: s.NewID(), EmpireID: empire.ID, PlanetID: homeworld.ID}
	empire.Capital = colony.ID
	homeworld.ColonyID = colony.ID
	s.Empires = append(s.Empires, empire)
	s.Colonies = append(s.Colonies, colony)
	s.AddEvent("game_started", "deterministic small fixture created")
	return s
}
