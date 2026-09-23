# Slice 17.2 Gate 4 - Acceptance Close

Date: 2026-09-23

Status: **accepted; Slice 17.2 closed with Gates 1-4 complete**

Parent slice: `docs/slices/PLANNED_17_2_BEAM_WEAPONS_MOUNT_MODIFIERS.md`

Gate-4 review evidence: `docs/research/SLICE_17_2_GATE4_WEAPON_SELECTOR_REVIEW_2026-09-22.md`

## Acceptance

The user explicitly accepted the long-running Weapon Selector / Ship Designer review and requested Gate 4 closure so the Slice-17 program can proceed to missiles/ordnance.

This acceptance closes the **design-side Beam breadth** delivered by Slice 17.2. It does not redefine the downstream Tactical boundary.

## Accepted Slice-17.2 product surface

- 13 frozen Beam weapon identities are available through the authoritative Ship Designer catalog.
- Beam technology, base damage, Space/PP, miniaturization and per-weapon modifier legality use the frozen Gate-2 contract.
- Standard / Heavy / Point Defense mount snapshots and supported Beam modifier snapshots are preserved losslessly.
- Multiple installed weapon rows can be added and edited through the authoritative selector.
- Quantity changes live on installed weapon cards rather than duplicating count controls inside the selector.
- Mount/modifier selection uses the full option card, with compact non-overlapping help controls.
- Base damage is labelled honestly as **Base damage / Basisschaden** while downstream Tactical-derived damage remains deferred.
- Heavy and Point Defense update authoritative Space/PP previews.
- Construction, persistence, design revision and immutable built-Ship/BattleSession Beam snapshots remain intact.
- Mobile and desktop Designer/selector behavior is covered by the accepted browser regression.

## Final QA evidence

The final implemented UI baseline was validated before closure with:

- `node --check web/scripts/check-reference-lab-browser.mjs`;
- reference-lab browser smoke on a fresh temporary server, including 390 px/mobile coverage;
- `npm --prefix web run build`;
- `go test ./... -count=1 -timeout 120s`;
- `go vet ./...`;
- `git diff --check`.

The last UI change additionally proves:

1. the selector has no quantity +/- controls;
2. installed weapon cards retain +/- quantity controls;
3. an installed weapon can be raised from count 1 to 2;
4. edit/reopen with Point Defense + Autofire preserves count 2.

After that gameplay/UI QA, subsequent commits up to this closeout changed only roadmap/documentation. No gameplay or web code changed after the accepted Gate-4 baseline.

## Explicit downstream boundary

Slice 17.2 does **not** claim complete Tactical Beam behavior.

- Broad Tactical combatant admission / hull and mandatory-system combat behavior is owned by planned **Slice 17.5**.
- Complete Beam hit/range/damage resolution, Heavy/Point Defense Tactical semantics, supported Beam modifier effects, Weapon Selector configured/effective-stat parity and Beam presentation/VFX are owned by planned **Slice 17.6**.
- Ordnance design breadth starts in Slice 17.3; complete Tactical ordnance is owned later by Slice 17.7.

These deferred consumers are explicit product scope, not Gate-4 defects in Slice 17.2.

## Close result

Slice 17.2 is closed. Its `_OPEN_` marker is removed.

The next child in dependency order is **Slice 17.3 - Missiles, Torpedoes and Bomb Ordnance**.
