# Slice 15.5 Gate 3 Block 3 - Interactive Tactical battlefield UI

Date: 2026-09-10

## Goal

Replace the Slice-15.4 Tactical placeholder with the first real browser-playable server-authoritative Tactical battlefield. The browser must render only participant-safe Tactical projection and submit only server command intent; movement legality, facing, range, RNG, hit/damage and round progression remain Go-owned.

## Battlefield shell

`BattleRouteView` still preserves the strategic Encounter entry/return contract. `Taktischen Kampf betreten` now opens `TacticalBattlefield` when a live supported Tactical view is projected.

The battlefield is a scalable SVG world-coordinate plane:

- current ship X/Y come directly from `tactical.ships`;
- ship rotation is `facing * 22.5°` from authoritative 16-way Facing;
- server-projected `legal_moves` render as selectable movement points;
- server-projected `legal_fire_actions` render weapon/target actions;
- no global battle rectangle or gameplay boundary is drawn;
- the visible grid is regenerated for the current camera ViewBox and therefore visually continues while panning.

The current ship glyph is a compact semantic SVG fallback. Integration of the richer accepted 15.3 procedural ship visuals remains a later Gate-3 visual-polish block; gameplay does not depend on the fallback art.

## Command lifecycle

`App.tsx` now owns Tactical mutation orchestration just like strategic mutations:

1. TacticalBattlefield builds a typed intent command from the latest projected Tactical sequence;
2. App submits it through `submitBattleCommand`;
3. App reloads the participant snapshot after success;
4. structured stale/conflict failures reuse the Slice-15.4 `refreshAfterConflict` path;
5. while the App lifecycle is not synchronized, Tactical commands are disabled.

The Battlefield never optimistically mutates X/Y, Facing, damage, readiness or active ship.

## Move mode

Move mode uses only `legal_moves[]` from the server.

- each projected destination has a visible point plus a larger transparent hit/touch target;
- click/tap/keyboard activation submits `battle.move_ship` with ship ID and destination X/Y;
- no client-side cost or collision calculation exists;
- active ship card shows current Movement, Facing and Position from the refreshed server snapshot.

## Fire mode

Fire mode uses only `legal_fire_actions[]`.

- supported ready Laser actions list authoritative target counts;
- each target displays the projected range index;
- legal enemy ships receive a target ring;
- click/tap target submits the existing `battle.fire_beam` command;
- after firing, the refreshed server projection removes a spent weapon action automatically.

## Scan mode

Scan is a non-mutating presentation mode matching the frozen Gate-2 contract.

Click/tap any participant-visible ship to inspect:

- owner / hull / drive;
- X/Y and Facing;
- Movement current/max;
- Beam Offense / Beam Defense;
- Armor current/max with health bar;
- Structure current/max with health bar;
- supported weapon name/count/damage range/readiness.

No composite strength number and no fake internal-system state are generated. Scan selection itself submits no Battle command.

## Activation / round flow

`End Activation` is enabled only when the participant owns the projected active ship and `can_end_activation` is true. It submits `battle.end_activation` through the same authoritative command lifecycle.

When another seat is active, move/fire actions are disabled while Scan remains available. Built-in AI can therefore act through its normal authoritative path while the human view remains read-only for commands.

## Camera / effectively unbounded field

Camera state is local presentation only:

- desktop: wheel zoom, drag pan;
- touch: one-finger pan, two-finger pinch zoom;
- explicit Zoom In / Zoom Out / Recenter controls are available as accessible fallbacks;
- camera updates only the SVG ViewBox;
- no camera state is transmitted as game input;
- there is no visible arena edge.

## Responsive UI

Desktop uses a wide battlefield with command/scan panel beside it.

Below 900 px the command area stacks beneath the battlefield. Below 700 px:

- mode controls become a three-column touch row;
- command panels become one column;
- primary Tactical controls inherit the project-wide 44-px mobile touch target floor;
- the viewport remains the dominant interaction surface.

All new Tactical labels are available in EN/DE through the shared i18n layer.

## Managed-browser end-to-end QA

QA used an isolated real GameSession on port 7181, imported from the current clean review seed and modified only as a test fixture:

- both two-Scout combat fleets placed in the same neutral System 02;
- both empires set to War;
- all four fixture Scouts given exactly one supported standard Laser;
- Seat1 remained local human and Seat2 builtin AI.

No browser mock was used.

### Entry

From Turn 1 / Planning, one normal `Fertig` click created a real 2-vs-2 Encounter and auto-routed to `/battle/1`. The Battle Entry showed two combat ships per side and `Taktischen Kampf betreten`.

Opening Tactical rendered:

- 4 authoritative ship markers;
- active Human Scout at `(10,9)` Facing 0;
- 515 server-projected legal move destinations for 20 movement points;
- Movement / Fire / Scan / End Activation controls.

### Move

The first projected move was selected through the SVG UI:

- `(10,9) -> (11,9)`;
- Movement `20 -> 19`;
- Facing remained 0;
- Tactical command sequence `1 -> 2`;
- projected legal move count refreshed from 515 to 443.

### Scan

Scan selected enemy ship 60 at `(14,9)` Facing 8 and displayed:

- Movement 20/20;
- Beam Offense 25;
- Beam Defense 0;
- Armor 4/4;
- Structure 4/4;
- Laser Cannon x1, damage 1-4, Ready.

Scan did not advance the Tactical command sequence.

### Fire / damage refresh

Fire mode exposed two server-projected legal targets:

- ship 60 at range index 1;
- ship 61 at range index 2.

Firing at ship 60 advanced sequence `2 -> 3`. The subsequent Scan showed authoritative damage:

- Armor `4/4 -> 0/4`;
- Structure remained `4/4`;
- the Human Laser action disappeared because readiness was spent.

### Activation / Built-in AI / round boundary

End Activation advanced sequence `3 -> 4` to the second Human Scout. Ending that activation allowed Built-in AI to perform its Tactical activations through the server. The browser returned to Human control at:

- Round 2;
- command sequence 9;
- first Human Scout active again;
- Movement reset to 20/20.

This demonstrates a real 2-vs-2 multi-ship round transition through the browser.

### Camera

Desktop zoom changed only the SVG ViewBox and Recenter restored the presentation. Emulated pointer QA then verified:

- one-finger pan changed ViewBox origin;
- two-finger pinch changed ViewBox size from 38x38 to 25.33x25.33;
- Tactical command sequence stayed at 9 throughout camera-only interaction.

### 320 CSS-px

Using the project mobile-media-rule QA technique at exactly 320 CSS px:

- Tactical root outer width: 320 px;
- client/scroll width: 318/318 px;
- no non-SVG horizontal overflow;
- viewport: 300 x 339 px;
- mode buttons: 44 px high;
- End Activation: 44 px high;
- Back to Encounter: 44 px high.

## Validation

Passed after implementation:

- `go test ./...`;
- `go vet ./...`;
- `npm run build`;
- `git diff --check`.

The temporary 7181 server, browser profile and QA snapshot files were removed/stopped after evidence collection.

## Next

Gate-3 gameplay UI is now present. Remaining work before Gate-4 close is primarily:

- integrate accepted 15.3 ship/battle visual primitives with safe fallback;
- add/strengthen automated browser regression where useful;
- final battle-resolution/return QA through a full browser battle;
- independent Gate-4 desktop/mobile close pass.
