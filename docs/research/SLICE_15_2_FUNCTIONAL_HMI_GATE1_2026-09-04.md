# Slice 15.2 Gate 1 - functional strategic HMI authority/projection audit

Date: **2026-09-04**

Status: **Gate 1 complete; Gate 2 contract freeze pending**.

## Objective

Audit the accepted Slice-15.2 strategic browser product direction against the authoritative Go/Core/Game/Session/App/HTTP implementation. The goal is to let Gate 2 freeze the exact player-safe reads, legal actions and command envelopes required for an interactive 2D Galaxy, star-system dialog, Colony table/detail, location-grouped Fleets, eight-category Research and Diplomacy without moving game rules into React.

Permanent product input: `docs/research/SLICE_15_2_STRATEGIC_HMI_PRODUCT_DIRECTION_2026-09-04.md`.

## Executive result

The backend is much closer to Slice 15.2 than the current React HMI suggests.

`internal/session/decision_view.go` already contains `PlayerDecisionView`, explicitly documented as the complete input boundary for built-in AI and future strategic HMI clients. It already projects:

- own Empire and Colonies;
- Galaxy and materialized planets;
- own Ships, ShipDesigns, Fleets, Outposts and active Population transfers;
- player-safe foreign strategic contacts;
- legal Research, Construction, Population, Fleet movement, Colonization, Outpost and Diplomacy choices;
- Colony-Base/Invasion/Battle decision information.

The main transport gap is that normal HTTP `PlayerSnapshot` still exposes only the narrower `session.PlayerView` plus Battle views. Gate 2 should therefore prefer exposing the existing player-safe DecisionView data atomically at the same revision/change-sequence boundary instead of inventing per-screen legality APIs.

Four real data/mechanics gaps cannot be solved honestly in React:

1. non-colonizable orbital bodies and useful star/spectral metadata are not persisted;
2. Colony construction currently has one active project, not an ordered multi-item future queue;
3. Population transfer exists as mechanics but has no server-projected legal target catalog;
4. Espionage has modifiers but no gameplay state/rules/commands and needs a dedicated later mechanics slice.

## Current browser / HTTP surface

The browser currently wraps New Game, game listing, narrow player snapshot, `colony.assign_population`, Diplomacy immediate commands, Invasion immediate commands and WebSocket invalidation.

The player snapshot route is:

`GET /api/v1/games/{gameID}/seats/{seatID}/snapshot`

Go `session.PlayerView` already sends a full own `core.Empire` and full own `core.Colony`, but `web/src/api.ts` intentionally types only a small subset. Already-delivered but currently untyped/unrendered Colony/Empire fields include:

- Treasury, Command Points and Food Logistics;
- known Technology/Field IDs and current Research state;
- Buildings;
- Economy, AdjustedEconomy and EconomyContext;
- PopulationDynamics;
- current Construction project/progress.

Thus much of Colony Detail and Research summary needs TypeScript/UI expansion, not new browser calculations.

## PlayerDecisionView safety / authority

`StrategicView` already carries full Galaxy coordinates, own Outposts/ShipDesigns/Ships/Fleets/PopulationTransfers and compact foreign contacts.

Its construction deliberately strips tested foreign/internal information:

- foreign full Fleet/Ship composition is not exposed;
- foreign Colony internals are not exposed;
- raw planet ColonyID/OutpostID occupancy internals are zeroed;
- internal blockade Empire IDs are removed.

`internal/session/decision_view_test.go` already verifies this player-safe behavior, deterministic projection and deep-clone/non-mutating semantics.

`DecisionCatalog` already projects server-authoritative choices for Research, Construction, Population, FleetMoves, Colonization, OutpostDeployment and Diplomacy. These catalogs should be the primary legality source for the 15.2 browser.

## Strategic command map

### Planning batch

Existing planning commands cover:

- `colony.assign_population`;
- `empire.select_research`;
- all current Colony construction families;
- `colony.transfer_population`;
- `empire.move_fleet`, split/merge;
- `empire.colonize_planet`;
- `fleet.deploy_outpost`.

`protocol.CommandBatch` is revision- and turn-bound and requires command sequence `1..N`. `GameSession.SubmitTurn` only accepts it in Planning. Strategic resolution occurs after submissions and the Host drives built-in AI/non-human phases to the next human interactive boundary.

Gate-2 proposal: React owns only a presentation Planning draft. UI gestures add/replace draft commands; `Runde beenden` submits one complete CommandBatch. No server state is considered changed until the authoritative response/refetch succeeds.

### Immediate commands

Existing revision-bound immediate authority covers War/Peace Diplomacy, Invasion and Colony-Base resolution. These remain separate from the Planning draft.

Tactical commands remain Slice 15.5.

## Galaxy / knowledge audit

Core already persists authoritative StarSystem X/Y coordinates, so React need not invent map placement.

The current supported strategic baseline is explicitly **no-fog**. There is no persisted discovered/visited/known-system gameplay state. Slice 15.2 therefore must not infer discovery from Fleet locations or browser history.

Gate-2 proposal:

- current Small vertical slice uses the authoritative public Galaxy as the known map;
- projection types remain extensible for later visibility metadata;
- actual exploration/fog knowledge mechanics stay future scope.

New Game generation calculates a transient spectral class, but `core.StarSystem` does not persist it. A neutral functional star can be drawn now; meaningful original-like star class/color requires authoritative persisted/projected metadata.

## Star-system dialog / orbital body gap

Slice-09 evidence and current code confirm that original-derived body-type RNG is consumed, but only body type 3 is materialized as `core.Planet`. Non-colonizable body types are discarded. Gas giants/asteroid belts therefore do not currently exist in authoritative saved state.

React must not manufacture them.

Gate 2 must decide whether to accept a bounded Core/state addition that persists non-colonizable orbital-body kind + orbit and likely star/spectral class. Exact original body-type-to-visible-kind mapping must be verified before naming the stored kinds. No gas/asteroid colonization mechanics are required for 15.2.

If this changes state schema, deterministic New Game/save/load/round-trip regressions are mandatory.

## Colonization flow

`DecisionCatalog.Colonization` already projects exact legal `FleetID + SystemID + PlanetID` tuples. The browser must show Colonize only when such a choice exists, never because it merely sees a Colony-Ship-looking Fleet.

The current authoritative command is a **Planning command**. Therefore the accepted confirm flow should be:

1. Galaxy -> System dialog -> Planet;
2. match projected legal Colonization choice;
3. show `Kolonisieren` only if legal;
4. confirm planet/system/source Fleet;
5. add `empire.colonize_planet` to the Planning draft;
6. show it as planned/pending, not already completed;
7. actual Colony appears after normal turn resolution/refetch;
8. then navigate/offer the canonical Colony Detail.

## Colonies / Colony Detail

The current player-safe data is already rich enough for a dense Colony table and detailed owned-planet page: Population, Buildings, Economy, PopulationDynamics and current Construction state are authoritative.

### Job controls

`colony.assign_population` is authoritative. `DecisionCatalog.Population` additionally carries server-derived preview choices with Farmers/Workers/Scientists, AdjustedEconomy, PopulationDynamics and FoodSafe.

The existing preview frontier is AI-oriented and does not enumerate every arbitrary Worker/Scientist split. Gate 2 should therefore distinguish:

- editing job counts/dragging: safe as local Planning input;
- showing arbitrary live Food/PP/RP forecasts: needs a pure server-side preview if required.

React must not reproduce economy formulas.

### Population transfer

`colony.transfer_population` already enforces source/destination ownership, one Population unit, cohort/job selection, source minimum population, destination capacity, same-system handling, interstellar Freighter reservation/ETA and active-transfer limits.

`StrategicView` projects active own transfers, but DecisionCatalog does not project legal transfer choices. If 15.2 exposes this workflow, Gate 2 should add a player-safe legal transfer-target/action projection rather than teaching React those rules.

### Canonical route

`Galaxy -> System -> owned Planet -> Colony Detail` and `Colonies table -> Colony Detail` must converge on the same GameID/ColonyID route. Returning from Galaxy should restore the previous map/system context where practical.

## Construction / build queue

Current Core has `Colony.Construction *ConstructionState`: one active project only. `AvailableConstructionChoices` returns no new choices while a project exists. Despite `queue_*` command names, no ordered future queue exists.

The accepted UX asks for current construction plus a right-side build queue, and the local MOO2 reference describes a build queue/current project.

Gate 2 must choose explicitly:

- **A:** display only the current authoritative project and choose a new one when idle; true queue deferred; or
- **B:** add a bounded authoritative ordered queue with append/reorder/remove semantics before presenting a real queue UI.

Product-fidelity recommendation is B if bounded; otherwise A must be visibly documented. Never emulate an authoritative queue only in React.

## Fleets

Strategic DecisionView already supports the desired Fleet screen:

- own full Fleet identities and ShipIDs;
- own concrete Ships/ShipDesigns;
- AtSystem/Destination/RemainingTurns/role/special kind;
- public foreign Fleet contacts without composition;
- legal FleetMove choices with destination, ETA, fuel/supply data and optional subset ShipIDs;
- separate legal Colonization/Outpost choices.

The UI can therefore group stationary Fleets by system and moving Fleets as in-transit without client legality logic. Galaxy and Fleets pages must cross-link by the same FleetID.

## Research

`game.ResearchChoice` already projects field IDs, previous/next links, RP cost, selection mode, Technology IDs/keys/name keys and Hyper-Advanced levels.

Server selection modes already represent normal/Creative/Uncreative behavior (`choose_one`, `all`, `fixed_one`, plus repeat-field). React therefore needs no race-trait Research legality.

The missing piece is player-facing **eight-category metadata**. Runtime TechnologyField data has IDs/links/cost/AIGroup but no normalized category ID/name.

Local reference confirms eight categories: Construction, Chemistry, Computers, Physics, Power, Sociology, Biology, Force Fields. There are eight root field chains. Chemistry root 22 is explicitly backed by existing evidence; the remaining root/category associations should be verified and normalized server/ruleset-side rather than hard-coded in React.

Gate 2 should freeze stable category IDs/order/name keys and field-chain membership.

## Diplomacy

Current 15.2 baseline is sufficient for a first-class screen:

- other Empire identity;
- Neutral/Peace/War stance;
- incoming/outgoing peace offer;
- legal Declare War / Offer Peace / Accept Peace actions.

Broader treaties/trade/tribute/technology exchange remain deferred.

## Espionage decision

Local MOO2 reference describes defensive/offensive Spy assignment, espionage and sabotage, affected by race and computer technology.

Current MOOX has normalized `SpyingBonus` plumbing (including Darlok +20), but no Spy gameplay state, counts, allocation, mission targets, command family, success/detection/capture rules, events or AI policy.

**Gate-1 decision:** actionable Espionage does not belong inside 15.2 implementation. It requires a dedicated later mechanics slice. 15.2 reserves the first-class navigation/product area and may show a clear non-actionable unavailable-baseline state, but must expose no fake controls.

## Built-in AI / busy state

`Host.mutate()` already runs `driveToInteractiveBoundary()` after successful mutations. It drives built-in AI Planning, Strategic Resolution and AI Encounter/Invasion/PostResolution decisions and stops at the next human-required boundary.

No browser `Run AI` API is needed.

Gate 2 should freeze:

- blocking `Runde wird aufgel├Âst / KI denkt ...` state while the end-turn request runs;
- duplicate submission disabled;
- receipt/WebSocket invalidation -> authoritative refetch;
- human decision surface when Host stops at a human boundary;
- rejection/error -> no assumed mutation, recover/refetch.

## Missing projection summary

### Reuse/expose

- existing `PlayerDecisionView` or an equivalent atomic HMI projection;
- full own Empire/Colony fields already present but missing TypeScript types;
- Strategic Galaxy/Fleet/Ship/Outpost/contact data;
- DecisionCatalog legal choices.

### Add if accepted

- persistent/projected non-colonizable orbital bodies and optionally spectral/star class;
- normalized eight-category Research metadata;
- legal Population-transfer choices if transfer UI is included.

### Gate-2 scope choice

- real ordered Construction queue vs. current single active project.

### Explicit future mechanics

- true exploration/fog-of-war knowledge;
- actionable Espionage;
- advanced Diplomacy;
- broad Ship Designer;
- Tactical Combat.

## Preferred Gate-2 transport contract

Gate 1 recommends one revision-coherent player HMI snapshot rather than screen-specific rule calls:

- retain current PlayerSnapshot/change-sequence/WebSocket lifecycle;
- add DecisionView-derived strategic/decision data under the same Host lock/revision;
- prefer an additive compatibility path during 15.2;
- all strategic screens consume the same player-safe revision;
- DecisionCatalog remains the sole legal-action source;
- ObserverView never reaches a player client.

Only high-frequency arbitrary Colony preview, if Gate 2 wants live forecasts for every drag position, justifies a separate pure/read-only server preview endpoint.

## Functional regression strategy

Gate 3/4 should prove:

- no foreign internals leak through the HMI snapshot;
- revision/change-sequence coherence;
- body/star metadata deterministic round-trip if added;
- projected legal choices generate accepted commands while forged illegal/stale commands reject without mutation;
- Galaxy uses server coordinates and projected bodies;
- Colonize is offered only from legal catalog and confirm creates a Planning draft order;
- multiple Colony job edits coexist in one draft without auto-ending the turn;
- Colony table/detail share identity;
- Construction/Fleet/Research/Diplomacy actions come only from server catalogs;
- Research displays all eight normalized categories without race logic in React;
- Espionage exposes no fake actionable mechanics;
- DE/EN and 15.1 responsive behavior remain intact;
- Human-vs-built-in-AI end-turn reaches the next interactive boundary and reconnect shows the same authoritative state.

## Gate-2 contract to freeze

Gate 2 should explicitly freeze these points before implementation:

1. expose DecisionView-derived player-safe strategic/legal data atomically;
2. local presentation-only Planning draft -> one revision-bound CommandBatch;
3. current no-fog/all-public Galaxy baseline;
4. persistent orbital-body/star-class scope;
5. Colonize confirm -> pending Planning command -> authoritative resolution;
6. Colony table + canonical Colony Detail and server-only economy derivation;
7. true ordered construction queue vs. explicit current single-project limitation;
8. Population-transfer legal projection or explicit UI deferral;
9. Fleet grouping/actions from own state + legal catalogs;
10. normalized eight-category Research metadata and server selection modes;
11. War/Peace Diplomacy baseline only;
12. Espionage mechanics deferred to a dedicated later slice;
13. Host AI driver + blocking browser busy/recovery behavior;
14. minimum server/browser/E2E acceptance scenarios.

## Gate-1 conclusion

Gate 1 is complete. No gameplay implementation was performed. The implementation strategy for Gate 2 is to surface the already-existing player-safe DecisionView to the browser and add only the bounded authoritative data/mechanics pieces the accepted UX cannot represent today.

## Gate-1 verification

- `go test ./internal/session ./internal/app ./internal/server` - PASS.
- `git diff --check` - PASS before commit.
- No production gameplay implementation was changed in Gate 1.
