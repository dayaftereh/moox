package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func canonicalNewGameSettings() NewGameSettings {
	return NewGameSettings{
		GalaxySize:      GalaxySizeSmall,
		GalaxyAge:       GalaxyAgeNormal,
		TechnologyLevel: NewGameTechnologyAverage,
		Players: []NewGamePlayerSpec{
			{SeatID: protocol.SeatID(1), EmpireName: "Human", RaceID: "human"},
			{SeatID: protocol.SeatID(2), EmpireName: "Darlok", RaceID: "darlok"},
		},
	}
}

func TestNewGameGoldenSeedStateFingerprint(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	result, err := rules.NewGame(0x8009, canonicalNewGameSettings())
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(result.State)
	if err != nil {
		t.Fatal(err)
	}
	got := fmt.Sprintf("%x", sha256.Sum256(data))
	const want = "ad89bf90bd8b111de9e28ed1e4fcb253cb9c695e7b861f9d3d44f463b77351f9"
	if got != want {
		t.Fatalf("golden seed state sha256=%s want=%s", got, want)
	}
}
func TestNewGameDeterministicAndRoundTrips(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	settings := canonicalNewGameSettings()
	first, err := rules.NewGame(0x8009, settings)
	if err != nil {
		t.Fatal(err)
	}
	second, err := rules.NewGame(0x8009, settings)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("same seed/settings did not produce an identical NewGameResult")
	}
	if got := []int{first.State.Empires[0].PlayerColorSlot, first.State.Empires[1].PlayerColorSlot}; !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("player color slots=%v want [1 2]", got)
	}

	data, err := json.Marshal(first.State)
	if err != nil {
		t.Fatal(err)
	}
	secondData, err := json.Marshal(second.State)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, secondData) {
		t.Fatal("same seed/settings did not produce byte-identical JSON state")
	}
	var loaded core.GameState
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}
	if err := loaded.Validate(); err != nil {
		t.Fatalf("round-tripped state invalid: %v", err)
	}
	if !reflect.DeepEqual(first.State, &loaded) {
		t.Fatal("state changed across JSON round trip")
	}

	different, err := rules.NewGame(0x800A, settings)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(first.State.Galaxy, different.State.Galaxy) || first.State.RNGState == different.State.RNGState {
		t.Fatal("different seed did not change generated galaxy/RNG state")
	}
}

func TestNewGameCanonicalBaselineInvariants(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	result, err := rules.NewGame(0x8009, canonicalNewGameSettings())
	if err != nil {
		t.Fatal(err)
	}
	state := result.State
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
	if state.Galaxy.ID != 1 {
		t.Fatalf("galaxy id=%d want=1", state.Galaxy.ID)
	}
	if len(state.Empires) != 2 || state.Empires[0].ID != 2 || state.Empires[1].ID != 3 {
		t.Fatalf("empire IDs=%v", []core.ID{state.Empires[0].ID, state.Empires[1].ID})
	}
	if len(state.Galaxy.Systems) != 20 {
		t.Fatalf("systems=%d want=20", len(state.Galaxy.Systems))
	}
	coords := map[[2]int]struct{}{}
	for i, system := range state.Galaxy.Systems {
		if want := core.ID(4 + i); system.ID != want {
			t.Fatalf("system[%d] id=%d want=%d", i, system.ID, want)
		}
		if system.X < 60 || system.X > 940 || system.Y < 60 || system.Y > 740 {
			t.Fatalf("system %d coords=(%d,%d) outside envelope", system.ID, system.X, system.Y)
		}
		key := [2]int{system.X, system.Y}
		if _, exists := coords[key]; exists {
			t.Fatalf("duplicate coordinates %v", key)
		}
		coords[key] = struct{}{}
		lastOrbit := -1
		for _, planet := range system.Planets {
			if planet.Orbit <= lastOrbit || planet.Orbit < 0 || planet.Orbit > 4 {
				t.Fatalf("system %d planet orbits not sorted/unique", system.ID)
			}
			lastOrbit = planet.Orbit
		}
	}

	homeSystems := make([]*core.StarSystem, 2)
	for i := range state.Empires {
		empire := &state.Empires[i]
		if empire.Treasury.BalanceBC != 50 || empire.Freighters != 0 {
			t.Fatalf("empire %d start treasury/freighters=%v/%d", empire.ID, empire.Treasury.BalanceBC, empire.Freighters)
		}
		wantTech := []int{32, 40, 41, 58, 63, 69, 100, 101, 103, 109, 119, 120, 121, 145, 157, 166, 167, 168, 187, 189}
		if !reflect.DeepEqual(empire.KnownTechnologyIDs, wantTech) {
			t.Fatalf("empire %d tech=%v want=%v", empire.ID, empire.KnownTechnologyIDs, wantTech)
		}
		colony := colonyByID(state, empire.Capital)
		if colony == nil {
			t.Fatalf("empire %d has no capital colony", empire.ID)
		}
		if colony.Population.Total() != 8 || colony.Population.Farmers() != 4 || colony.Population.Workers() != 2 || colony.Population.Scientists() != 2 {
			t.Fatalf("empire %d population=%+v", empire.ID, colony.Population)
		}
		if len(colony.Buildings) != 0 {
			t.Fatalf("empire %d starts with buildings %v", empire.ID, colony.Buildings)
		}
		planet := planetByID(state, colony.PlanetID)
		if planet == nil {
			t.Fatalf("capital planet missing")
		}
		if planet.SizeID != "medium" || planet.MineralID != "abundant" || planet.GravityID != "normal_g" || planet.ClimateID != "terran" {
			t.Fatalf("homeworld=%+v", *planet)
		}
		for si := range state.Galaxy.Systems {
			for pi := range state.Galaxy.Systems[si].Planets {
				if state.Galaxy.Systems[si].Planets[pi].ID == planet.ID {
					homeSystems[i] = &state.Galaxy.Systems[si]
				}
			}
		}
		if homeSystems[i] == nil || len(homeSystems[i].Planets) < 3 {
			t.Fatalf("empire %d home system has <3 planets", empire.ID)
		}
		if planet.Orbit != homeSystems[i].Planets[0].Orbit {
			t.Fatalf("homeworld is not lowest orbit")
		}
	}
	if homeSystems[0].ID >= homeSystems[1].ID {
		t.Fatalf("player home IDs not ascending: %d,%d", homeSystems[0].ID, homeSystems[1].ID)
	}
	maxDistance := int64(-1)
	var best [2]core.ID
	for i := 0; i < len(state.Galaxy.Systems); i++ {
		for j := i + 1; j < len(state.Galaxy.Systems); j++ {
			a, b := state.Galaxy.Systems[i], state.Galaxy.Systems[j]
			dx, dy := int64(a.X-b.X), int64(a.Y-b.Y)
			d := dx*dx + dy*dy
			pair := [2]core.ID{a.ID, b.ID}
			if d > maxDistance || (d == maxDistance && (pair[0] < best[0] || (pair[0] == best[0] && pair[1] < best[1]))) {
				maxDistance = d
				best = pair
			}
		}
	}
	if got := [2]core.ID{homeSystems[0].ID, homeSystems[1].ID}; got != best {
		t.Fatalf("home pair=%v want farthest=%v", got, best)
	}

	if len(state.ShipDesigns) != 2 || len(state.Ships) != 4 || len(state.StrategicFleets) != 4 {
		t.Fatalf("designs/ships/fleets=%d/%d/%d", len(state.ShipDesigns), len(state.Ships), len(state.StrategicFleets))
	}
	for i, empire := range state.Empires {
		design := state.ShipDesigns[i]
		if design.EmpireID != empire.ID || design.Name != "Scout" || design.Spec.HullID != "frigate" || design.Spec.WarpDriveID != "nuclear_drive" || design.Spec.ComputerID != "electronic_computer" || design.Spec.ArmorID != "titanium_armor" || design.Spec.FuelCellID != "standard_fuel_cells" || design.Spec.ShieldID != "" || len(design.Spec.Weapons) != 0 {
			t.Fatalf("empire %d scout design=%+v", empire.ID, design)
		}
		if design.VisualRevision != 1 || design.VisualGenome == nil || design.VisualGenome.HullID != "scout" {
			t.Fatalf("empire %d scout visual revision/genome=%d/%+v", empire.ID, design.VisualRevision, design.VisualGenome)
		}
		if err := core.ValidateShipVisualGenome(*design.VisualGenome); err != nil {
			t.Fatalf("empire %d scout visual genome invalid: %v", empire.ID, err)
		}
		visualShips := 0
		for j := range state.Ships {
			ship := &state.Ships[j]
			if ship.EmpireID != empire.ID || ship.SourceDesignID != design.ID {
				continue
			}
			visualShips++
			if ship.SourceVisualRevision != design.VisualRevision || ship.VisualGenome == nil || !reflect.DeepEqual(*ship.VisualGenome, *design.VisualGenome) {
				t.Fatalf("empire %d starting scout %d visual=%+v design=%+v", empire.ID, ship.ID, ship.VisualGenome, design.VisualGenome)
			}
			if ship.VisualGenome == design.VisualGenome {
				t.Fatalf("empire %d starting scout %d aliases design visual genome", empire.ID, ship.ID)
			}
		}
		if visualShips != 2 {
			t.Fatalf("empire %d visual starting scouts=%d want 2", empire.ID, visualShips)
		}
		combat := 0
		colonyShips := 0
		for _, fleet := range state.StrategicFleets {
			if fleet.EmpireID != empire.ID {
				continue
			}
			if fleet.Role == core.StrategicFleetRoleCombat {
				combat++
				if len(fleet.ShipIDs) != 2 {
					t.Fatalf("combat fleet ships=%v", fleet.ShipIDs)
				}
			}
			if fleet.SpecialKind == core.StrategicFleetSpecialColonyShip {
				colonyShips++
			}
		}
		if combat != 1 || colonyShips != 1 {
			t.Fatalf("empire %d combat/colony fleets=%d/%d", empire.ID, combat, colonyShips)
		}
	}
	if state.ShipDesigns[0].VisualGenome.Seed == state.ShipDesigns[1].VisualGenome.Seed {
		t.Fatal("starting scout visual seeds should differ by empire/design")
	}
	ids := append([]core.ID(nil), state.ShipDesigns[0].ID, state.ShipDesigns[1].ID)
	if !sort.SliceIsSorted(ids, func(i, j int) bool { return ids[i] < ids[j] }) {
		t.Fatal("design IDs not ordered")
	}
}

func TestNewGameRejectsUnsupportedSettings(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	cases := []NewGameSettings{
		{},
		func() NewGameSettings { s := canonicalNewGameSettings(); s.GalaxySize = "tiny"; return s }(),
		func() NewGameSettings { s := canonicalNewGameSettings(); s.GalaxyAge = "young"; return s }(),
		func() NewGameSettings {
			s := canonicalNewGameSettings()
			s.TechnologyLevel = NewGameTechnologyPreWarp
			return s
		}(),
		func() NewGameSettings { s := canonicalNewGameSettings(); s.StrategicCombat = true; return s }(),
		func() NewGameSettings { s := canonicalNewGameSettings(); s.Players[1].RaceID = "human"; return s }(),
		func() NewGameSettings { s := canonicalNewGameSettings(); s.Players[1].SeatID = 1; return s }(),
		func() NewGameSettings { s := canonicalNewGameSettings(); s.Players[1].EmpireName = "human"; return s }(),
	}
	for i, settings := range cases {
		if _, err := rules.NewGame(7, settings); err == nil {
			t.Fatalf("case %d unexpectedly accepted", i)
		}
	}
}

func TestNewGameAllGalaxySizeAgeTuplesAreDeterministicAndBounded(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	sizes := []struct {
		id          GalaxySize
		stars       int
		gridColumns int
		gridRows    int
	}{
		{GalaxySizeSmall, 20, 5, 4},
		{GalaxySizeMedium, 36, 6, 6},
		{GalaxySizeLarge, 54, 9, 6},
		{GalaxySizeHuge, 71, 9, 8},
	}
	ages := []GalaxyAge{GalaxyAgeMineralRich, GalaxyAgeNormal, GalaxyAgeOrganicRich}
	for _, size := range sizes {
		for _, age := range ages {
			t.Run(string(size.id)+"/"+string(age), func(t *testing.T) {
				settings := canonicalNewGameSettings()
				settings.GalaxySize = size.id
				settings.GalaxyAge = age
				first, err := rules.NewGame(0x8009, settings)
				if err != nil {
					t.Fatal(err)
				}
				second, err := rules.NewGame(0x8009, settings)
				if err != nil {
					t.Fatal(err)
				}
				firstJSON, err := json.Marshal(first.State)
				if err != nil {
					t.Fatal(err)
				}
				secondJSON, err := json.Marshal(second.State)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(firstJSON, secondJSON) {
					t.Fatal("same seed/settings did not produce byte-identical state")
				}
				if got := len(first.State.Galaxy.Systems); got != size.stars {
					t.Fatalf("systems=%d want=%d", got, size.stars)
				}
				for i, system := range first.State.Galaxy.Systems {
					column, row := i%size.gridColumns, i/size.gridColumns
					centerX, centerY := 100+200*column, 100+200*row
					if system.X < centerX-40 || system.X > centerX+40 || system.Y < centerY-40 || system.Y > centerY+40 {
						t.Fatalf("system[%d] coordinate=(%d,%d) outside frozen cell jitter around (%d,%d)", i, system.X, system.Y, centerX, centerY)
					}
				}
			})
		}
	}
}

func TestNewGameGalaxyAgeProfilesMatchFrozenGate2Data(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	if got := len(rules.NewGameGalaxy.GalaxyAges); got != 3 {
		t.Fatalf("galaxy ages=%d want=3", got)
	}
	wantSpectral := [][]int{
		{20, 25, 10, 10, 32, 1, 2},
		{10, 15, 16, 16, 37, 2, 4},
		{5, 5, 30, 21, 30, 3, 6},
	}
	wantNormalClimate := [][]int{
		{15, 55, 25, 5, 0, 0, 0, 0, 0, 0},
		{15, 50, 25, 10, 5, 0, 0, 0, 0, 0},
		{10, 15, 10, 10, 10, 10, 11, 11, 11, 2},
		{20, 0, 70, 0, 8, 2, 0, 0, 0, 0},
	}
	wantOrganicClimate := [][]int{
		{15, 40, 20, 25, 0, 0, 0, 0, 0, 0},
		{5, 30, 20, 25, 20, 0, 0, 0, 0, 0},
		{5, 8, 8, 13, 13, 13, 13, 13, 10, 4},
		{20, 0, 50, 0, 30, 0, 0, 0, 0, 0},
	}
	for i, age := range rules.NewGameGalaxy.GalaxyAges {
		if !reflect.DeepEqual(age.SpectralWeights, wantSpectral[i]) {
			t.Fatalf("age[%d] spectral=%v want=%v", i, age.SpectralWeights, wantSpectral[i])
		}
		wantClimate := wantNormalClimate
		if i == 2 {
			wantClimate = wantOrganicClimate
		}
		if !reflect.DeepEqual(age.ClimateWeights, wantClimate) {
			t.Fatalf("age[%d] climate=%v want=%v", i, age.ClimateWeights, wantClimate)
		}
	}
}

func TestNewGamePersistsOrbitalBodiesAndSpectralClass(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	result, err := rules.NewGame(0x8009, canonicalNewGameSettings())
	if err != nil {
		t.Fatal(err)
	}
	state := result.State
	counts := map[core.OrbitalBodyKind]int{}
	for si := range state.Galaxy.Systems {
		system := &state.Galaxy.Systems[si]
		if system.SpectralClass < 0 || system.SpectralClass > 6 {
			t.Fatalf("system %d spectral class=%d", system.ID, system.SpectralClass)
		}
		orbits := map[int]bool{}
		planetBodies := map[core.ID]bool{}
		for _, body := range system.Bodies {
			if body.ID == 0 || body.Name == "" {
				t.Fatalf("system %d has incomplete body %+v", system.ID, body)
			}
			if orbits[body.Orbit] {
				t.Fatalf("system %d has duplicate body orbit %d", system.ID, body.Orbit)
			}
			orbits[body.Orbit] = true
			counts[body.Kind]++
			if body.Kind == core.OrbitalBodyPlanet {
				if body.PlanetID == 0 || body.ID != body.PlanetID {
					t.Fatalf("system %d planet body=%+v", system.ID, body)
				}
				planetBodies[body.PlanetID] = true
			} else if body.PlanetID != 0 {
				t.Fatalf("system %d non-planet body references planet: %+v", system.ID, body)
			}
		}
		for _, planet := range system.Planets {
			if !planetBodies[planet.ID] {
				t.Fatalf("system %d planet %d has no persistent orbital body", system.ID, planet.ID)
			}
		}
	}
	if counts[core.OrbitalBodyPlanet] == 0 || counts[core.OrbitalBodyGasGiant] == 0 || counts[core.OrbitalBodyAsteroidBelt] == 0 {
		t.Fatalf("body counts=%v want planets, gas giants and asteroid belts", counts)
	}
	if err := state.Validate(); err != nil {
		t.Fatalf("new-game body state invalid: %v", err)
	}
}
