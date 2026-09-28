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
	for _, drive := range file.Drives {
		if !hasShipDrive(components.Drives, drive.DriveID) {
			return fmt.Errorf("tactical rules reference unknown drive %q", drive.DriveID)
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

func tacticalMetadataForEncounter(state *core.GameState, encounter Encounter, rules *ruleset.TacticalCombatFile) (*battle.TacticalSpec, string, error) {
	return tacticalMetadataForEncounterWithSpecials(state, encounter, rules, nil)
}

func tacticalMetadataForEncounterWithSpecials(state *core.GameState, encounter Encounter, rules *ruleset.TacticalCombatFile, specialRules *ruleset.ShipSpecialsFile) (*battle.TacticalSpec, string, error) {
	if state == nil {
		return nil, "", fmt.Errorf("tactical encounter state is nil")
	}
	if rules == nil {
		return nil, "", fmt.Errorf("tactical encounter rules are nil")
	}
	if len(encounter.Attacker.ShipIDs) < 1 || len(encounter.Defender.ShipIDs) < 1 {
		return nil, "tactical combat requires at least one combat Ship per side", nil
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
			if reason := baselineCombatantUnsupportedReason(*ship, rules); reason != "" {
				return reason, nil
			}
			drive, ok := tacticalDriveRule(rules, ship.Spec.WarpDriveID)
			if !ok {
				return "", fmt.Errorf("tactical rules have no drive %q", ship.Spec.WarpDriveID)
			}
			y := tacticalDeploymentY(len(side.ShipIDs), i)
			// Original MOO2 uses pristine/max combat speed for an undamaged ship; engine
			// damage reduces Current_Design_Base_Combat_Speed_ toward MinSpeed.
			tactical.Ships = append(tactical.Ships, baselineTacticalShip(*ship, side.SeatID, x, y, facing, drive.MaxSpeed, rules, specialRules))
		}
		return "", nil
	}

	multiShip := len(encounter.Attacker.ShipIDs) > 1 || len(encounter.Defender.ShipIDs) > 1
	attackerX, defenderX := 10, 11
	if multiShip {
		defenderX = 14
	}
	if reason, err := appendSide(encounter.Attacker, attackerX, 0); err != nil || reason != "" {
		return nil, reason, err
	}
	if reason, err := appendSide(encounter.Defender, defenderX, 8); err != nil || reason != "" {
		return nil, reason, err
	}

	sort.Slice(tactical.Ships, func(i, j int) bool { return tactical.Ships[i].ShipID < tactical.Ships[j].ShipID })
	return tactical, "", nil
}

func tacticalDeploymentY(count, index int) int {
	if count <= 1 {
		return 10
	}
	return 9 + index*2
}

func baselineCombatantUnsupportedReason(ship core.Ship, rules *ruleset.TacticalCombatFile) string {
	if ship.Spec.HullID != rules.Frigate.HullID || ship.Spec.ComputerID != rules.Computer.ComputerID || ship.Spec.ArmorID != rules.Frigate.ArmorID || ship.Spec.ShieldID != "" || ship.Spec.FuelCellID != rules.Frigate.FuelCellID {
		return "Slice 15.5 tactical combatant requires Frigate/Electronic/Titanium/no-shield/standard-fuel equipment"
	}
	if _, ok := tacticalDriveRule(rules, ship.Spec.WarpDriveID); !ok {
		return fmt.Sprintf("Slice 15.5 tactical combatant drive %q is unsupported", ship.Spec.WarpDriveID)
	}
	if len(ship.Spec.Weapons) > 8 {
		return "tactical combatant supports at most eight weapon mounts"
	}
	previousSlot := -1
	for i, weapon := range ship.Spec.Weapons {
		if weapon.Slot < 0 || weapon.Slot > 7 || (i > 0 && weapon.Slot <= previousSlot) {
			return "tactical combatant weapon slots must be unique, ascending and within 0..7"
		}
		canonical, err := canonicalWeaponMount(weapon)
		if err != nil {
			return fmt.Sprintf("tactical combatant weapon slot %d is invalid: %v", weapon.Slot, err)
		}
		if canonical.Count <= 0 {
			return "tactical combatant weapon count must be positive"
		}
		weaponRule, ok := rules.WeaponByID(canonical.WeaponID)
		if !ok {
			return fmt.Sprintf("tactical combatant Beam %q is unsupported", canonical.WeaponID)
		}
		if weaponRule.Kind != "beam" {
			if weaponRule.TacticalSupport != "metadata_only" {
				return fmt.Sprintf("tactical combatant ordnance %q has unsupported tactical mode %q", canonical.WeaponID, weaponRule.TacticalSupport)
			}
			previousSlot = weapon.Slot
			continue
		}
		if mountBit := beamMountBit(canonical.MountType); mountBit != 0 && weaponRule.AllowedModifierMask&mountBit == 0 {
			return fmt.Sprintf("tactical combatant Beam %q does not allow mount %q", canonical.WeaponID, canonical.MountType)
		}
		for _, modifierID := range canonical.ModifierIDs {
			modifier, ok := beamModifierByID(rules.BeamModifiers, modifierID)
			if !ok || modifier.Kind != "special" || weaponRule.AllowedModifierMask&modifier.Bit == 0 {
				return fmt.Sprintf("tactical combatant Beam %q does not allow modifier %q", canonical.WeaponID, modifierID)
			}
		}
		previousSlot = weapon.Slot
	}
	return ""
}

func baselineTacticalShip(ship core.Ship, seatID protocol.SeatID, x, y, facing, speed int, rules *ruleset.TacticalCombatFile, specialRules *ruleset.ShipSpecialsFile) battle.TacticalShipSpec {
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
		SourceDesignID:       ship.SourceDesignID,
		SourceDesignRevision: ship.SourceDesignRevision,
		StrategicPictureID:   ship.Spec.StrategicPictureID,
		SourceVisualRevision: ship.SourceVisualRevision,
		CurrentCombatSpeed:   speed,
		BeamOffense:          rules.Computer.BeamOffense,
		BeamDefense:          0,
		ArmorMax:             rules.Frigate.ArmorHits,
		StructureMax:         rules.Frigate.Structure,
	}
	if ship.VisualGenome != nil {
		genome := core.CloneShipVisualGenome(*ship.VisualGenome)
		out.VisualGenome = &genome
	}
	for _, mount := range ship.Spec.Weapons {
		canonical, _ := canonicalWeaponMount(mount)
		weaponRule, _ := rules.WeaponByID(canonical.WeaponID)
		out.Weapons = append(out.Weapons, battle.TacticalWeaponSpec{
			Slot: canonical.Slot, WeaponID: canonical.WeaponID, Count: canonical.Count,
			MountType: canonical.MountType, ModifierIDs: append([]core.ShipWeaponModifierID(nil), canonical.ModifierIDs...),
			Ammo: canonical.Ammo, FiringArcID: canonical.FiringArcID,
			MinDamage: weaponRule.MinDamage, MaxDamage: weaponRule.MaxDamage,
			InnateEffectIDs: append([]string(nil), weaponRule.InnateEffectIDs...),
			TacticalSupport: weaponRule.TacticalSupport,
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
	return out
}

func tacticalDriveRule(rules *ruleset.TacticalCombatFile, id string) (ruleset.TacticalDriveSpeedRule, bool) {
	for _, drive := range rules.Drives {
		if drive.DriveID == id {
			return drive, true
		}
	}
	return ruleset.TacticalDriveSpeedRule{}, false
}
