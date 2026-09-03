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

	generated, err := rules.NewGame(request.Seed, request.Settings)
	if err != nil {
		return CreateGameResult{}, err
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
