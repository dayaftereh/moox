package game

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"moox/internal/core"
	"moox/internal/protocol"
	"moox/internal/ruleset"
)

var ErrInvalidNewGameSettings = errors.New("invalid new game settings")

type GalaxySize string
type GalaxyAge string

const (
	GalaxySizeSmall  GalaxySize = "small"
	GalaxySizeMedium GalaxySize = "medium"
	GalaxySizeLarge  GalaxySize = "large"
	GalaxySizeHuge   GalaxySize = "huge"

	GalaxyAgeMineralRich GalaxyAge = "mineral_rich"
	GalaxyAgeNormal      GalaxyAge = "normal"
	GalaxyAgeOrganicRich GalaxyAge = "organic_rich"
)

var SupportedGalaxySizeIDs = []GalaxySize{GalaxySizeSmall, GalaxySizeMedium, GalaxySizeLarge, GalaxySizeHuge}
var SupportedGalaxyAgeIDs = []GalaxyAge{GalaxyAgeMineralRich, GalaxyAgeNormal, GalaxyAgeOrganicRich}

type NewGamePlayerSpec struct {
	SeatID              protocol.SeatID `json:"seat_id"`
	EmpireName          string          `json:"empire_name"`
	RaceID              string          `json:"race_id"`
	BuiltinAIControlled bool            `json:"-"`
}

type NewGameSettings struct {
	DifficultyID    core.DifficultyID      `json:"difficulty_id,omitempty"`
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
	Bodies   []newGameBodyPlan
}

type newGameBodyPlan struct {
	Orbit int
	Kind  core.OrbitalBodyKind
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
	if settings.DifficultyID == "" {
		settings.DifficultyID = core.DifficultyNormal
	}
	if r == nil || r.NewGameGalaxy == nil {
		return NewGameResult{}, fmt.Errorf("new game galaxy rules are unavailable")
	}
	if err := r.validateNewGameSettings(settings); err != nil {
		return NewGameResult{}, fmt.Errorf("%w: %v", ErrInvalidNewGameSettings, err)
	}

	state := core.NewGameState(seed)
	state.DifficultyID = settings.DifficultyID
	rng := state.RNG()
	state.Galaxy.ID = state.NewID()

	players := make([]NewGamePlayerResult, len(settings.Players))
	for i, player := range settings.Players {
		empire := core.Empire{
			ID:                  state.NewID(),
			Name:                strings.TrimSpace(player.EmpireName),
			BuiltinAIControlled: player.BuiltinAIControlled,
			RaceID:              player.RaceID,
			PlayerColorSlot:     i + 1,
			Freighters:          r.NewGameGalaxy.Start.Freighters,
			Treasury:            core.EmpireTreasuryState{BalanceBC: r.NewGameGalaxy.Start.TreasuryBC},
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

	sizeProfile, err := r.newGameGalaxySizeProfile(settings.GalaxySize)
	if err != nil {
		return NewGameResult{}, err
	}
	ageProfile, err := r.newGameGalaxyAgeProfile(settings.GalaxyAge)
	if err != nil {
		return NewGameResult{}, err
	}
	systems, err := r.generateNewGameSystems(state, rng, sizeProfile, ageProfile)
	if err != nil {
		return NewGameResult{}, err
	}
	homeIndexes := farthestNewGameSystemPair(systems)
	for _, index := range homeIndexes {
		if err := r.fillNewGameHomeSystem(&systems[index], rng, ageProfile); err != nil {
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

	if err := r.materializeNewGameSystems(state, systems); err != nil {
		return NewGameResult{}, err
	}

	if err := r.initializeNewGameStartingAssets(state, homeIndexes[:]); err != nil {
		return NewGameResult{}, err
	}
	// Allocate non-planet body IDs only after every legacy New Game object has
	// received its ID. This preserves the established deterministic IDs for
	// empires, planets, colonies, designs, ships and fleets.
	if err := materializeNewGameBodies(state, systems); err != nil {
		return NewGameResult{}, err
	}

	if err := r.finalizeNewGameEconomy(state); err != nil {
		return NewGameResult{}, err
	}

	state.AddEvent("game_started", "deterministic new game created")
	if err := state.Validate(); err != nil {
		return NewGameResult{}, fmt.Errorf("validate generated new game: %w", err)
	}
	state.CommitRNG(rng)
	return NewGameResult{State: state, Players: players}, nil
}

func (r *EconomyRules) materializeNewGameSystems(state *core.GameState, systems []newGameSystemPlan) error {
	state.Galaxy.Systems = make([]core.StarSystem, len(systems))
	for si := range systems {
		plan := &systems[si]
		sort.Slice(plan.Planets, func(i, j int) bool { return plan.Planets[i].Orbit < plan.Planets[j].Orbit })
		system := core.StarSystem{ID: plan.ID, Name: plan.Name, X: plan.X, Y: plan.Y, SpectralClass: plan.Spectral}
		for pi, planetPlan := range plan.Planets {
			if err := r.validateNewGamePlanetPlan(planetPlan); err != nil {
				return fmt.Errorf("system %d planet %d: %w", plan.ID, pi, err)
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
	return nil
}

func (r *EconomyRules) initializeNewGameStartingAssets(state *core.GameState, homeIndexes []int) error {
	if len(homeIndexes) != len(state.Empires) {
		return fmt.Errorf("home system count %d does not match empire count %d", len(homeIndexes), len(state.Empires))
	}
	for playerIndex, homeIndex := range homeIndexes {
		if homeIndex < 0 || homeIndex >= len(state.Galaxy.Systems) {
			return fmt.Errorf("empire %d home system index %d is out of range", state.Empires[playerIndex].ID, homeIndex)
		}
		system := &state.Galaxy.Systems[homeIndex]
		if len(system.Planets) < r.NewGameGalaxy.Homeworld.MinimumPlanets {
			return fmt.Errorf("home system %d has %d planets, expected at least %d", system.ID, len(system.Planets), r.NewGameGalaxy.Homeworld.MinimumPlanets)
		}
		homeworld := &system.Planets[0]
		empire := &state.Empires[playerIndex]
		empire.MarkSystemVisited(system.ID)
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
			return fmt.Errorf("create scout design for empire %d: %w", state.Empires[i].ID, err)
		}
		scoutSpecs[i] = spec
		designID := state.NewID()
		visual := generateStartingScoutVisualGenome(fmt.Sprintf("new-game:%016x:empire:%d:design:%d:scout:v4", state.Seed, state.Empires[i].ID, designID))
		if err := core.ValidateShipVisualGenome(visual); err != nil {
			return fmt.Errorf("create scout visual genome for empire %d: %w", state.Empires[i].ID, err)
		}
		state.ShipDesigns = append(state.ShipDesigns, core.ShipDesign{
			ID: designID, EmpireID: state.Empires[i].ID, Revision: 1, VisualRevision: 1, Name: "Scout", Spec: spec,
			VisualGenome: func() *core.ShipVisualGenome { clone := core.CloneShipVisualGenome(visual); return &clone }(),
		})
	}

	for i := range state.Empires {
		empire := &state.Empires[i]
		design := &state.ShipDesigns[i]
		homeSystem := &state.Galaxy.Systems[homeIndexes[i]]
		shipIDs := make([]core.ID, 0, r.NewGameGalaxy.Start.ScoutCount)
		for n := 0; n < r.NewGameGalaxy.Start.ScoutCount; n++ {
			var visual *core.ShipVisualGenome
			if design.VisualGenome != nil {
				clone := core.CloneShipVisualGenome(*design.VisualGenome)
				visual = &clone
			}
			ship := core.Ship{
				ID: state.NewID(), EmpireID: empire.ID, SourceDesignID: design.ID, SourceDesignRevision: design.Revision, SourceVisualRevision: design.VisualRevision,
				Name: fmt.Sprintf("Scout %d", n+1), Spec: scoutSpecs[i], VisualGenome: visual,
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
	return nil
}
func (r *EconomyRules) finalizeNewGameEconomy(state *core.GameState) error {
	resolver, err := NewEconomyResolver(r)
	if err != nil {
		return err
	}
	for i := range state.Colonies {
		if err := resolver.recalculateColony(state, &state.Colonies[i]); err != nil {
			return fmt.Errorf("initialize colony economy: %w", err)
		}
	}
	if _, err := resolver.materializeFoodLogistics(state, false); err != nil {
		return fmt.Errorf("initialize food logistics: %w", err)
	}
	for i := range state.Empires {
		points, err := r.deriveEmpireCommandPoints(state, &state.Empires[i])
		if err != nil {
			return fmt.Errorf("initialize command points: %w", err)
		}
		state.Empires[i].CommandPoints = points
	}
	return nil
}
func (r *EconomyRules) validateNewGameSettings(settings NewGameSettings) error {
	if !core.IsSupportedDifficultyID(settings.DifficultyID) {
		return fmt.Errorf("unsupported difficulty_id %q", settings.DifficultyID)
	}
	if _, err := r.newGameGalaxySizeProfile(settings.GalaxySize); err != nil {
		return err
	}
	if _, err := r.newGameGalaxyAgeProfile(settings.GalaxyAge); err != nil {
		return err
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

func (r *EconomyRules) newGameGalaxySizeProfile(id GalaxySize) (*ruleset.NewGameGalaxySize, error) {
	for i := range r.NewGameGalaxy.GalaxySizes {
		if r.NewGameGalaxy.GalaxySizes[i].ID == string(id) {
			return &r.NewGameGalaxy.GalaxySizes[i], nil
		}
	}
	return nil, fmt.Errorf("unsupported galaxy_size %q", id)
}

func (r *EconomyRules) newGameGalaxyAgeProfile(id GalaxyAge) (*ruleset.NewGameGalaxyAge, error) {
	for i := range r.NewGameGalaxy.GalaxyAges {
		if r.NewGameGalaxy.GalaxyAges[i].ID == string(id) {
			return &r.NewGameGalaxy.GalaxyAges[i], nil
		}
	}
	return nil, fmt.Errorf("unsupported galaxy_age %q", id)
}

func (r *EconomyRules) generateNewGameSystems(state *core.GameState, rng *core.RNG, size *ruleset.NewGameGalaxySize, age *ruleset.NewGameGalaxyAge) ([]newGameSystemPlan, error) {
	if size == nil || age == nil {
		return nil, fmt.Errorf("new game galaxy size/age profile is required")
	}
	plans := make([]newGameSystemPlan, size.Stars)
	spectral := make([]int, size.Stars)
	for i := 0; i < size.Stars; i++ {
		value, err := weightedNewGameChoice(rng, age.SpectralWeights)
		if err != nil {
			return nil, err
		}
		spectral[i] = value
	}
	for i := 0; i < size.Stars; i++ {
		jx, err := rng.Intn(81)
		if err != nil {
			return nil, err
		}
		jy, err := rng.Intn(81)
		if err != nil {
			return nil, err
		}
		column, row := i%size.GridColumns, i/size.GridColumns
		plans[i] = newGameSystemPlan{
			ID: state.NewID(), Name: fmt.Sprintf("System %02d", i+1),
			X: 100 + 200*column + jx - 40, Y: 100 + 200*row + jy - 40, Spectral: spectral[i],
		}
	}
	for i := range plans {
		if err := r.generateNewGamePlanets(&plans[i], rng, age); err != nil {
			return nil, fmt.Errorf("generate planets for system %d: %w", plans[i].ID, err)
		}
	}
	return plans, nil
}

func (r *EconomyRules) generateNewGamePlanets(system *newGameSystemPlan, rng *core.RNG, age *ruleset.NewGameGalaxyAge) error {
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
		kind, err := newGameOrbitalBodyKind(bodyType)
		if err != nil {
			return err
		}
		system.Bodies = append(system.Bodies, newGameBodyPlan{Orbit: orbit, Kind: kind})
		if kind == core.OrbitalBodyPlanet {
			planet, err := r.newGamePlanetProperties(rng, system.Spectral, orbit, age)
			if err != nil {
				return err
			}
			system.Planets = append(system.Planets, planet)
		}
	}
	sort.Slice(system.Planets, func(i, j int) bool { return system.Planets[i].Orbit < system.Planets[j].Orbit })
	return nil
}

func newGameOrbitalBodyKind(bodyType int) (core.OrbitalBodyKind, error) {
	switch bodyType {
	case 1:
		return core.OrbitalBodyAsteroidBelt, nil
	case 2:
		return core.OrbitalBodyGasGiant, nil
	case 3:
		return core.OrbitalBodyPlanet, nil
	default:
		return "", fmt.Errorf("unsupported materialized satellite body type %d", bodyType)
	}
}

func materializeNewGameBodies(state *core.GameState, plans []newGameSystemPlan) error {
	if state == nil {
		return fmt.Errorf("game state must not be nil")
	}
	if len(state.Galaxy.Systems) != len(plans) {
		return fmt.Errorf("system/body plan count mismatch: state=%d plans=%d", len(state.Galaxy.Systems), len(plans))
	}
	for si := range plans {
		system := &state.Galaxy.Systems[si]
		plan := &plans[si]
		sort.Slice(plan.Bodies, func(i, j int) bool { return plan.Bodies[i].Orbit < plan.Bodies[j].Orbit })
		if len(plan.Bodies) == 0 {
			system.Bodies = nil
			continue
		}
		system.Bodies = make([]core.OrbitalBody, 0, len(plan.Bodies))
		for _, bodyPlan := range plan.Bodies {
			body := core.OrbitalBody{Orbit: bodyPlan.Orbit, Kind: bodyPlan.Kind}
			switch bodyPlan.Kind {
			case core.OrbitalBodyPlanet:
				var planet *core.Planet
				for pi := range system.Planets {
					if system.Planets[pi].Orbit == bodyPlan.Orbit {
						planet = &system.Planets[pi]
						break
					}
				}
				if planet == nil {
					return fmt.Errorf("system %d orbit %d planet body has no materialized planet", system.ID, bodyPlan.Orbit)
				}
				body.ID = planet.ID
				body.PlanetID = planet.ID
				body.Name = planet.Name
			case core.OrbitalBodyAsteroidBelt:
				body.ID = state.NewID()
				body.Name = fmt.Sprintf("%s Asteroid Belt %d", system.Name, bodyPlan.Orbit+1)
			case core.OrbitalBodyGasGiant:
				body.ID = state.NewID()
				body.Name = fmt.Sprintf("%s Gas Giant %d", system.Name, bodyPlan.Orbit+1)
			default:
				return fmt.Errorf("system %d orbit %d has unsupported body kind %q", system.ID, bodyPlan.Orbit, bodyPlan.Kind)
			}
			system.Bodies = append(system.Bodies, body)
		}
	}
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

func (r *EconomyRules) newGamePlanetProperties(rng *core.RNG, spectral, orbit int, age *ruleset.NewGameGalaxyAge) (newGamePlanetPlan, error) {
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
	if age == nil {
		return newGamePlanetPlan{}, fmt.Errorf("new game galaxy age profile is required")
	}
	climateIndex, err := weightedNewGameChoice(rng, age.ClimateWeights[group])
	if err != nil {
		return newGamePlanetPlan{}, err
	}
	if _, err := rng.Intn(3); err != nil {
		return newGamePlanetPlan{}, err
	}
	return newGamePlanetPlan{Orbit: orbit, SizeID: newGameSizeIDs[sizeIndex], MineralID: newGameMineralIDs[mineralIndex], GravityID: newGameGravityIDs[gravityIndex], ClimateID: newGameClimateIDs[climateIndex]}, nil
}

func (r *EconomyRules) fillNewGameHomeSystem(system *newGameSystemPlan, rng *core.RNG, age *ruleset.NewGameGalaxyAge) error {
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
		planet, err := r.newGamePlanetProperties(rng, system.Spectral, orbit, age)
		if err != nil {
			return err
		}
		system.Planets = append(system.Planets, planet)
		replaced := false
		for i := range system.Bodies {
			if system.Bodies[i].Orbit == orbit {
				system.Bodies[i].Kind = core.OrbitalBodyPlanet
				replaced = true
				break
			}
		}
		if !replaced {
			system.Bodies = append(system.Bodies, newGameBodyPlan{Orbit: orbit, Kind: core.OrbitalBodyPlanet})
		}
		sort.Slice(system.Planets, func(i, j int) bool { return system.Planets[i].Orbit < system.Planets[j].Orbit })
		sort.Slice(system.Bodies, func(i, j int) bool { return system.Bodies[i].Orbit < system.Bodies[j].Orbit })
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
