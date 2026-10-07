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
  'projectile_moved',
]) {
  requireFragment(battlefield, eventKind, `event-driven ${eventKind} presentation`)
}
requireFragment(battlefield, 'tactical-projectile-hit-cell', 'projectile marker')
requireFragment(battlefield, "battlefield.ammoRemaining", 'Missile ammo presentation')
requireFragment(battlefield, "battlefield.cooldownRemaining", 'Torpedo cooldown presentation')
requireFragment(styles, '.tactical-projectile', 'projectile marker style')
requireFragment(styles, '.tactical-projectile-animation', 'projectile event animation style')
requireFragment(battlefield, '<animateTransform', 'projectile flight animation')
requireFragment(battlefield, 'tactical-projectile-flight-body', 'moving projectile glyph')
requireFragment(styles, '.tactical-projectile-flight-body', 'moving projectile style')
requireFragment(battlefield, 'tacticalMiniFormation(projectile.multiplicity)', 'exact miniature projectile salvo rendering')
requireFragment(battlefield, 'tacticalMiniFormation(projectileAnimation.multiplicity)', 'exact miniature projectile flight rendering')
requireFragment(battlefield, 'type TacticalPresentationCue', 'authoritative event presentation queue')
requireFragment(battlefield, 'presentationQueueRef', 'combat presentation queue state')
requireFragment(battlefield, 'focusCombatSpan', 'combat action camera framing')
requireFragment(battlefield, 'COMBAT_CAMERA_MAX_ZOOM = 2.8', 'restrained combat camera zoom')
requireFragment(battlefield, 'COMBAT_CUE_BEAM_MS = 1320', '20-percent slower combat camera beam hold')
requireFragment(battlefield, 'COMBAT_CUE_FIGHTER_ATTACK_MS = 1500', '20-percent slower Fighter attack hold')
requireFragment(battlefield, 'COMBAT_CUE_REACTIVE_PD_MS = 624', '20-percent slower reactive PD hold')
requireFragment(battlefield, 'COMBAT_CUE_MOVE_MS = 840', '20-percent slower movement hold')
requireFragment(battlefield, 'COMBAT_CUE_PROJECTILE_MS = 900', '20-percent slower projectile hold')
requireFragment(battlefield, 'dur="984ms"', '20-percent slower projectile flight animation')
requireFragment(styles, 'tactical-beam-flash 1.62s', '20-percent slower beam animation')
requireFragment(styles, 'tactical-fighter-shot-flash 1.104s', '20-percent slower Fighter shot animation')
requireFragment(styles, 'tactical-move-trail-fade .864s', '20-percent slower move animation')
requireFragment(battlefield, 'DRAG_THRESHOLD_PX = 4', 'mouse click drag tolerance')
requireFragment(battlefield, 'dragDistanceRef.current += Math.hypot(dx, dy)', 'cumulative pointer drag threshold')
requireFragment(battlefield, "event.kind === 'fighter_group_moved'", 'Fighter movement camera cue')
requireFragment(battlefield, "event.kind === 'fighter_attack_resolved'", 'Fighter attack camera cue')
requireFragment(battlefield, "transform={'translate(' + projectileAnimation.fromX + ' ' + projectileAnimation.fromY + ')'}", 'projectile flight static start transform')
requireFragment(battlefield, 'from="0 0"', 'projectile flight relative animation origin')
requireFragment(battlefield, 'projectileAnimation.toX - projectileAnimation.fromX', 'projectile flight relative X delta')
requireFragment(battlefield, 'projectileAnimation.toY - projectileAnimation.fromY', 'projectile flight relative Y delta')
requireFragment(battlefield, 'data-multiplicity={projectileAnimation.multiplicity}', 'animated salvo multiplicity')
requireFragment(styles, '.tactical-projectile-flight-flame', 'missile exhaust flame style')
requireFragment(battlefield, 'tactical-projectile-static-plume', 'static missile propulsion plume')
requireFragment(styles, '.tactical-projectile-static-plume', 'static missile propulsion plume style')
requireFragment(battlefield, 'tactical-projectile-energy-spine', 'missile sci-fi energy spine')
requireFragment(battlefield, 'tactical-fighter-exhaust', 'Fighter-family propulsion plume')
requireFragment(styles, '.tactical-fighter-exhaust', 'Fighter-family propulsion plume style')
if (battlefield.includes('tactical-projectile-count') || battlefield.includes('tactical-projectile-flight-count')) {
  throw new Error('Tactical ordnance browser contract forbids visible projectile xN count labels; multiplicity must be rendered as miniature missile glyphs.')
}


console.log('Tactical ordnance browser contract passed: authoritative projectiles, server-projected targets, launch/interception event presentation, ammo and cooldown wiring.')
