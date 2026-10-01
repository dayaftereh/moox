package beam

import (
	"fmt"

	"moox/internal/core"
	"moox/internal/ruleset"
)

type RangeProfile string

const (
	RangeProfileStandard     RangeProfile = "standard"
	RangeProfileHeavy        RangeProfile = "heavy"
	RangeProfilePointDefense RangeProfile = "point_defense"
)

type ConfiguredPreview struct {
	Family                string       `json:"family"`
	BaseMinDamage         int          `json:"base_min_damage"`
	BaseMaxDamage         int          `json:"base_max_damage"`
	ConfiguredMinDamage   int          `json:"configured_min_damage"`
	ConfiguredMaxDamage   int          `json:"configured_max_damage"`
	ShotsPerWeapon        int          `json:"shots_per_weapon"`
	RangeProfile          RangeProfile `json:"range_profile"`
	StaticToHitAdjustment int          `json:"static_to_hit_adjustment"`
	EffectIDs             []string     `json:"effect_ids,omitempty"`
}

// RoundPercent reproduces the signed nearest-integer percentage scaling used by
// the original Beam damage path before the minimum/max clamps are applied.
func RoundPercent(value, percent int) int {
	return (value*percent + 50) / 100
}

func ScaleDamageBounds(baseMin, baseMax, percent int) (int, int) {
	minDamage := RoundPercent(baseMin, percent)
	maxDamage := RoundPercent(baseMax, percent)
	if minDamage < 1 {
		minDamage = 1
	}
	if maxDamage < minDamage {
		maxDamage = minDamage
	}
	return minDamage, maxDamage
}

func DeriveConfiguredPreview(weapon ruleset.TacticalWeaponRule, mount core.ShipWeaponMount) (ConfiguredPreview, error) {
	if weapon.Kind != "beam" {
		return ConfiguredPreview{}, fmt.Errorf("configured Beam preview requires Beam weapon, got %q", weapon.Kind)
	}

	mountType := mount.MountType
	if mountType == "" {
		mountType = core.ShipWeaponMountStandard
	}

	damagePercent := 100
	rangeProfile := RangeProfileStandard
	staticToHit := 0
	switch mountType {
	case core.ShipWeaponMountStandard:
	case core.ShipWeaponMountHeavy:
		damagePercent = 150
		rangeProfile = RangeProfileHeavy
	case core.ShipWeaponMountPointDefense:
		damagePercent = 50
		rangeProfile = RangeProfilePointDefense
		staticToHit += 25
	default:
		return ConfiguredPreview{}, fmt.Errorf("unsupported Beam mount type %q", mountType)
	}

	effects := make([]string, 0, len(mount.ModifierIDs)+len(weapon.InnateEffectIDs)+2)
	appendEffect := func(id string) {
		if id == "" {
			return
		}
		for _, existing := range effects {
			if existing == id {
				return
			}
		}
		effects = append(effects, id)
	}

	autofire := false
	continuous := false
	for _, modifierID := range mount.ModifierIDs {
		id := string(modifierID)
		appendEffect(id)
		switch modifierID {
		case core.ShipWeaponModifierAutofire:
			autofire = true
		case core.ShipWeaponModifierContinuous:
			continuous = true
		}
	}

	for _, innateID := range weapon.InnateEffectIDs {
		switch innateID {
		case "natural_shield_piercing":
			appendEffect("shield_piercing")
		case "armor_structure_bypass":
			appendEffect("armor_first_bypass")
		case "monster_antaran_immunity_exception":
			appendEffect("monster_antaran_quarter_damage")
		case "all_four_shields":
			appendEffect("enveloping")
		case "double_range_penalty", "double_range_dissipation":
			appendEffect("doubled_range_to_hit")
		default:
			appendEffect(innateID)
		}
	}

	// Two executable-native effects are not fully represented by the older
	// metadata rows but are frozen by Slice 17.6 Gate 2.
	switch weapon.ID {
	case "fusion_beam":
		appendEffect("doubled_range_to_hit")
	case "death_ray":
		appendEffect("continuous")
		continuous = true
	}

	shots := 1
	if autofire {
		shots = 3
		staticToHit -= 20
	}
	if continuous {
		staticToHit += 25
	}

	minDamage, maxDamage := ScaleDamageBounds(weapon.MinDamage, weapon.MaxDamage, damagePercent)
	return ConfiguredPreview{
		Family:                "beam",
		BaseMinDamage:         weapon.MinDamage,
		BaseMaxDamage:         weapon.MaxDamage,
		ConfiguredMinDamage:   minDamage,
		ConfiguredMaxDamage:   maxDamage,
		ShotsPerWeapon:        shots,
		RangeProfile:          rangeProfile,
		StaticToHitAdjustment: staticToHit,
		EffectIDs:             effects,
	}, nil
}
