# Race Designer Pick costs - direct 1.31 extraction

Checkpoint: 2026-08-26

## Direct source

`RACESTUF.LBX` block 0 contains the English Race Designer group/option labels. `RACESTUF.LBX` block 6 is 57 bytes long and contains:

- 53 signed 8-bit Pick costs,
- followed by four zero bytes.

The 53 costs line up one-for-one with the 53 selectable options in block-0 order.

Examples:

- `population_growth_minus_50`: offset 0 = `0xFC` = -4
- `farming_minus_half`: offset 3 = `0xFD` = -3
- `creative`: offset 44 = `0x08` = 8
- `tolerant`: offset 45 = `0x0A` = 10
- `warlord`: offset 52 = `0x04` = 4

Block 6 SHA-256:

```text
b4a92e2b68be83aa6f51746c01988a8749828675fbfab6973252ba139f318423
```

The normalized `race_traits.json` now reads these signed bytes directly and records a per-option `pick_cost_source` pointing to `moo2-1.31-racestuf-block6` plus byte offset. `verification.pick_cost` is now `original-observed`.

The secondary StrategyWiki source remains only for Race Designer facts not yet located directly in the binary data, such as the starting/negative Pick budget and some behavioral cross-checks.

## CUSTMSTR variant

`CUSTMSTR.LBX` blocks 0..2 contain a related but different Race Designer presentation:

- Farming disadvantage label is `-1 Food` rather than `-1/2 Food`.
- `Poor Home World` is absent.
- its 56-byte block 3 contains a signed-byte cost sequence that therefore diverges from the `RACESTUF` ordering.

The first divergence is Farming:

```text
RACESTUF: -3
CUSTMSTR: -5
```

After the point where `Poor Home World` exists in `RACESTUF` but not `CUSTMSTR`, the remaining special-ability costs are shifted by one position in the CUSTMSTR byte sequence.

This data is preserved as a distinct historical/interface variant and is **not** used to overwrite the official MOOX `moo2-1.31` Race Designer ruleset until its original runtime role is established.
