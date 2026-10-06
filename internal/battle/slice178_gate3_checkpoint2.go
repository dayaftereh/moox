package battle

import (
	"fmt"

	"moox/internal/core"
)

const (
	TacticalInternalPoolEngine          = "engine"
	TacticalInternalPoolSecondary       = "secondary"
	TacticalInternalPoolWeapon          = "weapon"
	TacticalInternalPoolSpecial         = "special"
	TacticalInternalPoolShieldGenerator = "shield_generator"
)

type TacticalInternalWeaponState struct {
	Slot    int `json:"slot"`
	Current int `json:"current"`
}

type TacticalInternalSpecialState struct {
	Slot    int `json:"slot"`
	Current int `json:"current"`
}

type TacticalInternalState struct {
	EngineCurrent          int                            `json:"engine_current"`
	SecondaryCurrent       int                            `json:"secondary_current"`
	Weapons                []TacticalInternalWeaponState  `json:"weapons,omitempty"`
	Specials               []TacticalInternalSpecialState `json:"specials,omitempty"`
	ShieldGeneratorCurrent int                            `json:"shield_generator_current"`
}

type TacticalInternalWeaponView struct {
	Slot     int `json:"slot"`
	Current  int `json:"current"`
	Max      int `json:"max"`
	UnitSize int `json:"unit_size"`
}

type TacticalInternalSpecialView struct {
	Slot    int `json:"slot"`
	Current int `json:"current"`
	Max     int `json:"max"`
}

type TacticalInternalView struct {
	EngineCurrent          int                           `json:"engine_current"`
	EngineMax              int                           `json:"engine_max"`
	SecondaryCurrent       int                           `json:"secondary_current"`
	SecondaryMax           int                           `json:"secondary_max"`
	Weapons                []TacticalInternalWeaponView  `json:"weapons,omitempty"`
	Specials               []TacticalInternalSpecialView `json:"specials,omitempty"`
	ShieldGeneratorCurrent int                           `json:"shield_generator_current"`
	ShieldGeneratorMax     int                           `json:"shield_generator_max"`
}

func initialTacticalInternalState(spec *TacticalInternalSpec) TacticalInternalState {
	if spec == nil {
		return TacticalInternalState{}
	}
	out := TacticalInternalState{
		EngineCurrent:          spec.EngineMax,
		SecondaryCurrent:       spec.SecondaryMax,
		ShieldGeneratorCurrent: spec.ShieldGeneratorMax,
		Weapons:                make([]TacticalInternalWeaponState, len(spec.Weapons)),
		Specials:               make([]TacticalInternalSpecialState, len(spec.Specials)),
	}
	for i, item := range spec.Weapons {
		out.Weapons[i] = TacticalInternalWeaponState{Slot: item.Slot, Current: item.Max}
	}
	for i, item := range spec.Specials {
		out.Specials[i] = TacticalInternalSpecialState{Slot: item.Slot, Current: item.Max}
	}
	return out
}

func cloneTacticalInternalState(in TacticalInternalState) TacticalInternalState {
	out := in
	out.Weapons = append([]TacticalInternalWeaponState(nil), in.Weapons...)
	out.Specials = append([]TacticalInternalSpecialState(nil), in.Specials...)
	return out
}

func tacticalNonDamageableSpecial(id string) bool {
	switch id {
	case "battle_pods", "extended_fuel_tanks", "heavy_armor", "quantum_detonator", "reinforced_hull":
		return true
	default:
		return false
	}
}

func TacticalSpecialFunctional(shipSpec TacticalShipSpec, shipState TacticalShipState, legacy17_8Disabled bool, specialID string) bool {
	if legacy17_8Disabled || shipState.Destroyed {
		return false
	}
	slot := -1
	for _, special := range shipSpec.Specials {
		if special.SpecialID == specialID {
			slot = special.Slot
			break
		}
	}
	if slot < 0 {
		return false
	}
	if tacticalNonDamageableSpecial(specialID) {
		return true
	}
	if shipSpec.Internal == nil {
		return false
	}
	maxFound := false
	for _, special := range shipSpec.Internal.Specials {
		if special.Slot == slot {
			maxFound = special.Max > 0
			break
		}
	}
	if !maxFound {
		return false
	}
	for _, special := range shipState.Internal.Specials {
		if special.Slot == slot {
			return special.Current > 0
		}
	}
	return false
}

func tacticalCurrentWeaponCount(shipSpec TacticalShipSpec, shipState TacticalShipState, slot int) int {
	immutableCount := 0
	for _, weapon := range shipSpec.Weapons {
		if weapon.Slot == slot {
			immutableCount = weapon.Count
			break
		}
	}
	for _, state := range shipState.Weapons {
		if state.Slot == slot && state.CountCurrent >= 0 {
			if state.CountCurrent < immutableCount {
				return state.CountCurrent
			}
			return immutableCount
		}
	}
	return immutableCount
}

func tacticalCurrentCombatSpeed(spec TacticalSpec, shipSpec TacticalShipSpec, shipState TacticalShipState, legacy17_8Disabled bool) int {
	if legacy17_8Disabled || shipSpec.Internal == nil || shipSpec.Internal.EngineMax <= 0 {
		if shipState.CurrentCombatSpeed >= 0 {
			return shipState.CurrentCombatSpeed
		}
		return shipSpec.CurrentCombatSpeed
	}
	working := shipSpec.CurrentCombatSpeed
	if hasTacticalSpecial(shipSpec, "augmented_engines") && !TacticalSpecialFunctional(shipSpec, shipState, false, "augmented_engines") {
		if spec.Rules.Specials != nil {
			working -= spec.Rules.Specials.AugmentedEnginesSpeedBonus
		}
	}
	if working < 0 {
		working = 0
	}
	maxEngine := shipSpec.Internal.EngineMax
	currentEngine := shipState.Internal.EngineCurrent
	if currentEngine < 0 {
		currentEngine = 0
	}
	if currentEngine > maxEngine {
		currentEngine = maxEngine
	}
	half := maxEngine / 2
	denominator := maxEngine - half
	if denominator <= 0 || currentEngine <= half {
		return 0
	}
	speed := ((currentEngine - half) * working) / denominator
	if speed < 0 {
		return 0
	}
	return speed
}

func tacticalBeamOffenseForState(spec TacticalSpec, shipSpec TacticalShipSpec, shipState TacticalShipState, legacy17_8Disabled bool) int {
	offense := shipSpec.BeamOffense
	if hasTacticalSpecial(shipSpec, "battle_scanner") && !TacticalSpecialFunctional(shipSpec, shipState, legacy17_8Disabled, "battle_scanner") {
		if spec.Rules.Specials != nil {
			offense -= spec.Rules.Specials.BattleScannerOffenseBonus
		}
	}
	if offense < 0 {
		return 0
	}
	return offense
}

func tacticalBeamDefenseForState(spec TacticalSpec, shipSpec TacticalShipSpec, shipState TacticalShipState, legacy17_8Disabled bool) int {
	speed := tacticalCurrentCombatSpeed(spec, shipSpec, shipState, legacy17_8Disabled)
	if shipSpec.CurrentCombatSpeed <= 0 {
		return 0
	}
	defense := (shipSpec.BeamDefense * speed) / shipSpec.CurrentCombatSpeed
	if !legacy17_8Disabled && shipState.CloakState == TacticalCloakStateCloaked && spec.Rules.Specials != nil &&
		(TacticalSpecialFunctional(shipSpec, shipState, false, "cloaking_device") || TacticalSpecialFunctional(shipSpec, shipState, false, "phasing_cloak")) {
		defense += spec.Rules.Specials.CloakBeamDefenseBonus
	}
	return defense
}

func tacticalShieldArcMaxForState(spec TacticalSpec, shipSpec TacticalShipSpec, shipState TacticalShipState, legacy17_8Disabled bool) int {
	if shipState.Internal.ShieldGeneratorCurrent == 0 && shipSpec.Internal != nil && shipSpec.Internal.ShieldGeneratorMax > 0 && !legacy17_8Disabled {
		return 0
	}
	hull, ok := tacticalHullValidationRuleFor(shipSpec.HullID)
	if !ok {
		return shipSpec.ShieldArcMax
	}
	base := shipSpec.ShieldStrength * 5 * hull.ShieldFactor
	if legacy17_8Disabled {
		return shipSpec.ShieldArcMax
	}
	if hasTacticalSpecial(shipSpec, "multi_phased_shields") && TacticalSpecialFunctional(shipSpec, shipState, false, "multi_phased_shields") {
		if spec.Rules.Specials != nil && spec.Rules.Specials.MultiPhasedShieldDenominator > 0 {
			return (base * spec.Rules.Specials.MultiPhasedShieldNumerator) / spec.Rules.Specials.MultiPhasedShieldDenominator
		}
	}
	return base
}

func tacticalMissileDefenseForState(spec TacticalSpec, runtime *tacticalRuntime, targetSpec TacticalShipSpec, targetState TacticalShipState) int {
	defense := targetSpec.MissileDefense
	if runtime == nil || runtime.legacy17_8Disabled || spec.Rules.Specials == nil {
		return defense
	}
	rules := spec.Rules.Specials
	self := 0
	if TacticalSpecialFunctional(targetSpec, targetState, false, "ecm_jammer") {
		self = rules.ECMJammerDefenseBonus
	}
	if TacticalSpecialFunctional(targetSpec, targetState, false, "multi_wave_ecm_jammer") && self < rules.MultiWaveECMJammerDefenseBonus {
		self = rules.MultiWaveECMJammerDefenseBonus
	}
	if TacticalSpecialFunctional(targetSpec, targetState, false, "wide_area_jammer") && self < rules.WideAreaJammerSelfDefenseBonus {
		self = rules.WideAreaJammerSelfDefenseBonus
	}
	fleet := 0
	for _, friendlySpec := range spec.Ships {
		if friendlySpec.SeatID != targetSpec.SeatID || friendlySpec.ShipID == targetSpec.ShipID {
			continue
		}
		friendlyState := runtime.shipState(friendlySpec.ShipID)
		if friendlyState != nil && TacticalSpecialFunctional(friendlySpec, *friendlyState, false, "wide_area_jammer") {
			fleet = rules.WideAreaJammerFleetDefenseBonus
			break
		}
	}
	jammer := self
	if fleet > jammer {
		jammer = fleet
	}
	return defense + jammer
}

func hasTacticalSpecial(shipSpec TacticalShipSpec, id string) bool {
	for _, special := range shipSpec.Specials {
		if special.SpecialID == id {
			return true
		}
	}
	return false
}

func reconcileTacticalShipDerivedState(spec TacticalSpec, runtime *tacticalRuntime, shipID core.ID) {
	if runtime == nil {
		return
	}
	shipSpec := tacticalShipSpec(spec, shipID)
	shipState := runtime.shipState(shipID)
	if shipSpec == nil || shipState == nil {
		return
	}
	after := tacticalCurrentCombatSpeed(spec, *shipSpec, *shipState, runtime.legacy17_8Disabled)
	shipState.CurrentCombatSpeed = after
	shipState.MovementMax = after
	if shipState.MovementCurrent > after {
		shipState.MovementCurrent = after
	}
	shieldMax := tacticalShieldArcMaxForState(spec, *shipSpec, *shipState, runtime.legacy17_8Disabled)
	for i := range shipState.ShieldArcCurrent {
		if shipState.ShieldArcCurrent[i] > shieldMax {
			shipState.ShieldArcCurrent[i] = shieldMax
		}
	}
}

func applyTacticalInternalDamage(spec TacticalSpec, runtime *tacticalRuntime, shipID core.ID, pool string, slot int, amount int, source string) error {
	if runtime == nil {
		return fmt.Errorf("tactical internal damage requires runtime")
	}
	if amount <= 0 {
		return fmt.Errorf("tactical internal damage amount must be positive")
	}
	shipSpec := tacticalShipSpec(spec, shipID)
	shipState := runtime.shipState(shipID)
	if shipSpec == nil || shipState == nil {
		return fmt.Errorf("tactical internal damage ship %d is missing", shipID)
	}
	if runtime.legacy17_8Disabled || shipSpec.Internal == nil {
		return fmt.Errorf("tactical internal damage is unavailable for legacy-disabled battle")
	}
	before, after := 0, 0
	poolMax := 0
	found := false
	switch pool {
	case TacticalInternalPoolEngine:
		before = shipState.Internal.EngineCurrent
		poolMax = shipSpec.Internal.EngineMax
		after = maxInt(0, before-amount)
		shipState.Internal.EngineCurrent = after
		found = true
	case TacticalInternalPoolSecondary:
		before = shipState.Internal.SecondaryCurrent
		poolMax = shipSpec.Internal.SecondaryMax
		after = maxInt(0, before-amount)
		shipState.Internal.SecondaryCurrent = after
		found = true
	case TacticalInternalPoolWeapon:
		for i := range shipState.Internal.Weapons {
			if shipState.Internal.Weapons[i].Slot != slot {
				continue
			}
			before = shipState.Internal.Weapons[i].Current
			after = maxInt(0, before-amount)
			shipState.Internal.Weapons[i].Current = after
			found = true
			unitSize := 1
			for _, immutable := range shipSpec.Internal.Weapons {
				if immutable.Slot == slot {
					unitSize = immutable.UnitSize
					poolMax = immutable.Max
					break
				}
			}
			liveCount := 0
			if after > 0 {
				liveCount = (after + unitSize - 1) / unitSize
			}
			for j := range shipState.Weapons {
				if shipState.Weapons[j].Slot == slot {
					shipState.Weapons[j].CountCurrent = liveCount
					if liveCount == 0 {
						shipState.Weapons[j].Ready = false
					}
					break
				}
			}
			break
		}
	case TacticalInternalPoolSpecial:
		for i := range shipState.Internal.Specials {
			if shipState.Internal.Specials[i].Slot != slot {
				continue
			}
			before = shipState.Internal.Specials[i].Current
			after = maxInt(0, before-amount)
			shipState.Internal.Specials[i].Current = after
			for _, immutable := range shipSpec.Internal.Specials {
				if immutable.Slot == slot {
					poolMax = immutable.Max
					break
				}
			}
			found = true
			break
		}
	case TacticalInternalPoolShieldGenerator:
		before = shipState.Internal.ShieldGeneratorCurrent
		poolMax = shipSpec.Internal.ShieldGeneratorMax
		after = maxInt(0, before-amount)
		shipState.Internal.ShieldGeneratorCurrent = after
		found = true
		if after == 0 {
			shipState.ShieldArcCurrent = [4]int{}
		}
	default:
		return fmt.Errorf("unknown tactical internal pool %q", pool)
	}
	if !found {
		return fmt.Errorf("tactical internal pool %q slot %d is not present on ship %d", pool, slot, shipID)
	}
	speedBefore := shipState.CurrentCombatSpeed
	reconcileTacticalShipDerivedState(spec, runtime, shipID)
	speedAfter := shipState.CurrentCombatSpeed
	event := map[string]any{
		"ship_id": shipID, "pool": pool, "slot": slot, "source": source,
		"damage": amount, "damage_consumed": before - after, "current_before": before, "current_after": after, "max": poolMax,
		"combat_speed_before": speedBefore, "combat_speed_after": speedAfter,
	}
	switch pool {
	case TacticalInternalPoolWeapon:
		event["weapon_count_current"] = tacticalCurrentWeaponCount(*shipSpec, *shipState, slot)
	case TacticalInternalPoolSpecial:
		for _, special := range shipSpec.Specials {
			if special.Slot == slot {
				event["special_id"] = special.SpecialID
				event["special_functional"] = TacticalSpecialFunctional(*shipSpec, *shipState, false, special.SpecialID)
				break
			}
		}
	case TacticalInternalPoolShieldGenerator:
		event["shield_generator_destroyed"] = after == 0
		event["shield_arc_current"] = shipState.ShieldArcCurrent
	}
	return runtime.appendEvent("internal_damage_applied", 0, 0, event)
}

func tacticalInternalView(spec *TacticalInternalSpec, state TacticalInternalState) TacticalInternalView {
	if spec == nil {
		return TacticalInternalView{}
	}
	out := TacticalInternalView{
		EngineCurrent: state.EngineCurrent, EngineMax: spec.EngineMax,
		SecondaryCurrent: state.SecondaryCurrent, SecondaryMax: spec.SecondaryMax,
		ShieldGeneratorCurrent: state.ShieldGeneratorCurrent, ShieldGeneratorMax: spec.ShieldGeneratorMax,
		Weapons:  make([]TacticalInternalWeaponView, 0, len(spec.Weapons)),
		Specials: make([]TacticalInternalSpecialView, 0, len(spec.Specials)),
	}
	for _, immutable := range spec.Weapons {
		current := 0
		for _, live := range state.Weapons {
			if live.Slot == immutable.Slot {
				current = live.Current
				break
			}
		}
		out.Weapons = append(out.Weapons, TacticalInternalWeaponView{Slot: immutable.Slot, Current: current, Max: immutable.Max, UnitSize: immutable.UnitSize})
	}
	for _, immutable := range spec.Specials {
		current := 0
		for _, live := range state.Specials {
			if live.Slot == immutable.Slot {
				current = live.Current
				break
			}
		}
		out.Specials = append(out.Specials, TacticalInternalSpecialView{Slot: immutable.Slot, Current: current, Max: immutable.Max})
	}
	return out
}

func validateTacticalInternalState(spec *TacticalInternalSpec, state TacticalInternalState) error {
	if spec == nil {
		if state.EngineCurrent != 0 || state.SecondaryCurrent != 0 || state.ShieldGeneratorCurrent != 0 || len(state.Weapons) != 0 || len(state.Specials) != 0 {
			return fmt.Errorf("internal state exists without immutable spec")
		}
		return nil
	}
	if state.EngineCurrent < 0 || state.EngineCurrent > spec.EngineMax || state.SecondaryCurrent < 0 || state.SecondaryCurrent > spec.SecondaryMax || state.ShieldGeneratorCurrent < 0 || state.ShieldGeneratorCurrent > spec.ShieldGeneratorMax {
		return fmt.Errorf("fixed internal pool outside immutable bounds")
	}
	if len(state.Weapons) != len(spec.Weapons) || len(state.Specials) != len(spec.Specials) {
		return fmt.Errorf("internal slot-state count mismatch")
	}
	for i := range spec.Weapons {
		if state.Weapons[i].Slot != spec.Weapons[i].Slot || state.Weapons[i].Current < 0 || state.Weapons[i].Current > spec.Weapons[i].Max {
			return fmt.Errorf("internal weapon slot %d outside immutable bounds", spec.Weapons[i].Slot)
		}
	}
	for i := range spec.Specials {
		if state.Specials[i].Slot != spec.Specials[i].Slot || state.Specials[i].Current < 0 || state.Specials[i].Current > spec.Specials[i].Max {
			return fmt.Errorf("internal special slot %d outside immutable bounds", spec.Specials[i].Slot)
		}
	}
	return nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
