# Slice 16.6 Gate 4 close - Player composition and final New Game integration

Date: 2026-09-18

Status: **Gate 4 complete; Slice 16.6 closed**.

## User acceptance and final HMI baseline

The live New Game flow was reviewed through the Gate-4 polish pass and accepted as the Slice-16.6 baseline.

The final composition presentation keeps `Player Race` and `Opponent Count` as equal-height image-led selector cards. Resolved player/AI composition is shown separately in the full-width `Players & Empires` roster. Real race tiles expose a `?` information control using the existing race-trait facts; planned opponent slots remain neutral and do not fabricate unavailable races.

The final New Game text form intentionally exposes only Seed and Player Empire. Game ID is server-owned. The Seed field includes an accessible dice-icon reroll that creates a new explicit 64-bit hexadecimal seed through browser crypto; deterministic gameplay continues from that visible seed.

## Supported composition acceptance

Gate 4 adds `internal/server/new_game_gate4_test.go`.

`TestHTTPGate4SupportedClassCreatePlaySmoke` exercises public HTTP create/play coverage across every supported class in the frozen Slice-16.6 breadth:

- Difficulty: Easy, Normal, Hard, Very Hard, Impossible.
- Galaxy size: Small, Medium, Large, Huge.
- Galaxy age: Mineral Rich, Normal, Organic Rich.
- Starting technology: Pre-Warp and Average.
- Local preset race: Human and Klackon.
- Opponent count: 1 and 2, including three-player/two-built-in-AI sessions.

Every case creates through `POST /api/v1/games`, reads the local seat snapshot, submits the local human turn and proves the built-in AI controllers finish resolution so the game advances to the next turn.

## Unsupported boundary

`TestHTTPGate4UnsupportedCombinationsRejected` proves the public create API rejects without creating a game:

- four total players / three opponents;
- Advanced starting technology;
- planned Darlok as the local player race.

The real browser smoke separately proves planned opponent counts stay locked in the HMI, including the seven-opponent view with eight visible roster slots and seven neutral planned KI slots.

## Determinism and persistence

The existing Gate-3 deterministic matrix remains the authoritative complete-tuple regression:

- 4 galaxy sizes x 3 ages x 5 difficulties x 2 local races x 2 supported technology starts x 2 supported opponent counts = **480 tuples**;
- each accepted tuple is created twice from the same seed/settings and must match deterministically.

Gate 4 re-ran that matrix successfully.

Three-player persistence round trips were re-run for both Pre-Warp and Average. Export/import/restore preserves the Human/Darlok/Klackon composition and local/AI/AI controller ownership.

A live server acceptance on `127.0.0.1:7171` also created `gate4-live-20260918` as Human + Darlok AI + Klackon AI on Huge / Organic Rich / Impossible / Average, advanced turn 1 -> 2 through built-in AI resolution, exported the live snapshot, restored it and verified turn 2 plus the Human local race remained intact.

## Browser / responsive closure

The real Chrome smoke passes at 390 px mobile and desktop. The accepted flow verifies:

- all New Game visual selectors remain browsable without horizontal overflow;
- Player Race and Opponent Count remain equal-height selector cards on desktop;
- the full-width `Players & Empires` roster uses four desktop columns and two columns at 390 px;
- one opponent resolves to two Empire tiles; two opponents resolve to three;
- seven planned opponents render eight total tiles without inventing races;
- `?` race information opens/closes accessibly and displays the existing race portrait/facts;
- supported compositions leave Create enabled and planned compositions lock it;
- the launch briefing tracks the submitted galaxy, difficulty, player race, technology, resolved opponents and exact seed;
- the server-owned Game ID and redundant Technology / Combat text fields remain absent;
- the dice seed control is an accessible icon-only touch target and updates both Seed and launch briefing.

## Verification

Gate-4 verification completed successfully:

- `go test ./internal/server -run "TestHTTPGate4SupportedClassCreatePlaySmoke|TestHTTPGate4UnsupportedCombinationsRejected|TestHTTPPersistenceThreePlayerCompositionRoundTrip" -count=1`
- `go test ./internal/game -run "TestNewGameGate3AcceptedCompositionMatrixDeterministic|TestNewGameThreePlayerCompositionCreatesFrozenStartingState" -count=1`
- `npm run check:new-game-selector:browser`
- live `7171` create -> snapshot -> turn -> AI resolution -> export/restore acceptance
- `go test ./... -count=1`
- `go vet ./...`
- `npm run build`
- `git diff --check`

## Closure

Slice 16.6 now satisfies its Gate-4 exit criterion: the accepted player-composition breadth is deterministic and server-authoritative, supported games can be created and played through multi-AI turn resolution, unsupported combinations are blocked in UI/API, persistence remains sound, and the final New Game composition HMI is usable on desktop and 390 px mobile.

Next prepared step: **Slice 16.7 Gate 1 - Advanced starting technology parity and final Slice-16 closure audit**.

Slice 16.7 is prepared only. It is **not opened by this closure**.
