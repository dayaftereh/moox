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

### 5C - Galaxy fleet marker + floating picker

- Split fleet marker interaction from the system/star button without nested interactive elements.
- Compact draggable panel with fleet grouping and selectable ship thumbnails.
- Exact persisted ship SVGs in the picker.
- Click/tap `Ziel wÃƒÂ¤hlen` fallback.

### 5D - Drag/drop targeting + route feedback

- Drop hit-testing against star targets.
- Legal/illegal target states driven exclusively by server projection.
- Stage `empire.move_fleet` with selected `ship_ids`.
- Show authoritative route/ETA/range feedback and rejection messaging.
- Preserve pan/zoom and responsive behavior.

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
