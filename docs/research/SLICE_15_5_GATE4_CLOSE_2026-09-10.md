# Slice 15.5 Gate 4 - Independent Tactical QA close

Date: 2026-09-10
Status: **CLOSED - Gates 1-4 complete**

## Close decision

Slice 15.5 - Interactive 2D Tactical Combat is closed. The frozen 1v1-2v2 Tactical baseline is browser-playable, server-authoritative, participant-safe, responsive, recoverable after rejected/stale commands, and reconciles exact surviving/destroyed strategic ship identities through the frozen Slice-15.4 Battle Return path.

The next queued integration slice is **15.6 - Full browser vertical slice / polish / complete-game QA**.

## Frozen scope delivered

The closed baseline includes:

- 1v1 through 2v2 supported Tactical combat, with 2v2 as the primary acceptance shape;
- authoritative integer X/Y Tactical coordinates;
- 16 discrete Facings and original movement cost: `ceil(Euclidean distance) + facing steps * turn cost`;
- normal/Stabilizer/Nullifier turning-cost model in Battle core;
- current/max movement points and per-round reset;
- server-projected legal movement destinations;
- `battle.move_ship` intent command with server-derived cost/facing/legality;
- existing deterministic standard-Laser fire path integrated with moved positions;
- authoritative Armor -> aggregate Structure damage and ship destruction while internal subsystem damage remains deferred;
- one-ship-at-a-time initiative/activation flow through multi-ship rounds;
- participant-safe Tactical projection with RNG redaction and seat-owned legal action catalogs;
- read-only Tactical Scan for player-visible ships, including weapons/readiness, Armor/Structure, movement/facing and normalized combat ratings;
- responsive SVG battlefield with effectively unbounded presentation, no normal arena edge, pan/zoom/pinch and accessible camera controls;
- accepted Slice-15.3 procedural ship glyphs with semantic SVG fallback;
- stale/illegal command rejection with zero mutation, one refetch on 409 and no automatic resubmission;
- terminal Battle result reconciliation through Slice-15.4 return UI.

## Gate-4 acceptance matrix

### Normal browser entry

Evidence: `SLICE_15_5_GATE3_BLOCK3_TACTICAL_BATTLEFIELD_UI_2026-09-10.md` and `SLICE_15_5_GATE3_BLOCK4B_TACTICAL_VISUALS_FULL_RETURN_QA_2026-09-10.md`.

A real GameSession was driven from Turn 1 / Planning by normal `Fertig` through Encounter creation, Battle Entry and `Taktischen Kampf betreten`. No API-only Battle injection or React-owned Battle construction was used for the interaction path.

### Authoritative movement

Evidence: Block 1 + Block 3.

The browser selected a server-projected legal destination. The authoritative snapshot changed the active Human ship from `(10,9)` to `(11,9)`, Movement from 20 to 19 and command sequence from 1 to 2. React submitted only ship ID and destination X/Y.

### Authoritative Laser fire / damage

Evidence: Blocks 2-4b.

The browser selected only a server-projected Laser target. Fire consumed readiness, advanced command sequence and refreshed damage. Full-battle QA additionally proved Armor exhaustion continues into aggregate Structure and can destroy ships without fabricating deferred subsystem state.

### Multi-ship activation / round progression

Evidence: Blocks 1, 3 and 4b.

A real 2v2 browser Battle progressed through Human activations, Built-in-AI activations and multiple round boundaries. Movement/readiness reset from server authority on new rounds.

### Tactical -> strategic result reconciliation

Evidence: Block 4b.

Full browser Battle completed as Human Tactical Victory. Battle Return showed two destroyed enemies and two visible Human survivors. The authoritative `battle_completed` summary preserved exact IDs:

- destroyed: `[60,61]`;
- surviving: `[57,58]`.

`Weiter zur Strategie` returned to Galaxy / Turn 2 / Planning with no live Battle remaining.

### Rejected/stale/illegal commands

Evidence: `SLICE_15_5_GATE3_BLOCK5_TACTICAL_REJECTION_UX_2026-09-10.md`.

Managed-browser QA covered stale positive sequence, occupied destination and own-ship Laser target. Each 409 case produced:

- exactly one Battle POST;
- exactly one participant snapshot GET/refetch;
- zero authoritative position/sequence mutation;
- no automatic resubmission;
- visible global and in-Battle rejection feedback.

The repeatable server/API regression deep-compares pre/post participant Battle projection and verifies unchanged ChangeSequence/revision for all three rejected command classes.

### Tactical Scan

Evidence: Blocks 2 and 3.

Friendly/enemy ships can be inspected without consuming movement, readiness, activation or RNG. Scan exposes only normalized participant-safe data; no client-side composite strength or fake internal subsystem health is generated.

### Desktop / mobile / open field

Evidence: Block 3.

Desktop drag/wheel and emulated touch one-finger pan/two-finger pinch modify only camera ViewBox. Tactical command sequence remains unchanged during camera interaction. At exactly 320 CSS px there was no horizontal overflow; battlefield remained usable and primary Tactical controls met the 44-px touch target floor. No visible artificial arena edge is rendered.

### Visual integration

Evidence: Block 4b.

A real 2v2 Tactical Battle rendered four combatants as four accepted Slice-15.3 procedural ship glyphs. Facing remains authoritative world-marker rotation; procedural art is presentation only and retains a fallback silhouette.

## Authority boundary close

The closed browser surface does not calculate or own:

- movement legality or cost;
- resulting Facing;
- target legality;
- beam range legality;
- hit roll / RNG;
- damage or destruction;
- active ship / round progression;
- Battle winner/result;
- strategic fleet reconciliation.

Camera, selected mode, selected Scan target and transient UI warning state remain presentation/orchestration state only.

## Explicit deferrals preserved

Closure does not claim support for:

- more than two combat ships per side as an accepted Tactical shape;
- broad shield-facing / firing-arc systems;
- missiles, torpedoes, bombs, fighters or point defense;
- boarding/capture;
- retreat command;
- deployment editor;
- broad special-system/internal-damage simulation;
- planet/station Tactical combat;
- monsters/Antarans;
- animation timing as authority.

These remain future breadth/content work unless explicitly pulled into a later accepted contract.

## Independent final validation

A fresh Gate-4 session, separate from Gate-3 implementation sessions, ran on commit `0effcde` before this close marker:

- `go test ./...` - PASS;
- `go vet ./...` - PASS;
- `npm run build` - PASS;
- `git diff --check` - PASS;
- working tree clean before close-document edits.

No implementation blocker was found in the independent close pass.

## Final status

**Slice 15.5: CLOSED - Gates 1, 2, 3 and 4 complete.**

Milestone achieved: **first interactive server-authoritative 2D Tactical Combat battle playable from the browser and reconciled back into the strategic match**.

Next queued slice: **15.6 - Full browser vertical slice / polish / complete-game QA**.
