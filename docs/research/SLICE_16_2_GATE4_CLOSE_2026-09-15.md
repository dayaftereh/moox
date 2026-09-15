# Slice 16.2 Gate 4 - Difficulty visual selector close

Date: 2026-09-15
Status: **closed / accepted**

## Acceptance decision

The user explicitly accepted the V2 Difficulty threat-sigil visual family and approved closing Gate 4 before any Slice 16.3 work begins.

Slice 16.2 is therefore closed with Gates 1-4 complete.

## Final product result

Difficulty is now a first-class authoritative New Game/server/savegame contract with five stable IDs:

- `easy`
- `normal`
- `hard`
- `very_hard`
- `impossible`

Default/effective legacy value: `normal`.

The frontend uses the shared typed Slice-16.1 selector grammar and loads the authoritative Difficulty catalog from the server. Numeric explanatory facts are formatted from server-owned normalized fields rather than from a duplicated frontend balance table.

Tutor remains outside the main Difficulty ladder and is reserved for a future tutorial/onboarding concept.

## Frozen V1 gameplay effects

Built-in-AI empires receive the frozen additive per-role economy deltas and command-deficit rate from the server-owned profile. Human empires do not receive direct Difficulty economy modifiers.

| Difficulty | Food / farmer | Production / worker | Research / scientist | Tax BC / population | AI CP deficit / excess CP |
| --- | ---: | ---: | ---: | ---: | ---: |
| Easy | -0.25 | -0.50 | -0.50 | -0.25 | 11 BC |
| Normal | 0 | 0 | 0 | 0 | 10 BC |
| Hard | +0.25 | +0.50 | +0.50 | +0.25 | 9 BC |
| Very Hard | +0.50 | +1.00 | +1.00 | +0.375 | 8.5 BC |
| Impossible | +0.75 | +1.50 | +1.50 | +0.50 | 8 BC |

Later AI slices may consume `DifficultyID` for planning/diplomacy/tactical behavior, but they do not own or redefine this V1 server contract.

## Final visual acceptance

The initial command-crest family was replaced during Gate 4 after user feedback that the final/Impossible icon had too many overlays.

Accepted V2 uses minimalist deterministic threat sigils:

- Easy: open orbital arc + core;
- Normal: balanced hex contact + core;
- Hard: triangular threat marker + central diamond;
- Very Hard: razor diamond + inner diamond;
- Impossible: eight-point singularity flare + dark core + small energy center.

The accepted redesign removes the old shield/inner-shield stack, rank chevrons, crown/spike stacks, broken multi-ring overlay and duplicate glow-outline layers.

The five production assets remain deterministic 1200 x 675 SVGs with no embedded localized text. Server ID `very_hard` maps to kebab-case asset option/path `very-hard` while all other IDs map directly.

Detailed redesign evidence:

`docs/research/SLICE_16_2_GATE4_DIFFICULTY_ICON_REDESIGN_V2_2026-09-15.md`

## Gate-4 deterministic equality

Permanent Go coverage proves for every supported Difficulty:

- same seed + same complete settings + same Difficulty produces byte-identical authoritative New Game state;
- the exact selected Difficulty ID is persisted;
- unknown IDs are rejected;
- legacy/fixture states with an omitted Difficulty resolve to Normal without a state-schema bump.

## Gate-4 intended-effect isolation

Permanent regression `TestDifficultyChangesOnlyFrozenInitialStateEffects` compares Normal against every supported Difficulty with the same seed/settings.

After neutralizing only:

- `difficulty_id`;
- built-in-AI colony `Economy`;
- built-in-AI colony `AdjustedEconomy`;
- the direct food-output derivative `PopulationDynamics`;
- the corresponding built-in-AI `FoodLogistics` projection;

the remaining authoritative initial states are byte-identical.

This proves there are no hidden initial-state changes outside the frozen Difficulty contract and its direct normal food/economy derivatives.

Separate command-deficit coverage proves rates 11 / 10 / 9 / 8.5 / 8 BC for built-in AI and preserves the standard 10 BC human rate.

## Desktop/mobile acceptance

The final real-Chrome smoke passes both Galaxy Size and Difficulty selectors with:

- 390 px mobile layout;
- desktop multi-card layout;
- no horizontal overflow;
- all ten deterministic New Game SVG assets;
- mouse interaction;
- real CDP touch interaction;
- ArrowLeft/ArrowRight/Home/End keyboard navigation;
- info-dialog focus/Escape behavior;
- server-derived Difficulty facts;
- all five Difficulty choices supported while Small galaxy remains selected.

The user explicitly accepted the V2 icon review candidate before closure.

## Final regression

The final Gate-4 run passed:

- `go test ./...`;
- `npm run build`;
- UTF-8/mojibake guard;
- deterministic ten-asset selector contract check;
- TypeScript + Vite production build;
- `npm run check:new-game-selector:browser -- http://127.0.0.1:7171`;
- `git diff --check`;
- `/healthz` through NetBird `100.120.252.216:7171`;
- `/healthz` through LAN `192.168.5.27:7171`.

The preview server remains bound to `0.0.0.0:7171` with the explicit non-loopback flag so review is reachable externally.

## Gate-4 checklist result

- [x] Equal seed + equal settings remains byte/deterministically equal.
- [x] Difficulty changes produce only their frozen intended effects.
- [x] Desktop/mobile carousel QA.
- [x] Full tests/build/diff checks.

## Closure protocol

- Slice 16.2 `_OPEN_` marker is removed in the closing commit.
- `docs/slices/HISTORY.md` records the closure.
- `docs/research/ACTIVE_RESEARCH.md` moves to no active slice.
- Slice 16.3 becomes the next prepared objective but is **not opened** during this closure.
