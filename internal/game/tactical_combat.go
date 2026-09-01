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
	if state == nil {
		return nil, "", fmt.Errorf("tactical encounter state is nil")
	}
	if rules == nil {
		return nil, "", fmt.Errorf("tactical encounter rules are nil")
	}
	if len(encounter.Attacker.ShipIDs) != 1 || len(encounter.Defender.ShipIDs) != 1 {
		return nil, "Slice 07 tactical baseline requires exactly one combat Ship per side", nil
	}
	if len(encounter.Attacker.CivilianFleetIDs) != 0 || len(encounter.Defender.CivilianFleetIDs) != 0 {
		return nil, "Slice 07 tactical baseline does not support civilian Fleet context", nil
	}
	if len(encounter.DefenderColonyIDs) != 0 {
		return nil, "Slice 07 tactical baseline does not support Colony or planet defense", nil
	}
	attacker := shipByID(state, encounter.Attacker.ShipIDs[0])
	defender := shipByID(state, encounter.Defender.ShipIDs[0])
	if attacker == nil || defender == nil {
		return nil, "", fmt.Errorf("encounter tactical snapshot references missing strategic Ship")
	}
	if attacker.EmpireID != encounter.Attacker.EmpireID || defender.EmpireID != encounter.Defender.EmpireID {
		return nil, "", fmt.Errorf("encounter tactical Ship ownership differs from strategic side identity")
	}
	fusion, ok := tacticalDriveRule(rules, "fusion_drive")
	if !ok {
		return nil, "", fmt.Errorf("tactical rules have no fusion_drive")
	}
	nuclear, ok := tacticalDriveRule(rules, "nuclear_drive")
	if !ok {
		return nil, "", fmt.Errorf("tactical rules have no nuclear_drive")
	}
	if reason := baselineAttackerUnsupportedReason(*attacker, rules); reason != "" {
		return nil, reason, nil
	}
	if reason := baselineDefenderUnsupportedReason(*defender, rules); reason != "" {
		return nil, reason, nil
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
		Ships: []battle.TacticalShipSpec{
			baselineTacticalShip(*attacker, encounter.Attacker.SeatID, 10, 10, fusion.MaxSpeed, rules, true),
			baselineTacticalShip(*defender, encounter.Defender.SeatID, 11, 10, nuclear.MaxSpeed, rules, false),
		},
	}
	sort.Slice(tactical.Ships, func(i, j int) bool { return tactical.Ships[i].ShipID < tactical.Ships[j].ShipID })
	return tactical, "", nil
}

func tacticalDriveRule(rules *ruleset.TacticalCombatFile, id string) (ruleset.TacticalDriveSpeedRule, bool) {
	for _, drive := range rules.Drives {
		if drive.DriveID == id {
			return drive, true
		}
	}
	return ruleset.TacticalDriveSpeedRule{}, false
}

func baselineAttackerUnsupportedReason(ship core.Ship, rules *ruleset.TacticalCombatFile) string {
	if ship.Spec.HullID != rules.Frigate.HullID || ship.Spec.WarpDriveID != "fusion_drive" || ship.Spec.ComputerID != rules.Computer.ComputerID || ship.Spec.ArmorID != rules.Frigate.ArmorID || ship.Spec.ShieldID != "" || ship.Spec.FuelCellID != rules.Frigate.FuelCellID {
		return "Slice 07 tactical attacker requires Frigate/Fusion/Electronic/Titanium/no-shield/standard-fuel equipment"
	}
	if len(ship.Spec.Weapons) != 1 || ship.Spec.Weapons[0].Slot != 0 || ship.Spec.Weapons[0].WeaponID != rules.Weapon.ID || ship.Spec.Weapons[0].Count != 1 {
		return "Slice 07 tactical attacker requires exactly one slot-0 standard Laser"
	}
	return ""
}

func baselineDefenderUnsupportedReason(ship core.Ship, rules *ruleset.TacticalCombatFile) string {
	if ship.Spec.HullID != rules.Frigate.HullID || ship.Spec.WarpDriveID != "nuclear_drive" || ship.Spec.ComputerID != rules.Computer.ComputerID || ship.Spec.ArmorID != rules.Frigate.ArmorID || ship.Spec.ShieldID != "" || ship.Spec.FuelCellID != rules.Frigate.FuelCellID {
		return "Slice 07 tactical defender requires Frigate/Nuclear/Electronic/Titanium/no-shield/standard-fuel equipment"
	}
	if len(ship.Spec.Weapons) != 0 {
		return "Slice 07 tactical defender must be unarmed"
	}
	return ""
}

func baselineTacticalShip(ship core.Ship, seatID protocol.SeatID, x, y, speed int, rules *ruleset.TacticalCombatFile, armed bool) battle.TacticalShipSpec {
	out := battle.TacticalShipSpec{
		ShipID:             ship.ID,
		EmpireID:           ship.EmpireID,
		SeatID:             seatID,
		X:                  x,
		Y:                  y,
		HullID:             ship.Spec.HullID,
		WarpDriveID:        ship.Spec.WarpDriveID,
		ComputerID:         ship.Spec.ComputerID,
		ArmorID:            ship.Spec.ArmorID,
		CurrentCombatSpeed: speed,
		BeamOffense:        rules.Computer.BeamOffense,
		BeamDefense:        0,
		ArmorMax:           rules.Frigate.ArmorHits,
		StructureMax:       rules.Frigate.Structure,
	}
	if armed {
		out.Weapons = []battle.TacticalWeaponSpec{{Slot: 0, WeaponID: rules.Weapon.ID, Count: 1, MinDamage: rules.Weapon.MinDamage, MaxDamage: rules.Weapon.MaxDamage}}
	}
	return out
}
