# Slice 15.6 Gate 3 - final visible-browser acceptance run

Date: 2026-09-14
Status: **IN PROGRESS - Save/Load/Resume checkpoint green**

## Acceptance environment

- isolated current server: `127.0.0.1:7192`
- browser profile: `moox-15-6-final-acceptance`
- production code from current `main`
- canonical New Game settings entered through the visible browser New Game surface:
  - seed `0x8009`
  - Small / Normal galaxy
  - Average technology
  - Tactical / Strategic Combat disabled
  - Human vs Darlok built-in AI
- acceptance game ID: `slice-15-6-acceptance-20260914`
- canonical 7171 Triangle user playtest is not mutated by this run.

## Checkpoint A - New Game -> meaningful progress -> Save -> advance -> Load/Resume

**PASS**

Visible browser sequence:

1. Main Menu -> `Neues Spiel`.
2. The New Game screen visibly showed canonical defaults (`0x8009`, Small/Normal, Average/Tactical, Human, Darlok); only the game ID was changed to the dedicated acceptance ID.
3. `Spiel erstellen` opened Round 1 Planning through the normal game route.
4. Research HUD -> canonical Chemistry field -> `Deuterium Fuel Cells` was selected through the visible Research overlay.
5. `Fertig` submitted the normal Planning turn and reached Round 2 Planning. Visible state included `62 BC` and Deuterium progress `4% · 27T`.
6. Game menu -> `Spiel speichern` downloaded `moox-slice-15-6-acceptance-20260914-turn-2.json` (32101 bytes).
7. `Fertig` advanced normally to Round 3 (`74 BC`, research `7% · 26T`).
8. The saved file was supplied to the existing browser file-input handler. For automation only, the OS file-picker interaction was replaced by setting the same `<input type=file>`; no game API or state mutation was used.
9. The normal visible confirmation dialog identified the save as `slice-15-6-acceptance-20260914, Runde 2` and warned that the hosted game would only be replaced after server validation.
10. `Spiel wiederherstellen` restored the save through the normal browser persistence path.
11. The game returned to Round 2 Planning with the exact saved meaningful state: `62 BC`, research `4% · 27T`, same game route/identity.

This satisfies the frozen Gate-2 mid-game Save/Load/Resume requirement for the canonical visible-browser journey. Automatic process crash recovery is intentionally a separate non-contract concern documented in `SLICE_15_6_CLOSEOUT_PERSISTENCE_RESTART_DISPOSITION_2026-09-14.md`.

## Remaining final-run acceptance

- continue the canonical visible-browser game through the required supply/research progression;
- enter and execute at least one supported interactive Tactical battle through ordinary strategic play;
- continue strategically after Battle Return;
- use normal Troop Transport / Invasion UX and reach authoritative Conquest/Victory;
- desktop/mobile runtime and responsive smoke checks;
- full regression/build gates;
- Gate-4 close.
