# Slice 17.2A Gate 4 - Acceptance Close
Date: 2026-09-26
Status: **accepted and closed; Gates 1-4 complete**
Parent: Slice 17 - Military Ship Design + Tactical Combat
Follow-up: Slice 17.2A - Beam Weapon Firing Arcs
## 1. Acceptance basis
The user manually reviewed the implemented Beam firing-arc Ship Designer on the canonical 7171 review server, including the radial compass HMI and its mobile refinements, judged the result acceptable, and explicitly requested progression to Gate 4.
Gate 4 therefore validates the frozen Gate-2 contract and the completed Gate-3 implementation rather than widening scope.
Authoritative evidence remains:
- Gate 1 original audit:
  docs/research/SLICE_17_2A_GATE1_BEAM_FIRING_ARC_AUDIT_2026-09-25.md
- Gate 2 frozen contract:
  docs/research/SLICE_17_2A_GATE2_BEAM_FIRING_ARC_CONTRACT_2026-09-25.md
- Gate 3 implementation evidence:
  docs/research/SLICE_17_2A_GATE3_BEAM_FIRING_ARC_IMPLEMENTATION_2026-09-25.md
## 2. Gate-4 acceptance results
### 2.1 Original-supported Beam arcs are selectable
Accepted Beam arcs remain exactly:
- forward
- forward_extended
- back_extended
- rear
- 360
The Weapon Selector exposes them through the server-owned catalog metadata. The accepted radial compass HMI presents F / Fx / Bx / B around the current ship silhouette and 360 in the center.
Slice 17.2 remains closed/accepted and was not reopened.
### 2.2 Verified 360-degree surcharge
The authoritative Beam quote path retains the frozen original surcharge matrix:
- Forward: 0%
- Forward Extended: +25%
- Back Extended: +25%
- Rear: 0%
- 360: +50%
The surcharge remains additive with Beam modifier percentages after mount arithmetic and before miniaturization.
Regression coverage includes:
- Baseline Laser Cannon arc matrix;
- Mid-Tech Neutron Blaster 360 reference vector: Space 12 / PP 9 at miniaturization level 1;
- All-Tech Phasor 360 reference vector with Point Defense + Shield Piercing + Autofire.
### 2.3 Lossless persistence and BattleSession handoff
Regression coverage proves the selected Beam FiringArcID survives:
- authoritative quote canonicalization;
- save;
- reload/revision;
- construction;
- built-Ship immutable snapshot;
- BattleSession TacticalWeaponSpec / metadata projection.
A non-default 360 Beam arc is preserved exactly rather than inferred or replaced.
### 2.4 Accepted desktop/mobile Weapon Selector HMI
The user accepted the radial firing-arc compass after iterative review.
Final accepted HMI includes:
- current procedural ship silhouette centered in the compass;
- firing-arc coverage glow outside the ship center;
- Fx positioned above the horizontal midline;
- Bx positioned below the horizontal midline;
- additional mobile ship scaling/clearance;
- fixed-360 presentation for Missile/Torpedo/Bomb rows;
- no overlap regression at the existing 390px mobile reference width.
The browser reference-lab smoke remains green for desktop and mobile.
### 2.5 Server authority remains intact
The browser does not own Beam firing-arc legality or quote arithmetic.
The client:
- renders server-projected firing_arc_choices;
- displays server-projected labels and percentage metadata;
- sends the selected firing_arc_id;
- consumes authoritative preview/quote results.
There is no browser-local alternate Space/PP quote formula.
### 2.6 Tactical boundary remains honest
BattleSession carries Beam FiringArcID as metadata, but Slice 17.2A does not implement:
- target-bearing geometry;
- ship facing/rotation checks;
- Tactical fire eligibility by arc;
- Tactical arc rendering/enforcement.
Beam tactical support remains metadata-only until Slice 17.6 consumes the frozen arc metadata.
## 3. Final verification
Gate-4 close verification passed on the accepted final state:
- go test ./... -count=1
- go vet ./...
- npm --prefix web run build
- npm --prefix web run check:reference-lab:browser against the 7171 reference-games server
- git diff --check
The accepted browser smoke covers the compact/mobile Ship Designer, Beam arc selection/requote/persistence, and fixed-360 Ordnance presentation.
## 4. Gate-4 checklist
- [x] Original-supported Beam arcs can be selected without reopening 17.2.
- [x] 360-degree Beam configuration applies the verified original design surcharge.
- [x] Save -> reload -> revise -> build -> BattleSession preserves the chosen arc exactly.
- [x] Weapon Selector remains compact/non-overlapping on desktop and mobile.
- [x] No browser-local pricing/legality formula diverges from server authority.
- [x] Tactical support status is honest until Slice 17.6 consumes the arc metadata.
## 5. Close result
Slice 17.2A is accepted and closed with Gates 1-4 complete.
The exit criterion is satisfied:
- Beam firing arcs are original-evidenced;
- the contract is frozen;
- quote/legality is server-authoritative;
- supported arcs are selectable and costed;
- FiringArcID persists losslessly into immutable Ship/BattleSession snapshots;
- Tactical directional enforcement remains explicitly owned by Slice 17.6.
No later Slice-17 child is opened by this closeout.
The next planned child is Slice 17.4 - Ship Specials and Fighter Systems, which remains planned/not open.