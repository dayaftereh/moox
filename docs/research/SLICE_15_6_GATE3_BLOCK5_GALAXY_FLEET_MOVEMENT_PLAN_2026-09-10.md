# Slice 15.6 Gate 3 Block 5 - Galaxy fleet direct manipulation + micro tactical fixture

Status: planned from user feedback; implementation not started.

## Why this block exists

The current Galaxy map shows fleet presence but does not let the player interact with the fleet marker itself. At the Human home system the two visible markers are:

- colony marker: the owned colony at the system;
- own-fleet marker: aggregate count of owned ships at the system (currently 3 = 2 combat ships + 1 colony ship).

The own-fleet marker is currently `aria-hidden` inside the star/system button, so it cannot be clicked independently. The separate Fleets screen already consumes server-projected `fleet_moves` and can submit `empire.move_fleet`, but the current canonical `game-1` has no legal move choice because the starting 4 pc fuel range cannot reach the nearest generated system (8 pc in seed `0x8009`).

This blocks the normal Galaxy -> fleet movement -> encounter -> Tactical Combat browser journey required by Slice 15.6 Gate 3.

## User-facing interaction contract

1. The fleet marker on the Galaxy map becomes a real independent control. Clicking/tapping the star still opens the system. Clicking/tapping the fleet marker opens the fleet picker.
2. The fleet picker is a compact floating dialog/palette, not a full-map modal. It should remain small enough that the Galaxy and target stars stay visible.
3. If several friendly fleets are present in one system, the panel groups ships by source fleet. Movement is planned for one source fleet at a time.
4. Every concrete ship is shown with its persisted `ProceduralShipGlyph`, name/design revision and a compact role/weapon cue. Special civilian ships use their appropriate existing icon/art fallback.
5. Ships are individually selectable. Default may select all ships in the active source fleet. Subset movement continues to use authoritative `ship_ids` on `empire.move_fleet`.
6. The panel has a drag handle/header. Dragging the panel across the map and releasing it on a star attempts to use that star as the movement target. Releasing over empty map space only repositions the panel.
7. While dragging, candidate destination stars remain visible and the current hovered/drop target is highlighted. A route/target preview may be drawn without taking over the map.
8. A non-drag equivalent is mandatory for touch/accessibility: `Ziel wÃƒÂ¤hlen` enters target-pick mode, then the player taps/clicks a star.
9. Map pan/zoom must continue to work outside the fleet panel. Panel dragging/ship selection must not accidentally pan the Galaxy.
10. A legal target stages the existing `empire.move_fleet` planning order and shows an authoritative confirmation such as `System 02 Ã‚Â· ETA 1 Runde Ã‚Â· 2 pc / 4 pc`.
11. An illegal target must not create a draft order. The player receives a concise authoritative reason, e.g. `Ziel auÃƒÅ¸er Reichweite: 5 pc, verfÃƒÂ¼gbar 4 pc.` The UI must not recompute strategic distance, fuel range or legality itself.

## Visited-system / player-knowledge contract

This block must close the existing strategic information leak before fleet-target UX is built on top of it. The current `buildStrategicView()` copies the complete authoritative galaxy into every player decision view, including true system names and planet/body data for systems the empire has never visited. `PublicEmpires` and `Diplomacy` likewise currently expose every empire from game start.

Use `visited` as the canonical system-knowledge term for this vertical slice.

### Unvisited star

An unvisited star remains present on the Galaxy map so the player can navigate and send a fleet toward it, but the player-safe projection may reveal only information visible from the galactic map:

- stable opaque system ID required by commands;
- map position;
- visible stellar/spectral color/art.

It must NOT reveal before visit:

- the true system name (including in `title`, `aria-label`, hidden DOM text or client data);
- planet/body list or properties;
- colony/outpost ownership tied to that system;
- system specials, monsters/guardians or special-system identity such as Orion;
- other system-local details that would identify the destination in advance.

The UI labels such a target generically, e.g. `Unbekannter Stern` / `Nicht besucht`. The home system is visited from game creation and therefore named immediately.

### Visited system

A system becomes visited for an empire when an owned strategic fleet reaches it, or when the empire already owns a colony/outpost there. The server records this persistently before projecting the resulting system/encounter state. Once visited, the empire keeps the system knowledge across departure, save/load and reconnect.

After visit, the player-safe projection may reveal the true system name, planets/bodies and the supported local details. Planet potentials and any other planet-derived projection must be generated only for visited systems.

### Race discovery / first contact

System knowledge and diplomatic contact are separate concepts.

- Do not expose every race identity through `PublicEmpires` or every diplomacy row at game start.
- Before first contact, another empire's name/race/diplomatic state must not be available to the player UI merely because that empire exists in authoritative state.
- When the supported first-contact rule becomes true, the server records/projects that the race is known and the Diplomacy UI becomes available for that race.
- Visiting a system that actually reveals foreign presence can therefore lead into first contact, but merely sharing historical visit data for an otherwise empty system must not magically reveal a race.
- For MOO2 fidelity, the first-contact rule should be based on authoritative inter-empire reachability/encounter state rather than React. Original MOO2 documentation describes diplomacy beginning when either empire's ships are capable of reaching one of the other's colonies. Implement the narrow server-side condition required by the current two-empire vertical slice and cover it with focused tests; broader contact-loss/sensor/intelligence behavior can be deferred.
- Knowledge of an empire identity after first contact should not be erased merely by leaving a system; any later distinction between `known race` and currently active diplomatic contact is a separate expansion.

### Galaxy UX

- Unvisited stars remain visible and targetable, but have no true name label.
- Clicking/tapping an unvisited star must not open the normal detailed System dialog. It may open a compact `Nicht besucht` prompt and offer fleet target selection.
- In fleet target-pick mode an unvisited star remains a candidate. Reachability/range/ETA still come exclusively from the server target projection.
- Target/rejection copy must not accidentally reveal the hidden system name: use `Unbekannter Stern Â· 5 pc Â· Reichweite 4 pc`, for example.
- After first arrival, the star gains its real name on the map and the detailed System dialog becomes available immediately.
## Authority / projection change

Current `fleet_moves` contains legal choices only. That is insufficient for drop-on-invalid-target UX because React cannot explain why a visible star is illegal without duplicating simulation rules.

Add a player-safe authoritative target projection for each movable source fleet (name can be finalized during implementation), carrying at least:

- `fleet_id`;
- `destination_system_id`;
- `legal`;
- stable `reason` code when illegal;
- `distance_parsecs`;
- `fuel_range_parsecs`;
- `supply_distance_parsecs` when applicable;
- `eta_turns` when meaningful;
- supported `ship_ids`/subset context where required.

Initial reason vocabulary should cover at least `reachable`, `same_system`, `out_of_range`, `in_transit`, `unsupported_drive` and `no_supply`. Server validation of `empire.move_fleet` remains final authority.

The existing legal `fleet_moves` contract may either remain as the command-ready subset or be generalized carefully. Do not move formulas/legality into React.

## Deterministic micro tactical fixture

Do not change the normal Small-Galaxy generator and do not change the canonical seed `0x8009` to make this test pass. Add a separate reproducible QA fixture/game that can be recreated beside the normal game.

Proposed fixed layout (30 coordinate units = 1 pc):

- Darlok Home / Far Beacon: `(100, 100)` -> exactly 5 pc from Human Home, deliberately outside the initial 4 pc range; its true identity/name is hidden while unvisited.
- Human Home: `(250, 100)`
- Contact Star: `(310, 100)` -> exactly 2 pc from Human Home, reachable with standard 4 pc fuel cells; Nuclear Drive speed 2 gives ETA 1 turn. The fixture places a Darlok combat fleet here without exposing Darlok identity to Human before encounter.

This one three-star horizontal fixture tests both success and rejection without relying on random generation:

- drop on Darlok Home / Far Beacon -> rejected as 5 pc > 4 pc, no draft order and no hidden true name/race leak;
- drop on Contact Star -> legal, ETA 1, order staged while the target is still shown generically as unvisited;
- Darlok combat fleet at Contact Star allows Human arrival to establish the supported first-contact/known-race state and enter the already accepted Tactical path.
- the fixture must guarantee the Darlok combat fleet remains at Contact Star until the Human arrival resolves, so first contact/Tactical cannot disappear because both fleets move past one another; use fixture/controller setup rather than production-rule exceptions.

For fast movement/Tactical regression, the QA fixture may contain one prebuilt Human Laser test ship. This fixture is not a substitute for the final Slice 15.6/15 Gate-4 canonical browser playthrough, where Ship Designer -> Construction -> fleet -> Tactical remains exercised through normal UI.

The fixture should be clearly QA-only and reproducible (for example a versioned testdata snapshot/factory plus an explicit bootstrap path), not a new normal player-facing galaxy-size option. It should be possible to recreate it with a stable game ID such as `qa-fleet-tactical-v1` without disturbing `game-1` / seed `0x8009`.

## Implementation blocks

### 5A - Visited-system / player-knowledge boundary

- Add persistent per-empire visited-system knowledge and initialize each empire home system as visited.
- Mark destinations visited on authoritative fleet arrival and preserve through save/load/reconnect.
- Sanitize player strategic Galaxy projection: unvisited stars expose only opaque ID + position + stellar appearance; true name and planet/body/detail data are omitted until visited.
- Stop foreign race/diplomacy and static system-local contact leakage before authoritative first contact; filter `PublicEmpires`, `Diplomacy` and strategic contacts accordingly.
- Add resolver/session projection tests proving an unvisited system cannot leak its true name, Orion/special identity or planet details; prove it becomes named/inspectable after arrival; prove unknown empires stay out of diplomacy until first contact.

Implementation status (2026-09-10): **5A COMPLETE**

- `0f6c262` persists sorted per-empire `visited_system_ids`, initializes home-system knowledge and records fleet-arrival visits; unvisited player projection strips true names, planets/bodies and static local contacts.
- `0588a9f` adds sorted persistent `known_empire_ids`, symmetric first contact on authoritative arrival/shared-visit presence, pre-contact filtering for `PublicEmpires`, DecisionView diplomacy/commands and strategic contacts, plus anonymous Galaxy presentation and the compact `Nicht besucht` dialog. Built-in AI now explores only through server-projected legal fleet moves and the legacy full-conquest regressions explicitly preserve their prior post-contact/full-map fixture premise.
- `850350e` closes the older `PlayerView` leak as well: pre-contact foreign seat identities and diplomacy rows are omitted.
- Regression: `go test ./...` green; web `npm run build` green (main JS 258.20 kB, gzip 67.11 kB); `git diff --check` green before commits.
- Visible browser QA on fresh canonical seed `0x8009`: exactly one named system (Human home `System 01`) and 19 anonymous stars; selecting an anonymous star opens only `Nicht besucht / Unbekannter Stern` with no planet dialog; Diplomacy shows `Keine diplomatischen Kontakte.`
- Player-snapshot evidence on the same game: `public_empires` contains Human only; `view.seats` contains Human only; both DecisionView and PlayerView diplomacy properties are omitted pre-contact; `visited_system_ids=[4]`.
- Actual post-arrival visit + symmetric first-contact persistence is covered by `TestFleetArrivalEstablishesFirstContactAtPreviouslyVisitedSystem`; visible arrival/reveal QA remains naturally coupled to the upcoming fleet-movement blocks.

### 5B - Authority target projection

- Project legal and visible-illegal destinations with reason/distance/range/ETA data.
- Add focused resolver/session tests for 2 pc legal and 5 pc out-of-range targets.
- Preserve hidden-information boundaries and existing command validation.

Implementation status (2026-09-11): **5B COMPLETE**

- `09a5b03` adds server-authoritative `FleetMoveTarget` projection for every visible non-source star and every supported whole/subset fleet movement profile.
- Projected data is explicit and client-ready: `legal`, stable `reason` (`no_supply`, `out_of_fuel_range`, `invalid_eta`), direct `distance_parsecs`, `eta`, `fuel_range_parsecs`, `supply_distance_parsecs`, plus optional `ship_ids`.
- Existing legal `fleet_moves` are now derived from the same target authority, so legal buttons/AI and future invalid-target feedback cannot drift into separate range or ETA formulas.
- Focused 3-star fixture proves: 2 pc target is legal at 4 pc range with ETA 1; 5 pc target remains visible but illegal with `out_of_fuel_range`, 4 pc range and ETA 3; the actual `empire.move_fleet` validator rejects that same 5 pc command.
- `PlayerDecisionView.decisions.fleet_move_targets` projects the authority to clients; the web wire types mirror it without adding any client-side legality calculation.
- Session coverage proves a target can point at an unvisited/anonymous star while the StrategicView still omits its true name and detail data.
- Validation: focused resolver/session tests green; full `go test ./...` green; web `npm run build` green (main JS 258.20 kB, gzip 67.11 kB); `git diff --check` green.
- Canonical 7171 is intentionally not restarted for this backend-only block; deploy the new target wire together with the first consuming Galaxy fleet UI in 5C.
### 5C - Galaxy fleet marker + floating picker

- Split fleet marker interaction from the system/star button without nested interactive elements.
- Compact draggable panel with fleet grouping and selectable ship thumbnails.
- Exact persisted ship SVGs in the picker.
- Click/tap `Ziel wÃƒÂ¤hlen` fallback.

Implementation status (2026-09-11): **5C COMPLETE**

- `3338ead` splits each Galaxy system into sibling controls: the star button remains responsible for system inspection while the own-fleet marker is now an independent accessible button; browser DOM QA confirms zero nested buttons.
- Clicking the fleet marker opens a compact floating `Flottenauswahl` instead of a map-covering modal. The panel is draggable by its header, viewport-clamped and keeps the Galaxy visible behind it.
- Fleets at the same system are grouped independently. Canonical seed `0x8009` visibly exposes combat fleet `#59` and civilian colony fleet `#63` from the Human home system.
- Concrete ships render their exact persisted `ProceduralShipGlyph`; visible QA shows `Scout 1` and `Scout 2` separately selectable with independent `aria-pressed` state. Whole-fleet and single-ship selection profiles consume only server-projected `fleet_move_targets`.
- Fixed support fleets do not invent fake ship instances: `#63` renders `Colony Ship` as one fixed support unit.
- The footer shows only authoritative target counts. Canonical `0x8009` correctly reports `0 von 19 Zielen erreichbar`; React performs no distance/range/ETA legality calculation.
- Visible browser QA: fleet marker opens picker without opening SystemDialog; closing picker and clicking `System 01` still opens the normal system dialog; marker is outside any star button, desktop hit area is 24x24 px (30x30 px mobile CSS), picker measured about 380x264 px in a 791x605 viewport, and drag QA moved it from roughly `(78,60)` to `(178,140)` while remaining clamped.
- Validation: web `npm run build` green (main JS 266.43 kB / 69.07 kB gzip, below 350 KiB guardrail); full `go test ./...` green; `git diff --check` green.
- Canonical 7171 was refreshed for integrated 5B+5C QA and now runs as service `job-002278`. The known persistence seam reproduced on restart; `game-1` was recreated through the visible New Game UI with unchanged seed `0x8009`.
5C UX refinement (2026-09-11, `3503a6b`) supersedes the first picker layout:

- Combat/civilian fleet tabs are removed from the Galaxy picker. Every own ship/unit currently at the system is presented in one shared selectable pool.
- The pool is a compact four-column image grid. Canonical `game-1` visibly shows `Scout 1`, `Scout 2`, and `Kolonieschiff` side-by-side in the first row; larger sets scroll vertically inside the picker instead of growing the dialog. Ephemeral DOM layout QA with 32 tiles preserved four columns and produced body overflow/scroll.
- Concrete military ships keep their persisted `visual_genome` / `ProceduralShipGlyph`. Special strategic ships had no persisted per-instance visual genome in the current model, so `SpecialShipGlyph.tsx` now provides stable committed vector silhouettes for Colony Ship, Outpost Ship, and Troop Transport instead of the former flag/outpost/transport icon. A future per-instance special visual genome can replace this without changing the picker contract.
- Selection can span internal source fleets. Reachable/total counts are computed only by matching/intersecting the already authoritative per-source `fleet_move_targets`; no range, ETA, fuel, or legality formula moved into React. Unsupported partial multi-ship subsets remain explicitly unsupported rather than guessed.
- Own colony/outpost badges above the star are removed from this surface. A visited system name turns green when it contains an own colony; normal visited names remain white.
- CORRECTED by `7131046`: Galaxy fleet presence is aggregated **per empire/player, not per internal StrategicFleet**. Any number or mix of own Combat/Scout/Colony/Outpost/Transport fleets at one system produces exactly one own fleet icon; each visible foreign empire with fleet presence produces exactly one additional icon. Empire icons occupy distinct clockwise slots around the star (own presence first, then visible foreign empire IDs in stable order). The combined own system pool still retains source-fleet IDs internally for authoritative movement commands.
- `Empire.PlayerColorSlot` (1..8; 0 only as legacy/unset state) is now persisted at New Game creation and projected player-safely. Old saves without the field receive a deterministic projection fallback by authoritative empire order. Fleet icons use that player color; they are no longer colored by diplomacy stance.
- The picker is materially smaller: desktop max width 292 px; mobile max width 276 px and 76vw, leaving map space visible beside it. Canonical browser QA measured 292x201 px with a 4x66.5 px grid in a 791x605 viewport.
- Canonical browser QA also proved: no own colony marker remains, `System 01` computes to the success/green color, Scout 1 can be deselected while Scout 2 + Colony Ship stay selected (`2/3`, `Ziele 0/19`), Colony Ship renders a dedicated SVG rather than a `GameIcon`, and normal star inspection still opens only the SystemDialog after the picker closes.
- Validation after refinement: web `npm run build` green (main JS 268.83 kB / 69.90 kB gzip), full `go test ./...` green, staged `git diff --check` green.
OPEN FOLLOW-UP - persisted visual genomes for strategic special ships

- Colony Ship, Outpost Ship, and Troop Transport currently exist as fixed `StrategicFleet.SpecialKind` units without a concrete `Ship` instance and therefore without a persisted per-instance `visual_genome`.
- The current `SpecialShipGlyph.tsx` silhouettes are an intentional interim presentation layer only. They must not become the final persistence model.
- Required future contract: every constructed/starting strategic special ship receives a stable persisted visual-genome instance at creation time; save/load/reconnect must reproduce exactly the same geometry; projections expose that persisted genome; Galaxy/System/Construction views render the same instance everywhere.
- Migration/backfill must be deterministic for existing saves that predate the field. No client-side random regeneration on render.
- Acceptance requires focused tests for creation, snapshot projection, save/load round-trip, reconnect stability, and visual identity consistency across construction queue -> completed strategic unit -> Galaxy/System picker.
- Keep this follow-up open until the authoritative core/persistence representation is implemented; the committed static special silhouettes are not sufficient to close it.
### 5D - Drag/drop targeting + route feedback

- Drop hit-testing against star targets.
- Legal/illegal target states driven exclusively by server projection.
- Stage `empire.move_fleet` with selected `ship_ids`.
- Show authoritative route/ETA/range feedback and rejection messaging.
- Preserve pan/zoom and responsive behavior.

5D implementation status (2026-09-11): **COMPLETE** (`ee9f60e`)

- The compact shared system-unit picker now has a real `Ziel wählen` / `Choose target` mode. It remains available even when the canonical start has `0/19` legal destinations so the player can inspect why a destination is rejected.
- Candidate state is derived only from the already authoritative `fleet_move_targets`. Common destinations are intersected across every selected source-fleet profile; React does not calculate strategic distance, supply, fuel range, ETA, or legality.
- Legal common destinations receive a green target ring; server-rejected destinations receive a red dashed target ring. The source system is not projected as a movement target.
- Clicking/tapping a projected destination in target mode is intercepted before normal SystemDialog/unknown-star inspection. Hidden targets therefore remain `Unbekannter Stern` / `Unknown star`; no true system name is introduced through title, feedback, or target state.
- Illegal targets stage **no** draft order. Canonical seed `0x8009` visible QA selected the first anonymous target and showed `Unbekannter Stern · Außer Reichweite`, `8 pc · Reichweite 4 pc · ETA 4`, with no SystemDialog/unvisited dialog and no planning draft indicator.
- Legal target wiring stages one existing `empire.move_fleet` draft per selected authoritative source profile using the exact projected `target.fleet_id`, `target.destination_system_id`, and projected optional `target.ship_ids`. Existing draft keys remain `fleet:<fleet_id>`, so later selection changes replace the source fleet's planned move rather than duplicating it.
- Drag/drop uses the same `planFleetDestination` path as tap/click. Dragging the compact picker header more than 6 px temporarily exposes target highlighting; dropping over a star resolves its `data-galaxy-system-id` with `elementsFromPoint`. Dropping on empty Galaxy space only repositions the picker and clears stale feedback.
- Canonical visible drag QA reproduced the same 8 pc / range 4 / ETA 4 rejection with target mode off; empty-space drop moved the picker without target feedback or an order.
- Picker remains compact after feedback (canonical measured about 292x250 px) and still has zero nested buttons.
- Validation: web `npm run build` green (main JS 272.77 kB / 71.15 kB gzip, below the 350 KiB guardrail); focused authoritative fleet-target tests green; full `go test ./...` green; `git diff --check` green.
- The canonical `0x8009` start intentionally has no legal destination, so **positive visible green-target -> staged order -> turn resolution -> actual arrival acceptance remains the first 5E micro-fixture action**, not a reason to alter normal galaxy generation/range.
### 5E - Micro fixture + browser QA

- Add/recreate `qa-fleet-tactical-v1` beside the canonical game.
- Start with Human Home visited while Contact Star and Darlok Home/Far Beacon are visible-but-unvisited; prove their true names, Darlok identity and planet/system detail are hidden before first contact/visit.
- After Human arrival at Contact Star, prove it becomes visited, gains its real name/details, Darlok becomes a known race through the supported encounter/first-contact path, and Tactical opens.
- Verify fleet-marker click, ship selection, panel drag, invalid 5 pc drop, valid 2 pc drop, ETA 1, planned order, no pre-contact Darlok identity leak and post-encounter Diplomacy availability.
- Finish the turn through the visible browser and verify strategic arrival/encounter enters Tactical Combat.
- Exercise at least one Laser fire action and return to strategic view.
- Repeat the essential interaction on touch-sized viewport using the non-drag fallback and confirm the drag path remains usable.

## Acceptance for Block 5

Block 5 is complete when unvisited systems no longer leak their true name/special identity/planet detail, visited-system knowledge persists authoritatively, unknown races no longer leak through diplomacy before first contact, and a user can remain on the Galaxy map, select an owned fleet from its marker, inspect/select its ships, attempt an unvisited target by drag/drop or tap target mode, receive server-derived range/ETA feedback, stage a legal movement order, reveal/name the destination only upon arrival, establish supported first contact when appropriate, and use the deterministic micro fixture to reach the integrated Tactical battle without direct API/dev-tool milestone actions.

## Non-goals

- Do not rebalance or rescale the normal seed `0x8009` galaxy in this block.
- Do not hide unvisited star positions or stellar colors; unvisited stars must remain available as blind movement/visit targets.
- Do not implement broad sensor/spy/intelligence fog-of-war or full MOO2 contact-loss rules beyond the visited-system + initial-contact boundary required here.
- Do not make the QA micro fixture a normal game mode.
- Do not duplicate strategic distance, fuel, ETA or legality rules in React.
- Do not require drag as the only movement input.
- Do not broaden into full arbitrary multi-fleet command orchestration beyond what is needed for the accepted vertical slice.
