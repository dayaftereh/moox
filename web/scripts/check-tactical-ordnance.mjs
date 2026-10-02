import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))
const webRoot = path.resolve(here, '..')
const read = (relative) => fs.readFileSync(path.join(webRoot, relative), 'utf8')

const api = read('src/api.ts')
const battlefield = read('src/TacticalBattlefield.tsx')
const styles = read('src/styles.css')

const requireFragment = (source, fragment, label) => {
  if (!source.includes(fragment)) {
    throw new Error(`Tactical ordnance browser contract missing ${label}: ${fragment}`)
  }
}

requireFragment(api, 'target_projectile_id?: number', 'projectile target DTO')
requireFragment(api, 'ordnance: TacticalOrdnanceState', 'ordnance projection DTO')
requireFragment(api, "kind: 'battle.fire_ordnance'", 'ordnance command transport')
requireFragment(api, 'tacticalFireBeamProjectileCommand', 'Beam projectile command transport')

requireFragment(battlefield, 'tactical.ordnance?.projectiles', 'authoritative projectile projection')
requireFragment(battlefield, 'tactical.ordnance?.weapon_states', 'authoritative rack/cooldown projection')
requireFragment(battlefield, 'legalProjectileTargetByID', 'server-projected projectile target eligibility')
requireFragment(battlefield, 'tacticalFireOrdnanceCommand', 'ordnance launch UI')
requireFragment(battlefield, 'tacticalFireBeamProjectileCommand', 'manual Beam projectile targeting UI')
for (const eventKind of [
  'ordnance_launched',
  'projectile_moved',
  'projectile_retargeted',
  'projectile_interception_resolved',
  'projectile_mirv_split',
  'projectile_impact_resolved',
  'projectile_destroyed',
]) {
  requireFragment(battlefield, eventKind, `event-driven ${eventKind} presentation`)
}
requireFragment(battlefield, 'tactical-projectile-hit-cell', 'projectile marker')
requireFragment(battlefield, "battlefield.ammoRemaining", 'Missile ammo presentation')
requireFragment(battlefield, "battlefield.cooldownRemaining", 'Torpedo cooldown presentation')
requireFragment(styles, '.tactical-projectile', 'projectile marker style')
requireFragment(styles, '.tactical-projectile-animation', 'projectile event animation style')

console.log('Tactical ordnance browser contract passed: authoritative projectiles, server-projected targets, launch/interception event presentation, ammo and cooldown wiring.')
