# Slice 17.0 Gate 4 - Final Acceptance and Close

Date: 2026-09-19
Slice: 17.0 - Ship Designer / Triangle Reference Test Harness
Gate: 4 - QA + close
Baseline entering Gate 4: `3dada7a` (`fix(slice): allow same-turn drafted construction buyout`)
Session: `ses-20260918T152048-00240c450344`

## Result

PASS.

Slice 17.0 satisfies its QA-harness exit criterion and can close without opening Slice 17.1.

The durable Triangle reference family now provides deterministic Baseline, Mid-Tech and All-Tech Turn-1 states plus development-only controls while all gameplay after bootstrap continues through ordinary production authority.

## Acceptance matrix

### 1. All-Tech Designer Lab without research grinding

PASS for the Slice-17.0 harness contract.

- `game-triangle-all-tech` starts at Turn 1 with the frozen All-Tech technology profile.
- `TestReferenceTriangleTechnologyProfilesFreeze` passes.
- the All-Tech profile contains all normalized research-addressable technology fields/applications frozen by Gate 2;
- it does not fabricate unsupported Hyper-Advanced completion or unlock the deferred Advanced New Game path.

The Gate-4 checklist intentionally says Doom Star/high-tech Designer QA becomes directly available **once downstream 17.x mechanics exist**. Doom Star and broader mandatory systems remain Slice 17.1 scope. Gate 4 therefore verifies that no additional research grind or harness work will be required when those mechanics land; it does not implement them early.

### 2. Construction Lab uses normal production/turn resolution

PASS.

Independent acceptance tests pass:

- `TestReferenceAdvanceUsesNormalTurnPipeline`
- `TestReferenceAdvanceUntilConstructionRequiresRealCompletionEvent`
- `TestConstructionBuyoutDeductsBCAndCompletesOnNormalTurn`
- `TestHostConstructionBuyoutUsesNormalImmediateAuthorityAndNextTurnCompletion`
- `TestHostConstructionBuyoutAppliesSavedConstructionDraftBeforeBuyout`

The normal BC buyout remains game authority rather than a Reference cheat:

- the server deducts the original-style buyout price from the Empire treasury;
- an already visible saved Planning Draft may be bought in the same Planning phase;
- the Host atomically validates/applies that saved construction queue and then runs the ordinary buyout resolver;
- the project is fully funded but completes through the next normal turn resolution;
- Reference-only BC grants remain separate development controls.

### 3. Combat Lab ships use ordinary immutable ShipDesignSpec snapshots

PASS.

`TestReferenceMilitaryShipFixtureUsesProductionDesignAndImmutableSnapshot` passes.

Reference combat fixtures:

- create a normal validated military design;
- use the shared production Ship materializer;
- retain immutable built-Ship copies of the design snapshot;
- create ordinary Combat Fleets;
- do not introduce a fixture-only tactical Ship schema.

### 4. Reference controls unavailable in normal server mode

PASS.

Independent server acceptance tests pass:

- `TestReferenceAdvanceRouteAbsentWhenDevelopmentCapabilityDisabled`
- `TestReferenceAdvanceRejectsUntrustedGameOnEnabledServer`
- `TestReferenceAdvanceHTTPUsesTrustedReferenceMetadata`
- `TestReferenceGrantBCRouteIsolationAndTrustedGrant`

A fresh server started without `-reference-games` exposed no Reference games and did not accept the Reference POST endpoint. The server implementation registers the Reference mutation routes only when the explicit process capability is enabled, and application authority additionally requires trusted per-game Reference metadata.

### 5. Equal seed/profile produces identical state

PASS.

Independent deterministic tests pass:

- `TestReferenceTriangleGameIsDeterministic`
- `TestReferenceTriangleTechnologyProfilesFreeze`
- `TestReferenceTriangleMidTechBoundaryIsPredecessorClosed`

Baseline, Mid-Tech and All-Tech reuse the frozen Triangle seed/topology contract. Technology bootstrap does not consume simulation RNG or mint unrelated Core IDs.

### 6. Save/restore works for Reference scenarios

PASS.

Automated acceptance:

- `TestReferencePersistenceCannotEscalateImportAndRestoreRetainsTrustedMetadata`
- `TestHTTPPersistenceExportImportRestoreRoundTrip`

Fresh HTTP acceptance on an isolated `-enable-persistence -reference-games` server:

1. All-Tech snapshot began at 50 BC with profile `all_tech`;
2. live snapshot exported;
3. Reference development control changed treasury to 1050 BC;
4. live snapshot restored;
5. treasury returned exactly to 50 BC;
6. trusted profile remained `all_tech`;
7. scenario identity remained `triangle-2pc-all-tech-v1`.

Generic import still cannot mint trusted Reference capability; trusted metadata belongs to host registration authority.

### 7. Full regression/static/web/browser verification

PASS.

#### Focused Gate-4 acceptance

`go test ./internal/game -run "ReferenceTriangleTechnologyProfilesFreeze|ReferenceTriangleMidTechBoundaryIsPredecessorClosed|ReferenceTriangleGameIsDeterministic|ReferenceMilitaryShipFixtureUsesProductionDesignAndImmutableSnapshot|ConstructionBuyoutDeductsBCAndCompletesOnNormalTurn" -count=1 -v`

PASS.

`go test ./internal/app -run "ReferenceAdvanceUsesNormalTurnPipeline|ReferenceAdvanceUntilConstructionRequiresRealCompletionEvent|ReferencePersistenceCannotEscalateImportAndRestoreRetainsTrustedMetadata|HostConstructionBuyoutUsesNormalImmediateAuthorityAndNextTurnCompletion|HostConstructionBuyoutAppliesSavedConstructionDraftBeforeBuyout" -count=1 -v`

PASS.

`go test ./internal/server -run "ReferenceAdvanceRouteAbsentWhenDevelopmentCapabilityDisabled|ReferenceAdvanceRejectsUntrustedGameOnEnabledServer|ReferenceAdvanceHTTPUsesTrustedReferenceMetadata|ReferenceGrantBCRouteIsolationAndTrustedGrant|HTTPPersistenceExportImportRestoreRoundTrip" -count=1 -v`

PASS.

#### Broad Go

`go test ./...`

PASS.

`go vet ./...`

PASS.

#### Web production build

`npm run build`

PASS, including encoding, New Game selector, race assets, technology-start, opponent-composition, TypeScript and Vite production build checks.

#### Fresh browser acceptance

A fresh Reference server was started on loopback and:

`npm run check:reference-lab:browser -- http://127.0.0.1:7184`

PASS.

The browser workflow proves:

- More/Development-Tools placement;
- Reference-only +100/+1,000/+10,000 BC;
- +1/+5/+N turn controls;
- same-turn drafted construction buyout;
- real BC deduction;
- next-normal-turn completion;
- 390px mobile layout;
- desktop layout;
- ordinary-game isolation.

Temporary Gate-4 servers were shut down after testing. The persistent 7171 review service remains intentionally running.

## Exit criterion

Satisfied.

A developer/tester can load deterministic technology-rich designer, construction and combat reference scenarios in seconds. Reference bootstrap only prepares initial state; subsequent Designer validation, Construction, Turn Resolution, built-Ship snapshots, Fleet state, persistence and Tactical handoff continue through ordinary authoritative paths.

## Close decision

Slice 17.0 closes at Gate 4.

Next child in the parent program is Slice 17.1 - Hulls and Mandatory Ship Systems.

Slice 17.1 is **not opened by this close**. It requires explicit user release.
