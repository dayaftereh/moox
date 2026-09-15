# Slice 16.1 Gate 2 - Freeze

Date: 2026-09-15
Status: **complete / frozen**

## Decision

User review accepted the compact Galaxy Size selector and the astronomy-informed Galaxy Size artwork set. Gate 2 is therefore closed and the reusable Slice-16 visual-selection contract is frozen for downstream implementation.

## Frozen selector contract

- One compact setting card per visual setting.
- Setting title above the artwork.
- Artwork fills the card width and is clipped with rounded corners.
- Information is kept out of the permanent card body; detailed explanation opens from the artwork `?` control.
- Current selection is shown large and centered between previous/next arrows below the artwork.
- Small position dots may communicate option position without adding text-heavy chrome.
- The selector shell remains setting-neutral and reusable by later Slice-16 settings.

## Frozen responsive layout

- Narrow/mobile widths: one setting card per row.
- Tablet/desktop: multiple visual-setting cards may flow into the responsive settings grid when space permits.
- Current reference card remains compact rather than a full-width hero.
- Current validated artwork size is about 342 x 214 px at a 390 px viewport and about 606 x 341 px at desktop reference width.
- Artwork clipping is 16 px radius on mobile and 20 px on tablet/desktop.

## Frozen interaction/accessibility behavior

- Mouse/touch previous-next controls remain first-class.
- Keyboard selector navigation supports ArrowLeft, ArrowRight, Home and End.
- Mobile touch controls keep at least a 44 px target.
- The artwork info control is a 44 px mobile target.
- Detailed information opens as a bottom sheet on narrow/mobile widths and a centered modal on larger widths.
- The information surface is modal, Escape/close dismisses it, and focus returns to the artwork info control.
- Current dialog focus remains contained while open.

## Frozen asset/runtime convention

- Semantic asset IDs follow `new-game:<domain>:<option-id>`.
- Runtime setting assets live under `/assets/new-game/<domain>/<option-id>.<format>`.
- Galaxy Size assets use deterministic original SVG artwork under `/assets/new-game/galaxy-size/*.svg`.
- Vector/procedural illustrations use SVG where appropriate; painterly/organic raster artwork such as race portraits may use WebP/PNG.
- Localized UI/accessibility copy is not embedded in the artwork.
- Asset provenance is recorded before runtime acceptance.

## Frozen supported/planned presentation

- Unsupported choices may be browsed visually when useful for planning/review.
- Unsupported choices must never be presented as server-enabled.
- Create Game remains disabled whenever the selected option is not accepted by the authoritative server contract.
- Supported/planned status and explanatory detail belong in the on-demand information surface rather than permanently expanding the card.

## Galaxy Size reference implementation accepted at freeze

- Tiny: compact dwarf-irregular visual language.
- Small: compact flocculent spiral; currently the only server-supported galaxy-size option.
- Medium: barred spiral.
- Large: luminous grand-design spiral.
- Huge: extended giant spiral.
- Relative size is communicated by footprint, stellar density, disk extent, brightness and structural richness; spiral-arm count is not treated as a direct physical size metric.

Evidence:

- `docs/research/SLICE_16_1_GATE2_SINGLE_CARD_SELECTOR_REVIEW_2026-09-15.md`
- `docs/research/SLICE_16_1_GATE2_GALAXY_SIZE_ART_2026-09-15.md`

## Content refinement boundary

Gate-2 freeze does **not** freeze the final wording depth of every information dialog. Later work may expand or improve explanatory copy (for galaxy size, age, race abilities, difficulty, technology level, etc.) as long as the frozen selector structure, interaction, supported/planned semantics and asset contract are preserved.

## Next

Proceed to Slice 16.1 Gate 3 implementation using the frozen selector grammar as the common foundation for later Slice-16 setting screens.
