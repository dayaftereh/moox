# Planned slice 15.2 - Functional strategic gameplay HMI

Status: **closed; Gates 1-4 complete (independent Gate-4 QA passed 2026-09-07)**.

Queue position: **15.2 of Slice-15 family**.

## Objective

Make the currently authoritative strategic game genuinely operable from the browser using server-projected state, legal actions, targets and costs, while implementing the accepted game-specific strategic UX: 2D Galaxy map, star-system inspection, Colony table/detail management, location-grouped Fleets, eight-category Research, Diplomacy and a planned Espionage product area. Visual styling remains intentionally functional rather than final.

## Dependencies

- Slice 15.1 UX/navigation shell.
- Slice 13 built-in AI.
- Slice 14 live save/resume baseline.

## Scope guard

In scope:

- supported Human-vs-built-in-AI New Game entry using current narrow settings;
- **Galaxy as the primary interactive 2D map**, using the current no-fog/all-public authoritative system projection, player-safe ownership/fleet/travel information, pan/zoom and star selection;
- an in-map **star-system dialog** with persistent authoritative star metadata plus normal planets, gas giants and asteroid belts, Colony/Outpost state and present player Fleets/Ships;
- server-authoritative planet actions from the system dialog, including a **Colonize confirmation flow** only when a legal colonization choice is projected;
- **Colonies as a sortable management table** with quick population/job controls plus a canonical full **Colony Detail** screen reachable from both the table and owned planets in the star-system dialog;
- Colony Detail population/jobs header, **planet/Colony round-up with Population growth/capacity/output plus built Buildings/infrastructure**, and a real ordered authoritative construction queue with no user-visible item-count cap;
- **Research as a visual eight-category screen** (Construction, Power, Chemistry, Sociology, Computers, Biology, Physics, Force Fields), with Creative/Uncreative behavior driven by projected authoritative legal choices rather than browser rules;
- **Fleets as ship/fleet tiles grouped by system/location**, plus in-transit grouping, inspection, movement, colonization and Outpost workflows;
- **Diplomacy as a first-class strategic area** using the existing war/peace baseline;
- **Espionage as a first-class planned product/navigation area**, with actionable mechanics explicitly deferred to prepared Slice 20; 15.2 exposes no fake Spy actions;
- turn submission/end-turn and bounded built-in-AI driving;
- server-side read/legal-action additions plus a pure non-mutating Planning-preview surface required for live Food/PP/RP, build/research ETA and Population-transfer/Freighter feedback;
- player-safe projections only; no React-owned legality, costs, formulas or hidden AI state.

Defer:

- final visual identity/assets to 15.3;
- Encounter/Invasion/persistence rich presentation to 15.4;
- interactive Tactical battlefield/mechanics to 15.5;
- final complete-game polish and E2E acceptance to 15.6;
- broader New Game/race options to Slice 16;
- broad Ship Designer to Slice 17;
- advanced Diplomacy breadth;
- actionable Espionage mechanics to prepared Slice 20.

## Gate 1 - Functional workflow audit

- [x] Map every strategic command/decision in the canonical match to current App/HTTP surfaces.
- [x] Identify missing player-safe reads, legal actions, target catalogs and command envelopes.
- [x] Define player-safe projection/read requirements for the 2D Galaxy map, star-system dialog, planet bodies, present Fleets/Ships, Colony table and full Colony Detail screen.
- [x] Define the authoritative Colonize confirmation flow and the canonical cross-navigation `Galaxy -> System -> Planet -> Colony Detail` / `Colonies table -> Colony Detail`.
- [x] Define minimum Fleet grouping, eight-category Research and first-class Diplomacy data presentation.
- [x] Audit Espionage evidence/mechanics and decide whether the minimal authoritative baseline fits 15.2 or requires a dedicated sub-slice before actionable UI.
- [x] Define AI-turn driving boundaries and user-visible busy/ready states.
- [x] Define functional browser regression strategy for each command family.
- [x] Present Gate-2 authoritative HMI contract.

Gate-1 evidence: `docs/research/SLICE_15_2_FUNCTIONAL_HMI_GATE1_2026-09-04.md`.

Key Gate-1 findings: reuse the existing player-safe `PlayerDecisionView` as the HMI legality/read foundation; current baseline has no real fog/discovery state; persistent non-colonizable orbital bodies/star-class metadata and a true ordered construction queue are explicit Gate-2 scope decisions; actionable Espionage requires a dedicated later mechanics slice.

## Gate 2 - Functional authority freeze

- [x] Freeze required server read/legal-action additions.
- [x] Freeze browser command submission patterns and optimistic/non-optimistic behavior.
- [x] Freeze strategic page/panel responsibilities inherited from 15.1, including desktop direct Diplomacy/Espionage entries and the compact mobile `More` grouping.
- [x] Freeze 2D Galaxy interaction, star-system dialog, planet selection, Colonize-confirmation and Colony Detail navigation/interaction contracts.
- [x] Freeze Colony table/detail, location-grouped Fleet and eight-category Research functional contracts.
- [x] Freeze the Espionage scope decision: accepted minimal authoritative baseline or explicit dedicated follow-up sub-slice.
- [x] Freeze AI-turn driving and error recovery behavior.
- [x] Freeze minimum functional acceptance scenarios.

Gate-2 evidence: `docs/research/SLICE_15_2_FUNCTIONAL_HMI_GATE2_2026-09-04.md`.

Gate-2 freeze adds persistent gas-giant/asteroid/star metadata and generalized Outpost targets, a real no-visible-limit construction queue, same-row job vs cross-row Population-transfer table gestures, separate transfer source/destination jobs, a pure Planning preview for live economy/build/research timing, sticky BC/Freighters/Command-Point resources, and explicit Espionage deferral to Slice 20.

## Gate 3 - Implementation

- [x] Implement persistent star/orbital-body state plus the interactive 2D Galaxy map and in-map system dialog with planets, gas giants, asteroid belts, Outposts and present Fleet/Ship information.
- [x] Generalize Outposts to legal orbital bodies; implement server-projected Body/Planet actions, Colonize confirmation and post-resolution state refresh/navigation.
- [x] Implement the sortable Colonies management table, same-row job/cross-row Population-transfer gestures, canonical Colony Detail **planet overview with Population growth/next-Pop projection and built Buildings**, pure Planning preview and the real ordered construction queue/editor.
- [x] Implement location-grouped Fleet/Ship tiles and Fleet inspection/movement flows.
- [x] Implement normalized eight-category Research plus live authoritative RP/turn and ETA preview.
- [x] Implement movement/colonization/Outpost and first-class War/Peace Diplomacy workflows; keep Espionage non-actionable and reserved for Slice 20.
- [x] Add the sticky global BC/Freighters/Command-Points strip and integrate end-turn/built-in-AI automation through normal Host authority.
- [x] Add browser/server regressions for critical strategic commands.

## Gate 4 - QA + close

- [ ] Execute the canonical strategic Human-vs-AI workflow without direct API/dev-tool commands, including Galaxy -> star-system inspection, a legal Colonize confirmation where available, Colony table/detail management, Fleet and Research navigation and Diplomacy.
- [ ] Verify browser cannot bypass server legality/authority.
- [ ] Run full Go tests/vet/web build and `git diff --check`.
- [ ] Update evidence/status/HISTORY and close marker.
