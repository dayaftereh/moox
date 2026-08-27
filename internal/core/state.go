package core

import "fmt"

const StateSchemaVersion = 1

type ID uint64

type GameState struct {
	SchemaVersion int      `json:"schema_version"`
	Seed          uint64   `json:"seed"`
	RNGState      uint64   `json:"rng_state"`
	Turn          uint64   `json:"turn"`
	NextID        ID       `json:"next_id"`
	Galaxy        Galaxy   `json:"galaxy"`
	Empires       []Empire `json:"empires"`
	Colonies      []Colony `json:"colonies"`
	Events        []Event  `json:"events"`
}

type Galaxy struct {
	ID      ID           `json:"id"`
	Systems []StarSystem `json:"systems"`
}

type StarSystem struct {
	ID      ID       `json:"id"`
	Name    string   `json:"name"`
	X       int      `json:"x"`
	Y       int      `json:"y"`
	Planets []Planet `json:"planets"`
}

type Planet struct {
	ID        ID     `json:"id"`
	Name      string `json:"name"`
	Orbit     int    `json:"orbit"`
	SizeID    string `json:"size_id"`
	MineralID string `json:"mineral_id"`
	GravityID string `json:"gravity_id"`
	ClimateID string `json:"climate_id"`
	ColonyID  ID     `json:"colony_id,omitempty"`
}

type Empire struct {
	ID      ID     `json:"id"`
	Name    string `json:"name"`
	RaceID  string `json:"race_id"`
	Capital ID     `json:"capital_colony_id,omitempty"`
}

type Colony struct {
	ID       ID `json:"id"`
	EmpireID ID `json:"empire_id"`
	PlanetID ID `json:"planet_id"`
}

type Event struct {
	Turn    uint64 `json:"turn"`
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

func NewGameState(seed uint64) *GameState {
	return &GameState{SchemaVersion: StateSchemaVersion, Seed: seed, RNGState: seed, Turn: 1, NextID: 1}
}

func (s *GameState) NewID() ID {
	id := s.NextID
	s.NextID++
	return id
}

func (s *GameState) RNG() *RNG {
	return &RNG{state: s.RNGState}
}

func (s *GameState) CommitRNG(rng *RNG) {
	s.RNGState = rng.State()
}

func (s *GameState) AdvanceTurn() {
	s.Turn++
}

func (s *GameState) AddEvent(kind, message string) {
	s.Events = append(s.Events, Event{Turn: s.Turn, Kind: kind, Message: message})
}

func (s *GameState) Validate() error {
	if s.SchemaVersion != StateSchemaVersion {
		return fmt.Errorf("unsupported state schema version %d", s.SchemaVersion)
	}
	if s.NextID == 0 {
		return fmt.Errorf("next_id must be non-zero")
	}
	if s.Turn == 0 {
		return fmt.Errorf("turn must start at 1 or later")
	}

	seen := make(map[ID]string)
	planetIDs := make(map[ID]struct{})
	empireIDs := make(map[ID]struct{})
	colonyIDs := make(map[ID]struct{})
	checkID := func(id ID, label string) error {
		if id == 0 {
			return fmt.Errorf("%s has zero id", label)
		}
		if previous, ok := seen[id]; ok {
			return fmt.Errorf("duplicate id %d used by %s and %s", id, previous, label)
		}
		seen[id] = label
		return nil
	}

	if err := checkID(s.Galaxy.ID, "galaxy"); err != nil {
		return err
	}
	for si := range s.Galaxy.Systems {
		system := &s.Galaxy.Systems[si]
		if system.Name == "" {
			return fmt.Errorf("system %d has no name", si)
		}
		if err := checkID(system.ID, fmt.Sprintf("system[%d]", si)); err != nil {
			return err
		}
		for pi := range system.Planets {
			planet := &system.Planets[pi]
			if planet.Name == "" || planet.SizeID == "" || planet.MineralID == "" || planet.GravityID == "" || planet.ClimateID == "" {
				return fmt.Errorf("system[%d] planet[%d] has incomplete identity", si, pi)
			}
			if err := checkID(planet.ID, fmt.Sprintf("system[%d].planet[%d]", si, pi)); err != nil {
				return err
			}
			planetIDs[planet.ID] = struct{}{}
		}
	}
	for i := range s.Empires {
		empire := &s.Empires[i]
		if empire.Name == "" || empire.RaceID == "" {
			return fmt.Errorf("empire[%d] has incomplete identity", i)
		}
		if err := checkID(empire.ID, fmt.Sprintf("empire[%d]", i)); err != nil {
			return err
		}
		empireIDs[empire.ID] = struct{}{}
	}
	for i := range s.Colonies {
		colony := &s.Colonies[i]
		if colony.EmpireID == 0 || colony.PlanetID == 0 {
			return fmt.Errorf("colony[%d] has incomplete references", i)
		}
		if _, ok := empireIDs[colony.EmpireID]; !ok {
			return fmt.Errorf("colony[%d] references unknown empire %d", i, colony.EmpireID)
		}
		if _, ok := planetIDs[colony.PlanetID]; !ok {
			return fmt.Errorf("colony[%d] references unknown planet %d", i, colony.PlanetID)
		}
		if err := checkID(colony.ID, fmt.Sprintf("colony[%d]", i)); err != nil {
			return err
		}
		colonyIDs[colony.ID] = struct{}{}
	}
	for i := range s.Empires {
		if capital := s.Empires[i].Capital; capital != 0 {
			if _, ok := colonyIDs[capital]; !ok {
				return fmt.Errorf("empire[%d] references unknown capital colony %d", i, capital)
			}
		}
	}
	for si := range s.Galaxy.Systems {
		for pi := range s.Galaxy.Systems[si].Planets {
			if colonyID := s.Galaxy.Systems[si].Planets[pi].ColonyID; colonyID != 0 {
				if _, ok := colonyIDs[colonyID]; !ok {
					return fmt.Errorf("system[%d] planet[%d] references unknown colony %d", si, pi, colonyID)
				}
			}
		}
	}
	for id := range seen {
		if id >= s.NextID {
			return fmt.Errorf("id %d is not below next_id %d", id, s.NextID)
		}
	}
	return nil
}
