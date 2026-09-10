# Slice 15.6 Gate 2 - Full browser vertical slice / Ship Designer contract freeze

Date: 2026-09-10
Status: **FROZEN / explicitly accepted by user**

The user explicitly accepted the Slice-15.6 Gate-2 contract after the Ship Designer refinements, including persistent named designs, left/right hull stepping, technology locks, available/installed component lists, Production/Command-Point summaries, exact Colony construction handoff, and click/tap procedural SVG reroll.

Frozen implementation boundary:

- canonical milestone remains a complete browser journey from Main Menu/New Game to authoritative Human-vs-Built-in-AI Conquest Victory using seed `0x8009`, Small/Normal/Average, Human local_human vs Darlok builtin_ai, Strategic Combat disabled;
- the milestone acceptance run uses visible browser workflows only: no direct state mutation, imported snapshot, console editing, test-only opponent controller, direct Battle construction or hidden Tactical autoresolve in place of the required supported Tactical fight;
- Save/Load/Resume is mandatory mid-game and must preserve authoritative game identity, turn/revision and meaningful progress;
- at least one ordinary strategic encounter must enter the accepted interactive Tactical path with 1-2 visible combat ships per side, then perform Move/Scan/Laser Fire/activation progression and return through Battle Return before strategic continuation;
- the complete run continues through normal Troop Transport/Invasion UX to authoritative Conquest/Victory.

Frozen Ship Designer UX foundation:

- every Ship Design has a player-editable **name**;
- hull class is browsed with **left/right arrow controls**, not a hull image-tile chooser;
- canonical hull order is Frigate -> Destroyer -> Cruiser -> Battleship -> Titan -> Doom Star;
- the procedural ship graphic remains a large visual preview;
- clicking/tapping that preview explicitly **rerolls only the procedural SVG/visual genome** for the selected hull; design name, hull, installed equipment, cost/space, Command Points and gameplay legality remain unchanged;
- desktop click and mobile tap both support visual reroll;
- hull/component availability and lock reasons come from server authority; React must not infer gameplay unlock rules;
- unavailable technology-gated hulls/components may remain visible for progression context but are clearly locked/disabled/struck and cannot be saved/built;
- player-visible available weapons/components are shown in an **Available** list;
- installed weapons/components are shown separately below in an **Installed** list with stable slot/mount identity;
- the UI structure is future-ready for multiple separate mounts, per-mount quantity +/- and later weapon modifiers/upgrades, but controls may not imply functionality the server does not yet authorize;
- bottom summary displays authoritative **Production Cost**, **Command Point cost/impact**, and used/available design space where projected;
- saving/revising persists a named design and revision in the player's design catalog with visual identity/loadout;
- the exact saved design appears on Colonies as a Military Ship construction choice carrying `ship_design_id`, `ship_design_revision` and `ship_design_name`;
- completed ships retain source design identity/revision.

Frozen first functional breadth inside that durable shell:

- current authoritative save semantics remain the starting functional floor: named Frigate design with no weapon or exactly one slot-0 Laser Cannon count 1;
- server-derived mandatory drive/computer/armor/fuel/shield, cost and space remain authoritative;
- current normal-ship Command Point rule is server-owned and must be projected rather than recomputed as React gameplay authority;
- broader hull save support, multiple active mounts, weapon count >1, modifiers, specials, missiles/bombs/fighters and per-slot Tactical destruction are not to be faked in the browser; they require explicit later authority/Tactical expansion;
- nevertheless, the shell must be shaped now so those additions do not require a UI redesign.

Frozen responsive/runtime guardrails:

- 1440x900 desktop target, 390px representative mobile, 360px functional minimum, 320px catastrophic-overflow guard;
- no required horizontal page scrolling and no required hover;
- primary/secondary mobile workflow actions >=44px;
- same-host normal route + authoritative snapshot ready <=2 seconds;
- initial canonical-route transfer <=1 MiB;
- individual JS chunk <=350 KiB unless explicitly reviewed/justified;
- no uncaught browser exception/unhandled rejection, missing active production asset, permanent spinner, or navigation loop;
- stale HTTP 409 remains one submit -> visible rejection -> authoritative refetch, never auto-resubmit;
- refresh failure preserves readable last state and exposes Retry;
- Restore remains explicit confirmation;
- strategic and Tactical surfaces may never be simultaneously writable.

Implementation sequence after this freeze:

1. server-authoritative Ship Designer projection/command contract required by the frozen shell;
2. browser Ship Designer shell and narrow Frigate/Laser save path;
3. persistent design catalog + exact Colony build handoff QA;
4. canonical seed dry run against Built-in AI and repair only real integration blockers;
5. full visible-browser Save/Resume/Tactical/Victory acceptance plus responsive/runtime pass;
6. independent Gate-4 close.

Gate 3 may now implement within this boundary in small checkpointed/recovered blocks.
