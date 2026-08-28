# Uncreative external-acquisition repair - direct MOO2 1.31 evidence (2026-08-28)

## Scope

This checkpoint isolates the original repair path used when an Uncreative empire acquires a Technology application outside the normal fixed research choice. It extends, but does not change, the already-verified New Game Uncreative fixed-plan generation.

Relevant original functions:

- `Ensure_Uncreative_Field_OK_` - VA `0xE408F`
- `Player_Gets_Tech_App_` - VA `0xE4204`
- `Random_` - VA `0x1247A0`

Direct xref evidence shows `Player_Gets_Tech_App_` is the direct caller. After the acquired application is marked owned, Uncreative players run `Ensure_Uncreative_Field_OK_` for the application's TechField.

## Field scan

The helper walks the TechField's four application slots in original slot order.

For every concrete application:

- status `1`: count it as an already-available Uncreative research application;
- status other than `0`/`1`: ignore it;
- status `0`: consider it as a possible replacement only if all eligibility gates pass.

A replacement is committed only when **no status-1 application remains** in the field.

This means external acquisition of a non-fixed application does not replace the fixed application if that fixed application is still researchable.

## Replacement eligibility

A state-0 application is eligible for the repair reservoir when:

1. Technology 52 is allowed only when Random Events are enabled;
2. its TechField is not special government-evolution TechField `6`;
3. Strategic Combat availability is satisfied when Strategic Combat is enabled.

### Dimensional Portal

Technology `52` (`dimensional_portal`) is rejected when original global `0x21CAF == 0`.

Direct New Game UI and event-system evidence now identifies original global `0x21CAF` as **Random Events enabled**. The repair therefore excludes Technology 52 when `RandomEventsDisabled` is true. See `RANDOM_EVENTS_TECH_GATE_2026-08-28.md`.

### Government evolution field 6

TechField `6` contains:

- Technology 42 `confederation`;
- Technology 65 `federation`;
- Technology 77 `galactic_unification`;
- Technology 92 `imperium`.

`Ensure_Uncreative_Field_OK_` explicitly skips replacement candidates when the target TechField ID is `6`. Therefore external acquisition must not randomly repair an Uncreative fixed choice by switching to another government-evolution family.

## Reservoir-sampling RNG semantics

The original does **not** collect eligible candidates and perform one final random choice.

Instead it uses reservoir sampling while scanning application slots:

```text
replacement = none
eligible_count = 0
available_count = 0

for application in field_slot_order:
    if status(application) == 1:
        available_count++
        continue

    if status(application) != 0:
        continue

    if !replacement_eligible(application):
        continue

    eligible_count++
    if Random(eligible_count) == 1:
        replacement = application

if available_count == 0 and eligible_count > 0:
    status(replacement) = 1
```

Consequences for deterministic MOOX behavior:

- one RNG draw is consumed for **each eligible state-0 replacement candidate**;
- with one eligible candidate the draw is still consumed (`Random(1)`);
- a later candidate replaces the reservoir with probability `1/N` at candidate count `N`;
- the final choice is uniform across eligible candidates;
- eligible candidates can still consume RNG even when an existing status-1 application means no replacement is ultimately committed.

MOOX should preserve this draw pattern with its own caller-owned serializable RNG rather than claim original RNG identity.

## Semantic MOOX state mapping

MOOX does not persist original per-Technology status bytes. The equivalent semantic operation is on:

- `Empire.KnownTechnologyIDs`;
- `Empire.UncreativeResearchChoices`;
- the normalized TechField -> Technology membership;
- normalized Strategic Combat availability;
- semantic Random Events configuration for the Dimensional Portal gate.

A future external-Technology acquisition transition should therefore perform:

```text
1. grant the acquired Technology authoritatively;
2. if empire is not Uncreative: done;
3. locate the acquired Technology's TechField;
4. if the field is already completed/irrelevant: no legal-action repair is required;
5. inspect the current persisted fixed choice for that field;
6. if the fixed choice is still unknown and eligible: keep it;
7. otherwise reservoir-sample an eligible unknown replacement in original field-slot order;
8. update only that field's persisted FixedResearchChoice;
9. expose the repaired fixed_one legal action through the normal server ResearchChoices surface.
```

No client chooses the replacement Technology.

## Important distinction from New Game generation

Initial Uncreative fixed choices are generated during New Game initialization from the shared New Game RNG across TechFields `1..73`.

`Ensure_Uncreative_Field_OK_` is a **later repair path**. It must not be used to regenerate the complete initial fixed plan and must not introduce a separate private RNG stream.

## Implementation boundary

The clean integration point is the authoritative external-Technology grant operation used by future trade, espionage, conquest or scripted acquisition. The repair lives in the game/session domain layer, not in Wails, network, MCP or UI code.

MOOX now has that low-level external-acquisition operation. It remains server-owned and is not exposed as a fabricated client command.

## Runtime application-slot ordering

Direct `Init_Tech_` analysis at VA `0x5E1E3` resolves the remaining ordering question. The static TechField application slots start empty. `Init_Tech_` scans Technology/application IDs in ascending order and writes each application into the first empty slot of its TechField. Special TechField 74 is skipped by this normal slot-building path.

Therefore, for ordinary TechFields `1..73`, MOOX `TechnologyIDsByField` ascending-ID order is already the original runtime application-slot order used by `Ensure_Uncreative_Field_OK_`. No additional legacy slot array is needed in authoritative state or rules.

## Current MOOX game-layer support

MOOX now provides `EconomyRules.RepairUncreativeResearchChoiceAfterAcquisition`. It expects the acquired Technology to already be in `Empire.KnownTechnologyIDs`, consumes the caller-owned RNG with the verified reservoir pattern, persists a replacement `FixedResearchChoice` when one exists, removes the fixed choice when no legal replacement remains, and leaves TechField 6 without random government-family repair.

`ResearchChoices` treats a legitimately missing Uncreative fixed choice as an unselectable field instead of a projection error. Technology 52 now follows the resolved `RandomEventsDisabled` domain option instead of an opaque `0x21CAF` placeholder.

`EconomyResolver.GrantTechnology` and `GameSession.GrantTechnology` now provide the authoritative grant transition. The transition adds only the concrete Technology, does not invent TechField completion, commits any Uncreative repair RNG consumption through `GameState`, and emits `empire.technology_granted` into the strategic Observer/replay stream. Future trade/espionage/conquest systems can reuse this path without mutating Technology ownership directly.

## Verification targets

Runtime tests should lock:

- fixed choice unchanged when it remains unknown/researchable;
- replacement occurs when the fixed Technology becomes externally known;
- replacement never selects an already-known Technology;
- TechField 6 performs no random government-family repair;
- Strategic Combat-ineligible applications are excluded;
- one RNG draw per eligible replacement candidate in slot order;
- deterministic same-state/same-seed repair;
- no replacement when no eligible candidate exists;
- `ResearchChoices` immediately reflects the repaired `fixed_one` choice;
- the client never supplies replacement Technology IDs.
