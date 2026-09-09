# Slice 15.4 - Gate 2 rich interaction contract draft

Date: **2026-09-09**

Status: **accepted / frozen for Gate 3 on 2026-09-09**.

## Approval

Approved by the user on 2026-09-09 without requested changes. This document is the frozen Gate-2 interaction contract for Slice 15.4.

## 1. Interaction principle

15.4 presents authoritative human decision boundaries as unmistakable **priority states**. It must never create browser-only legality, costs, combat math or phase advancement.

Three presentation levels:

1. **Blocking decision** - user action is required before the server can continue (Encounter, Invasion, Colony Base).
2. **Transition summary** - server already resolved something and the player should understand what changed (Research breakthrough, battle return, loaded game, victory).
3. **Lifecycle status** - connection/save/load state that controls whether mutation is safe.

## 2. Encounter -> Tactical contract

Route: `#/game/<gameID>/battle/<battleID>`.

### Battle Entry mockup

```text
+------------------------------------------------------+
| ENCOUNTER - System 04                  Turn 17       |
| [spectral star]                                      |
|                                                      |
| Your Empire                    Darlok                 |
| [fleet/ship visuals]     VS     [fleet/ship visuals] |
| 2 combat ships                  3 combat ships        |
|                              Colony at risk           |
|                                                      |
| [ Enter Tactical Combat ]                            |
| Tactical unavailable: <server reason>   (if needed)  |
+------------------------------------------------------+
```

Requirements:

- System/sides/ship/fleet/colony context from Battle Spec and public identities.
- CTA exists only when the battle has a supported Tactical spec.
- 15.4 route/shell exists before 15.5; 15.5 supplies the battlefield interaction inside it.
- No autoresolve button unless a future authoritative command explicitly provides one.
- Browser cannot navigate past an unresolved human battle and pretend the turn continued.

### Tactical -> strategic return

On battle completion, show a one-time participant-safe summary keyed to authoritative result/event ID:

- winner/outcome;
- destroyed/surviving visible ships;
- colony/system consequence if known;
- Continue returns to the normal strategic route after acknowledging presentation only.

## 3. Invasion decision contract

Blocking modal/card over the strategic shell.

```text
INVASION OPPORTUNITY
System 06 - Orion II
Defender: Darlok
Available troop transports: 2

[ Invade ]   [ Decline ]
```

- Names/visuals replace raw IDs where public data exists.
- Invade submits exactly the server-projected eligible transport fleet IDs.
- Decline submits only the projected colony ID.
- Buttons lock while command is pending.
- Rejection shows the server reason and triggers refresh.

## 4. Colony Base decision contract

Blocking Post-Resolution decision.

```text
COLONY BASE READY - Alpha Prime
Choose one legal planet in this system:

[ Ocean II ] [ Barren III ] [ Terran IV ]

Or discard the Colony Base for +25 BC
[ Scrap Colony Base ]
```

- Target cards use existing OrbitalBodyArt and planet facts.
- Only projected legal target IDs are selectable.
- Refund is displayed from authoritative `trash_refund_bc`.
- Submit through existing immediate-command endpoint with current base revision.

## 5. Research breakthrough contract

Research completion is automatic server resolution. Presentation must not require a fake "Complete Research" command.

On the next safe snapshot, a one-time transition card/overlay may show:

```text
RESEARCH BREAKTHROUGH
Advanced Engineering completed
Unlocked: Automated Factory, ...

[ Choose Next Research ] [ Continue ]
```

- Content comes from a player-safe resolution summary/event projection.
- "Choose Next Research" navigates to existing Research UI; selection remains the existing authoritative Planning order.
- Creative/Uncreative behavior remains projected by server choices.

## 6. Save / Load contract

Expose persistence from the shell/menu without cloud-slot scope.

### Save

"Save Game" exports the current live snapshot and downloads a local JSON save file with a stable MOOX filename containing game ID/turn when available. No browser-created substitute state.

### Load

Single user-facing "Load Game" flow:

1. choose a local JSON file;
2. parse only enough metadata to determine game ID/schema for presentation; server remains final validator;
3. if the matching game is already hosted, explicitly ask to **Restore this game**;
4. otherwise Import as a hosted game;
5. never overwrite an existing hosted game implicitly;
6. after success navigate to the loaded game and wait for/refetch the authoritative snapshot.

### Restore confirmation mockup

```text
LOAD SAVE
This save belongs to game-1, Turn 12.
The currently hosted game-1 will be replaced.

[ Restore game-1 ] [ Cancel ]
```

## 7. Connection/state contract

A small global lifecycle indicator plus blocking behavior where necessary.

- Live: normal mutation.
- Refreshing: brief indicator; mutation gated until fresh snapshot.
- Offline/Reconnecting: stale snapshot may remain inspectable but is **read-only**.
- Refresh failed: visible Retry.
- Restoring/Importing: global mutation lock.
- Loaded: clear incompatible local drafts, show one-time loaded confirmation.

No required hover; mobile gets the same lifecycle semantics.

## 8. API error contract

Client introduces structured `APIError(status, code, message)`. UI branches on code, never on substring parsing. The raw authoritative server message remains visible in an expandable/detail line where useful.

For stale/session rejection:

1. display the rejection;
2. refetch;
3. never silently resubmit a mutation;
4. preserve only local planning intent that can safely remain a draft;
5. clearly report any dropped/incompatible draft after phase/load change.

## 9. Result/victory contract

Completed game gets a dedicated result surface, not only an inline card.

Minimum:

- winner public identity;
- completed turn;
- eliminated empires;
- Main Menu action;
- optional read-only Galaxy inspection.

No mutating End Turn or strategic actions in `completed`.

## 10. Minimum read/client additions

Freeze these additions for Gate 3:

1. TypeScript `ColonyBaseResolution` and `BattleDecision` matching existing DecisionView JSON.
2. Full player-safe Battle Spec/View TypeScript fields needed by Battle Entry/Return.
3. Public empire identity directory (ID/name/race ID only).
4. Player-safe recent resolution/transition summary with stable IDs for research/battle/major-result presentation.
5. Persistence API wrappers for export/import/restore.
6. Generic immediate-command wrapper or dedicated Colony Base helper using current revision.
7. Structured client `APIError`.
8. Explicit connection/lifecycle state separate from the current free-form status string.

15.5 may add Tactical-specific state/action types later; 15.4 must not pre-freeze movement/fire legality.

## 11. Responsive contract

- 320 CSS px remains the floor.
- Blocking decisions must fit without horizontal document scroll.
- Primary actions are at least 44px touch targets.
- Encounter side-by-side comparison may stack vertically on compact mobile.
- Save/load file controls must be touch-safe and not require drag/drop.
- Last known stale snapshot can be inspected on mobile while reconnecting, but mutation stays locked.

## 12. Gate-3 implementation order after approval

1. client types + structured API error + lifecycle state;
2. persistence wrappers and Save/Load UX;
3. Colony Base blocking decision;
4. Invasion rich decision;
5. resolution-summary projection + Research/result feedback;
6. Battle Entry/Return route/shell;
7. browser/server regressions and desktop/320px QA.

## Gate-2 decisions requested

Approve or revise:

- blocking-decision vs transition-summary model;
- stable Battle route/handoff to 15.5;
- Colony Base / Invasion layouts;
- automatic Research breakthrough presentation;
- local-file Save/Load/Restore semantics;
- read-only stale view while reconnecting;
- structured error/retry behavior;
- minimum new player-safe reads.
