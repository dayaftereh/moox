package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"moox/internal/app"
	"moox/internal/game"
)

func newPersistenceServer(t *testing.T, enabled bool) (*httptest.Server, *app.Host) {
	t.Helper()
	rules, err := game.LoadEconomyRules(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31"))
	if err != nil {
		t.Fatal(err)
	}
	host, err := app.NewHostWithNewGame(rules)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(Config{Host: host, ObserverEnabled: true, PersistenceEnabled: enabled})
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(handler), host
}

func createHTTPPersistenceGame(t *testing.T, serverURL, gameID string) {
	t.Helper()
	var created newGameResponse
	postJSON(t, serverURL+"/api/v1/games", serverNewGameRequest(gameID, "0x8009"), "", http.StatusCreated, &created)
}

func rawSnapshotRequest(t *testing.T, method, url string, body []byte, wantStatus int) []byte {
	t.Helper()
	request, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != wantStatus {
		t.Fatalf("%s %s status=%d want=%d body=%s", method, url, response.StatusCode, wantStatus, data)
	}
	return data
}

func TestHTTPPersistenceDisabledByDefault(t *testing.T) {
	server, _ := newPersistenceServer(t, false)
	defer server.Close()
	createHTTPPersistenceGame(t, server.URL, "persist-disabled")
	rawSnapshotRequest(t, http.MethodGet, server.URL+"/api/v1/games/persist-disabled/live-snapshot", nil, http.StatusForbidden)
	rawSnapshotRequest(t, http.MethodPost, server.URL+"/api/v1/games/import", []byte("{}"), http.StatusForbidden)
	rawSnapshotRequest(t, http.MethodPut, server.URL+"/api/v1/games/persist-disabled/live-snapshot", []byte("{}"), http.StatusForbidden)
}

func TestHTTPPersistenceExportImportRestoreRoundTrip(t *testing.T) {
	sourceServer, _ := newPersistenceServer(t, true)
	defer sourceServer.Close()
	createHTTPPersistenceGame(t, sourceServer.URL, "persist-http")

	saved := rawSnapshotRequest(t, http.MethodGet, sourceServer.URL+"/api/v1/games/persist-http/live-snapshot", nil, http.StatusOK)
	if len(saved) == 0 || saved[0] != '{' {
		t.Fatalf("exported snapshot=%q", saved)
	}

	targetServer, targetHost := newPersistenceServer(t, true)
	defer targetServer.Close()
	imported := rawSnapshotRequest(t, http.MethodPost, targetServer.URL+"/api/v1/games/import", saved, http.StatusCreated)
	if len(imported) == 0 {
		t.Fatal("empty import response")
	}
	if games := targetHost.ListGames(); len(games) != 1 || games[0].GameID != "persist-http" || games[0].ChangeSequence != 1 {
		t.Fatalf("imported games=%+v", games)
	}
	reexported := rawSnapshotRequest(t, http.MethodGet, targetServer.URL+"/api/v1/games/persist-http/live-snapshot", nil, http.StatusOK)
	if !bytes.Equal(saved, reexported) {
		t.Fatalf("HTTP import/re-export changed live bytes: %d vs %d", len(saved), len(reexported))
	}

	restored := rawSnapshotRequest(t, http.MethodPut, targetServer.URL+"/api/v1/games/persist-http/live-snapshot", saved, http.StatusOK)
	if len(restored) == 0 {
		t.Fatal("empty restore receipt")
	}
	if games := targetHost.ListGames(); len(games) != 1 || games[0].ChangeSequence != 2 {
		t.Fatalf("restore change sequence=%+v", games)
	}
}

func TestHTTPPersistenceInvalidSaveMapsToBadRequest(t *testing.T) {
	server, _ := newPersistenceServer(t, true)
	defer server.Close()
	rawSnapshotRequest(t, http.MethodPost, server.URL+"/api/v1/games/import", []byte("{"), http.StatusBadRequest)
}
