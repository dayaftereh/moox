package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"moox/internal/app"
	"moox/internal/core"
	"moox/internal/game"
	"moox/internal/server"
	"moox/internal/session"
)

const demoGameID = "demo"
const standardReferenceGameID = "game-1"
const triangleReferenceGameID = "game-triangle-2pc"
const defaultHTTPAddress = "127.0.0.1:7171"

func main() {
	var (
		addr                  = flag.String("addr", defaultHTTPAddress, "HTTP listen address")
		rulesDir              = flag.String("rules", filepath.Join("data", "rulesets", "moo2-1.31"), "normalized ruleset directory")
		webDir                = flag.String("web", filepath.Join("web", "dist"), "built web asset directory; omitted if index.html is absent")
		enableObserver        = flag.Bool("enable-observer", false, "enable privileged observer snapshot endpoint")
		enablePersistence     = flag.Bool("enable-persistence", false, "enable privileged live save/import/restore endpoints")
		demoFixture           = flag.Bool("demo-fixture", false, "development only: pre-register the legacy core.NewSmallFixture demo game")
		referenceGames        = flag.Bool("reference-games", false, "development only: pre-register game-1 plus the durable 3-player 2pc triangle reference game")
		allowInsecureNonLocal = flag.Bool("insecure-allow-nonloopback", false, "UNSAFE: allow unauthenticated direct non-loopback binding")
	)
	flag.Parse()
	if !*allowInsecureNonLocal && !isLoopbackAddress(*addr) {
		log.Fatalf("refusing non-loopback listen address %q without -insecure-allow-nonloopback; use a trusted VPN/reverse proxy/auth/TLS boundary for remote exposure", *addr)
	}
	host, err := newServerHost(*rulesDir, *demoFixture, *referenceGames)
	if err != nil {
		log.Fatal(err)
	}
	assets := optionalWebFS(*webDir)
	handler, err := server.NewHandler(server.Config{Host: host, StaticFS: assets, ObserverEnabled: *enableObserver, PersistenceEnabled: *enablePersistence})
	if err != nil {
		log.Fatal(err)
	}
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	httpServer := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP shutdown: %v", err)
		}
	}()
	log.Printf("MOOX development server listening on http://%s (games=%d observer=%t persistence=%t)", listener.Addr(), len(host.ListGames()), *enableObserver, *enablePersistence)
	if assets == nil {
		log.Printf("web assets not found at %s; API/WS only (use Vite dev server or build web/)", *webDir)
	}
	if err := httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func newServerHost(rulesDir string, demoFixture bool, referenceGames bool) (*app.Host, error) {
	rules, err := game.LoadEconomyRules(rulesDir)
	if err != nil {
		return nil, fmt.Errorf("load rules: %w", err)
	}
	host, err := app.NewHostWithNewGame(rules)
	if err != nil {
		return nil, fmt.Errorf("configure new game host: %w", err)
	}
	if referenceGames {
		if err := registerReferenceGames(host, rules); err != nil {
			return nil, err
		}
	}
	if !demoFixture {
		return host, nil
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		return nil, fmt.Errorf("create demo economy resolver: %w", err)
	}
	state := core.NewSmallFixture(0x8008)
	gameSession, err := session.NewGameSession(demoGameID, state, []session.Seat{{ID: 1, EmpireID: state.Empires[0].ID, Name: "Developer", Controller: session.ControllerLocalHuman}})
	if err != nil {
		return nil, fmt.Errorf("create demo session: %w", err)
	}
	if err := host.Register(app.Registration{Session: gameSession, Resolver: resolver, ImmediateResolver: resolver}); err != nil {
		return nil, fmt.Errorf("register demo session: %w", err)
	}
	return host, nil
}

func registerReferenceGames(host *app.Host, rules *game.EconomyRules) error {
	if host == nil || rules == nil {
		return fmt.Errorf("reference games require host and rules")
	}
	if _, err := host.CreateGame(app.CreateGameRequest{
		GameID: standardReferenceGameID,
		Seed:   0x8009,
		Settings: game.NewGameSettings{
			GalaxySize:      game.GalaxySizeSmall,
			GalaxyAge:       game.GalaxyAgeNormal,
			TechnologyLevel: game.NewGameTechnologyAverage,
			StrategicCombat: false,
			Players: []game.NewGamePlayerSpec{
				{SeatID: 1, EmpireName: "Human", RaceID: "human"},
				{SeatID: 2, EmpireName: "Darlok", RaceID: "darlok"},
			},
		},
		Controllers: []app.PlayerControllerSpec{
			{SeatID: 1, Controller: session.ControllerLocalHuman},
			{SeatID: 2, Controller: session.ControllerBuiltinAI},
		},
	}); err != nil {
		return fmt.Errorf("register standard reference game: %w", err)
	}

	generated, err := rules.NewReferenceTriangleGame(game.ReferenceTriangleSeed)
	if err != nil {
		return fmt.Errorf("build triangle reference game: %w", err)
	}
	resolver, err := game.NewEconomyResolver(rules)
	if err != nil {
		return fmt.Errorf("create triangle reference resolver: %w", err)
	}
	seats := make([]session.Seat, len(generated.Players))
	for i, player := range generated.Players {
		controller := session.ControllerBuiltinAI
		if player.SeatID == 1 {
			controller = session.ControllerLocalHuman
		}
		seats[i] = session.Seat{ID: player.SeatID, EmpireID: player.EmpireID, Name: player.Name, Controller: controller}
	}
	gameSession, err := session.NewGameSession(triangleReferenceGameID, generated.State, seats)
	if err != nil {
		return fmt.Errorf("create triangle reference session: %w", err)
	}
	if err := host.Register(app.Registration{Session: gameSession, Resolver: resolver, ImmediateResolver: resolver}); err != nil {
		return fmt.Errorf("register triangle reference session: %w", err)
	}
	return nil
}
func optionalWebFS(dir string) fs.FS {
	if strings.TrimSpace(dir) == "" {
		return nil
	}
	if info, err := os.Stat(filepath.Join(dir, "index.html")); err != nil || info.IsDir() {
		return nil
	}
	return os.DirFS(dir)
}

func isLoopbackAddress(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
