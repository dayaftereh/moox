# Original MOO2 1.31 Hyper-Advanced research - 2026-08-28

Status: the original repeated Hyper-Advanced field state, dynamic strategic research cost, completion increment and repeat behavior are directly verified from `Orion2.exe` 1.31 and implemented in MOOX.

## Reference binary

```text
C:\ASH\Temp\mastori2\Orion2.exe
SHA-256 7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5
```

The established bound-LE mapping remains:

```text
code object:      object 1
object-1 base VA: 0x10000
```

## Hyper-Advanced field set

The normalized original TechField table contains eight repeatable fields:

```text
TechField 75..82
```

All eight have:

```text
static table research cost = 15000 RP
next_id                    = 0
concrete Technology IDs    = none
```

Their predecessors are:

```text
75 <- 70
76 <- 40
77 <- 69
78 <- 58
79 <- 68
80 <- 48
81 <- 49
82 <- 32
```

Direct raw field-record inspection confirms that fields 75..82 have four zero application slots. They are therefore repeatable fields, not ordinary fields containing hidden concrete Technology records.

## Persistent original counters

Relevant helpers:

```text
Save_Hyper_Tech_Values_             VA 0x10F884
Restore_Hyper_Tech_Values_          VA 0x10F8B7
Promote_Hyper_Field_Values_         VA 0x10F8ED
Promote_A_Hyper_Value_By_Field_     VA 0x10F919
```

The original player structure stores eight one-byte counters at:

```text
player + 0x21C .. player + 0x223
```

The field-addressed form used by original code is:

```text
player + TechFieldID + 0x1D1
```

For field 75:

```text
75 + 0x1D1 = 0x21C
```

and field 82 maps to `0x223`.

A real original `SAVE10.GAM` was inspected using the already-established player-array layout. Its active players have all eight counters at zero, confirming zero as the persisted initial completed-level state.

## Authoritative original research cost

`Player_Research_Cost_` at VA `0xE1E96` reads the normal TechField table cost and, for field IDs >=75, adds the per-player counter:

```asm
cmp   dx, 0x4B
...
movzx edx, BYTE PTR [player + fieldID + 0x1D1]
imul  edx, edx, 0x2710
add   eax, edx
```

`0x2710` is decimal 10000.

Therefore the authoritative strategic cost rule is:

```text
research_cost_rp = static_field_cost_rp
                 + completed_levels * 10000
```

For the normalized 1.31 fields:

```text
completed 0 -> 15000 RP
completed 1 -> 25000 RP
completed 2 -> 35000 RP
completed 3 -> 45000 RP
...
```

MOOX stores the semantic `completed_levels` value as an integer. It does not reproduce the original one-byte storage width.

## Completion behavior

`Give_Player_Field_` at VA `0xE4410` has a dedicated Hyper-Advanced branch:

```asm
cmp fieldID, 0x4B
jl  normal_field_path
inc BYTE PTR [player + fieldID + 0x1D1]
...
cmp fieldID, 0x4B
jge return_without_normal_application_grant
```

So a Hyper-Advanced breakthrough:

1. increments only that field's completed-level counter;
2. does not grant a concrete Technology ID;
3. does not run normal field/application ownership acquisition.

`Check_For_Research_Breakthrough_` still follows the already-verified standard breakthrough mechanics:

- current RP is added to accumulated RP;
- one deterministic 1..100 roll is consumed for an active project;
- overflow is discarded on success;
- accumulated RP is reset to zero.

The Hyper-specific difference is the repeated-level increment instead of a normal field/Technology ownership transition.

## The active Hyper project remains repeatable

Normal field completion eventually clears the player's active field through the ordinary application/advance path.

The Hyper branch in `Give_Player_Field_` bypasses that path. The active field is not cleared by the strategic completion code; after breakthrough the same Hyper field therefore remains the active repeated research project while its RP progress returns to zero.

MOOX mirrors this:

```text
Hyper breakthrough
-> completed_levels++
-> ProgressRP = 0
-> ResearchState remains active on the same TechField
-> next turn uses the increased dynamic cost
```

The Hyper field is deliberately **not** added to `KnownTechnologyFieldIDs`, because that state means a permanently completed non-repeatable field in MOOX.

## Original technology-selection UI off-by-one

The original research-selection screen has a separate presentation behavior.

`_Tech_Select_` at VA `0x10DC12`:

1. saves all eight real counters;
2. temporarily increments all eight counters by one;
3. constructs the list/names/cost display;
4. restores the original counters before returning.

Representative flow:

```text
0x10DC62  save player+0x21C..0x223
0x10DC94  temporary counter++
...
0x10E48E  restore saved counters
```

Selecting a field writes the active TechField but does not persist the temporary promotion. Therefore:

```text
strategic runtime counter = completed_levels
selection-screen preview  = completed_levels + 1
```

This explains public tables/UI observations that show a first Hyper-Advanced item at 25000 RP even though the authoritative original breakthrough helper evaluates a zero-counter first project at 15000 RP.

### MOOX policy

MOOX intentionally exposes the **authoritative simulation cost** through `ResearchChoices` and events:

```text
first Hyper project = 15000 RP
second              = 25000 RP
third               = 35000 RP
```

MOOX does not reproduce the old selection-screen preview off-by-one. This is an explicit behavior/UI cleanup, not an accidental divergence in the simulation core.

## Original 20-level selection-list boundary

The original technology-list builder contains a Hyper-specific comparison against decimal 20. Because the UI has temporarily promoted counters by one, an inactive field with 20 completed levels can disappear from the selectable list.

However the active-field branch is evaluated separately and the strategic cost/completion helpers themselves do not contain the same hard stop.

Therefore this checkpoint does **not** claim a proven global gameplay cap of 20. MOOX currently does not impose a 20-level simulation cap. The original UI re-selection boundary remains documented for later UI-fidelity work.

## MOOX normalized ruleset

`technologies.json` advances to schema 3 and now includes:

```json
"hyper_advanced": {
  "tech_field_ids": [75, 76, 77, 78, 79, 80, 81, 82],
  "cost_increment_rp": 10000,
  "verification": "original-exe-player-research-cost-hyper-counter-runtime"
}
```

This keeps the original-derived increment out of UI/client logic and allows the authoritative rules layer to calculate dynamic cost.

## MOOX persistent state

`StateSchemaVersion` advances from 6 to 7.

Per Empire:

```go
HyperAdvancedResearch []HyperAdvancedResearchLevel
```

with semantic entries:

```go
TechFieldID     int
CompletedLevels int
```

Canonical storage is sparse:

- missing entry = 0 completed levels;
- stored entries must be field 75..82;
- stored `completed_levels` must be >=1;
- entries must be strictly sorted and unique.

This is intentionally not the original fixed eight-byte storage layout.

## `repeat_field` research mode

Hyper-Advanced research uses a new explicit server-authoritative mode:

```text
selection_mode = repeat_field
```

Properties:

- no concrete Technology IDs;
- client sends only `tech_field_id`;
- `technology_id` is rejected;
- field remains eligible after completed levels;
- field becomes available when its normal predecessor is known;
- current dynamic cost and level metadata are projected by `ResearchChoices`;
- Creative/Uncreative do not change this Hyper field mode.

`ResearchChoice` exposes:

```text
base_cost_rp      = authoritative current strategic threshold
completed_levels  = persisted completed count
research_level    = completed_levels + 1
```

Observer/replay selection/progress/completion metadata carries the same level information.

## Switching

Existing MOO2/MOOX full accumulated-RP transfer on research switching is retained.

Switching from a normal field into a Hyper field or away from a Hyper field transfers the complete accumulated `ProgressRP`; the destination then evaluates breakthrough probability against its current dynamic cost.

## Regression coverage

Tests now lock:

- Hyper fields 75..82 normalized with +10000 RP per completed level;
- initial authoritative cost 15000 RP;
- level 1 completed -> next cost 25000 RP;
- level 3 completed -> current choice cost 45000 RP;
- `repeat_field` contains no Technology IDs;
- clients cannot submit a Technology ID for Hyper selection;
- Hyper breakthrough increments only that field level;
- no Hyper field is added to permanent known-field ownership;
- active Hyper research remains selected after breakthrough;
- RP resets to zero with no overflow;
- subsequent turn uses the increased cost;
- Hyper level state round-trips exactly through schema 7;
- invalid/duplicate/unsorted Hyper level state is rejected;
- normal-to-Hyper switching preserves fractional RP;
- `GameSession.ResearchChoices`, authoritative selection and Observer share the same Hyper semantics.

## Remaining research work

Hyper-Advanced repeated-field runtime is now closed for the strategic core.

Next Research slice:

1. Advanced-start randomized/race-aware technology ownership;
2. later Uncreative external-acquisition replacement using `Ensure_Uncreative_Field_OK_`;
3. later identify the original `0x21CAF` Dimensional Portal gate;
4. UI-fidelity work may separately revisit the original temporary +1 preview and 20-level list boundary, but neither belongs in authoritative strategic cost calculation.