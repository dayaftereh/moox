package battle

import (
	"fmt"

	"moox/internal/core"
	"moox/internal/protocol"
)

const (
	CommandBoardShip = "battle.board_ship"

	TacticalBoardingModeRaid    = "raid"
	TacticalBoardingModeCapture = "capture"

	TacticalBoardingDeliveryPhysical    = "physical"
	TacticalBoardingDeliveryTransporter = "transporter"

	TacticalBoardingUnavailableAlreadyUsed     = "boarding_already_used"
	TacticalBoardingUnavailableNoMarines       = "boarding_no_marines"
	TacticalBoardingUnavailableTarget          = "boarding_target_unavailable"
	TacticalBoardingUnavailableTargetMoving    = "boarding_target_moving"
	TacticalBoardingUnavailableOutOfRange      = "boarding_out_of_range"
	TacticalBoardingUnavailableShieldBlocked   = "transporter_shield_blocked"
	TacticalBoardingUnavailableHardShields     = "hard_shields_block_transporters"
	TacticalBoardingUnavailableSpecialDisabled = "boarding_special_disabled"
)

type TacticalBoardingDoctrineSpec struct {
	SeatID             protocol.SeatID `json:"seat_id"`
	EmpireID           core.ID         `json:"empire_id"`
	GroundCombatBonus  int             `json:"ground_combat_bonus"`
	FleetCommandoBonus int             `json:"fleet_commando_bonus"`
	FleetSecurityBonus int             `json:"fleet_security_bonus"`
	FleetEngineerBonus int             `json:"fleet_engineer_bonus"`
	Telepathic         bool            `json:"telepathic,omitempty"`
}

type TacticalBoardingActionAvailability struct {
	ShipID            core.ID  `json:"ship_id"`
	TargetShipID      core.ID  `json:"target_ship_id"`
	DeliveryMode      string   `json:"delivery_mode"`
	MaxMarineCount    int      `json:"max_marine_count"`
	AllowedModes      []string `json:"allowed_modes,omitempty"`
	Available         bool     `json:"available"`
	UnavailableReason string   `json:"unavailable_reason,omitempty"`
}

type BoardShipPayload struct {
	ShipID       core.ID `json:"ship_id"`
	TargetShipID core.ID `json:"target_ship_id"`
	MarineCount  int     `json:"marine_count"`
	Mode         string  `json:"mode"`
}

func NewBoardShipCommand(sequence uint32, payload BoardShipPayload) (protocol.Command, error) {
	return protocol.NewCommand(sequence, CommandBoardShip, payload)
}

func validateTacticalBoardingDoctrines(spec TacticalSpec) error {
	if len(spec.BoardingDoctrines) == 0 {
		return nil
	}
	expected := make(map[protocol.SeatID]core.ID)
	for _, ship := range spec.Ships {
		if prior, ok := expected[ship.SeatID]; ok && prior != ship.EmpireID {
			return fmt.Errorf("tactical Boarding doctrine seat %d maps to multiple empires", ship.SeatID)
		}
		expected[ship.SeatID] = ship.EmpireID
	}
	var previous protocol.SeatID
	for i, doctrine := range spec.BoardingDoctrines {
		if doctrine.SeatID == 0 || doctrine.EmpireID == 0 || (i > 0 && doctrine.SeatID <= previous) {
			return fmt.Errorf("tactical Boarding doctrines must use non-zero identity and ascending unique seat ids")
		}
		if expected[doctrine.SeatID] != doctrine.EmpireID {
			return fmt.Errorf("tactical Boarding doctrine seat %d empire %d does not match battle authority", doctrine.SeatID, doctrine.EmpireID)
		}
		if doctrine.FleetCommandoBonus < 0 || doctrine.FleetSecurityBonus < 0 || doctrine.FleetEngineerBonus < 0 {
			return fmt.Errorf("tactical Boarding doctrine seat %d has negative fleet bonus", doctrine.SeatID)
		}
		previous = doctrine.SeatID
	}
	return nil
}

func tacticalBoardingDoctrine(spec TacticalSpec, seatID protocol.SeatID) *TacticalBoardingDoctrineSpec {
	for i := range spec.BoardingDoctrines {
		if spec.BoardingDoctrines[i].SeatID == seatID {
			return &spec.BoardingDoctrines[i]
		}
	}
	return nil
}

func tacticalEmpireForSeat(spec TacticalSpec, seatID protocol.SeatID) core.ID {
	for _, doctrine := range spec.BoardingDoctrines {
		if doctrine.SeatID == seatID {
			return doctrine.EmpireID
		}
	}
	for _, ship := range spec.Ships {
		if ship.SeatID == seatID {
			return ship.EmpireID
		}
	}
	return 0
}

func tacticalOwnerEmpire(ship TacticalShipSpec, state *TacticalShipState) core.ID {
	if state != nil && state.OwnerEmpireID != 0 {
		return state.OwnerEmpireID
	}
	return ship.EmpireID
}

func tacticalControllerSeat(ship TacticalShipSpec, state *TacticalShipState) protocol.SeatID {
	if state != nil && state.ControllerSeatID != 0 {
		return state.ControllerSeatID
	}
	return ship.SeatID
}

func tacticalRetreatBlockedByWarpDissipator(spec TacticalSpec, r *tacticalRuntime, retreatOwner core.ID) bool {
	if r == nil || retreatOwner == 0 || r.legacy17_8Disabled {
		return false
	}
	for i := range spec.Ships {
		ship := spec.Ships[i]
		state := r.shipState(ship.ShipID)
		if state == nil || state.Destroyed || state.PostCaptureState {
			continue
		}
		if tacticalOwnerEmpire(ship, state) == retreatOwner {
			continue
		}
		if TacticalSpecialFunctional(ship, *state, r.legacy17_8Disabled, "warp_dissipator") {
			return true
		}
	}
	return false
}

func tacticalBoardingActions(spec TacticalSpec, r *tacticalRuntime, source *TacticalShipSpec, sourceState *TacticalShipState) []TacticalBoardingActionAvailability {
	if r == nil || source == nil || sourceState == nil || r.legacy17_8Disabled || sourceState.Destroyed || sourceState.ActivationComplete {
		return nil
	}
	if spec.Rules.Boarding == nil || tacticalBoardingDoctrine(spec, tacticalControllerSeat(*source, sourceState)) == nil {
		return nil
	}
	maxMarines := (sourceState.MarineCurrent + 1) / 2
	transporters := TacticalSpecialFunctional(*source, *sourceState, r.legacy17_8Disabled, "transporters")
	delivery := TacticalBoardingDeliveryPhysical
	if transporters {
		delivery = TacticalBoardingDeliveryTransporter
	}
	sourceOwner := tacticalOwnerEmpire(*source, sourceState)
	out := make([]TacticalBoardingActionAvailability, 0, len(spec.Ships)-1)
	for i := range spec.Ships {
		target := &spec.Ships[i]
		targetState := r.shipState(target.ShipID)
		if target.ShipID == source.ShipID || targetState == nil || tacticalOwnerEmpire(*target, targetState) == sourceOwner {
			continue
		}
		action := TacticalBoardingActionAvailability{
			ShipID: source.ShipID, TargetShipID: target.ShipID, DeliveryMode: delivery,
			MaxMarineCount: maxMarines, AllowedModes: []string{TacticalBoardingModeRaid, TacticalBoardingModeCapture},
			Available: true,
		}
		switch {
		case !tacticalShipTargetable(targetState) || targetState.BeingCaptured:
			action.Available = false
			action.UnavailableReason = TacticalBoardingUnavailableTarget
		case sourceState.AlreadyBoarded || (transporters && sourceState.TeleporterUsed):
			action.Available = false
			action.UnavailableReason = TacticalBoardingUnavailableAlreadyUsed
		case maxMarines <= 0:
			action.Available = false
			action.UnavailableReason = TacticalBoardingUnavailableNoMarines
		case !transporters && targetState.CurrentCombatSpeed != 0:
			action.Available = false
			action.UnavailableReason = TacticalBoardingUnavailableTargetMoving
		default:
			distance := tacticalRangeIndex(sourceState.X, sourceState.Y, targetState.X, targetState.Y)
			maxRange := spec.Rules.Boarding.PhysicalRange
			if transporters {
				maxRange = spec.Rules.Boarding.TransporterRange
			}
			if distance > maxRange {
				action.Available = false
				action.UnavailableReason = TacticalBoardingUnavailableOutOfRange
			} else if transporters && TacticalSpecialFunctional(*target, *targetState, r.legacy17_8Disabled, "hard_shields") {
				action.Available = false
				action.UnavailableReason = TacticalBoardingUnavailableHardShields
			} else if transporters {
				arc := tacticalShieldImpactArc(*targetState, *sourceState)
				if targetState.ShieldArcCurrent[arc] > 0 {
					action.Available = false
					action.UnavailableReason = TacticalBoardingUnavailableShieldBlocked
				}
			}
		}
		out = append(out, action)
	}
	return out
}

func tacticalBoardingActionForTarget(actions []TacticalBoardingActionAvailability, targetID core.ID) *TacticalBoardingActionAvailability {
	for i := range actions {
		if actions[i].TargetShipID == targetID {
			return &actions[i]
		}
	}
	return nil
}

type tacticalRaidInternalChoice struct {
	Pool string
	Slot int
}

func tacticalBoardingModeAllowed(modes []string, mode string) bool {
	for _, candidate := range modes {
		if candidate == mode {
			return true
		}
	}
	return false
}

func tacticalInternalCurrentForSlot(items []TacticalInternalWeaponState, slot int) int {
	for _, item := range items {
		if item.Slot == slot {
			return item.Current
		}
	}
	return 0
}

func tacticalInternalSpecialCurrentForSlot(items []TacticalInternalSpecialState, slot int) int {
	for _, item := range items {
		if item.Slot == slot {
			return item.Current
		}
	}
	return 0
}

func tacticalRaidInternalChoiceForState(spec TacticalSpec, r *tacticalRuntime, target *TacticalShipSpec, targetState *TacticalShipState) (tacticalRaidInternalChoice, bool, int, error) {
	if r == nil || target == nil || targetState == nil {
		return tacticalRaidInternalChoice{}, false, 0, fmt.Errorf("Raid internal selection requires a live target")
	}
	if _, err := r.random(100, spec.Rules); err != nil {
		return tacticalRaidInternalChoice{}, false, 0, err
	}
	rngDraws := 1

	current := make([]int, 19)
	current[0] = targetState.Internal.EngineCurrent
	current[1] = targetState.Internal.SecondaryCurrent
	for slot := 0; slot < 8; slot++ {
		current[2+slot] = tacticalInternalCurrentForSlot(targetState.Internal.Weapons, slot)
		current[10+slot] = tacticalInternalSpecialCurrentForSlot(targetState.Internal.Specials, slot)
	}
	current[18] = targetState.Internal.ShieldGeneratorCurrent

	totalCurrent := 0
	for _, value := range current {
		if value > 0 {
			totalCurrent += value
		}
	}
	if totalCurrent <= 0 {
		return tacticalRaidInternalChoice{}, false, rngDraws, nil
	}

	weights := make([]int, len(current))
	totalWeight := 0
	for i, value := range current {
		if value <= 0 {
			continue
		}
		weight := 100 - (value*100)/totalCurrent
		if weight < 0 {
			weight = 0
		}
		weights[i] = weight
		totalWeight += weight
	}
	if totalWeight <= 0 {
		return tacticalRaidInternalChoice{}, false, rngDraws, nil
	}

	roll, err := r.random(uint32(totalWeight), spec.Rules)
	if err != nil {
		return tacticalRaidInternalChoice{}, false, rngDraws, err
	}
	rngDraws++
	index := -1
	for i, weight := range weights {
		if weight <= 0 {
			continue
		}
		if roll <= weight {
			index = i
			break
		}
		roll -= weight
	}
	if index < 0 {
		return tacticalRaidInternalChoice{}, false, rngDraws, nil
	}
	switch {
	case index == 0:
		return tacticalRaidInternalChoice{Pool: TacticalInternalPoolEngine, Slot: -1}, true, rngDraws, nil
	case index == 1:
		return tacticalRaidInternalChoice{Pool: TacticalInternalPoolSecondary, Slot: -1}, true, rngDraws, nil
	case index >= 2 && index <= 9:
		return tacticalRaidInternalChoice{Pool: TacticalInternalPoolWeapon, Slot: index - 2}, true, rngDraws, nil
	case index >= 10 && index <= 17:
		return tacticalRaidInternalChoice{Pool: TacticalInternalPoolSpecial, Slot: index - 10}, true, rngDraws, nil
	case index == 18:
		return tacticalRaidInternalChoice{Pool: TacticalInternalPoolShieldGenerator, Slot: -1}, true, rngDraws, nil
	default:
		return tacticalRaidInternalChoice{}, false, rngDraws, nil
	}
}

func tacticalBoardingModifiers(spec TacticalSpec, r *tacticalRuntime, source TacticalShipSpec, sourceState TacticalShipState, target TacticalShipSpec, targetState TacticalShipState, mode string) (int, int, error) {
	sourceDoctrine := tacticalBoardingDoctrine(spec, tacticalControllerSeat(source, &sourceState))
	targetDoctrine := tacticalBoardingDoctrine(spec, tacticalControllerSeat(target, &targetState))
	if sourceDoctrine == nil || targetDoctrine == nil {
		return 0, 0, fmt.Errorf("Boarding doctrine is unavailable")
	}
	switch mode {
	case TacticalBoardingModeRaid:
		attacker := sourceDoctrine.FleetCommandoBonus + source.CrewBoardingBonus
		defender := targetDoctrine.FleetCommandoBonus + targetDoctrine.FleetSecurityBonus + target.CrewBoardingBonus
		if TacticalSpecialFunctional(target, targetState, r.legacy17_8Disabled, "security_stations") {
			defender += spec.Rules.Boarding.SecurityStationsDefenseBonus
		}
		return attacker, defender, nil
	case TacticalBoardingModeCapture:
		attacker := sourceDoctrine.FleetCommandoBonus + sourceDoctrine.GroundCombatBonus
		defender := targetDoctrine.FleetCommandoBonus + targetDoctrine.FleetSecurityBonus + targetDoctrine.GroundCombatBonus
		if TacticalSpecialFunctional(target, targetState, r.legacy17_8Disabled, "security_stations") {
			defender += spec.Rules.Boarding.SecurityStationsDefenseBonus
		}
		return attacker, defender, nil
	default:
		return 0, 0, fmt.Errorf("unsupported Boarding mode %q", mode)
	}
}

func resolveTacticalRaid(spec TacticalSpec, r *tacticalRuntime, seatID protocol.SeatID, commandSequence uint32, source TacticalShipSpec, sourceState *TacticalShipState, target TacticalShipSpec, targetState *TacticalShipState, selectedParty, effectiveParty int, delivery string) error {
	targetMarineBefore := targetState.MarineCurrent
	attackerModifier, defenderModifier, err := tacticalBoardingModifiers(spec, r, source, *sourceState, target, *targetState, TacticalBoardingModeRaid)
	if err != nil {
		return err
	}
	temporaryAttacker := effectiveParty
	temporaryDefender := targetMarineBefore
	attackerTempLosses := 0
	defenderTempLosses := 0
	internalSelections := 0
	rngDraws := 0

	for iteration := 0; iteration < effectiveParty; iteration++ {
		attackerRoll, err := r.random(100, spec.Rules)
		if err != nil {
			return err
		}
		defenderRoll, err := r.random(100, spec.Rules)
		if err != nil {
			return err
		}
		rngDraws += 2
		diff := attackerModifier + attackerRoll - defenderModifier - defenderRoll

		if diff > spec.Rules.Boarding.RaidInternalDamageThreshold {
			for selection := 0; selection < 2; selection++ {
				choice, ok, draws, err := tacticalRaidInternalChoiceForState(spec, r, &target, targetState)
				rngDraws += draws
				if err != nil {
					return err
				}
				if !ok {
					continue
				}
				if err := applyTacticalInternalDamage(spec, r, target.ShipID, choice.Pool, choice.Slot, 50, "boarding_raid"); err != nil {
					return err
				}
				internalSelections++
			}
		}
		if diff > spec.Rules.Boarding.RaidDefenderLossThreshold && temporaryDefender > 0 {
			temporaryDefender--
			defenderTempLosses++
		}
		if diff < spec.Rules.Boarding.RaidAttackerLossThreshold && temporaryAttacker > 0 {
			temporaryAttacker--
			attackerTempLosses++
		}
	}

	targetState.MarineCurrent = targetMarineBefore
	sourceState.MarineCurrent += selectedParty
	return r.appendEvent("boarding_resolved", seatID, commandSequence, map[string]any{
		"source_ship_id":               source.ShipID,
		"target_ship_id":               target.ShipID,
		"mode":                         TacticalBoardingModeRaid,
		"delivery_mode":                delivery,
		"result":                       "raid_complete",
		"attacker_temporary_survivors": temporaryAttacker,
		"defender_temporary_survivors": temporaryDefender,
		"target_marine_before":         targetMarineBefore,
		"target_marine_after":          targetState.MarineCurrent,
		"source_marine_after":          sourceState.MarineCurrent,
		"attacker_modifier":            attackerModifier,
		"defender_modifier":            defenderModifier,
		"attacker_temporary_losses":    attackerTempLosses,
		"defender_temporary_losses":    defenderTempLosses,
		"internal_damage_selections":   internalSelections,
		"rng_draws":                    rngDraws,
	})
}

func reconcileTacticalCaptureMovingActors(spec TacticalSpec, r *tacticalRuntime, capturedShipID core.ID, newOwner core.ID, eventSeat protocol.SeatID, commandSequence uint32) error {
	for i := 0; i < len(r.ordnance.Projectiles); {
		projectile := &r.ordnance.Projectiles[i]
		if projectile.TargetShipID != capturedShipID {
			i++
			continue
		}
		sourceEmpire := tacticalEmpireForSeat(spec, projectile.SourceSeatID)
		if sourceEmpire == 0 {
			return fmt.Errorf("projectile %d source seat %d is invalid during capture reconciliation", projectile.ProjectileID, projectile.SourceSeatID)
		}
		if sourceEmpire != newOwner {
			i++
			continue
		}
		targets := tacticalOrdnanceRetargetIDs(spec, r, sourceEmpire)
		if len(targets) == 0 {
			projectileID := projectile.ProjectileID
			r.ordnance.Projectiles = append(r.ordnance.Projectiles[:i], r.ordnance.Projectiles[i+1:]...)
			if err := r.appendEvent("projectile_destroyed", eventSeat, commandSequence, map[string]any{
				"projectile_id": projectileID, "reason": "no_target",
			}); err != nil {
				return err
			}
			continue
		}
		previous := projectile.TargetShipID
		projectile.TargetShipID = targets[0]
		if err := r.appendEvent("projectile_retargeted", eventSeat, commandSequence, map[string]any{
			"projectile_id":           projectile.ProjectileID,
			"previous_target_ship_id": previous,
			"new_target_ship_id":      projectile.TargetShipID,
			"reason":                  "target_invalid",
		}); err != nil {
			return err
		}
		i++
	}

	for i := 0; i < len(r.fighters.Groups); {
		group := &r.fighters.Groups[i]
		if group.Phase != TacticalFighterPhaseOutbound || group.TargetShipID != capturedShipID {
			i++
			continue
		}
		sourceEmpire, ok := tacticalFighterSourceEmpire(spec, *group)
		if !ok {
			return fmt.Errorf("Fighter group %d source identity is invalid during capture reconciliation", group.GroupID)
		}
		if sourceEmpire != newOwner {
			i++
			continue
		}
		retargeted, err := tacticalFighterRetarget(spec, r, group, eventSeat, commandSequence)
		if err != nil {
			return err
		}
		if retargeted {
			i++
			continue
		}
		if tacticalFighterCarrierFunctional(spec, r, *group) {
			if err := tacticalFighterBeginReturn(spec, r, group, "no_target", eventSeat, commandSequence); err != nil {
				return err
			}
			i++
			continue
		}
		groupID := group.GroupID
		multiplicity := group.Multiplicity
		r.fighters.Groups = append(r.fighters.Groups[:i], r.fighters.Groups[i+1:]...)
		if err := r.appendEvent("fighter_group_destroyed", eventSeat, commandSequence, map[string]any{
			"group_id": groupID, "reason": "carrier_unavailable", "multiplicity_before": multiplicity,
		}); err != nil {
			return err
		}
	}
	return nil
}

func resolveTacticalCapture(spec TacticalSpec, r *tacticalRuntime, seatID protocol.SeatID, commandSequence uint32, source TacticalShipSpec, sourceState *TacticalShipState, target TacticalShipSpec, targetState *TacticalShipState, effectiveParty int, delivery string) error {
	targetMarineBefore := targetState.MarineCurrent
	previousOwner := tacticalOwnerEmpire(target, targetState)
	previousController := tacticalControllerSeat(target, targetState)
	attackerModifier, defenderModifier, err := tacticalBoardingModifiers(spec, r, source, *sourceState, target, *targetState, TacticalBoardingModeCapture)
	if err != nil {
		return err
	}

	attackerSurvivors := effectiveParty
	defenderSurvivors := targetMarineBefore
	rounds := 0
	rngDraws := 0
	for attackerSurvivors > 0 && defenderSurvivors > 0 {
		attackerRoll, err := r.random(100, spec.Rules)
		if err != nil {
			return err
		}
		defenderRoll, err := r.random(100, spec.Rules)
		if err != nil {
			return err
		}
		rngDraws += 2
		attackerScore := attackerModifier + attackerRoll
		defenderScore := defenderModifier + defenderRoll
		if attackerScore <= defenderScore {
			attackerSurvivors--
		}
		if attackerScore >= defenderScore {
			defenderSurvivors--
		}
		rounds++
	}

	if attackerSurvivors <= 0 {
		targetState.MarineCurrent = defenderSurvivors
		targetState.BeingCaptured = false
		return r.appendEvent("boarding_resolved", seatID, commandSequence, map[string]any{
			"source_ship_id":       source.ShipID,
			"target_ship_id":       target.ShipID,
			"mode":                 TacticalBoardingModeCapture,
			"delivery_mode":        delivery,
			"result":               "repelled",
			"attacker_survivors":   attackerSurvivors,
			"defender_survivors":   defenderSurvivors,
			"target_marine_before": targetMarineBefore,
			"target_marine_after":  targetState.MarineCurrent,
			"source_marine_after":  sourceState.MarineCurrent,
			"attacker_modifier":    attackerModifier,
			"defender_modifier":    defenderModifier,
			"ground_combat_rounds": rounds,
			"rng_draws":            rngDraws,
		})
	}

	targetState.MarineCurrent = attackerSurvivors
	targetState.TeleporterUsed = false
	targetState.AlreadyBoarded = false
	targetState.BeingCaptured = false

	if TacticalSpecialFunctional(target, *targetState, r.legacy17_8Disabled, "quantum_detonator") {
		roll, err := r.random(uint32(spec.Rules.Boarding.CaptureQuantumRandomBound), spec.Rules)
		if err != nil {
			return err
		}
		rngDraws++
		if roll <= spec.Rules.Boarding.CaptureQuantumDestroyMax {
			targetState.StructureDamage = target.StructureMax
			targetState.Destroyed = true
			if err := r.appendEvent("boarding_resolved", seatID, commandSequence, map[string]any{
				"source_ship_id":       source.ShipID,
				"target_ship_id":       target.ShipID,
				"mode":                 TacticalBoardingModeCapture,
				"delivery_mode":        delivery,
				"result":               "quantum_detonated",
				"attacker_survivors":   attackerSurvivors,
				"defender_survivors":   defenderSurvivors,
				"target_marine_before": targetMarineBefore,
				"target_marine_after":  targetState.MarineCurrent,
				"source_marine_after":  sourceState.MarineCurrent,
				"attacker_modifier":    attackerModifier,
				"defender_modifier":    defenderModifier,
				"ground_combat_rounds": rounds,
				"quantum_roll":         roll,
				"rng_draws":            rngDraws,
			}); err != nil {
				return err
			}
			return tacticalResolveShipDestruction(spec, r, target.ShipID, seatID, commandSequence, "quantum_detonator_capture")
		}
	}

	newOwner := tacticalOwnerEmpire(source, sourceState)
	newController := tacticalControllerSeat(source, sourceState)
	targetState.OwnerEmpireID = newOwner
	targetState.ControllerSeatID = newController
	if err := reconcileTacticalCaptureMovingActors(spec, r, target.ShipID, newOwner, seatID, commandSequence); err != nil {
		return err
	}
	newDoctrine := tacticalBoardingDoctrine(spec, newController)
	telepathicBypass := newDoctrine != nil && newDoctrine.Telepathic
	targetState.PostCaptureState = !telepathicBypass
	targetState.ActivationComplete = true

	if err := r.appendEvent("boarding_resolved", seatID, commandSequence, map[string]any{
		"source_ship_id":       source.ShipID,
		"target_ship_id":       target.ShipID,
		"mode":                 TacticalBoardingModeCapture,
		"delivery_mode":        delivery,
		"result":               "captured",
		"attacker_survivors":   attackerSurvivors,
		"defender_survivors":   defenderSurvivors,
		"target_marine_before": targetMarineBefore,
		"target_marine_after":  targetState.MarineCurrent,
		"source_marine_after":  sourceState.MarineCurrent,
		"attacker_modifier":    attackerModifier,
		"defender_modifier":    defenderModifier,
		"ground_combat_rounds": rounds,
		"rng_draws":            rngDraws,
	}); err != nil {
		return err
	}
	return r.appendEvent("ship_captured", seatID, commandSequence, map[string]any{
		"ship_id":                     target.ShipID,
		"previous_owner_empire_id":    previousOwner,
		"previous_controller_seat_id": previousController,
		"owner_empire_id":             newOwner,
		"controller_seat_id":          newController,
		"marine_current":              targetState.MarineCurrent,
		"officer_removed":             false,
		"post_capture_state":          targetState.PostCaptureState,
		"telepathic_bypass":           telepathicBypass,
	})
}

func prepareBoardShip(spec Spec, r *tacticalRuntime, seatID protocol.SeatID, commandSequence uint32, payload BoardShipPayload) (*Result, error) {
	if spec.Tactical == nil || r == nil {
		return nil, fmt.Errorf("battle has no tactical Boarding runtime")
	}
	active := tacticalShipSpec(*spec.Tactical, r.state.ActiveShipID)
	activeState := r.shipState(r.state.ActiveShipID)
	if active == nil || activeState == nil || activeState.Destroyed {
		return nil, fmt.Errorf("active tactical ship %d is unavailable", r.state.ActiveShipID)
	}
	if seatID == 0 || tacticalControllerSeat(*active, activeState) != seatID {
		return nil, fmt.Errorf("seat %d does not control active ship %d", seatID, r.state.ActiveShipID)
	}
	if payload.ShipID != r.state.ActiveShipID {
		return nil, fmt.Errorf("Boarding ship %d is not active ship %d", payload.ShipID, r.state.ActiveShipID)
	}
	if payload.Mode != TacticalBoardingModeRaid && payload.Mode != TacticalBoardingModeCapture {
		return nil, fmt.Errorf("unsupported Boarding mode %q", payload.Mode)
	}

	actions := tacticalBoardingActions(*spec.Tactical, r, active, activeState)
	action := tacticalBoardingActionForTarget(actions, payload.TargetShipID)
	if action == nil {
		return nil, fmt.Errorf("target ship %d is unavailable for Boarding", payload.TargetShipID)
	}
	if !action.Available {
		return nil, fmt.Errorf("Boarding action unavailable: %s", action.UnavailableReason)
	}
	if !tacticalBoardingModeAllowed(action.AllowedModes, payload.Mode) {
		return nil, fmt.Errorf("Boarding mode %q is not available for target ship %d", payload.Mode, payload.TargetShipID)
	}
	if payload.MarineCount <= 0 || payload.MarineCount > action.MaxMarineCount {
		return nil, fmt.Errorf("Boarding marine_count %d must be within 1..%d", payload.MarineCount, action.MaxMarineCount)
	}

	target := tacticalShipSpec(*spec.Tactical, payload.TargetShipID)
	targetState := r.shipState(payload.TargetShipID)
	if target == nil || targetState == nil || targetState.Destroyed {
		return nil, fmt.Errorf("target ship %d is not a live tactical ship", payload.TargetShipID)
	}

	effectiveParty := payload.MarineCount
	if action.DeliveryMode == TacticalBoardingDeliveryTransporter &&
		TacticalSpecialFunctional(*target, *targetState, r.legacy17_8Disabled, "damper_field") {
		effectiveParty /= spec.Tactical.Rules.Boarding.DamperTransporterDivisor
	}

	tacticalBreakCloakOnAction(activeState)
	sourceMarineBefore := activeState.MarineCurrent
	activeState.MarineCurrent -= payload.MarineCount
	activeState.AlreadyBoarded = true
	if action.DeliveryMode == TacticalBoardingDeliveryTransporter {
		activeState.TeleporterUsed = true
	}
	if payload.Mode == TacticalBoardingModeCapture {
		targetState.BeingCaptured = true
	}
	if err := r.appendEvent("boarding_started", seatID, commandSequence, map[string]any{
		"source_ship_id":               active.ShipID,
		"target_ship_id":               target.ShipID,
		"mode":                         payload.Mode,
		"delivery_mode":                action.DeliveryMode,
		"selected_marine_count":        payload.MarineCount,
		"effective_party":              effectiveParty,
		"source_marine_before":         sourceMarineBefore,
		"source_marine_after_dispatch": activeState.MarineCurrent,
	}); err != nil {
		return nil, err
	}

	switch payload.Mode {
	case TacticalBoardingModeRaid:
		if err := resolveTacticalRaid(*spec.Tactical, r, seatID, commandSequence, *active, activeState, *target, targetState, payload.MarineCount, effectiveParty, action.DeliveryMode); err != nil {
			return nil, err
		}
	case TacticalBoardingModeCapture:
		if err := resolveTacticalCapture(*spec.Tactical, r, seatID, commandSequence, *active, activeState, *target, targetState, effectiveParty, action.DeliveryMode); err != nil {
			return nil, err
		}
	}
	return nil, nil
}
