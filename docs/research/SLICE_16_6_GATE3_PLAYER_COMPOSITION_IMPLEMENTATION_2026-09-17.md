# Slice 16.6 Gate 3 - Player composition implementation

Date: 2026-09-17

Status: **Gate 3 complete; Gate 4 acceptance next**.

Frozen contract: `docs/research/SLICE_16_6_GATE2_PLAYER_COMPOSITION_FREEZE_2026-09-17.md`.

## Runtime and server implementation

Gate 3 implements the frozen 1-2 opponent V1 boundary without enabling the planned 3-7 counts.

### Composition catalog

Added the server-owned composition catalog and public route:

```text
GET /api/v1/new-game/compositions
```

The catalog exposes:

- original-facing opponent range 1-7;
- count 1/2 as `supported`;
- count 3-7 as `planned` with reason `additional_ai_race_breadth_required`;
- explicit max-supported-opponents = 2 for Small/Medium/Large/Huge;
- the four frozen Human/Klackon assignment entries;
- Built-in AI controller facts and canonical default opponent Empire names.

The create path remains authoritative. Catalog visibility does not enable a planned count.

### Frozen 2/3-player validation

Game-layer New Game now accepts only the exact frozen classes:

- 2 or 3 player specs;
- non-zero, unique, strictly ascending seat IDs;
- unique race IDs;
- non-empty, case-insensitively unique Empire names;
- local race Human or Klackon;
- seat 2 race Darlok;
- seat 3, when present, is exactly the other supported Human/Klackon race.

The public HTTP New Game path requires explicit controller assignments for every submitted player. Explicit controller roles are validated as:

```text
seat 1 -> local_human
seat 2 -> builtin_ai
seat 3 -> builtin_ai (when present)
```

Generic internal `Host.CreateGame` behavior without an explicit controller list remains backward compatible for non-HTTP/internal scenario callers.

## Deterministic home-system implementation

`selectNewGameHomeSystems` replaces the pair-only call site.

- First two homes use the existing `farthestNewGameSystemPair` unchanged.
- Two-player creation therefore preserves the existing home pair and ordering.
- A third home uses deterministic maximin distance to the already selected pair.
- Equal maximin distance uses lower System ID, then lower source index.
- Home-system selection consumes no additional RNG draws.

A dedicated regression freezes the maximin tie break and two-player pair compatibility.

## Deterministic game matrix

`TestNewGameGate3AcceptedCompositionMatrixDeterministic` exercises the complete accepted Game-layer product matrix at a canonical seed:

```text
4 galaxy sizes
x 3 galaxy ages
x 5 difficulties
x 2 local races
x 2 supported technology levels
x 2 opponent counts
= 480 tuples
```

Every tuple is created twice and required to be deeply equal, valid and to contain the frozen Empire/home count.

Additional game regressions verify:

- composition catalog range/support/reason contract;
- exact automatic opponent race assignment;
- three-player Pre-Warp zero-ship bootstrap;
- three-player Average 6 ships / 3 designs / 6 strategic fleets;
- AI flags for both opponent Empires.

## App/server and persistence coverage

App and HTTP regressions cover:

- Human + Darlok + Klackon 3-player creation;
- seat controller/race projection into player snapshots;
- rejection of a remote-human opponent controller;
- HTTP composition-catalog shape;
- HTTP 3-player creation.

Persistence coverage creates a 3-player / 2-AI game separately for Pre-Warp and Average and verifies:

- live snapshot export;
- import into a fresh persistence-enabled server;
- byte-identical re-export;
- preserved Human/Darlok/Klackon races and local/AI/AI controllers;
- live-snapshot restore and resumed change sequence.

## Production opponent-count artwork

Added seven deterministic production SVGs:

```text
/assets/new-game/opponent-count/1.svg
...
/assets/new-game/opponent-count/7.svg
```

The generator uses geometric/path/rect based numerals and count-dependent orbital nodes. Assets contain no runtime `<text>` and are registered in the New Game manifest.

`check:opponent-composition` freezes:

- all seven files;
- count/signature markers;
- manifest registration;
- no SVG text elements;
- composition API binding;
- selector and launch-briefing integration.

## New Game HMI integration

The Web client now fetches `/api/v1/new-game/compositions` and renders a shared VisualSelector for opponent count.

Behavior:

- browsable counts 1 through 7;
- 1 and 2 supported;
- 3 through 7 planned;
- planned states show the localized server reason and disable Create Game;
- supported states resolve opponent assignments from the server catalog;
- opponent cards show the existing Slice-16.4 race portraits and AI label;
- no per-opponent race editor;
- no random opponent option;
- the old editable Darlok-opponent-name field is removed;
- the create request is built from the resolved assignment and emits the exact local/AI controller list.

Changing Human/Klackon automatically changes the second opponent in the two-opponent class through the catalog assignment rather than frontend hard-coding.

## Launch briefing

A final launch briefing is rendered from the same selected state consumed by request construction.

It contains:

- galaxy size + age;
- difficulty;
- local race + editable player Empire name;
- starting technology;
- opponent count + resolved opponent races;
- exact submitted seed.

At 390 px the summary becomes a single-column compact layout. Opponent chips wrap without horizontal page overflow.

## Browser / responsive verification

The real Chrome New Game selector smoke now includes Opponent Composition in addition to Difficulty, Galaxy Size, Galaxy Age, Starting Technology and Player Race.

The smoke verifies at a real emulated 390 px viewport:

- count 1 -> supported, one opponent chip, Create enabled;
- count 2 -> supported, two opponent chips, Create enabled;
- count 3 -> planned, no fabricated chips, Create disabled, localized reason visible;
- correct 1200x675 opponent SVGs;
- no horizontal overflow;
- launch briefing contains the resolved opponents.

The same existing smoke continues through desktop layout verification.

Current browser result: **35 browsable New Game options, mobile + desktop pass**.

## Live runtime verification

The persistent backend service was recycled onto the Gate-3 code after compilation/tests. The live endpoint returns HTTP 200 and the served production JS bundle contains both the composition endpoint binding and the German launch briefing.

## Gate-3 verification

Completed checks:

- targeted game/app/server Gate-3 regressions - pass;
- 480-tuple deterministic Game matrix - pass;
- three-player Pre-Warp/Average persistence round trip - pass;
- `npm run build` including `check:opponent-composition` - pass;
- `npm run check:new-game-selector:browser` - pass;
- `go vet ./...` - pass;
- `git diff --check` - pass.

- `go test ./... -count=1` - pass across the complete repository.

## Gate-4 handoff

Gate 4 should verify the frozen/implemented product rather than redesign it:

- live desktop + 390 px walkthrough;
- actual create/play smoke for supported 1- and 2-opponent classes;
- Human and Klackon local starts;
- Pre-Warp and Average;
- representative Small/Huge and Easy/Impossible combinations;
- planned counts 3-7 cannot create through UI or API;
- launch briefing matches submitted state;
- persistence/resume remains intact;
- full repository tests/vet/Web/browser/diff checks;
- Slice 16.6 closure and handoff to Slice 16.7 Advanced parity.

Advanced Starting Technology remains intentionally outside this implementation.