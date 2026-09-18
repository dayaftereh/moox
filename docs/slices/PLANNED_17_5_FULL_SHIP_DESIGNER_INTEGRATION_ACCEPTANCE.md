# Planned Slice 17.5 - Full Ship Designer Integration and Acceptance

Status: **planned; child of Slice 17; not open**.

## Objective

Integrate the completed 17.1-17.4 design breadth into one polished server-authoritative Ship Designer and close the entire Slice-17 program with end-to-end reference-lab, construction, persistence and combat-handoff acceptance.

## Inputs

17.5 opens only after:

- 17.0 Reference/Test Harness is closed;
- 17.1 Hulls/Mandatory Systems is closed;
- 17.2 Beam Weapons is closed;
- 17.3 Ordnance is closed;
- 17.4 Specials/Fighters is closed;
- every deferred tactical feature has an explicit downstream owner.

## Integration scope

- unified server-owned Designer Catalog;
- hull + mandatory component + weapon + special selection;
- exact remaining Space and Production Cost;
- technology/buildability/compatibility locks;
- miniaturization results;
- stable design IDs/revisions;
- immutable built-Ship snapshots;
- save/edit/revise workflow;
- Colony Construction handoff and completion;
- Command Point/fleet integration;
- persistence/export/import/restore;
- reference-lab scenario selection;
- final desktop/mobile Ship Designer HMI;
- Tactical/Triangle handoff metadata and support status.

## Gate 1 - integration audit

- [ ] Reconcile all 17.1-17.4 catalogs into one legal-design model.
- [ ] Audit duplicate/conflicting cost/space/miniaturization paths.
- [ ] Audit revision and construction queue behavior when a design changes after queueing or after build.
- [ ] Audit all server rejection reasons and HMI projections.
- [ ] Audit persistence/migration/schema consequences.
- [ ] Audit reference-lab coverage for early/mid/high-tech and combat snapshots.
- [ ] Identify any remaining gaps that block claiming Slice-17 breadth complete.

## Gate 2 - final designer contract freeze

- [ ] Freeze complete supported design matrix.
- [ ] Freeze unified cost/space/miniaturization calculation.
- [ ] Freeze save/revision/delete/queue/built-snapshot semantics.
- [ ] Freeze construction handoff and Command Point effects.
- [ ] Freeze final Designer HMI interaction/responsive contract.
- [ ] Freeze reference-scenario QA matrix.
- [ ] Freeze Tactical/Triangle compatibility/support metadata.

## Gate 3 - integration implementation

- [ ] Merge server catalogs/legal actions into one designer projection.
- [ ] Complete HMI selectors, locks, summaries and preview.
- [ ] Complete design save/revise/construction workflow.
- [ ] Complete persistence/export/import/restore.
- [ ] Complete reference-lab controls and scenario navigation.
- [ ] Add cross-category regression matrices and crafted invalid requests.
- [ ] Ensure later Tactical work consumes the same built-Ship snapshots.

## Gate 4 - program QA + close

- [ ] Early-, mid- and high-tech reference labs expose exactly their frozen legal choices.
- [ ] Representative Frigate/Destroyer/Cruiser/Battleship/Titan/Doom Star designs can be authored where unlocked.
- [ ] Representative Beam/ordnance/special combinations validate exact cost/space.
- [ ] Doom Star + late-game systems can be tested without research/turn grinding.
- [ ] `advance until construction complete` uses normal turn resolution and yields the exact designed Ship snapshot.
- [ ] Built Ships do not mutate after later design revision.
- [ ] Save/export/import/restore preserves all design and built-Ship data.
- [ ] Reference controls remain absent outside explicit development mode.
- [ ] Desktop and 390 px Designer/Construction flows pass.
- [ ] Full Go tests, vet, Web build/browser smoke and `git diff --check` pass.
- [ ] Parent Slice 17 completion checklist is closed and roadmap/HISTORY updated.

## Exit criterion

Ship design is a broad, coherent product surface rather than a Frigate/Laser fixture: the server owns all supported legality/cost/space/technology decisions, the browser can author and build those designs, reference labs make late-game QA practical, persistence is stable, and later Tactical/Triangle systems receive exactly the immutable ship snapshots created by normal gameplay.
