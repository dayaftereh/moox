# Open slice 10 - Diplomacy, war and peace baseline

Opened: 2026-09-02
Status: **Gate 4 technical QA passed; implementation commit + closeout in progress**.

Planned specification: `docs/slices/PLANNED_10_DIPLOMACY_WAR_PEACE_BASELINE.md`
Permanent evidence: `docs/research/DIPLOMACY_WAR_PEACE_BASELINE_2026-09-02.md`
Starting HEAD: `9b2641d` (`docs: close deterministic new game slice`)

## Recovery checklist

- [x] Slice 09 confirmed closed and repository clean at open.
- [x] Exactly one Slice-10 `_OPEN_` marker created.
- [x] Gate 1: current-code/original-evidence checkup and Gate-2 contract proposal.
- [x] Gate 2: accept/freeze diplomacy v1 contract.
- [x] Gate 3: implementation + regressions + server/web proof.
- [ ] Gate 4: final QA + commit + close.

## Scope guard

Do not implement diplomacy state or commands during Gate 1. Gate 1 must first re-check the current directed hostility/blockade/encounter authority, original MOO2 war/peace transition timing, attack authorization/sneak-attack behavior and replay/UI projection needs. Present an exact narrow Gate-2 contract before implementation.

## Gate-1 result

- Original normal-Empire war is relation value `4`, written symmetrically by `Declare_War_`.
- Original peace is relation value `3`, written symmetrically by `Declare_Peace_` after a successful peace proposal.
- Ordinary attack authorization accepts relations `4..6`; relation `0` has a distinct sneak-attack/AI path and is deferred.
- War/accepted-peace transitions are immediate diplomacy actions and do not cancel existing Fleet movement orders.
- The exact proposed Core22 / command / immediate-transport / event / projection / same-system contract is frozen as a Gate-2 proposal in `docs/research/DIPLOMACY_WAR_PEACE_BASELINE_2026-09-02.md`.

Gate 2 accepted/froze the contract on 2026-09-03. Resume at Gate 3 implementation only; do not widen scope beyond the accepted contract.
## Gate-2 result

- Core schema 22 accepted for canonical `neutral`/`peace`/`war` semantics plus directional pending peace offers.
- Normal `peace`/`war` transitions are reciprocal/symmetric; storage stays directionally keyed.
- Commands frozen: `diplomacy.declare_war`, `diplomacy.offer_peace`, `diplomacy.accept_peace`.
- Immediate diplomacy is revision-bound and allowed only in `planning` **before any seat has submitted the current turn**; the first End-Turn submission closes diplomacy until the next planning phase.
- War alone authorizes normal strategic attacks/blockades/encounters; neutral/peace do not. Sneak attacks remain deferred.
- Event schema 1 and player-safe diplomacy projection accepted; broader treaties/AI/espionage/special hostile values and automatic peace expiry remain deferred.
- Gate 3 must implement the accepted contract and its same-system/movement/BattleSession proofs; no gameplay implementation occurred in Gate 2.
## Gate-3 result

- Core schema 22 now persists canonical reciprocal `peace|war` relations plus sorted directional pending peace offers; missing relation remains neutral.
- `diplomacy.declare_war`, `diplomacy.offer_peace` and `diplomacy.accept_peace` are server-authoritative immediate commands with strict payloads, deterministic events and exact `base_revision`.
- Planning diplomacy is rejected after the first current-turn submission and in later phases; a materialized tactical Battle remains committed.
- `MayAttackEmpire` is the single normal-Empire attack authority consumed by blockade and strategic encounter derivation; only `war` authorizes attack.
- Player snapshots expose only own diplomacy state/offers; Observer retains complete state. The React HMI can declare war, offer peace and accept an incoming offer using the immediate-command endpoint.
- Regression proofs cover Core22 roundtrip/noncanonical rejection, byte-identical deterministic lifecycle replay/events, stale revision, first-submission lock, neutral same-system -> war encounter, accepted peace suppressing the next encounter while preserving Fleets, and encounter-phase rejection.
- Gate-3 full checks passed: `go test ./... -count=1`, `go vet ./...`, `npm --prefix web run build`, and `git diff --check`.

Resume at Gate 4 for final independent QA, status/history update, commit and Slice close. Do not widen into deferred treaties/AI/sneak attacks/espionage.
## Gate-4 technical QA - 2026-09-03

- Focused `Diplomacy|Blockade|Encounter|TacticalBattle|NewGameGoldenSeedStateFingerprint` coverage across Core/Game/Session/Server passed five consecutive runs.
- `TestDiplomacyLifecycleDeterministicReplay` passed ten consecutive runs and reproduces byte-identical Core22 serialized state plus DeepEqual event sequences.
- Core22 save/load validation covers reciprocal war/peace and directional pending peace offers; noncanonical/asymmetric state is rejected.
- Authority scan found zero unexpected production touches of `DiplomaticRelations` / `DiplomaticPeaceOffers` outside Core state/query/validation and the Game diplomacy resolver. Production command path is Server -> Host -> Session -> Game resolver; acting Empire is derived from Seat authority.
- Web mutation transport has one central `fetch()` wrapper in `web/src/api.ts`; the Diplomacy UI only renders projection fields and submits server commands, with no local authoritative state mutation.
- Independent full Gate-4 checks passed: `go test ./... -count=1`, `go vet ./...`, `npm --prefix web run build`, `git diff --check`.

Technical QA is complete. Gate 4 remains open only for the implementation commit, final HISTORY/status synchronization and deletion of this OPEN marker.