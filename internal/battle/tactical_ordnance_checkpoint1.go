package battle

import (
	"fmt"
	"sort"

	"moox/internal/core"
	"moox/internal/protocol"
)

type TacticalOrdnanceRulesSnapshot struct {
	MovementUnitsPerField     int                                  `json:"movement_units_per_field"`
	FastSpeedBonus            int                                  `json:"fast_speed_bonus"`
	TorpedoCooldownRounds     int                                  `json:"torpedo_cooldown_rounds"`
	MIRVMultiplier            int                                  `json:"mirv_multiplier"`
	PlasmaDissipationPerField int                                  `json:"plasma_dissipation_per_field"`
	Weapons                   []TacticalOrdnanceWeaponRuleSnapshot `json:"weapons"`
}

type TacticalOrdnanceWeaponRuleSnapshot struct {
	WeaponID           string `json:"weapon_id"`
	BaseSpeed          int    `json:"base_speed"`
	BaseDurability     int    `json:"base_durability"`
	BeamInterceptorDCV int    `json:"beam_interceptor_dcv"`
}

type TacticalOrdnanceWeaponState struct {
	ShipID         core.ID `json:"ship_id"`
	WeaponSlot     int     `json:"weapon_slot"`
	AmmoRemaining  int     `json:"ammo_remaining"`
	CooldownRounds int     `json:"cooldown_rounds"`
}

type TacticalProjectileState struct {
	ProjectileID             uint64                      `json:"projectile_id"`
	SpawnOrder               uint64                      `json:"spawn_order,omitempty"`
	SourceSeatID             protocol.SeatID             `json:"source_seat_id"`
	SourceShipID             core.ID                     `json:"source_ship_id"`
	WeaponSlot               int                         `json:"weapon_slot"`
	TargetShipID             core.ID                     `json:"target_ship_id"`
	WeaponID                 string                      `json:"weapon_id"`
	WeaponFamily             string                      `json:"weapon_family"`
	ModifierIDs              []core.ShipWeaponModifierID `json:"modifier_ids,omitempty"`
	X                        int                         `json:"x"`
	Y                        int                         `json:"y"`
	Facing                   int                         `json:"facing"`
	Speed                    int                         `json:"speed"`
	MovementRemaining        int                         `json:"movement_remaining"`
	DistanceTraveled         int                         `json:"distance_traveled"`
	Multiplicity             int                         `json:"multiplicity"`
	UnitDurability           int                         `json:"unit_durability"`
	InterceptDamageRemainder int                         `json:"intercept_damage_remainder"`
	WarheadStrength          int                         `json:"warhead_strength"`
}

type TacticalOrdnanceState struct {
	NextProjectileID uint64                        `json:"next_projectile_id"`
	WeaponStates     []TacticalOrdnanceWeaponState `json:"weapon_states,omitempty"`
	Projectiles      []TacticalProjectileState     `json:"projectiles,omitempty"`
}

func expectedTacticalOrdnanceWeaponRules() []TacticalOrdnanceWeaponRuleSnapshot {
	return []TacticalOrdnanceWeaponRuleSnapshot{
		{WeaponID: "nuclear_missile", BaseSpeed: 12, BaseDurability: 4, BeamInterceptorDCV: 50},
		{WeaponID: "merculite_missile", BaseSpeed: 12, BaseDurability: 8, BeamInterceptorDCV: 75},
		{WeaponID: "pulson_missile", BaseSpeed: 12, BaseDurability: 12, BeamInterceptorDCV: 100},
		{WeaponID: "zeon_missile", BaseSpeed: 12, BaseDurability: 16, BeamInterceptorDCV: 130},
		{WeaponID: "anti_matter_torpedoes", BaseSpeed: 20, BaseDurability: 5, BeamInterceptorDCV: 100},
		{WeaponID: "proton_torpedoes", BaseSpeed: 20, BaseDurability: 5, BeamInterceptorDCV: 100},
		{WeaponID: "plasma_torpedoes", BaseSpeed: 24, BaseDurability: 5, BeamInterceptorDCV: 120},
	}
}

func validateTacticalOrdnanceRulesSnapshot(r TacticalOrdnanceRulesSnapshot) error {
	if r.MovementUnitsPerField != 20 || r.FastSpeedBonus != 4 || r.TorpedoCooldownRounds != 2 || r.MIRVMultiplier != 4 || r.PlasmaDissipationPerField != 5 {
		return fmt.Errorf("unsupported Slice 17.7 tactical ordnance constants")
	}
	want := expectedTacticalOrdnanceWeaponRules()
	if len(r.Weapons) != len(want) {
		return fmt.Errorf("tactical ordnance rule count %d, expected %d", len(r.Weapons), len(want))
	}
	for i := range want {
		if r.Weapons[i] != want[i] {
			return fmt.Errorf("tactical ordnance rule[%d]=%+v, expected %+v", i, r.Weapons[i], want[i])
		}
	}
	return nil
}

func tacticalSpecHasExecutableOrdnance(spec TacticalSpec) bool {
	for _, ship := range spec.Ships {
		for _, weapon := range ship.Weapons {
			if weapon.TacticalSupport == "ordnance" {
				return true
			}
		}
	}
	return false
}

func validateTacticalOrdnanceRulesForSpec(spec TacticalSpec) error {
	if spec.Rules.Ordnance == nil {
		if tacticalSpecHasExecutableOrdnance(spec) {
			return fmt.Errorf("executable tactical ordnance requires frozen ordnance rules")
		}
		return nil
	}
	return validateTacticalOrdnanceRulesSnapshot(*spec.Rules.Ordnance)
}

func cloneTacticalOrdnanceRulesSnapshot(in *TacticalOrdnanceRulesSnapshot) *TacticalOrdnanceRulesSnapshot {
	if in == nil {
		return nil
	}
	out := *in
	out.Weapons = append([]TacticalOrdnanceWeaponRuleSnapshot(nil), in.Weapons...)
	return &out
}

func initialTacticalOrdnanceState(spec TacticalSpec) TacticalOrdnanceState {
	out := TacticalOrdnanceState{NextProjectileID: 1}
	for _, ship := range spec.Ships {
		for _, weapon := range ship.Weapons {
			family := tacticalWeaponFamily(weapon)
			if family != "missile" && family != "torpedo" {
				continue
			}
			state := TacticalOrdnanceWeaponState{ShipID: ship.ShipID, WeaponSlot: weapon.Slot}
			if family == "missile" {
				state.AmmoRemaining = weapon.Ammo
			}
			out.WeaponStates = append(out.WeaponStates, state)
		}
	}
	sort.Slice(out.WeaponStates, func(i, j int) bool {
		if out.WeaponStates[i].ShipID != out.WeaponStates[j].ShipID {
			return out.WeaponStates[i].ShipID < out.WeaponStates[j].ShipID
		}
		return out.WeaponStates[i].WeaponSlot < out.WeaponStates[j].WeaponSlot
	})
	return out
}

func cloneTacticalOrdnanceState(in TacticalOrdnanceState) TacticalOrdnanceState {
	out := in
	out.WeaponStates = append([]TacticalOrdnanceWeaponState(nil), in.WeaponStates...)
	out.Projectiles = make([]TacticalProjectileState, len(in.Projectiles))
	for i := range in.Projectiles {
		out.Projectiles[i] = in.Projectiles[i]
		out.Projectiles[i].ModifierIDs = append([]core.ShipWeaponModifierID(nil), in.Projectiles[i].ModifierIDs...)
	}
	return out
}

func ordnanceWeaponSpec(spec TacticalSpec, shipID core.ID, slot int) (*TacticalShipSpec, *TacticalWeaponSpec) {
	ship := tacticalShipSpec(spec, shipID)
	if ship == nil {
		return nil, nil
	}
	weapon := weaponSpec(*ship, slot)
	if weapon == nil {
		return ship, nil
	}
	return ship, weapon
}

func validateTacticalOrdnanceState(spec TacticalSpec, state TacticalOrdnanceState) error {
	if state.NextProjectileID == 0 {
		return fmt.Errorf("tactical ordnance next projectile id must be non-zero")
	}
	want := initialTacticalOrdnanceState(spec).WeaponStates
	if len(state.WeaponStates) != len(want) {
		return fmt.Errorf("tactical ordnance weapon-state count %d, expected %d", len(state.WeaponStates), len(want))
	}
	for i := range want {
		got := state.WeaponStates[i]
		if got.ShipID != want[i].ShipID || got.WeaponSlot != want[i].WeaponSlot {
			return fmt.Errorf("tactical ordnance weapon-state[%d] identity mismatch", i)
		}
		_, weapon := ordnanceWeaponSpec(spec, got.ShipID, got.WeaponSlot)
		if weapon == nil {
			return fmt.Errorf("tactical ordnance weapon-state[%d] references missing weapon", i)
		}
		switch tacticalWeaponFamily(*weapon) {
		case "missile":
			if got.AmmoRemaining < 0 || got.AmmoRemaining > weapon.Ammo || got.CooldownRounds != 0 {
				return fmt.Errorf("tactical Missile state for ship %d slot %d is outside bounds", got.ShipID, got.WeaponSlot)
			}
		case "torpedo":
			if got.AmmoRemaining != 0 || got.CooldownRounds < 0 || got.CooldownRounds > 2 {
				return fmt.Errorf("tactical Torpedo state for ship %d slot %d is outside bounds", got.ShipID, got.WeaponSlot)
			}
		default:
			return fmt.Errorf("tactical ordnance state references non-ordnance weapon %q", weapon.WeaponID)
		}
	}

	var previous uint64
	for i, projectile := range state.Projectiles {
		if projectile.ProjectileID == 0 || projectile.ProjectileID <= previous || projectile.ProjectileID >= state.NextProjectileID {
			return fmt.Errorf("tactical projectile[%d] id/order invalid", i)
		}
		previous = projectile.ProjectileID
		source, weapon := ordnanceWeaponSpec(spec, projectile.SourceShipID, projectile.WeaponSlot)
		if source == nil || weapon == nil {
			return fmt.Errorf("tactical projectile %d references missing source weapon", projectile.ProjectileID)
		}
		if projectile.SourceSeatID != source.SeatID || projectile.WeaponID != weapon.WeaponID || projectile.WeaponFamily != tacticalWeaponFamily(*weapon) {
			return fmt.Errorf("tactical projectile %d source identity mismatch", projectile.ProjectileID)
		}
		target := tacticalShipSpec(spec, projectile.TargetShipID)
		if target == nil || target.SeatID == source.SeatID {
			return fmt.Errorf("tactical projectile %d has invalid target ship %d", projectile.ProjectileID, projectile.TargetShipID)
		}
		if projectile.Facing < 0 || projectile.Facing > 15 || absInt(projectile.X) > TacticalCoordinateSafetyLimit || absInt(projectile.Y) > TacticalCoordinateSafetyLimit {
			return fmt.Errorf("tactical projectile %d position/facing invalid", projectile.ProjectileID)
		}
		if projectile.Speed <= 0 || projectile.MovementRemaining < 0 || projectile.DistanceTraveled < 0 || projectile.Multiplicity <= 0 || projectile.UnitDurability <= 0 || projectile.InterceptDamageRemainder < 0 || projectile.InterceptDamageRemainder >= projectile.UnitDurability || projectile.WarheadStrength <= 0 {
			return fmt.Errorf("tactical projectile %d runtime values invalid", projectile.ProjectileID)
		}
	}
	return nil
}
