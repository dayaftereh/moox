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
	host, err := newServerHost(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"), false)
	if err != nil {
		t.Fatal(err)
	}
	if games := host.ListGames(); len(games) != 0 {
		t.Fatalf("production host started with games: %+v", games)
	}
}

func TestDemoFixtureRequiresExplicitFlag(t *testing.T) {
	host, err := newServerHost(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"), true)
	if err != nil {
		t.Fatal(err)
	}
	games := host.ListGames()
	if len(games) != 1 || games[0].GameID != demoGameID {
		t.Fatalf("explicit demo fixture games=%+v", games)
	}
}
