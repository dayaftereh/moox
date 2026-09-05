package session

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"moox/internal/game"
	"moox/internal/protocol"
)

func TestPlanningPreviewRecalculatesDraftWithoutMutatingSession(t *testing.T) {
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := rules.NewGame(0x8009, game.NewGameSettings{
		GalaxySize: game.GalaxySizeSmall, GalaxyAge: game.GalaxyAgeNormal, TechnologyLevel: game.NewGameTechnologyAverage,
		Players: []game.NewGamePlayerSpec{{SeatID: 1, EmpireName: "Human", RaceID: "human"}, {SeatID: 2, EmpireName: "Darlok", RaceID: "darlok"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewGameSession("planning-preview", generated.State, []Seat{
		{ID: 1, EmpireID: generated.Players[0].EmpireID, Name: "Human", Controller: ControllerLocalHuman},
		{ID: 2, EmpireID: generated.Players[1].EmpireID, Name: "Darlok", Controller: ControllerBuiltinAI},
	})
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	statusBefore := s.Status()
	observerBefore, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	observerBytesBefore, err := json.Marshal(observerBefore)
	if err != nil {
		t.Fatal(err)
	}
	colony := observerBefore.State.Colonies[0]
	if colony.EmpireID != generated.Players[0].EmpireID {
		for _, candidate := range observerBefore.State.Colonies {
			if candidate.EmpireID == generated.Players[0].EmpireID {
				colony = candidate
				break
			}
		}
	}
	farmers, workers, scientists := colony.Population.Farmers(), colony.Population.Workers(), colony.Population.Scientists()
	if workers < 1 {
		t.Fatalf("fixture workers=%v want >=1", workers)
	}
	baseBatch := protocol.CommandBatch{SchemaVersion: protocol.CommandSchemaVersion, GameID: "planning-preview", SeatID: 1, Turn: observerBefore.State.Turn, BaseRevision: statusBefore.Revision}
	base, err := s.PlanningPreview(baseBatch, resolver)
	if err != nil {
		t.Fatal(err)
	}
	command, err := game.NewAssignPopulationCommand(1, game.AssignPopulationPayload{ColonyID: colony.ID, Farmers: farmers, Workers: workers - 1, Scientists: scientists + 1})
	if err != nil {
		t.Fatal(err)
	}
	batch := baseBatch
	batch.Commands = []protocol.Command{command}
	preview, err := s.PlanningPreview(batch, resolver)
	if err != nil {
		t.Fatal(err)
	}
	if preview.BaseRevision != statusBefore.Revision || preview.Turn != observerBefore.State.Turn {
		t.Fatalf("preview boundary=%+v status=%+v", preview, statusBefore)
	}
	if preview.Projection.Research.RPPerTurn <= base.Projection.Research.RPPerTurn {
		t.Fatalf("research RP/turn draft=%v base=%v want increase", preview.Projection.Research.RPPerTurn, base.Projection.Research.RPPerTurn)
	}
	var projected *game.PlanningColonyPreview
	for i := range preview.Projection.Colonies {
		if preview.Projection.Colonies[i].Colony.ID == colony.ID {
			projected = &preview.Projection.Colonies[i]
			break
		}
	}
	if projected == nil || projected.Colony.Population.Workers() != workers-1 || projected.Colony.Population.Scientists() != scientists+1 {
		t.Fatalf("projected colony=%+v", projected)
	}
	statusAfter := s.Status()
	observerAfter, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	observerBytesAfter, err := json.Marshal(observerAfter)
	if err != nil {
		t.Fatal(err)
	}
	if statusAfter.Revision != statusBefore.Revision || statusAfter.Phase != statusBefore.Phase || !bytes.Equal(observerBytesBefore, observerBytesAfter) {
		t.Fatalf("planning preview mutated session: before=%+v after=%+v", statusBefore, statusAfter)
	}
}
