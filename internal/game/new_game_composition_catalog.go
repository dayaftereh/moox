package game

import "fmt"

type NewGameCompositionAvailability string

const (
	NewGameCompositionSupported NewGameCompositionAvailability = "supported"
	NewGameCompositionPlanned   NewGameCompositionAvailability = "planned"

	NewGameCompositionReasonAdditionalAIRaceBreadth = "additional_ai_race_breadth_required"
)

type NewGameOpponentCountProfile struct {
	OpponentCount int                            `json:"opponent_count"`
	Availability  NewGameCompositionAvailability `json:"availability"`
	ReasonID      string                         `json:"reason_id,omitempty"`
	Facts         []string                       `json:"facts"`
}

type NewGameOpponentGalaxyLimit struct {
	GalaxySize            GalaxySize `json:"galaxy_size"`
	MaxSupportedOpponents int        `json:"max_supported_opponents"`
}

type NewGameOpponentSpec struct {
	RaceID            string `json:"race_id"`
	Controller        string `json:"controller"`
	DefaultEmpireName string `json:"default_empire_name"`
}

type NewGameOpponentAssignment struct {
	PlayerRaceID  string                `json:"player_race_id"`
	OpponentCount int                   `json:"opponent_count"`
	Opponents     []NewGameOpponentSpec `json:"opponents"`
}

type NewGameCompositionCatalog struct {
	DefaultOpponentCount int                           `json:"default_opponent_count"`
	OriginalOpponentMin  int                           `json:"original_opponent_min"`
	OriginalOpponentMax  int                           `json:"original_opponent_max"`
	Counts               []NewGameOpponentCountProfile `json:"counts"`
	GalaxyLimits         []NewGameOpponentGalaxyLimit  `json:"galaxy_limits"`
	Assignments          []NewGameOpponentAssignment   `json:"assignments"`
}

func NewGameCompositionCatalogForRules(rules *EconomyRules) (NewGameCompositionCatalog, error) {
	if rules == nil || len(rules.RaceModifiers) == 0 {
		return NewGameCompositionCatalog{}, fmt.Errorf("new game composition rules are unavailable")
	}
	for _, raceID := range []string{"human", "klackon", FixedOpponentRaceID} {
		if _, ok := rules.RaceModifiers[raceID]; !ok {
			return NewGameCompositionCatalog{}, fmt.Errorf("ruleset has no race %q required by composition catalog", raceID)
		}
	}
	counts := make([]NewGameOpponentCountProfile, 0, 7)
	for count := 1; count <= 7; count++ {
		profile := NewGameOpponentCountProfile{OpponentCount: count}
		if count <= 2 {
			profile.Availability = NewGameCompositionSupported
			if count == 1 {
				profile.Facts = []string{"1 built-in AI opponent", "Darlok baseline"}
			} else {
				profile.Facts = []string{"2 built-in AI opponents", "Darlok + supported alternate race"}
			}
		} else {
			profile.Availability = NewGameCompositionPlanned
			profile.ReasonID = NewGameCompositionReasonAdditionalAIRaceBreadth
			profile.Facts = []string{"Original-range count", "Additional AI race breadth required"}
		}
		counts = append(counts, profile)
	}
	limits := make([]NewGameOpponentGalaxyLimit, 0, len(SupportedGalaxySizeIDs))
	for _, size := range SupportedGalaxySizeIDs {
		limits = append(limits, NewGameOpponentGalaxyLimit{GalaxySize: size, MaxSupportedOpponents: 2})
	}
	return NewGameCompositionCatalog{
		DefaultOpponentCount: 1,
		OriginalOpponentMin:  1,
		OriginalOpponentMax:  7,
		Counts:               counts,
		GalaxyLimits:         limits,
		Assignments: []NewGameOpponentAssignment{
			{PlayerRaceID: "human", OpponentCount: 1, Opponents: []NewGameOpponentSpec{{RaceID: FixedOpponentRaceID, Controller: "builtin_ai", DefaultEmpireName: "Darlok"}}},
			{PlayerRaceID: "human", OpponentCount: 2, Opponents: []NewGameOpponentSpec{{RaceID: FixedOpponentRaceID, Controller: "builtin_ai", DefaultEmpireName: "Darlok"}, {RaceID: "klackon", Controller: "builtin_ai", DefaultEmpireName: "Klackon"}}},
			{PlayerRaceID: "klackon", OpponentCount: 1, Opponents: []NewGameOpponentSpec{{RaceID: FixedOpponentRaceID, Controller: "builtin_ai", DefaultEmpireName: "Darlok"}}},
			{PlayerRaceID: "klackon", OpponentCount: 2, Opponents: []NewGameOpponentSpec{{RaceID: FixedOpponentRaceID, Controller: "builtin_ai", DefaultEmpireName: "Darlok"}, {RaceID: "human", Controller: "builtin_ai", DefaultEmpireName: "Human"}}},
		},
	}, nil
}

func ExpectedNewGameOpponentRaceIDs(playerRaceID string, opponentCount int) ([]string, bool) {
	if !IsSupportedPlayerRaceID(playerRaceID) || (opponentCount != 1 && opponentCount != 2) {
		return nil, false
	}
	result := []string{FixedOpponentRaceID}
	if opponentCount == 2 {
		if playerRaceID == "human" {
			result = append(result, "klackon")
		} else {
			result = append(result, "human")
		}
	}
	return result, true
}
