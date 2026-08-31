# Planned slice 01 - Colony Base and same-system colonization

Status: **planned / queued; not open**.

Queue position: **1 of 7**.

This file is a prepared slice specification. It is not an `_OPEN_*.md` recovery marker and does not mean implementation has started.

## Objective

Resolve and implement the original MOO2 Colony Base as the dedicated same-system expansion path, instead of letting `colony_base` fall through MOOX's generic persistent-Building completion behavior.

The intended vertical capability is:

```text
existing Colony
-> Colony Base becomes legally available from Technology 40
-> choose/lock a legal uncolonized target in the same StarSystem
-> spend the original Production cost
-> resolve the original Colony Base completion semantics
-> create/link a second Colony in that same system
```

The exact target-binding, founding-Population and consumption/persistence behavior must be established in Gate 1 rather than assumed from the Colony Ship path.

## Existing baseline

- Core `StateSchemaVersion` is 16.
- Technology `40` is normalized as `colony_base`.
- `buildings.json` contains `colony_base` with original-observed **200 PP** Production cost and **0 BC** Maintenance.
- The current generic Building choice path exposes normalized Buildings when their Technology is known.
- The current generic Building completion path would append `colony_base` to `Colony.Buildings` as if it were an ordinary persistent Building.
- There is no dedicated Colony Base research document, rule path or focused test today.
- Colony Ship interstellar founding already supplies a useful but not automatically reusable second-Colony state baseline.

## Dependencies

- completed Colony Ship / second-Colony state baseline (`db9d01e`);
- existing generic Construction and Building-choice infrastructure;
- current StarSystem/Planet/Colony ownership and validation model.

## Scope guard

In scope only after direct Gate-1 evidence:

- Technology/buildability of Colony Base;
- Production cost and any special cost adjustment;
- same-system target selection and legality;
- when the target becomes bound to the Construction project;
- exact completion semantics;
- source/target Population effects;
- creation/linking/recalculation of the new Colony;
- whether Colony Base is consumed as a project rather than persisted as a Building;
- Observer/replay/save-load behavior.

Explicitly defer:

- Outpost behavior and Outpost-to-Colony conversion (slice 02);
- interstellar Colony Ship behavior already closed unless a concrete shared-state defect is found;
- Transport/invasion;
- AI target selection;
- broad diplomacy/hostile-system combat;
- Native/special founding cases unless Gate 1 proves they are inseparable from the basic Colony Base rule.

## Gate 1 - Checkup + original analysis / reverse engineering

- [ ] Re-check Git status, `_OPEN_*.md`, Core schema and current Building/Construction handling.
- [ ] Create the permanent Colony Base evidence document.
- [ ] Verify Technology 40 availability and the original 200 PP / 0 BC table interpretation in the executable path that actually queues/completes Colony Base.
- [ ] Verify whether Colony Base is represented as a special Construction choice despite appearing in the original Building table.
- [ ] Resolve target-selection timing: before queue, during queue, or at completion.
- [ ] Resolve legal target rules: same StarSystem requirement, empty Planet requirement, environment/ownership/hostility restrictions and multiple-target handling.
- [ ] Resolve exact completion behavior: new Colony creation, Planet link and any source-Colony side effects.
- [ ] Resolve founding Population amount/origin/job assignment and whether Population is transferred/consumed from the source Colony or created under a separate original rule.
- [ ] Verify whether Colony Base ever persists as a Building after completion or is purely a consumed colonization project.
- [ ] Verify completion timing relative to Population Growth, Food, Construction and other same-turn Colony effects.
- [ ] Verify what happens if target legality changes while the project is in progress.
- [ ] Compare the resulting new-Colony initialization with the already evidenced Colony Ship founding path without assuming they are identical.
- [ ] Separate directly proven behavior from original-derived interpretation and deliberate MOOX semantic normalization.
- [ ] Present Gate-1 findings and a narrow implementation proposal; stop before gameplay implementation.

## Gate 2 - Implementation decision

Do not mark complete until discussed/accepted.

Candidate decisions to resolve:

- [ ] Construction project kind: dedicated `colony_base` project vs special semantics under Building construction.
- [ ] Target Planet ID persistence in Construction state and schema-version impact.
- [ ] Legal-action/command payload and target projection for Human/AI callers.
- [ ] Exact authoritative cost and progress semantics.
- [ ] New-Colony initialization and source-Population mutation contract.
- [ ] Whether `colony_base` is excluded from persistent `Colony.Buildings` by design.
- [ ] Behavior when a queued target becomes invalid.
- [ ] Observer/replay/save-load event contract.
- [ ] Accepted deliberate divergences/deferrals.

## Gate 3 - Implementation

Tentative checklist; implement only the accepted Gate-2 contract.

- [ ] Prevent Colony Base from incorrectly behaving as an ordinary persistent Building.
- [ ] Add target-aware Colony Base legal action/queue semantics.
- [ ] Add authoritative same-system Planet validation.
- [ ] Advance Production using the established Construction turn-order boundary.
- [ ] Resolve completion into the evidenced source-Population/new-Colony transition.
- [ ] Create/link/recalculate the new Colony and target Planet atomically.
- [ ] Emit dedicated queue/progress/completion/colonization events as accepted.
- [ ] Add exact save/load and Observer clone/replay coverage.
- [ ] Add rejection tests for wrong system, occupied Planet, wrong owner and target invalidation.
- [ ] Add a complete headless vertical test: legal Colony Base -> Production -> same-system second Colony.
- [ ] Keep Outposts, AI and combat out of scope.

## Gate 4 - Follow-up QA + commit + close

- [ ] `gofmt` all changed Go files.
- [ ] `go test ./... -count=1`.
- [ ] `go vet ./...`.
- [ ] `git diff --check`.
- [ ] Run focused Colony Base/Building Construction/Colony Ship/Population/economy/save-load regressions.
- [ ] Verify no persistent fake `colony_base` Building remains unless Gate 1 explicitly proves that behavior.
- [ ] Update permanent evidence, README/status/HISTORY as required.
- [ ] Commit gameplay implementation.
- [ ] Commit closing documentation and delete the single `_OPEN_` marker.

## Exit criterion

Technology-40 Colony Base can create a legal second Colony on another Planet in the same StarSystem through the original-evidence-backed Production and Population rules, with exact state/replay persistence and without masquerading as an ordinary Building unless the original proves it should.