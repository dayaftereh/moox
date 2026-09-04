# Slice 15.2 Gate 2 - functional strategic HMI authority contract

Date: **2026-09-04**

Status: **Gate 2 complete; Gate 3 implementation pending**.

Gate-1 input: `docs/research/SLICE_15_2_FUNCTIONAL_HMI_GATE1_2026-09-04.md`.

Accepted product direction: `docs/research/SLICE_15_2_STRATEGIC_HMI_PRODUCT_DIRECTION_2026-09-04.md`.

## Gate-2 decision

The player accepted the Gate-1 architecture and refined the functional UX. This document freezes the state/API/UI authority contract for Slice 15.2 before implementation.

The governing rule remains: React may own interaction state and a reversible Planning draft, but it never owns gameplay legality, economic formulas, build/research timing formulas, transfer legality or hidden state. All displayed strategic calculations come from authoritative server projections or pure non-mutating server previews.

## 1. Atomic player HMI snapshot

The browser continues to consume one revision-coherent player snapshot and the existing WebSocket invalidation/change-sequence lifecycle.

Gate 3 will extend the player snapshot additively with the player-safe strategic/decision information already produced by `GameSession.DecisionView`:

- strategic Galaxy/system projection;
- own Ships, ShipDesigns, Fleets, Outposts and active Population transfers;
- compact player-safe foreign strategic contacts;
- legal Research, Construction, Population, FleetMove, Colonization, OutpostDeployment and Diplomacy choices;
- human-required decision state where applicable.

The existing `view` and Battle fields remain compatible. The browser does not receive `ObserverView`.

Host generation must keep `View`, strategic data, decisions, Battles, revision and change sequence coherent under the same hosted-game lock.

## 2. Local Planning draft, authoritative submission

During `Planning`, browser interactions edit a local presentation-only draft. The authoritative GameSession does not change on each tap/drag.

The draft may contain multiple simultaneous plans across Colonies, Research and Fleets. At `Runde beenden / End Turn`, the browser serializes the effective draft as one existing revision/turn-bound `protocol.CommandBatch` with command sequences `1..N`.

Rules:

- editing the same semantic setting replaces the earlier draft value rather than stacking contradictory commands;
- the exact effective ordered draft used for preview is the one submitted at End Turn;
- no optimistic authoritative success state is shown;
- while submission + Host AI driving runs, the shell shows a blocking `Runde wird aufgelâ”œÃ‚st / KI denkt ...` state and prevents duplicate submission;
- receipt/WebSocket invalidation causes authoritative refetch;
- stale/rejected submission does not silently apply anything and triggers refresh/reconciliation.

## 3. Pure Planning preview API

The user requires build/research/output times to update immediately when Population jobs or queue plans change.

React will **not** reproduce MOO2 formulas. Gate 3 will add a pure, non-mutating planning-preview surface that accepts the current base revision plus effective draft commands, clones authoritative state, validates/applies previewable planning intent and returns derived values without advancing the turn, incrementing change sequence or publishing invalidations.

Preferred HTTP shape:

`POST /api/v1/games/{gameID}/seats/{seatID}/planning-preview`

The exact Go envelope name is implementation detail, but the contract is frozen:

- request is seat/base-revision bound;
- stale/illegal draft returns a normal structured rejection;
- preview never mutates the hosted GameSession;
- response identifies the same base revision;
- browser may debounce rapid drag/tap changes, but visible values update automatically after every effective draft change.

### Required preview outputs

At minimum the preview returns authoritative projected values needed by the UI:

Empire/global:

- Treasury balance;
- modeled net BC income per turn;
- Freighters total;
- Freighters used for Food;
- Freighters reserved for Population transfer;
- Freighters available;
- Command Points used/capacity and overage indication;
- total projected Research RP per turn.

Per Colony:

- projected Farmers / Workers / Scientists;
- Food, Production PP and Research RP output;
- population/growth/capacity values already modeled by Core/Game;
- food-safe/starvation indication;
- current construction progress;
- ordered build queue;
- current-rate estimated build completion turns for the queue.

Research:

- active field/application selection;
- progress RP / authoritative field cost;
- projected empire RP per turn;
- current-rate estimated turns to completion.

Population transfer:

- legal source/destination/job action;
- Population amount;
- Freighters required/reserved;
- same-system flag;
- ETA turns;
- destination-capacity legality.

`estimated_turns` is explicitly a **current-plan/current-rate estimate**, recalculated after every draft edit and after every authoritative turn refresh. It is not a promise that future Population growth, newly completed buildings, events or other future state changes will leave the rate unchanged.

This gives the player the immediate local-feeling UI requested while preserving one authoritative formula implementation.

## 4. Persistent orbital bodies and star metadata

The accepted star-system dialog requires bodies that currently disappear after New Game generation.

Gate 2 freezes a real authoritative orbital-body model for 15.2.

Every generated visible orbit body required by the current MOO2 baseline must be persistent and have stable identity. The minimum player-facing body kinds are:

- normal colonizable `planet`;
- `gas_giant`;
- `asteroid_belt`.

The implementation must verify the original body-type table mapping before assigning those names to generator body types. Deferred special/wormhole materialization must not be guessed.

Star/system spectral/class metadata currently transient in New Game also becomes persistent/projectable so the system dialog/map has authoritative star identity/color metadata. Final artwork remains Slice 15.3.

Existing colonizable Planet environmental properties (size/mineral/gravity/climate) remain authoritative only for normal planets.

## 5. Outposts on planets, gas giants and asteroid belts

This is now **in Slice 15.2**, not deferred.

Original MOO2 manual page 77 explicitly permits Outposts on gas giants and in asteroid belts. Local extracted original help (`reference/original/text/manifest.json`, Outpost Ship/help records) also describes Outpost Ships and the Outpost placement workflow. The current MOOX `Outpost.PlanetID`/`DeployOutpostPayload.PlanetID` model is therefore too narrow.

Gate 3 must generalize Outpost targeting to stable orbital-body identity while retaining save migration/compatibility.

Frozen behavior:

- an Outpost Ship may establish an Outpost on a server-projected legal unoccupied normal planet, gas giant or asteroid belt;
- legal targets come only from the authoritative OutpostDeployment choice catalog;
- the body/system displays the Outpost after resolution;
- the Outpost supplies its normal strategic range/system role independent of body habitability;
- a normal planet with a friendly Outpost can still later convert to a Colony through the existing supported path;
- a gas giant/asteroid Outpost does **not** become a Colony without future Artificial Planet mechanics;
- future Artificial Planet work must be able to preserve/convert body identity rather than requiring a presentation-only fake body.

Colonization remains legal only on normal server-projected colonizable planets.

## 6. Galaxy / system interaction

Current Slice-15.2 knowledge baseline remains intentionally no-fog/all-public because no authoritative discovered/known-system state exists yet.

Galaxy:

- authoritative X/Y coordinates;
- pan/zoom;
- stars/systems and player-safe contacts;
- Fleet selections/routes;
- tap/click star opens in-map system dialog.

System dialog:

- central authoritative star;
- all projected persistent orbital bodies by orbit;
- Colony/Outpost presence through player-safe projection;
- present own Fleets/Ships and public contacts;
- legal Planet/Body actions only from server choice catalogs.

No React-inferred scouting state.

## 7. Colonize confirmation

Frozen flow:

1. Galaxy -> System dialog -> normal planet;
2. match selected Planet against projected legal Colonization choices;
3. show `Kolonisieren` only when legal;
4. confirmation identifies system/planet and source Colony-Ship Fleet;
5. confirm adds `empire.colonize_planet` to the Planning draft;
6. UI marks the action as planned/pending only;
7. actual Colony appears after End Turn / strategic resolution / authoritative refetch;
8. new owned Colony can open the canonical Colony Detail screen.

## 8. Colony management table

The Colonies screen remains a real table, including on mobile. A contained table viewport may scroll horizontally; the page itself must not require horizontal scrolling.

Minimum columns/content:

- Colony/System name - opens canonical Colony Detail;
- total Population;
- Nahrung/Farmers - assigned Population plus Food output;
- Produktion/Workers - assigned Population plus PP output;
- Forschung/Scientists - assigned Population plus RP output;
- current build item + estimated turns;
- compact status/warnings where needed.

The table is sortable by useful visible values.

### Population drag/tap semantics

Same Colony row:

- moving a Population unit between Farmer/Worker/Scientist columns is a local job reassignment;
- no confirmation dialog;
- effective draft `colony.assign_population` is updated;
- Planning preview automatically recalculates that Colony plus empire-wide dependent values such as total Research, Food/Freighter usage and ETAs.

Different Colony row:

- moving a Population unit to another Colony is Population transfer;
- before adding the transfer to the draft, show confirmation with source, destination, source job, target job, Freighters required and ETA;
- server legal preview controls whether Confirm is enabled;
- same-system transfer shows zero interstellar Freighter requirement where authoritative rules say so;
- cross-system transfer shows reserved Freighters + travel time.

Drag-and-drop is never the only control. Mobile also supports source-tap -> destination-tap and direct +/- job controls.

## 9. Population transfer destination job

Current MOOX transfer mechanics preserve the source job because the payload/state carries one `Job`. That is insufficient for the accepted Colony-table UX, where dropping onto the destination Production/Farming/Research cell should determine the job at arrival.

Gate 3 will extend the authoritative transfer model with separate source and destination job semantics.

Minimum command/projection meaning:

- SourceColonyID;
- DestinationColonyID;
- source cohort;
- SourceJob;
- DestinationJob;
- one Population unit per transfer action in the initial browser workflow.

For same-system transfer, the unit is added to DestinationJob during resolution. For interstellar transfer, DestinationJob is persisted with the transit and used on arrival.

All existing capacity, minimum-source-Population, Freighter and ETA rules remain server-side.

## 10. Real Colony construction queue

Gate 2 chooses the real queue, not the current single-project limitation.

Original MOO2 manual page 64 has a finite seven-slot queue. The player deliberately chooses a MOOX modernization: **no user-visible item-count limit**. This is an intentional divergence from original UI capacity, not an uncertain parity claim.

Protocol/body-size safety limits may still bound pathological requests, but there is no product rule such as 7 or 10 queue entries.

### Authoritative queue state

A Colony gains an ordered persistent construction queue. Each entry identifies the existing authoritative project kind/ID and, where relevant, ShipDesign ID/revision.

The head entry is the active project. Remaining entries are future work.

Gate 3 should prefer one atomic queue-edit Planning command (for example `colony.set_construction_queue`) over a long list of UI-specific append/remove/reorder commands. Existing construction commands may remain as compatibility/internal helpers if useful.

Server validation applies to every entry. React never decides whether a building/ship/project is buildable.

### Production progress semantics

The queue must support original-like production stock behavior:

- accumulated production is Colony-level authoritative production progress, not disposable browser state;
- reordering/changing the top project does not silently throw away already accumulated PP;
- when an item completes, valid overflow production carries into subsequent queue entries in order during the same resolution until production is exhausted or an ongoing/non-completing process becomes the head;
- saved state round-trips queue order and progress exactly.

Any parity nuance exposed during Gate-3 implementation must be resolved in Game/Core, never in React.

### Construction editor UX

From Colony Detail or the Colony's system-body view, tapping/clicking current construction or `Neues bauen / Change` opens the functional construction editor.

Desktop:

- left: searchable/grouped list of all currently legal buildables;
- right: ordered queue;
- tap/click legal item -> append;
- drag/reorder queue;
- remove entry;
- current head visibly distinct with progress and ETA.

Mobile:

- same semantic lists using tabs/stacked panes/sheets as needed;
- tap operations are complete without precision drag.

Leaving the editor keeps the **Planning draft** queue. It does not submit the turn.

Normal Colony Detail shows the current build prominently plus upcoming queue items; Colony table shows at least current build + ETA.

## 11. Build/research ETA live behavior

The user explicitly wants timing to react to Population job changes during the current Planning turn.

Frozen behavior:

- changing Workers immediately updates projected Colony PP and build ETA;
- changing Scientists on any owned Colony immediately updates empire projected RP/turn and active Research ETA;
- Farmer changes may indirectly alter Food/Freighter availability and therefore other preview status;
- queue edits immediately update each entry's sequential current-rate ETA;
- all values update again after authoritative turn advancement.

Research ETA is based on authoritative remaining RP and projected empire RP/turn. Build ETA is based on authoritative remaining/queued PP and projected current Colony PP rate. Zero-rate cases show no finite ETA instead of dividing/client-guessing.

## 12. Persistent global resource strip

The running-game shell gains an always-reachable resource summary in addition to Turn/Phase/connection state.

Minimum permanent values:

- BC Treasury balance plus modeled net income/turn;
- Freighters available / total, with detail for Food usage and Population-transfer reservation;
- Command Points used / capacity, with over-cap warning.

Desktop: visible in the sticky top/status region.

Mobile: compact sticky resource chips/second row without causing page-wide horizontal overflow; details may open on tap.

Research RP/turn is required on Research and relevant planning previews, but is not mandated as a permanently visible fourth global chip.

## 13. Colony Detail canonical screen

Both entry paths resolve to the same GameID + ColonyID identity:

- Galaxy -> System -> owned normal Planet -> Colony Detail;
- Colonies table -> Colony Detail.

Functional layout:

- top: Population jobs with the same draft/preview behavior as the table;
- planet/environment/status + authoritative Colony economy;
- Buildings/infrastructure;
- current construction + queue summary;
- `Neues bauen / Change` -> construction editor;
- desktop queue/build summary on the right, mobile below/sheet as appropriate;
- back navigation preserves useful Galaxy/system context where entered from Galaxy.

## 14. Fleets

Frozen from Gate 1:

- own Fleet/Ship tiles grouped by authoritative system/location;
- separate in-transit group;
- FleetID is the same identity on Galaxy and Fleets screens;
- movement targets, ETA/range and split subset choices only from server legal choices;
- Colonize and Outpost actions only from their catalogs;
- foreign contacts never expose hidden composition.

## 15. Research

Research uses eight player-facing categories:

1. Construction
2. Chemistry
3. Computers
4. Physics
5. Power
6. Sociology
7. Biology
8. Force Fields

Gate 3 must normalize category ID/order/name-key and map all field chains server/ruleset-side. React must not hard-code field-ID/category relationships.

Existing server selection modes remain the sole normal/Creative/Uncreative authority:

- `choose_one`;
- `all`;
- `fixed_one`;
- `repeat_field`.

Research screen shows current progress, projected empire RP/turn and Planning-preview ETA that changes immediately when Scientist assignments change anywhere in the empire.

## 16. Diplomacy

Slice 15.2 freezes the existing authoritative baseline only:

- Neutral / Peace / War;
- incoming/outgoing peace offer;
- Declare War;
- Offer Peace;
- Accept Peace.

Broader treaties/trade/tribute/technology exchange remain later work.

## 17. Espionage -> dedicated Slice 20

The player explicitly requests a separate gameplay slice for Espionage rather than a partial 15.2 implementation.

A new prepared `PLANNED_20_ESPIONAGE_INTELLIGENCE_BASELINE.md` owns the mechanics.

Slice 15.2 only reserves the accepted first-class navigation/product location and presents an honest unavailable-baseline state. No fake Spy controls are permitted.

The number 20 is intentionally reserved for this future work even though Slices 18/19 are not currently specified.

## 18. Display metadata / localization

All new player-visible structural copy remains DE/EN under the Slice-15.1 i18n layer.

Server choice/projection records must carry enough stable display metadata (IDs/name keys/category keys/costs) for React to present buildables, technologies, bodies and actions without reconstructing rule databases or exposing observer state.

Proper/generated system/planet/ship names remain data, not UI translations.

## 19. Minimum Gate-3 implementation acceptance

Before Gate 3 can be considered implemented, automated/manual evidence must cover at least:

1. player snapshot exposes strategic + decision data atomically without foreign-detail leaks;
2. 2D Galaxy uses authoritative coordinates;
3. generated gas giants/asteroid belts survive persistence and render in the system dialog;
4. an Outpost can be legally projected/deployed on a gas giant and asteroid belt and affects normal Outpost strategic range;
5. Colony Ship colonization remains limited to legal normal planets;
6. construction queue persists order/progress and accepts more than the original seven entries without an item-count rejection;
7. queue add/remove/reorder comes from legal build choices;
8. same-row Colony job movement updates previewed Food/PP/RP and ETAs without incrementing hosted Game revision/change sequence;
9. Scientist change on one Colony updates empire Research RP/turn + Research ETA;
10. Worker change updates that Colony build ETA;
11. cross-Colony Population drag/tap produces legal transfer confirmation with Freighters/ETA and target job, and confirmed command survives transit/arrival;
12. global sticky shell shows BC/net income, Freighters and Command Points;
13. Colony table -> Colony Detail and Galaxy -> System -> Planet -> same Colony Detail identity;
14. Fleet movement/Colonize/Outpost choices are catalog-driven;
15. eight Research categories and Creative/Uncreative modes are data/server driven;
16. War/Peace Diplomacy works from the first-class screen;
17. Espionage has no actionable fake controls and points to future mechanics scope only;
18. End Turn submits the full effective draft once, drives built-in AI to the next human boundary and refetches authoritative state;
19. reload/save-resume preserves bodies, Outposts, construction queue/progress and normal strategic state;
20. DE/EN + 15.1 desktop/mobile shell remain functional.

## 20. Intentional MOO2 divergence recorded in Gate 2

The original MOO2 manual documents a **seven-slot** construction queue. MOOX intentionally removes the user-visible queue-length cap and presents an ordered list limited only by normal protocol/system safety constraints.

This modernization does not change the one-active-head production semantics; it removes a UI capacity restriction.

## Gate-2 conclusion

Gate 2 is complete and the implementation shape is frozen.

Gate 3 may now implement the accepted strategic HMI plus the authoritative data/model additions required by that HMI: player DecisionView transport, pure Planning preview, persistent system bodies/star metadata, generalized Outposts, real construction queue, Population-transfer target-job/legal preview, Research category metadata and the responsive Colony/Galaxy/Fleet/Research/Diplomacy workflows.

Espionage mechanics are explicitly outside Gate 3 and move to prepared Slice 20.

## Gate-2 verification

- Original manual cross-check: Outpost on gas giants/asteroid belts confirmed; original construction queue documented as seven slots.
- `git diff --check` - PASS before commit.
- Live development preview `/healthz` on port 7171 - PASS after Gate-session handoff.
- Gate 2 is documentation/contract freeze only; no production gameplay implementation changed.
