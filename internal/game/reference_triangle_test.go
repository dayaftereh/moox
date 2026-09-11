package game

import (
	"bytes"
	"encoding/json"
	"testing"

	"moox/internal/core"
)

func TestReferenceTriangleGameContract(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	generated, err := rules.NewReferenceTriangleGame(ReferenceTriangleSeed)
	if err != nil {
		t.Fatal(err)
	}
	state := generated.State
	if len(state.Galaxy.Systems) != 3 {
		t.Fatalf("systems=%d want 3", len(state.Galaxy.Systems))
	}
	if len(state.Empires) != 3 || len(generated.Players) != 3 {
		t.Fatalf("empires=%d players=%d want 3/3", len(state.Empires), len(generated.Players))
	}
	wantRaces := []string{"human", "darlok", "psilon"}
	for i, empire := range state.Empires {
		if empire.RaceID != wantRaces[i] {
			t.Fatalf("empire[%d] race=%q want %q", i, empire.RaceID, wantRaces[i])
		}
		if empire.PlayerColorSlot != i+1 {
			t.Fatalf("empire[%d] color slot=%d want %d", i, empire.PlayerColorSlot, i+1)
		}
		if len(empire.VisitedSystemIDs) != 1 || empire.VisitedSystemIDs[0] != state.Galaxy.Systems[i].ID {
			t.Fatalf("empire[%d] visited=%v want only system %d", i, empire.VisitedSystemIDs, state.Galaxy.Systems[i].ID)
		}
		if len(empire.KnownEmpireIDs) != 0 {
			t.Fatalf("empire[%d] starts with known empires %v", i, empire.KnownEmpireIDs)
		}
	}

	for i := 0; i < len(state.Galaxy.Systems); i++ {
		for j := i + 1; j < len(state.Galaxy.Systems); j++ {
			if got := strategicDistanceParsecs(state.Galaxy.Systems[i], state.Galaxy.Systems[j]); got != 2 {
				t.Fatalf("distance %s -> %s = %d pc want 2", state.Galaxy.Systems[i].Name, state.Galaxy.Systems[j].Name, got)
			}
		}
	}

	for i, system := range state.Galaxy.Systems {
		if len(system.Planets) != 3 {
			t.Fatalf("system[%d] planets=%d want 3", i, len(system.Planets))
		}
		if len(system.Bodies) != 3 {
			t.Fatalf("system[%d] bodies=%d want 3", i, len(system.Bodies))
		}
		home := system.Planets[0]
		if home.ColonyID == 0 {
			t.Fatalf("system[%d] home planet has no colony", i)
		}
		colony := colonyByID(state, home.ColonyID)
		if colony == nil || colony.EmpireID != state.Empires[i].ID {
			t.Fatalf("system[%d] colony=%+v want empire %d", i, colony, state.Empires[i].ID)
		}
	}

	resolver, err := NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	for i, empire := range state.Empires {
		combatFleets := 0
		colonyFleets := 0
		for _, fleet := range state.StrategicFleets {
			if fleet.EmpireID != empire.ID || fleet.AtSystemID != state.Galaxy.Systems[i].ID {
				continue
			}
			switch {
			case fleet.Role == core.StrategicFleetRoleCombat:
				combatFleets++
				if len(fleet.ShipIDs) != rules.NewGameGalaxy.Start.ScoutCount {
					t.Fatalf("empire[%d] combat ship count=%d want %d", i, len(fleet.ShipIDs), rules.NewGameGalaxy.Start.ScoutCount)
				}
			case fleet.SpecialKind == core.StrategicFleetSpecialColonyShip:
				colonyFleets++
			}
		}
		if combatFleets != 1 || colonyFleets != 1 {
			t.Fatalf("empire[%d] combat fleets=%d colony fleets=%d want 1/1", i, combatFleets, colonyFleets)
		}

		targets, err := resolver.AvailableFleetMoveTargets(state, empire.ID)
		if err != nil {
			t.Fatal(err)
		}
		for destinationIndex, destination := range state.Galaxy.Systems {
			if destinationIndex == i {
				continue
			}
			legalProfiles := 0
			for _, target := range targets {
				if target.DestinationSystemID != destination.ID {
					continue
				}
				if !target.Legal || target.DistanceParsecs != 2 || target.ETA != 1 || target.FuelRangeParsecs < 2 {
					t.Fatalf("empire[%d] destination=%s target=%+v", i, destination.Name, target)
				}
				legalProfiles++
			}
			if legalProfiles < 2 {
				t.Fatalf("empire[%d] destination=%s legal profiles=%d want at least combat+colony", i, destination.Name, legalProfiles)
			}
		}
	}
}

func TestReferenceTriangleGameIsDeterministic(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	first, err := rules.NewReferenceTriangleGame(ReferenceTriangleSeed)
	if err != nil {
		t.Fatal(err)
	}
	second, err := rules.NewReferenceTriangleGame(ReferenceTriangleSeed)
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
		t.Fatalf("reference triangle state differs: first=%d bytes second=%d bytes", len(firstJSON), len(secondJSON))
	}
}
