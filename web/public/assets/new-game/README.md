# New Game runtime assets

Stable manifest: `/assets/new-game/manifest.json` (repository path `web/public/assets/new-game/manifest.json`).

## Semantic IDs

New Game setting art uses stable lowercase kebab-case IDs:

`new-game:<domain>:<option-id>`

Frozen domains for Slice 16:

- `difficulty`
- `galaxy-size`
- `galaxy-age`
- `technology-level`
- `player-count`

## Runtime paths

Setting art is stored as:

`web/public/assets/new-game/<domain>/<option-id>.<svg|webp|png>`

and served as:

`/assets/new-game/<domain>/<option-id>.<svg|webp|png>`

Source paths remain semantic and stable; do not put build hashes, display copy, translated labels or ordinal positions into filenames.

Race identity remains canonical under the Slice-16.4 race asset tree rather than being duplicated under New Game:

- `web/public/assets/races/<race-id>/portrait.webp` (or `.png` when intentionally chosen)
- `web/public/assets/races/<race-id>/emblem.svg`

The New Game manifest may reference those race assets when composition/race selectors land.

## Format policy

- SVG: frames, arrows, badges, symbolic setting art, diagrams, emblems, number art and deterministic vector-native procedural illustrations such as the Galaxy Size set.
- WebP/PNG: painterly or organic raster artwork such as race portraits and texture-heavy illustrations.
- Accessibility text lives in the localized UI/catalog and never depends on raster text embedded in artwork.
- Generated/original-art provenance belongs in the relevant Slice evidence/art workflow before a runtime asset is accepted.

Slice 16.1 now registers the five generated Galaxy Size SVGs in the manifest. Later 16.x slices add entries only for assets that actually exist in the repository.
