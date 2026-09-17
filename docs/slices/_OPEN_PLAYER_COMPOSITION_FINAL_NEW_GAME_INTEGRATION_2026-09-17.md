# Open slice 16.6 - Player composition and final New Game integration

Status: **Gate 3 complete / Gate 4 next**

Parent: `docs/slices/PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md`
Plan: `docs/slices/PLANNED_16_6_PLAYER_COMPOSITION_FINAL_NEW_GAME_INTEGRATION.md`
Permanent Gate-1 evidence: `docs/research/SLICE_16_6_GATE1_PLAYER_COMPOSITION_AUDIT_2026-09-17.md`
Prototype evidence: `docs/research/prototypes/SLICE_16_6_OPPONENT_COMPOSITION_2026-09-17.svg`
Permanent Gate-2 freeze: `docs/research/SLICE_16_6_GATE2_PLAYER_COMPOSITION_FREEZE_2026-09-17.md`
Permanent Gate-3 implementation: `docs/research/SLICE_16_6_GATE3_PLAYER_COMPOSITION_IMPLEMENTATION_2026-09-17.md`

## Gate 1 checklist

- [x] Audit original/runtime player-count limits and galaxy-size dependencies.
- [x] Audit built-in AI support for more than one opponent.
- [x] Freeze deterministic race-assignment/randomization behavior candidates for Gate 2.
- [x] Prototype original-range number tiles 1-7 and compact opponent portrait layout.
- [x] Verify how disabled counts/combinations should be explained at 390 px mobile.

## Gate 1 result

- Original documented setup range is **2-8 total players = 1-7 opponents**. The earlier 1-9 visual target was incorrect for original parity.
- Production MOOX remains exactly two players and uses a pair-only generic home-star allocator, so no runtime breadth was enabled in Gate 1.
- Core/session/controller evidence already supports a three-player class; an existing regression runs one local human plus two Builtin-AI seats on the reference triangle.
- Human/Klackon remain the only server-supported local preset races; Darlok is a grandfathered fixed AI baseline with known parity gaps.
- Strongest Gate-2 V1 candidate: 1 opponent = Darlok; 2 opponents = Darlok + the other supported Human/Klackon race, all AI, unique race IDs.
- Counts 3-7 remain original-range targets but require wider honest AI race support, duplicate-policy change, or deferral.
- Random opponent assignment should remain absent in V1 unless a server-owned seeded candidate set/order/RNG contract is explicitly frozen.

## Gate 2 result

- V1 supports exactly 1 or 2 Builtin-AI opponents on every supported galaxy size.
- Counts 3-7 remain visible/planned with reason `additional_ai_race_breadth_required`.
- Two-player home placement is byte/topology compatible with the current farthest-pair algorithm; seat 3 uses a deterministic maximin extension with no additional selection RNG draw.
- Automatic composition is Human -> Darlok[/Klackon] or Klackon -> Darlok[/Human].
- Duplicate preset races and random opponent assignment remain unsupported.
- `GET /api/v1/new-game/compositions` is the frozen server-owned availability/assignment contract.
- The final New Game launch briefing renders galaxy, difficulty, player race/name, starting technology, opponents and seed from the same selected state submitted to Create Game.
## Gate 3 result

- Server-owned `/api/v1/new-game/compositions` is live with supported 1/2 and planned 3-7 counts.
- Production New Game supports the frozen 2/3-player classes and deterministic farthest-pair + maximin home placement.
- Public HTTP controller roles are local-human + Builtin-AI opponents; wrong explicit roles are rejected.
- The complete 480 accepted Game-layer tuple matrix is deterministic.
- Seven production opponent-count SVGs, opponent cards, server reasons and final launch briefing are integrated.
- Real Chrome smoke passes on 390 px + desktop with 35 browsable New Game options.
- Three-player Pre-Warp and Average persistence export/import/re-export/restore regressions pass.

## Gate 3 checklist

- [x] Composition catalog/API and exact 2/3-player validation.
- [x] Backward-compatible deterministic N-player home selection.
- [x] Frozen controller/race composition and rejection coverage.
- [x] Seven opponent-count assets and manifest/contract checks.
- [x] Opponent selector/cards/disabled reasons and dynamic request binding.
- [x] Final launch briefing.
- [x] 480-tuple determinism plus app/server/persistence regressions.
- [x] Real 390 px + desktop browser selector smoke.
## Guardrails

- Gate 1 is research/prototype only. Do not widen New Game validation before Gate 2.
- Do not treat all 13 catalog races as AI-runtime-supported merely because portraits/catalog metadata exist.
- Do not silently repair an unsupported count/race/galaxy tuple in the frontend.
- Do not implement client-side random race assignment.
- Advanced Starting Technology remains Slice 16.7 scope.

## Gate 2 checklist

- [x] Freeze supported opponent/player-count x galaxy-size matrix.
- [x] Freeze deterministic N-player home-star normalization.
- [x] Freeze opponent race/controller assignment rules.
- [x] Freeze duplicate/random policy.
- [x] Freeze server-owned composition catalog/disabled-reason contract.
- [x] Freeze final cross-setting validation and visual launch-summary layout.
- [x] Freeze deterministic create/play/save/resume fixture matrix.

## Next

Run Gate 4 acceptance against the implemented contract: live create/play walkthroughs for supported 1/2 opponent classes, representative Human/Klackon, Pre-Warp/Average, Small/Huge and Easy/Impossible settings, confirm planned counts 3-7 remain blocked, verify launch briefing/persistence, then close Slice 16.6 and hand off to Slice 16.7.