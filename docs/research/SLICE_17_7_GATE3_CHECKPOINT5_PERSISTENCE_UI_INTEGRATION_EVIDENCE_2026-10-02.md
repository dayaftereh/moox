# Slice 17.7 Gate 3 - Checkpoint 5 Persistence / UI / Integration Evidence

Date: 2026-10-02
Status: COMPLETE

## Scope

Checkpoint 5 closes Slice 17.7 Gate 3 by accepting the frozen ordnance contract across persistence, replay, participant-safe projection, browser transport/presentation and strategic Design -> Build -> BattleSession integration.

Implemented / accepted here:

- active projectile live-snapshot persistence including partial interception remainder;
- deterministic restore/continuation with active projectiles;
- detached Observer/Player ordnance projection with Player RNG redaction preserved;
- corrupt projectile remainder restore rejection;
- browser DTOs for authoritative ordnance weapon/projectile state;
- additive browser Beam projectile target transport using `target_projectile_id`;
- browser `battle.fire_ordnance` launch transport;
- live server-projected Missile ammo and Torpedo cooldown presentation;
- authoritative projectile markers rendered from projected projectile x/y only;
- server-projected projectile target eligibility for manual Beam fire;
- event-driven launch/move/retarget/interception/MIRV/impact/destruction feedback;
- projectile-target Beam feedback using authoritative `target_x/target_y`;
- Selector -> Save -> Build -> BattleSession Nuclear Missile execution;
- GameSession live snapshot roundtrip and deterministic post-restore ordnance launch;
- explicit bombardment-only Bomb exclusion vector;
- permanent web contract guard for Tactical ordnance wiring.

No Slice-17.8 active Special behavior is introduced.

## 1. Active projectile persistence / replay

A dedicated CP5 BattleSession vector launches a Nuclear Missile at a sufficiently distant target so the actor remains live after the launch movement quantum.

The active actor is then given a valid partial interception remainder of 2 against unit durability 4. The live snapshot must preserve:

- projectile ID;
- current authoritative x/y;
- source/target identities;
- weapon identity/family/modifiers;
- speed and remaining movement;
- distance traveled;
- multiplicity;
- unit durability;
- `intercept_damage_remainder`;
- warhead strength;
- next projectile ID;
- live ordnance weapon state.

The test JSON-marshals the snapshot, unmarshals it, restores a BattleSession and requires the re-exported live snapshot to be byte-identical.

Both the original and restored BattleSessions then receive the same activation-ending command stream. Their continued live snapshots must again marshal byte-identically.

This proves that active actor IDs, ordnance state, event order and Tactical RNG continue deterministically after restore.

## 2. Projection safety

The active-projectile vector is projected through both Observer `View()` and participant `PlayerView()`.

Acceptance requires:

- both projections expose the same ordnance weapon/projectile facts;
- PlayerView still zeroes authoritative Tactical RNG / initial RNG;
- mutating projected projectile x/y or interception remainder does not mutate BattleSession authority;
- mutating projected ammo does not mutate BattleSession authority.

The existing CP1 detached clone path therefore remains valid for the complete CP5 actor shape.

## 3. Snapshot corruption rejection

CP5 explicitly corrupts an active projectile snapshot by setting:

`intercept_damage_remainder = unit_durability`

The frozen invariant requires remainder to stay in `[0, unit_durability)` while the actor survives.

`RestoreLiveSnapshot` rejects the corrupted snapshot.

## 4. Browser transport

`web/src/api.ts` now mirrors the accepted BattleSession wire contract for:

- `TacticalOrdnanceWeaponState`;
- `TacticalProjectileState`;
- `TacticalOrdnanceState`;
- optional `TacticalFireTarget.target_projectile_id`;
- full Tactical Beam mount/support metadata needed by presentation.

`TacticalView` carries authoritative `ordnance`.

Two additive command helpers are provided:

- `tacticalFireOrdnanceCommand` -> `battle.fire_ordnance`;
- `tacticalFireBeamProjectileCommand` -> `battle.fire_beam` with `target_projectile_id`.

No client formula determines target legality.

## 5. Browser legal-target behavior

The Tactical battlefield consumes only the server's `weapon_actions` projection.

For the selected weapon it builds separate maps of:

- projected legal Ship target IDs;
- projected legal projectile target IDs.

Ship clicks therefore dispatch:

- `battle.fire_beam` for a legal Beam Ship target;
- `battle.fire_ordnance` for a legal Missile/Torpedo Ship target.

Projectile clicks dispatch manual Beam fire only when:

- the selected action is `battle.fire_beam`; and
- the server action projection contains that exact `target_projectile_id`.

Heavy Beam exclusion remains server-authoritative because the browser cannot synthesize an omitted projectile target.

## 6. Authoritative projectile presentation

Active projectile markers are rendered directly from:

`tactical.ordnance.projectiles`

including authoritative:

- x/y;
- facing;
- source side;
- multiplicity;
- weapon identity.

The marker renderer does not integrate speed over client time and does not extrapolate past the last authoritative state.

The Tactical camera bounds include current projectile coordinates so live actors remain visible.

## 7. Event-driven presentation

The browser recognizes these authoritative event identities:

- `ordnance_launched`;
- `projectile_moved`;
- `projectile_retargeted`;
- `projectile_interception_resolved`;
- `projectile_mirv_split`;
- `projectile_impact_resolved`;
- `projectile_destroyed`.

Event feedback derives position from event payloads/current projected state/earlier authoritative projectile events. It never calculates a predicted projectile position.

For projectile-target Beam fire, existing Beam animation now accepts:

- `target_projectile_id`;
- authoritative `target_x`;
- authoritative `target_y`.

Projectile Beam damage feedback is classified separately from Ship Shield/Armor/Structure layers.

State mutation is never delayed until animation completion.

## 8. Ammo / cooldown presentation

The Tactical weapon strip now joins each active weapon slot with the server-projected `ordnance.weapon_states` row for the active Ship.

It presents:

- Missile `ammo_remaining`;
- Torpedo `cooldown_rounds`.

Server unavailable reasons are also surfaced for:

- `ammo_exhausted`;
- `weapon_cooldown`;
- `ordnance_effect_unavailable`.

No browser-side ammo decrement or cooldown countdown is performed.

## 9. Browser regression contract

A new:

`web/scripts/check-tactical-ordnance.mjs`

is included in the normal web `build` chain.

It freezes the presence of:

- projectile target DTO transport;
- ordnance projection DTO;
- `battle.fire_ordnance`;
- Beam projectile command transport;
- authoritative projectile and weapon-state reads;
- server-projected projectile target eligibility;
- all accepted ordnance event identities;
- marker/event style hooks;
- live ammo/cooldown presentation.

The production web build passes this contract, the UTF-8/mojibake guard, TypeScript build and Vite production bundle.

## 10. Selector -> Save -> Build -> BattleSession integration

A dedicated GameSession CP5 test uses the all-tech deterministic session fixture and saves a military design with:

- Nuclear Missile;
- Count 1;
- Ammo 5;
- Standard mount;
- 360-degree arc.

Acceptance follows the real pipeline:

1. save the design through the normal military-design command;
2. queue/build the design through the normal strategic session path;
3. enter the generated BattleSession;
4. verify the Tactical weapon projection:
   - family `missile`;
   - ammo 5;
   - damage 8-8;
   - support `ordnance`;
5. verify the projected weapon action is `battle.fire_ordnance`;
6. verify initial authoritative live ammo is 5;
7. marshal/unmarshal the active GameSession and require byte-identical live snapshot;
8. require restored BattleSession projection to match;
9. submit the same `battle.fire_ordnance` command to original/restored sessions;
10. require post-command Battle projections and phase to remain identical;
11. require authoritative `ordnance_launched`;
12. require projected live ammo to become 4.

This is the final Design/Selector -> persisted Ship -> Tactical execution bridge for Slice 17.7.

## 11. Frozen reference vectors A-G

Gate 3's frozen reference vectors are covered across Checkpoints 1-5:

### A - Nuclear Missile baseline

Covered by CP1/CP2 plus CP5 integration:

- designed ammo initializes live ammo;
- launch decrements one volley;
- multiplicity/count, speed 12, durability 4, DCV 50 and warhead 8 remain frozen;
- no MIRV without modifier.

### B - Zeon / Fast / Armored / ECCM

Covered by CP2/CP3 regressions:

- Fast speed 16;
- Armored durability 32;
- DCV remains 130;
- scanner reduction / ECCM defense ordering;
- one impact RNG per surviving unit.

### C - Nuclear / MIRV

Covered by CP3/CP4 regressions:

- carrier launches grouped;
- reactive defenses occur before split;
- surviving multiplicity expands x4;
- impact RNG iterates expanded units;
- intercepted carrier does not split.

### D - Plasma Torpedo variants

Covered by CP1-CP3 regressions:

- speed 24;
- durability 5;
- DCV 120;
- Torpedo cooldown 2 -> 1 -> 0;
- base / Overloaded strength;
- field dissipation;
- No Range Dissipation suppression.

### E - reactive Beam Point Defense

Covered by CP4:

- projectile DCV feeds frozen Slice-17.6 Beam resolver;
- PD slots process ascending;
- Beam events/RNG precede projectile mutation;
- returned damage mutates multiplicity/remainder;
- ordinary readiness is preserved.

### F - manual Beam projectile target

Covered by CP4 plus CP5 browser integration:

- Standard and Point Defense mounts may receive projected projectile targets;
- Heavy does not;
- normal Beam readiness is consumed;
- Beam event carries projectile identity/authoritative coordinates;
- projectile-side interception event follows.

### G - Bomb exclusion

CP5 adds a direct BattleSession regression for a Nuclear Bomb snapshot with `bombardment_only` support.

Acceptance requires:

- Bomb metadata survives Tactical spec/view;
- no ordnance weapon runtime row;
- no projectile;
- no available `battle.fire_ordnance` action;
- an explicit attempted ship-Tactical ordnance command is rejected without state/event mutation.

This closes ordinary ship-Tactical Bomb ownership truthfully without inventing colony combat.

## 12. Scope boundary

Checkpoint 5 does not activate:

- Anti-Missile Rockets;
- Lightning Field;
- Dauntless Guidance;
- Fast Missile Racks;
- fighter/bomber interception;
- any other Slice-17.8 combat Special.

The browser does not simulate any such effect.

## 13. QA

Dedicated CP5 BattleSession persistence/projection/corruption/Bomb tests:

`go test ./internal/battle -run TestSlice177Gate3Checkpoint5 -count=1`

- PASS.

All Slice-17.7 Gate-3 Battle tests:

`go test ./internal/battle -run TestSlice177Gate3 -count=1`

- PASS.

CP5 GameSession Design -> Build -> BattleSession integration:

`go test ./internal/session -run TestSlice177Gate3Checkpoint5 -count=1`

- PASS.

Web production build:

`npm run build`

- PASS, including `check:tactical-ordnance`, UTF-8 guard, TypeScript and Vite production bundle.

`git diff --check`

- PASS; only the repository's normal Windows LF/CRLF warnings are emitted.

Focused integration: `go test ./internal/ruleset ./internal/battle ./internal/game ./internal/session -count=1` - PASS. `go vet ./...` - PASS.

Repository-wide `go test ./... -count=1` - PASS on the final Checkpoint-5 candidate.

## 14. Gate 3 result

**COMPLETE / PASS - all five implementation checkpoints complete.**

Gate 4 now owns final QA + close:

- normal UI launch acceptance;
- authoritative ammo/readiness visibility;
- frozen vector verification;
- interception/non-interception QA;
- impact damage QA;
- verified Bomb context;
- animation acceptance;
- active-projectile persistence/replay;
- full Go/vet/web/browser/reference-game acceptance;
- Slice 17.7 close.
