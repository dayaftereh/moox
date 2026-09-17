package app

import (
	"fmt"
	"strings"

	"moox/internal/game"
	"moox/internal/protocol"
	"moox/internal/session"
)

type PlayerControllerSpec struct {
	SeatID     protocol.SeatID        `json:"seat_id"`
	Controller session.ControllerType `json:"controller"`
}

type CreateGameRequest struct {
	GameID      string
	Seed        uint64
	Settings    game.NewGameSettings
	Controllers []PlayerControllerSpec
}

type CreateGameResult struct {
	Game    GameSummary                `json:"game"`
	Players []game.NewGamePlayerResult `json:"players"`
}

func NewHostWithNewGame(rules *game.EconomyRules) (*Host, error) {
	if rules == nil {
		return nil, fmt.Errorf("new game rules must not be nil")
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		return nil, fmt.Errorf("create economy resolver: %w", err)
	}
	host := NewHost()
	host.newGameRules = rules
	host.newGameResolver = resolver
	host.newGameImmediateResolver = resolver
	return host, nil
}

func (h *Host) NewGameRaceCatalog() (game.PresetRaceCatalog, error) {
	if h == nil {
		return game.PresetRaceCatalog{}, fmt.Errorf("host is nil")
	}
	h.mu.RLock()
	rules := h.newGameRules
	h.mu.RUnlock()
	if rules == nil {
		return game.PresetRaceCatalog{}, fmt.Errorf("new game creation is not configured")
	}
	return game.PresetRaceCatalogFromRules(rules)
}

func (h *Host) NewGameGalaxyCatalog() (game.GalaxyCatalog, error) {
	if h == nil {
		return game.GalaxyCatalog{}, fmt.Errorf("host is nil")
	}
	h.mu.RLock()
	rules := h.newGameRules
	h.mu.RUnlock()
	if rules == nil {
		return game.GalaxyCatalog{}, fmt.Errorf("new game creation is not configured")
	}
	return game.GalaxyCatalogFromRules(rules)
}

func (h *Host) NewGameTechnologyCatalog() (game.NewGameTechnologyCatalog, error) {
	if h == nil {
		return game.NewGameTechnologyCatalog{}, fmt.Errorf("host is nil")
	}
	h.mu.RLock()
	rules := h.newGameRules
	h.mu.RUnlock()
	if rules == nil {
		return game.NewGameTechnologyCatalog{}, fmt.Errorf("new game creation is not configured")
	}
	return game.TechnologyLevelCatalog(), nil
}
func (h *Host) NewGameCompositionCatalog() (game.NewGameCompositionCatalog, error) {
	if h == nil {
		return game.NewGameCompositionCatalog{}, fmt.Errorf("host is nil")
	}
	h.mu.RLock()
	rules := h.newGameRules
	h.mu.RUnlock()
	if rules == nil {
		return game.NewGameCompositionCatalog{}, fmt.Errorf("new game creation is not configured")
	}
	return game.NewGameCompositionCatalogForRules(rules)
}

func validateNewGameControllerComposition(settings game.NewGameSettings, controllers []PlayerControllerSpec) error {
	playerSeats := make(map[protocol.SeatID]int, len(settings.Players))
	for i, player := range settings.Players {
		playerSeats[player.SeatID] = i
	}
	seen := make(map[protocol.SeatID]struct{}, len(controllers))
	for _, controller := range controllers {
		if controller.SeatID == 0 {
			return fmt.Errorf("controller seat ID must be non-zero")
		}
		if _, exists := seen[controller.SeatID]; exists {
			return fmt.Errorf("duplicate controller assignment for seat %d", controller.SeatID)
		}
		seen[controller.SeatID] = struct{}{}
		index, ok := playerSeats[controller.SeatID]
		if !ok {
			return fmt.Errorf("controller assignment references unknown player seat %d", controller.SeatID)
		}
		expected := session.ControllerBuiltinAI
		if index == 0 {
			expected = session.ControllerLocalHuman
		}
		if controller.Controller != expected {
			return fmt.Errorf("seat %d controller %q does not match Slice 16.6 composition; want %q", controller.SeatID, controller.Controller, expected)
		}
	}
	return nil
}

func (h *Host) CreateGame(request CreateGameRequest) (CreateGameResult, error) {
	if h == nil {
		return CreateGameResult{}, fmt.Errorf("host is nil")
	}
	gameID := strings.TrimSpace(request.GameID)
	if gameID == "" {
		return CreateGameResult{}, fmt.Errorf("game ID must not be empty")
	}
	if gameID != request.GameID {
		return CreateGameResult{}, fmt.Errorf("game ID must not have leading or trailing whitespace")
	}

	h.mu.RLock()
	_, exists := h.games[gameID]
	rules := h.newGameRules
	resolver := h.newGameResolver
	immediateResolver := h.newGameImmediateResolver
	h.mu.RUnlock()
	if exists {
		return CreateGameResult{}, fmt.Errorf("%w: game %q", ErrGameExists, gameID)
	}
	if rules == nil || resolver == nil || immediateResolver == nil {
		return CreateGameResult{}, fmt.Errorf("new game creation is not configured")
	}

	if err := validateNewGameControllerComposition(request.Settings, request.Controllers); err != nil {
		return CreateGameResult{}, fmt.Errorf("%w: %v", game.ErrInvalidNewGameSettings, err)
	}

	controllers := make(map[protocol.SeatID]session.ControllerType, len(request.Controllers))
	for _, controller := range request.Controllers {
		if controller.SeatID == 0 {
			return CreateGameResult{}, fmt.Errorf("controller seat ID must be non-zero")
		}
		if _, exists := controllers[controller.SeatID]; exists {
			return CreateGameResult{}, fmt.Errorf("duplicate controller assignment for seat %d", controller.SeatID)
		}
		controllers[controller.SeatID] = controller.Controller
	}
	settings := request.Settings
	settings.Players = append([]game.NewGamePlayerSpec(nil), request.Settings.Players...)
	for i := range settings.Players {
		if controller, ok := controllers[settings.Players[i].SeatID]; ok {
			settings.Players[i].BuiltinAIControlled = controller == session.ControllerBuiltinAI
		}
	}
	generated, err := rules.NewGame(request.Seed, settings)
	if err != nil {
		return CreateGameResult{}, err
	}
	seats := make([]session.Seat, len(generated.Players))
	seenControllerSeats := make(map[protocol.SeatID]struct{}, len(generated.Players))
	for i, player := range generated.Players {
		controller := session.ControllerLocalHuman
		if configured, ok := controllers[player.SeatID]; ok {
			controller = configured
			seenControllerSeats[player.SeatID] = struct{}{}
		}
		seats[i] = session.Seat{
			ID:         player.SeatID,
			EmpireID:   player.EmpireID,
			Name:       player.Name,
			Controller: controller,
		}
	}
	for seatID := range controllers {
		if _, ok := seenControllerSeats[seatID]; !ok {
			return CreateGameResult{}, fmt.Errorf("controller assignment references unknown generated seat %d", seatID)
		}
	}
	gameSession, err := session.NewGameSession(gameID, generated.State, seats)
	if err != nil {
		return CreateGameResult{}, fmt.Errorf("create game session: %w", err)
	}
	if err := h.Register(Registration{Session: gameSession, Resolver: resolver, ImmediateResolver: immediateResolver}); err != nil {
		return CreateGameResult{}, err
	}
	summary, err := h.gameSummary(gameID)
	if err != nil {
		return CreateGameResult{}, err
	}
	return CreateGameResult{Game: summary, Players: generated.Players}, nil
}

func (h *Host) gameSummary(gameID string) (GameSummary, error) {
	hosted, err := h.lookup(gameID)
	if err != nil {
		return GameSummary{}, err
	}
	hosted.mu.Lock()
	defer hosted.mu.Unlock()
	status := hosted.session.Status()
	return GameSummary{
		SchemaVersion:  SchemaVersion,
		GameID:         status.GameID,
		ChangeSequence: hosted.changeSequence,
		Revision:       status.Revision,
		Turn:           status.Turn,
		Phase:          status.Phase,
	}, nil
}
