package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"moox/internal/app"
	"moox/internal/protocol"
)

const maxJSONBodyBytes int64 = 1 << 20

type Config struct {
	Host               *app.Host
	StaticFS           fs.FS
	ObserverEnabled    bool
	PersistenceEnabled bool
}

type apiServer struct {
	host               *app.Host
	observerEnabled    bool
	persistenceEnabled bool
}

type commandRequest struct {
	SchemaVersion int              `json:"schema_version"`
	SeatID        protocol.SeatID  `json:"seat_id"`
	BaseRevision  uint64           `json:"base_revision"`
	Command       protocol.Command `json:"command"`
}

type apiErrorEnvelope struct {
	SchemaVersion int      `json:"schema_version"`
	Error         apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewHandler(cfg Config) (http.Handler, error) {
	if cfg.Host == nil {
		return nil, fmt.Errorf("application host must not be nil")
	}
	server := &apiServer{host: cfg.Host, observerEnabled: cfg.ObserverEnabled, persistenceEnabled: cfg.PersistenceEnabled}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", server.handleHealth)
	mux.HandleFunc("GET /api/v1/new-game/difficulties", server.handleDifficultyCatalog)
	mux.HandleFunc("GET /api/v1/new-game/galaxy", server.handleGalaxyCatalog)
	mux.HandleFunc("GET /api/v1/new-game/races", server.handleRaceCatalog)
	mux.HandleFunc("GET /api/v1/new-game/technologies", server.handleTechnologyCatalog)
	mux.HandleFunc("GET /api/v1/new-game/compositions", server.handleCompositionCatalog)
	mux.HandleFunc("GET /api/v1/games", server.handleGames)
	mux.HandleFunc("POST /api/v1/games", server.handleCreateGame)
	mux.HandleFunc("GET /api/v1/games/{gameID}/seats/{seatID}/snapshot", server.handlePlayerSnapshot)
	mux.HandleFunc("POST /api/v1/games/{gameID}/seats/{seatID}/planning-preview", server.handlePlanningPreview)
	mux.HandleFunc("PUT /api/v1/games/{gameID}/seats/{seatID}/planning-draft", server.handlePlanningDraft)
	mux.HandleFunc("GET /api/v1/games/{gameID}/observer/snapshot", server.handleObserverSnapshot)
	mux.HandleFunc("GET /api/v1/games/{gameID}/live-snapshot", server.handleLiveSnapshotExport)
	mux.HandleFunc("POST /api/v1/games/import", server.handleLiveSnapshotImport)
	mux.HandleFunc("PUT /api/v1/games/{gameID}/live-snapshot", server.handleLiveSnapshotRestore)
	mux.HandleFunc("POST /api/v1/games/{gameID}/turn-submissions", server.handleTurnSubmission)
	mux.HandleFunc("POST /api/v1/games/{gameID}/immediate-commands", server.handleImmediateCommand)
	mux.HandleFunc("POST /api/v1/games/{gameID}/battles/{battleID}/commands", server.handleBattleCommand)
	mux.HandleFunc("GET /api/v1/games/{gameID}/stream", server.handleStream)
	if cfg.StaticFS != nil {
		mux.Handle("GET /", spaHandler{fsys: cfg.StaticFS})
	}
	return mux, nil
}

func (s *apiServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (s *apiServer) handleGames(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.host.ListGames())
}

func (s *apiServer) handlePlayerSnapshot(w http.ResponseWriter, r *http.Request) {
	seatID, err := parseSeatID(r.PathValue("seatID"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	snapshot, err := s.host.PlayerSnapshot(r.PathValue("gameID"), seatID)
	if err != nil {
		writeHostError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (s *apiServer) handlePlanningPreview(w http.ResponseWriter, r *http.Request) {
	if !validateMutationRequest(w, r) {
		return
	}
	seatID, err := parseSeatID(r.PathValue("seatID"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	var batch protocol.CommandBatch
	if err := decodeJSON(w, r, &batch); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	gameID := r.PathValue("gameID")
	if batch.GameID != gameID || batch.SeatID != seatID {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "body game_id/seat_id does not match planning-preview route")
		return
	}
	preview, err := s.host.PlanningPreview(gameID, seatID, batch)
	if err != nil {
		writeHostError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, preview)
}

func (s *apiServer) handlePlanningDraft(w http.ResponseWriter, r *http.Request) {
	if !validateMutationRequest(w, r) {
		return
	}
	seatID, err := parseSeatID(r.PathValue("seatID"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	var draft app.PlanningDraft
	if err := decodeJSON(w, r, &draft); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	gameID := r.PathValue("gameID")
	if draft.GameID != gameID || draft.SeatID != seatID {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "body game_id/seat_id does not match planning-draft route")
		return
	}
	snapshot, err := s.host.SavePlanningDraft(gameID, seatID, draft)
	if err != nil {
		writeHostError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (s *apiServer) handleObserverSnapshot(w http.ResponseWriter, r *http.Request) {
	if !s.observerEnabled {
		writeAPIError(w, http.StatusForbidden, "forbidden", "observer API is disabled")
		return
	}
	snapshot, err := s.host.ObserverSnapshot(r.PathValue("gameID"))
	if err != nil {
		writeHostError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (s *apiServer) handleTurnSubmission(w http.ResponseWriter, r *http.Request) {
	if !validateMutationRequest(w, r) {
		return
	}
	var batch protocol.CommandBatch
	if err := decodeJSON(w, r, &batch); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	gameID := r.PathValue("gameID")
	if batch.GameID != gameID {
		writeAPIError(w, http.StatusBadRequest, "bad_request", fmt.Sprintf("body game_id %q does not match path game %q", batch.GameID, gameID))
		return
	}
	receipt, err := s.host.SubmitTurn(gameID, batch)
	if err != nil {
		writeHostError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, receipt)
}

func (s *apiServer) handleImmediateCommand(w http.ResponseWriter, r *http.Request) {
	if !validateMutationRequest(w, r) {
		return
	}
	var request commandRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if err := validateImmediateCommandRequest(request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	receipt, err := s.host.SubmitImmediateCommand(r.PathValue("gameID"), request.SeatID, request.BaseRevision, request.Command)
	if err != nil {
		writeHostError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, receipt)
}

func (s *apiServer) handleBattleCommand(w http.ResponseWriter, r *http.Request) {
	if !validateMutationRequest(w, r) {
		return
	}
	battleID, err := strconv.ParseUint(r.PathValue("battleID"), 10, 64)
	if err != nil || battleID == 0 {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "battleID must be a positive integer")
		return
	}
	var request commandRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if err := validateCommandRequest(request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	receipt, err := s.host.SubmitBattleCommand(r.PathValue("gameID"), battleID, request.SeatID, request.Command)
	if err != nil {
		writeHostError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, receipt)
}

func (s *apiServer) handleStream(w http.ResponseWriter, r *http.Request) {
	notifications, cancelSubscription, err := s.host.Subscribe(r.PathValue("gameID"))
	if err != nil {
		writeHostError(w, err)
		return
	}
	defer cancelSubscription()

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")
	closed := conn.CloseRead(context.Background())

	for {
		select {
		case <-closed.Done():
			return
		case notification, ok := <-notifications:
			if !ok {
				return
			}
			writeCtx, cancel := context.WithTimeout(closed, 10*time.Second)
			err := wsjson.Write(writeCtx, conn, notification)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

func validateImmediateCommandRequest(request commandRequest) error {
	if request.SchemaVersion != app.SchemaVersion {
		return fmt.Errorf("unsupported request schema_version %d", request.SchemaVersion)
	}
	if request.SeatID == 0 {
		return fmt.Errorf("seat_id must be non-zero")
	}
	if request.BaseRevision == 0 {
		return fmt.Errorf("base_revision must be non-zero")
	}
	if err := request.Command.Validate(1); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}
	return nil
}

func validateCommandRequest(request commandRequest) error {
	if request.SchemaVersion != app.SchemaVersion {
		return fmt.Errorf("unsupported request schema_version %d", request.SchemaVersion)
	}
	if request.SeatID == 0 {
		return fmt.Errorf("seat_id must be non-zero")
	}
	if err := request.Command.Validate(request.Command.Sequence); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}
	return nil
}
func validateMutationRequest(w http.ResponseWriter, r *http.Request) bool {
	if !sameOrigin(r) {
		writeAPIError(w, http.StatusForbidden, "forbidden", "request origin does not match host")
		return false
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "Content-Type must be application/json")
		return false
	}
	return true
}

func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" {
		return false
	}
	return strings.EqualFold(parsed.Host, r.Host)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("decode JSON: trailing value")
		}
		return fmt.Errorf("decode JSON trailing data: %w", err)
	}
	return nil
}

func parseSeatID(raw string) (protocol.SeatID, error) {
	value, err := strconv.ParseUint(raw, 10, 32)
	if err != nil || value == 0 {
		return 0, fmt.Errorf("seatID must be a positive integer")
	}
	return protocol.SeatID(value), nil
}

func writeHostError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, app.ErrNotFound):
		writeAPIError(w, http.StatusNotFound, "not_found", err.Error())
	case errors.Is(err, app.ErrSessionRejected):
		writeAPIError(w, http.StatusConflict, "session_rejected", err.Error())
	case errors.Is(err, app.ErrInvalidSave):
		writeAPIError(w, http.StatusBadRequest, "invalid_save", err.Error())
	case errors.Is(err, app.ErrRulesetMismatch):
		writeAPIError(w, http.StatusConflict, "ruleset_mismatch", err.Error())
	case errors.Is(err, app.ErrGameExists):
		writeAPIError(w, http.StatusConflict, "game_exists", err.Error())
	case errors.Is(err, app.ErrGameIDMismatch):
		writeAPIError(w, http.StatusConflict, "game_id_mismatch", err.Error())
	case errors.Is(err, app.ErrPersistenceUnsupported):
		writeAPIError(w, http.StatusConflict, "persistence_unsupported", err.Error())
	default:
		writeAPIError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	writeJSONStatus(w, status, apiErrorEnvelope{SchemaVersion: app.SchemaVersion, Error: apiError{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	writeJSONStatus(w, status, value)
}

func writeJSONStatus(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

type spaHandler struct {
	fsys fs.FS
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/healthz" {
		http.NotFound(w, r)
		return
	}
	assetPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if assetPath == "." || assetPath == "" {
		assetPath = "index.html"
	}
	if info, err := fs.Stat(h.fsys, assetPath); err != nil || info.IsDir() {
		assetPath = "index.html"
	}
	data, err := fs.ReadFile(h.fsys, assetPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	info, _ := fs.Stat(h.fsys, assetPath)
	modTime := time.Time{}
	if info != nil {
		modTime = info.ModTime()
	}
	http.ServeContent(w, r, assetPath, modTime, bytes.NewReader(data))
}
