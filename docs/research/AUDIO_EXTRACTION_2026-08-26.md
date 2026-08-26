# MOO2 RIFF/WAVE extraction - checkpoint 2026-08-26

The evidence-based block scan found 96 RIFF blocks. All 96 use `WAVE` as RIFF form type, so they can be copied losslessly without transcoding.

Private output:

```text
reference/original/audio/
```

## Result

- 96 WAVE files
- 52.54 MiB original audio payload
- all detected format tags are PCM (`1`)
- all are 8-bit samples

Format distribution:

| Count | Channels | Sample rate | Bits |
| ---: | ---: | ---: | ---: |
| 57 | 2 | 22050 Hz | 8 |
| 35 | 1 | 22050 Hz | 8 |
| 2 | 1 | 10989 Hz | 8 |
| 1 | 2 | 44100 Hz | 8 |
| 1 | 2 | 11000 Hz | 8 |

Source distribution:

- `SOUND.LBX`: 68 WAV blocks
- `STREAM.LBX`: 8 WAV blocks
- `STREAMHD.LBX`: 20 WAV blocks

## Provenance

`manifest.json` records for each WAV:

- source archive and block,
- original byte size,
- SHA-256,
- output path,
- format tag,
- channel count,
- sample rate,
- byte rate,
- block alignment,
- bits per sample.

A sample post-export hash comparison for `SOUND.LBX` block 1 matches the original block exactly.

These files are original copyrighted game audio and stay under ignored `reference/original/`; they are reference material only, not distributable MOOX assets.
