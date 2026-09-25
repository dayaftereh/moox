# Slice 17.2A Gate 3 - Beam Firing-Arc Implementation Evidence
Date: 2026-09-25
Status: **Gate 3 complete; Gate 4 acceptance ready / not started**
Parent: Slice 17 - Military Ship Design + Tactical Combat
Follow-up: Slice 17.2A - Beam Weapon Firing Arcs
## 1. Scope
Gate 3 implements only the contract frozen in:
docs/research/SLICE_17_2A_GATE2_BEAM_FIRING_ARC_CONTRACT_2026-09-25.md
Slice 17.2 and Slice 17.3 remain closed/accepted. Tactical target-direction geometry and fire eligibility remain deferred to Slice 17.6.
## 2. Core and authoritative quote path
Implemented canonical Beam firing-arc IDs:
- forward
- forward_extended
- back_extended
- rear
- 360
Authoritative Beam quote behavior:
- an empty Beam firing arc canonicalizes to forward;
- an unknown non-empty arc rejects with weapon_firing_arc_unknown;
- all five frozen Beam arcs are accepted for every otherwise legal Beam row;
- surcharge percentages are 0 / 25 / 25 / 0 / 50;
- the arc surcharge is added to the Beam modifier percentage after mount arithmetic and before miniaturization;
- existing Count/grouped-row semantics remain unchanged.
The successful quote returns the canonical FiringArcID and save/revision paths persist that canonical value.
## 3. Catalog, API and Weapon Selector
The authoritative designer catalog now exposes Beam firing_arc_choices and default_firing_arc_id.
Each choice projects:
- canonical ID;
- English label;
- German label;
- Space percentage;
- production-cost percentage.
The web API type accepts all five canonical IDs.
The Weapon Selector:
- defaults a new Beam to Forward;
- renders the ordered choices supplied by the server catalog;
- uses the server-projected localized labels and cost percentages;
- sends the selected firing_arc_id through the existing authoritative preview path;
- requotes whenever the selected arc changes;
- reopens an edited row with its persisted arc selected.
Fixed-360 Ordnance behavior from Slice 17.3 remains separate and unchanged.
## 4. Persistence and BattleSession handoff
The implementation retains core.ShipWeaponMount.FiringArcID as the single persisted design field.
Regression coverage proves a non-default 360 Beam arc survives:
- save/revision;
- built-Ship snapshot construction;
- TacticalWeaponSpec / BattleSession metadata projection and clone paths.
17.2A still does not interpret the metadata as Tactical directional eligibility. That remains Slice 17.6.
## 5. Frozen reference vectors
### Baseline - Laser Cannon
Count 1, Standard, no Beam modifiers:
- Forward: Space 10 / PP 5
- Forward Extended: Space 12 / PP 6
- Back Extended: Space 12 / PP 6
- Rear: Space 10 / PP 5
- 360: Space 15 / PP 7
### Mid-Tech - Neutron Blaster
Count 1, Standard, no modifiers, Mid-Tech miniaturization level 1, 360:
- Space 12
- PP 9
This is covered explicitly by TestSlice172AMidTechNeutron360ReferenceVector.
### All-Tech - Phasor
Count 3, Point Defense, Shield Piercing + Autofire, miniaturization level 4, 360:
- unit Space 6
- unit PP 3
- row Space 18
- row PP 9
This preserves the frozen additive modifier+arc ordering from Gate 2.
## 6. Browser / HMI proof
web/scripts/check-reference-lab-browser.mjs now verifies on the real All-Tech reference game that:
- the Beam selector exposes exactly forward, forward_extended, back_extended, rear, 360 in frozen order;
- Forward is selected by default;
- selecting 360 updates the authoritative Space/PP preview;
- applying the edit and reopening the row preserves 360;
- the existing 390px mobile and desktop selector/layout invariants continue to pass.
No browser-local Beam legality or pricing formula was added.
## 7. Verification
Gate-3 verification completed successfully:
- go test ./... -count=1
- go vet ./...
- npm --prefix web run build
- npm --prefix web run check:reference-lab:browser against a -reference-games server
- focused Baseline/Mid-Tech/All-Tech Beam arc regression tests
- git diff --check
## 8. Gate-3 result
Gate 3 is complete.
Implemented surface:
- Core IDs and canonicalization;
- authoritative quote/validation;
- catalog/API metadata;
- server-owned preview;
- Weapon Selector;
- save/revision;
- built Ship;
- BattleSession metadata;
- Baseline/Mid-Tech/All-Tech regressions;
- mobile/desktop browser proof.
Gate 4 acceptance is ready / not started. Tactical direction enforcement remains explicitly owned by Slice 17.6.