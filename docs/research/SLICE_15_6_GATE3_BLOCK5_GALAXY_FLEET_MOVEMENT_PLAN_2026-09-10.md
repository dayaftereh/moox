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
8. A non-drag equivalent is mandatory for touch/accessibility: `Ziel wählen` enters target-pick mode, then the player taps/clicks a star.
9. Map pan/zoom must continue to work outside the fleet panel. Panel dragging/ship selection must not accidentally pan the Galaxy.
10. A legal target stages the existing `empire.move_fleet` planning order and shows an authoritative confirmation such as `System 02 · ETA 1 Runde · 2 pc / 4 pc`.
11. An illegal target must not create a draft order. The player receives a concise authoritative reason, e.g. `Ziel außer Reichweite: 5 pc, verfügbar 4 pc.` The UI must not recompute strategic distance, fuel range or legality itself.

## Exploration / player-knowledge contract

This block must close the existing strategic information leak before fleet-target UX is built on top of it. The current `buildStrategicView()` copies the complete authoritative galaxy into every player decision view, including planet/body data for systems the empire has never visited, and currently also projects foreign strategic contacts without an exploration boundary.

Use two explicit knowledge levels for the current vertical slice:

- `charted`: the star itself is known on the Galaxy map. Its stable system ID, map position, display name and visible stellar/spectral presentation may be projected, so it can be selected as a blind exploration destination.
- `explored`: at least one owned fleet has reached the system (or the empire owns a colony/outpost there). Only then may detailed system contents such as planets/bodies and their properties be projected for normal inspection.

Required persistence/authority behavior:

- Exploration knowledge belongs to the empire/server state, not React local state, so save/load/reconnect preserves it deterministically.
- Each empire's starting/home system is explored at New Game creation.
- When an owned strategic fleet arrives at a destination system, that system becomes explored for that empire before the resulting system/encounter UI is projected.
- Once explored, a system remains explored even after the fleet leaves.
- Planet potentials and other planet-derived player-safe projections must not include unexplored planets.
- Static foreign colony/outpost/fleet presence tied to an unexplored system must not be leaked through the player snapshot merely because it exists in authoritative state. Full long-range sensor/intelligence rules are outside this block; the safe baseline is to reveal system-local detail on exploration/arrival.
- Built-in AI should consume the same player-safe knowledge boundary where it uses `PlayerDecisionView`; it must not require React-only or privileged planet knowledge to function.

Galaxy UX:

- Charted but unexplored stars stay visible on the map, visually marked as `Nicht erkundet`/unknown.
- A normal click/tap on an unexplored star must not show its planet list or hidden ownership details. A compact unexplored-state dialog/message is acceptable and should offer the movement workflow when the player has a selectable fleet.
- In fleet target-pick mode, an unexplored star remains a valid candidate. Reachability/range/ETA still come from the server target projection.
- After first arrival, the same star immediately transitions to explored presentation and its authoritative planet/system details become inspectable.
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

- Far Beacon: `(100, 100)` -> exactly 5 pc from Human Home, deliberately outside the initial 4 pc range.
- Human Home: `(250, 100)`
- Darlok Contact: `(310, 100)` -> exactly 2 pc from Human Home, reachable with standard 4 pc fuel cells; Nuclear Drive speed 2 gives ETA 1 turn.

This one three-star horizontal fixture tests both success and rejection without relying on random generation:

- drop on Far Beacon -> rejected as 5 pc > 4 pc, no draft order;
- drop on Darlok Contact -> legal, ETA 1, order staged;
- Darlok combat fleet at Darlok Contact allows the arrival to enter the already accepted Tactical path.
- the fixture must guarantee the Darlok combat fleet remains at Darlok Contact until the Human arrival resolves, so the Tactical encounter cannot disappear because both fleets move past one another; use fixture/controller setup rather than production-rule exceptions.

For fast movement/Tactical regression, the QA fixture may contain one prebuilt Human Laser test ship. This fixture is not a substitute for the final Slice 15.6/15 Gate-4 canonical browser playthrough, where Ship Designer -> Construction -> fleet -> Tactical remains exercised through normal UI.

The fixture should be clearly QA-only and reproducible (for example a versioned testdata snapshot/factory plus an explicit bootstrap path), not a new normal player-facing galaxy-size option. It should be possible to recreate it with a stable game ID such as `qa-fleet-tactical-v1` without disturbing `game-1` / seed `0x8009`.

## Implementation blocks

### 5A - Exploration / player-knowledge boundary

- Add persistent per-empire explored-system knowledge and initialize each empire home system.
- Mark destinations explored on authoritative fleet arrival and preserve through save/load/reconnect.
- Sanitize player strategic Galaxy projection: charted star metadata remains; unexplored planet/body/detail data is omitted.
- Stop static system-local foreign contact leakage from unexplored systems within the current supported baseline.
- Add resolver/session projection tests proving an unexplored system cannot leak its planet details and becomes visible after arrival.

### 5B - Authority target projection

- Project legal and visible-illegal destinations with reason/distance/range/ETA data.
- Add focused resolver/session tests for 2 pc legal and 5 pc out-of-range targets.
- Preserve hidden-information boundaries and existing command validation.

### 5C - Galaxy fleet marker + floating picker

- Split fleet marker interaction from the system/star button without nested interactive elements.
- Compact draggable panel with fleet grouping and selectable ship thumbnails.
- Exact persisted ship SVGs in the picker.
- Click/tap `Ziel wählen` fallback.

### 5D - Drag/drop targeting + route feedback

- Drop hit-testing against star targets.
- Legal/illegal target states driven exclusively by server projection.
- Stage `empire.move_fleet` with selected `ship_ids`.
- Show authoritative route/ETA/range feedback and rejection messaging.
- Preserve pan/zoom and responsive behavior.

### 5E - Micro fixture + browser QA

- Add/recreate `qa-fleet-tactical-v1` beside the canonical game.
- Start with Human Home explored while Darlok Contact and Far Beacon are charted-but-unexplored; prove their planet/system detail is hidden before arrival.
- After Human arrival at Darlok Contact, prove it becomes explored and detailed system data is then visible.
- Verify fleet-marker click, ship selection, panel drag, invalid 5 pc drop, valid 2 pc drop, ETA 1 and planned order.
- Finish the turn through the visible browser and verify strategic arrival/encounter enters Tactical Combat.
- Exercise at least one Laser fire action and return to strategic view.
- Repeat the essential interaction on touch-sized viewport using the non-drag fallback and confirm the drag path remains usable.

## Acceptance for Block 5

Block 5 is complete when unexplored systems no longer leak planet/system detail, exploration knowledge persists authoritatively, and a user can remain on the Galaxy map, select an owned fleet from its marker, inspect/select its ships, attempt a charted target by drag/drop or tap target mode, receive server-derived range/ETA feedback, stage a legal movement order, reveal the destination only upon arrival, and use the deterministic micro fixture to reach the integrated Tactical battle without direct API/dev-tool milestone actions.

## Non-goals

- Do not rebalance or rescale the normal seed `0x8009` galaxy in this block.
- Do not hide charted star positions entirely; unexplored stars must remain available as blind movement/exploration targets.
- Do not implement broad sensor/spy/intelligence fog-of-war rules beyond the system-exploration knowledge boundary required here.
- Do not make the QA micro fixture a normal game mode.
- Do not duplicate strategic distance, fuel, ETA or legality rules in React.
- Do not require drag as the only movement input.
- Do not broaden into full arbitrary multi-fleet command orchestration beyond what is needed for the accepted vertical slice.
