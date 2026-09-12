# Slice 15.6 Gate 3 - Tactical UI refinement review
Date: 2026-09-12
Status: **OPEN - review focus after Block5E completion**
User workstream shorthand: **5A Tactical UI**
## Naming / history guard
The user refers to this resumed workstream as `5A Tactical UI`. Existing Slice 15.6 documentation already uses `Block5A` for the completed Galaxy knowledge/exploration authority block, so the repository history is not renumbered or rewritten. This document is the active Tactical UI refinement review within Slice 15.6 Gate 3.
## Entry condition now satisfied
Block5E is complete. The Triangle reference scenario now provides a real browser gameplay path from strategic play into Tactical combat, including the previously blocked larger-Fleet case:
- co-located known foreign Fleets can be inspected;
- `Angreifen` performs authoritative war declaration;
- encounter creation happens on the next `Fertig` strategic-resolution boundary;
- Tactical accepts 3v2+ concrete combat-Ship counts;
- unarmed Scouts are valid Tactical participants;
- strategic Colony/Outpost/Troop vessels do not materialize as Tactical ships;
- explicit `R├╝ckzug` exists for a no-weapon/retreat outcome;
- on 2026-09-12 the user played the Triangle scenario through the real encounter flow and successfully reached the Battle/Tactical UI.
This is the stable entry point required before judging or redesigning the Tactical presentation.
## Existing Tactical baseline to preserve while reviewing UI
Slice 15.5 remains the closed functional authority baseline. Current Tactical presentation already has:
- participant-safe server-projected SVG battlefield;
- authoritative active ship / initiative sequencing;
- server-projected legal movement points;
- server-projected legal fire actions/targets;
- read-only Scan/ship inspection;
- End Activation and Retreat commands;
- server-authoritative damage/readiness/state refresh;
- desktop wheel zoom + drag pan;
- touch one-finger pan + two-finger pinch zoom;
- explicit zoom/recenter controls;
- effectively unbounded camera presentation without a normal arena edge;
- responsive desktop/mobile layout and shared DE/EN localization;
- BattleSession -> strategic return contract.
These are behavior/authority contracts, not a freeze on visual design.
## Review scope
The next pass is deliberately UI/HMI-first and should be evaluated in the real Triangle Tactical encounter. Review and refine, without moving authority into React:
1. overall battlefield composition and visual hierarchy;
2. ship scale, orientation, selection state and active-ship emphasis;
3. friendly/enemy distinction and fleet readability for 3v2+ battles;
4. movement visualization and destination affordances;
5. attack/target visualization, weapon affordances and readiness feedback;
6. initiative/turn-order presentation;
7. ship health, armor/structure and damage feedback;
8. Scan/details presentation and information density;
9. action controls, mode switching, End Activation and Retreat placement;
10. camera controls, default framing, pan/zoom ergonomics and recenter behavior;
11. desktop use of available width/height;
12. mobile/touch layout, overlays/sheets and no-hover operation;
13. battle-entry, active-battle and battle-result/return transitions;
14. animation/polish only where it remains presentation-only and does not affect deterministic authority.
## Acceptance discipline
- Use the real `game-triangle-2pc` strategic path to reach Tactical whenever practical; do not substitute a React-only mock for final acceptance.
- Preserve server authority for movement legality, target legality, range, RNG, damage, readiness, initiative and terminal result.
- Support 3v2+ presentation; do not regress to assumptions that exactly two ships exist per side.
- Unarmed ships and Retreat must remain valid.
- Civilian strategic vessels must remain excluded from Tactical materialization.
- Desktop and mobile/touch are both required acceptance surfaces.
- Block5F scanner/deep-space Fleet intelligence remains queued until this Tactical UI review has a stable direction.
## Immediate next step
Review the live Tactical screen with the user and convert the requested visual/interaction changes into small implementation blocks, each with browser-visible acceptance against the real Triangle encounter.