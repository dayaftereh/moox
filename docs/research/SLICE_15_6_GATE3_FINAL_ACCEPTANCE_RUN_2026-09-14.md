# Slice 15.6 Gate 3 - final visible-browser acceptance run

Date: 2026-09-14
Status: **PASS - Gate 3 canonical visible-browser acceptance complete**

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

## Final-run continuation and result

### Checkpoint B - canonical research/supply prerequisite

**PASS**

- Continued the restored canonical game only through normal Planning resolution and visible Research choices.
- Completed the required fuel/supply chain through Deuterium Fuel Cells, the field-2 intermediate, Iridium Fuel Cells, Microlite Construction and finally Urridium Fuel Cells / 12 pc range.
- Population was reassigned through the visible Colony UI to accelerate research and later returned to production.
- Expanded normally with a Human Colony on planet 30 in system 10.
- The final authoritative state contained the four canonical Human supply Outposts on planets 32, 33, 34 and 43.

### Checkpoint C - ordinary strategic encounter -> interactive Tactical -> Battle Return

**PASS**

- Normal strategic play produced Battle #1, Human Combat Fleet #59 versus Darlok Combat Fleet #62, two ships per side.
- Entered Tactical from the visible Encounter screen using `Taktischen Kampf betreten`.
- Executed a visible Tactical Scan action. The canonical startfleet in this fixture has no weapon mounts (0/0 ready), so this specific acceptance battle validates the interactive Tactical lifecycle rather than Laser damage.
- The battle resolved authoritatively as `Tactical Retreat`, Darlok winner, zero destroyed ships and four surviving ships.
- Reloading the battle route projected the authoritative completed result and `Weiter zur Strategie` returned to normal Planning/Galaxy play.
- Laser slots/grouped volleys are separately covered by the earlier real user playtest and focused browser/server tests: grouped Laser x2 damage aggregation, split-slot readiness and repeated beam animation all passed before this final run.

### Checkpoint D - Troop Transport / Invasion / Conquest Victory

**PASS**

- The authoritative strategic state contained Human Colony #80 on planet 30, Human Outposts on planets 32/33/34/43, Troop Transport #89 and active Human-Darlok war.
- Human Combat Fleet #59 and Troop Transport #89 reached the final Darlok colony system through normal strategic Fleet movement.
- The browser entered the blocking Invasion decision and projected exactly one eligible transport.
- Visible `Mit allen Transportern invasieren` resolved the decision.
- Final authoritative result: `kind=conquest`, winner Empire 2 / Seat 1 (Human), eliminated Empire 3 (Darlok), completed Round 562 / Revision 1137.
- The browser visibly showed `SPIEL ABGESCHLOSSEN`, `Eroberungssieg`, `Human gewinnt durch Eroberung` and `Eliminierte Imperien: Darlok`.

The unusually high displayed round number is not a game-rule requirement. During automation, an old timed-out managed Chrome process continued issuing already-scripted `Fertig` clicks. Its process tree was identified and killed; the isolated 7192 server was then verified stable before the controlled Tactical/Return/Invasion/Victory steps. The extra empty turns affected only this isolated acceptance fixture's round count and did not bypass any required state transition or mutate the user's canonical 7171 Triangle game.

### Responsive/runtime smoke

**PASS**

- Final desktop Victory view: 791x605 CSS viewport, no horizontal document overflow.
- The same current Chrome activated the <=980px responsive path without horizontal overflow.
- Real 390x844 mobile metrics were already exercised in this Slice after the weapon-slot work:
  - Ship Designer: no page overflow; installed mount row and 40px quantity controls fit.
  - Tactical: 390px-wide HUD, compact slot strip, four 32px action buttons, no horizontal overflow; independent S1/S2 selection and spent-slot fallback passed.
- Therefore the final closeout does not invent a second browser emulation mechanism solely for the terminal Victory card; the mobile-critical interactive surfaces already have real-device-metric evidence.

### Final regression gates

**PASS**

- `go test ./... -count=1` - PASS across all packages, including `internal/app`, `internal/battle`, `internal/game`, `internal/server` and `internal/session`.
- `npm run build` - PASS (TypeScript + Vite production build).
- `git diff --check` - required again after closeout-document edits before the final commit.
- User canonical Triangle on 7171 remained isolated from the acceptance run; final verification during closeout showed Turn 9 / Encounters / Revision 23 / seed 32778.

## Gate-3 conclusion

The frozen Gate-2 browser vertical-slice contract is satisfied. New Game, meaningful progress, Save/Load/Resume, research/supply progression, normal strategic encounter, interactive Tactical, Battle Return, Troop Transport/Invasion and authoritative Conquest/Victory were all exercised on the canonical seed-0x8009 Human-vs-Darlok journey. Gate 3 is closed and the slice may proceed to Gate 4 closure.
