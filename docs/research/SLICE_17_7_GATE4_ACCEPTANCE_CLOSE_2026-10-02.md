# Slice 17.7 Gate 4 - Tactical Ordnance Acceptance / Close

Date: 2026-10-02

Status: **ACCEPTED / CLOSED - Gates 1-4 complete**.

## Scope accepted

Slice 17.7 closes the server-authoritative Tactical consumer for the supported missile/torpedo matrix frozen in Gate 2:

- authoritative launch, ammo/rack, readiness and cooldown state;
- persistent projectile actors with deterministic movement/ETA/target state;
- missile/torpedo impact, Shield/Armor/Structure damage and configured ordnance effects;
- manual/reactive Beam Point Defense handoff and deterministic interception ordering;
- participant-safe Tactical action/projectile projection;
- persistence/live-snapshot/replay-safe projectile state and events;
- event-driven browser launch/flight/interception/impact presentation;
- Selector -> Save -> Build -> BattleSession -> Tactical ordnance integration.

Bomb/Bio rows remain deliberately **`bombardment_only`**. Gate-1/2 evidence established that they do not belong to ordinary ship-target Tactical combat. Slice 17.7 therefore accepts correct exclusion from ship Tactical plus lossless handoff metadata; the authoritative planetary bombardment consumer is reserved downstream in Slice 19 Orbital / Planetary Combat.

Fighter/Bomber/Interceptor actors and active combat-special consumers remain Slice 17.8.

## Gate-4 hardening completed during live QA

Live All-Tech/reference-game QA exposed several integration issues after Gate 3 checkpoint 5. They were fixed before close rather than waived:

1. **Original-scale Tactical deployment**
   - strategic encounters now project opposing fleets at the evidenced original X anchors 25/55, preserving a 30-field separation;
   - missiles/torpedoes can therefore exist as visible in-flight actors instead of collapsing into near-instant contact;
   - Beam-only unit tests explicitly reposition fixtures when a near-range Beam assertion is the subject of the test rather than weakening production deployment.

2. **Built-in Tactical AI movement and ordnance actions**
   - the AI Battle decision catalog now consumes authoritative `WeaponActions`, including `battle.fire_ordnance`;
   - when no legal weapon target exists, AI receives a validated approach move chosen from server-projected `LegalMoves` before ending activation;
   - regression proves 30-field Beam AI moves into range and then receives Beam fire;
   - checkpoint-5 integration proves a launch-ready Nuclear Missile appears to AI as `battle.fire_ordnance`.

3. **Strategic AI deadlock prevention**
   - Built-in AI construction replacement uses the authoritative construction-queue contract rather than the legacy one-shot queue command;
   - Built-in AI strategic batches are preflighted through Planning Preview before submission;
   - invalid AI planning commands can no longer be stored and strand the session in `strategic_resolution`.

4. **Missile browser presentation**
   - salvo projectiles render as separate missile actors;
   - flight direction follows authoritative source->target geometry;
   - exhaust is explicitly rearward with a rear-moving exhaust particle so thrust direction is visually unambiguous;
   - browser animation remains presentation-only and does not simulate projectile authority.

5. **30-field transport/session regression alignment**
   - persistence, continuation, parallel Battle ordering and HTTP transport fixtures were updated to use legal approach movement under the accepted deployment contract;
   - the server continues to reject stale/illegal commands atomically.

## Gate-4 acceptance matrix

- **Normal Tactical launch:** PASS - supported missile/torpedo actions project through normal Tactical weapon actions and browser controls.
- **Ammo/readiness/salvo authority:** PASS - live ordnance weapon state remains server-owned and projected.
- **Projectile vectors:** PASS - deterministic launch/movement/impact regressions remain green.
- **Point Defense:** PASS - interception/non-interception ordering and projectile-side handoff remain server-authoritative.
- **Damage:** PASS - ordnance impacts resolve through authoritative Shield/Armor/Structure damage.
- **Bomb ownership:** PASS - Bomb/Bio do not become illegal ship-Tactical attacks; they remain `bombardment_only` for Slice 19.
- **Presentation:** PASS - launch/flight/interception/impact/exhaust presentation consumes authoritative events/state only.
- **Persistence/replay:** PASS - active Tactical/live-snapshot continuation remains deterministic.
- **AI:** PASS - AI can approach from original deployment distance and can consume executable ordnance actions.
- **Reference/live smoke:** PASS - 7171 All-Tech QA exercised visible missiles, multiple salvos, reload/restart/reference reset flow and subsequent Tactical/strategic continuation.

## Final QA

Final close candidate was verified after all Gate-4 hardening:

- `go test ./... -count=1` - **PASS**
  - includes `internal/server` PASS;
  - includes full `internal/session` PASS;
  - includes ordnance, Beam/shared Tactical, strategic integration and persistence regressions.
- `go vet ./...` - **PASS**
- `web: npm run build` - **PASS**
  - UTF-8/mojibake check PASS;
  - New Game selector check PASS;
  - race/technology/opponent checks PASS;
  - `check:tactical-ordnance` PASS;
  - TypeScript build PASS;
  - Vite production build PASS.
- `git diff --check` - **PASS** (line-ending notices only; no whitespace errors).
- targeted server transport regressions after 30-field update - **PASS**.
- targeted AI move/fire and AI ordnance catalog regressions - **PASS**.

## Close decision

Slice 17.7 satisfies its frozen Gate-2 contract and Gate-4 acceptance criteria and is **accepted/closed**.

The Slice-17.7 `_OPEN_` marker is removed by this closeout. Slice 17.8 - Tactical Fighters / Bombers / Combat Specials is the next prepared Slice-17 child and remains **planned/not open** until explicitly started.
