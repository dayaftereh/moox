# Slice 15.6 Gate 3 - Weapon Slots, Quantity and Tactical Selection Plan

Date: 2026-09-14
Status: implemented and validated
Baseline before this planning document: `12e7cbc` (`fix: defer tactical weapon selector`)

## Why this block exists

The Tactical UI is now clean enough that the remaining weapon interaction gap is explicit: an active ship must be able to choose which installed weapon group fires. The current one-line `Laser Cannon` strip was intentionally removed because it represented only the old single-weapon baseline and did not model the intended Master of Orion-style slot behavior.

The user approved completing this as a narrow vertical slice through Ship Designer -> persisted Ship Design -> built Ship -> Tactical projection -> per-slot fire selection, rather than postponing the entire interaction until the broad Slice 17 Ship Designer work.

This is an explicit amendment to the Slice 15.6 Gate-2 freeze. The old freeze allowed only `none | slot 0 Laser Cannon count 1` and reserved multiple live mounts/count > 1 for later. We are now deliberately lifting only that weapon-mount restriction while keeping the rest of the broader Ship Designer scope deferred.

## Product semantics agreed with the user

A weapon **slot is a firing group**. `count` is the number of physical weapons grouped into that one slot.

Examples:

- `Slot 0: Laser Cannon x5` means five Laser Cannons installed in one group. Selecting that slot in Tactical fires the five weapons as one grouped volley. The player cannot split those five weapons across different targets during that activation.
- Five separate mounts, `Slot 0..4: Laser Cannon x1`, mean five independently selectable firing groups. The player may fire the slots individually and may therefore use different targets if legal.
- In the Ship Designer, pressing `+` on a mount increases the **count in the same slot**. It does not create a new slot.
- Choosing/addition of the weapon again creates a **new slot** with count 1.
- `-` decreases the count in the current slot; when count would reach zero the row should be removed rather than persisting an invalid zero-count mount.
- Hull space and production/design cost are consumed per physical weapon, therefore weapon contribution is multiplied by `count`.
- Slots retain stable identity/order. The current core contract already supports slots 0..7, so this work keeps the existing maximum of eight weapon mounts unless a later design decision changes it.

This distinction is intentionally future-facing for missiles, ammunition, limited-shot weapons, special weapons and other systems where grouping versus independent slots matters tactically.

## Tactical firing semantics

For the first implementation, a grouped mount with `count = N` represents N physical weapons firing in the same selected volley.

Agreed direction:

- selecting a slot selects the whole mount/group;
- firing consumes readiness for that **slot**, not for every mount of the same weapon type;
- all weapons in the selected slot fire at the selected target;
- each physical weapon receives its own hit/damage resolution rather than collapsing `count` into one oversized hit;
- the event/result representation should retain enough information to explain the volley (shots/hits/damage) without breaking deterministic replay;
- after firing, that slot is no longer ready for the current activation;
- other ready slots remain independently selectable and fireable until activation rules end the ship's turn.

This preserves the important gameplay difference between `1 slot x5` and `5 slots x1`.

## Repo facts verified before implementation

The architecture is already prepared for most of this model:

### Core ship design

`internal/core/military.go` already defines `ShipWeaponMount` with:

- `Slot int`
- `WeaponID string`
- `Count int`

`ShipDesignSpec.Weapons` already persists `[]ShipWeaponMount`.

Core validation already permits up to 8 mounts, requires slots in `0..7`, strictly ascending slot order, non-empty weapon id and positive count.

### Web/API contract

`web/src/api.ts` already exposes `ShipWeaponMount[]` in both the design spec and save payload. No new conceptual API shape is required merely to represent multiple slots/counts.

### Tactical protocol/model

`internal/battle/tactical.go` already has:

- `TacticalWeaponSpec { Slot, WeaponID, Count, MinDamage, MaxDamage }`
- per-slot `TacticalWeaponState { Slot, Ready }`
- `TacticalWeaponView` including `Slot`, `WeaponID`, `Count` and `Ready`
- `TacticalFireAction` addressed by `weapon_slot`
- `FireBeamPayload` addressed by `weapon_slot`

The web command `tacticalFireBeamCommand(...)` already sends `weapon_slot`, and the Tactical UI already carries `selectedWeaponSlot` state internally.

Therefore the work is an expansion of existing authority, not a replacement data model.

## Current deliberate restrictions that must be lifted

The following guards are the actual blockers and must be changed deliberately with regression tests:

1. `internal/game/military_design.go`
   - `validateBaselineMilitaryWeapons` currently allows zero weapons or exactly one mount only.
   - It further requires exactly slot 0 / `laser_cannon` / count 1.
   - `clearedMilitaryDesignSpec` currently charges weapon space/cost only for the single supported weapon; count multiplication and multiple mounts are not yet applied.

2. `web/src/ShipBuilderView.tsx`
   - current UI state is just a `laserInstalled` boolean.
   - save payload emits only `{ slot: 0, weapon_id: 'laser_cannon', count: 1 }` when enabled.
   - this must become an installed-mount list with stable slots and per-row quantity controls.

3. `internal/battle/tactical.go`
   - Tactical spec validation currently rejects more than one weapon and rejects count != 1.
   - fire resolution currently rejects any weapon outside the one-Laser/count-1 fixture.
   - this must accept the newly supported Laser mount list/counts and resolve grouped volleys deterministically.

4. Tactical UI
   - the temporary single-weapon strip was removed in commit `12e7cbc`.
   - `selectedWeaponSlot` and legal per-slot fire actions remain in place and should be reconnected to a proper multi-slot selector rather than restoring the old strip unchanged.

## Scope boundary

This block intentionally does **not** pull all Slice 17 Ship Designer work into 15.6.

In scope now:

- multiple Laser Cannon mounts;
- per-mount Laser Cannon count > 1;
- stable mount slot identity;
- `+/-` quantity control in Ship Designer;
- adding the same weapon again as a new independent slot;
- correct per-count hull-space and cost accounting;
- persistence/design/build propagation;
- Tactical projection of every mount;
- Tactical per-slot selection/readiness;
- grouped Laser volley resolution;
- desktop and mobile usable weapon-slot selector;
- deterministic tests and real-browser QA.

Still deferred unless separately approved:

- broad additional weapon catalog/mechanics;
- missiles/ammunition implementation itself;
- weapon modifiers/upgrades;
- arc/facing systems;
- per-weapon targeting inside one grouped slot;
- per-weapon damage/destruction within a mount;
- general component-slot redesign beyond the weapon-mount rows;
- broad Slice 17 designer systems not required for this vertical slice.

## Proposed implementation sequence

### Block A - Contract amendment and rules tests

Document the explicit 15.6 freeze amendment and add failing/target tests before broad UI work.

Required cases include at least:

- one slot Laser x1 remains valid;
- one slot Laser x5 is valid if space allows;
- five slots Laser x1 are valid if space allows;
- duplicate slot ids are rejected;
- count <= 0 is rejected;
- more than 8 mounts is rejected;
- insufficient hull space is rejected using `count * weapon space`;
- design cost increases using `count * weapon cost`;
- persisted/reloaded design preserves slots and counts exactly.

### Block B - Authoritative Ship Designer rules

Generalize the current baseline validation for the supported Laser Cannon without opening unrelated weapon types.

For every mount:

- verify supported/known weapon technology;
- calculate `BaseSpace * Count`;
- calculate `BaseCostPP * Count`;
- preserve stable slot/count in the resulting design snapshot.

### Block C - Ship Designer UI

Replace the boolean Laser toggle with the already-planned available/installed structure.

Minimum interaction:

- available weapon row: `Laser Cannon` with an Add action;
- Add creates next free stable slot with count 1;
- installed rows show slot, weapon name, count, space/cost contribution;
- `-` / count / `+` controls modify quantity in that same slot;
- explicit remove action or decrement-to-zero removes the mount;
- adding Laser again creates a new slot instead of incrementing another slot automatically;
- live remaining-space feedback prevents/clearly rejects invalid increases/additions;
- mobile layout remains usable without horizontal page overflow.

### Block D - Tactical authority and grouped volleys

Lift the one-mount/count-1 Tactical guards only as far as the supported Laser baseline requires.

For each slot:

- project one legal fire action when ready and legal targets exist;
- preserve independent readiness;
- resolve `count` physical Laser shots for the selected slot using deterministic RNG ordering;
- aggregate/publish enough result data for UI feedback and replay/debugging;
- mark only the fired slot not ready;
- leave other mounts ready.

### Block E - Tactical weapon-slot UI

When an own active ship has weapons, expose a compact slot selector directly above the current action row.

Suggested compact labels:

- `S1 Laser x5`
- `S2 Laser x1`
- `S3 Laser x2`

Behavior:

- selected slot visibly highlighted;
- spent/not-ready slot visibly disabled/completed;
- slots with no legal target must not imply they can fire;
- selecting a slot changes the target legality projection used by battlefield clicks;
- after firing, selection should move to a sensible remaining ready slot or remain visibly spent with clear state;
- on mobile, use a compact horizontally scrollable slot strip rather than increasing HUD height excessively;
- retain the existing bottom action row: `Scannen / Warten / Fertig / Rückzug`.

### Block F - End-to-end QA

Mandatory comparison scenario:

1. Design A: one mount `Laser x5`.
2. Design B: five independent mounts `Laser x1`.
3. Build/use both through normal authoritative game flow.
4. Verify Design A fires one grouped five-shot volley from one selected slot and then that slot is spent.
5. Verify Design B exposes five independently selectable slots and can fire them separately.
6. Save/reload/reconnect must preserve design slots/counts and unresolved Tactical slot readiness.
7. Run desktop and 390px-class mobile browser QA.

## Acceptance criteria for this Gate-3 block

The block is complete only when all of the following are true:

- Ship Designer can create multiple Laser slots and change count per slot with `+/-`.
- Hull space and cost scale by physical weapon count.
- Saved designs preserve exact slot/count identity.
- Built ships carry the exact design mounts into Tactical.
- Tactical exposes every mount as its own selectable slot.
- `Laser xN` in one slot fires N deterministic physical shots as one grouped volley.
- Firing one slot consumes only that slot's readiness.
- Multiple separate slots can be fired independently.
- Mobile Tactical HUD remains compact and the existing four action buttons remain intact.
- tests cover designer validation/accounting, persistence, Tactical projection, grouped firing and independent readiness.
- `go test ./...`, web build and `git diff --check` are green.
- real-browser desktop/mobile QA passes.

## Crash / resume handoff

If this chat/session is lost, resume as follows:

1. Repository: `C:\ASH\Workspace\Projects\moox` on `main`.
2. Read this file first: `docs/research/SLICE_15_6_GATE3_WEAPON_SLOTS_TACTICAL_SELECTION_PLAN_2026-09-14.md`.
3. Also read the current session checkpoint for the weapon-slot design session if it is still available.
4. Confirm `git status` before editing; do not discard unrelated work.
5. Implementation is complete through commit `72508de`; `main` was clean/synced before this closure update.
6. The weapon-slot/count vertical slice is implemented end to end and validated on server rules, Ship Designer, Tactical authority and Tactical UI.
7. Resume normal remaining Slice 15.6 Gate-3 work; reopen this block only for new weapon mechanics or follow-up polish.

## Decision record

User approval on 2026-09-14:

- proceed with this narrow 15.6/G3 vertical slice now;
- use MOO-style distinction between quantity within one slot and repeated independent slots;
- finish the Tactical weapon-selection workflow by first expanding the Ship Designer/rules foundation;
- keep the architecture suitable for later missile/ammunition mechanics.
## Implementation completion - 2026-09-14

The approved weapon-slot vertical slice is complete.

Implemented commits:

- `f4d4be6` - authoritative military design accepts up to eight positive-count Laser mounts and accounts space/cost per physical weapon;
- `51fb804` - Ship Designer adds stable weapon-slot rows, add-new-slot behavior, per-slot quantity +/-/remove and dynamic cost/space preview;
- `f83a212` - design mounts propagate into Tactical, split slots retain independent readiness, and grouped mounts resolve one deterministic physical shot per weapon while spending only the selected slot;
- `80e1377` - Tactical HUD exposes compact selectable S1/S2/... weapon groups above Scannen/Warten/Fertig/Rückzug, including spent state and mobile horizontal layout;
- `72508de` - regression fixture updated so unsupported manual-lifecycle coverage continues to exercise a genuinely unsupported shield case instead of the newly supported Laser x2 case.

Validation:

- `go test ./... -count=1` passes;
- `npm run build` passes;
- `git diff --check` passes before closure;
- isolated real-browser Ship Designer QA proved both one slot `Laser x2` and two slots `Laser x1`, with identical 20/25 weapon-space usage and 35 PP total preview on the current Frigate baseline;
- 390x844 mobile Ship Designer QA showed no horizontal page overflow and 40px quantity-control touch row;
- isolated 390x844 Tactical QA showed S1/S2 above the unchanged four 32px action buttons in a 191px HUD;
- selecting S2 and firing changed the authoritative/UI state from 2/2 to 1/2 ready, marked only S2 as `Verbraucht`/disabled and automatically selected still-ready S1;
- QA used isolated port 7187 only; the canonical Triangle game on 7171 remained Turn 1 / planning / revision 1 / seed 32778 with zero battles.

Future missiles/ammunition, modifiers, arcs, per-weapon damage/destruction inside a mount and broader Slice-17 designer systems remain intentionally out of scope.
