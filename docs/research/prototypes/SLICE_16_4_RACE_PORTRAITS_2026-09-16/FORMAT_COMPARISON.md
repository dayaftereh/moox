# Slice 16.4 Gate-1 portrait prototype format/crop comparison

Date: 2026-09-16
Subjects: Alkari / Meklar / Silicoid
Purpose: research-only composition, crop and raster-format evaluation; **not final production portrait art**.

## Sources

All three portraits are generated deterministically from `generate.mjs` as 1200 x 1500 SVG composition prototypes with one shared dark cinematic frame/light language.

The SVGs intentionally stress three different rendering/material cases:

- Alkari: curved organic avian/reptilian anatomy plus armor;
- Meklar: organic biological core plus hard-surface cybernetic exoskeleton;
- Silicoid: non-humanoid crystalline facets and internal refractive glow.

The final production portrait contract is still expected to be raster-first/painterly. SVG remains appropriate for emblems and for this deterministic technical prototype only.

## Raster comparison

Reference rasterization: headless Chrome 152 at 1200 x 1500, device scale 1.
WebP conversion: ffmpeg/libwebp quality 82.

| Race | PNG bytes | WebP bytes | WebP / PNG | SSIM PNG vs WebP |
| --- | ---: | ---: | ---: | ---: |
| Alkari | 623,008 | 41,268 | 6.6% | 0.995993 |
| Meklar | 658,644 | 41,930 | 6.4% | 0.996002 |
| Silicoid | 721,556 | 40,188 | 5.6% | 0.996042 |

Result:

- conservative WebP is roughly **15-18x smaller** than the lossless PNG reference for these prototypes;
- SSIM is ~0.996 for all three, with no prototype-specific collapse on curved organic edges, hard machinery or faceted crystal geometry;
- WebP is therefore the preferred runtime portrait candidate;
- PNG remains useful as lossless review/export/reference when needed, but should not be duplicated into production by default.

This result is specific to the prototypes; Gate 2/3 should repeat a spot check on the first final painterly asset because painterly texture/noise may compress differently.

## Square/card crop experiment

A centered 1200 x 1200 crop beginning at Y=120 was downscaled to 640 x 640.

Resulting files:

| Race | Square PNG | Square WebP |
| --- | ---: | ---: |
| Alkari | 179,554 B | 18,148 B |
| Meklar | 182,255 B | 18,528 B |
| Silicoid | 196,260 B | 17,602 B |

Visual review confirms:

- Alkari retains the avian/reptilian head, long neck, shoulder armor and artifact cue;
- Meklar retains the small biological core inside the much larger machine frame;
- Silicoid remains unmistakably non-humanoid and crystalline;
- no primary sensory/focal region is clipped;
- all three remain readable at compact card size.

The 4:5 master should remain canonical; 1:1 is a presentation crop, not a separate source asset.

## 390 px mobile hierarchy experiment

Research-only mobile review pages were rendered in headless Chrome at an exact 390 x 844 viewport.

Test card geometry:

- outer page padding 16 px;
- portrait `min(82vw, 320px)`;
- portrait aspect ratio 4:5;
- 16 px portrait radius;
- race name and two fact lines outside artwork;
- 52 x 52 previous/next controls.

Visual review result:

- all content fits the 390 px width without horizontal overflow;
- the portrait remains the dominant element;
- race name is readable before compact facts;
- two concise fact lines fit without pushing navigation off-screen;
- 52 px controls comfortably exceed the 44 px mobile target;
- Alkari/Meklar/Silicoid remain visually distinct despite the smaller presentation.

## Gate-1 art/pipeline conclusion

Freeze candidate for Gate 2:

- master portrait composition: 4:5;
- production format: WebP, with optional PNG only when lossless export/review is deliberately needed;
- no baked localized text;
- same safe central focal zone for all species, but no requirement for humanoid facial geometry;
- species identity must survive a centered 1:1 crop;
- final painterly rendering may be richer than these prototypes but should retain their hierarchy/safe-zone discipline.
