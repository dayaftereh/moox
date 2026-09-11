# Slice 15.6 Gate3 Block5F — Foreign Fleet Detection / Scanner Intelligence

Status: **PLANNED FOLLOW-UP after Block5E arrival/contact/Tactical acceptance**

## Why this block exists

Block5D/5E now establishes the own-Fleet route lifecycle on the Galaxy map: pre-submit planning routes, immutable committed transit, a travelling Fleet marker, server-projected remaining parsecs/ETA, and the compact read-only transit Fleet dialog. The same presentation language must also support **foreign Fleets moving through deep space**, but foreign visibility and ship detail are an intelligence problem, not merely a UI problem.

This block therefore owns scanner-governed foreign Fleet detection and presentation. It must not leak complete foreign Fleet state to React and then hide it cosmetically.

## Current MOO2 1.31 ruleset evidence already normalized in the repository

`data/rulesets/moo2-1.31/technologies.json` currently contains scanner technology descriptions that establish the intended research basis:

- `space_scanner`: detects enemy ships; base detection range **1 pc**; ships in transit can be detected farther based on ship size class.
- `tachyon_scanner`: detects enemy ships; base detection range **3 pc**; ships in transit can be detected farther based on ship size class; also modifies enemy missile evasion.
- `neutron_scanner`: detects enemy ships; base detection range **5 pc**; ships in transit can be detected farther based on ship size class; also modifies enemy missile evasion.
- `battle_scanner`: in addition to its tactical beam-to-hit effect, the current normalized description grants **+2 pc galactic scanning range**.

These descriptions are evidence for the implementation/research plan, not yet a complete normalized strategic scanner mechanic. Exact stacking, scanner source locations, size-class range contribution, contact persistence and information depth still require authority-side research/normalization before implementation.

## Frozen product direction

### 1. Detection is server authority

- A foreign Fleet may exist and travel authoritatively without being projected to the observing player.
- The server determines whether that Fleet is currently detectable from the observing empire's scanner coverage.
- React receives only permitted contacts and permitted intelligence fields.
- No hidden foreign Fleet/ship list, destination, route, ETA or equipment may be shipped to the browser merely to be hidden with CSS/conditional rendering.

### 2. Galaxy presentation once a foreign transit contact is detectable

- A detectable foreign Fleet in transit receives a Galaxy Fleet marker at its authoritative visible transit position.
- The **Fleet marker keeps the foreign empire/player color**. Threat state must not replace empire color.
- A detectable hostile/incoming route uses a **red dashed route line** toward the projected destination when that destination is permitted intelligence.
- If the destination itself is not permitted intelligence, the projection/UI must not reveal it solely because the Fleet was detected; a reduced contact presentation is required instead.
- ETA / remaining turns and remaining parsecs are displayed only when the authoritative scanner-intelligence projection permits those values.
- The existing own-transit route/marker implementation should be reused rather than building a second unrelated Galaxy widget.

### 3. Incoming warning case

Primary acceptance case: an AI Fleet is detected while travelling toward a Human-owned/known system.

Expected player-facing result when scanner authority permits it:

- foreign player-colored Fleet marker in deep space;
- red dashed line from that moving Fleet marker to the threatened Human system;
- visible remaining ETA and/or remaining pc according to projected intelligence;
- the route advances each turn until arrival;
- no player action can cancel or redirect the foreign Fleet.

### 4. Reuse the compact read-only Fleet dialog

Clicking a detected foreign travelling Fleet should open the same compact Fleet-dialog language used by own transit, but with **scanner/intelligence redaction**:

- no selection controls;
- no target/retarget/cancel controls;
- no foreign movement mutation actions;
- only the ships/attributes currently permitted by scanner intelligence are rendered;
- `?` details remain the single inspection affordance where individual-ship intelligence is available.

The dialog must support graduated intelligence rather than assuming all-or-nothing visibility. Candidate levels to research/normalize include:

1. contact only — Fleet exists / rough size or ship count;
2. classes/hulls visible;
3. individual ships/design identity visible;
4. equipment/weapons/damage/readiness visible where the original rules and available intelligence support it.

These levels are a product architecture direction; exact MOO2 thresholds/content must be researched before freezing gameplay values.

### 5. Scanner coverage must be explicit strategic state/derivation

Block5F must normalize which assets contribute galactic scanner coverage, for example technologies and any station/colony/ship bonuses that are actually supported by ruleset/research evidence. Coverage must be derived server-side from authoritative empire assets and technology.

Do not hard-code scanner circles or detection range in React.

### 6. Last-known contact behavior is an explicit research question

Before implementation, determine what should happen when a previously detected foreign Fleet leaves current scanner coverage:

- disappear immediately;
- persist at last-known position;
- persist with stale/estimated route/ETA;
- or another original-game behavior.

The UI must visibly distinguish current detection from stale intelligence if stale contacts are supported. Do not invent this behavior without research.

### 7. Relationship to First Contact / empire identity

Fleet detection and diplomatic First Contact are separate knowledge axes.

- Detecting a transit contact must not automatically reveal empire/race identity unless the authoritative knowledge rules say it should.
- A marker may therefore need an unknown-contact presentation while still showing only scanner-permitted movement data.
- After established contact, player-color/known-empire presentation can use the existing empire identity projection.

### 8. Fleet grouping needs an explicit transit contract

Stationary system markers are already frozen as **one Fleet icon per empire/system**. For deep-space detected traffic, Block5F must research/freeze whether contacts aggregate by empire + shared route/ETA or remain separate authoritative travelling Fleet groups. Do not blindly reuse stationary-system aggregation if it would merge independently moving contacts.

## Block5F acceptance targets

At minimum, the eventual implementation should prove:

1. foreign Fleet outside scanner coverage is absent from player projection and DOM;
2. crossing scanner range creates a server-authorized contact;
3. detectable incoming hostile Fleet renders foreign player-colored marker + red dashed route to a permitted known target;
4. remaining pc/ETA come from server projection only;
5. clicking contact opens compact read-only Fleet dialog with no mutation controls;
6. ship/detail content is limited to scanner-authorized intelligence;
7. leaving range follows the researched current-vs-last-known contract without hidden-data leakage;
8. save/load/reconnect preserves only authoritative scanner/contact state that is intended to persist;
9. mobile/touch path can open/inspect the same contact safely;
10. normal own-Fleet green route behavior remains unchanged.

## Slice ordering

- **First finish Block5E:** Human arrival -> visited reveal -> First Contact -> shared-system Fleet markers -> Tactical fire/return.
- **Then Block5F:** scanner coverage + detected foreign transit contacts + incoming warnings/read-only inspection.

Block5F is intentionally adjacent to 5E so the current Fleet/route architecture is extended while still fresh, rather than postponed to a distant slice and forcing another Galaxy-map redesign.