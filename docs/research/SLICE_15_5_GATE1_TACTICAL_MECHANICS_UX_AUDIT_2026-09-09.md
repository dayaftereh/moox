# Slice 15.5 Gate 1 - Tactical mechanics / UX audit

Date: 2026-09-09
Status: **Gate 1 complete**

## Objective

Re-open the deliberately narrow Slice-07 Tactical baseline as the foundation for a credible interactive browser battle. Close the original MOO2 movement/facing facts that Slice 07 explicitly deferred, preserve Go/server authority, define a practical first multi-ship breadth, and establish desktop/touch interaction plus deterministic acceptance strategy before Gate-2 freeze.

## Inputs audited

- `docs/research/TACTICAL_SHIP_COMBAT_BASELINE_2026-09-01.md`;
- `docs/research/SLICE_15_5_INTERACTIVE_TACTICAL_COMBAT_PRODUCT_DIRECTION_2026-09-04.md`;
- `docs/slices/PLANNED_15_5_INTERACTIVE_TACTICAL_COMBAT.md`;
- current `internal/battle` Tactical/BattleSession state and commands;
- current Session/App/browser Battle handoff from Slice 15.4;
- original MOO2 HELP block 0;
- original `Orion2.exe` Watcom tactical symbols and object-1 code.

Private executable extracts used for disassembly were deleted after analysis and are not repository artifacts.

## Existing authoritative MOOX baseline

Slice 07 already owns a deterministic BattleSession-local tactical state for one deliberately narrow fixture:

- Tactical ships with fixed post-deployment integer X/Y positions;
- initiative/activation order;
- exact deterministic RNG;
- Armor/Structure damage and destruction;
- one standard Laser path;
- `battle.fire_beam`;
- `battle.end_activation`;
- terminal Battle result reconciliation back into the strategic Session.

The current browser Battle route from Slice 15.4 already provides the stable strategic entry/return shell. The missing credible gameplay layer is principally movement/facing/legal-action projection plus a real battlefield UI around the existing fire/activation path.

## Original HELP evidence

Original HELP directly establishes the following tactical semantics:

- combat speed determines how far a ship may travel in a combat round;
- the current-ship display exposes **remaining movement points**;
- the combat map represents the whole combat area and supports viewport recentering;
- a missile warning can interrupt an already ordered move and then allow the ship to **continue moving**;
- `Wait` moves to another active ship while keeping the current ship in the active list;
- `Done` moves on and removes the current ship from the active list;
- weapon readiness is an explicit visible state;
- tactical distances are described in **squares**;
- a hyperspace jump preserves direction and the ship must rotate normally to change facing;
- **Inertial Stabilizer** halves movement cost for turning;
- **Inertial Nullifier** allows direction changes without movement cost.

## Original Tactical Scan / ship inspection

Original HELP has a dedicated **Scan Button** in Tactical Combat. It toggles scan mode; while scan mode is active, clicking **any ship** opens a detailed display of that ship, and moving the viewport is the only other accepted command. This is distinct from the `Battle Scanner` ship technology: HELP describes Battle Scanner as a beam-to-hit / galactic scanning bonus, not as a prerequisite for the Tactical Scan UI.

The surrounding original Tactical UI evidence also exposes:

- a Damage Bar with separate Armor, Structure and internal-system damage;
- an Internals/Target display that switches to target information when pointing at a ship;
- a Weapons Display with per-weapon readiness/status;
- executable functions `Scan_Ship_Info_`, `Scan_Ship_D_Info_` and `Get_Ship_Strength_Info_`.

For MOOX this supports a participant-safe **read-only scan/inspection mode**. The first Tactical slice should expose concrete authoritative combat facts rather than invent an opaque client-side strength score:

- ship identity / design / size where player-safe;
- weapon loadout and supported readiness;
- Armor and Structure current/max plus clear damage percentages;
- internal-system damage when that subsystem is later normalized;
- current movement allowance, position and facing;
- authoritative offense/defense or other normalized combat ratings when present.

A single composite `combat_strength` number should only be added after its original/runtime formula is separately closed; the UI can already answer "how strong is it?" through the concrete offensive, defensive, weapon and survivability data above.

Scan itself does not need to be a mutating battle command. It can be a client presentation mode over player-safe server projection. To mirror the original interaction safely, scan mode may disable move/fire target selection while remaining pannable/zoomable; exiting Scan restores normal command selection.
The initiative baseline remains:

`initiative = modified beam offense / 10 + current combat speed`

with highest initiative acting first when Tactical Initiative is enabled.

## Executable movement symbols

The original executable exposes adjacent Watcom symbols for the movement subsystem. Corrected object-1 offsets are:

| Function | Object-1 offset |
| --- | ---: |
| `Move_Ship_` | `0x2EE0F` |
| `Get_Facing_` | `0x2F5F1` |
| `Rotate_Ship_` | `0x2F628` |
| `Turn_Cost_` | `0x2F923` |
| `Move_Cost_` | `0x2F95F` |
| `Turn_To_Face_` | `0x2FA61` |

The same symbol neighborhood includes `Ship_In_Legal_Square_`, confirming that the original movement system distinguishes movement legality from mere rendering.

## Exact original facing model

`Get_Facing_` computes a direction angle from source/destination coordinates, normalizes it to 0..359 degrees, then quantizes it as:

`facing = ((angle * 4 + 45) / 90) mod 16`

This is the nearest of **16 discrete facing sectors**, each approximately 22.5 degrees.

The shortest facing difference therefore lives in `0..8` steps around the 16-way circle.

## Exact original turning cost

`Turn_Cost_` returns an integer movement-point cost per facing step:

- normal ship: **2**;
- Inertial Stabilizer path: **1**;
- Inertial Nullifier path: **0**.

The executable checks two tactical special/component identities (`0x12` for the half-cost path and `0x11` for the zero-cost path); HELP independently identifies the matching Stabilizer/Nullifier behaviors above.

## Exact original destination movement cost

`Move_Cost_` reads current X, Y and facing from the Tactical ship record, returns zero for the current coordinate, and computes destination cost in two parts.

### Translation component

For `dx = destinationX - currentX` and `dy = destinationY - currentY`:

`translationCost = ceil(sqrt(dx^2 + dy^2))`

The executable implements the integer ceiling by increasing an integer radius until `radius^2 >= dx^2 + dy^2`.

### Turning component

The server derives the destination facing through `Get_Facing_`, computes the shortest circular difference on the 16-sector facing ring, and multiplies it by `Turn_Cost_`.

### Total

`moveCost = ceil(sqrt(dx^2 + dy^2)) + facingSteps * turnCostPerStep`

This means facing/turning is not decorative: it is part of authentic movement legality and budget consumption.

## Movement distance is not weapon range

The existing Slice-07 beam range evidence remains separate:

`rawRange = max(abs(dx), abs(dy)) + floor(min(abs(dx), abs(dy)) / 2)`

and ship-to-ship beam range index is:

`(rawRange + 2) / 3`

Movement therefore must **not** reuse the beam-range metric. The movement formula is Euclidean-ceiling plus turning cost; weapon range is the separate original range approximation already used by the supported Laser path.

## Gate-1 facing decision

Facing must enter the minimum Slice-15.5 movement model. The earlier plan allowed Gate 2 to pull facing/turning in only if evidence proved it inseparable; the original executable now proves exactly that.

Minimum accepted facing scope should be:

- authoritative current facing `0..15` per Tactical ship;
- server-derived resulting facing for a destination;
- turning cost included in movement legality;
- visual ship rotation from authoritative facing.

Still deferred unless separately accepted later:

- shield-facing mechanics;
- weapon firing arcs;
- advanced rotate-only commands;
- broad special systems that depend on facing.

The current Laser/no-shield baseline does not require those broader systems merely because movement now has facing.

## Minimum credible combat breadth

Slice 07 intentionally accepts exactly one combat ship per side. That is no longer a credible browser gameplay floor because a normal fresh MOOX game starts each empire with two scouts; the user already reached a 2-vs-2 encounter that was rejected by the Slice-07 support predicate.

Gate 1 therefore recommends **1-2 combat ships per side** as the first Slice-15.5 supported breadth, with 2-vs-2 as the main acceptance fixture.

Why 2-vs-2:

- normal early-game fleets can enter Tactical without a crafted 1-vs-1 save;
- initiative/active-ship cycling becomes real gameplay rather than a placeholder;
- Wait/Done/end-activation semantics can be exercised meaningfully;
- UI/data structures are forced not to hard-code attacker/defender singletons;
- it remains narrow enough to avoid solving the final large-fleet battlefield in one slice.

Implementation structures should nevertheless use slices/IDs and remain N-vs-M shaped so a later breadth expansion does not require a protocol redesign.

## Required player-safe Tactical projection

Gate 1 recommends the following server-owned shape, finalized by Gate 2:

- Battle round;
- active seat and active ship ID;
- initiative/activation order in player-safe form;
- Tactical ships:
  - stable strategic ship ID / empire / seat;
  - X/Y;
  - facing `0..15`;
  - current/max movement points;
  - Armor/Structure current/max;
  - destroyed/active state;
  - supported weapon readiness;
- `legal_moves[]` for the active player ship:
  - destination X/Y;
  - authoritative movement cost;
  - authoritative resulting facing;
  - remaining movement points after move;
- legal fire actions / supported weapon identity;
- player-safe legal target IDs and authoritative range information;
- `can_end_activation`.

The browser may visualize only what is projected. It must not reproduce hidden legality or cost math to decide what is clickable.

## Command envelope recommendation

Continue using the existing battle-local `protocol.Command` sequencing/transaction model.

Add:

`battle.move_ship`

Minimum trusted payload:

- `ship_id`;
- `x`;
- `y`.

The client does **not** submit trusted movement cost, remaining points or facing. The BattleSession re-derives legality, cost and resulting facing and commits atomically or rejects with zero mutation.

Keep and integrate:

- `battle.fire_beam`;
- `battle.end_activation`.

## Desktop interaction target

- battlefield is the primary surface;
- active ship is visually emphasized;
- server-projected legal destinations are visible and selectable;
- click a legal destination to move;
- action/status panel shows movement remaining, facing, Armor/Structure and supported weapon state;
- selecting the Laser reveals only server-projected legal targets;
- click target to fire;
- End Activation advances authoritative initiative;
- optional hover may preview information but can never be required.

## Mobile/touch target

- battlefield consumes most of the viewport;
- one-finger pan and pinch zoom where battlefield size requires it;
- tap legal destination / ship / target;
- selected ship/action details in a compact bottom sheet or overlay;
- primary actions meet the existing 44-px touch target baseline;
- no workflow depends on hover;
- the server-projected move/target overlays remain the source of interactive legality.

## Battlefield extent: original evidence versus MOOX direction

Original HELP says the reduced combat map shows the **entire area of the combat**, and the executable contains `Ship_In_Legal_Square_` / `Set_Legal_Moves_`. The original therefore has a defined combat area/legal-square concept; an actually unbounded battlefield is **not** an original-fidelity claim.

For MOOX the preferred product direction is intentionally different: the battlefield should feel effectively **unbounded**.

- do not put a fixed `battle_width` / `battle_height` into the first public Tactical contract;
- do not render a visible arena edge or wall;
- Tactical X/Y remain authoritative integer world coordinates;
- initial deployment is relative to an encounter origin, not to permanent board corners;
- the camera may pan/zoom freely and initially fits the participating ships;
- legal destinations are still a finite server-projected local set because movement points are finite each activation;
- collision/occupancy/other accepted legality still applies;
- implementation may retain a very large technical integer safety envelope, but it must not function as a normal gameplay boundary.

This keeps authentic movement economics while avoiding an artificial small-arena feel. It also scales better to later larger fleets because the camera/coordinate model is not tied to a fixed board rectangle.
## MOOX modernization versus original fidelity

Evidence-backed mechanics to preserve:

- integer Tactical X/Y authority;
- movement-point budget tied to combat speed;
- 16 facings;
- original translation + turn-cost formula;
- initiative and current Laser range/hit/damage chain;
- activation semantics and result reconciliation.

Presentation that may be modernized without changing authority:

- a smooth scalable, effectively unbounded 2D coordinate plane rather than reproducing the original finite combat-area pixels;
- highlighted finite server-projected legal destinations around the active ship, with no visible global arena edge;
- pan/zoom gestures;
- mobile bottom sheet instead of original fixed side panels;
- optional desktop hover alongside equivalent click/tap selection;
- richer MOOX procedural ship art rotated to the authoritative facing.

## Regression / acceptance strategy

The Gate-2 minimum acceptance set should require:

1. enter a supported Tactical battle from normal strategic browser gameplay, not API-only injection;
2. server projects active ship plus legal movements;
3. click/tap one legal destination and observe authoritative X/Y, facing and movement-point update;
4. stale/illegal/over-budget move is rejected with zero Tactical mutation and clear client refresh/error handling;
5. select the supported Laser and fire at a server-projected legal target;
6. observe authoritative readiness, range outcome, damage/destruction and refreshed state;
7. progress activation/initiative through a real 2-vs-2 fixture and reach a new round;
8. verify round reset of movement/readiness according to authority;
9. resolve the Battle and reconcile exact destroyed/surviving strategic ship IDs through the frozen Slice-15.4 return surface;
10. reconnect/restore mid-Tactical and recover exact active ship, positions, facings, movement budget, readiness and sequence without duplicate commands;
11. verify desktop plus phone/touch behavior, pan/zoom and no required hover.

## Gate-1 conclusion

Gate 1 is complete. The largest new original-fidelity result is that **movement and facing must be designed together**: original MOO2 destination cost is Euclidean-ceiling travel plus 16-way turn cost. The recommended first credible browser breadth is **up to 2 ships per side**, with 2-vs-2 acceptance. Gate 2 must freeze this contract before implementation begins.
