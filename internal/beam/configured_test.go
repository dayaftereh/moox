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

func TestEffectiveDamageReferenceVectorsABCD(t *testing.T) {
	tests := []struct {
		name          string
		weaponID      string
		baseMin       int
		baseMax       int
		mount         core.ShipWeaponMountType
		innate        []string
		rangeModifier int
		effectiveRoll int
		threshold     int
		wantPercent   int
		wantMin       int
		wantMax       int
		wantDamage    int
	}{
		{
			name: "A Laser Standard rounding", weaponID: "laser_cannon",
			baseMin: 1, baseMax: 4, mount: core.ShipWeaponMountStandard,
			rangeModifier: -10, effectiveRoll: 100, threshold: 50,
			wantPercent: 90, wantMin: 1, wantMax: 4, wantDamage: 4,
		},
		{
			name: "B Particle Heavy", weaponID: "particle_beam",
			baseMin: 10, baseMax: 30, mount: core.ShipWeaponMountHeavy,
			innate:        []string{"natural_shield_piercing"},
			rangeModifier: -10, effectiveRoll: 80, threshold: 50,
			wantPercent: 140, wantMin: 14, wantMax: 42, wantDamage: 31,
		},
		{
			name: "C Mauler Heavy", weaponID: "mauler_device",
			baseMin: 100, baseMax: 100, mount: core.ShipWeaponMountHeavy,
			innate:        []string{"always_hit", "double_range_dissipation"},
			rangeModifier: -10, effectiveRoll: 100, threshold: 60,
			wantPercent: 140, wantMin: 140, wantMax: 140, wantDamage: 140,
		},
		{
			name: "D Laser Point Defense ship target", weaponID: "laser_cannon",
			baseMin: 1, baseMax: 4, mount: core.ShipWeaponMountPointDefense,
			rangeModifier: -30, effectiveRoll: 50, threshold: 45,
			wantPercent: 20, wantMin: 1, wantMax: 1, wantDamage: 1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config, err := DeriveConfigured(tc.weaponID, tc.baseMin, tc.baseMax, tc.mount, nil, tc.innate)
			if err != nil {
				t.Fatal(err)
			}
			minDamage, maxDamage, percent := EffectiveDamageBounds(config, tc.rangeModifier, 100, 0)
			if percent != tc.wantPercent || minDamage != tc.wantMin || maxDamage != tc.wantMax {
				t.Fatalf("bounds=%d-%d percent=%d want %d-%d percent=%d", minDamage, maxDamage, percent, tc.wantMin, tc.wantMax, tc.wantPercent)
			}
			if got := InterpolateDamage(minDamage, maxDamage, tc.effectiveRoll, tc.threshold, 100); got != tc.wantDamage {
				t.Fatalf("damage=%d want %d", got, tc.wantDamage)
			}
		})
	}
}

func TestEffectiveDamageNoRangeDissipationSuppressesRangePenalty(t *testing.T) {
	config, err := DeriveConfigured("mass_driver", 6, 6, core.ShipWeaponMountStandard, nil, []string{"no_range_dissipation"})
	if err != nil {
		t.Fatal(err)
	}
	minDamage, maxDamage, percent := EffectiveDamageBounds(config, -65, 100, 0)
	if percent != 100 || minDamage != 6 || maxDamage != 6 {
		t.Fatalf("no-range-dissipation bounds=%d-%d percent=%d", minDamage, maxDamage, percent)
	}
}

func TestDeriveConfiguredSelectedContinuousAndEnveloping(t *testing.T) {
	config, err := DeriveConfigured(
		"fusion_beam",
		2,
		6,
		core.ShipWeaponMountStandard,
		[]core.ShipWeaponModifierID{
			core.ShipWeaponModifierContinuous,
			core.ShipWeaponModifierEnveloping,
		},
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if config.StaticToHitAdjustment != 25 || config.ShotsPerWeapon != 1 {
		t.Fatalf("Fusion selected modifier config=%+v", config)
	}
	if !reflect.DeepEqual(config.EffectIDs, []string{"continuous", "enveloping", "doubled_range_to_hit"}) {
		t.Fatalf("Fusion selected modifier effects=%v", config.EffectIDs)
	}
}
