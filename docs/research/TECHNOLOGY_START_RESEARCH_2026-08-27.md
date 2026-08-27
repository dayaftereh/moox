# New-game technology ownership and research acquisition - 2026-08-27

Status: first deterministic technology-ownership slice implemented. Original MOO2 1.31 technology/tech-field tables and the staged new-game field list are normalized from the private reference executable. Research breakthrough probability/overflow remains deliberately deferred.

## Original 1.31 technology application table

The clean-room decoder now reads the concrete technology table directly from `Orion2.exe` 1.31:

- file offset: `0x1FC720`
- records: 203
- record size: 13 bytes
- byte `+0`: technology field ID (`0xFF` is normalized as `-1` for the exceptional non-field technology)

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
