package game

import (
	"fmt"
	"sort"

	"moox/internal/battle"
	"moox/internal/core"
	"moox/internal/protocol"
	"moox/internal/ruleset"
)

func validateTacticalCombatReferences(file *ruleset.TacticalCombatFile, technologyKeyByID map[int]string, hulls map[string]ruleset.ShipHull, components ruleset.ShipMandatoryComponents) error {
	if file == nil {
		return fmt.Errorf("tactical combat rules are nil")
	}
	if key, ok := technologyKeyByID[file.Weapon.TechnologyID]; !ok || key != file.Weapon.ID {
		return fmt.Errorf("weapon %q technology %d resolves to %q", file.Weapon.ID, file.Weapon.TechnologyID, key)
	}
	if _, ok := hulls[file.Frigate.HullID]; !ok {
		return fmt.Errorf("tactical Frigate references unknown hull %q", file.Frigate.HullID)
	}
	if !hasShipArmor(components.Armors, file.Frigate.ArmorID) {
		return fmt.Errorf("tactical Frigate references unknown armor %q", file.Frigate.ArmorID)
	}
	if !hasShipFuelCell(components.FuelCells, file.Frigate.FuelCellID) {
		return fmt.Errorf("tactical Frigate references unknown fuel cell %q", file.Frigate.FuelCellID)
	}
	if !hasShipComputer(components.Computers, file.Computer.ComputerID) {
		return fmt.Errorf("tactical rules reference unknown computer %q", file.Computer.ComputerID)
	}
	for _, hull := range file.Hulls {
		if _, ok := hulls[hull.HullID]; !ok {
			return fmt.Errorf("tactical rules reference unknown hull %q", hull.HullID)
		}
	}
	for _, armor := range file.Armors {
		if !hasShipArmor(components.Armors, armor.ArmorID) {
			return fmt.Errorf("tactical rules reference unknown armor %q", armor.ArmorID)
		}
	}
	for _, drive := range file.Drives {
		if !hasShipDrive(components.Drives, drive.DriveID) {
			return fmt.Errorf("tactical rules reference unknown drive %q", drive.DriveID)
		}
	}
	for _, computer := range file.Computers {
		if !hasShipComputer(components.Computers, computer.ComputerID) {
			return fmt.Errorf("tactical rules reference unknown computer %q", computer.ComputerID)
		}
	}
	for _, shield := range file.Shields {
		if !hasShipShield(components.Shields, shield.ShieldID) {
			return fmt.Errorf("tactical rules reference unknown shield %q", shield.ShieldID)
		}
	}
	for _, fuel := range file.FuelCells {
		if !hasShipFuelCell(components.FuelCells, fuel.FuelCellID) {
			return fmt.Errorf("tactical rules reference unknown fuel cell %q", fuel.FuelCellID)
		}
	}
	return nil
}

func hasShipDrive(items []ruleset.ShipDrive, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func hasShipComputer(items []ruleset.ShipComputer, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func hasShipArmor(items []ruleset.ShipArmor, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func hasShipFuelCell(items []ruleset.ShipFuelCell, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func hasShipShield(items []ruleset.ShipShield, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}
	return false
}

func tacticalMetadataForEncounter(state *core.GameState, encounter Encounter, rules *ruleset.TacticalCombatFile) (*battle.TacticalSpec, string, error) {
	return tacticalMetadataForEncounterWithSpecials(state, encounter, rules, nil)
}

func tacticalMetadataForEncounterWithSpecials(state *core.GameState, encounter Encounter, rules *ruleset.TacticalCombatFile, specialRules *ruleset.ShipSpecialsFile, economyRules ...*EconomyRules) (*battle.TacticalSpec, string, error) {
	if state == nil {
		return nil, "", fmt.Errorf("tactical encounter state is nil")
	}
	if rules == nil {
		return nil, "", fmt.Errorf("tactical encounter rules are nil")
	}
	if len(encounter.Attacker.ShipIDs) < 1 || len(encounter.Defender.ShipIDs) < 1 {
		return nil, "tactical combat requires at least one combat Ship per side", nil
	}
	var economy *EconomyRules
	if len(economyRules) > 0 {
		economy = economyRules[0]
	}

	tactical := &battle.TacticalSpec{
		Rules: battle.TacticalRulesSnapshot{
			SchemaVersion:                rules.SchemaVersion,
			InitiativeBeamOffenseDivisor: rules.Initiative.BeamOffenseDivisor,
			RNGMultiplier:                rules.RNG.Multiplier,
			RNGIncrement:                 rules.RNG.Increment,
			BeamBaseHitThreshold:         rules.Beam.BaseHitThreshold,
			BeamMaxHitThreshold:          rules.Beam.MaxHitThreshold,
			BeamEffectiveRollCap:         rules.Beam.EffectiveRollCap,
			BeamToHitRangeModifiers:      append([]int(nil), rules.Beam.ToHitRangeModifiers...),
			BeamDamageRangeModifiers:     append([]int(nil), rules.Beam.DamageRangeModifiers...),
			Ordnance:                     tacticalOrdnanceRulesSnapshot(rules),
			Fighter:                      tacticalFighterRulesSnapshot(rules),
			Boarding:                     tacticalBoardingRulesSnapshot(rules),
			Specials:                     tacticalSpecialRulesSnapshot(rules),
		},
		InitiativeEnabled: true,
		InitialRNGState:   battle.BaselineTacticalRNGState,
		Ships:             make([]battle.TacticalShipSpec, 0, len(encounter.Attacker.ShipIDs)+len(encounter.Defender.ShipIDs)),
	}

	appendSide := func(side EncounterSide, x, facing int) (string, error) {
		for i, shipID := range side.ShipIDs {
			ship := shipByID(state, shipID)
			if ship == nil {
				return "", fmt.Errorf("encounter tactical snapshot references missing strategic Ship %d", shipID)
			}
			if ship.EmpireID != side.EmpireID {
				return "", fmt.Errorf("encounter tactical Ship %d ownership differs from strategic side identity", shipID)
			}
			if err := validateTacticalCombatant(*ship, rules, specialRules); err != nil {
				return "", err
			}
			y := tacticalDeploymentY(len(side.ShipIDs), i)
			projected, err := projectTacticalShip(*ship, side.SeatID, x, y, facing, rules, specialRules)
			if err != nil {
				return "", err
			}
			tactical.Ships = append(tactical.Ships, projected)
		}
		return "", nil
	}

	// Original MOO2 Deploy_Ships_ anchors opposing fleets at X=25 and X=55.
	// Keep the absolute Y origin simplified, but preserve the original 30-field
	// separation so missiles can exist as in-flight actors and ships must close
	// before short-range Beam fire.
	attackerX, defenderX := 25, 55
	if reason, err := appendSide(encounter.Attacker, attackerX, 0); err != nil || reason != "" {
		return nil, reason, err
	}
	if reason, err := appendSide(encounter.Defender, defenderX, 8); err != nil || reason != "" {
		return nil, reason, err
	}

	if economy != nil {
		doctrines, err := tacticalFighterDoctrinesForEncounter(state, encounter, tactical.Ships, economy)
		if err != nil {
			return nil, "", err
		}
		tactical.FighterDoctrines = doctrines
	}
	sort.Slice(tactical.Ships, func(i, j int) bool { return tactical.Ships[i].ShipID < tactical.Ships[j].ShipID })
	return tactical, "", nil
}

func bestKnownFighterBeamPayload(economy *EconomyRules, empire *core.Empire) *battle.TacticalFighterPayloadSpec {
	if economy == nil || economy.TacticalCombat == nil || empire == nil {
		return nil
	}
	var best *ruleset.TacticalWeaponRule
	for _, weapon := range economy.TacticalCombat.BeamWeapons() {
		if weapon.AllowedModifierMask&0x04 == 0 || !empireKnowsTechnology(empire, weapon.TechnologyID) {
			continue
		}
		if best == nil || weapon.OriginalIndex > best.OriginalIndex {
			copy := weapon
			best = &copy
		}
	}
	if best == nil {
		return nil
	}
	return &battle.TacticalFighterPayloadSpec{
		WeaponID: best.ID, MinDamage: best.MinDamage, MaxDamage: best.MaxDamage,
		InnateEffectIDs: append([]string(nil), best.InnateEffectIDs...), PointDefenseMode: true,
	}
}

func bestKnownFighterBombPayload(economy *EconomyRules, empire *core.Empire) *battle.TacticalFighterPayloadSpec {
	if economy == nil || economy.TacticalCombat == nil || empire == nil {
		return nil
	}
	var best *ruleset.TacticalWeaponRule
	for _, weapon := range economy.TacticalCombat.AllWeapons() {
		// Original Fighter_Bomb uses the ordinary Bomb table. Bio weapons are a
		// distinct bombardment family and are not Fighter payload candidates.
		if weapon.Kind != "bomb" || weapon.OriginalIndex < 21 || weapon.OriginalIndex > 24 ||
			!empireKnowsTechnology(empire, weapon.TechnologyID) {
			continue
		}
		if best == nil || weapon.OriginalIndex > best.OriginalIndex {
			copy := weapon
			best = &copy
		}
	}
	if best == nil {
		return nil
	}
	return &battle.TacticalFighterPayloadSpec{
		WeaponID: best.ID, MinDamage: best.MinDamage, MaxDamage: best.MaxDamage,
		InnateEffectIDs: append([]string(nil), best.InnateEffectIDs...),
	}
}

type tacticalFighterDoctrineNeeds struct {
	beam bool
	bomb bool
}

func tacticalFighterDoctrineNeedsForSeat(ships []battle.TacticalShipSpec, seat protocol.SeatID) tacticalFighterDoctrineNeeds {
	var out tacticalFighterDoctrineNeeds
	for _, ship := range ships {
		if ship.SeatID != seat {
			continue
		}
		for _, weapon := range ship.Weapons {
			if weapon.TacticalSupport != "fighter" {
				continue
			}
			switch weapon.WeaponID {
			case "fighter_bays":
				out.beam = true
			case "bomber_bays":
				out.bomb = true
			case "heavy_fighter_bays":
				out.beam, out.bomb = true, true
			}
		}
	}
	return out
}

func tacticalFighterDoctrinesForEncounter(state *core.GameState, encounter Encounter, ships []battle.TacticalShipSpec, economy *EconomyRules) ([]battle.TacticalFighterDoctrineSpec, error) {
	if state == nil || economy == nil || economy.TacticalCombat == nil {
		return nil, nil
	}
	out := make([]battle.TacticalFighterDoctrineSpec, 0, 2)
	for _, side := range []EncounterSide{encounter.Attacker, encounter.Defender} {
		needs := tacticalFighterDoctrineNeedsForSeat(ships, side.SeatID)
		if !needs.beam && !needs.bomb {
			continue
		}
		empire := empireByID(state, side.EmpireID)
		if empire == nil {
			return nil, fmt.Errorf("Fighter doctrine empire %d is missing", side.EmpireID)
		}
		drive, ok := bestKnownShipDrive(economy.ShipDrives, empire)
		if !ok || drive.ComponentIndex < 1 || drive.ComponentIndex > 6 {
			return nil, fmt.Errorf("Fighter doctrine empire %d has no supported Warp Drive", side.EmpireID)
		}
		armor, ok := bestKnownShipArmor(economy.ShipArmors, empire)
		if !ok {
			return nil, fmt.Errorf("Fighter doctrine empire %d has no supported Armor", side.EmpireID)
		}
		durability := shipArmorDurabilityMultiple(armor.ComponentIndex)
		if durability == 0 {
			return nil, fmt.Errorf("Fighter doctrine empire %d has unsupported Armor index %d", side.EmpireID, armor.ComponentIndex)
		}
		computer, ok := bestKnownShipComputer(economy.ShipComputers, empire)
		if !ok {
			return nil, fmt.Errorf("Fighter doctrine empire %d has no supported Computer", side.EmpireID)
		}
		computerRule, ok := economy.TacticalCombat.ComputerByID(computer.ID)
		if !ok {
			return nil, fmt.Errorf("Fighter doctrine empire %d computer %q has no Tactical offense rule", side.EmpireID, computer.ID)
		}
		transDimensional := false
		if modifiers, ok := economy.RaceModifiers[empire.RaceID]; ok {
			transDimensional = modifiers.TransDimensional
		}
		researchModifiers := economy.RaceResearchModifiers[empire.RaceID]
		beamPayload := bestKnownFighterBeamPayload(economy, empire)
		bombPayload := bestKnownFighterBombPayload(economy, empire)
		if needs.beam && beamPayload == nil {
			return nil, fmt.Errorf("Fighter doctrine empire %d has no Point-Defense-capable Fighter Beam", side.EmpireID)
		}
		if needs.bomb && bombPayload == nil {
			return nil, fmt.Errorf("Fighter doctrine empire %d has no supported Fighter Bomb", side.EmpireID)
		}

		// The Original also adds fleet-offensive and fleet-pilot leader bonuses.
		// MOOX has no authoritative leader/officer runtime yet, so only currently
		// represented immutable inputs are frozen here; absent leader terms are 0.
		out = append(out, battle.TacticalFighterDoctrineSpec{
			SeatID: side.SeatID, EmpireID: side.EmpireID, BestWarpDriveIndex: drive.ComponentIndex,
			ArmorDurabilityMultiple: durability, TransDimensional: transDimensional,
			FighterBeamOffense:      50 + computerRule.BeamOffense + researchModifiers.ShipAttackBonus,
			MovingActorDefenseBonus: researchModifiers.ShipDefenseBonus,
			BeamPayload:             beamPayload,
			BombPayload:             bombPayload,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SeatID < out[j].SeatID })
	return out, nil
}

func tacticalOrdnanceRulesSnapshot(rules *ruleset.TacticalCombatFile) *battle.TacticalOrdnanceRulesSnapshot {
	if rules == nil {
		return nil
	}
	out := &battle.TacticalOrdnanceRulesSnapshot{
		MovementUnitsPerField:     rules.Ordnance.MovementUnitsPerField,
		FastSpeedBonus:            rules.Ordnance.FastSpeedBonus,
		TorpedoCooldownRounds:     rules.Ordnance.TorpedoCooldownRounds,
		MIRVMultiplier:            rules.Ordnance.MIRVMultiplier,
		PlasmaDissipationPerField: rules.Ordnance.PlasmaDissipationPerField,
	}
	for _, weapon := range rules.OrdnanceWeapons() {
		if weapon.TacticalSupport != "ordnance" {
			continue
		}
		out.Weapons = append(out.Weapons, battle.TacticalOrdnanceWeaponRuleSnapshot{
			WeaponID: weapon.ID, BaseSpeed: weapon.ProjectileSpeed, BaseDurability: weapon.ProjectileDurability, BeamInterceptorDCV: weapon.BeamInterceptorDCV,
		})
	}
	return out
}

func tacticalFighterRulesSnapshot(rules *ruleset.TacticalCombatFile) *battle.TacticalFighterRulesSnapshot {
	if rules == nil {
		return nil
	}
	return &battle.TacticalFighterRulesSnapshot{
		MovementUnitsPerField: rules.Fighter.MovementUnitsPerField, CraftPerBay: rules.Fighter.CraftPerBay,
		FighterBaseSpeed: rules.Fighter.FighterBaseSpeed, BomberBaseSpeed: rules.Fighter.BomberBaseSpeed, HeavyFighterBaseSpeed: rules.Fighter.HeavyFighterBaseSpeed,
		FighterBaseHits: rules.Fighter.FighterBaseHits, BomberBaseHits: rules.Fighter.BomberBaseHits, HeavyFighterBaseHits: rules.Fighter.HeavyFighterBaseHits,
		FighterAttackStock: rules.Fighter.FighterAttackStock, BomberAttackStock: rules.Fighter.BomberAttackStock, HeavyFighterAttackStock: rules.Fighter.HeavyFighterAttackStock,
		WarpDriveSpeedStep: rules.Fighter.WarpDriveSpeedStep, TransDimensionalSpeedBonus: rules.Fighter.TransDimensionalSpeedBonus,
		ArmorDurabilityMultiples: append([]int(nil), rules.Fighter.ArmorDurabilityMultiples...), RecoveryDivisor: rules.Fighter.RecoveryDivisor,
	}
}

func tacticalBoardingRulesSnapshot(rules *ruleset.TacticalCombatFile) *battle.TacticalBoardingRulesSnapshot {
	if rules == nil {
		return nil
	}
	return &battle.TacticalBoardingRulesSnapshot{
		PhysicalRange: rules.Boarding.PhysicalRange, TransporterRange: rules.Boarding.TransporterRange,
		DamperTransporterDivisor: rules.Boarding.DamperTransporterDivisor, SecurityStationsDefenseBonus: rules.Boarding.SecurityStationsDefenseBonus,
		RaidInternalDamageThreshold: rules.Boarding.RaidInternalDamageThreshold, RaidDefenderLossThreshold: rules.Boarding.RaidDefenderLossThreshold,
		RaidAttackerLossThreshold: rules.Boarding.RaidAttackerLossThreshold, CaptureQuantumRandomBound: rules.Boarding.CaptureQuantumRandomBound,
		CaptureQuantumDestroyMax: rules.Boarding.CaptureQuantumDestroyMax, CrewQualityBonuses: append([]int(nil), rules.Boarding.CrewQualityBonuses...),
	}
}

func tacticalSpecialRulesSnapshot(rules *ruleset.TacticalCombatFile) *battle.TacticalSpecialRulesSnapshot {
	if rules == nil {
		return nil
	}
	x := rules.Specials
	return &battle.TacticalSpecialRulesSnapshot{
		AugmentedEnginesSpeedBonus: x.AugmentedEnginesSpeedBonus, BattleScannerOffenseBonus: x.BattleScannerOffenseBonus,
		ECMJammerDefenseBonus: x.ECMJammerDefenseBonus, MultiWaveECMJammerDefenseBonus: x.MultiWaveECMJammerDefenseBonus,
		WideAreaJammerSelfDefenseBonus: x.WideAreaJammerSelfDefenseBonus, WideAreaJammerFleetDefenseBonus: x.WideAreaJammerFleetDefenseBonus,
		CloakBeamDefenseBonus: x.CloakBeamDefenseBonus, CloakMissileForcedMissThreshold: x.CloakMissileForcedMissThreshold,
		PhasingInitialLifetime: x.PhasingInitialLifetime, EnergyAbsorberDivisor: x.EnergyAbsorberDivisor, DamperDamageDivisor: x.DamperDamageDivisor,
		DisplacementOrdnanceMissThreshold: x.DisplacementOrdnanceMissThreshold, DisplacementBeamMissThreshold: x.DisplacementBeamMissThreshold,
		ShieldRechargeNormalFactor: x.ShieldRechargeNormalFactor, ShieldRechargeCapacitorFactor: x.ShieldRechargeCapacitorFactor,
		HardShieldsAdditionalAbsorption: x.HardShieldsAdditionalAbsorption, MultiPhasedShieldNumerator: x.MultiPhasedShieldNumerator,
		MultiPhasedShieldDenominator: x.MultiPhasedShieldDenominator, LightningDestroyThreshold: x.LightningDestroyThreshold,
		QuantumExplosionMultiplier: x.QuantumExplosionMultiplier, AutomatedRepairBasePercent: x.AutomatedRepairBasePercent,
		SelfTeleportMaxRange: x.SelfTeleportMaxRange,
	}
}

func tacticalDeploymentY(count, index int) int {
	if count <= 1 {
		return 10
	}
	return 9 + index*2
}

func validateTacticalCombatant(ship core.Ship, rules *ruleset.TacticalCombatFile, specialRules *ruleset.ShipSpecialsFile) error {
	if _, err := tacticalFoundationForShip(ship, rules); err != nil {
		return fmt.Errorf("tactical combatant ship %d: %w", ship.ID, err)
	}
	if len(ship.Spec.Weapons) > 8 {
		return fmt.Errorf("tactical combatant ship %d has %d weapon mounts, maximum is 8", ship.ID, len(ship.Spec.Weapons))
	}
	previousSlot := -1
	for i, weapon := range ship.Spec.Weapons {
		if weapon.Slot < 0 || weapon.Slot > 7 || (i > 0 && weapon.Slot <= previousSlot) {
			return fmt.Errorf("tactical combatant ship %d weapon slots must be unique, ascending and within 0..7", ship.ID)
		}
		canonical, err := canonicalWeaponMount(weapon)
		if err != nil {
			return fmt.Errorf("tactical combatant ship %d weapon slot %d: %w", ship.ID, weapon.Slot, err)
		}
		if canonical.Count <= 0 {
			return fmt.Errorf("tactical combatant ship %d weapon slot %d count must be positive", ship.ID, weapon.Slot)
		}
		modifiers, err := canonicalWeaponModifierIDs(canonical.ModifierIDs)
		if err != nil {
			return fmt.Errorf("tactical combatant ship %d weapon slot %d modifiers: %w", ship.ID, weapon.Slot, err)
		}
		canonical.ModifierIDs = modifiers
		weaponRule, ok := rules.WeaponByID(canonical.WeaponID)
		if !ok {
			return fmt.Errorf("tactical combatant ship %d weapon %q is unknown", ship.ID, canonical.WeaponID)
		}
		switch weaponRule.Kind {
		case "beam":
			if mountBit := beamMountBit(canonical.MountType); mountBit != 0 && weaponRule.AllowedModifierMask&mountBit == 0 {
				return fmt.Errorf("tactical combatant Beam %q does not allow mount %q", canonical.WeaponID, canonical.MountType)
			}
			for _, modifierID := range canonical.ModifierIDs {
				modifier, ok := beamModifierByID(rules.BeamModifiers, modifierID)
				if !ok || modifier.Kind != "special" || weaponRule.AllowedModifierMask&modifier.Bit == 0 {
					return fmt.Errorf("tactical combatant Beam %q does not allow modifier %q", canonical.WeaponID, modifierID)
				}
			}
		case "missile", "torpedo":
			if weaponRule.TacticalSupport != "ordnance" {
				return fmt.Errorf("tactical combatant ordnance weapon %q has invalid tactical support %q", canonical.WeaponID, weaponRule.TacticalSupport)
			}
		case "bomb":
			if weaponRule.TacticalSupport != "bombardment_only" {
				return fmt.Errorf("tactical combatant Bomb/Bio weapon %q has invalid tactical support %q", canonical.WeaponID, weaponRule.TacticalSupport)
			}
		case "fighter":
			if weaponRule.TacticalSupport != "fighter" || weaponRule.CraftPerUnit != 4 {
				return fmt.Errorf("tactical combatant Fighter weapon %q has invalid tactical support %q/craft %d", canonical.WeaponID, weaponRule.TacticalSupport, weaponRule.CraftPerUnit)
			}
		default:
			if weaponRule.TacticalSupport != "metadata_only" {
				return fmt.Errorf("tactical combatant later-family weapon %q has invalid tactical support %q", canonical.WeaponID, weaponRule.TacticalSupport)
			}
		}
		previousSlot = weapon.Slot
	}

	if len(ship.Spec.Specials) > 8 {
		return fmt.Errorf("tactical combatant ship %d has %d special mounts, maximum is 8", ship.ID, len(ship.Spec.Specials))
	}
	previousSpecialSlot := -1
	seenSpecialIDs := make(map[string]struct{}, len(ship.Spec.Specials))
	for i, special := range ship.Spec.Specials {
		if special.Slot < 0 || special.Slot > 7 || (i > 0 && special.Slot <= previousSpecialSlot) {
			return fmt.Errorf("tactical combatant ship %d special slots must be unique, ascending and within 0..7", ship.ID)
		}
		if special.SpecialID == "" {
			return fmt.Errorf("tactical combatant ship %d special slot %d requires special_id", ship.ID, special.Slot)
		}
		if _, ok := seenSpecialIDs[special.SpecialID]; ok {
			return fmt.Errorf("tactical combatant ship %d has duplicate special %q", ship.ID, special.SpecialID)
		}
		if specialRules != nil {
			if _, ok := specialRules.ByID(special.SpecialID); !ok {
				return fmt.Errorf("tactical combatant ship %d special %q is unknown", ship.ID, special.SpecialID)
			}
		}
		seenSpecialIDs[special.SpecialID] = struct{}{}
		previousSpecialSlot = special.Slot
	}
	return nil
}

type tacticalFoundationValues struct {
	MinimumCombatSpeed int
	CurrentCombatSpeed int
	BeamOffense        int
	BeamDefense        int
	ArmorMax           int
	StructureMax       int
	MarineMax          int
	ShieldStrength     int
	ShieldArcMax       int
}

func tacticalFoundationForShip(ship core.Ship, rules *ruleset.TacticalCombatFile) (tacticalFoundationValues, error) {
	if rules == nil {
		return tacticalFoundationValues{}, fmt.Errorf("tactical rules are nil")
	}
	hull, ok := rules.HullByID(ship.Spec.HullID)
	if !ok {
		return tacticalFoundationValues{}, fmt.Errorf("tactical rules have no hull %q", ship.Spec.HullID)
	}
	armor, ok := rules.ArmorByID(ship.Spec.ArmorID)
	if !ok {
		return tacticalFoundationValues{}, fmt.Errorf("tactical rules have no armor %q", ship.Spec.ArmorID)
	}
	drive, ok := rules.DriveByID(ship.Spec.WarpDriveID)
	if !ok {
		return tacticalFoundationValues{}, fmt.Errorf("tactical rules have no drive %q", ship.Spec.WarpDriveID)
	}
	computer, ok := rules.ComputerByID(ship.Spec.ComputerID)
	if !ok {
		return tacticalFoundationValues{}, fmt.Errorf("tactical rules have no computer %q", ship.Spec.ComputerID)
	}
	if _, ok := rules.FuelCellByID(ship.Spec.FuelCellID); !ok {
		return tacticalFoundationValues{}, fmt.Errorf("tactical rules have no fuel cell %q", ship.Spec.FuelCellID)
	}

	unusedBonus := rules.MaxUnusedSpaceCombatSpeedBonus
	if ship.Spec.HullSpace > 0 {
		unused := ship.Spec.HullSpace - ship.Spec.SpaceUsed
		if unused < 0 {
			unused = 0
		}
		if unused > ship.Spec.HullSpace {
			unused = ship.Spec.HullSpace
		}
		unusedBonus = (rules.MaxUnusedSpaceCombatSpeedBonus*unused + ship.Spec.HullSpace - 1) / ship.Spec.HullSpace
	}
	minimumSpeed := hull.BaseCombatSpeed + drive.CombatSpeedBonus
	currentSpeed := minimumSpeed + unusedBonus
	armorMax := hull.BaseArmor + (hull.BaseArmor*armor.BonusPercent)/100
	marineMax := hull.MarineMax
	for _, special := range ship.Spec.Specials {
		if special.SpecialID == "troop_pods" {
			marineMax *= 2
			break
		}
	}
	shieldStrength, shieldArcMax := 0, 0
	if ship.Spec.ShieldID != "" {
		shield, ok := rules.ShieldByID(ship.Spec.ShieldID)
		if !ok {
			return tacticalFoundationValues{}, fmt.Errorf("tactical rules have no shield %q", ship.Spec.ShieldID)
		}
		shieldStrength = shield.Strength
		shieldArcMax = shield.Strength * rules.ShieldCapacityPerStrength * hull.ShieldHullFactor
	}
	foundation := tacticalFoundationValues{
		MinimumCombatSpeed: minimumSpeed, CurrentCombatSpeed: currentSpeed,
		BeamOffense: computer.BeamOffense, BeamDefense: currentSpeed * rules.BeamDefensePerCombatSpeed,
		ArmorMax: armorMax, StructureMax: hull.Structure, MarineMax: marineMax,
		ShieldStrength: shieldStrength, ShieldArcMax: shieldArcMax,
	}
	return applySlice178PassiveFoundation(ship, rules, foundation), nil
}

func projectTacticalShip(ship core.Ship, seatID protocol.SeatID, x, y, facing int, rules *ruleset.TacticalCombatFile, specialRules *ruleset.ShipSpecialsFile) (battle.TacticalShipSpec, error) {
	foundation, err := tacticalFoundationForShip(ship, rules)
	if err != nil {
		return battle.TacticalShipSpec{}, err
	}
	out := battle.TacticalShipSpec{
		ShipID:               ship.ID,
		EmpireID:             ship.EmpireID,
		SeatID:               seatID,
		X:                    x,
		Y:                    y,
		Facing:               facing,
		TurningMode:          battle.TacticalTurningNormal,
		HullID:               ship.Spec.HullID,
		WarpDriveID:          ship.Spec.WarpDriveID,
		ComputerID:           ship.Spec.ComputerID,
		ArmorID:              ship.Spec.ArmorID,
		ShieldID:             ship.Spec.ShieldID,
		FuelCellID:           ship.Spec.FuelCellID,
		SourceDesignID:       ship.SourceDesignID,
		SourceDesignRevision: ship.SourceDesignRevision,
		StrategicPictureID:   ship.Spec.StrategicPictureID,
		SourceVisualRevision: ship.SourceVisualRevision,
		MinimumCombatSpeed:   foundation.MinimumCombatSpeed,
		CurrentCombatSpeed:   foundation.CurrentCombatSpeed,
		BeamOffense:          foundation.BeamOffense,
		BeamDefense:          foundation.BeamDefense,
		ArmorMax:             foundation.ArmorMax,
		StructureMax:         foundation.StructureMax,
		MarineMax:            foundation.MarineMax,
		ShieldStrength:       foundation.ShieldStrength,
		ShieldArcMax:         foundation.ShieldArcMax,
	}
	if ship.VisualGenome != nil {
		genome := core.CloneShipVisualGenome(*ship.VisualGenome)
		out.VisualGenome = &genome
	}
	for _, mount := range ship.Spec.Weapons {
		canonical, _ := canonicalWeaponMount(mount)
		weaponRule, _ := rules.WeaponByID(canonical.WeaponID)
		out.Weapons = append(out.Weapons, battle.TacticalWeaponSpec{
			Slot: canonical.Slot, WeaponID: canonical.WeaponID, WeaponFamily: weaponRule.Kind, Count: canonical.Count,
			MountType: canonical.MountType, ModifierIDs: append([]core.ShipWeaponModifierID(nil), canonical.ModifierIDs...),
			Ammo: canonical.Ammo, FiringArcID: canonical.FiringArcID,
			MinDamage: weaponRule.MinDamage, MaxDamage: weaponRule.MaxDamage,
			InnateEffectIDs: append([]string(nil), weaponRule.InnateEffectIDs...),
			TacticalSupport: weaponRule.TacticalSupport, CraftPerUnit: weaponRule.CraftPerUnit,
		})
	}
	for _, special := range ship.Spec.Specials {
		support := "metadata_only"
		if specialRules != nil {
			if rule, ok := specialRules.ByID(special.SpecialID); ok {
				support = rule.TacticalSupport
			}
		}
		out.Specials = append(out.Specials, battle.TacticalSpecialSpec{
			Slot: special.Slot, SpecialID: special.SpecialID, TacticalSupport: support,
		})
		switch special.SpecialID {
		case "inertial_stabilizer":
			out.TurningMode = battle.TacticalTurningInertialStabilizer
		case "inertial_nullifier":
			out.TurningMode = battle.TacticalTurningInertialNullifier
		}
	}
	internal, err := tacticalInternalSpecForShip(ship, rules, specialRules)
	if err != nil {
		return battle.TacticalShipSpec{}, err
	}
	out.Internal = internal
	return out, nil
}
