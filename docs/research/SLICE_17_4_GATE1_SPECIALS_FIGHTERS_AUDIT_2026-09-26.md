# Slice 17.4 Gate 1 - Ship Specials and Fighter Systems Audit
Date: 2026-09-26
Status: **Gate 1 complete; original 1.31 specials/fighter economics, miniaturization, compatibility, quantity and support ownership audited; Gate 2 ready / not started**
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
- Gate 1 has now recovered the original 1.31 selectable-special base Space/PP arrays, exclusions, class and strategic fields directly from Orion2.exe; final miniaturization-category mapping and separate fighter/bomber economics/quantity semantics remain open;
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
## 6. Original 1.31 economics / compatibility extraction
Durable original-EXE recovery after the initial Gate-1 candidate audit:
- embedded Watcom symbol _specials resolves to LE object 2, offset 0x6EE0;
- record width is 47 bytes;
- the table contains a dummy id=0 record followed by selectable special IDs 1..39;
- each record contains two prefix words, id, tech_id, 6 hull-indexed Space values, 6 hull-indexed Cost values, 6 exclusion IDs, one class byte and one signed strategic value;
- the local Orion2.exe SHA-256 used by the decoder is 7ae2ac2e5904ca330009af2827279d889906b0b9b7a8854c38eb707a56e955b5;
- hull-indexed Space/PP, exclusion, class and strategic values can therefore be frozen from the local 1.31 executable rather than inferred from HELP prose or a later mod table.
A classic 1.50 defaults table was used only as a row-name/cross-check aid. 36 of 39 selectable rows match the local 1.31 record fields exactly. The known local-1.31 differences are exclusion sets for:
- Cloaking Device;
- Phasing Cloak;
- Stealth Field.
Gate 2 must use the local 1.31 values for those rows.
Original design-path symbols additionally recovered:
- Special_Legal_For_Design_;
- Special_System_Exclusion_;
- Number_Of_Special_Systems_;
- Design_Template_Costs_;
- Design_Template_Space_Requirements_;
- Total_Design_Cost_;
- Total_Design_Space_.
Disassembly of the template Cost/Space loops shows:
- normal installed specials source hull-indexed base Cost and Space directly from _specials;
- normal special Cost is passed through the existing technology-level cost reducer;
- normal special Space is passed through the existing technology miniaturization-category space reducer;
- Battle Pods (id=4) are an explicit exception and bypass those normal special Cost/Space miniaturization reducers.
The next original anchor is _technology_applications at LE object 2 offset 0x6079. The interrupted research was resolving the exact application/category byte used by Miniaturization_Category_; the first attempted base address was rejected because it produced impossible all-zero category bytes.
Still required before Gate 1 close:
1. finish exact Miniaturization_Category_ mapping against _technology_applications and confirm rounding/reducer inputs for all special tech IDs;
2. decode/freeze Special_Legal_For_Design_ and Special_System_Exclusion_ behavior beyond the record exclusion arrays, including automatic-vs-selectable and duplicate/slot legality;
3. recover the separate Fighter/Bomber/Heavy Fighter design economics and quantity semantics from the original fighter/weapon path;
4. freeze Battle Pods / Megafluxers capacity ordering relative to mandatory equipment and final quote validation;
5. produce the final supported eligibility matrix.
No Gate-2 contract value should be guessed while those remaining original paths are unresolved.
### 6.1 Miniaturization closed
Direct disassembly of Miniaturization_Category_ at object-1 0x5F11C proves:
- the technology-application record stride is 13 bytes;
- byte +12 is read from _technology_applications;
- the routine defaults the returned Space class to 1;
- only application classes 3 and 9 enter exception branches that can change the result to Space class 0 or 2.
Direct evaluation of all 39 _specials tech IDs shows application byte +12 = 0 for every row. Therefore every normal selectable Special uses Space class 1.
The already-audited class-1 Space factors apply:
- mini 0 = 100%;
- mini 1 = 80%;
- mini 2 = 70%;
- mini 3 = 60%;
- mini 4 = 50%;
- mini 5+ = 40%.
The universal positive Cost reducer remains:
- mini 0 = 100%;
- mini 1 = 75%;
- mini 2 = 55%;
- mini 3 = 40%;
- mini 4 = 30%;
- mini 5+ = 25%.
Existing original rounding/minimum-one behavior from Slice 17.1 is reused; no parallel 17.4 approximation is needed.
Battle Pods id=4 remain the explicit exception already observed in Design_Template_Costs_ / Design_Template_Space_Requirements_: their hull-indexed Cost and negative Space values bypass normal Special miniaturization.
### 6.2 Slot, duplicate and exclusion legality closed
Original Special_System_Choices_ proves:
- the ordinary Designer candidate loop covers IDs 0..38;
- id=39 is not offered as a normal player-design Special;
- the current design has exactly 8 Special slots;
- an already-installed Special choice is marked unavailable, except when editing that same slot.
Original Number_Of_Special_Systems_ independently counts exactly those 8 slots.
Original Special_System_Exclusion_:
- reads the candidate row's 6 exclusion IDs from _specials;
- compares them against all 8 installed Special slots;
- ignores the currently edited slot;
- returns the colliding installed Special ID when present.
The local 1.31 table also self-excludes IDs 1..38. ID 39 does not self-exclude, but that record is not exposed by the normal Designer candidate loop.
Original Special_Legal_For_Design_(design, specialID) additionally:
- checks installed-Special exclusion relationships;
- rejects Hard Shields id=13, Multi-Phased Shields id=20 and Shield Capacitor id=30 when design+0x12 is zero.
The prior direct design-layout audit identifies design+0x12 as the installed/best shield, so those three shield-special rows require a nonzero primary shield.
Gate-2 legality can therefore be server-authoritative with:
- at most 8 installed Special rows;
- no duplicate normal Special IDs;
- local-1.31 6-ID exclusion arrays;
- the explicit shield prerequisite for IDs 13/20/30;
- id39 locked from normal Designer selection.
### 6.3 Fighter/Bomber/Heavy Fighter economics and quantity closed
The three carrier systems are not _specials rows. They are original _weapons rows 29..31 and therefore reuse the normal authoritative weapon-row quantity and miniaturization path.
Direct local 1.31 rows:
| Row | Normalized system | Tech | Base Space | Base PP | Application byte +12 | Space class |
| ---: | --- | ---: | ---: | ---: | ---: | ---: |
| 29 | Heavy Fighter Bay | 83 | 80 | 50 | 0 | 1 |
| 30 | Bomber Bay | 31 | 60 | 30 | 0 | 1 |
| 31 | Fighter Bay / Interceptor | 66 | 30 | 10 | 0 | 1 |
All three therefore use the same class-1 Space reducer and universal Cost reducer as other normal weapon rows.
Original HELP semantics remain:
- one Fighter Bay count carries a group of 4 Fighters;
- one Bomber Bay count carries a group of 4 Bombers;
- one Heavy Fighter Bay count carries a group of 4 Heavy Fighters.
Therefore the authoritative design quantity is the normal weapon-row count of Bays. There is no separate Special-slot count for fighter systems. Practical maximum count is constrained by normal weapon-row/Space validation rather than a new 17.4 hard cap.
Tactical launch, movement, interception, firing, return/reload and fighter damage remain explicitly Slice 17.8.
### 6.4 Battle Pods / Megafluxers capacity ordering closed
Original Total_Design_Space_:
- reads base hull Space from the 6-row hull table;
- tests researched technology application 89, which is Megafluxers;
- when present, returns floor(base_hull_space * 125 / 100).
Megafluxers is therefore an automatic researched capacity effect, not a selectable _specials row.
Battle Pods id=4 are separately represented as a selectable Special with fixed hull-indexed Space:
- Frigate -12;
- Destroyer -30;
- Cruiser -60;
- Battleship -125;
- Titan -250;
- Doom Star -600.
These values are added by Total_Space_Used_ with the other installed Special-space totals. Because they are negative, Battle Pods reduce used Space by approximately 50% of base hull Space and are not miniaturized.
The two effects therefore stack through separate original paths:
1. Megafluxers increases total hull capacity to floor(125% of base);
2. Battle Pods subtract their fixed negative Space from used Space.
Gate 2 must preserve that ordering and must not replace it with a multiplicative +25% then +50% approximation.
### 6.5 Explicit Gate-1 support classification
This is the audit classification that Gate 2 must freeze or narrow; it is not yet implementation.
| ID | System | Gate-1 class | Ownership statement |
| ---: | --- | ---: | --- |
| 1 | Achilles Targeting Unit | 3 | Design representation; combat targeting deferred to 17.8. |
| 2 | Augmented Engines | 3 | Design representation; combat movement consumer deferred to 17.8. |
| 3 | Automated Repair Unit | 3 | Design representation; in-combat repair deferred to 17.8. |
| 4 | Battle Pods | 1 | Design capacity effect can be complete in 17.4. |
| 5 | Battle Scanner | 3 | Design/strategic metadata may project; beam-hit combat consumer deferred to 17.8. |
| 6 | Cloaking Device | 3 | Design representation; Tactical cloak consumer deferred to 17.8. |
| 7 | Damper Field | 3 | Design representation; combat damage-field consumer deferred to 17.8. |
| 8 | Displacement Device | 3 | Design representation; combat defense consumer deferred to 17.8. |
| 9 | ECM Jammer | 3 | Design representation; missile-defense consumer deferred to 17.8. |
| 10 | Energy Absorber | 3 | Design representation; combat absorption consumer deferred to 17.8. |
| 11 | Extended Fuel Tanks | 1 | Strategic fuel-range effect can be complete in 17.4. |
| 12 | Fast Missile Racks | 3 | Design representation; missile firing/readiness consumer deferred to 17.8. |
| 13 | Hard Shields | 3 | Shield prerequisite is supported; special combat effect deferred to 17.8. |
| 14 | Heavy Armor | 3 | Derived armor can project; full armor-piercing interaction deferred to 17.8. |
| 15 | High Energy Focus | 3 | Design representation; beam combat consumer deferred to 17.8. |
| 16 | Hyper-X Capacitors | 3 | Design representation; beam readiness consumer deferred to 17.8. |
| 17 | Inertial Nullifier | 2 | Existing Tactical TurningMode consumer can be connected completely. |
| 18 | Inertial Stabilizer | 2 | Existing Tactical TurningMode consumer can be connected completely. |
| 19 | Lightning Field | 3 | Design representation; active area-defense consumer deferred to 17.8. |
| 20 | Multi-Phased Shields | 3 | Shield prerequisite is supported; special shield combat consumer deferred to 17.8. |
| 21 | Multi-Wave ECM Jammer | 3 | Design representation; missile-defense consumer deferred to 17.8. |
| 22 | Phase Shifter | 3 | Design representation; phase-shift Tactical consumer deferred to 17.8. |
| 23 | Phasing Cloak | 3 | Design representation; Tactical cloak consumer deferred to 17.8. |
| 24 | Quantum Detonator | 3 | Design representation; active detonation consumer deferred to 17.8. |
| 25 | Rangemaster Unit | 3 | Design representation; beam-range combat consumer deferred to 17.8. |
| 26 | Reflection Field | 3 | Design representation; reflected-damage consumer deferred to 17.8. |
| 27 | Reinforced Hull | 3 | Derived Structure may project; full drive/subsystem protection remains deferred to 17.8. |
| 28 | Scout Lab | 3 | Design/strategic metadata may project; complete encounter/combat bonus has no current consumer. |
| 29 | Security Stations | 3 | Design representation; boarding-defense consumer deferred to 17.8. |
| 30 | Shield Capacitor | 3 | Shield prerequisite is supported; recharge/combat consumer deferred to 17.8. |
| 31 | Stealth Field | 3 | Design/strategic metadata may project; complete detection consumer is not yet present. |
| 32 | Structural Analyzer | 3 | Design representation; combat damage consumer deferred to 17.8. |
| 33 | Sub-Space Teleporter | 3 | Design representation; active Tactical teleport consumer deferred to 17.8. |
| 34 | Time Warp Facilitator | 3 | Design representation; Tactical time/action consumer deferred to 17.8. |
| 35 | Transporters | 3 | Design representation; boarding transfer consumer deferred to 17.8. |
| 36 | Troop Pods | 3 | Marine-capacity metadata may project; boarding resolution remains deferred to 17.8. |
| 37 | Warp Dissipator | 3 | Design representation; Tactical movement/retreat consumer deferred to 17.8. |
| 38 | Wide Area Jammer | 3 | Design representation; missile-defense consumer deferred to 17.8. |
| 39 | Regeneration record | 4 | Original record exists but is not exposed by normal Special_System_Choices_; locked as Designer row. |
| W29 | Heavy Fighter Bay | 3 | Design/count/payload metadata complete; fighter Tactical consumer deferred to 17.8. |
| W30 | Bomber Bay | 3 | Design/count/payload metadata complete; bomber Tactical consumer deferred to 17.8. |
| W31 | Fighter Bay / Interceptor | 3 | Design/count/payload metadata complete; fighter Tactical consumer deferred to 17.8. |
Automatic/non-slot technology effects are separate from the 8 selectable Special slots:
- Megafluxers tech 89: automatic +25% base hull capacity path; 17.4 design/quote effect.
- Advanced Damage Control: not a selectable _specials row; any post-battle repair effect must be owned by its appropriate strategic/combat lifecycle rather than rendered as a fake Special slot.

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
Additional original-EXE progress now durable:
- _specials table located at object 2 / 0x6EE0;
- 47-byte record layout decoded, including 6 hull Space, 6 hull PP, exclusions, class and strategic fields;
- selectable IDs 1..39 plus dummy id=0 confirmed;
- local 1.31 exclusion differences versus later classic defaults isolated;
- original legality/exclusion and design Cost/Space routines located;
- normal special Cost/Space miniaturization path and Battle Pods exception identified.
Gate-1 close findings:
- all 39 Special records have exact original Space/PP/exclusion/class/strategic evidence;
- all normal selectable Specials resolve to Space miniaturization class 1; Battle Pods are the explicit non-miniaturized exception;
- 8 Special slots, duplicate rejection, 6-ID exclusion rules and shield prerequisite IDs 13/20/30 are original-evidenced;
- Special id39 is not offered by the normal Designer choice loop;
- Fighter/Bomber/Heavy Fighter rows 29..31 have exact local 1.31 economics and normal weapon-row count semantics;
- Megafluxers and Battle Pods capacity paths/order are directly original-evidenced;
- every Special/Fighter candidate now has an explicit Gate-1 support class.
Gate 1 is complete. Gate 2 must now freeze the supported subset, normalized data representation, server-owned compatibility/quote contract, immutable snapshot shape and exact 17.8 handoffs before implementation begins.