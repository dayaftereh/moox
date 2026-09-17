package game

import (
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/protocol"
)

func compositionNewGameSettings(playerRace string, opponents int) NewGameSettings {
	settings := canonicalNewGameSettings()
	settings.Players[0].RaceID = playerRace
	settings.Players[0].EmpireName = map[string]string{"human": "Human", "klackon": "Klackon"}[playerRace]
	settings.Players[1].BuiltinAIControlled = true
	if opponents == 2 {
		other := "human"
		name := "Human"
		if playerRace == "human" {
			other = "klackon"
			name = "Klackon"
		}
		settings.Players = append(settings.Players, NewGamePlayerSpec{SeatID: protocol.SeatID(3), EmpireName: name, RaceID: other, BuiltinAIControlled: true})
	}
	return settings
}

func TestNewGameCompositionCatalogGate3Contract(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	catalog, err := NewGameCompositionCatalogForRules(rules)
	if err != nil {
		t.Fatal(err)
	}
	if catalog.DefaultOpponentCount != 1 || catalog.OriginalOpponentMin != 1 || catalog.OriginalOpponentMax != 7 {
		t.Fatalf("catalog range/default=%+v", catalog)
	}
	if len(catalog.Counts) != 7 || len(catalog.GalaxyLimits) != 4 || len(catalog.Assignments) != 4 {
		t.Fatalf("catalog breadth counts/limits/assignments=%d/%d/%d", len(catalog.Counts), len(catalog.GalaxyLimits), len(catalog.Assignments))
	}
	for i, profile := range catalog.Counts {
		wantCount := i + 1
		if profile.OpponentCount != wantCount {
			t.Fatalf("count[%d]=%d want=%d", i, profile.OpponentCount, wantCount)
		}
		if wantCount <= 2 {
			if profile.Availability != NewGameCompositionSupported || profile.ReasonID != "" {
				t.Fatalf("supported count %d profile=%+v", wantCount, profile)
			}
		} else if profile.Availability != NewGameCompositionPlanned || profile.ReasonID != NewGameCompositionReasonAdditionalAIRaceBreadth {
			t.Fatalf("planned count %d profile=%+v", wantCount, profile)
		}
	}
	for _, limit := range catalog.GalaxyLimits {
		if limit.MaxSupportedOpponents != 2 {
			t.Fatalf("galaxy %s max opponents=%d want=2", limit.GalaxySize, limit.MaxSupportedOpponents)
		}
	}
	for _, tc := range []struct {
		player string
		count  int
		want   []string
	}{
		{"human", 1, []string{"darlok"}},
		{"human", 2, []string{"darlok", "klackon"}},
		{"klackon", 1, []string{"darlok"}},
		{"klackon", 2, []string{"darlok", "human"}},
	} {
		got, ok := ExpectedNewGameOpponentRaceIDs(tc.player, tc.count)
		if !ok || !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("assignment %s+%d got=%v ok=%v want=%v", tc.player, tc.count, got, ok, tc.want)
		}
	}
}

func TestSelectNewGameHomeSystemsPreservesPairAndUsesMaximinTieBreak(t *testing.T) {
	systems := []newGameSystemPlan{
		{ID: 10, X: 0, Y: 0},
		{ID: 20, X: 10, Y: 0},
		{ID: 30, X: 5, Y: 4},
		{ID: 25, X: 5, Y: -4},
	}
	pair := farthestNewGameSystemPair(systems)
	two, err := selectNewGameHomeSystems(systems, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(two, []int{pair[0], pair[1]}) {
		t.Fatalf("two-player homes=%v pair=%v", two, pair)
	}
	three, err := selectNewGameHomeSystems(systems, 3)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(three, []int{0, 1, 3}) {
		t.Fatalf("three-player homes=%v want=[0 1 3] (lower System.ID wins maximin tie)", three)
	}
}

func TestNewGameThreePlayerCompositionCreatesFrozenStartingState(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	for _, playerRace := range []string{"human", "klackon"} {
		for _, technology := range []NewGameTechnologyLevel{NewGameTechnologyPreWarp, NewGameTechnologyAverage} {
			t.Run(playerRace+"/"+string(technology), func(t *testing.T) {
				settings := compositionNewGameSettings(playerRace, 2)
				settings.TechnologyLevel = technology
				result, err := rules.NewGame(0x1663, settings)
				if err != nil {
					t.Fatal(err)
				}
				if len(result.Players) != 3 || len(result.State.Empires) != 3 || len(result.State.Colonies) != 3 {
					t.Fatalf("players/empires/colonies=%d/%d/%d want=3/3/3", len(result.Players), len(result.State.Empires), len(result.State.Colonies))
				}
				wantRaces := []string{playerRace, "darlok", "human"}
				if playerRace == "human" {
					wantRaces[2] = "klackon"
				}
				for i, empire := range result.State.Empires {
					if empire.RaceID != wantRaces[i] {
						t.Fatalf("empire[%d] race=%q want=%q", i, empire.RaceID, wantRaces[i])
					}
					if (i > 0) != empire.BuiltinAIControlled {
						t.Fatalf("empire[%d] BuiltinAIControlled=%v", i, empire.BuiltinAIControlled)
					}
				}
				if technology == NewGameTechnologyPreWarp {
					if len(result.State.Ships) != 0 || len(result.State.ShipDesigns) != 0 || len(result.State.StrategicFleets) != 0 {
						t.Fatalf("Pre-Warp ships/designs/fleets=%d/%d/%d want=0/0/0", len(result.State.Ships), len(result.State.ShipDesigns), len(result.State.StrategicFleets))
					}
				} else {
					if len(result.State.Ships) != 6 || len(result.State.ShipDesigns) != 3 || len(result.State.StrategicFleets) != 6 {
						t.Fatalf("Average ships/designs/fleets=%d/%d/%d want=6/3/6", len(result.State.Ships), len(result.State.ShipDesigns), len(result.State.StrategicFleets))
					}
				}
			})
		}
	}
}

func TestNewGameGate3AcceptedCompositionMatrixDeterministic(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	technologies := []NewGameTechnologyLevel{NewGameTechnologyPreWarp, NewGameTechnologyAverage}
	playerRaces := []string{"human", "klackon"}
	for _, size := range SupportedGalaxySizeIDs {
		for _, age := range SupportedGalaxyAgeIDs {
			for _, difficulty := range core.SupportedDifficultyIDs {
				for _, playerRace := range playerRaces {
					for _, technology := range technologies {
						for _, opponents := range []int{1, 2} {
							settings := compositionNewGameSettings(playerRace, opponents)
							settings.GalaxySize = size
							settings.GalaxyAge = age
							settings.DifficultyID = difficulty
							settings.TechnologyLevel = technology
							first, err := rules.NewGame(0x1663, settings)
							if err != nil {
								t.Fatalf("size=%s age=%s difficulty=%s race=%s tech=%s opponents=%d: %v", size, age, difficulty, playerRace, technology, opponents, err)
							}
							second, err := rules.NewGame(0x1663, settings)
							if err != nil {
								t.Fatal(err)
							}
							if !reflect.DeepEqual(first, second) {
								t.Fatalf("nondeterministic tuple size=%s age=%s difficulty=%s race=%s tech=%s opponents=%d", size, age, difficulty, playerRace, technology, opponents)
							}
							if len(first.State.Empires) != opponents+1 || len(first.State.Colonies) != opponents+1 {
								t.Fatalf("tuple empire/colony count=%d/%d want=%d", len(first.State.Empires), len(first.State.Colonies), opponents+1)
							}
						}
					}
				}
			}
		}
	}
}
