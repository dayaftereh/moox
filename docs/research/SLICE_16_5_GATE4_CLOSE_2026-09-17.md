# Slice 16.5 Gate 4 close - Starting Technology visual selector

Date: 2026-09-17

Status: **Gate 4 complete; Slice 16.5 closed**.

## User acceptance

The live New Game selector was reviewed after the Gate-3 artwork and scaling fixes. The user accepted the current three-image direction as the Slice-16.5 baseline. Deeper artwork refinement is explicitly future polish and is not a closure blocker.

Frozen visible labels:

- Pre-Warp
- Durchschnitt
- Fortschrittlich

Advanced/Fortschrittlich remains catalog-visible but planned/locked. Runtime parity remains owned by Slice 16.7.

## Exact frozen runtime state

Gate 4 adds `internal/game/new_game_technology_gate4_test.go` so the accepted runtime levels are checked against the Gate-2 freeze rather than only broad ship counts.

For Human and Klackon player starts, both accepted technology levels freeze the common non-Advanced baseline per Empire:

- exactly one home colony and that colony is the Empire capital;
- Population 8;
- 4 Farmers / 2 Workers / 2 Scientists;
- 50 BC treasury balance;
- 0 freighters.

Pre-Warp additionally freezes:

- completed/known starting fields exactly `0, 29`;
- known normalized Technology IDs exactly `32, 40, 103, 145, 166, 168`;
- zero starting `Ships`;
- zero `ShipDesigns`;
- zero `StrategicFleets`.

Average additionally freezes:

- completed/known starting fields exactly `0, 22, 23, 28, 29, 55, 57`;
- exactly 20 known Tactical-combat applications under the frozen Tactical setting;
- two Scout ship entries per Empire;
- one combat strategic fleet per Empire containing those two Scouts;
- one Colony Ship special strategic fleet per Empire;
- one starting Scout design per Empire.

Advanced is still rejected by the authoritative New Game runtime and must not silently fall back to Average.

## Determinism

`TestNewGameTechnologyGate4RepeatedDeterminism` repeats the accepted Human/Klackon x Pre-Warp/Average matrix for seeds `0x1653`, `0x1654`, and `0x1655`.

For every case it requires both:

- deep equality of the returned result;
- byte-equal JSON encoding of the authoritative game state.

The existing Gate-3 Advanced-rejection and catalog-boundary tests remain in force.

## Browser / HMI closure

`web/scripts/check-new-game-selector-browser.mjs` now includes Starting Technology in the real Chrome smoke path.

At a real emulated 390 x 844 mobile viewport it verifies all three options:

- correct German one-line labels;
- correct versioned SVG path;
- native 1200 x 675 SVG load;
- responsive image scaling into the selector frame;
- no horizontal overflow;
- at least 44 px arrow targets;
- server-owned `supported` / `planned` availability;
- Create Game remains enabled for Pre-Warp and Average and locked for Advanced;
- distinct 64 x 36 rendered pixel fingerprints for all three artworks.

The same smoke then switches to a 1280 x 900 desktop viewport and verifies responsive Technology art sizing, 20 px desktop radius, one-line title fit, at least 52 px arrow targets, and no horizontal overflow.

Observed stable artwork fingerprints in the accepted live build:

- Pre-Warp: `34bb2c49`
- Durchschnitt / Average: `ba1e4ffe`
- Fortschrittlich / Advanced: `2715f5f4`

The browser smoke passed with 28 browsable New Game options across Difficulty, Galaxy Size, Galaxy Age, Starting Technology, and Player Race.

## Verification

Gate-4 verification completed successfully:

- `go test ./internal/game -run TestNewGameTechnologyGate4 -count=1`
- `go test ./... -count=1`
- `npm run check:new-game-selector:browser`
- `go vet ./...`
- `npm run build`
- `git diff --check`

The final repository closeout keeps the working tree clean after commit.

## Closure

Slice 16.5 satisfies all four Gate-4 checks:

- exact accepted start states are frozen and tested;
- equal settings remain deterministic across repeated multi-seed fixtures;
- the three progression artworks are distinct and usable on desktop and 390 px mobile;
- repository tests, vet, browser checks, web build and diff checks pass.

Next prepared step: **Slice 16.6 Gate 1 - player composition and final New Game integration audit**.

Advanced Starting Technology parity remains explicitly deferred to **Slice 16.7** so it cannot be forgotten or accidentally implemented as a UI-only toggle.