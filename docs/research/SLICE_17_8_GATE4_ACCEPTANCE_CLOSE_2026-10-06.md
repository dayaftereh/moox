# Slice 17.8 Gate 4 - Tactical Fighters / Bombers / Combat Specials Acceptance / Close

Date: 2026-10-06

Status: **ACCEPTED / CLOSED - Gates 1-4 complete**.

## Scope accepted

Slice 17.8 closes the server-authoritative Tactical consumers frozen in Gate 2 for Fighter/Bomber/Heavy-Fighter actors, Boarding/Capture and the accepted combat-Special matrix:

- authoritative Fighter Bay readiness, launch, moving-actor lifecycle, targeting, attack, interception, loss, return and recovery;
- shared Fighter/projectile moving-actor ordering and Beam Point-Defense handoff;
- physical / Transporter Boarding, Raid internal damage, Capture authority transfer and actor reconciliation;
- authoritative Internal-system pools and damage-driven live Weapon/Special/Shield/speed consequences;
- accepted combat-Special targeting, defense, damage, readiness, cloak/phasing, teleport, recharge/repair and Time-Warp scheduler consumers;
- persistent/replay-safe Fighter, Internal, mutable authority and Special lifecycle state;
- participant-safe Tactical action/state projection and common built-in AI actions;
- event-driven browser Fighter/Boarding/Special presentation with no client-side combat authority;
- Selector -> Save -> Build -> BattleSession -> Fighter launch integration.

Strategic-only rows and explicitly deferred/unproven behavior remain outside Slice 17.8. Gate 4 accepts those ownership boundaries rather than presenting unsupported behavior as Tactical-working.

## Gate-4 review hardening

The final implementation review immediately before the Checkpoint-6 commit found one real integration omission and closed it before Gate 3 was committed:

1. **Achilles Targeting Unit Beam wiring**
   - the original-compatible Internal selector and 3x Weapon/Shield weighting were already implemented and regression-covered;
   - review found that the firing ship's functional `achilles_targeting_unit` had not yet been assigned into the Beam dependency context;
   - the wiring was added before the Gate-3 Checkpoint-6 commit;
   - a direct Beam regression now proves functional attacker Achilles bypasses Armor and consumes authoritative Internal state.

No additional gameplay hardening was required after Gate 3 commit. Gate 4 is therefore an acceptance/close pass over the completed implementation rather than a second mechanics checkpoint.

## Gate-4 acceptance matrix

- **Fighter launch / readiness:** PASS - Fighter/Bomber/Heavy-Fighter bays are server-authoritative and launch through normal Tactical weapon actions.
- **Fighter lifecycle:** PASS - movement, attack cadence, losses/interception, retarget, return/recovery and carrier-loss handling are deterministic and BattleSession-owned.
- **Cross-family interactions:** PASS - shared moving-actor order, Beam Point Defense, Lightning/interception and Special-aware damage use common server authority.
- **Boarding / Capture:** PASS - physical/Transporter legality, Raid reconciliation, Capture ownership/controller mutation and actor reconciliation are authoritative and persisted.
- **Internal systems / Achilles:** PASS - Internal pools drive live consumers; Vector K preserves the original Internal threshold/RNG/weight transform and real Beam integration.
- **Combat Specials:** PASS - accepted passive/active Tactical consumers use server-owned functional state, lifecycle and event ordering; unsupported/strategic-only behavior is not synthesized.
- **Scheduler / movement:** PASS - self-teleport and Time-Warp use normal BattleSession command/pass authority and persisted scheduler state.
- **Persistence / replay / reconnect:** PASS - combined modern Fighter/Special/Capture state round-trips, restores, projects detached views and continues deterministically; legacy schema compatibility remains explicit.
- **AI:** PASS - common Battle decisions consume server-projected Fighter, Beam, Ordnance, Boarding, Move, End, Wait and Retreat legality.
- **Presentation:** PASS - browser renders Fighter actors, Boarding controls, teleport, Internal/Special scan state and Retreat legality from authoritative state/events only.
- **Configured integration:** PASS - All-Tech Cruiser Save -> Build -> Encounter -> BattleSession projects Fighter doctrine/Internal/Special state and completes a real `battle.launch_fighters` command/event.
- **Reference/runtime smoke:** PASS - a fresh current-HEAD server build on isolated port 7172 served the current HMI, listed Reference Games and opened `game-triangle-all-tech` through the Runde-1 Galaxy view in managed Chrome.

## Final QA

Final Gate-4 close candidate was verified from committed Gate-3 HEAD `34f2dcc`:

- `go test ./... -count=1` - **PASS**;
- `go vet ./...` - **PASS**;
- `web: npm run build` - **PASS**;
  - UTF-8/mojibake check PASS;
  - New Game selector check PASS;
  - race/technology/opponent checks PASS;
  - Tactical ordnance browser contract check PASS;
  - TypeScript build PASS;
  - Vite production build PASS;
- `git diff --check` - **PASS**;
- fresh current-HEAD Reference/Game browser smoke on temporary `127.0.0.1:7172` - **PASS**;
- persistent reference service on 7171 was not restarted or disturbed.

No rules-data files changed in Gate 4. The accepted `moo2-1.31` ruleset fingerprint remains:

`68c06a57bb9fdc18c0ff8045b54241e51cc9cba41298898fbfb535b6a338bd46`

## Close decision

Slice 17.8 satisfies its frozen Gate-2 contract and Gate-4 acceptance criteria and is **accepted/closed**.

The Slice-17.8 `_OPEN_` marker is removed by this closeout. Slice 17.9 - Full Ship Designer + Tactical Integration and Acceptance is the next prepared Slice-17 child and remains **planned/not open** until explicitly started.
