# Planned slice 15 - Playable browser strategic HMI epic

Status: **parent epic / prepared; no direct implementation gate**.

Queue position: **15 family of 17**.

## Objective

Turn the transport-proof React HMI into the first genuinely playable, visually coherent browser game for the supported Human-vs-built-in-AI match while keeping all gameplay authority in Go.

Slice 15 is deliberately split into six gated sub-slices so functional authority, UX architecture, visual exploration, rich decision/persistence presentation, interactive Tactical Combat and final end-to-end acceptance do not collapse into one oversized implementation gate.

## Dependencies

- Slice 08 server/web authority baseline.
- Slice 13 built-in AI baseline.
- Slice 14 live save/resume baseline.

## Sub-slice sequence

1. **15.1 - UX architecture / navigation / design system contract**
   - screen map, information architecture, navigation, responsive shell, interaction grammar and baseline design tokens.
2. **15.2 - Functional strategic gameplay HMI**
   - authoritative browser workflows for the accepted 2D Galaxy + star-system dialog, Colony table/detail, location-grouped Fleets, eight-category Research, Diplomacy and turn control using server-projected legality; Espionage is a first-class product area whose missing mechanics must be audited/split rather than faked in React.
3. **15.3 - MOOX visual identity / graphics / asset pipeline**
   - corporate identity, typography/color/icon language, imagery/model/render strategy, asset provenance and runtime asset integration.
4. **15.4 - Rich gameplay decisions / persistence UX**
   - Encounter entry/return, invasion/research/colony decision presentation plus save/load/reconnect/invalidation/error UX; the interactive battlefield is explicitly deferred to 15.5.
5. **15.5 - Interactive 2D tactical combat**
   - extend the narrow Slice-07 fixed-position Laser baseline with server-authoritative movement/legal targets and a mouse/touch interactive 2D battlefield.
6. **15.6 - Full browser vertical slice / polish / complete-game QA**
   - complete Main Menu -> New Game -> strategic play -> save/reload/resume -> interactive Tactical battle -> victory path without developer tooling.

## Epic scope guard

The 15.x family owns browser playability and presentation for already-authoritative gameplay. Accepted strategic product areas are **Galaxy, Colonies, Fleets, Research, Diplomacy and Espionage**. Desktop should expose them directly in the left rail; mobile keeps the compact five-slot bottom navigation with Diplomacy/Espionage as first-class entries under `More`. It must not move rules, legality, costs, target selection, AI policy or deterministic simulation into React. From Slice 15.1 onward the browser HMI is also **multi-language by contract**, initially German and English with runtime switching; locale remains presentation-only and all later 15.x UI work must use the shared localization layer rather than hard-coded player-facing strings.

Still deferred beyond the family unless explicitly pulled forward by a sub-slice contract:

- broader New Game/preset-race breadth from Slice 16;
- broad military Ship Designer component/weapon breadth from Slice 17;
- tactical breadth beyond the explicit interactive baseline accepted by Slice 15.5;
- full Race Designer;
- native/Wails packaging;
- copyrighted original-game asset reuse without explicit provenance/license permission.

## Parent milestone acceptance

Slice 15 is complete only when all sub-slices 15.1-15.6 are closed and a human can:

- start the supported Human-vs-built-in-AI game from the browser;
- understand and execute all currently supported strategic decisions without direct API/dev-tool use;
- save, reload/reconnect and resume deterministically;
- resolve supported Encounter/Invasion/Colony-Base/Research decisions and play the accepted interactive Tactical Combat path;
- reach and understand the final result;
- do so through one coherent MOOX visual/interaction language.

Milestone after **15.6**: first saveable, resumable and finishable Human-vs-built-in-AI MOOX browser vertical slice with an integrated interactive Tactical Combat path.
