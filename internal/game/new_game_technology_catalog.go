package game

type NewGameTechnologyAvailability string

const (
	NewGameTechnologySupported NewGameTechnologyAvailability = "supported"
	NewGameTechnologyPlanned   NewGameTechnologyAvailability = "planned"
)

type NewGameTechnologyProfile struct {
	ID           NewGameTechnologyLevel        `json:"id"`
	NameKey      string                        `json:"name_key"`
	Availability NewGameTechnologyAvailability `json:"availability"`
	Facts        []string                      `json:"facts"`
}

type NewGameTechnologyCatalog struct {
	DefaultID NewGameTechnologyLevel     `json:"default_id"`
	Profiles  []NewGameTechnologyProfile `json:"profiles"`
}

func TechnologyLevelCatalog() NewGameTechnologyCatalog {
	return NewGameTechnologyCatalog{
		DefaultID: NewGameTechnologyAverage,
		Profiles: []NewGameTechnologyProfile{
			{ID: NewGameTechnologyPreWarp, NameKey: "newGame.techPreWarp", Availability: NewGameTechnologySupported, Facts: []string{"Single star", "No starting ships", "Interstellar capability must be researched", "2 completed starting fields / 6 known applications"}},
			{ID: NewGameTechnologyAverage, NameKey: "newGame.techAverage", Availability: NewGameTechnologySupported, Facts: []string{"Single star", "2 Scouts + 1 Colony Ship", "7 completed starting fields / 20 Tactical applications"}},
			{ID: NewGameTechnologyAdvanced, NameKey: "newGame.techAdvanced", Availability: NewGameTechnologyPlanned, Facts: []string{"Larger starting empire", "Average baseline + 19 extra research fields", "Starting fleet fills Command Points", "Exact grants vary by seed, race and initialization context"}},
		},
	}
}
