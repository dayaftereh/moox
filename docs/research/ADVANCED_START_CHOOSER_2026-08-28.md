# Advanced Start chooser - direct MOO2 1.31 evidence (2026-08-28)

## Scope

This checkpoint isolates the original default-game `Choose_Tech_Application_` selection path used by Advanced New Game technology grants. It is clean-room behavioral evidence for MOOX. It does not require MOOX to reproduce original storage layouts or the original RNG implementation.

Reference executable: private `Orion2.exe` 1.31 installation used by the existing research corpus.

Key functions:

- `All_Gray_Boxes_Researched_` - VA `0xFD2F9`
- `Choose_Tech_Application_` - VA `0xFD335`
- `Calc_Tech_Value_` - VA `0xFC845`
- `Set_Competition_Tech_Values_` - VA `0xFD219`
- `Get_Weighted_Choice_Int_` - VA `0xFE96F`
- `Random_` - VA `0x1247A0`
- fallback `Choose_Hyper_Advanced_Tech_` - VA `0xFC734`

## Candidate scan

`Choose_Tech_Application_` scans Technology/application indexes `1..211` in ascending order. Normal Advanced candidates pass all of the following gates:

1. application status is `1` (researchable);
2. Strategic Combat availability is satisfied when that game mode is enabled;
3. the application's TechField is `< 75`;
4. the owning TechField status is `2` (open research frontier).

For each legal candidate it calls `Calc_Tech_Value_` and stores an integer raw weight.

After the six Average-start fields are already researched, `All_Gray_Boxes_Researched_` is true and the chooser applies the verified start bonuses:

- TechField `4`: weight x2;
- Technology `114`: weight x2;
- Technology `51`: weight x5;
- TechField `73`: weight x2;
- if the resulting raw weight is zero, promote it to `1`.

The alternative not-all-gray-boxes path is not the Advanced-start path and remains separate from this checkpoint.

## Cost-tier admission

The original chooser uses two 212-entry integer arrays: raw weights and tier-scaled weights.

The cost divisor read by this path is normalized from zero to one. During New Game initialization the observed divisor is zero, therefore Advanced-start grants use divisor `1`.

The threshold starts at `15`. For each candidate with positive raw weight:

```text
normalized_cost = max(1, floor(field_cost_rp / divisor))

if normalized_cost <= threshold:
    scaled_weight = floor(raw_weight * threshold / normalized_cost)
else:
    scaled_weight = 0
```

If no positive candidate is admitted, the threshold is replaced by:

```text
floor(3 * threshold / 2)
```

so the sequence begins:

```text
15, 22, 33, 49, 73, 109, ...
```

The threshold expansion itself consumes no RNG.

## Default-game secondary filter

After tier scaling the original has a second filter guarded by global byte `0x21CB0`. Existing executable analysis proves the default game setting is `0`, so the first MOOX Advanced-start implementation must bypass this layer exactly as the original default path does.

The non-default `0x21CB0 > 0` path remains deliberately deferred until that game setting has a semantic public option in MOOX.

## Weighted selection and RNG consumption

When at least one scaled weight is positive, `Choose_Tech_Application_` calls `Get_Weighted_Choice_Int_` with the 212-entry scaled-weight array.

`Get_Weighted_Choice_Int_` performs:

```text
total = sum(weights[0..211])
if total == 0:
    return -1

roll = Random(total)  // 1..total inclusive
for index in 0..211:
    if roll <= weights[index]:
        return index
    roll -= weights[index]
```

Because index zero has no normal candidate weight and Technology IDs are scanned into their matching array indexes, the returned weighted-choice index is the selected Technology ID.

For MOOX this means the deterministic semantic contract is:

1. calculate all integer weights without advancing RNG;
2. sum final scaled weights in Technology-ID order;
3. consume exactly one caller-owned New Game RNG draw when total weight is positive;
4. choose the first Technology bucket containing that one-based roll.

MOOX should use its own serializable RNG (`rng.Intn(total) + 1`) rather than claim original RNG identity.

## Zero-weight fallback

If the final scaled maximum/weight set is empty, the original falls back to `Choose_Hyper_Advanced_Tech_` at VA `0xFC734`.

This fallback exists in the general chooser, but Advanced New Game's 19 extra grants are expected to remain on the ordinary `<75` frontier. MOOX should keep the fallback explicit rather than silently inventing a normal Technology when no candidate exists.

## Implementation contract for MOOX

The default-path Advanced chooser can therefore be implemented as a pure deterministic weight build plus one authoritative RNG choice:

```text
candidates = legal_open_normal_technologies(empire)
raw[id] = Calc_Tech_Value_default(...)
raw[id] = apply_advanced_start_bonuses(raw[id], id, field)

threshold = 15
repeat:
    scaled = tier_scale(raw, threshold, divisor=1)
    if any(scaled > 0):
        break
    threshold = floor(3 * threshold / 2)

selected = weighted_choice_one_rng_draw(scaled, shared_NewGameRNG)
```

The full-state initializer must continue processing Empires in stable original-equivalent order because `Set_Competition_Tech_Values_` lets later Empires observe technologies already granted to earlier Empires.

## Verification targets

Tests for the runtime port should lock at least:

- threshold sequence and integer truncation;
- no RNG draw while thresholds expand;
- one RNG draw for a positive weighted choice;
- Technology-ID-stable bucket ordering;
- deterministic same-state/same-seed result;
- different earlier-Empire ownership can change a later Empire's competition-derived weights;
- Strategic Combat filtering occurs before weighted selection;
- exactly 19 Advanced extra grant iterations;
- ordinary, Creative and Uncreative acquisition semantics remain server-owned;
- no client-selected Technology ownership enters New Game generation.
