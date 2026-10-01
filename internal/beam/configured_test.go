package beam

import (
	"reflect"
	"testing"

	"moox/internal/core"
	"moox/internal/ruleset"
)

func TestScaleDamageBoundsUsesOriginalNearestIntegerRounding(t *testing.T) {
	minDamage, maxDamage := ScaleDamageBounds(1, 4, 90)
	if minDamage != 1 || maxDamage != 4 {
		t.Fatalf("Laser 1-4 at 90%% = %d-%d want 1-4", minDamage, maxDamage)
	}

	minDamage, maxDamage = ScaleDamageBounds(1, 4, 20)
	if minDamage != 1 || maxDamage != 1 {
		t.Fatalf("Laser 1-4 at 20%% = %d-%d want clamped 1-1", minDamage, maxDamage)
	}
}

func TestDeriveConfiguredPreviewPointDefenseAutofire(t *testing.T) {
	preview, err := DeriveConfiguredPreview(ruleset.TacticalWeaponRule{
		ID: "phasor", Kind: "beam", MinDamage: 5, MaxDamage: 20,
	}, core.ShipWeaponMount{
		MountType: core.ShipWeaponMountPointDefense,
		ModifierIDs: []core.ShipWeaponModifierID{
			core.ShipWeaponModifierShieldPiercing,
			core.ShipWeaponModifierAutofire,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if preview.Family != "beam" || preview.ConfiguredMinDamage != 3 || preview.ConfiguredMaxDamage != 10 ||
		preview.ShotsPerWeapon != 3 || preview.RangeProfile != RangeProfilePointDefense || preview.StaticToHitAdjustment != 5 {
		t.Fatalf("configured PD Autofire preview=%+v", preview)
	}
	if !reflect.DeepEqual(preview.EffectIDs, []string{"shield_piercing", "autofire"}) {
		t.Fatalf("effect IDs=%v", preview.EffectIDs)
	}
}

func TestDeriveConfiguredPreviewNormalizesFrozenInnateEffects(t *testing.T) {
	particle, err := DeriveConfiguredPreview(ruleset.TacticalWeaponRule{
		ID: "particle_beam", Kind: "beam", MinDamage: 10, MaxDamage: 30,
		InnateEffectIDs: []string{"natural_shield_piercing"},
	}, core.ShipWeaponMount{MountType: core.ShipWeaponMountHeavy})
	if err != nil {
		t.Fatal(err)
	}
	if particle.ConfiguredMinDamage != 15 || particle.ConfiguredMaxDamage != 45 || particle.RangeProfile != RangeProfileHeavy {
		t.Fatalf("Particle Heavy preview=%+v", particle)
	}
	if !reflect.DeepEqual(particle.EffectIDs, []string{"shield_piercing"}) {
		t.Fatalf("Particle effects=%v", particle.EffectIDs)
	}

	fusion, err := DeriveConfiguredPreview(ruleset.TacticalWeaponRule{
		ID: "fusion_beam", Kind: "beam", MinDamage: 2, MaxDamage: 6,
	}, core.ShipWeaponMount{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(fusion.EffectIDs, []string{"doubled_range_to_hit"}) {
		t.Fatalf("Fusion effects=%v", fusion.EffectIDs)
	}

	deathRay, err := DeriveConfiguredPreview(ruleset.TacticalWeaponRule{
		ID: "death_ray", Kind: "beam", MinDamage: 50, MaxDamage: 100,
		InnateEffectIDs: []string{"marine_kill"},
	}, core.ShipWeaponMount{})
	if err != nil {
		t.Fatal(err)
	}
	if deathRay.StaticToHitAdjustment != 25 || !reflect.DeepEqual(deathRay.EffectIDs, []string{"marine_kill", "continuous"}) {
		t.Fatalf("Death Ray preview=%+v", deathRay)
	}
}
