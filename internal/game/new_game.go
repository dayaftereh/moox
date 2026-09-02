package game

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"moox/internal/core"
	"moox/internal/protocol"
)

var ErrInvalidNewGameSettings = errors.New("invalid new game settings")

type GalaxySize string
type GalaxyAge string

const (
	GalaxySizeSmall GalaxySize = "small"
	GalaxyAgeNormal GalaxyAge  = "normal"
)

type NewGamePlayerSpec struct {
	SeatID     protocol.SeatID `json:"seat_id"`
	EmpireName string          `json:"empire_name"`
	RaceID     string          `json:"race_id"`
}

type NewGameSettings struct {
	GalaxySize      GalaxySize             `json:"galaxy_size"`
	GalaxyAge       GalaxyAge              `json:"galaxy_age"`
	TechnologyLevel NewGameTechnologyLevel `json:"technology_level"`
	StrategicCombat bool                   `json:"strategic_combat"`
	Players         []NewGamePlayerSpec    `json:"players"`
}

type NewGamePlayerResult struct {
	SeatID   protocol.SeatID `json:"seat_id"`
	EmpireID core.ID         `json:"empire_id"`
	RaceID   string          `json:"race_id"`
	Name     string          `json:"name"`
}

type NewGameResult struct {
	State   *core.GameState       `json:"state"`
	Players []NewGamePlayerResult `json:"players"`
}

type newGameSystemPlan struct {
	ID       core.ID
	Name     string
	X        int
	Y        int
	Spectral int
	Planets  []newGamePlanetPlan
}

type newGamePlanetPlan struct {
	Orbit     int
	SizeID    string
	MineralID string
	GravityID string
	ClimateID string
}

var newGameSizeIDs = []string{"tiny", "small", "medium", "large", "huge"}
var newGameSizeThresholds = []int{1, 3, 7, 9, 10}
var newGameMineralIDs = []string{"ultra_poor", "poor", "abundant", "rich", "ultra_rich"}
var newGameGravityIDs = []string{"low_g", "normal_g", "heavy_g"}
var newGameClimateIDs = []string{"toxic", "radiated", "barren", "desert", "tundra", "ocean", "swamp", "arid", "terran", "gaia"}

func (r *EconomyRules) NewGame(seed uint64, settings NewGameSettings) (NewGameResult, error) {
	if r == nil || r.NewGameGalaxy == nil {
		return NewGameResult{}, fmt.Errorf("new game galaxy rules are unavailable")
	}
	if err := r.validateNewGameSettings(settings); err != nil {
		return NewGameResult{}, fmt.Errorf("%w: %v", ErrInvalidNewGameSettings, err)
	}

	state := core.NewGameState(seed)
	rng := state.RNG()
	state.Galaxy.ID = state.NewID()

	players := make([]NewGamePlayerResult, len(settings.Players))
	for i, player := range settings.Players {
		empire := core.Empire{
			ID:         state.NewID(),
			Name:       strings.TrimSpace(player.EmpireName),
			RaceID:     player.RaceID,
			Freighters: r.NewGameGalaxy.Start.Freighters,
			Treasury:   core.EmpireTreasuryState{BalanceBC: r.NewGameGalaxy.Start.TreasuryBC},
		}
		state.Empires = append(state.Empires, empire)
		players[i] = NewGamePlayerResult{SeatID: player.SeatID, EmpireID: empire.ID, RaceID: empire.RaceID, Name: empire.Name}
	}

	if err := r.InitializeNewGameTechnologies(state, NewGameTechnologyStateOptions{
		Level:           settings.TechnologyLevel,
		StrategicCombat: settings.StrategicCombat,
		NewGameRNG:      rng,
	}); err != nil {
		return NewGameResult{}, fmt.Errorf("initialize new game technologies: %w", err)
	}

	systems, err := r.generateNewGameSystems(state, rng)
	if err != nil {
		return NewGameResult{}, err
	}
	homeIndexes := farthestNewGameSystemPair(systems)
	for _, index := range homeIndexes {
		if err := r.fillNewGameHomeSystem(&systems[index], rng); err != nil {
			return NewGameResult{}, err
		}
	}
	for _, index := range homeIndexes {
		sort.Slice(systems[index].Planets, func(i, j int) bool { return systems[index].Planets[i].Orbit < systems[index].Planets[j].Orbit })
		if len(systems[index].Planets) == 0 {
			return NewGameResult{}, fmt.Errorf("home system %d has no materialized planet", systems[index].ID)
		}
		home := &systems[index].Planets[0]
		home.SizeID = r.NewGameGalaxy.Homeworld.SizeID
		home.MineralID = r.NewGameGalaxy.Homeworld.MineralID
		home.GravityID = r.NewGameGalaxy.Homeworld.GravityID
		home.ClimateID = r.NewGameGalaxy.Homeworld.ClimateID
	}

	state.Galaxy.Systems = make([]core.StarSystem, len(systems))
	for si := range systems {
		plan := &systems[si]
		sort.Slice(plan.Planets, func(i, j int) bool { return plan.Planets[i].Orbit < plan.Planets[j].Orbit })
		system := core.StarSystem{ID: plan.ID, Name: plan.Name, X: plan.X, Y: plan.Y}
		for pi, planetPlan := range plan.Planets {
			if err := r.validateNewGamePlanetPlan(planetPlan); err != nil {
				return NewGameResult{}, fmt.Errorf("system %d planet %d: %w", plan.ID, pi, err)
			}
			system.Planets = append(system.Planets, core.Planet{
				ID:        state.NewID(),
				Name:      newGamePlanetName(plan.Name, planetPlan.Orbit),
				Orbit:     planetPlan.Orbit,
				SizeID:    planetPlan.SizeID,
				MineralID: planetPlan.MineralID,
				GravityID: planetPlan.GravityID,
				ClimateID: planetPlan.ClimateID,
			})
		}
		state.Galaxy.Systems[si] = system
	}

	for playerIndex, homeIndex := range homeIndexes {
		system := &state.Galaxy.Systems[homeIndex]
		if len(system.Planets) < r.NewGameGalaxy.Homeworld.MinimumPlanets {
			return NewGameResult{}, fmt.Errorf("home system %d has %d planets, expected at least %d", system.ID, len(system.Planets), r.NewGameGalaxy.Homeworld.MinimumPlanets)
		}
		homeworld := &system.Planets[0]
		empire := &state.Empires[playerIndex]
		colony := core.Colony{
			ID:         state.NewID(),
			EmpireID:   empire.ID,
			PlanetID:   homeworld.ID,
			Population: core.NewAssimilatedPopulation(empire.ID, r.NewGameGalaxy.Start.Farmers, r.NewGameGalaxy.Start.Workers, r.NewGameGalaxy.Start.Scientists),
		}
		empire.Capital = colony.ID
		homeworld.ColonyID = colony.ID
		state.Colonies = append(state.Colonies, colony)
	}

	scoutSpecs := make([]core.ShipDesignSpec, len(state.Empires))
	for i := range state.Empires {
		spec, err := r.newGameScoutSpec(&state.Empires[i])
		if err != nil {
			return NewGameResult{}, fmt.Errorf("create scout design for empire %d: %w", state.Empires[i].ID, err)
		}
		scoutSpecs[i] = spec
		state.ShipDesigns = append(state.ShipDesigns, core.ShipDesign{
			ID: state.NewID(), EmpireID: state.Empires[i].ID, Revision: 1, Name: "Scout", Spec: spec,
		})
	}

	for i := range state.Empires {
		empire := &state.Empires[i]
		design := &state.ShipDesigns[i]
		homeSystem := &state.Galaxy.Systems[homeIndexes[i]]
		shipIDs := make([]core.ID, 0, r.NewGameGalaxy.Start.ScoutCount)
		for n := 0; n < r.NewGameGalaxy.Start.ScoutCount; n++ {
			ship := core.Ship{
				ID: state.NewID(), EmpireID: empire.ID, SourceDesignID: design.ID, SourceDesignRevision: design.Revision,
				Name: fmt.Sprintf("Scout %d", n+1), Spec: scoutSpecs[i],
			}
			state.Ships = append(state.Ships, ship)
			shipIDs = append(shipIDs, ship.ID)
		}
		state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
			ID: state.NewID(), EmpireID: empire.ID, Role: core.StrategicFleetRoleCombat, AtSystemID: homeSystem.ID,
			ShipIDs: shipIDs,
		})
	}
	for i := range state.Empires {
		homeSystem := &state.Galaxy.Systems[homeIndexes[i]]
		state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
			ID: state.NewID(), EmpireID: state.Empires[i].ID, Role: core.StrategicFleetRoleCivilian,
			SpecialKind: core.StrategicFleetSpecialColonyShip, AtSystemID: homeSystem.ID, FTLSpeed: 2,
		})
	}

	resolver, err := NewEconomyResolver(r)
	if err != nil {
		return NewGameResult{}, err
	}
	for i := range state.Colonies {
		if err := resolver.recalculateColony(state, &state.Colonies[i]); err != nil {
			return NewGameResult{}, fmt.Errorf("initialize colony economy: %w", err)
		}
	}
	if _, err := resolver.materializeFoodLogistics(state, false); err != nil {
		return NewGameResult{}, fmt.Errorf("initialize food logistics: %w", err)
	}
	for i := range state.Empires {
		points, err := r.deriveEmpireCommandPoints(state, &state.Empires[i])
		if err != nil {
			return NewGameResult{}, fmt.Errorf("initialize command points: %w", err)
		}
		state.Empires[i].CommandPoints = points
	}

	state.AddEvent("game_started", "deterministic new game created")
	if err := state.Validate(); err != nil {
		return NewGameResult{}, fmt.Errorf("validate generated new game: %w", err)
	}
	state.CommitRNG(rng)
	return NewGameResult{State: state, Players: players}, nil
}

func (r *EconomyRules) validateNewGameSettings(settings NewGameSettings) error {
	if settings.GalaxySize != GalaxySizeSmall {
		return fmt.Errorf("unsupported galaxy_size %q; Slice 09 supports only %q", settings.GalaxySize, GalaxySizeSmall)
	}
	if settings.GalaxyAge != GalaxyAgeNormal {
		return fmt.Errorf("unsupported galaxy_age %q; Slice 09 supports only %q", settings.GalaxyAge, GalaxyAgeNormal)
	}
	if settings.TechnologyLevel != NewGameTechnologyAverage {
		return fmt.Errorf("unsupported technology_level %q; Slice 09 supports only %q", settings.TechnologyLevel, NewGameTechnologyAverage)
	}
	if settings.StrategicCombat {
		return fmt.Errorf("strategic_combat is not supported in Slice 09")
	}
	if len(settings.Players) != 2 {
		return fmt.Errorf("Slice 09 requires exactly two players, got %d", len(settings.Players))
	}
	seenSeats := map[protocol.SeatID]struct{}{}
	seenRaces := map[string]struct{}{}
	seenNames := map[string]struct{}{}
	previousSeat := protocol.SeatID(0)
	for i, player := range settings.Players {
		if player.SeatID == 0 {
			return fmt.Errorf("player[%d] seat_id must be non-zero", i)
		}
		if i > 0 && player.SeatID <= previousSeat {
			return fmt.Errorf("player seat IDs must be strictly ascending")
		}
		if _, exists := seenSeats[player.SeatID]; exists {
			return fmt.Errorf("duplicate seat_id %d", player.SeatID)
		}
		seenSeats[player.SeatID] = struct{}{}
		previousSeat = player.SeatID
		if player.RaceID != "human" && player.RaceID != "darlok" {
			return fmt.Errorf("player[%d] unsupported race_id %q", i, player.RaceID)
		}
		if _, exists := seenRaces[player.RaceID]; exists {
			return fmt.Errorf("race_id %q must appear exactly once", player.RaceID)
		}
		if _, ok := r.RaceModifiers[player.RaceID]; !ok {
			return fmt.Errorf("ruleset has no race %q", player.RaceID)
		}
		seenRaces[player.RaceID] = struct{}{}
		name := strings.TrimSpace(player.EmpireName)
		if name == "" {
			return fmt.Errorf("player[%d] empire_name must not be empty", i)
		}
		nameKey := strings.ToLower(name)
		if _, exists := seenNames[nameKey]; exists {
			return fmt.Errorf("empire names must be unique")
		}
		seenNames[nameKey] = struct{}{}
	}
	if _, ok := seenRaces["human"]; !ok {
		return fmt.Errorf("Slice 09 requires one human player")
	}
	if _, ok := seenRaces["darlok"]; !ok {
		return fmt.Errorf("Slice 09 requires one darlok player")
	}
	return nil
}

func (r *EconomyRules) generateNewGameSystems(state *core.GameState, rng *core.RNG) ([]newGameSystemPlan, error) {
	const stars = 20
	plans := make([]newGameSystemPlan, stars)
	spectral := make([]int, stars)
	weights := make([]int, 7)
	for i := range weights {
		weights[i] = r.NewGameGalaxy.SpectralWeights[i][1]
	}
	for i := 0; i < stars; i++ {
		value, err := weightedNewGameChoice(rng, weights)
		if err != nil {
			return nil, err
		}
		spectral[i] = value
	}
	for i := 0; i < stars; i++ {
		jx, err := rng.Intn(81)
		if err != nil {
			return nil, err
		}
		jy, err := rng.Intn(81)
		if err != nil {
			return nil, err
		}
		column, row := i%5, i/5
		plans[i] = newGameSystemPlan{
			ID: state.NewID(), Name: fmt.Sprintf("System %02d", i+1),
			X: 100 + 200*column + jx - 40, Y: 100 + 200*row + jy - 40, Spectral: spectral[i],
		}
	}
	for i := range plans {
		if err := r.generateNewGamePlanets(&plans[i], rng); err != nil {
			return nil, fmt.Errorf("generate planets for system %d: %w", plans[i].ID, err)
		}
	}
	return plans, nil
}

func (r *EconomyRules) generateNewGamePlanets(system *newGameSystemPlan, rng *core.RNG) error {
	if system.Spectral < 0 || system.Spectral > 6 {
		return fmt.Errorf("invalid spectral class %d", system.Spectral)
	}
	if system.Spectral == 6 {
		return nil
	}
	roll, err := rng.Intn(10)
	if err != nil {
		return err
	}
	count := r.NewGameGalaxy.SatelliteCounts[roll][system.Spectral]
	unused := []int{0, 1, 2, 3, 4}
	for body := 0; body < count; body++ {
		choice, err := rng.Intn(len(unused))
		if err != nil {
			return err
		}
		orbit := unused[choice]
		unused = append(unused[:choice], unused[choice+1:]...)
		bodyType, err := r.newGameBodyType(rng, orbit)
		if err != nil {
			return err
		}
		if bodyType != 3 {
			continue
		}
		planet, err := r.newGamePlanetProperties(rng, system.Spectral, orbit)
		if err != nil {
			return err
		}
		system.Planets = append(system.Planets, planet)
	}
	sort.Slice(system.Planets, func(i, j int) bool { return system.Planets[i].Orbit < system.Planets[j].Orbit })
	return nil
}

func (r *EconomyRules) newGameBodyType(rng *core.RNG, orbit int) (int, error) {
	for {
		roll, err := rng.Intn(10)
		if err != nil {
			return 0, err
		}
		bodyType := r.NewGameGalaxy.BodyTypes[roll][orbit]
		if bodyType == 4 {
			// Original Generate_Satellite_Type_ consumes Random(100) for type 4.
			// Wormhole/special materialization is deferred, so the result remains non-colonizable here.
			if _, err := rng.Intn(100); err != nil {
				return 0, err
			}
			bodyType = 1
		}
		if bodyType == 2 && orbit == 0 {
			continue
		}
		return bodyType, nil
	}
}

func (r *EconomyRules) newGamePlanetProperties(rng *core.RNG, spectral, orbit int) (newGamePlanetPlan, error) {
	if spectral < 0 {
		return newGamePlanetPlan{}, fmt.Errorf("invalid spectral class %d", spectral)
	}
	// Original normal-body tables contain six spectral columns. Class 6 is a special
	// object class with no normal satellite path; Homeworld underflow fill uses the
	// nearest normal column while still consuming the shared property RNG sequence.
	if spectral > 5 {
		spectral = 5
	}
	sizeRoll, err := rng.Intn(10)
	if err != nil {
		return newGamePlanetPlan{}, err
	}
	sizeRoll++
	sizeIndex := 0
	for sizeIndex < len(r.NewGameGalaxy.SizeRollUpperThresholds)-1 && sizeRoll > r.NewGameGalaxy.SizeRollUpperThresholds[sizeIndex] {
		sizeIndex++
	}
	mineralRoll, err := rng.Intn(10)
	if err != nil {
		return newGamePlanetPlan{}, err
	}
	mineralIndex := r.NewGameGalaxy.MineralClasses[mineralRoll][spectral]
	gravityIndex := r.NewGameGalaxy.GravityByMineralSize[mineralIndex][sizeIndex]
	group := r.NewGameGalaxy.PlanetGroupBySpectralOrbit[spectral][orbit]
	climateIndex, err := weightedNewGameChoice(rng, r.NewGameGalaxy.NormalClimateWeights[group])
	if err != nil {
		return newGamePlanetPlan{}, err
	}
	if _, err := rng.Intn(3); err != nil {
		return newGamePlanetPlan{}, err
	}
	return newGamePlanetPlan{Orbit: orbit, SizeID: newGameSizeIDs[sizeIndex], MineralID: newGameMineralIDs[mineralIndex], GravityID: newGameGravityIDs[gravityIndex], ClimateID: newGameClimateIDs[climateIndex]}, nil
}

func (r *EconomyRules) fillNewGameHomeSystem(system *newGameSystemPlan, rng *core.RNG) error {
	for len(system.Planets) < r.NewGameGalaxy.Homeworld.MinimumPlanets {
		used := make(map[int]struct{}, len(system.Planets))
		for _, planet := range system.Planets {
			used[planet.Orbit] = struct{}{}
		}
		orbit := -1
		for candidate := 0; candidate < 5; candidate++ {
			if _, ok := used[candidate]; !ok {
				orbit = candidate
				break
			}
		}
		if orbit < 0 {
			return fmt.Errorf("cannot fill home system %d to %d planets", system.ID, r.NewGameGalaxy.Homeworld.MinimumPlanets)
		}
		planet, err := r.newGamePlanetProperties(rng, system.Spectral, orbit)
		if err != nil {
			return err
		}
		system.Planets = append(system.Planets, planet)
		sort.Slice(system.Planets, func(i, j int) bool { return system.Planets[i].Orbit < system.Planets[j].Orbit })
	}
	return nil
}

func farthestNewGameSystemPair(systems []newGameSystemPlan) [2]int {
	best := [2]int{0, 1}
	bestDistance := int64(-1)
	for i := 0; i < len(systems); i++ {
		for j := i + 1; j < len(systems); j++ {
			dx := int64(systems[i].X - systems[j].X)
			dy := int64(systems[i].Y - systems[j].Y)
			distance := dx*dx + dy*dy
			if distance > bestDistance || (distance == bestDistance && (systems[i].ID < systems[best[0]].ID || (systems[i].ID == systems[best[0]].ID && systems[j].ID < systems[best[1]].ID))) {
				best = [2]int{i, j}
				bestDistance = distance
			}
		}
	}
	return best
}

func weightedNewGameChoice(rng *core.RNG, weights []int) (int, error) {
	total := 0
	for _, weight := range weights {
		if weight < 0 {
			return 0, fmt.Errorf("negative weight %d", weight)
		}
		total += weight
	}
	if total <= 0 {
		return 0, fmt.Errorf("weighted choice has no positive weight")
	}
	roll, err := rng.Intn(total)
	if err != nil {
		return 0, err
	}
	for i, weight := range weights {
		if roll < weight {
			return i, nil
		}
		roll -= weight
	}
	return 0, fmt.Errorf("weighted choice exhausted total %d", total)
}

func (r *EconomyRules) validateNewGamePlanetPlan(plan newGamePlanetPlan) error {
	if plan.Orbit < 0 || plan.Orbit > 4 {
		return fmt.Errorf("orbit %d outside 0..4", plan.Orbit)
	}
	if _, ok := r.PopulationSizeClass[plan.SizeID]; !ok {
		return fmt.Errorf("unknown size_id %q", plan.SizeID)
	}
	if _, ok := r.MineralIndustryPerWorker[plan.MineralID]; !ok {
		return fmt.Errorf("unknown mineral_id %q", plan.MineralID)
	}
	switch plan.GravityID {
	case "low_g", "normal_g", "heavy_g":
	default:
		return fmt.Errorf("unknown gravity_id %q", plan.GravityID)
	}
	if _, ok := r.ClimateFoodPerFarmer[plan.ClimateID]; !ok {
		return fmt.Errorf("unknown climate_id %q", plan.ClimateID)
	}
	return nil
}

func (r *EconomyRules) newGameScoutSpec(empire *core.Empire) (core.ShipDesignSpec, error) {
	if empire == nil {
		return core.ShipDesignSpec{}, fmt.Errorf("empire must not be nil")
	}
	hull, ok := r.ShipHulls["frigate"]
	if !ok {
		return core.ShipDesignSpec{}, fmt.Errorf("ruleset has no frigate hull")
	}
	if hull.SizeIndex != 0 || len(hull.StrategicPictureIDs) == 0 {
		return core.ShipDesignSpec{}, fmt.Errorf("frigate hull baseline is invalid")
	}
	drive, ok := newGameDriveByID(r, "nuclear_drive")
	if !ok {
		return core.ShipDesignSpec{}, fmt.Errorf("ruleset has no nuclear_drive")
	}
	computer, ok := newGameComputerByID(r, "electronic_computer")
	if !ok {
		return core.ShipDesignSpec{}, fmt.Errorf("ruleset has no electronic_computer")
	}
	armor, ok := newGameArmorByID(r, "titanium_armor")
	if !ok {
		return core.ShipDesignSpec{}, fmt.Errorf("ruleset has no titanium_armor")
	}
	fuel, ok := newGameFuelByID(r, "standard_fuel_cells")
	if !ok {
		return core.ShipDesignSpec{}, fmt.Errorf("ruleset has no standard_fuel_cells")
	}
	for _, technologyID := range []int{drive.TechnologyID, computer.TechnologyID, armor.TechnologyID, fuel.TechnologyID} {
		if !empireKnowsTechnology(empire, technologyID) {
			return core.ShipDesignSpec{}, fmt.Errorf("empire %d lacks Scout component technology %d", empire.ID, technologyID)
		}
	}
	index := hull.SizeIndex
	baseCost := hull.BaseCostPP + drive.CostByHullPP[index] + computer.CostByHullPP[index] + (hull.BaseCostPP*armor.CostPercent)/100
	spaceUsed := drive.SpaceByHull[index]
	return core.ShipDesignSpec{HullID: hull.ID, StrategicPictureID: hull.StrategicPictureIDs[0], WarpDriveID: drive.ID, FTLSpeed: drive.FTLSpeed, ComputerID: computer.ID, ArmorID: armor.ID, FuelCellID: fuel.ID, FuelRangeParsecs: fuel.RangeParsecs, HullBaseCostPP: hull.BaseCostPP, HullSpace: hull.BaseSpace, SpaceUsed: spaceUsed, BaseDesignCostPP: baseCost, ProductionCostPP: militaryShipProductionCostPP(empire, baseCost, r.RaceModifiers)}, nil
}

func newGameDriveByID(r *EconomyRules, id string) (drive struct {
	ID                        string
	TechnologyID, FTLSpeed    int
	SpaceByHull, CostByHullPP []int
}, ok bool) {
	for _, item := range r.ShipDrives {
		if item.ID == id {
			return struct {
				ID                        string
				TechnologyID, FTLSpeed    int
				SpaceByHull, CostByHullPP []int
			}{item.ID, item.TechnologyID, item.FTLSpeed, item.SpaceByHull, item.CostByHullPP}, true
		}
	}
	return drive, false
}

func newGameComputerByID(r *EconomyRules, id string) (computer struct {
	ID           string
	TechnologyID int
	CostByHullPP []int
}, ok bool) {
	for _, item := range r.ShipComputers {
		if item.ID == id {
			return struct {
				ID           string
				TechnologyID int
				CostByHullPP []int
			}{item.ID, item.TechnologyID, item.CostByHullPP}, true
		}
	}
	return computer, false
}

func newGameArmorByID(r *EconomyRules, id string) (armor struct {
	ID                        string
	TechnologyID, CostPercent int
}, ok bool) {
	for _, item := range r.ShipArmors {
		if item.ID == id {
			return struct {
				ID                        string
				TechnologyID, CostPercent int
			}{item.ID, item.TechnologyID, item.CostPercent}, true
		}
	}
	return armor, false
}

func newGameFuelByID(r *EconomyRules, id string) (fuel struct {
	ID                         string
	TechnologyID, RangeParsecs int
}, ok bool) {
	for _, item := range r.ShipFuelCells {
		if item.ID == id {
			return struct {
				ID                         string
				TechnologyID, RangeParsecs int
			}{item.ID, item.TechnologyID, item.RangeParsecs}, true
		}
	}
	return fuel, false
}

func newGamePlanetName(system string, orbit int) string {
	roman := []string{"I", "II", "III", "IV", "V"}
	if orbit >= 0 && orbit < len(roman) {
		return system + " " + roman[orbit]
	}
	return fmt.Sprintf("%s %d", system, orbit+1)
}
