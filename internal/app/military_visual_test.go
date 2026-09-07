package app

import (
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/game"
)

func appVisualGenome() core.ShipVisualGenome {
	return core.ShipVisualGenome{
		Version: core.ShipVisualGenomeVersion,
		HullID:  "battleship", StyleID: "sleek", MorphologyID: "hammer", Seed: "app:visual:v4",
		Length: 94, Beam: 31, StationCount: 5,
		StationWidths: []float64{7, 22, 29, 16, 0}, NotchDepths: []float64{0, 0, 7, 0, 0},
		EngineCount: 3, DetailCount: 4,
		Primitives: []core.ShipVisualPrimitive{{Kind: "wedge", T: .68, Length: 22, Width: 14, Sweep: .5}},
	}
}

func TestMilitaryDesignVisualImmediatePersistsAcrossHostExportImport(t *testing.T) {
	host := loadNewGameHost(t)
	const gameID = "visual-server-persistence"
	created, err := host.CreateGame(CreateGameRequest{GameID: gameID, Seed: 0x8011, Settings: appNewGameSettings()})
	if err != nil {
		t.Fatal(err)
	}
	seatID := created.Players[0].SeatID
	snapshot, err := host.PlayerSnapshot(gameID, seatID)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Decision == nil || len(snapshot.Decision.Strategic.ShipDesigns) == 0 {
		t.Fatalf("new game has no military design for persistence test: %+v", snapshot.Decision)
	}
	design := snapshot.Decision.Strategic.ShipDesigns[0]
	gameplayRevision := design.Revision
	genome := appVisualGenome()
	command, err := game.NewSetMilitaryDesignVisualCommand(1, game.SetMilitaryDesignVisualPayload{DesignID: design.ID, VisualGenome: genome})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := host.SubmitImmediateCommand(gameID, seatID, snapshot.View.Revision, command); err != nil {
		t.Fatal(err)
	}

	after, err := host.PlayerSnapshot(gameID, seatID)
	if err != nil {
		t.Fatal(err)
	}
	stored := after.Decision.Strategic.ShipDesigns[0]
	if stored.Revision != gameplayRevision || stored.VisualRevision != 1 || stored.VisualGenome == nil || !reflect.DeepEqual(*stored.VisualGenome, genome) {
		t.Fatalf("host visual design=%+v", stored)
	}

	saved, err := host.ExportLiveSnapshot(gameID)
	if err != nil {
		t.Fatal(err)
	}
	fresh := loadNewGameHost(t)
	if _, err := fresh.ImportLiveSnapshot(saved); err != nil {
		t.Fatal(err)
	}
	restored, err := fresh.PlayerSnapshot(gameID, seatID)
	if err != nil {
		t.Fatal(err)
	}
	restoredDesign := restored.Decision.Strategic.ShipDesigns[0]
	if restoredDesign.Revision != gameplayRevision || restoredDesign.VisualRevision != 1 || restoredDesign.VisualGenome == nil || !reflect.DeepEqual(*restoredDesign.VisualGenome, genome) {
		t.Fatalf("restored host visual design=%+v", restoredDesign)
	}
}
