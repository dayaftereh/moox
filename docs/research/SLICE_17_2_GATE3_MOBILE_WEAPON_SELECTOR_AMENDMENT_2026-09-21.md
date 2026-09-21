# Slice 17.2 Gate 3 - Mobile Weapon Selector HMI Amendment

Date: 2026-09-21

Result: **PASS**

## User feedback

The first Gate-3 Weapon Selector implementation was functionally correct but still too desktop-oriented on a phone:

- the separate "Available Weapons" card duplicated the selector entry point and consumed vertical space;
- the weapon filters were always visible and compressed the actual weapon list;
- the configuration area was too tall;
- the lower Install action could fall outside the immediately visible mobile area;
- the empty Installed Weapons area did not itself act as the obvious Add target.

## Main Designer change

The separate Available Weapons section was removed.

The Installed Weapons panel now owns the complete weapon entry point:

- "Space Available" is shown in the Installed Weapons heading;
- a compact `+` remains available in the heading whenever another mount can be added;
- when no weapon is installed, the entire empty Installed Weapons surface is a large Add-weapon target;
- existing installed rows keep the compact `- count +` and remove controls;
- the reserved mobile Installed Weapons panel is reduced from 164 px to 150 px.

This keeps the remaining-space information next to the weapon workflow while removing a redundant card from the main Ship Designer.

## Filter behavior

Weapon Selector filters are now collapsed by default.

A compact Filter button in the dialog header toggles:

- search;
- availability;
- modifier support;
- sort key/direction.

The weapon list therefore receives the available height first instead of losing space to controls that are only occasionally needed.

## Mobile selector layout

At <=700 px the Weapon Selector now uses an almost full-viewport compact layout.

Before a weapon is selected:

- the configuration pane is hidden;
- the weapon list receives the full body height.

After selection:

- the upper portion remains a scrollable weapon list;
- the lower portion contains the selected weapon configuration;
- configuration content has its own scroll region;
- the Install action is outside that scroll region and remains visible at the bottom.

Additional mobile compaction:

- smaller header, close/filter controls and dialog padding;
- 40 px weapon rows;
- reduced gaps and typography;
- three-column mount-choice row where possible;
- two-column special-modifier grid;
- compact quantity and preview rows;
- footer Close button hidden on mobile because the header already contains Close;
- Install expands to the available footer width.

## 390x844 regression measurements

The updated reference-lab browser smoke uses the existing 390x844 emulated phone viewport and now verifies:

- no separate Available Weapons card exists;
- the empty Installed Weapons surface opens the Weapon Selector;
- filters are absent initially and the Filter toggle reports `aria-expanded=false`;
- before weapon selection the list height is at least 500 px; current deterministic measurement is about 780 px;
- the configuration pane is not displayed before selection;
- opening Filter reveals search plus the three selector controls;
- all 13 Beam rows remain available;
- Standard / Heavy / Point Defense and special modifiers remain selectable;
- after selecting a weapon the Install button is fully inside the 844 px viewport without scrolling;
- Standard and Heavy Laser installation still succeeds;
- installed rows remain height-stable, compact and scrollable.

## QA

Final amendment verification:

- `go test ./... -count=1 -timeout 120s` - **PASS**
- `go vet ./...` - **PASS**
- `npm run build` - **PASS**
- `node --check scripts/check-reference-lab-browser.mjs` - **PASS**
- `npm run check:reference-lab:browser -- http://127.0.0.1:7201` - **PASS**
- `git diff --check` - **PASS** (line-ending warning only)

Gate 4 remains **not started**. Slice 17.3+ remains closed.
