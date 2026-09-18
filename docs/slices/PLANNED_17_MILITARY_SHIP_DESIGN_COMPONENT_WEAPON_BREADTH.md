# Planned Slice 17 - Military Ship Design Program

Status: **active parent program; child 17.0 open at Gates 1-2 complete / Gate 3 ready**.

Queue position: **current gameplay program after closed Slice 16**. Slice 18 is reserved downstream for Galaxy/Strategic Navigation fidelity and is not a prerequisite for Slice 17.

## Objective

Expand the intentionally narrow military-design baseline into a complete server-authoritative ship-construction foundation that later Tactical/Triangle combat can consume without hard-coded fixture ships.

Slice 17 is a **parent program**, not one monolithic implementation slice. The original single 17-plan became too broad once full hull, mandatory-component, weapon, ordnance, special/fighter, designer, construction and QA requirements were considered.

## Binding decomposition

Slice 17 is executed in dependency order as:

1. **17.0 - Ship Designer / Triangle Reference Test Harness**
   - durable development-only reference scenarios;
   - high-tech/all-tech designer laboratories;
   - deterministic construction acceleration using normal turn resolution;
   - direct built-Ship / battle fixture entry for Tactical/Triangle QA.
2. **17.1 - Hulls and Mandatory Ship Systems**
   - Frigate through Doom Star;
   - drives, computers, armor, shields, fuel;
   - hull/buildability technology and infrastructure restrictions;
   - cost/space/Command-Point and miniaturization foundation.
3. **17.2 - Beam Weapons and Beam Mount Modifiers**
   - Beam weapon breadth;
   - mount count/slots;
   - applicable Beam modifiers;
   - cost/space/unlock/miniaturization metadata;
   - compatibility with the existing Laser tactical vertical slice.
4. **17.3 - Missiles, Torpedoes and Bomb Ordnance**
   - design-side ordnance representation, ammunition/counts and legal mounting;
   - authoritative snapshots and costs;
   - full tactical projectile behavior remains owned by later Tactical work unless explicitly pulled in by Gate 1.
5. **17.4 - Ship Specials and Fighter Systems**
   - design-side ship specials, Fighter/Bomber-related systems and special-derived ship stats;
   - exact tactical effects may remain downstream where their combat subsystem is not yet implemented.
6. **17.5 - Full Ship Designer Integration and Acceptance**
   - unified catalog/UI;
   - design revision and immutable built-Ship snapshots;
   - Colony Construction handoff;
   - persistence/replay;
   - broad designer/reference-lab/browser QA.

Only one 17.x child may be active unless a checkpoint explicitly documents a justified parallel exception.

## Existing foundation

The program starts from substantial already-implemented infrastructure:

- persisted Empire ShipDesigns;
- built Ships carry immutable ShipDesignSpec snapshots;
- design IDs and revisions survive persistence;
- Colony Construction can build a selected military design;
- Command Points and Combat Fleets already consume built ships;
- normalized six-hull catalog exists;
- mandatory drive/computer/armor/shield/fuel fields already exist in ShipDesignSpec;
- weapon mount snapshots already exist but current runtime validation intentionally permits only the narrow supported baseline;
- Slice 15.6 already provides the browser Ship Designer shell and Colony Build -> Designer -> Construction handoff;
- Slice 07 provides a deterministic tactical Laser vertical slice and original combat evidence;
- development-only `-reference-games` already registers normal reference games plus the durable 3-player 2-pc Triangle scenario.

## Program architecture rules

- A reference/test harness may alter **initial state**, not gameplay formulas.
- Test scenarios must use the same designer, validation, construction, turn, fleet and battle code paths as ordinary games after bootstrap.
- No debug-only technology or free-build rule may leak into normal New Game/API behavior.
- Built Ship snapshots remain immutable when a design slot is later revised.
- Design legality is server-owned; HMI consumes catalogs and rejection reasons.
- Slice 17 defines what is installed on a ship. Later Tactical/Triangle work may own the full combat-time behavior of systems whose tactical mechanics are not yet implemented.
- A design must never be advertised as tactically complete merely because it can be persisted.

## Parent opening guard

Before opening 17.0 Gate 1:

- Git clean;
- zero unrelated `_OPEN_` markers;
- Slice 16 closed;
- Slice 18 remains planned only;
- current reference-game/Triangle scenario still green;
- current narrow military design/construction/tactical Laser regressions green.

## Parent completion contract

- [ ] 17.0 closed: durable Ship/Combat QA harness exists.
- [ ] 17.1 closed: hulls + mandatory systems are fully supported to the frozen breadth.
- [ ] 17.2 closed: Beam design breadth is complete to the frozen contract.
- [ ] 17.3 closed: ordnance design breadth is complete to the frozen contract.
- [ ] 17.4 closed: supported specials/fighter design breadth is complete.
- [ ] 17.5 closed: unified designer/construction/persistence/browser acceptance passes.
- [ ] Every deferred tactical effect has an explicit downstream owner and is not represented as already working.
- [ ] Final full Go/vet/web/browser/persistence/reference-game QA passes.

## Exit criterion

Slice 17 is complete when a player can construct and persist a broad, technology-gated military ship design across the frozen hull/component/weapon/special matrix, build that exact immutable design through normal Colony Construction, inspect/test it through durable reference scenarios, and hand the resulting authoritative ship snapshot to later Tactical/Triangle systems without special-case fixture-only design data.
