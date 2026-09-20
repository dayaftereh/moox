import type { MilitaryDesignerWeaponChoice, ShipWeaponModifierID, ShipWeaponMountType } from './api'
import type { TranslationKey, TranslationVars } from './i18n'

type Translator = (key: TranslationKey, vars?: TranslationVars) => string

const weaponTranslationKeys: Record<string, TranslationKey> = {
  mass_driver: 'shipbuilder.weapon.massDriver',
  gauss_cannon: 'shipbuilder.weapon.gaussCannon',
  laser_cannon: 'shipbuilder.weapon.laserCannon',
  particle_beam: 'shipbuilder.weapon.particleBeam',
  fusion_beam: 'shipbuilder.weapon.fusionBeam',
  ion_pulse_cannon: 'shipbuilder.weapon.ionPulseCannon',
  graviton_beam: 'shipbuilder.weapon.gravitonBeam',
  neutron_blaster: 'shipbuilder.weapon.neutronBlaster',
  phasor: 'shipbuilder.weapon.phasor',
  disrupter_cannon: 'shipbuilder.weapon.disrupterCannon',
  death_ray: 'shipbuilder.weapon.deathRay',
  plasma_cannon: 'shipbuilder.weapon.plasmaCannon',
  mauler_device: 'shipbuilder.weapon.maulerDevice',
}

export const mountTranslationKeys: Record<ShipWeaponMountType, TranslationKey> = {
  standard: 'shipbuilder.mount.standard',
  heavy: 'shipbuilder.mount.heavy',
  point_defense: 'shipbuilder.mount.pointDefense',
}

export const modifierTranslationKeys: Record<ShipWeaponModifierID, TranslationKey> = {
  armor_piercing: 'shipbuilder.modifier.armorPiercing',
  continuous: 'shipbuilder.modifier.continuous',
  no_range_dissipation: 'shipbuilder.modifier.noRangeDissipation',
  shield_piercing: 'shipbuilder.modifier.shieldPiercing',
  autofire: 'shipbuilder.modifier.autofire',
  enveloping: 'shipbuilder.modifier.enveloping',
}

export function weaponLabel(t: Translator, weapon: Pick<MilitaryDesignerWeaponChoice, 'id'>): string {
  const key = weaponTranslationKeys[weapon.id]
  return key ? t(key) : weapon.id.replaceAll('_', ' ')
}

export function mountLabel(t: Translator, mountType: ShipWeaponMountType | undefined): string {
  return t(mountTranslationKeys[mountType || 'standard'])
}

export function modifierLabel(t: Translator, modifierID: ShipWeaponModifierID): string {
  return t(modifierTranslationKeys[modifierID])
}

export function weaponDamageLabel(weapon: Pick<MilitaryDesignerWeaponChoice, 'min_damage' | 'max_damage'>): string {
  return weapon.min_damage === weapon.max_damage
    ? String(weapon.max_damage)
    : `${weapon.min_damage}–${weapon.max_damage}`
}
