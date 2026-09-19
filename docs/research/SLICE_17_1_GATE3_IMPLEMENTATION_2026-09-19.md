# Slice 17.1 Gate 3 - Hulls and Mandatory Ship Systems Implementation

Date: 2026-09-19
Slice: 17.1 - Hulls and Mandatory Ship Systems
Gate: 3 - implementation
Baseline entering Gate 3: `dc1829d` (`docs(slice): freeze slice 17.1 gate 2 contract`)
Recovery session: `ses-20260919T170516-8d6f090dee1f`

## Result

PASS.

Gate 3 implements the frozen Slice-17.1 hull/mandatory-system contract without opening Gate 4 and without entering Slice 17.2 weapon breadth.

## Implemented gameplay/domain behavior

### Six-hull Designer legality

The military Designer now supports the complete frozen hull set:

- Frigate;
- Destroyer;
- Cruiser;
- Battleship;
- Titan;
- Doom Star.

Frigate through Battleship require no dedicated hull technology.
Titan requires technology 186 (`titan_construction`).
Doom Star requires technology 55 (`doom_star_construction`).

The former Frigate-only `current_design_scope` gate is no longer the authority for these six hulls.

### Mandatory-system selection

The server/compiler implements the category-specific Gate-2 contract:

- Warp Drive: automatic current-best Empire capability;
- Fuel Cell: automatic current-best Empire capability;
- Armor: best-known material is applied on explicit design save/refit;
- Computer: explicitly selectable among known legal technologies, defaulting to best known when omitted;
- Shield: explicitly selectable among known legal technologies, including explicit no-Shield.

The save command now accepts:

- optional `computer_id`;
- optional `shield_id`, preserving absent/default versus explicit empty/no-Shield semantics.

Unknown or unowned selections are rejected by domain authority.

### Direct-original miniaturization

A shared compiler now implements the verified MOO2 1.31 miniaturization path:

- successor-field traversal;
- Hyper-Advanced repeat contribution;
- saturation at level 5+;
- component cost factors 100/75/55/40/30/25%;
- Shield Space factors 100/80/65/50/35/25%;
- original positive integer rounding/minimum behavior.

Miniaturization is applied per mandatory component before total design cost is assembled.

### Dynamic effective saved-design projection

Saved design identity/revision and explicit component choices remain stable.

Current effective cost/Space is recompiled from current Empire research when projected to the player.

Research-driven miniaturization can therefore reduce effective cost/Space without incrementing the design revision.

Drive/Fuel strategic movement remains current-Empire authority rather than historical built-design authority.

### Queue-time construction snapshot

Military construction now persists a compiled `MilitaryShipSpec` snapshot in `ConstructionState`.

When a military design enters active construction or the normal construction queue, the server freezes:

- design ID;
- design revision;
- compiled effective ShipDesignSpec;
- effective ProductionCostPP.

Later research or design edits do not reprice an already queued/running project.

Completion materializes the concrete Ship from that frozen construction snapshot.

The state validator accepts the new military snapshot only on military Ship projects and rejects it on non-military construction.

### Cruiser+ Colony infrastructure authority

The original large-hull restriction is enforced server-side:

- Frigate/Destroyer: no orbital-dock requirement from this rule;
- Cruiser/Battleship/Titan/Doom Star: Star Base or better required.

The rule is enforced in both:

- construction-choice projection;
- authoritative military queue resolution.

Star Base, Battlestation and Star Fortress all satisfy the requirement.

A forged/stale queue command cannot bypass the rule.

## Catalog/API/HMI

### Designer catalog

The authoritative catalog now exposes:

- all six hulls and technology locks;
- Computer choices with availability, miniaturization level and effective PP by hull;
- Shield choices with availability, miniaturization level, effective PP and effective Space by hull;
- current automatic Drive/Fuel/Armor selections;
- server-derived production-cost ratio;
- per-hull authoritative default variants.

### Web API

`web/src/api.ts` now carries the expanded mandatory-system catalog and save payload types.

### Ship Designer HMI

The normal Designer now exposes dedicated mandatory-system controls:

- Computer selector;
- Shield selector;
- explicit no-Shield option.

The UI preview uses only server-projected effective component cost/Space tables for miniaturization. It does not reimplement research-chain miniaturization in TypeScript.

Automatic Drive/Fuel and save-time Armor behavior are explained in the HMI.

German and English labels were added for the mandatory-system controls and Computer/Shield component names.

Responsive selector layout collapses to one column on narrow screens.

## Deterministic acceptance fixtures

### Mid-Tech golden vector

Cruiser + Electronic Computer + Class I Shield:

- Electronic Computer mini level 3;
- Class I Shield mini level 2;
- effective design cost 275 PP;
- Shield Space 13;
- CP 3.

### All-Tech golden vector

Cruiser + Electronic Computer + Class I Shield:

- reduction saturated at level 5+;
- effective design cost 264 PP;
- Shield Space 5;
- CP 3.

All-Tech additionally verifies Doom Star availability with:

- Interphased Drive;
- Moleculartronic Computer default;
- Xentronium Armor;
- Class X Shield default;
- Thorium Fuel.

A deliberate All-Tech edit can downgrade to Electronic Computer and explicitly select no Shield.

## Tests added/updated

Focused Gate-3 tests cover:

- direct miniaturization golden vectors;
- Mid-Tech Titan versus Doom-Star tech lock;
- All-Tech late hull/system defaults;
- explicit Computer selection;
- explicit no-Shield selection;
- unknown/unowned Computer rejection;
- Frigate/Destroyer stationless buildability;
- Cruiser+ orbital-dock rejection;
- Star Base/Battlestation/Star Fortress acceptance;
- dynamic saved-design recompilation after research;
- stable queue-time construction snapshot after research;
- JSON persistence round-trip for in-progress military construction;
- immutable built Ship snapshot;
- expanded catalog/decision/server expectations.

Existing immediate-command/server tests were updated only where the old Frigate-only variant count was intentionally superseded by the six-hull contract.

## Verification

### Focused backend

`go test ./internal/game ./internal/core ./internal/session -run "Slice171|MilitaryDesign|ShipDesigner|ConstructionQueue|DecisionView" -count=1`

PASS.

Additional focused military/construction tests also passed during recovery.

### Full Go regression

`go test ./...`

PASS.

### Static analysis

`go vet ./...`

PASS.

### Web production build

`npm run build`

PASS, including encoding and existing New Game/race/technology/opponent contract checks, TypeScript and Vite production build.

### Fresh browser acceptance

Fresh Reference server on loopback port 7191:

`npm run check:reference-lab:browser -- http://127.0.0.1:7191`

PASS.

The browser smoke now verifies, at 390px mobile width:

- All-Tech Designer exposes Computer and Shield selectors;
- Electronic and Moleculartronic Computers are available;
- explicit no-Shield is available;
- Class X Shield is available;
- hull stepper reaches Doom Star;
- Doom Star can be saved with Electronic Computer and no Shield;
- saved design is authoritative in the returned snapshot;
- no horizontal overflow.

The same smoke then verifies the already accepted Development-Tools/BC/buyout workflow and desktop layout/ordinary-game isolation.

During diagnosis, the original pre-G3 browser smoke from repository HEAD was also run against the current backend and passed unchanged, proving the later buyout failure seen in the first combined attempt was test-order state interaction rather than a backend regression. The combined script was reordered to isolate the two workflows and then passed.

Temporary Gate-3 servers were stopped after testing.

### Diff hygiene

`git diff --check`

PASS before close preparation.

## Scope boundary

Gate 3 does not broaden the weapon catalog beyond the existing no-weapon/Laser baseline.

It does not implement broad late Computer/Armor/Shield/Drive Tactical effects.

Designs containing mandatory systems that Tactical combat does not yet support remain honestly rejected at the Tactical boundary rather than approximated.

Slice 17.2+ remain unopened.

## Gate 3 result

The frozen 17.1 implementation contract is now present end-to-end across domain authority, construction/persistence snapshots, player projection, API types, HMI and deterministic browser/backend tests.

Gate 4 QA/close is next but has not been started.


## HMI information amendment

User review of the desktop Designer identified that the first Gate-3 HMI exposed technically correct Computer/Shield selectors but did not explain what the systems actually do.

The Gate-3 implementation was therefore refined before Gate 4:

- Computer and Shield are now two independent visible cards rather than two dropdowns inside one mandatory-systems card.
- Every Computer choice shows its verified original Beam Attack effect directly: Electronic +25, Optronic +50, Positronic +75, Cybertronic +100, Moleculartronic +125.
- Every Shield choice shows its verified original per-hit damage block and capacity factor directly: Class I 1/5x, Class III 3/15x, Class V 5/25x, Class VII 7/35x, Class X 10/50x.
- No-Shield is a first-class visible choice showing 0 PP / 0 Space and no Shield protection.
- Actual miniaturized PP and Shield Space for the selected hull remain visible on every choice.
- The authoritative Designer catalog now projects the effect values so the HMI does not infer them from component names.
- The complete installed-system summary sits directly below the Ship preview and includes Drive, Computer, Armor, Shield, Fuel and installed weapons with their gameplay effects.
- Automatic systems use human-readable names rather than raw IDs.
- Drive shows strategic FTL and the verified original tactical speed contribution.
- Armor shows the verified original durability multiplier.
- Fuel shows strategic range or unlimited range.
- A visible scope note distinguishes currently active strategic behavior from verified original tactical effects that broader MOOX Tactical support does not yet simulate.
- Mobile keeps one-column cards; desktop renders Computer and Shield as two separate side-by-side cards.

Additional browser acceptance verifies no Computer/Shield dropdown remains, the two cards are separate, effect text is visible, the installed-system block is present, 390 px has no overflow, 1280 px renders two readable side-by-side cards without overflow, and the smoke is repeatable against the same reference server.

Post-amendment verification: focused Designer tests PASS; npm run build PASS; reference browser smoke PASS; go test ./... PASS; go vet ./... PASS; git diff --check PASS.

Gate 4 remains not started.


## Ship information / popup HMI amendment (2026-09-20)

The previous always-expanded Computer/Shield card layout was superseded after further desktop review. The final Gate-3 Designer UX now uses one compact Ship-information card directly below the Ship preview. It contains Hull/PP/CP/Space plus Drive, Computer, Armor, Shield, Fuel and installed weapons. Every mandatory-system row shows its effect and has a dedicated `?` detail control. Computer and Shield rows open rich popup/card selectors; Drive, Armor and Fuel remain automatic and info-only.

Computer popup cards show verified Beam Attack plus miniaturized PP/level. Shield popup cards show per-hit block, actual capacity for the current hull class, miniaturized PP/Space/level, and explicit no-Shield. Mobile uses a one-column bottom-sheet selector; desktop uses two-column popup cards. No mandatory-system dropdown is used.

Evidence: `docs/research/SLICE_17_1_GATE3_SHIP_INFO_POPUP_AMENDMENT_2026-09-20.md`.

Post-amendment verification: `npm run build` PASS; combined reference/buyout/ship-designer browser smoke PASS on fresh loopback 7192; `go test ./...` PASS; `go vet ./...` PASS; `git diff --check` PASS before staging. Gate 4 remains not started.
