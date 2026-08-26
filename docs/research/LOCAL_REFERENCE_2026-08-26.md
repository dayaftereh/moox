# Local Master of Orion II 1.31 Reference

Observed on 2026-08-26 at:

`C:\ASH\Temp\mastori2`

This is treated as a private research input and is **not** copied into the MOOX Git repository.

## Identification

The local `README.TXT` identifies the installation as:

- Master of Orion II: Battle at Antares
- Version 1.31
- Readme dated 11 April 1997

The directory contains both DOS/primary and Windows 95 executables:

- `Orion2.exe` - 2,644,842 bytes
- `ORION95.EXE` - 1,564,160 bytes

It also contains `PATCH13.LBX`, consistent with the official 1.31 patched data set.

## Inventory summary

- Total files: 420
- Total size: approximately 322.38 MiB
- `.LBX` files: 373
- `.GAM` save files: 1 (`SAVE10.GAM`)
- Additional sound-driver subdirectories: `MT32`, `SB16`, `SC55`

## Important archives already identified

Examples relevant to exact-rule and content research:

- `RACESTUF.LBX` - race-designer labels/trait vocabulary; printable content already confirms numerical race modifiers and special-ability names.
- `RACEOPT.LBX`, `RACENAME.LBX`, `RACES.LBX`, `RACESEL.LBX` - race selection/design presentation and related data.
- `HERODATA.LBX`, `OFFICER.LBX`, `SKILDESC.LBX` - leader/officer data and skills.
- `SCIENCE.LBX` - science/research screen/data payloads.
- `COLBLDG.LBX`, `BLDG0.LBX` ... `BLDG5.LBX`, `BILLTEXT.LBX` - colony/building-related content.
- `DESIGN.LBX`, `SHIPS.LBX`, `SHIPNAME.LBX`, `BEAMS.LBX` - ship design, ships, names and combat weapon presentation/data.
- `PLANETS.LBX`, `PLNTSUM.LBX` - planets and planet UI/data.
- `DIPLOMAT.LBX`, `DIPLOM*.LBX` - diplomacy content.
- `EVENTS.LBX`, `EVENTM*.LBX`, `MONSTER.LBX` - events and monsters.
- `COUNCIL.LBX`, `COUNCMSG.LBX` - Galactic Council.
- `ANTAROOM.LBX`, `ANTARMSG.LBX`, `ANATKFIN.LBX`, `ANWINFIN.LBX` - Antaran content.
- `HELP.LBX`, `MAINTEXT.LBX`, `ESTRINGS.LBX`, `MSGENG.LBX` - in-game text/help/string sources.

## LBX format confirmation

Multiple files were checked directly and match the documented SimTex LBX signature/layout:

- 16-bit block count at offset 0
- bytes `AD FE` at offset 2
- data offset table beginning with `0x00000800`

Observed examples:

| Archive | Blocks | Size |
| --- | ---: | ---: |
| `RACESTUF.LBX` | 14 | 9,471 B |
| `RACEOPT.LBX` | 7 | 657,404 B |
| `HERODATA.LBX` | 1 | 6,005 B |
| `SCIENCE.LBX` | 3 | 965,928 B |
| `COLBLDG.LBX` | 7 | 116,904 B |
| `SHIPS.LBX` | 449 | 585,096 B |
| `PLANETS.LBX` | 30 | 5,513,320 B |
| `PATCH13.LBX` | 2 | 30,052 B |

This means an MOOX-owned read-only LBX inventory/extraction tool can work against this installation immediately.

## Direct data proof

A simple printable-string inspection of `RACESTUF.LBX` already exposes the original race-designer vocabulary and visible numeric modifiers, including categories such as population growth, farming, industry, science, money, ship attack/defense, ground combat, spying, governments and special abilities such as Creative, Uncreative, Lithovore, Cybernetic, Tolerant, Telepathic, Warlord and others.

The archive also contains multiple language variants (English, German, French and more). This gives us a useful way to distinguish display strings from underlying rule structures.

## Reference hashes

SHA-256 values are recorded solely to identify this exact local reference set:

- `Orion2.exe`: `7AE2AC2E5904CA330009AF2827279D889906B0B9B7A8854C38EB707A56E955B5`
- `ORION95.EXE`: `6E19AFDC98F1AEDCB8D2F974D5B658B0C855F54529BDABDDE193F5266E275185`
- `PATCH13.LBX`: `79AA4BE4825299C7CF2AC4886069979E3979157304EF1769C59B2082BE26D913`
- `RACESTUF.LBX`: `A7942F4F13CA081C6D8D4F53266A3EE79198F92C786402BD41C33ABBEECBB5B4`
- `HERODATA.LBX`: `38705F160EE493F3BC4D2D5F1B980B0D26DFD784D80BFE072EC88575DF938E18`

## Recommended next extraction order

1. Build a generic LBX inventory/parser in `tools/lbx/`.
2. Catalog all 373 archives and block offsets/sizes into a generated local JSON report.
3. Decode text/string archives first because they provide names and labels for later binary structures.
4. Decode race-design and race data.
5. Decode technology/building tables.
6. Decode ship components and hull data.
7. Decode leaders and events.
8. Inspect `SAVE10.GAM` separately as a possible concrete game-state fixture.
9. Only after structure is understood, add graphics/audio decoders for private UI reference.

The key result is that MOOX now has a concrete official-1.31 reference data set against which exact gameplay behavior can be researched and tested.
