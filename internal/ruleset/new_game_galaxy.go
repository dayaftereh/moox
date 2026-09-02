package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const NewGameGalaxySchemaVersion = 1

type NewGameGalaxyFile struct {
	SchemaVersion              int                   `json:"schema_version"`
	Ruleset                    string                `json:"ruleset"`
	Sources                    []Source              `json:"sources"`
	GalaxySizes                []NewGameGalaxySize   `json:"galaxy_sizes"`
	GalaxyAges                 []NewGameGalaxyAge    `json:"galaxy_ages"`
	SizeRollUpperThresholds    []int                 `json:"size_roll_upper_thresholds"`
	SpectralWeights            [][]int               `json:"spectral_weights"`
	SatelliteCounts            [][]int               `json:"satellite_counts"`
	BodyTypes                  [][]int               `json:"body_types"`
	MineralClasses             [][]int               `json:"mineral_classes"`
	GravityByMineralSize       [][]int               `json:"gravity_by_mineral_size"`
	PlanetGroupBySpectralOrbit [][]int               `json:"planet_group_by_spectral_orbit"`
	NormalClimateWeights       [][]int               `json:"normal_climate_weights"`
	Homeworld                  NewGameHomeworldRules `json:"homeworld"`
	Start                      NewGameStartRules     `json:"start"`
	Algorithms                 NewGameAlgorithms     `json:"algorithms"`
}

type NewGameGalaxySize struct {
	ID          string `json:"id"`
	Index       int    `json:"index"`
	Stars       int    `json:"stars"`
	GridColumns int    `json:"grid_columns"`
	GridRows    int    `json:"grid_rows"`
}

type NewGameGalaxyAge struct {
	ID    string `json:"id"`
	Index int    `json:"index"`
}

type NewGameHomeworldRules struct {
	MinimumPlanets int    `json:"minimum_planets"`
	SizeID         string `json:"size_id"`
	MineralID      string `json:"mineral_id"`
	GravityID      string `json:"gravity_id"`
	ClimateID      string `json:"climate_id"`
}

type NewGameStartRules struct {
	Population      float64 `json:"population"`
	Farmers         float64 `json:"farmers"`
	Workers         float64 `json:"workers"`
	Scientists      float64 `json:"scientists"`
	TreasuryBC      float64 `json:"treasury_bc"`
	Freighters      int     `json:"freighters"`
	ScoutCount      int     `json:"scout_count"`
	ColonyShipCount int     `json:"colony_ship_count"`
}

type NewGameAlgorithms struct {
	Coordinates        string `json:"coordinates"`
	HomeworldSelection string `json:"homeworld_selection"`
	OrbitSelection     string `json:"orbit_selection"`
	PopulationJobs     string `json:"population_jobs"`
	StartingBuildings  string `json:"starting_buildings"`
	ScoutSnapshot      string `json:"scout_snapshot"`
}

func LoadNewGameGalaxy(path string) (*NewGameGalaxyFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file NewGameGalaxyFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *NewGameGalaxyFile) Validate() error {
	if f == nil {
		return fmt.Errorf("new game galaxy rules are nil")
	}
	if f.SchemaVersion != NewGameGalaxySchemaVersion {
		return fmt.Errorf("unsupported new game galaxy schema version %d", f.SchemaVersion)
	}
	if f.Ruleset == "" || len(f.Sources) == 0 {
		return fmt.Errorf("new game galaxy ruleset and sources are required")
	}
	if len(f.GalaxySizes) != 4 {
		return fmt.Errorf("galaxy size count=%d, expected 4", len(f.GalaxySizes))
	}
	expectedStars := []int{20, 36, 54, 71}
	expectedCols := []int{5, 6, 9, 9}
	expectedRows := []int{4, 6, 6, 8}
	for i, size := range f.GalaxySizes {
		if size.Index != i || size.ID == "" || size.Stars != expectedStars[i] || size.GridColumns != expectedCols[i] || size.GridRows != expectedRows[i] {
			return fmt.Errorf("galaxy size[%d] is invalid", i)
		}
	}
	if !equalNewGameInts(f.SizeRollUpperThresholds, []int{1, 3, 7, 9, 10}) {
		return fmt.Errorf("size roll thresholds are invalid: %v", f.SizeRollUpperThresholds)
	}
	expectedAges := []string{"mineral_rich", "normal", "organic_rich"}
	if len(f.GalaxyAges) != len(expectedAges) {
		return fmt.Errorf("galaxy age count=%d, expected 3", len(f.GalaxyAges))
	}
	for i, age := range f.GalaxyAges {
		if age.Index != i || age.ID != expectedAges[i] {
			return fmt.Errorf("galaxy age[%d]=%q/%d is invalid", i, age.ID, age.Index)
		}
	}
	if err := validateMatrix("spectral_weights", f.SpectralWeights, 7, 3, 0, 100); err != nil {
		return err
	}
	for col := 0; col < 3; col++ {
		sum := 0
		for row := 0; row < 7; row++ {
			sum += f.SpectralWeights[row][col]
		}
		if sum != 100 {
			return fmt.Errorf("spectral_weights column %d sums to %d, expected 100", col, sum)
		}
	}
	if err := validateMatrix("satellite_counts", f.SatelliteCounts, 10, 6, 0, 5); err != nil {
		return err
	}
	if err := validateMatrix("body_types", f.BodyTypes, 10, 5, 1, 4); err != nil {
		return err
	}
	if err := validateMatrix("mineral_classes", f.MineralClasses, 10, 6, 0, 4); err != nil {
		return err
	}
	if err := validateMatrix("gravity_by_mineral_size", f.GravityByMineralSize, 5, 5, 0, 2); err != nil {
		return err
	}
	if err := validateMatrix("planet_group_by_spectral_orbit", f.PlanetGroupBySpectralOrbit, 6, 5, 0, 3); err != nil {
		return err
	}
	if err := validateMatrix("normal_climate_weights", f.NormalClimateWeights, 4, 10, 0, 100); err != nil {
		return err
	}
	for i, weights := range f.NormalClimateWeights {
		sum := 0
		for _, w := range weights {
			sum += w
		}
		if sum <= 0 {
			return fmt.Errorf("normal_climate_weights group %d has no weight", i)
		}
	}
	if f.Homeworld.MinimumPlanets != 3 || f.Homeworld.SizeID != "medium" || f.Homeworld.MineralID != "abundant" || f.Homeworld.GravityID != "normal_g" || f.Homeworld.ClimateID != "terran" {
		return fmt.Errorf("homeworld baseline is invalid")
	}
	if f.Start.Population != 8 || f.Start.Farmers != 4 || f.Start.Workers != 2 || f.Start.Scientists != 2 || f.Start.TreasuryBC != 50 || f.Start.Freighters != 0 || f.Start.ScoutCount != 2 || f.Start.ColonyShipCount != 1 {
		return fmt.Errorf("starting baseline is invalid")
	}
	if f.Algorithms.Coordinates != "moox_grid_jitter_v1" || f.Algorithms.HomeworldSelection != "farthest_pair_v1" || f.Algorithms.OrbitSelection != "uniform_unused_orbit_v1" || f.Algorithms.PopulationJobs != "continuous_4_2_2_v1" || f.Algorithms.StartingBuildings != "none_v1" || f.Algorithms.ScoutSnapshot != "current_compatible_scout_v1" {
		return fmt.Errorf("new game algorithm identifiers are invalid")
	}
	return nil
}

func validateMatrix(name string, matrix [][]int, rows, cols, min, max int) error {
	if len(matrix) != rows {
		return fmt.Errorf("%s rows=%d, expected %d", name, len(matrix), rows)
	}
	for r, row := range matrix {
		if len(row) != cols {
			return fmt.Errorf("%s row %d columns=%d, expected %d", name, r, len(row), cols)
		}
		for c, value := range row {
			if value < min || value > max {
				return fmt.Errorf("%s[%d][%d]=%d outside %d..%d", name, r, c, value, min, max)
			}
		}
	}
	return nil
}

func equalNewGameInts(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
