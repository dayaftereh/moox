import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { execFileSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const webRoot = path.resolve(here, '..')
const repoRoot = path.resolve(webRoot, '..')
const appPath = path.join(webRoot, 'src', 'App.tsx')
const apiPath = path.join(webRoot, 'src', 'api.ts')
const i18nPath = path.join(webRoot, 'src', 'i18n.tsx')
const stylesPath = path.join(webRoot, 'src', 'styles.css')
const manifestPath = path.join(webRoot, 'public', 'assets', 'new-game', 'manifest.json')
const generatorPath = path.join(webRoot, 'scripts', 'generate-technology-level-art.mjs')
const assetDir = path.join(webRoot, 'public', 'assets', 'new-game', 'technology-level')
const expected = [
  ['pre_warp', 'pre-warp', 'surface-launchpad'],
  ['average', 'average', 'orbital-fleet'],
  ['advanced', 'advanced', 'hyperlane-network'],
]

function fail(message) {
  console.error(`Technology start contract failed: ${message}`)
  process.exit(1)
}

function assert(condition, message) {
  if (!condition) fail(message)
}

const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'))
const entries = manifest.entries.filter((entry) => entry.domain === 'technology-level')
assert(entries.length === 3, `manifest technology-level entries=${entries.length} want=3`)

for (const [id, assetID, signature] of expected) {
  const entry = entries.find((candidate) => candidate.option_id === assetID)
  assert(entry, `missing manifest entry for ${assetID}`)
  assert(entry.id === `new-game:technology-level:${assetID}`, `unexpected semantic id for ${assetID}: ${entry.id}`)
  assert(entry.format === 'svg', `${assetID} format=${entry.format} want=svg`)
  assert(entry.generator === 'web/scripts/generate-technology-level-art.mjs', `${assetID} generator mismatch`)
  const assetPath = path.join(assetDir, `${assetID}.svg`)
  const svg = fs.readFileSync(assetPath, 'utf8')
  assert(svg.includes('width="1200" height="675"'), `${assetID} dimensions are not 1200x675`)
  assert(svg.includes('<title>') && svg.includes('<desc>'), `${assetID} missing title/desc`)
  assert(!svg.includes('<text'), `${assetID} must not contain rasterizing text elements`)
  assert(svg.includes(`data-art-signature="${signature}"`), `${assetID} missing distinct composition signature ${signature}`)
  assert(id === 'pre_warp' || assetID === id, `unexpected id mapping ${id} -> ${assetID}`)
}

const app = fs.readFileSync(appPath, 'utf8')
const api = fs.readFileSync(apiPath, 'utf8')
const i18n = fs.readFileSync(i18nPath, 'utf8')
const styles = fs.readFileSync(stylesPath, 'utf8')
assert(app.includes('VisualSelector<NewGameTechnologyLevel>'), 'typed technology VisualSelector binding missing')
assert(app.includes('getTechnologyCatalog(controller.signal)'), 'technology catalog fetch missing')
assert(app.includes('technology_level: technologyLevelID'), 'Create Game is not bound to selected technology level')
assert(app.includes("const technologyLevelAssetVersion = 'slice16-5-g3-art2'"), 'technology artwork cache-busting version missing')
assert(app.includes("technologyProfilesByID.get(technologyLevelID)?.availability !== 'supported'"), 'Create Game support-boundary lock missing')
assert(api.includes("export type NewGameTechnologyLevel = 'pre_warp' | 'average' | 'advanced'"), 'typed technology-level API union missing')
assert(api.includes("'/api/v1/new-game/technologies'"), 'technology catalog endpoint binding missing')
assert(i18n.includes("'newGame.techAverage': 'Durchschnitt'"), 'compact German Average label missing')
assert(i18n.includes("'newGame.techAdvanced': 'Fortschrittlich'"), 'compact German Advanced label missing')
assert(styles.includes('.visual-selector[data-setting-id="technology-level"] .visual-selector-current h3'), 'technology-specific one-line title rule missing')
assert(styles.includes('white-space: nowrap'), 'technology title no-wrap rule missing')
assert(/\.new-game-galaxy-art img,\s*\.new-game-difficulty-art img,\s*\.new-game-technology-art img,\s*\.new-game-opponent-count-art img\s*\{[^}]*width:\s*100%;[^}]*height:\s*100%;[^}]*object-fit:\s*contain;[^}]*object-position:\s*center;/s.test(styles), 'selector artwork must scale fully without cropping')

const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'moox-tech-art-'))
try {
  execFileSync(process.execPath, [generatorPath], {
    cwd: repoRoot,
    env: { ...process.env, MOOX_TECHNOLOGY_ART_OUT_DIR: tmp },
    stdio: 'pipe',
  })
  for (const [, assetID] of expected) {
    const committed = fs.readFileSync(path.join(assetDir, `${assetID}.svg`))
    const regenerated = fs.readFileSync(path.join(tmp, `${assetID}.svg`))
    assert(committed.equals(regenerated), `${assetID} is not reproducible from generator`)
  }
} finally {
  fs.rmSync(tmp, { recursive: true, force: true })
}

console.log('Technology start contract passed: 3 server-bound options, 3 distinct/reproducible SVG compositions, compact one-line labels, and Advanced planned/locked.')