package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"moox/internal/app"
	"moox/internal/game"
)

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
	if request.GameID == "" || strings.TrimSpace(request.GameID) != request.GameID {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "game_id must be non-empty without leading or trailing whitespace")
		return
	}
	seed, err := parseNewGameSeed(request.Seed)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	created, err := s.host.CreateGame(app.CreateGameRequest{GameID: request.GameID, Seed: seed, Settings: request.Settings, Controllers: request.Controllers})
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
