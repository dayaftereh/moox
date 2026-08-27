# Original research breakthrough mechanic - 2026-08-27

Status: verified directly against the private Master of Orion II 1.31 `Orion2.exe` and implemented in the deterministic Master of Orion X runtime for standard technology fields.

## Naming convention

Research events carry both original numeric IDs and stable speaking internal keys. Example:

```text
TechField 56
Technology 155 (research_laboratory, Research Laboratory)
Building research_laboratory (Production ID 35)
```

The numeric IDs remain the clean-room fidelity anchor; the speaking key is used by logs, Observer/debug output and future Human/AI APIs. User-visible names continue to come from language keys.

## Original breakthrough chance

The original integer chance helper is at code VA `0xE1EC6` in MOO2 1.31. Its behavior is:

```text
if base_cost <= 0 or projected_rp <= base_cost:
    chance = 0
else:
    chance = floor((projected_rp - base_cost) * 100 / base_cost)
    chance = min(chance, 100)
    if chance == 0:
        chance = 1
```

Examples:

| Base cost | Projected RP | Chance |
| ---: | ---: | ---: |
| 50 | 50 | 0% |
| 50 | 51 | 2% |
| 150 | 151 | 1% |
| 150 | 225 | 50% |
| 150 | 300 | 100% |
| 150 | 450 | 100% |

The explicit minimum-1 branch matters for expensive fields where a small positive excess would otherwise truncate to zero.

## Inputs to the chance

The field-cost helper at VA `0xE1E96` reads the active TechField's original base RP cost from the 23-byte technology-field table.

The research chance caller at VA `0xE1EF4` supplies:

```text
projected_rp = accumulated_research + empire_research_this_turn
```

This confirms that the displayed/rolled breakthrough chance already includes the current turn's research before that research is committed to the accumulated total.

For standard fields this uses the normalized base RP cost. Original hyper-advanced fields (`TechField >= 75`) add player-specific repeated-field cost scaling. That state is not modeled yet, so automatic breakthrough resolution explicitly rejects those fields instead of silently using an incorrect static cost.

## Original RNG comparison

The original `random(N)` helper is at VA `0x1247A0`. For `N=100` it returns a uniform integer in the inclusive range:

```text
1..100
```

The research turn resolver at VA `0xE44E0` succeeds when:

```text
roll <= chance_percent
```

Therefore:

- 0% can never break through,
- 1% is exactly one successful roll out of 100,
- 100% is guaranteed,
- the 100% case still consumes the random roll.

The original executable uses its own LCG. Master of Orion X intentionally keeps the simulation-owned SplitMix64 RNG defined by the deterministic core; `Intn(100)+1` reproduces the 1..100 distribution and comparison semantics without claiming original RNG-stream identity.

## RNG consumption order

A subtle but important replay rule is proven by the original turn resolver: every empire with an active nonzero research field consumes one `random(100)` call each research turn, even when the calculated chance is 0%.

The runtime mirrors this. Empires with no active `ResearchState` consume no research RNG. Multiple active empires are processed in stable ascending Empire ID order, keeping multiplayer/replay results independent from slice/network timing.

## MOOX numeric architecture

The original executable uses integer Colony and Empire research storage. That fact remains part of the clean-room evidence above, but it is no longer copied as a MOOX storage constraint.

Per `docs/architecture/ADR-0002-research-float64.md`, MOOX stores active research as RP-native `float64`:

```text
ResearchState.ProgressRP = 153.75
```

The Colony economy now also uses domain-native `float64` values under ADR-0003. Empire research therefore sums Colony RP directly without any milli-unit conversion or per-Colony truncation:

```text
1.9 RP + 1.9 RP = 3.8 RP
```

This intentionally diverges from the old integer storage behavior. It avoids losing fractional bonuses merely because the 1996 executable used narrow integer fields.

The verified breakthrough curve remains useful as a gameplay rule. MOOX computes it from the full floating-point projected RP and floors only the final percentage used by the discrete 1..100 roll. Therefore rounding occurs at the named probability boundary, not during RP accumulation.

`GameState.Validate` rejects negative, NaN and infinite research progress. Persistent state schema 4 serializes the value as numeric `progress_rp`; schema 4 additionally materializes Population sustenance/growth dynamics while retaining the same domain-native numeric policy.
## Turn resolution and overflow

The original turn resolver at VA `0xE44E0` performs this order:

```text
chance = breakthrough_chance(accumulated + current_turn_rp)
accumulated += current_turn_rp
roll = random(100)  // 1..100, even when chance == 0
if roll <= chance:
    breakthrough = true
    accumulated = 0
    complete_technology_field()
```

The `accumulated = 0` assignment occurs before the original field-completion routine. Classic research overflow is therefore discarded on breakthrough rather than transferred into the next field.

The runtime mirrors this by clearing the completed `ResearchState`; no RP carry is created.

## Runtime events / Observer

Every active research turn emits authoritative `empire.research_progressed` data containing:

- Empire ID,
- TechField ID,
- Technology IDs,
- speaking Technology keys,
- base cost RP,
- previous accumulated RP,
- current-turn research RP,
- projected RP,
- breakthrough chance percent,
- actual 1..100 roll,
- breakthrough result.

On success, `empire.research_completed` follows and carries both numeric Technology IDs and speaking keys.

For the current regression example:

```text
TechField 56
Technology 155 (research_laboratory)
```

completes and immediately makes Building `research_laboratory` (Production ID 35) available through the existing authority-filtered `BuildingChoices` projection.

## Determinism tests

The checkpoint verifies:

- the exact integer chance curve,
- minimum 1% after positive excess/rounding,
- guaranteed 100% at double base cost,
- one RNG draw at active 0% chance,
- no RNG draw without active research,
- fractional Colony RP preservation through deterministic Empire aggregation,
- speaking Technology key propagation,
- Technology 155 -> Building research_laboratory unlock,
- identical state/event/RNG result from identical seed + state,
- authoritative GameSession/Observer end-to-end breakthrough.

## Deliberately deferred

This checkpoint does not yet implement or claim:

- original research-target/TechField selection UI and command semantics,
- Creative/Uncreative selection rules,
- automatic selection of the next research field,
- leader/global research sources not yet present in the economy model,
- hyper-advanced repeated-field cost state/scaling,
- Advanced-start randomized/race-aware extra fields.

That legal-action slice is now implemented in `docs/research/RESEARCH_SELECTION_2026-08-27.md`: Human UI and AI consume the same `ResearchChoices`, and `empire.select_research` sends only the selected TechField ID.
