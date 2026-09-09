# Slice 15.4 Gate 3 - Block 4 rich Invasion decision UX

Date: **2026-09-09**

Status: **implemented and end-to-end verified**.

## Goal

Replace the functional but ID-heavy Invasion priority card with the accepted blocking-decision presentation while keeping all legality and ground-combat resolution on the server.

## Player-safe public empire identity

The existing Invasion projection contains attacker/defender empire IDs but the browser previously lacked a reliable player-safe empire-name/race directory. Inferring identity from seat labels would mix controller/player naming with empire identity.

`PlayerDecisionView` now includes a deliberately minimal `public_empires` projection:

- empire ID;
- empire name;
- race ID.

It exposes no Treasury, Research, known technologies, queues, command points or other hidden foreign state. The list is sorted by empire ID for deterministic JSON.

`TestPlayerDecisionViewIsPlayerSafeDeepCopyAndDeterministic` now verifies the public Darlok identity is present while existing hidden-state isolation tests remain intact.

## Rich Invasion context

The browser derives presentation context only from already player-safe projections:

- target System from `invasion.system_id`;
- foreign Colony contact matching `invasion.colony_id`;
- target Planet through the contact's public planet ID;
- defender through `public_empires`;
- eligible own Transport fleets strictly from `eligible_transport_fleet_ids`.

The blocking screen shows:

- spherical target Planet art using `OrbitalBodyArt`;
- System and Planet name;
- climate / size / mineral facts;
- defender public name and race;
- exact count of eligible troop transports;
- individual projected Fleet IDs / special kind;
- explicit note that Invade uses exactly the server-projected eligible transports;
- Invade / Decline actions.

No troop strength, combat odds or hidden formula is invented in React.

## Command behavior

The existing `submitInvasion` remains authoritative:

- Invade submits exactly all currently projected `eligible_transport_fleet_ids`;
- Decline submits the projected Colony ID.

The rich UI adds:

- synchronized-lifecycle mutation gating;
- 44px minimum action targets;
- structured rejection display;
- automatic authoritative refetch after HTTP 409 stale/session conflict.

## Real session fixture

An isolated temporary Go test used the existing `newInvasionSession` and `submitEmptyInvasionTurn` helpers to create a true `invasion_decisions` live snapshot.

Fixture:

- game ID: `browser-invasion`;
- two local-human seats;
- attacker at war with Defender;
- target: System Beta / Beta I;
- Defender public empire: `Defender`, race `human`;
- one attacker combat fleet at target;
- two eligible troop transports, Fleet IDs **16 and 17**;
- phase: `invasion_decisions`;
- revision: 2.

The fixture was marshalled through the production live-snapshot path, then the temporary test source was removed and never committed.

## Isolated browser E2E

The fixture was imported to a temporary persistence server on **127.0.0.1:7193**. Canonical 7171 was not modified during the scenario tests.

### Initial decision

Desktop browser showed:

- System: Beta;
- Planet: Beta I;
- facts: Desert / Small / Rich;
- Defender: Defender / human;
- two eligible transport chips: Fleet 16 and Fleet 17, both `troop_transport`;
- Invade and Decline buttons: 44px high;
- End Turn disabled;
- lifecycle `synced`;
- no danger notice;
- no horizontal overflow.

### Decline path

Clicking Decline through the browser:

- submitted the real server command;
- decision dialog disappeared;
- game advanced to **Turn 2 / Planning**;
- End Turn re-enabled;
- lifecycle remained `synced`;
- no client-side continuation logic was required.

### Invade path

The original fixture was atomically restored, then Invade was clicked.

The server:

- used the two projected eligible transport fleets;
- captured the Defender Colony;
- transferred the target Colony to the attacker;
- consumed the applicable transport force;
- eliminated the defender empire;
- correctly completed the game as conquest because this fixture had no remaining opponent holdings.

Verified result:

- phase: `completed`;
- turn: 1;
- revision: 4;
- winner empire: 9;
- winner seat: 1;
- eliminated empire: 11;
- target Colony now belongs to empire 9;
- one troop-transport fleet remains.

This also provides useful evidence for the upcoming Victory/result transition block.

## 320px QA

Viewport: **320x646**.

- decision sheet width: exactly 320px;
- height: about 534px for two transports;
- target art: 74x74px;
- both action buttons: 298x44px;
- document scroll width: exactly 320px;
- no horizontal overflow;
- no danger notice.

## Regression

- `go test ./internal/session -run 'DecisionView|Invasion' -count=1` PASS;
- `go test ./internal/app ./internal/server -count=1` PASS;
- `go vet ./internal/session ./internal/app ./internal/server` PASS;
- TypeScript project build PASS;
- Vite production build PASS;
- `git diff --check` PASS;
- temporary 7193 server closed and fixture removed.
