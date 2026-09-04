# Planned slice 15.1 - UX architecture / navigation / design system contract

Status: **open; Gate 1 audit complete; Gate 2 interaction/design-system freeze pending user approval**.

Queue position: **15.1 of Slice-15 family**.

## Objective

Define the complete browser information architecture and reusable interaction/design-system contract before broad gameplay screens or visual asset production begin.

## Dependencies

- Slice 15 parent epic.
- Existing Slice-08 React/server authority baseline.
- Slice 14 save/resume transport semantics for lifecycle planning.

## Scope guard

In scope:

- screen map from Main Menu through New Game, live game, save/load and result;
- galaxy/system/colony/fleet/research/diplomacy/battle navigation model;
- persistent game shell, primary/secondary navigation and contextual panels;
- responsive desktop/tablet/mobile behavior and minimum viewport rules;
- modal/dialog/toast/tooltip/loading/error/empty-state interaction grammar;
- initial design tokens for spacing, typography scale, color roles, elevation, borders and control states;
- component primitives for buttons, tabs, cards, lists, data rows, resource badges and dialogs;
- keyboard/focus/accessibility baseline where practical;
- wireframes or screenshot-level mockups for the canonical match journey.

Defer:

- production graphic assets, portraits, ship art and rich backgrounds to 15.3;
- full strategic gameplay wiring to 15.2;
- final rich decision/persistence presentation to 15.4;
- complete-game polish to 15.5.

## Gate 1 - UX audit

- [x] Inventory current App.tsx/styles.css structure and existing browser routes/state transitions.
- [x] Map every canonical Human-vs-AI user journey and required screen/panel.
- [x] Identify navigation ambiguity, modal collisions and responsive constraints.
- [x] Audit existing server invalidation/reconnect states that need explicit UX states.
- [x] Produce low-fidelity screen map and component inventory.
- [x] Present Gate-2 UX contract before implementation.

## Gate 2 - Interaction/design-system freeze

- [ ] Freeze screen hierarchy and navigation model.
- [ ] Freeze game-shell layout regions and responsive behavior.
- [ ] Freeze interaction grammar for dialogs, toasts, errors, loading and focus.
- [ ] Freeze baseline design-token/component API without locking final art direction.
- [ ] Freeze wireframe acceptance set for canonical workflows.

## Gate 3 - Implementation

- [ ] Refactor monolithic browser shell into the accepted layout/component structure.
- [ ] Implement navigation, responsive shell and reusable primitives.
- [ ] Implement lifecycle/loading/error/invalidation presentation skeletons.
- [ ] Add component/navigation regressions where useful.
- [ ] Keep all gameplay rules and legality server-authoritative.

## Gate 4 - QA + close

- [ ] Verify every accepted wireframe/navigation path in browser runtime.
- [ ] Verify responsive/focus/loading/error states at representative phone and desktop viewports.
- [ ] Run the built UI on `0.0.0.0:7171`, verify it through the current NetBird IP, and keep the managed `moox-ui-review` Chrome session on the same live build for desktop QA.
- [ ] Run web build plus relevant Go/server regressions and `git diff --check`.
- [ ] Update evidence/status/HISTORY and close marker.
