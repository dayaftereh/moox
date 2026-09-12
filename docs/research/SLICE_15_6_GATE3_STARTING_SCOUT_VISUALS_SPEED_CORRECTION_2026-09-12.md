# Slice 15.6 Gate3 - Starting Scout visuals and Tactical speed correction

Status: **IMPLEMENTED / isolated fresh-game browser QA green**
Date: 2026-09-12

## Why this follow-up exists
After the Tactical mobile polish, the user questioned the temporary reduction of Nuclear Drive movement from 20 to 10 and asked for new games whose starting Scouts already have real persistent generated ship visuals.

## Original MOO2 combat-speed evidence
The previously decoded original MOO2 1.31 evidence in `TACTICAL_SHIP_COMBAT_BASELINE_2026-09-01.md` is explicit. Drive tables are per hull from Frigate through Doom Star:

- Nuclear minimum: `10,8,6,5,4,3`
- Nuclear pristine/max: `20,18,16,15,14,13`
- Fusion minimum: `12,10,8,7,6,5`
- Fusion pristine/max: `22,20,18,17,16,15`

`Current_Design_Base_Combat_Speed_` reduces the pristine maximum toward the minimum as engine damage accumulates. Therefore an undamaged Frigate baseline uses Nuclear 20 or Fusion 22. The temporary MinSpeed interpretation from the prior polish block was wrong and has been reverted in Tactical metadata, Battle validation, movement tests and scan tests.

## Persistent starting Scout visual contract
New-game creation now generates a fully resolved `ShipVisualGenome` v4 for every empire's initial Scout design.

- visual hull is `scout`, while authoritative gameplay hull remains the current Frigate rules baseline;
- `ShipDesign.VisualRevision = 1` at game creation;
- the fully resolved geometry is persisted on `ShipDesign.VisualGenome`, not just a seed;
- each starting Scout receives its own clone of the design genome plus `SourceVisualRevision = 1`;
- Scout 1 and Scout 2 of one empire therefore intentionally look identical because they are instances of the same design;
- different empire/design seeds generate independent visual identities;
- later browser-generator changes cannot alter the already persisted geometry.

The server-side generator mirrors the current browser v4 Scout profile and deterministic seed/PRNG contract. The seed includes game seed, empire ID and design ID, preserving deterministic reference games.

## Isolated fresh-reference-game evidence on 7178
Full live snapshots proved persisted visuals before any player action.

### `game-triangle-2pc`, seed `0x800A`
- Human Scout design: v4 `scout`, `sleek / bulb`, VisualRevision 1
- Darlok Scout design: v4 `scout`, `spear / fork`, VisualRevision 1
- Psilon Scout design: v4 `scout`, `organic / manta`, VisualRevision 1
- every starting Scout has SourceVisualRevision 1 and the exact persisted design genome

### `game-1`, seed `0x8009`
- Human Scout design: v4 `scout`, `spear / hammer`, VisualRevision 1
- second empire Scout design also has its own persistent seed/revision

Browser fleet-picker QA confirms the starting Scout SVGs are rendered from persisted v4 genomes (`data-hull-id="scout"`, `data-genome-version="4"`) rather than the previous Frigate fallback.

## Acceptance / authority notes
- Tactical combat speed remains server-authoritative.
- Visual genomes are presentation-only authoritative state and do not alter military stats.
- The existing Colony Ship/other special civilian visuals remain their dedicated special-ship presentation path; this change gives ordinary starting military Scouts equally stable visual identity through the military visual-genome path.
- Existing save/reload cloning rules remain unchanged.
