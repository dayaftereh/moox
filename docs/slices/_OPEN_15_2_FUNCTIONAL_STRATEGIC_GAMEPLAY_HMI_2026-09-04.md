# OPEN - Slice 15.2 functional strategic gameplay HMI

Opened: **2026-09-04**

Status: **Gates 1-3 complete; Gate 4 blocker remediation in progress before independent QA/closure**.

Primary direction: implement the accepted strategic browser UX on top of server-authoritative player-safe projections: interactive 2D Galaxy, star-system dialog, Colony table + Colony Detail, location-grouped Fleets, eight-category Research, first-class Diplomacy and an explicit Espionage scope decision.

Permanent evidence:
- Product direction: `docs/research/SLICE_15_2_STRATEGIC_HMI_PRODUCT_DIRECTION_2026-09-04.md`
- Gate 1 audit: `docs/research/SLICE_15_2_FUNCTIONAL_HMI_GATE1_2026-09-04.md`
- Gate 2 contract: `docs/research/SLICE_15_2_FUNCTIONAL_HMI_GATE2_2026-09-04.md`
- Gate 3 implementation evidence: `docs/research/SLICE_15_2_FUNCTIONAL_HMI_GATE3_2026-09-05.md`
- Gate 4 blocker remediation: `docs/research/SLICE_15_2_GATE4_BLOCKER_REMEDIATION_2026-09-05.md`

- [x] Gate 1: functional workflow / authority-projection audit.
- [x] Gate 2: freeze functional HMI/server contract.
- [x] Gate 3: implementation.
- [ ] Gate 4: independent QA and close.
  - [x] G4-A: direct Galaxy pan/zoom interaction plus persistent Main/End-Turn actions.
  - [x] G4-B: explicit star-system dialog/overlay plus Colony usability/Population presentation blockers.
  - [x] G4-B2: table-first Colony UI, drag/drop Population, dedicated build route, fullscreen Galaxy/orbital System layout, denser typography and smaller bottom-right End Turn.
    - [x] B2-A: table-first Colony overview + multi-unit desktop/touch Population drag/drop.
    - [x] B2-B: dedicated Colony build-management route/page.
    - [x] B2-C: fullscreen Galaxy/orbital System + denser typography/help + smaller bottom-right End Turn.
    - [x] B2-D: compact resource-first top bar + full-width strategic content + seven-tab mobile nav + bottom-right End Turn; left Main/navigation retained.
    - [x] B2-E: three-dot in-game Main menu + language/settings + Espionage/Advanced relocation + five-tab primary navigation.
    - [x] B2-F: clickable BC/Food/Freighter/Command/Research HUD with signed surplus/deficit states and detail popovers.
  - [ ] G4-C: independent mobile/desktop workflow QA and Slice-15.2 closure.

Live preview:
- preview is run in a separate read-only ASH session on port `7171`;
- restart that preview after each committed remediation block so the embedded web bundle matches HEAD;
- device review remains available over the current LLO Desktop NetBird address.
