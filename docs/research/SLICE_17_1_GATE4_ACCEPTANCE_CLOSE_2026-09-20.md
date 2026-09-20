# Slice 17.1 Gate 4 - Acceptance and Close

Date: 2026-09-20
Result: **PASS**
Tested gameplay/UI baseline: `d9bab26 fix(ui): compact installed weapon controls`

## Scope

Slice 17.1 closes the Hulls and Mandatory Ship Systems child of the Slice-17 military-design program.

Accepted breadth:

- six supported hull classes from Frigate through Doom Star;
- hull technology locks and Cruiser+ orbital-dock buildability authority;
- automatic Drive/Fuel behavior;
- save-time Armor authority;
- explicit selectable Computer;
- explicit selectable Shield including no-Shield;
- original-evidenced mandatory-component miniaturization;
- current-research effective saved-design projection;
- fixed queue-time construction snapshots;
- command-point derivation and stable rejection semantics;
- durable Baseline/Mid/All-Tech reference fixtures;
- mobile/desktop Ship Designer HMI, including compact preview, mandatory-system information, effect/help surfaces, defense ratings, stable installed-weapon baseline and construction/save handoff.

Broader Beam, ordnance and special-system breadth remains downstream in Slices 17.2-17.4.

## Gate-4 acceptance matrix

### 1. Every frozen hull can be designed only when buildable/unlocked - PASS

Focused coverage:

- `TestSlice171CatalogProfilesAndLateHullLocks`;
- `TestSlice171OrbitalDockBuildabilityAndQueueSnapshot`;
- `TestSlice171FrigateDestroyerNeedNoOrbitalDock`.

The Baseline/Mid/All-Tech catalog matrices preserve Titan/Doom Star research locks and Cruiser+ orbital-dock construction authority. Frigate/Destroyer remain independent of the Cruiser+ dock gate.

### 2. Every frozen mandatory component obeys technology/space/cost rules - PASS

Focused coverage:

- `TestSlice171MiniaturizationGoldenVectors`;
- `TestSlice171ExplicitComputerAndNoShieldSelection`;
- full Slice-17.1 catalog/save validation tests.

The frozen original miniaturization factors, Shield space factors, component technology legality, current effective cost/space projection and explicit Computer/Shield semantics remain deterministic.

### 3. Doom Star and late-game configurations are testable through Slice 17.0 - PASS

The durable All-Tech Triangle reference game exposes the six-hull Designer directly, including Titan and Doom Star, without research grinding.

The browser acceptance traverses all six hulls, verifies the late-hull Designer path and keeps the mobile/desktop preview, information/help and save flows operational.

### 4. Revision never mutates already-built Ships - PASS

Focused coverage:

- `TestBuiltMilitaryShipKeepsWeaponSnapshotAcrossDesignRevision`;
- `TestReferenceMilitaryShipFixtureUsesProductionDesignAndImmutableSnapshot`;
- queue-time snapshot coverage inside `TestSlice171OrbitalDockBuildabilityAndQueueSnapshot`.

Built/reference Ships and queued construction preserve immutable authoritative snapshots rather than re-reading later design revisions.

### 5. Construction and persistence round trips pass - PASS

Focused coverage includes:

- `TestMilitaryConstructionChoiceQueueCompletionAndSnapshot`;
- `TestReferenceMilitaryShipFixtureUsesProductionDesignAndImmutableSnapshot`;
- `TestCombatFleetTransitRoundTripsSchema23`;
- complete `go test ./...`.

No Slice-17.1-specific persistence or construction regression remains open.

### 6. Full tests/vet/web/browser/diff checks pass - PASS

Fresh Gate-4 commands:

- focused Slice-17.1/snapshot/reference/persistence test selection: PASS;
- `go test ./...`: PASS;
- `go vet ./...`: PASS;
- `npm run build` from `web`: PASS;
- `node --check scripts/check-reference-lab-browser.mjs`: PASS;
- `npm run check:reference-lab:browser -- http://127.0.0.1:7171`: PASS;
- `git diff --check`: PASS.

The browser acceptance includes:

- six-hull All-Tech Designer;
- late-hull selection;
- compact/contained and progressively scaled ship preview;
- mandatory system information/help;
- Computer/Shield card selectors and explicit no-Shield;
- Beam Defense/Missile Evasion values and help formulas;
- compact stable installed-weapon baseline;
- Space Available beside component installation;
- production cost beside Save Design;
- mobile 390px and desktop layout;
- same-turn drafted construction buyout and next-turn completion;
- ordinary-game isolation from reference-only development controls.

## Exit criterion

**Satisfied.**

The normal Designer can author and build the frozen hull + mandatory-system combinations through one authoritative rules pipeline, including late-game configurations through durable reference laboratories, without claiming complete Beam/ordnance/special breadth.

## Downstream handoff

Slice 17.2 owns Beam Weapons and Beam Mount Modifiers.

The accepted Slice-17.1 HMI leaves a compact installed-weapon surface and stable scroll region that Slice 17.2 may extend. The user-approved direction is to keep the main Designer compact and move broad weapon selection/configuration into a dedicated selection dialog, while weapon-specific modifiers remain owned by the Beam contract rather than being invented in 17.1.

Slice 17.1 is closed. No further 17.1 gameplay work is required before opening Slice 17.2.
