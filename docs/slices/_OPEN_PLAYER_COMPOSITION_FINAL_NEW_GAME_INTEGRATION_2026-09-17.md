# Open slice 16.6 - Player composition and final New Game integration

Status: **Gate 1 complete / Gate 2 next**

Parent: `docs/slices/PLANNED_16_NEW_GAME_PRESET_RACE_BREADTH.md`
Plan: `docs/slices/PLANNED_16_6_PLAYER_COMPOSITION_FINAL_NEW_GAME_INTEGRATION.md`
Permanent Gate-1 evidence: `docs/research/SLICE_16_6_GATE1_PLAYER_COMPOSITION_AUDIT_2026-09-17.md`
Prototype evidence: `docs/research/prototypes/SLICE_16_6_OPPONENT_COMPOSITION_2026-09-17.svg`

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

## Guardrails

- Gate 1 is research/prototype only. Do not widen New Game validation before Gate 2.
- Do not treat all 13 catalog races as AI-runtime-supported merely because portraits/catalog metadata exist.
- Do not silently repair an unsupported count/race/galaxy tuple in the frontend.
- Do not implement client-side random race assignment.
- Advanced Starting Technology remains Slice 16.7 scope.

## Gate 2 checklist

- [ ] Freeze supported opponent/player-count x galaxy-size matrix.
- [ ] Freeze deterministic N-player home-star normalization.
- [ ] Freeze opponent race/controller assignment rules.
- [ ] Freeze duplicate/random policy.
- [ ] Freeze server-owned composition catalog/disabled-reason contract.
- [ ] Freeze final cross-setting validation and visual launch-summary layout.
- [ ] Freeze deterministic create/play/save/resume fixture matrix.

## Next

Freeze the first production composition boundary from the Gate-1 evidence before changing runtime behavior. The conservative evidence-backed option is one local Human/Klackon plus one or two Builtin-AI opponents; Gate 2 must explicitly accept that boundary or document the additional work required for a wider range.