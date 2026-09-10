# Slice 15.5 Gate 3 Block 4b - Tactical ship visuals and full Battle Return QA

Date: 2026-09-10

## Goal

Finish the visual handoff from Slice 15.3 and prove the interactive Tactical path through a real 2v2 Battle all the way back to strategic play after the Block-4a Armor-to-Structure fix.

## 15.3 ship visual integration

The simple semantic Tactical triangle remains as a low-opacity SVG fallback, but each live Tactical ship now renders the accepted Slice-15.3 `ProceduralShipGlyph` inside the world-coordinate marker.

The procedural seed is stable per empire/ship identity. Tactical Facing still rotates the authoritative world marker; the art itself adds no gameplay state. Own/enemy palettes are applied through the existing procedural vector CSS variables.

Managed-browser QA confirmed four Tactical combatants render four procedural glyphs in the real 2v2 encounter.

## Real full-battle fixture

QA used an isolated real GameSession on port 7181 based on the current canonical seed. Test-only setup:

- both two-Scout combat fleets placed in the same neutral system;
- Human and Darlok set to War;
- both Human Scouts armed with one supported standard Laser;
- both Darlok Scouts left unarmed;
- Seat1 local human, Seat2 built-in AI.

This fixture intentionally makes the battle resolvable through short repeated Human Laser activations while still exercising a true 2v2 initiative/round flow. No browser mock was used.

## Browser path

From Turn 1 / Planning:

1. clicked normal `Fertig`;
2. app created Battle 1 and auto-routed to the Battle Entry surface;
3. clicked `Taktischen Kampf betreten`;
4. confirmed 4 ship markers and 4 procedural ship glyphs;
5. Human activations repeatedly selected the server-projected Laser action/target and then End Activation;
6. built-in AI handled its own unarmed activations;
7. the battle progressed through multiple rounds without the former Armor-overflow rejection;
8. both Darlok ships were destroyed;
9. Slice-15.4 Battle Return appeared automatically;
10. `Weiter zur Strategie` returned to Galaxy / Turn 2 / Planning.

## Authoritative result evidence

The Battle Return surface showed:

- winner: Human;
- outcome: Tactical Victory;
- destroyed ships: 2;
- visible survivors: 2;
- destroyed identities: ship 60 and ship 61;
- survivors: Scout 1 and Scout 2.

The participant snapshot after acknowledgement contained no live Battles and its authoritative `battle_completed` summary contained:

- `destroyed_ship_ids: [60,61]`;
- `surviving_ship_ids: [57,58]`;
- winner seat 1 / winner empire 2;
- outcome `tactical_victory`.

The browser then showed Turn 2 / Planning with no Battle/Return overlay.

## Block-4a regression confirmation

Before Block 4a, the same style of full browser battle failed once Armor had been depleted because Slice-07 rejected Armor overflow/internal selection. This rerun completed normally, proving the supported aggregate Structure path is usable end to end rather than only in unit tests.

## Cleanup

The isolated 7181 QA server, managed Chrome session and temporary imported snapshot files were removed/stopped after evidence collection. Canonical 7171 was not used as the mutation fixture.

## Remaining Gate-3 work

Still intentionally open after this block:

- browser-visible stale/illegal Tactical command rejection/refetch behavior;
- repeatable deterministic browser regression coverage for the critical Tactical path;
- final Gate-4 close pass / full validation close marker.
