import crypto from 'node:crypto'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { spawnSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const webRoot = path.resolve(here, '..')
const repoRoot = path.resolve(webRoot, '..')
const assetDir = path.join(webRoot, 'public', 'assets', 'new-game', 'galaxy-size')
const manifestPath = path.join(webRoot, 'public', 'assets', 'new-game', 'manifest.json')
const appPath = path.join(webRoot, 'src', 'App.tsx')
const selectorPath = path.join(webRoot, 'src', 'components', 'VisualSelector.tsx')
const stylesPath = path.join(webRoot, 'src', 'styles.css')
const generatorPath = path.join(here, 'generate-galaxy-size-art.mjs')
const difficultyAssetDir = path.join(webRoot, 'public', 'assets', 'new-game', 'difficulty')
const difficultyGeneratorPath = path.join(here, 'generate-difficulty-art.mjs')
const galaxyAgeAssetDir = path.join(webRoot, 'public', 'assets', 'new-game', 'galaxy-age')
const galaxyAgeGeneratorPath = path.join(here, 'generate-galaxy-age-art.mjs')

const ids = ['tiny', 'small', 'medium', 'large', 'huge']
const difficultyAssets = [['easy', 'easy'], ['normal', 'normal'], ['hard', 'hard'], ['very_hard', 'very-hard'], ['impossible', 'impossible']]
const difficultyIds = difficultyAssets.map(([id]) => id)
const galaxyAgeAssets = [['mineral_rich', 'mineral-rich'], ['normal', 'normal'], ['organic_rich', 'organic-rich']]
const failures = []

function assert(condition, message) {
  if (!condition) failures.push(message)
}

function sha256(buffer) {
  return crypto.createHash('sha256').update(buffer).digest('hex')
}

const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'))
const entries = manifest.entries.filter((entry) => entry.domain === 'galaxy-size')
assert(entries.length === ids.length, `expected ${ids.length} galaxy-size manifest entries, found ${entries.length}`)

for (const id of ids) {
  const expectedPath = `/assets/new-game/galaxy-size/${id}.svg`
  const entry = entries.find((candidate) => candidate.option_id === id)
  assert(Boolean(entry), `missing manifest entry for ${id}`)
  if (entry) {
    assert(entry.id === `new-game:galaxy-size:${id}`, `invalid semantic id for ${id}`)
    assert(entry.format === 'svg', `invalid format for ${id}: ${entry.format}`)
    assert(entry.path === expectedPath, `invalid runtime path for ${id}: ${entry.path}`)
    assert(entry.provenance === 'original-procedural-mox', `invalid provenance for ${id}`)
  }

  const assetPath = path.join(assetDir, `${id}.svg`)
  assert(fs.existsSync(assetPath), `missing SVG asset ${assetPath}`)
  if (!fs.existsSync(assetPath)) continue
  const svg = fs.readFileSync(assetPath, 'utf8')
  assert(svg.includes('width="1200" height="675"'), `${id}.svg missing frozen native dimensions`)
  assert(svg.includes('viewBox="0 0 1200 675"'), `${id}.svg missing frozen 16:9 viewBox`)
  assert(svg.includes('<title'), `${id}.svg missing title metadata`)
  assert(svg.includes('<desc'), `${id}.svg missing description metadata`)
  assert(!/<text\b/i.test(svg), `${id}.svg embeds text; localized UI copy must stay outside artwork`)
}

const difficultyEntries = manifest.entries.filter((entry) => entry.domain === 'difficulty')
assert(difficultyEntries.length === difficultyIds.length, `expected ${difficultyIds.length} difficulty manifest entries, found ${difficultyEntries.length}`)
for (const [id, assetId] of difficultyAssets) {
  const expectedPath = `/assets/new-game/difficulty/${assetId}.svg`
  const entry = difficultyEntries.find((candidate) => candidate.option_id === assetId)
  assert(Boolean(entry), `missing difficulty manifest entry for ${id}`)
  if (entry) {
    assert(entry.id === `new-game:difficulty:${assetId}`, `invalid difficulty semantic id for ${id}`)
    assert(entry.format === 'svg', `invalid difficulty format for ${id}: ${entry.format}`)
    assert(entry.path === expectedPath, `invalid difficulty runtime path for ${id}: ${entry.path}`)
    assert(entry.provenance === 'original-procedural-mox', `invalid difficulty provenance for ${id}`)
    assert(entry.generator === 'web/scripts/generate-difficulty-art.mjs', `invalid difficulty generator for ${id}`)
  }
  const assetPath = path.join(difficultyAssetDir, `${assetId}.svg`)
  assert(fs.existsSync(assetPath), `missing Difficulty SVG asset ${assetPath}`)
  if (!fs.existsSync(assetPath)) continue
  const svg = fs.readFileSync(assetPath, 'utf8')
  assert(svg.includes('width="1200" height="675"'), `difficulty ${id}.svg missing frozen native dimensions`)
  assert(svg.includes('viewBox="0 0 1200 675"'), `difficulty ${id}.svg missing frozen 16:9 viewBox`)
  assert(svg.includes('<title'), `difficulty ${id}.svg missing title metadata`)
  assert(svg.includes('<desc'), `difficulty ${id}.svg missing description metadata`)
  assert(!/<text\b/i.test(svg), `difficulty ${id}.svg embeds text; localized UI copy must stay outside artwork`)
}


const galaxyAgeEntries = manifest.entries.filter((entry) => entry.domain === 'galaxy-age')
assert(galaxyAgeEntries.length === galaxyAgeAssets.length, `expected ${galaxyAgeAssets.length} galaxy-age manifest entries, found ${galaxyAgeEntries.length}`)
for (const [id, assetId] of galaxyAgeAssets) {
  const expectedPath = `/assets/new-game/galaxy-age/${assetId}.svg`
  const entry = galaxyAgeEntries.find((candidate) => candidate.option_id === assetId)
  assert(Boolean(entry), `missing Galaxy Age manifest entry for ${id}`)
  if (entry) {
    assert(entry.id === `new-game:galaxy-age:${assetId}`, `invalid Galaxy Age semantic id for ${id}`)
    assert(entry.format === 'svg', `invalid Galaxy Age format for ${id}: ${entry.format}`)
    assert(entry.path === expectedPath, `invalid Galaxy Age runtime path for ${id}: ${entry.path}`)
    assert(entry.provenance === 'original-procedural-mox', `invalid Galaxy Age provenance for ${id}`)
    assert(entry.generator === 'web/scripts/generate-galaxy-age-art.mjs', `invalid Galaxy Age generator for ${id}`)
  }
  const assetPath = path.join(galaxyAgeAssetDir, `${assetId}.svg`)
  assert(fs.existsSync(assetPath), `missing Galaxy Age SVG asset ${assetPath}`)
  if (!fs.existsSync(assetPath)) continue
  const svg = fs.readFileSync(assetPath, 'utf8')
  assert(svg.includes('width="1200" height="675"'), `Galaxy Age ${id}.svg missing frozen native dimensions`)
  assert(svg.includes('viewBox="0 0 1200 675"'), `Galaxy Age ${id}.svg missing frozen 16:9 viewBox`)
  assert(svg.includes('<title'), `Galaxy Age ${id}.svg missing title metadata`)
  assert(svg.includes('<desc'), `Galaxy Age ${id}.svg missing description metadata`)
  assert(!/<text\b/i.test(svg), `Galaxy Age ${id}.svg embeds text; localized UI copy must stay outside artwork`)
}

const app = fs.readFileSync(appPath, 'utf8')
const selector = fs.readFileSync(selectorPath, 'utf8')
const styles = fs.readFileSync(stylesPath, 'utf8')

assert(app.includes('<VisualSelector<GalaxySizeID>'), 'App does not use the typed GalaxySizeID VisualSelector contract')
assert(app.includes('<VisualSelector<GalaxyAgeID>'), 'App does not use the typed GalaxyAgeID VisualSelector contract')
assert(!app.includes('GalaxyPrototypeSize'), 'App still contains the retired GalaxyPrototypeSize prototype contract')
assert(app.includes("const galaxySizeIDs: readonly GalaxySizeID[] = ['small', 'medium', 'large', 'huge']"), 'App bound Galaxy Size IDs do not match the frozen four-option contract')
assert(!app.includes("['tiny', t('newGame.sizeTiny')"), 'App still binds Tiny into the production Galaxy Size selector')
assert(app.includes("newGameAssetPath('galaxy-size', size, 'svg')"), 'App does not resolve Galaxy Size art through newGameAssetPath')
assert(app.includes("newGameAssetPath('galaxy-age', galaxyAgeAssetOptionID(age), 'svg')"), 'App does not resolve Galaxy Age art through newGameAssetPath')
assert(app.includes('getGalaxyCatalog(controller.signal)'), 'App does not load the authoritative Galaxy catalog')
assert(app.includes('galaxy_size: galaxySizeID'), 'Create Game does not submit selected galaxy_size')
assert(app.includes('galaxy_age: galaxyAgeID'), 'Create Game does not submit selected galaxy_age')
assert(app.includes('<VisualSelector<PresetRaceID>'), 'App does not use the typed PresetRaceID VisualSelector contract')
assert(app.includes('getRaceCatalog(controller.signal)'), 'App does not load the authoritative Race catalog')
assert(app.includes('settingId="player-race"'), 'App does not bind the semantic player-race selector')
assert(app.includes('race_id: playerRaceID'), 'Create Game does not submit selected player race')
assert(app.includes("raceProfilesByID.get(playerRaceID)?.player_availability !== 'supported'"), 'Create Game is not locked by authoritative race support state')
assert(styles.includes('.visual-selector[data-setting-id="player-race"] .visual-selector-art'), '4:5 player-race artwork style contract missing')
assert(selector.includes('VisualSelectorOption<TId extends string = string>'), 'VisualSelector option ID contract is not generic')
assert(selector.includes('VisualSelectorProps<TId extends string>'), 'VisualSelector props are not typed by option ID')
assert(selector.includes("event.key === 'ArrowLeft'"), 'VisualSelector missing ArrowLeft navigation')
assert(selector.includes("event.key === 'ArrowRight'"), 'VisualSelector missing ArrowRight navigation')
assert(selector.includes("event.key === 'Home'"), 'VisualSelector missing Home navigation')
assert(selector.includes("event.key === 'End'"), 'VisualSelector missing End navigation')
assert(selector.includes('aria-modal="true"'), 'VisualSelector info surface is not modal')
assert(selector.includes('aria-roledescription="carousel"'), 'VisualSelector carousel semantics missing')
assert(styles.includes('grid-template-columns: repeat(auto-fit, minmax(min(100%, 340px), 1fr));'), 'responsive New Game settings grid contract changed')
assert(styles.includes('.visual-selector-info-button'), 'info-button style contract missing')
assert(styles.includes('min-width: 44px;'), '44px mobile touch target contract missing')

const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'moox-galaxy-art-'))
try {
  const generated = spawnSync(process.execPath, [generatorPath], {
    cwd: webRoot,
    env: { ...process.env, MOOX_GALAXY_ART_OUT_DIR: tempDir },
    encoding: 'utf8',
  })
  assert(generated.status === 0, `isolated galaxy-art generation failed: ${generated.stderr || generated.stdout}`)
  if (generated.status === 0) {
    for (const id of ids) {
      const currentPath = path.join(assetDir, `${id}.svg`)
      const regeneratedPath = path.join(tempDir, `${id}.svg`)
      assert(fs.existsSync(regeneratedPath), `generator did not produce ${id}.svg`)
      if (!fs.existsSync(currentPath) || !fs.existsSync(regeneratedPath)) continue
      const currentHash = sha256(fs.readFileSync(currentPath))
      const regeneratedHash = sha256(fs.readFileSync(regeneratedPath))
      assert(currentHash === regeneratedHash, `${id}.svg is not reproducible from the committed generator`)
    }
  }
} finally {
  fs.rmSync(tempDir, { recursive: true, force: true })
}

const difficultyTempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'moox-difficulty-art-'))
try {
  const generated = spawnSync(process.execPath, [difficultyGeneratorPath], {
    cwd: webRoot,
    env: { ...process.env, MOOX_DIFFICULTY_ART_OUT_DIR: difficultyTempDir },
    encoding: 'utf8',
  })
  assert(generated.status === 0, `isolated difficulty-art generation failed: ${generated.stderr || generated.stdout}`)
  if (generated.status === 0) {
    for (const [id, assetId] of difficultyAssets) {
      const currentPath = path.join(difficultyAssetDir, `${assetId}.svg`)
      const regeneratedPath = path.join(difficultyTempDir, `${assetId}.svg`)
      assert(fs.existsSync(regeneratedPath), `difficulty generator did not produce ${id}.svg`)
      if (!fs.existsSync(currentPath) || !fs.existsSync(regeneratedPath)) continue
      assert(sha256(fs.readFileSync(currentPath)) === sha256(fs.readFileSync(regeneratedPath)), `difficulty ${id}.svg is not reproducible from the committed generator`)
    }
  }
} finally {
  fs.rmSync(difficultyTempDir, { recursive: true, force: true })
}

const galaxyAgeTempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'moox-galaxy-age-art-'))
try {
  const generated = spawnSync(process.execPath, [galaxyAgeGeneratorPath], {
    cwd: webRoot,
    env: { ...process.env, MOOX_GALAXY_AGE_ART_OUT_DIR: galaxyAgeTempDir },
    encoding: 'utf8',
  })
  assert(generated.status === 0, `isolated Galaxy Age art generation failed: ${generated.stderr || generated.stdout}`)
  if (generated.status === 0) {
    for (const [id, assetId] of galaxyAgeAssets) {
      const currentPath = path.join(galaxyAgeAssetDir, `${assetId}.svg`)
      const regeneratedPath = path.join(galaxyAgeTempDir, `${assetId}.svg`)
      assert(fs.existsSync(regeneratedPath), `Galaxy Age generator did not produce ${assetId}.svg`)
      if (!fs.existsSync(currentPath) || !fs.existsSync(regeneratedPath)) continue
      assert(sha256(fs.readFileSync(currentPath)) === sha256(fs.readFileSync(regeneratedPath)), `Galaxy Age ${id}.svg is not reproducible from the committed generator`)
    }
  }
} finally {
  fs.rmSync(galaxyAgeTempDir, { recursive: true, force: true })
}

if (failures.length > 0) {
  console.error(`New Game selector contract check failed (${failures.length}):`)
  for (const failure of failures) console.error(`- ${failure}`)
  process.exit(1)
}

console.log(`New Game selector contract check passed: ${ids.length + difficultyIds.length + galaxyAgeAssets.length} deterministic SVG assets across Galaxy Size, Difficulty and Galaxy Age plus authoritative typed Player Race binding, responsive/accessibility invariants.`)
console.log(`Repository: ${repoRoot}`)
