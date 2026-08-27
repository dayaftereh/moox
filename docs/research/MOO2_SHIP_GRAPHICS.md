# MOO2 ship graphics mapping

Baseline: official Master of Orion II 1.31, local reference `C:\ASH\Temp\mastori2`.

This note records the evidence used to map the six player hull classes and the three civilian ship types to strategic `SHIPS.LBX` graphics, the standard player tactical `CMBTSHP.LBX` layout, and the independently proven 20-frame tactical structure (five folded orientations x four neutral animation phases).

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

## Tactical CMBTSHP.LBX layout and frame semantics

`CMBTSHP.LBX` contains exactly 360 entries, matching 8 player colors x 45 slots.

Original executable evidence establishes:

```text
tactical_block = color_index * 45 + combat_picture_id
palette_block  = color_index * 45 + 44
```

Every combat picture slot 0..43 in all eight player-color groups is a 59x60 graphic with exactly 20 frames. Slot 44 is the per-color palette.

Archive-layout evidence:

- `Load_Combat_Ship_Palette_`: object 1 offset `0x39F99`, length `0x191`, SHA-256 `5676f51a6b7233a14a6d1efed4d10d9158fa30cab740c042039e8fb4083ccc64`,
- `Load_Individual_Ship_Pictures_`: object 1 offset `0x3A12A`, length `0x3BA`, SHA-256 `73532630ee0c7225343a07ec2c061f690f6a94b0f3f068612980ca28fbaf3784`.

### Strategic picture ID -> tactical picture ID

`Load_Combat_Ship_` proves that normal player designs copy their design picture byte into the combat-ship picture field. For size index 5 with a player owner, the function hard-codes picture ID 43, matching the normalized Doom Star strategic picture ID.

- `Load_Combat_Ship_`: object 1 offset `0x3954A`, length `0x4F7`, SHA-256 `27ca9b7efcbc52930d68c4de81fc42d4f7b27aaf6942e4f096ff1a143bd20a3d`.

Therefore the six normalized military hulls use the same picture IDs for strategic `SHIPS.LBX` and tactical `CMBTSHP.LBX` lookup. Civilian strategic picture IDs 45..47 are outside the CMBTSHP picture range 0..43 and are not assigned tactical assets by this evidence.

### Twenty frames = five folded orientations x four animation phases

The original drawing paths establish the meaning of all 20 frames.

`Draw_Ship_To_Bitmap_` uses static base frames 0, 4, 8, 12 and 16. It folds the 16 combat facings into five stored orientation indices and relies on draw mirroring for the opposite quadrants.

`Draw_Ship_` uses the same folded orientation calculation and adds an animation phase 0..3:

```text
frame = orientation_index * 4 + animation_phase
orientation_index = 0..4
animation_phase   = 0..3
```

Evidence:

- `Draw_Ship_`: object 1 offset `0x20062`, length `0x5CF`, SHA-256 `d7ac88d365eba8857bf93fb1e7c7b4bab587828502343346ac878ee0c87bce0e`,
- `Draw_Ship_To_Bitmap_`: object 1 offset `0x22B26`, length `0x26E`, SHA-256 `61ea32e4db51f981e77f30b73cc412e8e83ac933d0fecc3ac5f3e031603511ff`.

The term `animation_phase` is intentionally neutral. The code proves a four-phase frame cycle but this layer does not rename it to engine/thrust/damage animation without additional evidence.

### Tactical semantic assets

Each military hull has one tactical semantic set:

```text
ship_hull.<hull-id>.tactical
```

Each variant records:

- `color_index` 0..7,
- `style_index`,
- original `picture_id`,
- `orientation_index` 0..4,
- `animation_phase` 0..3,
- source `CMBTSHP.LBX` block,
- exact frame `orientation_index*4 + animation_phase`,
- block SHA-256,
- decoded 59x60 dimensions.

Variant counts:

- Frigate: 8 colors x 8 styles x 20 frames = 1,280,
- Destroyer: 1,280,
- Cruiser: 1,280,
- Battleship: 1,280,
- Titan: 1,280,
- Doom Star: 8 colors x 1 picture x 20 frames = 160,
- total tactical hull frame references: **6,560**.

Picture slots 40, 41 and 42 remain deliberately unlabelled. Slot 44 is the palette. No tactical meaning is assigned to strategic-only picture IDs 45..47.

The semantic layer therefore now covers the complete standard player military hull graphics in both strategic and tactical contexts without selecting arbitrary tactical frames.