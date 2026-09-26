# Slice 17.4 Gate 1 - Ship Specials and Fighter Systems Audit
Date: 2026-09-26
Status: **Gate 1 in progress; candidate/support audit durable, exact ship Space/PP/miniaturization and general compatibility table still open**
Parent: Slice 17 - Military Ship Design + Tactical Combat
Child: Slice 17.4 - Ship Specials and Fighter Systems
## 1. Scope and evidence boundary
This audit does not implement Slice 17.4.
It establishes:
- the high-confidence original/normalized candidate families;
- original HELP effects already present in the committed moo2-1.31 technology extraction;
- current MOOX design/persistence/Tactical gaps;
- which gameplay effects can be owned in 17.4 versus which must remain deferred to Slice 17.8;
- reference vectors for later Gate-3 implementation/QA.
The committed technology extraction provides original technology IDs, technology-field IDs, research cost and HELP descriptions with provenance back to the original 1.31 assets/Orion2.exe.
Important boundary:
- research RP is technology research cost, not Ship Designer PP cost;
- Gate 1 has not yet recovered the exact selectable-special Space/PP/miniturization table from the original design path;
- no Gate-2 economics or compatibility values may be inferred from HELP prose.
## 2. Current MOOX representation gap
Current core.ShipDesignSpec contains:
- hull / strategic picture;
- drive;
- computer;
- armor;
- shield;
- fuel;
- authoritative hull/base/production cost and space;
- weapons.
It has no authoritative Specials/Fighter collection.
Consequences:
- no special ID/count snapshot can survive save -> revision -> construction -> built Ship today;
- no ship_specials.json normalized ruleset exists yet;
- Designer catalog/preview has no generic special/fighter rows;
- current Tactical projection sets TurningMode to normal even though the Battle core already understands inertial_stabilizer and inertial_nullifier;
- no Tactical consumers were found for repair/regeneration, cloaking, stasis, fighters/bombers, teleporters, tractor/pulsar/plasma-web/black-hole/spatial-compressor or related active specials.
This means Slice 17.4 requires a real persisted design object, not browser-only metadata.
## 3. High-confidence candidate families
The technology AI classes provide a useful completeness cross-check around the original HELP catalog.
### 3.1 Capacity / automatic / survivability
Candidate IDs:
- battle_pods
- megafluxers
- reinforced_hull
- heavy_armor
- automated_repair_unit
- advanced_damage_control
- survival_pods
- time_warp_facilitator
- quantum_detonator
Original HELP facts already recovered:
- Battle Pods: +50% available ship space.
- Megafluxers: +25% equipment capacity and automatically used on all ships once acquired.
- Reinforced Hull: triples ship structural integrity and drive-system protection.
- Heavy Armor: triples absorbable armor damage and negates armor-piercing effects.
- Automated Repair Unit: repairs during combat; HELP states 20% armor/structure and 10% systems per repair cycle.
- Advanced Damage Control: completely repairs ships between battles and is automatic once researched.
Gate-2 note:
Megafluxers and Advanced Damage Control are automatic technology effects, not automatically assumed to be selectable special slots.
### 3.2 Fighter / bomber carrier systems
Candidate IDs:
- fighter_bays
- bomber_bays
- heavy_fighter_bays
Original HELP facts:
- Fighter Bays carry groups of 4 short-range fighters; each carries the best point-defense-capable beam and can fire it four times before returning.
- Bomber Bays carry groups of 4; each bomber carries one best available bomb and moves at speed 10.
- Heavy Fighter Bays carry heavy fighters with 2 best point-defense beams and 2 best bombs; heavy fighters have 5 hits modified by best armor and speed 12.
17.4 owns:
- legal design representation;
- Space/PP quote;
- bay/count snapshot;
- payload metadata required to hand off exactly.
17.8 owns:
- launch/movement/targeting/interception;
- fighter weapon/bomb firing;
- return/reload behavior;
- fighter damage/destruction and active Tactical UI.
### 3.3 Mobility
Candidate IDs:
- augmented_engines
- inertial_stabilizer
- inertial_nullifier
- sub_space_teleporter
- phase_shifter
Existing Battle-core consumer:
- normal turning cost = 2 movement points per facing step;
- Inertial Stabilizer turning cost = 1;
- Inertial Nullifier turning cost = 0.
Current strategic-to-Tactical projection always supplies normal TurningMode.
Therefore Stabilizer/Nullifier are the clearest category-2 candidates:
design representation is missing, but an already-supported Tactical consumer exists and can be connected without inventing Slice-17.8 combat behavior.
Sub-Space Teleporter and Phase Shifter require dedicated active Tactical consumers and remain deferred.
### 3.4 Sensors / range / strategic utility
Candidate IDs:
- battle_scanner
- scout_lab
- extended_fuel_tanks
- stealth_field
- warp_dissipater
- security_stations
Recovered HELP facts:
- Battle Scanner: +50 beam chance-to-hit and +2 parsecs galactic scanning range.
- Scout Lab: research by ship size 1/2/4/8/16/32 and +10 per size class to fleet attack roll against monsters/Antarans.
- Extended Fuel Tanks: +50% ship range at a considerable Space cost.
- Stealth Field: hides equipped ships from main-screen detection.
These are mixed strategic/Tactical effects. Gate 2 must split the owned derived-stat portion from deferred Tactical consumers instead of labelling the entire special complete.
### 3.5 Beam / targeting support
Candidate IDs:
- achilles_targeting_unit
- high_energy_focus
- hyper_x_capacitors
- rangemaster_unit
- structural_analyzer
- battle_scanner
These alter Beam/Tactical resolution or readiness. They may be represented in 17.4, but their full combat consumption remains downstream unless an already-supported consumer is explicitly proven.
### 3.6 ECM / area defense
Candidate IDs:
- ecm_jammer
- multi_wave_ecm_jammer
- wide_area_jammer
- lightning_field
- pulsar
- spatial_compressor
- anti_missile_rockets
HELP exposes missile-evasion and active-area effects, but current Battle/Tactical code has no named consumer for these systems. Design representation can be accepted only with explicit metadata/defer status.
### 3.7 Shield / damage-field specials
Candidate IDs:
- hard_shields
- multi_phased_shields
- shield_capacitors
- reflection_field
- damper_field
- energy_absorber
- displacement_device
The ordinary Class I/III/V/VII/X shield family remains mandatory-system territory from Slice 17.1 and is not reclassified as a 17.4 special.
Special shield modifiers/effects above remain candidates because they alter shield strength, recharge, damage or hit resolution beyond the primary shield component.
### 3.8 Boarding / security
Candidate IDs:
- assault_shuttles
- transporters
- troop_pods
- tractor_beam
- security_stations
Recovered HELP includes:
- Troop Pods double marines aboard a ship.
- Assault Shuttles move groups of marines to enemy ships.
- Transporters perform ranged marine transfer when shield conditions permit.
- Security Stations improve anti-boarding defense.
Design/marine-capacity metadata may be 17.4-owned; capture/boarding action resolution remains deferred until the responsible Tactical consumer exists.
### 3.9 Active Tactical specials
Candidate IDs:
- black_hole_generator
- gyro_destabilizer
- plasma_web
- stasis_field
- sub_space_teleporter
- phase_shifter
- time_warp_facilitator
- quantum_detonator
These require active combat state/action resolution and are therefore category-3 candidates by default:
design representation complete, Tactical consumer explicitly deferred to Slice 17.8 unless Gate 2 proves a bounded existing consumer.
## 4. Candidate exclusions / separate owners
The AI-class cross-check also returns technologies that must not silently become 17.4 selectable specials:
- Class I/III/V/VII/X Shield: primary mandatory shield family, already owned by Slice 17.1.
- Mauler Device: weapon, not a ship-special row.
- Jump Gate / Star Gate: strategic infrastructure, not an installed Ship Designer special.
- Microlite Construction: technology/economic effect, not assumed to be an installed special.
- normal ECM/missile weapon families must be distinguished from special rows before Gate 2.
Gate 1 must recover the original design-special eligibility table before the final supported matrix is frozen.
## 5. Technology/research audit already complete
Committed technology extraction now gives unlock identity and research-field RP for the candidate catalog.
Representative research costs:
- Extended Fuel Tanks 50 RP
- Reinforced Hull 80 RP
- Fighter Bays 80 RP
- Heavy Armor 150 RP
- Battle Pods 250 RP
- Battle Scanner 250 RP
- Inertial Stabilizer 650 RP
- Automated Repair Unit 1500 RP
- Bomber Bays 2000 RP
- Megafluxers / Cloaking Device / High Energy Focus / Stasis Field 3500 RP
- Heavy Fighter Bays / Achilles Targeting Unit 6000 RP
- Inertial Nullifier / Sub-Space Teleporter / Displacement Device 7500 RP
- Damper Field / Phasing Cloak / Reflection Field / Time Warp Facilitator / Black Hole Generator / Spatial Compressor / Quantum Detonator 15000 RP
Again: these are research costs, not ship PP costs.
## 6. Economics / compatibility still open
Before Gate 1 can close, original evidence is still required for every supported selectable row:
1. base Ship Designer Space;
2. base production PP;
3. miniaturization class and exact rounding;
4. whether cost/space scale with hull size or quantity;
5. maximum count / duplicate legality;
6. mutual exclusions;
7. automatic-vs-selectable distinction;
8. fighter-bay quantity semantics;
9. interaction of Battle Pods / Megafluxers with base hull capacity;
10. ordering of capacity modifiers relative to mandatory equipment and normal quote validation.
Orion2.exe is present locally under reference/original/support/files.
Standard wdis/wdump/objdump tools are not installed in the current environment, so these values remain an explicit Gate-1 extraction task rather than guesses.
## 7. Provisional support classification
This is audit classification, not Gate-2 acceptance.
### Category 1 candidate - design + strategic/derived effect can be completed in 17.4
- Battle Pods
- Reinforced Hull
- Heavy Armor
- Extended Fuel Tanks
- Scout Lab
- strategic portion of Battle Scanner
- strategic portion of Stealth Field
- Troop Pods marine-capacity portion
- automatic Megafluxers / Advanced Damage Control if Gate 2 confirms they belong in the design projection rather than selectable slots
### Category 2 candidate - design + already-supported Tactical effect
- Inertial Stabilizer
- Inertial Nullifier
The Battle core already has exact turning behavior; 17.4 must only connect a frozen persisted design snapshot to that existing consumer.
### Category 3 candidate - design representation, Tactical deferred to 17.8
- Fighter Bays
- Bomber Bays
- Heavy Fighter Bays
- Cloaking Device / Phasing Cloak
- Automated Repair Unit during-combat behavior
- Sub-Space Teleporter / Phase Shifter
- Damper / Lightning / Reflection / Energy Absorber / Displacement effects
- Stasis Field
- Black Hole Generator
- Gyro Destabilizer
- Pulsar / Plasma Web / Spatial Compressor
- Time Warp Facilitator / Quantum Detonator
- Assault Shuttles / Transporters / Tractor Beam
- Beam-support specials whose combat resolution is not already implemented
- ECM/area-defense specials without a current Battle consumer
### Category 4 candidate - locked/not accepted until original eligibility is proven
- any technology returned by broad AI-class scanning that is not proven to be a Ship Designer special row;
- Phase Shifter if its player-design availability remains absent from the normal research-field path;
- any special whose exact Space/PP/miniaturization/compatibility cannot be recovered.
## 8. Reference vectors defined for 17.0 / Gate 3
Existing game-triangle-all-tech already exposes all six hulls including Doom Star and has deterministic Doom-Star save/browser coverage.
Gate 3 should materialize at least these special-focused vectors on that harness:
### Vector A - Doom Star capacity / survivability
- All-Tech Doom Star;
- Battle Pods;
- Reinforced Hull;
- Heavy Armor;
- account for automatic Megafluxers if Gate 2 confirms global application.
Purpose:
prove capacity ordering, derived Structure/Armor and immutable snapshot persistence.
### Vector B - Doom Star carrier
- All-Tech Doom Star;
- Fighter Bays;
- Bomber Bays;
- Heavy Fighter Bays.
Purpose:
prove bay legality/counts, payload metadata and exact handoff to Slice 17.8 without fake Tactical launch behavior.
### Vector C - mobility consumer
- medium/large all-tech combat ship;
- Inertial Stabilizer, then Inertial Nullifier.
Purpose:
prove persisted special -> Tactical TurningMode -> existing exact turn-cost path 1 / 0.
### Vector D - strategic utility
- Extended Fuel Tanks;
- Battle Scanner;
- Scout Lab.
Purpose:
prove supported strategic derived effects without accidentally claiming deferred Tactical behavior.
## 9. Gate-1 progress
Completed:
- candidate family inventory and AI-class completeness cross-check;
- original HELP effect audit for major design/fighter candidates;
- technology IDs / field IDs / research-cost availability confirmed;
- current ShipDesignSpec / Designer / Battle handoff gap identified;
- existing Tactical consumer audit completed;
- 17.8 fighter/combat-special ownership boundary established;
- high-tech/Doom-Star reference vectors defined.
Still required before Gate 1 close:
- exact original Ship Designer Space / PP table;
- miniaturization/rounding;
- selectable-vs-automatic table;
- duplicate/quantity/exclusivity rules;
- final eligibility matrix.
No implementation should begin before those are frozen in Gate 2.