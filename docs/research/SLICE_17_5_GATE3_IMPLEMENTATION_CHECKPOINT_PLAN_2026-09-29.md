# Slice 17.5 Gate 3 - Implementation Checkpoint Plan

Date: 2026-09-29
Status: **Checkpoints 1-2 complete; Checkpoint 3 ready / not started**
Base commit: `4939a04 docs(slice): freeze 17.5 gate 2 tactical foundation`
Gate 2 contract: `docs/research/SLICE_17_5_GATE2_TACTICAL_FOUNDATION_CONTRACT_FREEZE_2026-09-29.md`
OPEN slice: `docs/slices/_OPEN_TACTICAL_COMBAT_FOUNDATION_SHIP_COMBATANT_BREADTH_2026-09-28.md`

## Execution policy

Gate 3 is deliberately split into five bounded checkpoints so that implementation, review and chat/session handoff stay recoverable.

After **every** checkpoint:

1. run the checkpoint-specific focused tests;
2. run `git diff --check`;
3. review the complete diff for that checkpoint;
4. commit only the coherent checkpoint block;
5. verify `git status --short --branch` is clean;
6. persist an ASH session checkpoint naming the completed checkpoint, HEAD, tests and exact next step;
7. report back to the user and **stop**. Do not start the next checkpoint until the user explicitly says to continue.

No push unless explicitly requested. Slice 17.5 remains OPEN until Gate 4 acceptance/close work is explicitly completed.

## Checkpoint 1 - Tactical Ruleset and Ship Projection

Status: **COMPLETE**

Scope:

- extend the authoritative Tactical ruleset with the frozen Hull, Armor, Computer, Drive, Shield and Fuel foundation rows;
- derive truthful `StructureMax`, `ArmorMax`, minimum/current combat speed, Beam Offense, Beam Defense, Shield Strength and Shield Arc Max;
- extend immutable `TacticalShipSpec` / projection with the frozen provenance and derived fields;
- preserve current admission guards during this checkpoint;
- do not implement full Beam, ordnance, fighter or combat-Special family behavior.

Acceptance:

- Early/Mid/Late representative vectors across all six hull classes pass;
- existing Slice-15.5 Tactical baseline remains regression-green;
- no old admission guard is removed yet;
- working tree is clean after a dedicated checkpoint commit.

## Checkpoint 2 - Broad Combatant Admission

Status: **COMPLETE**

Scope:

- implement the frozen structural/derived combatant validator;
- replace `baselineCombatantUnsupportedReason` and `validateBaselineCombatant` together;
- admit legal Slice-17.1 Hull/Drive/Computer/Armor/Shield/Fuel combinations without hidden baseline substitution;
- keep malformed/unknown identities and malformed weapon slots as encounter-level errors;
- allow legal ships carrying later-family weapons whose Tactical consumer is not implemented.

Acceptance:

- Frigate through Doom Star with accepted mandatory systems can enter Tactical;
- no legal admitted combatant silently receives Frigate/Electronic/Titanium/no-Shield/baseline-drive/zero-defense substitutions;
- unsupported later-family weapon presence alone never rejects the encounter;
- regression tests remain green and the checkpoint is committed cleanly.

## Checkpoint 3 - Shields and Shared Damage Foundation

Status: **READY / NOT STARTED**

Scope:

- add four mutable Shield-facing pools;
- implement authoritative facing selection;
- implement Shield per-hit block and per-facing capacity;
- implement shared Shield -> Armor -> Structure damage application;
- implement deterministic Shield recharge in stable arc order;
- extend common damage events with authoritative Shield/Armor/Structure before/after data.

Acceptance:

- frozen no-Shield and Class I/III/V/VII/X behavior is covered;
- Gate-2 damage vectors pass, including Class III Frigate/Cruiser examples and deterministic recharge;
- destruction and overflow are deterministic;
- existing Laser damage regressions remain green;
- checkpoint is committed cleanly.

## Checkpoint 4 - Generic Weapon Action Availability and UI Consumption

Status: **PENDING CHECKPOINT 3**

Scope:

- project one authoritative action-availability descriptor per installed weapon slot;
- preserve supported current Laser/Beam command behavior;
- expose unsupported later-family slots as visible but unavailable with `tactical_consumer_unavailable`;
- guarantee rejected/unavailable actions consume zero RNG and zero readiness;
- expose authoritative target lists only for available actions;
- update Tactical UI to consume server-projected availability/targets without client gameplay formulas.

Acceptance:

- mixed supported/unsupported weapon designs enter and render in Tactical;
- supported slots stay usable;
- unsupported slots are disabled explicitly and cannot become client pseudo-actions;
- browser does not derive hit/damage/readiness/target legality;
- checkpoint is committed cleanly.

## Checkpoint 5 - Integration, Regression and Gate 3 Evidence

Status: **PENDING CHECKPOINT 4**

Scope:

- cover all frozen Early/Mid/Late matrix vectors;
- cover Shield/damage/recharge and unsupported-action vectors;
- verify Observer/replay/public projections and persistence boundaries affected by the foundation;
- run full Go and web regression suites plus browser smoke where applicable;
- write Gate-3 implementation evidence;
- update live Slice status to Gate 3 complete / Gate 4 review ready only after all evidence is green.

Minimum final QA:

- focused Slice-17.5 tests;
- `go test ./... -count=1`;
- `go vet ./...`;
- `npm --prefix web run build`;
- relevant Reference-Lab browser smoke;
- `git diff --check`;
- final Git/status/open-marker verification.

Acceptance:

- implementation matches the Gate-2 frozen contract;
- no ownership leakage into full 17.6 Beam, 17.7 Ordnance or 17.8 Fighter/Combat-Special behavior;
- complete Gate-3 evidence exists;
- working tree is clean after a dedicated Gate-3 evidence commit;
- report back and stop for user review before any later gate/close action.

## Current resume point

Resume at **Checkpoint 3 - Shields and Shared Damage Foundation** only after explicit user continuation. Checkpoints 1-2 are complete and committed; verify the current clean HEAD with `git log -1 --oneline` before continuing.

Do not start Checkpoint 3 until the user explicitly says to continue.
