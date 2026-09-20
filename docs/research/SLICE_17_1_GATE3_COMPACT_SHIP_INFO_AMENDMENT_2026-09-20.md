# Slice 17.1 Gate 3 - Compact Ship Information Amendment

Date: 2026-09-20
Slice: 17.1 - Hulls and Mandatory Ship Systems
Gate: 3 HMI amendment
Baseline: `eda5b1c` (`feat(ui): add ship system info popups`)
Session: `ses-20260920T090827-205c7a6fbea5`

## User review goal

Keep the useful Ship-information block below the preview, but make it substantially more compact and player-oriented:

- do not repeat effect prose such as Beam Attack or strategic range in the always-visible rows;
- show only the system category and the concrete installed component;
- keep the trailing `?` as the place for detailed explanation;
- keep Computer and Shield as rich card selectors rather than dropdowns;
- remove technical miniaturization notation such as `M0`, `M1`, etc.;
- replace it with information that tells the player what the research advantage actually does.

## Implemented compact summary

The five mandatory-system rows now show only:

- Drive -> installed Drive;
- Computer -> installed Computer;
- Armor -> installed Armor;
- Shield -> installed Shield or None;
- Fuel -> installed Fuel Cell;

plus the existing trailing `?` detail action.

The header hint and always-visible effect copy were removed from the summary rows.

The rows were tightened to a two-column label/value layout, smaller vertical padding and a narrower help column.

Detailed effects remain available through the `?` dialogs, so no information was lost.

## Player-facing miniaturization information

The UI no longer exposes internal `M0/M1/M2...` terminology.

Instead, Computer cards display:

`current PP · build-cost reduction`

Examples:

- no reduction: `5 PP · Basiswerte`;
- saturated late-game reduction: `1 PP · -75 % Baukosten`.

Shield cards display:

`current PP · current Space · build-cost reduction · Space reduction`

Examples are derived from the exact original miniaturization factors frozen in Gate 2.

## Server-authoritative projection

The percentage labels are not re-derived from research rules in TypeScript.

The authoritative Designer catalog now projects:

- `cost_reduction_percent`;
- `space_reduction_percent`.

Those values are computed in Go from the same factor helpers used by the authoritative miniaturized cost/Space compiler.

This keeps the frontend display aligned with gameplay authority and avoids duplicating miniaturization rules in the HMI.

The existing internal `miniaturization_level` remains available for deterministic/debug/test purposes, but is no longer presented as player terminology.

## Regression coverage

Focused Go coverage now verifies:

- Mid-Tech Electronic Computer reports 60% cost reduction at mini level 3;
- Mid-Tech Class I Shield reports 45% cost reduction and 35% Space reduction at mini level 2;
- All-Tech saturated Electronic Computer/Class I Shield projections report 75% reductions as appropriate.

Browser acceptance verifies at 390px:

- five compact system rows;
- no verbose effect copy in the always-visible summary;
- no `M<number>` or `Miniaturisierung` notation in summary or Computer card choices;
- Electronic Computer choice still shows its actual gameplay effect;
- the same choice also shows the player-facing `-75 % Baukosten` reduction;
- all existing Ship Designer, save, no-Shield, Development Tools and buyout workflows remain functional.

## Verification

- `npm run build` - PASS
- `go test ./internal/game -run "Slice171CatalogProfilesAndLateHullLocks|Slice171MiniaturizationGoldenVectors" -count=1` - PASS
- `go test ./...` - PASS
- `go vet ./...` - PASS
- fresh Reference server browser smoke on port 7193 - PASS
- `git diff --check` - PASS apart from expected Git line-ending conversion warnings

## Scope

No gameplay balance or original-rule behavior changed.

This amendment changes only:

- player-facing projection metadata;
- Ship Designer information density/wording/layout;
- browser and Go assertions for that projection.

Gate 4 remains not started.
Slice 17.2+ remain closed.
