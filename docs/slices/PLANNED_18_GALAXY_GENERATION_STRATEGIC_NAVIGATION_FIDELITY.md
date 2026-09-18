# Planned slice 18 - Galaxy Generation and Strategic Navigation Fidelity

Status: **planned / queued; not open**.

Queue position: **reserved future Slice 18 by product decision**. Slice 18 follows Slice 17 and owns the galaxy-scale / strategic-route fidelity gaps found immediately after Slice 16. Slice 19 remains intentionally unassigned.

## Objective

Bring New Game galaxy geometry and strategic fleet-route legality materially closer to the owned MOO2 1.31 behavior without mixing Tactical Combat or ship-design work into this slice.

The slice must address the complete related contract rather than patching only one visible symptom:

- original-like strategic coordinate scale and star density;
- star-placement rejection rules;
- non-black-hole galaxy connectivity;
- early homeworld reachability;
- black-hole safety around stars/homeworlds;
- authoritative black-hole route blocking;
- one shared server rule for move target projection and actual command validation;
- clear HMI feedback for blocked strategic routes;
- deterministic New Game / persistence behavior after geometry changes.

## Why Slice 18 exists

Post-Slice-16 audit found that the current `moox_grid_jitter_v1` baseline preserved star count but not original strategic density/topology.

Current MOOX uses the correct conversion of **30 coordinate units = 1 parsec**, but expands the physical coordinate envelope with Galaxy Size. As a result, sampled Small/Medium/Large/Huge games all remain similarly sparse instead of larger galaxies becoming denser.

A research-only 1,200-start audit across supported Galaxy Size/Age and 2/3-player classes found **0 of 1,200** Human starts with another planet-bearing system inside the original 4-pc base-fuel neighborhood. Nearest planet-bearing systems were commonly around 7+ pc and reached as far as 16 pc in the sampled matrix.

This is not just a homeworld-neighbor bug. Fresh original disassembly shows several related galaxy rules that the current generator does not yet enforce.

## Proven original findings inherited by Slice 18

### Parsec scale

`Parsecs_Between_Points_` confirms **30 original map-coordinate units per parsec**. The current MOOX parsec conversion is therefore not the primary defect.

### Fixed strategic map envelope / density

Original `Set_Star_XYs_` uses one strategic coordinate space rather than growing the physical map with Galaxy Size. A directly observed fallback placement path stores coordinates approximately in:

- X: 20..412;
- Y: 20..360.

Gate 1 must finish the complete original placement audit before Gate 2 freezes exact production constants, but the important already-proven direction is that larger galaxy sizes add more stars to the same strategic-scale map instead of scaling the map envelope proportionally with star count.

### Homeworld early expansion

Original `Modify_Home_Worlds_` calls:

`Guarantee_A_Planet_At_Min_Dist_(homeStar, 4)`

once per homeworld.

The helper accepts an existing other planet-bearing star when one is at <=4 pc. If none exists, it selects an eligible nearby star, ensures planets exist where needed, and shifts that target toward the home star until it satisfies the requested threshold.

Therefore the hard original contract currently proven is:

- at least **one** other planet-bearing star <=4 pc from each home star;
- not a hard guarantee of two such systems.

Multiple nearby systems are nevertheless expected to be common because the original global star density is much higher than the current MOOX layout, especially at larger Galaxy Sizes.

### Global connectivity

Original `Map_Is_Connected_`:

- excludes black holes from the normal-star graph;
- connects normal stars when `Parsecs_Between_Stars <= 8`;
- accepts the generated galaxy only when all normal stars form one connected component.

This is a separate topology rule from the <=4-pc home-neighbor guarantee.

### Black-hole star spacing / homeworld safety

Original `Star_XY_Invalid_` proves that when **either** compared object is spectral class 6 (Black Hole), placement is rejected if `Parsecs_Between_Stars < 5`.

Therefore a black hole must be at least **5 pc from every star**, not only from other black holes.

Consequences:

- a home star cannot have a black hole inside its 4-pc starting neighborhood;
- the first guaranteed <=4-pc expansion target cannot itself be geometrically crowded by a black hole closer than the original minimum star spacing;
- black-hole placement belongs to galaxy generation, not only movement validation.

### Black-hole route blocking

Original HELP explicitly documents a fleet move as invalid when it is **blocked by a black hole**.

Owned 1.31 executable functions include:

- `Collect_Black_Holes_` - identifies spectral class 6 systems;
- `Black_Hole_Blocks_Points_` - checks each black hole against a source/destination segment;
- `Initialize_Black_Hole_Blocks_` - precomputes blocked star links.

`Black_Hole_Blocks_Points_` uses `Point_Within_D_Of_Seg_` with a directly observed clearance of **45 original coordinate units**.

Because map scaling/placement is being corrected in the same slice, Gate 1/2 must freeze the exact MOOX normalization of this clearance before implementation. Do not copy 45 into a different coordinate scale blindly.

## Current MOOX gaps

At the Slice-18 planning baseline:

- `moox_grid_jitter_v1` expands the map envelope with Galaxy Size;
- average nearest-star distances remain around 6.6-7.0 pc across all sizes instead of becoming denser;
- the current generator has no equivalent <=4-pc planet-bearing home-neighbor normalization;
- the current generator does not enforce the original <=8-pc non-black-hole connected-component rule;
- spectral class 6 is generated/rendered as a Black Hole but current placement does not yet enforce the proven >=5-pc black-hole-to-any-star contract;
- strategic move target projection validates destination/range/supply/ETA but not black-hole route geometry;
- actual strategic move command validation likewise has no shared black-hole segment predicate;
- HMI therefore cannot report an authoritative `blocked_by_black_hole` reason.

## Scope guard

Slice 18 owns **strategic** galaxy geometry and route legality.

It does **not** own:

- Military Ship Designer/component/weapon breadth - Slice 17;
- Tactical Combat movement/pathfinding;
- combat maneuvering around black holes;
- later Triangle-Game combat/fleet presentation beyond consuming the authoritative strategic route contract;
- Espionage - Slice 20;
- Colony/Building/Planet/Pollution breadth - Slice 21;
- race completion/designer - Slice 22;
- full Advanced start - Slice 23.

Slice 17 can be completed independently before Slice 18 opens. Slice 18 should nevertheless leave a reusable authoritative route predicate that later fleet/combat presentation can consume where appropriate.

## Binding architecture direction

- New Game seed remains the only entropy source; equal complete settings + equal seed must reproduce identical galaxy coordinates/topology.
- Strategic star coordinates are authoritative Core state.
- Parsec conversion remains server-owned and shared by supply/range/route logic.
- Galaxy-size breadth should primarily change star count/density, not silently invent a different parsec scale.
- Home-neighbor correction must be deterministic and must not bypass normal planet/system invariants.
- Black-hole placement safety is a generation invariant.
- Black-hole route blocking is a server gameplay invariant.
- Available move targets and actual move command validation must call the same route-legality rule.
- HMI must consume server-projected legality/rejection facts rather than recompute geometry independently.
- Any generation retry/rejection loop must have deterministic bounded behavior and explicit failure handling.
- Existing save/resume/replay and fixed-seed tests must be updated for the intentionally changed galaxy coordinate fixtures.

## Gate 1 - original / current geometry audit

- [ ] Re-audit the complete original `Set_Star_XYs_` / star-placement call graph and exact coordinate bounds for all supported Galaxy Sizes.
- [ ] Freeze evidence for all relevant `Star_XY_Invalid_` branches, including normal-star proximity and black-hole >=5-pc separation.
- [ ] Re-audit `Parsecs_Between_Points_` / `Parsecs_Between_Stars_` rounding and confirm the 30-units-per-pc contract end to end.
- [ ] Re-audit `Map_Is_Connected_` and exact handling of black holes / excluded objects.
- [ ] Re-audit `Modify_Home_Worlds_` and `Guarantee_A_Planet_At_Min_Dist_(...,4)` including eligible-star selection, planet creation and coordinate shifting.
- [ ] Quantify original-like expected neighborhood density by Galaxy Size from direct algorithm/evidence where feasible; distinguish hard guarantees from emergent density.
- [ ] Re-audit black-hole count/generation by Galaxy Size and any additional placement restrictions.
- [ ] Re-audit `Collect_Black_Holes_`, `Black_Hole_Blocks_Points_`, `Initialize_Black_Hole_Blocks_` and exact segment-clearance semantics.
- [ ] Map the observed 45-unit black-hole route clearance onto the final strategic coordinate contract.
- [ ] Audit current MOOX target-projection and actual-move paths to define one shared route-legality owner.
- [ ] Audit UI projection/reason surfaces required for blocked routes.
- [ ] Produce fixed-seed before/after fixture requirements for Small/Medium/Large/Huge and all Galaxy Ages.

## Gate 2 - galaxy / navigation contract freeze

- [ ] Freeze exact authoritative strategic coordinate envelope and Galaxy-Size density behavior.
- [ ] Freeze star-placement rejection/minimum-spacing rules.
- [ ] Freeze deterministic generation retry/order semantics and RNG consumption.
- [ ] Freeze normal-star connectivity: all relevant normal stars connected through <=8-pc edges, with black holes excluded exactly as proven.
- [ ] Freeze homeworld guarantee: >=1 other planet-bearing star <=4 pc from every home star; do **not** invent a hard >=2 guarantee.
- [ ] Freeze deterministic target selection/planet creation/coordinate adjustment for the home-neighbor correction.
- [ ] Freeze black-hole placement: >=5 pc from every star under the final parsec metric.
- [ ] Freeze exact black-hole route-blocking segment geometry and clearance.
- [ ] Freeze one authoritative route-legality API used by both available-target projection and move validation.
- [ ] Freeze stable rejection reason(s), including `blocked_by_black_hole` or equivalent.
- [ ] Freeze HMI behavior for blocked/unreachable routes on desktop and 390 px mobile.
- [ ] Freeze all changed deterministic New Game fixtures and persistence expectations.

## Gate 3 - implementation

- [ ] Replace/upgrade the current size-dependent sparse coordinate layout with the frozen original-like strategic-scale generator.
- [ ] Implement deterministic placement rejection and bounded regeneration.
- [ ] Enforce the <=8-pc non-black-hole connectivity contract.
- [ ] Implement the <=4-pc planet-bearing home-neighbor guarantee after home selection at the correct original-equivalent phase.
- [ ] Enforce >=5-pc black-hole-to-any-star placement.
- [ ] Implement the shared authoritative black-hole route-blocking predicate.
- [ ] Apply that predicate to available strategic fleet targets.
- [ ] Apply the same predicate to actual move command validation so crafted requests cannot bypass projection.
- [ ] Project stable route rejection reasons to App/Server/HMI.
- [ ] Add/update deterministic generation tests for all Galaxy Sizes/Ages and representative player counts/seeds.
- [ ] Add crafted black-hole route tests for clear path, tangent/near-clearance, blocked segment and endpoint-adjacent cases.
- [ ] Update intentionally changed fixed-seed coordinate/snapshot fixtures with documented provenance.

## Gate 4 - QA + close

- [ ] Representative Small/Medium/Large/Huge maps visually and numerically exhibit the frozen density progression.
- [ ] Every tested home has >=1 other planet-bearing star <=4 pc.
- [ ] No generated black hole is <5 pc from any star.
- [ ] All relevant normal-star graphs are connected under the frozen <=8-pc rule.
- [ ] Black-hole-blocked target routes are absent/disabled with the frozen authoritative reason.
- [ ] Crafted blocked move commands are rejected server-side even if submitted directly.
- [ ] Clear routes continue to honor Fuel/Supply/ETA semantics.
- [ ] Equal seed/settings produce identical galaxy topology and route legality.
- [ ] Save/export/import/restore preserves corrected coordinates and movement behavior.
- [ ] Desktop and 390 px galaxy-map route feedback remain usable.
- [ ] Full Go tests, vet, Web build/browser smoke and `git diff --check` pass.
- [ ] Re-audit downstream Slice 20/21/22/23 assumptions after the changed galaxy geometry.

## Exit criterion

Galaxy Size changes star density within one evidence-backed strategic coordinate scale; the normal-star map is connected under the original-style <=8-pc topology rule; every homeworld has an immediately reachable planet-bearing system within 4 pc; black holes stay at least 5 pc from every star and authoritatively block strategic flight paths; and the same deterministic server rule governs both move suggestions and actual commands without Tactical/Combat scope leakage.
