# Slice 15.6 Gate 3 - Closeout deferrals

Date: 2026-09-14
Status: **FROZEN FOR 15.6 CLOSEOUT**

## Purpose

Slice 15.6 has reached the accepted playable browser baseline for Ship Design -> Colony construction -> fleet movement/encounter -> interactive Tactical -> strategic continuation. The following items are useful future breadth, but they are **not dependencies for Slice 15.6 Gate 3/Gate 4 closure**. They must not remain ambiguous `OPEN` items under the active 15.6 milestone.

## Deferred breadth and ownership

### 1. Foreign Fleet detection / scanner intelligence (former Block5F)

Deferred beyond Slice 15.6.

The planned foreign deep-space Fleet marker, scanner-authorized route/ETA projection, redacted ship-detail depth, stacking/source rules and last-known-contact behavior require an explicit server-authority research/freeze step. This belongs to later strategic intelligence/scanner breadth; Slice 20 Intelligence is a natural candidate, but exact ownership must be frozen when that breadth starts.

The already-implemented co-located/visited-system foreign Fleet inspection from Block5E remains part of the accepted 15.6 baseline.

### 2. Persistent per-instance visual genomes for special strategic ships

Deferred beyond Slice 15.6.

Colony Ship, Outpost Ship and Troop Transport currently use the accepted stable special-vessel presentation path but do not yet persist a per-instance `visual_genome` through construction, save/load and reconnect. This remains a visual/persistence backlog item. Static special silhouettes are sufficient for the 15.6 complete-game acceptance and do not block Gate 4.

### 3. Tactical activation batching / preview and fire-all semantics

Deferred beyond Slice 15.6.

Current Tactical commands intentionally mutate server authority immediately. `Wait`, `Finish`, direct movement and selected weapon-slot fire are authoritative and browser-tested. A staged activation commit model, no-selection `fire all`, split-fire within one grouped mount, and other deeper activation semantics require later Tactical authority work. They must not be simulated in React.

Per `PROJECT_STATUS.md`, additional Tactical breadth is intentionally re-audited after the currently prepared roadmap rather than being pulled into this milestone.

### 4. Broad Military Ship Designer component/weapon breadth

Deferred to **Slice 17**.

Slice 15.6 keeps the now-proven Frigate/Laser vertical slice, including persistent named designs, grouped Laser counts, multiple independent Laser slots, authoritative cost/space and exact Tactical slot readiness. Additional hull save support, missiles/ammunition, bombs/fighters, specials, broader components, modifiers/upgrades, arcs/facings and per-weapon destruction belong to Slice 17.

### 5. Further Tactical visual polish

Deferred beyond Slice 15.6 unless a Gate-4 regression exposes a concrete usability blocker.

The current Tactical screen is explicitly accepted by the user as a playable baseline, not a final presentation. Cosmetic hierarchy, animation and layout improvements may continue later without reopening the closed 15.6 Tactical review.

## Closeout rule

Only failures against the frozen Gate-2 acceptance contract may reopen Slice 15.6 implementation work now:

- visible browser start/new-game journey;
- mid-game Save/Load/Resume;
- authoritative Ship Design -> Colony handoff;
- normal fleet movement/encounter path;
- at least one supported interactive Tactical battle;
- strategic continuation through normal Invasion/Conquest/Victory;
- desktop/mobile runtime guardrails;
- server authority / stale-client / regression guarantees.

Everything else listed above is explicitly deferred and must not block the final canonical acceptance run or Gate-4 closure.
