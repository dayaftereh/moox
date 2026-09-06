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
    - [x] B2-G: remove Research primary tab; HUD opens MOO2-inspired authoritative research-selection overlay and returns to Galaxy after selection.
    - [x] B2-H: explicit choose-one technology buttons + selected technology emphasis + visible selection-mode explanation.
    - [x] B2-I: selected research technology uses immediate draft state + unmistakable bold/check/highlight styling.
    - [x] B2-J: relaxed top-HUD spacing + connection status integrated into the three-dot menu frame.
    - [ ] B2-K: Galaxy pan bounds + MOO2-inspired framing review.
      - [x] Pan bounds with responsive edge offset on mouse/touch.
      - [x] Map-only Galaxy surface; remove visible zoom/help/title controls while preserving wheel/pinch zoom.
      - [x] Restore Espionage as fifth primary tab; shorten turn-submit action to Fertig/Done and remove duplicate Espionage from the game menu.
      - [ ] Review supplied MOO2 Galaxy reference and refine framing/density.
    - [x] B2-L: MOO2-inspired star-system dialog with clickable bodies, authoritative planet inspection and legal selected-body actions; special-resource markers deferred until core metadata exists.
    - [x] B2-M: true circular system orbits + clickable system-level fleet/ship inspection; planet-local fleet anchoring deferred until core exposes an authoritative body location.
    - [x] B2-N: keep the orbital picture fleet-free; expose present fleets/special vessels through a contextual Fleet/Ships dialog with per-Ship technical loadout drill-down; persistent partial damage remains a core gap.
    - [x] B2-O: clicking a colonized planet opens its Colony detail directly; uncolonized/non-planet bodies remain inspectable in the System dialog.
    - [x] B2-P: MOO2-inspired responsive construction catalog/detail/queue workspace + confirmed current-build cancellation with preserved authoritative PP; future Ship Designer affordance reserved but disabled until Slice 17.
    - [x] B2-Q: MOO2-inspired responsive Colony command hierarchy plus server-derived player-relative PlanetPotential for uncolonized System inspection and understandable climate/size/mineral/gravity effects.
    - [x] B2-R: compact Colony toolbar + deduplicated BC/population/output presentation + dense planet-trait/job tables, preserving authoritative drag/drop and PlanetPotential.
    - [x] B2-S: split Colony Detail into compact Planet Profile + tap/click Planet Info popover and a separate always-visible Colony Profile below.
    - [x] B2-T: merge Growth + next-Pop ETA into one signed row; apply shared green-positive/red-negative tones and document the existing core-Morale / missing player-safe Morale-projection boundary.
    - [x] B2-U: compact Build-management toolbar + show current construction once with PP/%/ETA and only future projects in the Queue; lock queue reordering from silently replacing the current build.
  - [ ] G4-C: independent mobile/desktop workflow QA and Slice-15.2 closure.

Live preview:
- preview is run in a separate read-only ASH session on port `7171`;
- restart that preview after each committed remediation block so the embedded web bundle matches HEAD;
- device review remains available over the current LLO Desktop NetBird address.
