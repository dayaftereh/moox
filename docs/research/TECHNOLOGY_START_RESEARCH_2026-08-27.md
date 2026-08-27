# New-game technology ownership and research acquisition - 2026-08-27

Status: first deterministic technology-ownership slice implemented. Original MOO2 1.31 technology/tech-field tables and the staged new-game field list are normalized from the private reference executable. Exact research breakthrough probability/rounding and automatic per-turn integration remain deliberately deferred.

## Original 1.31 technology application table

The clean-room decoder now reads the concrete technology table directly from `Orion2.exe` 1.31:

- file offset: `0x1FC720`
- records: 203
- record size: 13 bytes
- byte `+0`: technology field ID (`0xFF` is normalized as `-1` for the exceptional non-field technology)
- byte `+5`: Strategic Combat availability flag

The first records reproduce the known MOO2 technology-field sequence exactly, and each normalized technology now stores `tech_field_id` with file-offset provenance. The existing TECHNAME.LBX source remains authoritative for the 203 technology names.

## Original 1.31 technology-field table

The technology-field table was located at:

- file offset: `0x1FBFB5`
- fields: 1..82
- record size: 23 bytes

Fields normalized in this checkpoint:

```text
+0x00 uint16  previous tech-field ID
+0x02 uint16  next tech-field ID
+0x0C uint32  base research cost in RP
+0x10 uint8   AI group
```

For example, field 1 decodes as previous=18, next=34, cost=400 RP, AI group=5. These values are stored in technology schema v2 and converted to the runtime fixed-point scale where needed.

## New-game staged technology fields

The private 1.31 executable contains the six-entry staged field list as `uint16` values at file offset `0x1FF7B0`:

```text
29, 55, 22, 57, 28, 23
```

Classic reverse-engineering documentation identifies these fields as Engineering, Nuclear Fission, Chemistry, Physics, Electronics and Cold Fusion, with technology field 0 as the always-known Starting Technology field.

The runtime therefore supports the two starts that are deterministic from this evidence:

### Pre-Warp

Known fields:

```text
0, 29
```

Derived normalized Technology IDs:

```text
32, 40, 103, 145, 166, 168
```

These are the field-0 technologies plus Engineering technologies.

### Average

Known fields:

```text
0, 29, 55, 22, 57, 28, 23
```

Derived normalized Technology IDs:

```text
32, 40, 41, 58, 63, 69, 100, 101, 103, 109,
119, 120, 121, 145, 157, 166, 167, 168, 187, 189
```

`InitializeEmpireTechnologies` materializes both `KnownTechnologyFieldIDs` and sorted `KnownTechnologyIDs`. Strategic-combat filtering is kept explicit in the new-game options so the same original technology table can support that game option without client-side filtering.

## Original SAVE10.GAM new-game cross-check

The private reference installation also contains an original auto-save at stardate 3500.0:

- file: `SAVE10.GAM`
- SHA-256: `ECE2EB06D782078DD0A6F746020A05691355303CEB02BBFBBE2233E987272BE1`
- player array start: `0x1AA0F`
- player record stride: `0xEA9`
- 203 technology-status bytes: player `+0x117`
- research progress: player `+0x1EA`
- research area/item: player `+0x320/+0x321`

All five active player records have zero research progress and no selected research item. Their status-3 technology set is identical:

```text
32, 40, 41, 58, 69, 100, 101, 103, 109, 119,
120, 121, 145, 157, 166, 167, 168, 187, 189
```

This is exactly the normalized Average start above except Technology 63 (Extended Fuel Tanks). The original `Orion2.exe` technology record for ID 63 belongs to Chemistry field 22 but has Strategic Combat availability byte `+5 = 0`. Across all 203 technologies in this save, every technology whose original Strategic flag is 0 also has save status 0. Therefore the 19-ID save observation is consistent with an Average **Strategic Combat** start, while the unfiltered Average start contains 20 IDs.

This is used as an independent original-save regression check; the save itself is not committed.
### Advanced

Advanced start is intentionally rejected for now. Its extra field selection is randomized/race-aware and has not yet been normalized sufficiently from the original generator. No deterministic-looking substitute is invented.

## Research state and ownership transition

`Empire` now persists:

- `KnownTechnologyFieldIDs`
- `KnownTechnologyIDs`
- optional active `ResearchState`

`ResearchState` contains:

- active `tech_field_id`
- the selected technology IDs belonging to that field
- fixed-point RP progress

`EconomyResolver.CompleteResearchField` is the authoritative ownership transition after a breakthrough has already been established. It:

1. validates the active field and its original RP cost,
2. rejects completion below that base cost,
3. validates that every selected technology belongs to the active field,
4. rejects fields/technologies already known,
5. adds the field and acquired technologies in stable sorted order,
6. clears active research,
7. emits `empire.research_completed` as a system strategic event.

A regression test confirms that completing the field containing Technology 155 makes the normalized Research Laboratory building become a legal production choice immediately through the existing server-side buildability projection.

## Session boundary

`GameSession.CompleteResearchField` commits this transition only in `post_resolution` phase. The session:

- clones the authoritative state,
- invokes the game-layer transition,
- validates the resulting state/event,
- commits atomically,
- increments the session revision,
- appends the `empire.research_completed` event to the strategic observer/replay stream.

This keeps future Wails, network and AI adapters away from direct `GameState` mutation.

## Breakthrough boundary established

The original/manual behavior and classic 1.50 documentation establish three boundaries that the runtime can rely on without yet implementing the random roll:

1. RP accumulates against the selected field's base cost.
2. Once the base cost has been exceeded, a per-turn breakthrough chance applies; the exact timing remains random.
3. At twice the base cost the breakthrough is guaranteed. Classic behavior spends the accumulated RP on breakthrough rather than carrying excess into the next project.

`CompleteResearchField` therefore remains a post-breakthrough ownership operation. It does not decide the random success roll. The exact probability/rounding rule between 1x and 2x cost remains the active reverse-engineering target.
## Deliberately deferred

This checkpoint does not yet claim or implement the original:

- research-target selection rules,
- research-point accumulation timing across colonies/empire,
- breakthrough probability above the nominal field cost,
- guaranteed-completion threshold,
- RP overflow/carry behavior after a breakthrough,
- Creative/Uncreative field-choice semantics beyond the explicit selected IDs already stored in `ResearchState`,
- Advanced-start randomized/race-aware extra fields.

The next research slice should isolate the original breakthrough/overflow algorithm before automatic per-turn research resolution is connected to `CompleteResearchField`.
