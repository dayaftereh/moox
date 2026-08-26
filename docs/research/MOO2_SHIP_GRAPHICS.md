# MOO2 ship graphics mapping

Baseline: official Master of Orion II 1.31, local reference `C:\ASH\Temp\mastori2`.

This note records the evidence used to map the six player hull classes and the three civilian ship types to strategic `SHIPS.LBX` graphics. It also records the proven container layout of `CMBTSHP.LBX`, while deliberately deferring tactical multi-frame semantics until their frame meanings are independently decoded.

## Player hull identities

English `TECHNAME.LBX` block 0 contains the consecutive hull-name sequence:

1. Frigate
2. Destroyer
3. Cruiser
4. Battleship
5. Titan
6. Doom Star

`data/rulesets/moo2-1.31/ship_hulls.json` preserves these as size indices 0..5.

The original `Auto_Design_Ship_` picture-selection code is hash-verified before normalization. The relevant 1.31 range is LE object 1 `0x51731..0x5175A`:

- SHA-256: `6daa37d29c2818c73add925fd145b49728a015dcd16c9d00e87486eee6605de2`
- size indices 0..4 use eight picture styles each: `size_index*8 + style`, style 0..7,
- size index 5 uses hard-coded picture ID 43.

This yields:

- Frigate: picture IDs 0..7,
- Destroyer: 8..15,
- Cruiser: 16..23,
- Battleship: 24..31,
- Titan: 32..39,
- Doom Star: 43.

`Load_Doom_Star_Ship_Design_` independently passes ship size 5 to the auto-design path.

## Strategic SHIPS.LBX layout

`SHIPS.LBX` contains 449 entries in the local 1.31 reference.

The original `Get_Ship_Picture_Seg_` and `Load_Player_Ship_Palette_` functions establish the player-color layout:

```text
strategic_block = color_index * 50 + picture_id
palette_block   = color_index * 50 + 49
```

where player `color_index` is 0..7.

Evidence ranges:

- `Get_Ship_Picture_Seg_`, object 1 offset `0x48697`, length `0x3C`, SHA-256 `af7430b189af870cec1848fea4649c27dfb00d8af6c7f737a5c611f6c54075ef`,
- `Load_Player_Ship_Palette_`, object 1 offset `0x485E0`, length `0x98`, SHA-256 `a78aa3af973e4f20a1bbc4c844dc094d24b8cfc75a711e845cc9d34215b3c1bb`.

The eight player groups therefore occupy blocks 0..399. The remaining SHIPS entries 400..448 are outside this standard player-color formula and are not assigned by this mapping.

All strategic player ship picture blocks used here are single-frame graphics:

- ordinary hull/civilian pictures are 52x48,
- Doom Star picture 43 is 52x52,
- player palette slot 49 is not treated as a ship picture.

## Civilian strategic picture IDs

The original design initialization functions directly assign these picture IDs:

- Colony Ship: 45,
- Outpost Ship: 46,
- Transport: 47.

The hash-validated original function ranges are:

- `Load_Colony_Ship_Design_`: object 1 `0x464CD`, length `0xB9`, SHA-256 `3e468fc441a98fccc8214b88c60daf661ab01b31431d988155f5d785d3ee56cc`,
- `Load_Outpost_Ship_Design_`: object 1 `0x46586`, length `0xB8`, SHA-256 `b5a6c0dbe3e1a8aa471bc3ba71d07f05631ebfc67fb9974b1e17812878156c9d`,
- `Load_Transport_Ship_Design_`: object 1 `0x4663E`, length `0xE8`, SHA-256 `bddf60805ca545e5ee0580426fcb34c76a840d3c14279be4780c03ebb6a0e17b`.

Using the same 50-slot color formula, the semantic strategic asset layer contains:

- `ship.colony.strategic`: 8 color variants,
- `ship.outpost.strategic`: 8 color variants,
- `ship.transport.strategic`: 8 color variants.

## Strategic semantic assets

Each hull has one semantic strategic set:

```text
ship_hull.<hull-id>.strategic
```

Every variant stores:

- player color index,
- style index,
- original picture ID,
- source `SHIPS.LBX` block,
- frame 0,
- original block SHA-256,
- decoded dimensions.

Counts:

- first five hulls: 5 x 8 colors x 8 styles = 320 variants,
- Doom Star: 8 colors = 8 variants,
- three civilian ships: 3 x 8 colors = 24 variants,
- total strategic ship references: **352**.

Picture IDs 40, 41, 42, 44 and 48 are not assigned a hull/civilian meaning by this evidence and remain deliberately unlabelled.

## Tactical CMBTSHP.LBX layout

`CMBTSHP.LBX` contains exactly 360 entries, matching 8 player colors x 45 slots.

Original executable evidence establishes:

```text
tactical_block = color_index * 45 + combat_picture_id
palette_block  = color_index * 45 + 44
```

Evidence:

- `Load_Combat_Ship_Palette_`: object 1 offset `0x39F99`, length `0x191`, SHA-256 `5676f51a6b7233a14a6d1efed4d10d9158fa30cab740c042039e8fb4083ccc64`,
- `Load_Individual_Ship_Pictures_`: object 1 offset `0x3A12A`, length `0x3BA`, SHA-256 `73532630ee0c7225343a07ec2c061f690f6a94b0f3f068612980ca28fbaf3784`.

Unlike strategic SHIPS pictures, CMBTSHP blocks are multi-frame graphics with varying frame counts. Therefore MOOX does **not** currently create tactical assets by arbitrarily selecting frame 0. The next reverse-engineering step is to identify frame meaning/orientation/damage or animation usage from the original combat drawing code and then represent the whole tactical sprite set semantically.

This separation is intentional: archive/block layout is confirmed, but tactical frame semantics are not yet claimed.
