import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const webRoot = path.resolve(here, '..')
const assetDir = path.join(webRoot, 'public', 'assets', 'new-game', 'opponent-count')
const manifest = JSON.parse(fs.readFileSync(path.join(webRoot, 'public', 'assets', 'new-game', 'manifest.json'), 'utf8'))
const api = fs.readFileSync(path.join(webRoot, 'src', 'api.ts'), 'utf8')
const app = fs.readFileSync(path.join(webRoot, 'src', 'App.tsx'), 'utf8')
const assert = (value, message) => { if (!value) throw new Error(message) }
for (let count = 1; count <= 7; count++) {
  const svg = fs.readFileSync(path.join(assetDir, `${count}.svg`), 'utf8')
  assert(svg.includes(`data-opponent-count="${count}"`), `opponent ${count} count marker missing`)
  assert(svg.includes(`data-art-signature="opponent-count-${count}"`), `opponent ${count} signature missing`)
  assert(!svg.includes('<text'), `opponent ${count} must not use runtime text elements`)
  assert(manifest.assets?.some((entry) => entry.domain === 'opponent-count' && entry.option_id === String(count) && entry.path === `/assets/new-game/opponent-count/${count}.svg`), `manifest missing opponent-count ${count}`)
}
assert(api.includes("'/api/v1/new-game/compositions'"), 'composition API binding missing')
assert(app.includes('settingId="opponent-count"'), 'opponent count selector missing')
assert(app.includes('new-game-launch-briefing'), 'launch briefing missing')
assert(app.includes('new-game-empire-roster-card'), 'full-width empire roster missing')
assert(app.includes('data-role="planned-opponent"'), 'planned opponent roster slots missing')
assert(app.includes('function RaceInfoButton'), 'per-race roster information button missing')
assert(!app.includes('new-game-opponent-list'), 'opponent selector card must not contain the old embedded roster')
console.log('Opponent composition contract passed: 7 deterministic SVGs, manifest, composition API, selector, full-width 8-slot empire roster, race info and launch briefing.')