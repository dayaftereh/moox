# Slice 17.8 User Review - Unexpected Tactical Retreat and Missing Result

Date: 2026-10-08
Status: IMPLEMENTED - awaiting user review

## Reproduced user issue

- In game-triangle-combat-review on 7171, the Human player fired at the Darlok fleet and ended the third Human activation.
- The game unexpectedly transitioned directly to strategy. Authoritative recent_resolutions reported battle_completed / tactical_retreat, Human winner, zero destroyed ships and six survivors, not a kill victory.
- A deterministic session regression replayed the same three-ship weapons sequence. The builtin AI exhausted multiple weapon actions, then reached a Fighter group whose source carrier was unavailable.
- That Fighter group's attack stock ran out and tacticalFighterBeginReturn rejected the unavailable carrier. EndActivation therefore failed, leaving retreat as the AI's first/only permitted command.

## Gameplay fix

- When an exhausted Fighter group has no functional source carrier, do not attempt to begin a return to that carrier.
- The arrival resolver produces a carrier_unavailable destruction outcome through the established Fighter group event mechanism. Normal returns to functional carriers remain unchanged.
- Permanent integration regression in internal/session/tactical_fighter_carrier_retreat_regression_test.go fires Human weapons at the opposing ship and verifies that the AI completes its actions and hands a regular active Tactical turn back to Human instead of retreating.

## Result-display fix

- Repeated resets of the reference game reuse battle-completion event IDs (observed event-16). Session-storage result acknowledgements can therefore refer to an earlier playthrough.
- On a direct battle route, an authoritative completed-battle summary now keeps the result view visible even when its event ID was already acknowledged. Navigation to strategy requires the explicit result-view Continue button.
- A tactical_retreat outcome has a localized Retreat/Rueckzug label and explains which empire withdrew; other outcomes retain their existing display.
- The Web contract test checks the completed-result route guard and retreat rendering.
- Live Chrome QA on 7171 with an artificially persisted old event-16 acknowledgement showed Kampfergebnis, Rueckzug, Darlok withdrew, winner Human, zero destroyed and six surviving ships; Continue returned to strategy only when clicked.

## Guardrails

- No AI retreat decisions or strategic casualty reporting are synthesized by the browser.
- Real voluntary retreat remains supported and must still show a Battle Review.
- No repository reset, clean, or push.
- Keep Slice 17.8 open pending user acceptance, and do not open Slice 17.9.
