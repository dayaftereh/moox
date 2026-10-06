package battle

import "fmt"

type tacticalInternalChoice struct {
	Pool         string
	Slot         int
	Structure    bool
	Threshold    int
	WeightedRoll int
	Weights      [19]int
}

func tacticalBeamInternalWeights(state TacticalInternalState, achilles bool) [19]int {
	var weights [19]int
	weights[0] = maxInt(0, state.EngineCurrent)
	weights[1] = maxInt(0, state.SecondaryCurrent)
	weaponMultiplier := 1
	shieldMultiplier := 1
	if achilles {
		weaponMultiplier = 3
		shieldMultiplier = 3
	}
	for _, weapon := range state.Weapons {
		if weapon.Slot >= 0 && weapon.Slot < 8 {
			weights[2+weapon.Slot] = maxInt(0, weapon.Current) * weaponMultiplier
		}
	}
	for _, special := range state.Specials {
		if special.Slot >= 0 && special.Slot < 8 {
			weights[10+special.Slot] = maxInt(0, special.Current)
		}
	}
	weights[18] = maxInt(0, state.ShieldGeneratorCurrent) * shieldMultiplier
	return weights
}

func tacticalBeamInternalTotal(state TacticalInternalState) int {
	total := maxInt(0, state.EngineCurrent) / 2
	total += maxInt(0, state.SecondaryCurrent)
	for _, weapon := range state.Weapons {
		total += maxInt(0, weapon.Current)
	}
	for _, special := range state.Specials {
		total += maxInt(0, special.Current)
	}
	total += maxInt(0, state.ShieldGeneratorCurrent)
	return total
}

func tacticalBeamInternalThreshold(targetSpec TacticalShipSpec, targetState TacticalShipState, achilles bool, targetMonsterAntaranBand bool) int {
	if achilles {
		return 100
	}
	internalTotal := tacticalBeamInternalTotal(targetState.Internal)
	remainingStructure := targetSpec.StructureMax - targetState.StructureDamage
	if remainingStructure < 0 {
		remainingStructure = 0
	}
	threshold := 0
	if denominator := internalTotal + remainingStructure; denominator > 0 {
		threshold = (internalTotal * 100) / denominator
	}
	if targetMonsterAntaranBand && threshold > 25 {
		threshold = 25
	}
	return threshold
}

func tacticalSelectBeamInternal(spec TacticalSpec, r *tacticalRuntime, targetSpec TacticalShipSpec, targetState TacticalShipState, selectionRoll int, achilles bool, targetMonsterAntaranBand bool) (tacticalInternalChoice, error) {
	choice := tacticalInternalChoice{Threshold: tacticalBeamInternalThreshold(targetSpec, targetState, achilles, targetMonsterAntaranBand)}
	if selectionRoll < 1 || selectionRoll > 100 {
		return choice, fmt.Errorf("tactical internal selection roll %d outside 1..100", selectionRoll)
	}
	if choice.Threshold < selectionRoll || tacticalBeamInternalTotal(targetState.Internal) <= 0 {
		choice.Structure = true
		return choice, nil
	}
	choice.Weights = tacticalBeamInternalWeights(targetState.Internal, achilles)
	totalWeight := 0
	for _, weight := range choice.Weights {
		totalWeight += weight
	}
	if totalWeight <= 0 {
		choice.Structure = true
		return choice, nil
	}
	weightedRoll, err := r.random(uint32(totalWeight), spec.Rules)
	if err != nil {
		return choice, err
	}
	choice.WeightedRoll = weightedRoll
	cumulative := 0
	selected := -1
	for index, weight := range choice.Weights {
		cumulative += weight
		if weightedRoll <= cumulative {
			selected = index
			break
		}
	}
	if selected < 0 {
		return choice, fmt.Errorf("tactical internal weighted roll %d exceeds total %d", weightedRoll, totalWeight)
	}
	switch {
	case selected == 0:
		choice.Pool = TacticalInternalPoolEngine
	case selected == 1:
		choice.Pool = TacticalInternalPoolSecondary
	case selected >= 2 && selected <= 9:
		choice.Pool = TacticalInternalPoolWeapon
		choice.Slot = selected - 2
	case selected >= 10 && selected <= 17:
		choice.Pool = TacticalInternalPoolSpecial
		choice.Slot = selected - 10
	case selected == 18:
		choice.Pool = TacticalInternalPoolShieldGenerator
	default:
		return choice, fmt.Errorf("unsupported tactical internal weighted index %d", selected)
	}
	return choice, nil
}

func tacticalInternalPoolCurrent(state TacticalInternalState, pool string, slot int) int {
	switch pool {
	case TacticalInternalPoolEngine:
		return maxInt(0, state.EngineCurrent)
	case TacticalInternalPoolSecondary:
		return maxInt(0, state.SecondaryCurrent)
	case TacticalInternalPoolWeapon:
		for _, item := range state.Weapons {
			if item.Slot == slot {
				return maxInt(0, item.Current)
			}
		}
	case TacticalInternalPoolSpecial:
		for _, item := range state.Specials {
			if item.Slot == slot {
				return maxInt(0, item.Current)
			}
		}
	case TacticalInternalPoolShieldGenerator:
		return maxInt(0, state.ShieldGeneratorCurrent)
	}
	return 0
}

func classifyTacticalDamageLayer(targetSpec TacticalShipSpec, targetState TacticalShipState, result *tacticalDamageApplication) {
	shieldChanged := result.ShieldBefore > result.ShieldAfter
	armorChanged := result.ArmorBefore > result.ArmorAfter
	structureChanged := result.StructureAfter > result.StructureBefore
	switch {
	case shieldChanged && armorChanged && structureChanged:
		result.Layer = "shield_armor_structure"
	case shieldChanged && armorChanged:
		result.Layer = "shield_armor"
	case shieldChanged && structureChanged:
		result.Layer = "shield_structure"
	case shieldChanged:
		result.Layer = "shield"
	case targetSpec.ShieldStrength > 0 && result.IncomingDamage > 0 && result.PostBlockDamage == 0:
		result.Layer = "shield_block"
	case armorChanged && structureChanged:
		result.Layer = "armor_structure"
	case armorChanged:
		result.Layer = "armor"
	case structureChanged:
		result.Layer = "structure"
	case targetState.StructureDamage >= targetSpec.StructureMax && result.IncomingDamage > 0:
		result.Layer = "structure_overkill"
	default:
		result.Layer = ""
	}
}

func applyTacticalBeamInternalLowerLayer(spec TacticalSpec, r *tacticalRuntime, targetSpec TacticalShipSpec, targetState *TacticalShipState, result *tacticalDamageApplication, options tacticalDamageOptions, selectionRoll int, achilles bool, targetMonsterAntaranBand bool) error {
	remaining := result.LowerLayerOverflow
	if remaining <= 0 || targetState == nil || targetState.Destroyed {
		return nil
	}
	roll := selectionRoll
	if !options.ArmorFirstBypass && targetState.ArmorCurrent > 0 {
		absorbed := remaining
		if absorbed > targetState.ArmorCurrent {
			absorbed = targetState.ArmorCurrent
		}
		targetState.ArmorCurrent -= absorbed
		remaining -= absorbed
		result.ArmorAfter = targetState.ArmorCurrent
		if remaining <= 0 {
			classifyTacticalDamageLayer(targetSpec, *targetState, result)
			return nil
		}
		var err error
		roll, err = r.random(100, spec.Rules)
		if err != nil {
			return err
		}
	}
	for remaining > 0 && !targetState.Destroyed {
		choice, err := tacticalSelectBeamInternal(spec, r, targetSpec, *targetState, roll, achilles, targetMonsterAntaranBand)
		if err != nil {
			return err
		}
		if choice.Structure {
			targetState.StructureDamage += remaining
			if targetState.StructureDamage > targetSpec.StructureMax {
				targetState.StructureDamage = targetSpec.StructureMax
			}
			remaining = 0
			break
		}
		before := tacticalInternalPoolCurrent(targetState.Internal, choice.Pool, choice.Slot)
		if before <= 0 {
			return fmt.Errorf("selected empty tactical internal pool %q slot %d", choice.Pool, choice.Slot)
		}
		if err := applyTacticalInternalDamage(spec, r, targetSpec.ShipID, choice.Pool, choice.Slot, remaining, "beam"); err != nil {
			return err
		}
		consumed := before
		if consumed > remaining {
			consumed = remaining
		}
		remaining -= consumed
		if remaining <= 0 {
			break
		}
		roll, err = r.random(100, spec.Rules)
		if err != nil {
			return err
		}
	}
	result.ArmorAfter = targetState.ArmorCurrent
	result.StructureAfter = targetState.StructureDamage
	if targetState.StructureDamage >= targetSpec.StructureMax {
		targetState.Destroyed = true
	}
	classifyTacticalDamageLayer(targetSpec, *targetState, result)
	return nil
}

func applyTacticalBeamDamageWithSpecials(spec TacticalSpec, r *tacticalRuntime, targetSpec TacticalShipSpec, targetState *TacticalShipState, sourceState TacticalShipState, incomingDamage int, options tacticalDamageOptions, selectionRoll int, achilles bool, targetMonsterAntaranBand bool) (tacticalDamageApplication, error) {
	if r == nil || r.legacy17_8Disabled || targetSpec.Internal == nil {
		return applyTacticalDamageWithSpecials(spec, r, targetSpec, targetState, sourceState, incomingDamage, options), nil
	}
	options.DeferLowerLayer = true
	result := applyTacticalDamageWithSpecials(spec, r, targetSpec, targetState, sourceState, incomingDamage, options)
	if err := applyTacticalBeamInternalLowerLayer(spec, r, targetSpec, targetState, &result, options, selectionRoll, achilles, targetMonsterAntaranBand); err != nil {
		return result, err
	}
	return result, nil
}
