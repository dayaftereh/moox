package main

import (
	"path/filepath"
	"testing"
)

func TestDefaultHTTPAddress(t *testing.T) {
	if defaultHTTPAddress != "127.0.0.1:7171" {
		t.Fatalf("default HTTP address=%q, want 127.0.0.1:7171", defaultHTTPAddress)
	}
}

func TestLoopbackAddressPolicy(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:7171", "[::1]:7171", "localhost:7171"} {
		if !isLoopbackAddress(addr) {
			t.Fatalf("expected loopback address %q", addr)
		}
	}
	for _, addr := range []string{"0.0.0.0:7171", ":7171", "192.168.1.20:7171", "bad"} {
		if isLoopbackAddress(addr) {
			t.Fatalf("unexpected loopback approval for %q", addr)
		}
	}
}

func TestProductionHostStartsEmpty(t *testing.T) {
	host, err := newServerHost(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"), false, false)
	if err != nil {
		t.Fatal(err)
	}
	if games := host.ListGames(); len(games) != 0 {
		t.Fatalf("production host started with games: %+v", games)
	}
}

func TestDemoFixtureRequiresExplicitFlag(t *testing.T) {
	host, err := newServerHost(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"), true, false)
	if err != nil {
		t.Fatal(err)
	}
	games := host.ListGames()
	if len(games) != 1 || games[0].GameID != demoGameID {
		t.Fatalf("explicit demo fixture games=%+v", games)
	}
}
func TestReferenceGamesRequireExplicitFlag(t *testing.T) {
	host, err := newServerHost(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"), false, true)
	if err != nil {
		t.Fatal(err)
	}
	games := host.ListGames()
	if len(games) != 2 {
		t.Fatalf("reference games=%+v, want exactly two", games)
	}
	seen := map[string]bool{}
	for _, game := range games {
		seen[game.GameID] = true
	}
	if !seen[standardReferenceGameID] || !seen[triangleReferenceGameID] {
		t.Fatalf("reference game IDs=%v, want %q and %q", seen, standardReferenceGameID, triangleReferenceGameID)
	}

	standardHuman, err := host.PlayerSnapshot(standardReferenceGameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	standardAI, err := host.PlayerSnapshot(standardReferenceGameID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if standardHuman.View.Seat.Seat.Controller != "local_human" || standardAI.View.Seat.Seat.Controller != "builtin_ai" {
		t.Fatalf("standard controllers human=%q ai=%q", standardHuman.View.Seat.Seat.Controller, standardAI.View.Seat.Seat.Controller)
	}

	triangleHuman, err := host.PlayerSnapshot(triangleReferenceGameID, 1)
	if err != nil {
		t.Fatal(err)
	}
	triangleAI1, err := host.PlayerSnapshot(triangleReferenceGameID, 2)
	if err != nil {
		t.Fatal(err)
	}
	triangleAI2, err := host.PlayerSnapshot(triangleReferenceGameID, 3)
	if err != nil {
		t.Fatal(err)
	}
	if triangleHuman.View.Seat.Seat.Controller != "local_human" || triangleAI1.View.Seat.Seat.Controller != "builtin_ai" || triangleAI2.View.Seat.Seat.Controller != "builtin_ai" {
		t.Fatalf("triangle controllers human=%q ai1=%q ai2=%q", triangleHuman.View.Seat.Seat.Controller, triangleAI1.View.Seat.Seat.Controller, triangleAI2.View.Seat.Seat.Controller)
	}
}
