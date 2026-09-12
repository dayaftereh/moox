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
			// Baseline ships do not yet model extra maneuvering/engine allocation. Start at
			// the drive minimum; MaxSpeed remains the future design ceiling.
			tactical.Ships = append(tactical.Ships, baselineTacticalShip(*ship, side.SeatID, x, y, facing, drive.MinSpeed, rules))
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
	if len(ship.Spec.Weapons) > 1 {
		return "Slice 15.5 tactical combatant supports at most one standard Laser"
	}
	if len(ship.Spec.Weapons) == 1 {
		weapon := ship.Spec.Weapons[0]
		if weapon.Slot != 0 || weapon.WeaponID != rules.Weapon.ID || weapon.Count != 1 {
			return "Slice 15.5 tactical combatant weapon must be exactly one slot-0 standard Laser"
		}
	}
	return ""
}

func baselineTacticalShip(ship core.Ship, seatID protocol.SeatID, x, y, facing, speed int, rules *ruleset.TacticalCombatFile) battle.TacticalShipSpec {
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
	if len(ship.Spec.Weapons) == 1 {
		out.Weapons = []battle.TacticalWeaponSpec{{Slot: 0, WeaponID: rules.Weapon.ID, Count: 1, MinDamage: rules.Weapon.MinDamage, MaxDamage: rules.Weapon.MaxDamage}}
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
