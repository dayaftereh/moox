import crypto from 'node:crypto'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { spawnSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const webRoot = path.resolve(here, '..')
const repoRoot = path.resolve(webRoot, '..')
const manifestPath = path.join(webRoot, 'public', 'assets', 'races', 'manifest.json')
const sourceDir = path.join(repoRoot, 'docs', 'research', 'prototypes', 'SLICE_16_4_RUNTIME_PORTRAITS_2026-09-16')
const generatorPath = path.join(here, 'generate-race-portrait-sources.mjs')
const ids = ['human', 'klackon', 'darlok']
const failures = []
const assert = (condition, message) => { if (!condition) failures.push(message) }
const sha256 = (buffer) => crypto.createHash('sha256').update(buffer).digest('hex')

assert(fs.existsSync(manifestPath), `missing race manifest ${manifestPath}`)
if (fs.existsSync(manifestPath)) {
  const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'))
  assert(manifest.schema_version === 1, `race manifest schema_version=${manifest.schema_version}`)
  assert(Array.isArray(manifest.entries), 'race manifest entries missing')
  if (Array.isArray(manifest.entries)) {
    assert(manifest.entries.length === ids.length, `expected ${ids.length} race manifest entries, found ${manifest.entries.length}`)
    for (const id of ids) {
      const entry = manifest.entries.find((candidate) => candidate.race_id === id && candidate.kind === 'portrait')
      assert(Boolean(entry), `missing portrait manifest entry for ${id}`)
      if (!entry) continue
      assert(entry.id === `race:${id}:portrait`, `${id} invalid semantic id ${entry.id}`)
      assert(entry.format === 'webp', `${id} invalid format ${entry.format}`)
      assert(entry.path === `/assets/races/${id}/portrait.webp`, `${id} invalid path ${entry.path}`)
      assert(entry.alt_name_key === `race.${id}.name`, `${id} invalid alt_name_key ${entry.alt_name_key}`)
      assert(entry.provenance === 'original-procedural-mox', `${id} invalid provenance ${entry.provenance}`)
      for (const key of ['x', 'y']) assert(Number.isFinite(entry.focal_point?.[key]), `${id} missing focal_point.${key}`)
      for (const key of ['x', 'y', 'width', 'height']) assert(Number.isFinite(entry.safe_area?.[key]), `${id} missing safe_area.${key}`)
      const runtimePath = path.join(webRoot, 'public', entry.path.replace(/^\/assets\//, 'assets/'))
      assert(fs.existsSync(runtimePath), `missing runtime portrait ${runtimePath}`)
      if (fs.existsSync(runtimePath)) {
        const data = fs.readFileSync(runtimePath)
        assert(data.length > 10000, `${id} portrait suspiciously small: ${data.length} bytes`)
        assert(data.subarray(0, 4).toString('ascii') === 'RIFF' && data.subarray(8, 12).toString('ascii') === 'WEBP', `${id} is not RIFF/WEBP`)
      }
    }
  }
}

for (const id of ids) {
  const sourcePath = path.join(sourceDir, `${id}.svg`)
  assert(fs.existsSync(sourcePath), `missing source SVG ${sourcePath}`)
  if (fs.existsSync(sourcePath)) {
    const svg = fs.readFileSync(sourcePath, 'utf8')
    assert(svg.includes('width="1200" height="1500"'), `${id}.svg missing 1200x1500 dimensions`)
    assert(svg.includes('viewBox="0 0 1200 1500"'), `${id}.svg missing 4:5 viewBox`)
    assert(!/<text\b/i.test(svg), `${id}.svg contains baked text`)
  }
}

const temp = fs.mkdtempSync(path.join(os.tmpdir(), 'moox-race-portrait-check-'))
const generated = spawnSync(process.execPath, [generatorPath, temp], { encoding: 'utf8' })
assert(generated.status === 0, `race portrait source generator failed: ${generated.stderr || generated.stdout}`)
if (generated.status === 0) {
  for (const id of ids) {
    const committed = fs.readFileSync(path.join(sourceDir, `${id}.svg`))
    const regenerated = fs.readFileSync(path.join(temp, `${id}.svg`))
    assert(sha256(committed) === sha256(regenerated), `${id}.svg is not deterministic from generator`)
  }
}
fs.rmSync(temp, { recursive: true, force: true })

if (failures.length) {
  console.error(failures.map((failure) => `- ${failure}`).join('\n'))
  process.exit(1)
}
console.log(`Race asset contract passed: ${ids.length} deterministic 4:5 sources + WebP runtime portraits.`)
