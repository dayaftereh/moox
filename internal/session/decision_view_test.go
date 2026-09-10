package session

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"moox/internal/battle"
	"moox/internal/game"
	"moox/internal/protocol"
)

func TestPlayerDecisionViewIsPlayerSafeDeepCopyAndDeterministic(t *testing.T) {
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	generated, err := rules.NewGame(0x8009, game.NewGameSettings{
		GalaxySize:      game.GalaxySizeSmall,
		GalaxyAge:       game.GalaxyAgeNormal,
		TechnologyLevel: game.NewGameTechnologyAverage,
		Players: []game.NewGamePlayerSpec{
			{SeatID: 1, EmpireName: "Human", RaceID: "human"},
			{SeatID: 2, EmpireName: "Darlok", RaceID: "darlok"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	seats := []Seat{
		{ID: 1, EmpireID: generated.Players[0].EmpireID, Name: "Human", Controller: ControllerBuiltinAI},
		{ID: 2, EmpireID: generated.Players[1].EmpireID, Name: "Darlok", Controller: ControllerBuiltinAI},
	}
	s, err := NewGameSession("decision-view", generated.State, seats)
	if err != nil {
		t.Fatal(err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		t.Fatal(err)
	}
	observerBefore, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	observerBytesBefore, err := json.Marshal(observerBefore)
	if err != nil {
		t.Fatal(err)
	}

	view, err := s.DecisionView(protocol.SeatID(1), resolver)
	if err != nil {
		t.Fatal(err)
	}
	if view.Empire.ID != generated.Players[0].EmpireID {
		t.Fatalf("decision empire=%d want=%d", view.Empire.ID, generated.Players[0].EmpireID)
	}
	if len(view.PublicEmpires) != len(generated.State.Empires) {
		t.Fatalf("public empire identities=%d want=%d", len(view.PublicEmpires), len(generated.State.Empires))
	}
	foundPublicDarlok := false
	for _, identity := range view.PublicEmpires {
		if identity.ID == generated.Players[1].EmpireID {
			foundPublicDarlok = true
			if identity.Name != "Darlok" || identity.RaceID != "darlok" {
				t.Fatalf("public Darlok identity=%+v", identity)
			}
		}
	}
	if !foundPublicDarlok {
		t.Fatal("decision view missing public Darlok identity")
	}
	for _, colony := range view.Colonies {
		if colony.EmpireID != view.Empire.ID {
			t.Fatalf("decision view leaked foreign colony %+v", colony)
		}
	}
	for _, ship := range view.Strategic.Ships {
		if ship.EmpireID != view.Empire.ID {
			t.Fatalf("decision view leaked foreign ship %+v", ship)
		}
	}
	for _, fleet := range view.Strategic.Fleets {
		if fleet.EmpireID != view.Empire.ID {
			t.Fatalf("decision view leaked foreign fleet composition %+v", fleet)
		}
	}
	foreignColonyContact, foreignFleetContact := false, false
	for _, contact := range view.Strategic.Contacts {
		if contact.EmpireID == view.Empire.ID {
			t.Fatalf("own asset projected redundantly as foreign contact %+v", contact)
		}
		switch contact.Kind {
		case StrategicContactColony:
			foreignColonyContact = true
		case StrategicContactFleet:
			foreignFleetContact = true
		}
	}
	if !foreignColonyContact || !foreignFleetContact {
		t.Fatalf("symmetric strategic contacts missing colony=%v fleet=%v contacts=%+v", foreignColonyContact, foreignFleetContact, view.Strategic.Contacts)
	}
	planetCount := 0
	for _, system := range view.Strategic.Galaxy.Systems {
		if len(system.BlockadedEmpireIDs) != 0 {
			t.Fatalf("decision galaxy leaked blockade internals for system %d: %v", system.ID, system.BlockadedEmpireIDs)
		}
		for _, planet := range system.Planets {
			planetCount++
			if planet.ColonyID != 0 || planet.OutpostID != 0 {
				t.Fatalf("decision galaxy leaked occupancy internals planet=%d colony=%d outpost=%d", planet.ID, planet.ColonyID, planet.OutpostID)
			}
		}
	}
	if len(view.Strategic.PlanetPotentials) != planetCount {
		t.Fatalf("planet potential count=%d want=%d", len(view.Strategic.PlanetPotentials), planetCount)
	}
	for _, potential := range view.Strategic.PlanetPotentials {
		if potential.PlanetID == 0 || potential.PopulationCapacity <= 0 {
			t.Fatalf("invalid player-relative planet potential %+v", potential)
		}
	}
	if len(view.Decisions.Research) == 0 || len(view.Decisions.Construction) == 0 || len(view.Decisions.Population) == 0 {
		t.Fatalf("decision catalog incomplete: %+v", view.Decisions)
	}
	if len(view.Decisions.ShipDesigner.Hulls) != 6 || view.Decisions.ShipDesigner.Hulls[0].ID != "frigate" || !view.Decisions.ShipDesigner.Hulls[0].SaveAvailable {
		t.Fatalf("decision Ship Designer hulls=%+v", view.Decisions.ShipDesigner.Hulls)
	}
	if len(view.Decisions.ShipDesigner.Weapons) != 1 || view.Decisions.ShipDesigner.Weapons[0].ID != "laser_cannon" || !view.Decisions.ShipDesigner.Weapons[0].Available {
		t.Fatalf("decision Ship Designer weapons=%+v", view.Decisions.ShipDesigner.Weapons)
	}
	if len(view.Decisions.ShipDesigner.Variants) != 2 || view.Decisions.ShipDesigner.Variants[1].Spec.ProductionCostPP <= view.Decisions.ShipDesigner.Variants[0].Spec.ProductionCostPP {
		t.Fatalf("decision Ship Designer variants=%+v", view.Decisions.ShipDesigner.Variants)
	}

	firstBytes, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	view2, err := s.DecisionView(1, resolver)
	if err != nil {
		t.Fatal(err)
	}
	secondBytes, err := json.Marshal(view2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatal("same state/seat produced different DecisionView bytes")
	}

	view.Empire.Treasury.BalanceBC = -999999
	if len(view.Colonies) != 0 {
		view.Colonies[0].PlanetID = 999999
	}
	if len(view.Strategic.Galaxy.Systems) != 0 && len(view.Strategic.Galaxy.Systems[0].Planets) != 0 {
		view.Strategic.Galaxy.Systems[0].Planets[0].ClimateID = "mutated"
	}
	if len(view.Strategic.Fleets) != 0 && len(view.Strategic.Fleets[0].ShipIDs) != 0 {
		view.Strategic.Fleets[0].ShipIDs[0] = 999999
	}
	observerAfter, err := s.ObserverView()
	if err != nil {
		t.Fatal(err)
	}
	observerBytesAfter, err := json.Marshal(observerAfter)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(observerBytesBefore, observerBytesAfter) {
		t.Fatal("mutating returned DecisionView changed authoritative session state")
	}
}

func TestDecisionViewCatalogsSupportedTacticalActions(t *testing.T) {
	state, seats, _ := makeTacticalStrategicFixture(t, 1)
	resolver := loadTacticalEconomyResolver(t)
	s, err := NewGameSession("decision-battle", state, seats)
	if err != nil {
		t.Fatal(err)
	}
	submitStubTurn(t, s, "decision-battle")
	if err := s.ResolveStrategic(resolver); err != nil {
		t.Fatal(err)
	}
	status := s.Status()
	if status.Phase != PhaseEncounters {
		t.Fatalf("phase=%q", status.Phase)
	}
	view, err := s.DecisionView(1, resolver)
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Decisions.Battles) != 1 || len(view.Decisions.Battles[0].Actions) == 0 {
		t.Fatalf("battle decision catalog=%+v", view.Decisions.Battles)
	}
	if view.Decisions.Battles[0].Actions[0].Kind != battle.CommandFireBeam {
		t.Fatalf("first tactical AI action kind=%q", view.Decisions.Battles[0].Actions[0].Kind)
	}
	before := s.Status()
	if before.Revision != status.Revision {
		t.Fatalf("building battle catalog mutated session revision: before=%d after=%d", status.Revision, before.Revision)
	}
}
