# Slice 17.1 Gate 3 - Original Ship Designer summary/layout amendment

Date: 2026-09-20
Status: Gate-3 HMI amendment

## Trigger

A mobile/Desktop review against a supplied original Master of Orion II Ship Designer screenshot identified four information-hierarchy issues in the current MOOX Designer:

1. Hull class and Command Points were repeated in the Ship Information card even though the hull selector already exposes them.
2. Production cost was shown near mandatory systems instead of next to the final Save/Build action.
3. Used/total design space was shown near mandatory systems even though remaining space is primarily relevant while adding weapons and later modifications.
4. The original screen exposes Beam Defense and Missile Evasion as player-facing derived ship ratings.

The supplied original screenshot shows a Titan with Anti-Matter Drive, 500 total space, 121 Space Available, Combat Speed 13, Beam Defense +65 and Missile Evasion 0%.

## Original-rule evidence

The established MOO2 mechanics evidence used for this amendment gives the following design-screen relationship:

- hull base combat speed by size: Frigate 10, Destroyer 8, Cruiser 6, Battleship 5, Titan 4, Doom Star 3;
- drive combat-speed bonus: Nuclear +0, Fusion +2, Ion +4, Anti-Matter +6, Hyper +8, Interphased +10;
- unused-space speed bonus: ceil(10 * unusedSpace / hullSpace);
- Beam Defense contribution from combat speed: combatSpeed * 5;
- current Slice-17.1 scope has no ECM jammer, inertial system, crew-level or racial modifier wired into the Designer, so base Missile Evasion is 0%.

The supplied screenshot is a useful independent sanity vector:

- Titan base combat speed = 4;
- Anti-Matter Drive bonus = 6;
- unused-space bonus = ceil(10 * 121 / 500) = 3;
- displayed combat speed = 4 + 6 + 3 = 13;
- displayed Beam Defense = 13 * 5 = +65;
- displayed Missile Evasion = 0%.

This reproduces the visible original values exactly.

## HMI change

The Ship Information card now focuses on installed mandatory systems and derived combat ratings:

- Drive;
- Computer;
- Armor;
- Shield;
- Fuel Cell;
- Beam Defense;
- Missile Evasion.

The redundant four-stat summary row is removed:

- Hull class: removed here because the hull selector already shows it;
- Command Points: removed here because the hull selector already shows it;
- Production Cost: moved to the Save Design footer;
- Design Space: replaced by remaining Space Available in the weapon/component section.

The weapon/component area now keeps remaining space beside the place where the player consumes it. The Save Design footer now keeps total production cost beside the final save action, mirroring the original information hierarchy more closely without copying its desktop-only layout literally.

## Authority boundary

The original-rule constants needed by the Designer are projected from the Go catalog:

- per-hull base combat speed;
- maximum unused-space combat-speed bonus = 10;
- Beam Defense per combat-speed point = 5;
- base Missile Evasion = 0 for the current supported design scope.

The current Slice-15.5 tactical implementation is deliberately unchanged. It still has its narrower combat baseline, including its existing Beam Defense behavior, and it does not yet consume a Missile Evasion field. This amendment therefore makes the Ship Designer display the verified original derived design values without claiming unsupported tactical simulation.

Future tactical slices may reuse the verified rule data when those systems are implemented authoritatively.

## Acceptance

Mobile and desktop browser QA must verify:

- no old Hull / Production Cost / Command Points / Design Space summary grid remains inside Ship Information;
- all five mandatory-system rows and their detail affordances remain;
- Beam Defense and Missile Evasion are visible under Ship Information;
- Space Available is visible in the weapon/component section;
- Production Cost is visible beside Save Design;
- no horizontal overflow at 390px;
- existing Computer/Shield popup selectors, explicit no-Shield save and hull-size progression remain intact;
- Gate 4 remains closed and Slice 17.2+ remain unopened.
