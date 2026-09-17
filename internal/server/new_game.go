package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"moox/internal/app"
	"moox/internal/core"
	"moox/internal/game"
)

type difficultyCatalogResponse struct {
	SchemaVersion int                      `json:"schema_version"`
	DefaultID     core.DifficultyID        `json:"default_id"`
	Profiles      []game.DifficultyProfile `json:"profiles"`
}

func (s *apiServer) handleDifficultyCatalog(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, difficultyCatalogResponse{
		SchemaVersion: app.SchemaVersion,
		DefaultID:     game.DefaultDifficultyID,
		Profiles:      game.DifficultyProfiles(),
	})
}

type galaxyCatalogResponse struct {
	SchemaVersion int                      `json:"schema_version"`
	DefaultSizeID game.GalaxySize          `json:"default_size_id"`
	DefaultAgeID  game.GalaxyAge           `json:"default_age_id"`
	Sizes         []game.GalaxySizeProfile `json:"sizes"`
	Ages          []game.GalaxyAgeProfile  `json:"ages"`
}

func (s *apiServer) handleGalaxyCatalog(w http.ResponseWriter, _ *http.Request) {
	catalog, err := s.host.NewGameGalaxyCatalog()
	if err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "service_unavailable", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, galaxyCatalogResponse{
		SchemaVersion: app.SchemaVersion,
		DefaultSizeID: catalog.DefaultSizeID,
		DefaultAgeID:  catalog.DefaultAgeID,
		Sizes:         catalog.Sizes,
		Ages:          catalog.Ages,
	})
}

type raceCatalogResponse struct {
	SchemaVersion       int                      `json:"schema_version"`
	DefaultPlayerRaceID string                   `json:"default_player_race_id"`
	FixedOpponentRaceID string                   `json:"fixed_opponent_race_id"`
	Profiles            []game.PresetRaceProfile `json:"profiles"`
}

func (s *apiServer) handleRaceCatalog(w http.ResponseWriter, _ *http.Request) {
	catalog, err := s.host.NewGameRaceCatalog()
	if err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "service_unavailable", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, raceCatalogResponse{
		SchemaVersion:       app.SchemaVersion,
		DefaultPlayerRaceID: catalog.DefaultPlayerRaceID,
		FixedOpponentRaceID: catalog.FixedOpponentRaceID,
		Profiles:            catalog.Profiles,
	})
}

type technologyCatalogResponse struct {
	SchemaVersion int                             `json:"schema_version"`
	DefaultID     game.NewGameTechnologyLevel     `json:"default_id"`
	Profiles      []game.NewGameTechnologyProfile `json:"profiles"`
}

func (s *apiServer) handleTechnologyCatalog(w http.ResponseWriter, _ *http.Request) {
	catalog, err := s.host.NewGameTechnologyCatalog()
	if err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "service_unavailable", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, technologyCatalogResponse{
		SchemaVersion: app.SchemaVersion,
		DefaultID:     catalog.DefaultID,
		Profiles:      catalog.Profiles,
	})
}

type compositionCatalogResponse struct {
	SchemaVersion        int                                `json:"schema_version"`
	DefaultOpponentCount int                                `json:"default_opponent_count"`
	OriginalOpponentMin  int                                `json:"original_opponent_min"`
	OriginalOpponentMax  int                                `json:"original_opponent_max"`
	Counts               []game.NewGameOpponentCountProfile `json:"counts"`
	GalaxyLimits         []game.NewGameOpponentGalaxyLimit  `json:"galaxy_limits"`
	Assignments          []game.NewGameOpponentAssignment   `json:"assignments"`
}

func (s *apiServer) handleCompositionCatalog(w http.ResponseWriter, _ *http.Request) {
	catalog, err := s.host.NewGameCompositionCatalog()
	if err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "service_unavailable", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, compositionCatalogResponse{
		SchemaVersion:        app.SchemaVersion,
		DefaultOpponentCount: catalog.DefaultOpponentCount,
		OriginalOpponentMin:  catalog.OriginalOpponentMin,
		OriginalOpponentMax:  catalog.OriginalOpponentMax,
		Counts:               catalog.Counts,
		GalaxyLimits:         catalog.GalaxyLimits,
		Assignments:          catalog.Assignments,
	})
}

type newGameRequest struct {
	SchemaVersion int                        `json:"schema_version"`
	GameID        string                     `json:"game_id"`
	Seed          string                     `json:"seed"`
	Settings      game.NewGameSettings       `json:"settings"`
	Controllers   []app.PlayerControllerSpec `json:"controllers,omitempty"`
}

type newGameResponse struct {
	SchemaVersion int                        `json:"schema_version"`
	Game          app.GameSummary            `json:"game"`
	Players       []game.NewGamePlayerResult `json:"players"`
}

func (s *apiServer) handleCreateGame(w http.ResponseWriter, r *http.Request) {
	if !validateMutationRequest(w, r) {
		return
	}
	var request newGameRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if request.SchemaVersion != app.SchemaVersion {
		writeAPIError(w, http.StatusBadRequest, "bad_request", fmt.Sprintf("unsupported schema_version %d", request.SchemaVersion))
		return
	}
	if request.GameID != "" && strings.TrimSpace(request.GameID) != request.GameID {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "game_id must not have leading or trailing whitespace")
		return
	}
	if len(request.Controllers) != len(request.Settings.Players) {
		writeAPIError(w, http.StatusBadRequest, "bad_request", fmt.Sprintf("Slice 16.6 requires explicit controller assignments for every player; got %d controllers for %d players", len(request.Controllers), len(request.Settings.Players)))
		return
	}
	seed, err := parseNewGameSeed(request.Seed)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	gameID := request.GameID
	if gameID == "" {
		gameID = nextServerGameID(s.host.ListGames())
	}
	created, err := s.host.CreateGame(app.CreateGameRequest{GameID: gameID, Seed: seed, Settings: request.Settings, Controllers: request.Controllers})
	if request.GameID == "" && errors.Is(err, app.ErrGameExists) {
		for retries := 0; retries < 32 && errors.Is(err, app.ErrGameExists); retries++ {
			gameID = nextServerGameID(s.host.ListGames())
			created, err = s.host.CreateGame(app.CreateGameRequest{GameID: gameID, Seed: seed, Settings: request.Settings, Controllers: request.Controllers})
		}
	}
	if err != nil {
		switch {
		case errors.Is(err, game.ErrInvalidNewGameSettings):
			writeAPIError(w, http.StatusBadRequest, "bad_request", err.Error())
		case errors.Is(err, app.ErrGameExists):
			writeAPIError(w, http.StatusConflict, "game_exists", err.Error())
		default:
			writeAPIError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		}
		return
	}
	writeJSONStatus(w, http.StatusCreated, newGameResponse{SchemaVersion: app.SchemaVersion, Game: created.Game, Players: created.Players})
}

func nextServerGameID(games []app.GameSummary) string {
	used := make(map[int]struct{}, len(games))
	for _, summary := range games {
		if !strings.HasPrefix(summary.GameID, "game-") {
			continue
		}
		n, err := strconv.Atoi(strings.TrimPrefix(summary.GameID, "game-"))
		if err == nil && n > 0 {
			used[n] = struct{}{}
		}
	}
	for n := 1; ; n++ {
		if _, exists := used[n]; !exists {
			return fmt.Sprintf("game-%d", n)
		}
	}
}

func parseNewGameSeed(raw string) (uint64, error) {
	if raw == "" || strings.TrimSpace(raw) != raw {
		return 0, fmt.Errorf("seed must be a non-empty decimal or 0x hexadecimal string")
	}
	base := 10
	digits := raw
	if strings.HasPrefix(raw, "0x") || strings.HasPrefix(raw, "0X") {
		base = 16
		digits = raw[2:]
		if digits == "" {
			return 0, fmt.Errorf("seed hexadecimal value is empty")
		}
	}
	value, err := strconv.ParseUint(digits, base, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid seed %q: %w", raw, err)
	}
	return value, nil
}
