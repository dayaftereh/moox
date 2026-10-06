package battle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"moox/internal/beam"
	"moox/internal/core"
	"moox/internal/protocol"
)

const (
	CommandMoveShip        = "battle.move_ship"
	CommandFireBeam        = "battle.fire_beam"
	CommandEndActivation   = "battle.end_activation"
	CommandWaitActivation  = "battle.wait_activation"
	CommandRetreat         = "battle.retreat"
	TacticalOutcomeVictory = "tactical_victory"
	TacticalOutcomeRetreat = "tactical_retreat"

	TacticalWeaponUnavailableConsumer  = "tactical_consumer_unavailable"
	TacticalWeaponUnavailableNotReady  = "weapon_not_ready"
	TacticalWeaponUnavailableNoTargets = "no_legal_targets"

	BaselineTacticalRNGState uint32 = 0x001500BD

	TacticalTurningNormal             = "normal"
	TacticalTurningInertialStabilizer = "inertial_stabilizer"
	TacticalTurningInertialNullifier  = "inertial_nullifier"
	TacticalCoordinateSafetyLimit     = 1_000_000
)

type TacticalRulesSnapshot struct {
	SchemaVersion                int                            `json:"schema_version"`
	InitiativeBeamOffenseDivisor int                            `json:"initiative_beam_offense_divisor"`
	RNGMultiplier                uint32                         `json:"rng_multiplier"`
	RNGIncrement                 uint32                         `json:"rng_increment"`
	BeamBaseHitThreshold         int                            `json:"beam_base_hit_threshold"`
	BeamMaxHitThreshold          int                            `json:"beam_max_hit_threshold"`
	BeamEffectiveRollCap         int                            `json:"beam_effective_roll_cap"`
	BeamToHitRangeModifiers      []int                          `json:"beam_to_hit_range_modifiers"`
	BeamDamageRangeModifiers     []int                          `json:"beam_damage_range_modifiers"`
	Ordnance                     *TacticalOrdnanceRulesSnapshot `json:"ordnance,omitempty"`
	Fighter                      *TacticalFighterRulesSnapshot  `json:"fighter,omitempty"`
	Boarding                     *TacticalBoardingRulesSnapshot `json:"boarding,omitempty"`
	Specials                     *TacticalSpecialRulesSnapshot  `json:"specials,omitempty"`
}

type TacticalFighterRulesSnapshot struct {
	MovementUnitsPerField      int   `json:"movement_units_per_field"`
	CraftPerBay                int   `json:"craft_per_bay"`
	FighterBaseSpeed           int   `json:"fighter_base_speed"`
	BomberBaseSpeed            int   `json:"bomber_base_speed"`
	HeavyFighterBaseSpeed      int   `json:"heavy_fighter_base_speed"`
	FighterBaseHits            int   `json:"fighter_base_hits"`
	BomberBaseHits             int   `json:"bomber_base_hits"`
	HeavyFighterBaseHits       int   `json:"heavy_fighter_base_hits"`
	FighterAttackStock         int   `json:"fighter_attack_stock"`
	BomberAttackStock          int   `json:"bomber_attack_stock"`
	HeavyFighterAttackStock    int   `json:"heavy_fighter_attack_stock"`
	WarpDriveSpeedStep         int   `json:"warp_drive_speed_step"`
	TransDimensionalSpeedBonus int   `json:"trans_dimensional_speed_bonus"`
	ArmorDurabilityMultiples   []int `json:"armor_durability_multiples"`
	RecoveryDivisor            int   `json:"recovery_divisor"`
}

type TacticalBoardingRulesSnapshot struct {
	PhysicalRange                int   `json:"physical_range"`
	TransporterRange             int   `json:"transporter_range"`
	DamperTransporterDivisor     int   `json:"damper_transporter_divisor"`
	SecurityStationsDefenseBonus int   `json:"security_stations_defense_bonus"`
	RaidInternalDamageThreshold  int   `json:"raid_internal_damage_threshold"`
	RaidDefenderLossThreshold    int   `json:"raid_defender_loss_threshold"`
	RaidAttackerLossThreshold    int   `json:"raid_attacker_loss_threshold"`
	CaptureQuantumRandomBound    int   `json:"capture_quantum_random_bound"`
	CaptureQuantumDestroyMax     int   `json:"capture_quantum_destroy_max"`
	CrewQualityBonuses           []int `json:"crew_quality_bonuses"`
}

type TacticalSpecialRulesSnapshot struct {
	AugmentedEnginesSpeedBonus        int `json:"augmented_engines_speed_bonus"`
	BattleScannerOffenseBonus         int `json:"battle_scanner_offense_bonus"`
	ECMJammerDefenseBonus             int `json:"ecm_jammer_defense_bonus"`
	MultiWaveECMJammerDefenseBonus    int `json:"multi_wave_ecm_jammer_defense_bonus"`
	WideAreaJammerSelfDefenseBonus    int `json:"wide_area_jammer_self_defense_bonus"`
	WideAreaJammerFleetDefenseBonus   int `json:"wide_area_jammer_fleet_defense_bonus"`
	CloakBeamDefenseBonus             int `json:"cloak_beam_defense_bonus"`
	CloakMissileForcedMissThreshold   int `json:"cloak_missile_forced_miss_threshold"`
	PhasingInitialLifetime            int `json:"phasing_initial_lifetime"`
	EnergyAbsorberDivisor             int `json:"energy_absorber_divisor"`
	DamperDamageDivisor               int `json:"damper_damage_divisor"`
	DisplacementOrdnanceMissThreshold int `json:"displacement_ordnance_miss_threshold"`
	DisplacementBeamMissThreshold     int `json:"displacement_beam_miss_threshold"`
	ShieldRechargeNormalFactor        int `json:"shield_recharge_normal_factor"`
	ShieldRechargeCapacitorFactor     int `json:"shield_recharge_capacitor_factor"`
	HardShieldsAdditionalAbsorption   int `json:"hard_shields_additional_absorption"`
	MultiPhasedShieldNumerator        int `json:"multi_phased_shield_numerator"`
	MultiPhasedShieldDenominator      int `json:"multi_phased_shield_denominator"`
	LightningDestroyThreshold         int `json:"lightning_destroy_threshold"`
	QuantumExplosionMultiplier        int `json:"quantum_explosion_multiplier"`
	AutomatedRepairBasePercent        int `json:"automated_repair_base_percent"`
	SelfTeleportMaxRange              int `json:"self_teleport_max_range"`
}

type TacticalWeaponSpec struct {
	Slot            int                         `json:"slot"`
	WeaponID        string                      `json:"weapon_id"`
	WeaponFamily    string                      `json:"weapon_family"`
	Count           int                         `json:"count"`
	MountType       core.ShipWeaponMountType    `json:"mount_type,omitempty"`
	ModifierIDs     []core.ShipWeaponModifierID `json:"modifier_ids,omitempty"`
	Ammo            int                         `json:"ammo,omitempty"`
	FiringArcID     core.ShipWeaponFiringArcID  `json:"firing_arc_id,omitempty"`
	MinDamage       int                         `json:"min_damage"`
	MaxDamage       int                         `json:"max_damage"`
	InnateEffectIDs []string                    `json:"innate_effect_ids,omitempty"`
	TacticalSupport string                      `json:"tactical_support,omitempty"`
	CraftPerUnit    int                         `json:"craft_per_unit,omitempty"`
}

type TacticalSpecialSpec struct {
	Slot            int    `json:"slot"`
	SpecialID       string `json:"special_id"`
	TacticalSupport string `json:"tactical_support"`
}

type TacticalFighterPayloadSpec struct {
	WeaponID         string   `json:"weapon_id"`
	MinDamage        int      `json:"min_damage"`
	MaxDamage        int      `json:"max_damage"`
	InnateEffectIDs  []string `json:"innate_effect_ids,omitempty"`
	PointDefenseMode bool     `json:"point_defense_mode,omitempty"`
}

type TacticalFighterDoctrineSpec struct {
	SeatID                  protocol.SeatID             `json:"seat_id"`
	EmpireID                core.ID                     `json:"empire_id"`
	BestWarpDriveIndex      int                         `json:"best_warp_drive_index"`
	ArmorDurabilityMultiple int                         `json:"armor_durability_multiple"`
	TransDimensional        bool                        `json:"trans_dimensional,omitempty"`
	BeamPayload             *TacticalFighterPayloadSpec `json:"beam_payload,omitempty"`
	BombPayload             *TacticalFighterPayloadSpec `json:"bomb_payload,omitempty"`
	FighterBeamOffense      int                         `json:"fighter_beam_offense"`
	MovingActorDefenseBonus int                         `json:"moving_actor_defense_bonus"`
}

type TacticalInternalWeaponSpec struct {
	Slot     int `json:"slot"`
	Max      int `json:"max"`
	UnitSize int `json:"unit_size"`
}

type TacticalInternalSpecialSpec struct {
	Slot int `json:"slot"`
	Max  int `json:"max"`
}

type TacticalInternalSpec struct {
	EngineMax          int                           `json:"engine_max"`
	SecondaryMax       int                           `json:"secondary_max"`
	Weapons            []TacticalInternalWeaponSpec  `json:"weapons,omitempty"`
	Specials           []TacticalInternalSpecialSpec `json:"specials,omitempty"`
	ShieldGeneratorMax int                           `json:"shield_generator_max"`
}

type TacticalShipSpec struct {
	ShipID                  core.ID                `json:"ship_id"`
	EmpireID                core.ID                `json:"empire_id"`
	SeatID                  protocol.SeatID        `json:"seat_id"`
	X                       int                    `json:"x"`
	Y                       int                    `json:"y"`
	Facing                  int                    `json:"facing"`
	TurningMode             string                 `json:"turning_mode,omitempty"`
	HullID                  string                 `json:"hull_id"`
	WarpDriveID             string                 `json:"warp_drive_id"`
	ComputerID              string                 `json:"computer_id"`
	ArmorID                 string                 `json:"armor_id"`
	ShieldID                string                 `json:"shield_id,omitempty"`
	FuelCellID              string                 `json:"fuel_cell_id"`
	SourceDesignID          core.ID                `json:"source_design_id,omitempty"`
	SourceDesignRevision    uint32                 `json:"source_design_revision,omitempty"`
	StrategicPictureID      int                    `json:"strategic_picture_id,omitempty"`
	SourceVisualRevision    uint32                 `json:"source_visual_revision,omitempty"`
	VisualGenome            *core.ShipVisualGenome `json:"visual_genome,omitempty"`
	Weapons                 []TacticalWeaponSpec   `json:"weapons,omitempty"`
	Specials                []TacticalSpecialSpec  `json:"specials,omitempty"`
	Internal                *TacticalInternalSpec  `json:"internal,omitempty"`
	MinimumCombatSpeed      int                    `json:"minimum_combat_speed"`
	CurrentCombatSpeed      int                    `json:"current_combat_speed"`
	BeamOffense             int                    `json:"beam_offense"`
	BeamDefense             int                    `json:"beam_defense"`
	MissileDefense          int                    `json:"missile_defense"`
	MissileScannerReduction int                    `json:"missile_scanner_reduction"`
	ArmorMax                int                    `json:"armor_max"`
	StructureMax            int                    `json:"structure_max"`
	MarineMax               int                    `json:"marine_max"`
	CrewBoardingBonus       int                    `json:"crew_boarding_bonus"`
	ShieldStrength          int                    `json:"shield_strength"`
	ShieldArcMax            int                    `json:"shield_arc_max"`
}

type TacticalSpec struct {
	Rules             TacticalRulesSnapshot          `json:"rules"`
	InitiativeEnabled bool                           `json:"initiative_enabled"`
	InitialRNGState   uint32                         `json:"initial_rng_state,omitempty"`
	Ships             []TacticalShipSpec             `json:"ships"`
	FighterDoctrines  []TacticalFighterDoctrineSpec  `json:"fighter_doctrines,omitempty"`
	BoardingDoctrines []TacticalBoardingDoctrineSpec `json:"boarding_doctrines,omitempty"`
}

type TacticalWeaponState struct {
	Slot             int    `json:"slot"`
	Ready            bool   `json:"ready"`
	CountCurrent     int    `json:"count_current"`
	SpecialReadiness string `json:"special_readiness,omitempty"`
}

type TacticalShipState struct {
	ShipID                core.ID               `json:"ship_id"`
	X                     int                   `json:"x"`
	Y                     int                   `json:"y"`
	Facing                int                   `json:"facing"`
	CurrentCombatSpeed    int                   `json:"current_combat_speed"`
	MovementCurrent       int                   `json:"movement_current"`
	MovementMax           int                   `json:"movement_max"`
	ActivationComplete    bool                  `json:"activation_complete"`
	ArmorCurrent          int                   `json:"armor_current"`
	ShieldArcCurrent      [4]int                `json:"shield_arc_current"`
	StructureDamage       int                   `json:"structure_damage"`
	MarineCurrent         int                   `json:"marine_current"`
	OwnerEmpireID         core.ID               `json:"owner_empire_id"`
	ControllerSeatID      protocol.SeatID       `json:"controller_seat_id"`
	CloakState            string                `json:"cloak_state"`
	PhasingTurnsRemaining int                   `json:"phasing_turns_remaining,omitempty"`
	EnergyAbsorberStored  int                   `json:"energy_absorber_stored,omitempty"`
	TeleporterUsed        bool                  `json:"teleporter_used,omitempty"`
	AlreadyBoarded        bool                  `json:"already_boarded,omitempty"`
	BeingCaptured         bool                  `json:"being_captured,omitempty"`
	PostCaptureState      bool                  `json:"post_capture_state,omitempty"`
	Weapons               []TacticalWeaponState `json:"weapons,omitempty"`
	Internal              TacticalInternalState `json:"internal"`
	Destroyed             bool                  `json:"destroyed"`
}

type TacticalState struct {
	Round               uint32              `json:"round"`
	InitiativeOrder     []core.ID           `json:"initiative_order"`
	ActiveShipID        core.ID             `json:"active_ship_id"`
	NextCommandSequence uint32              `json:"next_command_sequence"`
	RNGState            uint32              `json:"rng_state,omitempty"`
	PassKind            string              `json:"pass_kind,omitempty"`
	TimeWarpSeatID      protocol.SeatID     `json:"time_warp_seat_id,omitempty"`
	TimeWarpQueue       []core.ID           `json:"time_warp_queue,omitempty"`
	Ships               []TacticalShipState `json:"ships"`
}

type TacticalEvent struct {
	Sequence        uint64          `json:"sequence"`
	Kind            string          `json:"kind"`
	SeatID          protocol.SeatID `json:"seat_id,omitempty"`
	CommandSequence uint32          `json:"command_sequence,omitempty"`
	Data            json.RawMessage `json:"data,omitempty"`
}

type TacticalMoveOption struct {
	MovementKind           string `json:"movement_kind"`
	X                      int    `json:"x"`
	Y                      int    `json:"y"`
	MoveCost               int    `json:"move_cost"`
	ResultingFacing        int    `json:"resulting_facing"`
	MovementRemainingAfter int    `json:"movement_remaining_after"`
}

type TacticalWeaponView struct {
	Slot             int                         `json:"slot"`
	WeaponID         string                      `json:"weapon_id"`
	Count            int                         `json:"count"`
	MountType        core.ShipWeaponMountType    `json:"mount_type,omitempty"`
	ModifierIDs      []core.ShipWeaponModifierID `json:"modifier_ids,omitempty"`
	Ammo             int                         `json:"ammo,omitempty"`
	FiringArcID      core.ShipWeaponFiringArcID  `json:"firing_arc_id,omitempty"`
	MinDamage        int                         `json:"min_damage"`
	MaxDamage        int                         `json:"max_damage"`
	InnateEffectIDs  []string                    `json:"innate_effect_ids,omitempty"`
	TacticalSupport  string                      `json:"tactical_support,omitempty"`
	Ready            bool                        `json:"ready"`
	SpecialReadiness string                      `json:"special_readiness,omitempty"`
}

type TacticalShipView struct {
	ShipID                  core.ID                `json:"ship_id"`
	EmpireID                core.ID                `json:"empire_id"`
	SeatID                  protocol.SeatID        `json:"seat_id"`
	HullID                  string                 `json:"hull_id"`
	WarpDriveID             string                 `json:"warp_drive_id"`
	ComputerID              string                 `json:"computer_id"`
	ArmorID                 string                 `json:"armor_id"`
	ShieldID                string                 `json:"shield_id,omitempty"`
	SourceDesignID          core.ID                `json:"source_design_id,omitempty"`
	SourceDesignRevision    uint32                 `json:"source_design_revision,omitempty"`
	StrategicPictureID      int                    `json:"strategic_picture_id,omitempty"`
	SourceVisualRevision    uint32                 `json:"source_visual_revision,omitempty"`
	VisualGenome            *core.ShipVisualGenome `json:"visual_genome,omitempty"`
	X                       int                    `json:"x"`
	Y                       int                    `json:"y"`
	Facing                  int                    `json:"facing"`
	CurrentCombatSpeed      int                    `json:"current_combat_speed"`
	MovementCurrent         int                    `json:"movement_current"`
	MovementMax             int                    `json:"movement_max"`
	ActivationComplete      bool                   `json:"activation_complete"`
	ArmorCurrent            int                    `json:"armor_current"`
	ArmorMax                int                    `json:"armor_max"`
	MarineCurrent           int                    `json:"marine_current"`
	MarineMax               int                    `json:"marine_max"`
	TeleporterUsed          bool                   `json:"teleporter_used,omitempty"`
	AlreadyBoarded          bool                   `json:"already_boarded,omitempty"`
	BeingCaptured           bool                   `json:"being_captured,omitempty"`
	PostCaptureState        bool                   `json:"post_capture_state,omitempty"`
	CloakState              string                 `json:"cloak_state"`
	PhasingTurnsRemaining   int                    `json:"phasing_turns_remaining,omitempty"`
	ShieldStrength          int                    `json:"shield_strength"`
	ShieldArcCurrent        [4]int                 `json:"shield_arc_current"`
	ShieldArcMax            int                    `json:"shield_arc_max"`
	StructureCurrent        int                    `json:"structure_current"`
	StructureMax            int                    `json:"structure_max"`
	BeamOffense             int                    `json:"beam_offense"`
	BeamDefense             int                    `json:"beam_defense"`
	MissileDefense          int                    `json:"missile_defense"`
	MissileScannerReduction int                    `json:"missile_scanner_reduction"`
	Weapons                 []TacticalWeaponView   `json:"weapons,omitempty"`
	Internal                TacticalInternalView   `json:"internal"`
	Destroyed               bool                   `json:"destroyed"`
}

type TacticalFireTarget struct {
	TargetShipID         core.ID `json:"target_ship_id,omitempty"`
	TargetProjectileID   uint64  `json:"target_projectile_id,omitempty"`
	TargetFighterGroupID uint64  `json:"target_fighter_group_id,omitempty"`
	RangeIndex           int     `json:"range_index"`
}

type TacticalFireAction struct {
	ShipID     core.ID              `json:"ship_id"`
	WeaponSlot int                  `json:"weapon_slot"`
	WeaponID   string               `json:"weapon_id"`
	Targets    []TacticalFireTarget `json:"targets"`
}

type TacticalWeaponActionAvailability struct {
	ShipID            core.ID              `json:"ship_id"`
	WeaponSlot        int                  `json:"weapon_slot"`
	WeaponID          string               `json:"weapon_id"`
	WeaponFamily      string               `json:"weapon_family"`
	CommandKind       string               `json:"command_kind,omitempty"`
	Available         bool                 `json:"available"`
	UnavailableReason string               `json:"unavailable_reason,omitempty"`
	Targets           []TacticalFireTarget `json:"targets,omitempty"`
}

type TacticalView struct {
	State             TacticalState                        `json:"state"`
	Ordnance          TacticalOrdnanceState                `json:"ordnance"`
	Fighters          TacticalFighterView                  `json:"fighters"`
	Events            []TacticalEvent                      `json:"events"`
	Ships             []TacticalShipView                   `json:"ships"`
	LegalMoves        []TacticalMoveOption                 `json:"legal_moves,omitempty"`
	WeaponActions     []TacticalWeaponActionAvailability   `json:"weapon_actions,omitempty"`
	LegalFireActions  []TacticalFireAction                 `json:"legal_fire_actions,omitempty"`
	BoardingActions   []TacticalBoardingActionAvailability `json:"boarding_actions,omitempty"`
	CanEndActivation  bool                                 `json:"can_end_activation"`
	CanRetreat        bool                                 `json:"can_retreat"`
	CanWaitActivation bool                                 `json:"can_wait_activation"`
	WaitTargetShipIDs []core.ID                            `json:"wait_target_ship_ids,omitempty"`
}

type MoveShipPayload struct {
	ShipID core.ID `json:"ship_id"`
	X      int     `json:"x"`
	Y      int     `json:"y"`
}

type FireBeamPayload struct {
	ShipID               core.ID `json:"ship_id"`
	TargetShipID         core.ID `json:"target_ship_id,omitempty"`
	TargetProjectileID   uint64  `json:"target_projectile_id,omitempty"`
	TargetFighterGroupID uint64  `json:"target_fighter_group_id,omitempty"`
	WeaponSlot           int     `json:"weapon_slot"`
}

type ReactiveBeamPointDefenseRequest struct {
	ShipID            core.ID `json:"ship_id"`
	WeaponSlot        int     `json:"weapon_slot"`
	TargetX           int     `json:"target_x"`
	TargetY           int     `json:"target_y"`
	TargetBeamDefense int     `json:"target_beam_defense"`
}

type ReactiveBeamPointDefenseShot struct {
	PhysicalWeaponIndex int  `json:"physical_weapon_index"`
	PhysicalWeaponCount int  `json:"physical_weapon_count"`
	BurstShotIndex      int  `json:"burst_shot_index"`
	BurstShotCount      int  `json:"burst_shot_count"`
	ShotIndex           int  `json:"shot_index"`
	ShotCount           int  `json:"shot_count"`
	RangeIndex          int  `json:"range_index"`
	RawHitRoll          int  `json:"raw_hit_roll"`
	EffectiveHitRoll    int  `json:"effective_hit_roll"`
	Threshold           int  `json:"threshold"`
	Hit                 bool `json:"hit"`
	Damage              int  `json:"damage"`
}

type ReactiveBeamPointDefenseResult struct {
	ShipID          core.ID                        `json:"ship_id"`
	WeaponSlot      int                            `json:"weapon_slot"`
	WeaponID        string                         `json:"weapon_id"`
	MountType       core.ShipWeaponMountType       `json:"mount_type"`
	ModifierIDs     []core.ShipWeaponModifierID    `json:"modifier_ids,omitempty"`
	InnateEffectIDs []string                       `json:"innate_effect_ids,omitempty"`
	TargetX         int                            `json:"target_x"`
	TargetY         int                            `json:"target_y"`
	Shots           []ReactiveBeamPointDefenseShot `json:"shots"`
}

type EndActivationPayload struct {
	ShipID core.ID `json:"ship_id"`
}
type WaitActivationPayload struct {
	ShipID       core.ID `json:"ship_id"`
	TargetShipID core.ID `json:"target_ship_id,omitempty"`
}
type RetreatPayload struct {
	ShipID core.ID `json:"ship_id"`
}

type tacticalRuntime struct {
	state                TacticalState
	ordnance             TacticalOrdnanceState
	fighters             TacticalFighterState
	events               []TacticalEvent
	nextEventSequence    uint64
	nextMovingActorOrder uint64
	legacy17_8Disabled   bool
}

type PreparedCommand struct {
	baseRevision uint64
	runtime      tacticalRuntime
	result       *Result
}

func (p *PreparedCommand) Result() *Result {
	if p == nil || p.result == nil {
		return nil
	}
	out := *p.result
	out.WinnerSeats = append([]protocol.SeatID(nil), p.result.WinnerSeats...)
	out.DestroyedShipIDs = append([]core.ID(nil), p.result.DestroyedShipIDs...)
	return &out
}

func NewMoveShipCommand(sequence uint32, payload MoveShipPayload) (protocol.Command, error) {
	return protocol.NewCommand(sequence, CommandMoveShip, payload)
}

func NewFireBeamCommand(sequence uint32, payload FireBeamPayload) (protocol.Command, error) {
	return protocol.NewCommand(sequence, CommandFireBeam, payload)
}

func NewLaunchFightersCommand(sequence uint32, payload LaunchFightersPayload) (protocol.Command, error) {
	return protocol.NewCommand(sequence, CommandLaunchFighters, payload)
}

func NewEndActivationCommand(sequence uint32, payload EndActivationPayload) (protocol.Command, error) {
	return protocol.NewCommand(sequence, CommandEndActivation, payload)
}
func NewWaitActivationCommand(sequence uint32, payload WaitActivationPayload) (protocol.Command, error) {
	return protocol.NewCommand(sequence, CommandWaitActivation, payload)
}
func NewRetreatCommand(sequence uint32, payload RetreatPayload) (protocol.Command, error) {
	return protocol.NewCommand(sequence, CommandRetreat, payload)
}

func validateTacticalSpec(spec Spec) error {
	if spec.Tactical == nil {
		return nil
	}
	if spec.SystemID == 0 || len(spec.Participants) != 2 {
		return fmt.Errorf("tactical battle requires a two-seat strategic battle")
	}
	if spec.TacticalUnsupportedReason != "" {
		return fmt.Errorf("tactical spec and unsupported reason are mutually exclusive")
	}
	t := spec.Tactical
	if !t.InitiativeEnabled {
		return fmt.Errorf("Slice 07 tactical fixture requires initiative")
	}
	if t.InitialRNGState != BaselineTacticalRNGState {
		return fmt.Errorf("Slice 07 tactical fixture requires initial RNG state 0x%08X", BaselineTacticalRNGState)
	}
	if err := validateBaselineRules(t.Rules); err != nil {
		return err
	}
	if err := validateTacticalOrdnanceRulesForSpec(*t); err != nil {
		return err
	}
	if err := validateTacticalFighterDoctrines(*t); err != nil {
		return err
	}
	if err := validateTacticalBoardingDoctrines(*t); err != nil {
		return err
	}
	if len(spec.Attacker.ShipIDs) < 1 || len(spec.Defender.ShipIDs) < 1 {
		return fmt.Errorf("tactical battle requires at least one combat Ship per side")
	}
	if len(t.Ships) != len(spec.Attacker.ShipIDs)+len(spec.Defender.ShipIDs) {
		return fmt.Errorf("tactical ship count does not match strategic sides")
	}

	byID := make(map[core.ID]TacticalShipSpec, len(t.Ships))
	occupied := make(map[[2]int]core.ID, len(t.Ships))
	var previous core.ID
	for i, ship := range t.Ships {
		if ship.ShipID == 0 || (i > 0 && ship.ShipID <= previous) {
			return fmt.Errorf("tactical ships must be strictly ascending by strategic ship id")
		}
		previous = ship.ShipID
		if ship.Facing < 0 || ship.Facing > 15 {
			return fmt.Errorf("tactical ship %d facing %d is outside 0..15", ship.ShipID, ship.Facing)
		}
		if absInt(ship.X) > TacticalCoordinateSafetyLimit || absInt(ship.Y) > TacticalCoordinateSafetyLimit {
			return fmt.Errorf("tactical ship %d coordinate is outside technical safety envelope", ship.ShipID)
		}
		if prior := occupied[[2]int{ship.X, ship.Y}]; prior != 0 {
			return fmt.Errorf("tactical ships %d and %d share coordinate (%d,%d)", prior, ship.ShipID, ship.X, ship.Y)
		}
		occupied[[2]int{ship.X, ship.Y}] = ship.ShipID
		if ship.VisualGenome == nil && ship.SourceVisualRevision != 0 {
			return fmt.Errorf("tactical ship %d has source_visual_revision %d without visual_genome", ship.ShipID, ship.SourceVisualRevision)
		}
		if ship.VisualGenome != nil {
			if ship.SourceVisualRevision == 0 {
				return fmt.Errorf("tactical ship %d visual_genome requires positive source_visual_revision", ship.ShipID)
			}
			if err := core.ValidateShipVisualGenome(*ship.VisualGenome); err != nil {
				return fmt.Errorf("tactical ship %d visual_genome: %w", ship.ShipID, err)
			}
		}
		if err := validateTacticalCombatantSnapshot(ship, t.Rules); err != nil {
			return err
		}
		byID[ship.ShipID] = ship
	}
	for _, id := range spec.Attacker.ShipIDs {
		ship, ok := byID[id]
		if !ok {
			return fmt.Errorf("tactical snapshot is missing attacker ship %d", id)
		}
		if ship.EmpireID != spec.Attacker.EmpireID || ship.SeatID != spec.Attacker.SeatID {
			return fmt.Errorf("tactical attacker ship %d authority does not match strategic side", id)
		}
	}
	for _, id := range spec.Defender.ShipIDs {
		ship, ok := byID[id]
		if !ok {
			return fmt.Errorf("tactical snapshot is missing defender ship %d", id)
		}
		if ship.EmpireID != spec.Defender.EmpireID || ship.SeatID != spec.Defender.SeatID {
			return fmt.Errorf("tactical defender ship %d authority does not match strategic side", id)
		}
	}
	return nil
}

func validateTacticalFighterDoctrines(spec TacticalSpec) error {
	expected := make(map[protocol.SeatID]core.ID)
	requiredBeam := make(map[protocol.SeatID]bool)
	requiredBomb := make(map[protocol.SeatID]bool)
	for _, ship := range spec.Ships {
		hasFighters := false
		for _, weapon := range ship.Weapons {
			if weapon.TacticalSupport != "fighter" {
				continue
			}
			hasFighters = true
			switch weapon.WeaponID {
			case "fighter_bays":
				requiredBeam[ship.SeatID] = true
			case "bomber_bays":
				requiredBomb[ship.SeatID] = true
			case "heavy_fighter_bays":
				requiredBeam[ship.SeatID] = true
				requiredBomb[ship.SeatID] = true
			default:
				return fmt.Errorf("tactical Fighter doctrine carrier ship %d has unsupported Fighter row %q", ship.ShipID, weapon.WeaponID)
			}
		}
		if !hasFighters {
			continue
		}
		if prior, ok := expected[ship.SeatID]; ok && prior != ship.EmpireID {
			return fmt.Errorf("tactical Fighter doctrine seat %d maps to multiple empires", ship.SeatID)
		}
		expected[ship.SeatID] = ship.EmpireID
	}
	if len(spec.FighterDoctrines) != len(expected) {
		return fmt.Errorf("tactical Fighter doctrine count %d, expected %d", len(spec.FighterDoctrines), len(expected))
	}
	var previousSeat protocol.SeatID
	for i, doctrine := range spec.FighterDoctrines {
		if doctrine.SeatID == 0 || doctrine.EmpireID == 0 || (i > 0 && doctrine.SeatID <= previousSeat) {
			return fmt.Errorf("tactical Fighter doctrines must use non-zero identity and ascending unique seat ids")
		}
		expectedEmpire, ok := expected[doctrine.SeatID]
		if !ok || expectedEmpire != doctrine.EmpireID {
			return fmt.Errorf("tactical Fighter doctrine seat %d empire %d does not match Fighter carrier authority", doctrine.SeatID, doctrine.EmpireID)
		}
		if doctrine.BestWarpDriveIndex < 1 || doctrine.BestWarpDriveIndex > 6 {
			return fmt.Errorf("tactical Fighter doctrine seat %d has invalid drive index %d", doctrine.SeatID, doctrine.BestWarpDriveIndex)
		}
		switch doctrine.ArmorDurabilityMultiple {
		case 1, 2, 4, 6, 8, 10:
		default:
			return fmt.Errorf("tactical Fighter doctrine seat %d has invalid armor durability multiple %d", doctrine.SeatID, doctrine.ArmorDurabilityMultiple)
		}
		if err := validateTacticalFighterPayload(doctrine.BeamPayload); err != nil {
			return fmt.Errorf("tactical Fighter doctrine seat %d Beam payload: %w", doctrine.SeatID, err)
		}
		if err := validateTacticalFighterPayload(doctrine.BombPayload); err != nil {
			return fmt.Errorf("tactical Fighter doctrine seat %d Bomb payload: %w", doctrine.SeatID, err)
		}
		if requiredBeam[doctrine.SeatID] {
			if doctrine.BeamPayload == nil {
				return fmt.Errorf("tactical Fighter doctrine seat %d requires a Beam payload", doctrine.SeatID)
			}
			if !doctrine.BeamPayload.PointDefenseMode {
				return fmt.Errorf("tactical Fighter doctrine seat %d Beam payload must freeze Point-Defense mode", doctrine.SeatID)
			}
		}
		if requiredBomb[doctrine.SeatID] && doctrine.BombPayload == nil {
			return fmt.Errorf("tactical Fighter doctrine seat %d requires a Bomb payload", doctrine.SeatID)
		}
		previousSeat = doctrine.SeatID
	}
	return nil
}

func validateTacticalFighterPayload(payload *TacticalFighterPayloadSpec) error {
	if payload == nil {
		return nil
	}
	if payload.WeaponID == "" || payload.MinDamage < 0 || payload.MaxDamage < payload.MinDamage {
		return fmt.Errorf("invalid Fighter payload")
	}
	return nil
}

func canonicalTacticalWeaponMountType(value core.ShipWeaponMountType) core.ShipWeaponMountType {
	if value == "" {
		return core.ShipWeaponMountStandard
	}
	return value
}

func tacticalWeaponModifierRank(id core.ShipWeaponModifierID) int {
	switch id {
	case core.ShipWeaponModifierArmorPiercing:
		return 0
	case core.ShipWeaponModifierContinuous:
		return 1
	case core.ShipWeaponModifierNoRangeDissipation:
		return 2
	case core.ShipWeaponModifierShieldPiercing:
		return 3
	case core.ShipWeaponModifierAutofire:
		return 4
	case core.ShipWeaponModifierEnveloping:
		return 5
	case core.ShipWeaponModifierMIRV:
		return 6
	case core.ShipWeaponModifierECCM:
		return 7
	case core.ShipWeaponModifierArmored:
		return 8
	case core.ShipWeaponModifierFast:
		return 9
	case core.ShipWeaponModifierEmissionsGuidance:
		return 10
	case core.ShipWeaponModifierOverloaded:
		return 11
	default:
		return -1
	}
}

func validateTacticalWeaponSnapshot(w TacticalWeaponSpec) error {
	if w.WeaponID == "" || w.Count <= 0 || w.MinDamage < 0 || w.MaxDamage < w.MinDamage {
		return fmt.Errorf("invalid tactical weapon snapshot %q", w.WeaponID)
	}
	if w.MinDamage == 0 && (w.MaxDamage != 0 || (w.TacticalSupport != "metadata_only" && w.TacticalSupport != "fighter")) {
		return fmt.Errorf("tactical weapon %q has invalid zero-damage metadata", w.WeaponID)
	}
	switch canonicalTacticalWeaponMountType(w.MountType) {
	case core.ShipWeaponMountStandard, core.ShipWeaponMountHeavy, core.ShipWeaponMountPointDefense:
	default:
		return fmt.Errorf("tactical weapon %q has invalid mount type %q", w.WeaponID, w.MountType)
	}
	previousRank := -1
	for _, id := range w.ModifierIDs {
		rank := tacticalWeaponModifierRank(id)
		if rank < 0 || rank <= previousRank {
			return fmt.Errorf("tactical weapon %q has invalid/noncanonical modifier %q", w.WeaponID, id)
		}
		previousRank = rank
	}
	switch w.TacticalSupport {
	case "", "baseline", "beam", "ordnance", "bombardment_only", "metadata_only", "fighter":
	default:
		return fmt.Errorf("tactical weapon %q has invalid support mode %q", w.WeaponID, w.TacticalSupport)
	}
	family := tacticalWeaponFamily(w)
	if w.TacticalSupport == "fighter" {
		if family != "fighter" || w.CraftPerUnit != 4 || w.MinDamage != 0 || w.MaxDamage != 0 {
			return fmt.Errorf("tactical weapon %q has invalid Fighter support metadata", w.WeaponID)
		}
	} else if w.CraftPerUnit != 0 {
		return fmt.Errorf("tactical weapon %q carries craft_per_unit outside Fighter support", w.WeaponID)
	}
	if w.TacticalSupport == "ordnance" && family != "missile" && family != "torpedo" {
		return fmt.Errorf("tactical weapon %q family %q cannot use ordnance support", w.WeaponID, family)
	}
	if w.TacticalSupport == "bombardment_only" && family != "bomb" {
		return fmt.Errorf("tactical weapon %q family %q cannot use bombardment-only support", w.WeaponID, family)
	}
	if w.TacticalSupport == "" && w.WeaponID != "laser_cannon" {
		return fmt.Errorf("tactical weapon %q is missing support metadata", w.WeaponID)
	}
	return nil
}

func tacticalBeamConfigured(w TacticalWeaponSpec) (beam.ConfiguredPreview, error) {
	if tacticalWeaponFamily(w) != "beam" {
		return beam.ConfiguredPreview{}, fmt.Errorf("weapon %q is not a Beam weapon", w.WeaponID)
	}
	return beam.DeriveConfigured(w.WeaponID, w.MinDamage, w.MaxDamage, canonicalTacticalWeaponMountType(w.MountType), w.ModifierIDs, w.InnateEffectIDs)
}

func tacticalBeamFireSupported(w TacticalWeaponSpec) bool {
	if w.Count <= 0 || tacticalWeaponFamily(w) != "beam" {
		return false
	}
	if w.TacticalSupport == "beam" {
		return true
	}
	return w.WeaponID == "laser_cannon" &&
		canonicalTacticalWeaponMountType(w.MountType) == core.ShipWeaponMountStandard &&
		len(w.ModifierIDs) == 0 &&
		w.MinDamage == 1 &&
		w.MaxDamage == 4 &&
		(w.TacticalSupport == "" || w.TacticalSupport == "baseline")
}

type tacticalHullValidationRule struct {
	BaseArmor                  int
	Structure                  int
	BaseCombatSpeed            int
	ShieldFactor               int
	MarineMax                  int
	InternalEngineMax          int
	InternalSecondaryMax       int
	InternalShieldGeneratorMax int
}

func tacticalHullValidationRuleFor(id string) (tacticalHullValidationRule, bool) {
	switch id {
	case "frigate":
		return tacticalHullValidationRule{BaseArmor: 4, Structure: 4, BaseCombatSpeed: 10, ShieldFactor: 1, MarineMax: 5, InternalEngineMax: 2, InternalSecondaryMax: 1, InternalShieldGeneratorMax: 1}, true
	case "destroyer":
		return tacticalHullValidationRule{BaseArmor: 10, Structure: 10, BaseCombatSpeed: 8, ShieldFactor: 2, MarineMax: 8, InternalEngineMax: 5, InternalSecondaryMax: 2, InternalShieldGeneratorMax: 2}, true
	case "cruiser":
		return tacticalHullValidationRule{BaseArmor: 30, Structure: 30, BaseCombatSpeed: 6, ShieldFactor: 3, MarineMax: 12, InternalEngineMax: 10, InternalSecondaryMax: 5, InternalShieldGeneratorMax: 6}, true
	case "battleship":
		return tacticalHullValidationRule{BaseArmor: 50, Structure: 50, BaseCombatSpeed: 5, ShieldFactor: 4, MarineMax: 20, InternalEngineMax: 15, InternalSecondaryMax: 7, InternalShieldGeneratorMax: 7}, true
	case "titan":
		return tacticalHullValidationRule{BaseArmor: 80, Structure: 80, BaseCombatSpeed: 4, ShieldFactor: 5, MarineMax: 30, InternalEngineMax: 20, InternalSecondaryMax: 10, InternalShieldGeneratorMax: 10}, true
	case "doom_star":
		return tacticalHullValidationRule{BaseArmor: 150, Structure: 150, BaseCombatSpeed: 3, ShieldFactor: 7, MarineMax: 50, InternalEngineMax: 40, InternalSecondaryMax: 20, InternalShieldGeneratorMax: 20}, true
	default:
		return tacticalHullValidationRule{}, false
	}
}

func tacticalArmorBonusPercentFor(id string) (int, bool) {
	switch id {
	case "titanium_armor":
		return 0, true
	case "tritanium_armor":
		return 100, true
	case "zortrium_armor":
		return 300, true
	case "neutronium_armor":
		return 500, true
	case "adamantium_armor":
		return 700, true
	case "xentronium_armor":
		return 900, true
	default:
		return 0, false
	}
}

func tacticalDriveCombatSpeedBonusFor(id string) (int, bool) {
	switch id {
	case "nuclear_drive":
		return 0, true
	case "fusion_drive":
		return 2, true
	case "ion_drive":
		return 4, true
	case "anti_matter_drive":
		return 6, true
	case "hyper_drive":
		return 8, true
	case "interphased_drive":
		return 10, true
	default:
		return 0, false
	}
}

func tacticalComputerBeamOffenseFor(id string) (int, bool) {
	switch id {
	case "electronic_computer":
		return 25, true
	case "optronic_computer":
		return 50, true
	case "positronic_computer":
		return 75, true
	case "cybertronic_computer":
		return 100, true
	case "moleculartronic_computer":
		return 125, true
	default:
		return 0, false
	}
}

func tacticalShieldStrengthFor(id string) (int, bool) {
	switch id {
	case "":
		return 0, true
	case "class_i_shield":
		return 1, true
	case "class_iii_shield":
		return 3, true
	case "class_v_shield":
		return 5, true
	case "class_vii_shield":
		return 7, true
	case "class_x_shield":
		return 10, true
	default:
		return 0, false
	}
}

func tacticalFuelCellAccepted(id string) bool {
	switch id {
	case "standard_fuel_cells", "deuterium_fuel_cells", "iridium_fuel_cells", "urridium_fuel_cells", "thorium_fuel_cells":
		return true
	default:
		return false
	}
}

func validateTacticalSpecialSnapshot(s TacticalSpecialSpec) error {
	if s.SpecialID == "" {
		return fmt.Errorf("tactical special requires special_id")
	}
	switch s.TacticalSupport {
	case "metadata_only", "strategic_only", "turning_mode", "tactical_passive", "tactical_active":
		return nil
	default:
		return fmt.Errorf("tactical special %q has invalid support mode %q", s.SpecialID, s.TacticalSupport)
	}
}

func validateTacticalInternalSpec(in *TacticalInternalSpec) error {
	if in == nil {
		return nil
	}
	if in.EngineMax < 0 || in.SecondaryMax < 0 || in.ShieldGeneratorMax < 0 || len(in.Weapons) > 8 || len(in.Specials) > 8 {
		return fmt.Errorf("invalid tactical internal maxima")
	}
	previous := -1
	for _, item := range in.Weapons {
		if item.Slot < 0 || item.Slot > 7 || item.Slot <= previous || item.Max < 0 || item.UnitSize <= 0 {
			return fmt.Errorf("invalid tactical internal weapon pool")
		}
		previous = item.Slot
	}
	previous = -1
	for _, item := range in.Specials {
		if item.Slot < 0 || item.Slot > 7 || item.Slot <= previous || item.Max < 0 {
			return fmt.Errorf("invalid tactical internal special pool")
		}
		previous = item.Slot
	}
	return nil
}

func validateTacticalCombatantSnapshot(s TacticalShipSpec, rules TacticalRulesSnapshot) error {
	hull, ok := tacticalHullValidationRuleFor(s.HullID)
	if !ok {
		return fmt.Errorf("tactical ship %d has unknown hull %q", s.ShipID, s.HullID)
	}
	armorBonus, ok := tacticalArmorBonusPercentFor(s.ArmorID)
	if !ok {
		return fmt.Errorf("tactical ship %d has unknown armor %q", s.ShipID, s.ArmorID)
	}
	driveBonus, ok := tacticalDriveCombatSpeedBonusFor(s.WarpDriveID)
	if !ok {
		return fmt.Errorf("tactical ship %d has unknown drive %q", s.ShipID, s.WarpDriveID)
	}
	beamOffense, ok := tacticalComputerBeamOffenseFor(s.ComputerID)
	if !ok {
		return fmt.Errorf("tactical ship %d has unknown computer %q", s.ShipID, s.ComputerID)
	}
	shieldStrength, ok := tacticalShieldStrengthFor(s.ShieldID)
	if !ok {
		return fmt.Errorf("tactical ship %d has unknown shield %q", s.ShipID, s.ShieldID)
	}
	if !tacticalFuelCellAccepted(s.FuelCellID) {
		return fmt.Errorf("tactical ship %d has unknown fuel cell %q", s.ShipID, s.FuelCellID)
	}

	minimumSpeed := hull.BaseCombatSpeed + driveBonus
	if s.MinimumCombatSpeed != minimumSpeed {
		return fmt.Errorf("tactical ship %d minimum combat speed %d does not match derived %d", s.ShipID, s.MinimumCombatSpeed, minimumSpeed)
	}
	minCurrentSpeed, maxCurrentSpeed := minimumSpeed, minimumSpeed+10
	if hasTacticalSpecial(s, "augmented_engines") {
		if rules.Specials == nil {
			return fmt.Errorf("tactical ship %d Augmented Engines requires Special rules", s.ShipID)
		}
		minCurrentSpeed += rules.Specials.AugmentedEnginesSpeedBonus
		maxCurrentSpeed += rules.Specials.AugmentedEnginesSpeedBonus
	}
	if s.CurrentCombatSpeed < minCurrentSpeed || s.CurrentCombatSpeed > maxCurrentSpeed {
		return fmt.Errorf("tactical ship %d current combat speed %d is outside derived %d..%d", s.ShipID, s.CurrentCombatSpeed, minCurrentSpeed, maxCurrentSpeed)
	}
	if hasTacticalSpecial(s, "battle_scanner") {
		if rules.Specials == nil {
			return fmt.Errorf("tactical ship %d Battle Scanner requires Special rules", s.ShipID)
		}
		beamOffense += rules.Specials.BattleScannerOffenseBonus
	}
	if s.BeamOffense != beamOffense {
		return fmt.Errorf("tactical ship %d Beam Offense %d does not match derived %d", s.ShipID, s.BeamOffense, beamOffense)
	}
	if s.BeamDefense != s.CurrentCombatSpeed*5 {
		return fmt.Errorf("tactical ship %d Beam Defense %d does not match derived %d", s.ShipID, s.BeamDefense, s.CurrentCombatSpeed*5)
	}
	if s.MissileDefense < 0 || s.MissileScannerReduction < 0 {
		return fmt.Errorf("tactical ship %d Missile Defense/scanner values must be non-negative", s.ShipID)
	}
	armorMax := hull.BaseArmor + (hull.BaseArmor*armorBonus)/100
	if hasTacticalSpecial(s, "heavy_armor") {
		armorMax *= 3
	}
	structureMax := hull.Structure
	if hasTacticalSpecial(s, "reinforced_hull") {
		structureMax *= 3
	}
	if s.ArmorMax != armorMax || s.StructureMax != structureMax {
		return fmt.Errorf("tactical ship %d durability %d/%d does not match derived %d/%d", s.ShipID, s.ArmorMax, s.StructureMax, armorMax, structureMax)
	}
	marineMax := hull.MarineMax
	for _, special := range s.Specials {
		if special.SpecialID == "troop_pods" {
			marineMax *= 2
			break
		}
	}
	if s.MarineMax != marineMax {
		return fmt.Errorf("tactical ship %d MarineMax %d does not match derived %d", s.ShipID, s.MarineMax, marineMax)
	}
	shieldArcMax := shieldStrength * 5 * hull.ShieldFactor
	if hasTacticalSpecial(s, "multi_phased_shields") && shieldArcMax > 0 {
		if rules.Specials == nil || rules.Specials.MultiPhasedShieldDenominator <= 0 {
			return fmt.Errorf("tactical ship %d Multi-Phased Shields requires valid Special rules", s.ShipID)
		}
		shieldArcMax = (shieldArcMax * rules.Specials.MultiPhasedShieldNumerator) / rules.Specials.MultiPhasedShieldDenominator
	}
	if s.ShieldStrength != shieldStrength || s.ShieldArcMax != shieldArcMax {
		return fmt.Errorf("tactical ship %d shield strength/arc %d/%d does not match derived %d/%d", s.ShipID, s.ShieldStrength, s.ShieldArcMax, shieldStrength, shieldArcMax)
	}

	switch s.TurningMode {
	case "", TacticalTurningNormal, TacticalTurningInertialStabilizer, TacticalTurningInertialNullifier:
	default:
		return fmt.Errorf("tactical ship %d turning mode %q is unsupported", s.ShipID, s.TurningMode)
	}
	if len(s.Weapons) > 8 {
		return fmt.Errorf("tactical ship %d has %d weapon mounts, maximum is 8", s.ShipID, len(s.Weapons))
	}
	previousSlot := -1
	for i, w := range s.Weapons {
		if w.Slot < 0 || w.Slot > 7 || (i > 0 && w.Slot <= previousSlot) {
			return fmt.Errorf("tactical ship %d weapon slots must be unique, ascending and within 0..7", s.ShipID)
		}
		if err := validateTacticalWeaponSnapshot(w); err != nil {
			return fmt.Errorf("tactical ship %d weapon slot %d: %w", s.ShipID, w.Slot, err)
		}
		previousSlot = w.Slot
	}
	if len(s.Specials) > 8 {
		return fmt.Errorf("tactical ship %d has %d special mounts, maximum is 8", s.ShipID, len(s.Specials))
	}
	previousSpecialSlot := -1
	seenSpecialIDs := make(map[string]struct{}, len(s.Specials))
	for i, special := range s.Specials {
		if special.Slot < 0 || special.Slot > 7 || (i > 0 && special.Slot <= previousSpecialSlot) {
			return fmt.Errorf("tactical ship %d special slots must be unique, ascending and within 0..7", s.ShipID)
		}
		if _, ok := seenSpecialIDs[special.SpecialID]; ok {
			return fmt.Errorf("tactical ship %d has duplicate special %q", s.ShipID, special.SpecialID)
		}
		if err := validateTacticalSpecialSnapshot(special); err != nil {
			return fmt.Errorf("tactical ship %d special slot %d: %w", s.ShipID, special.Slot, err)
		}
		seenSpecialIDs[special.SpecialID] = struct{}{}
		previousSpecialSlot = special.Slot
	}
	if err := validateTacticalInternalSpec(s.Internal); err != nil {
		return fmt.Errorf("tactical ship %d internal snapshot: %w", s.ShipID, err)
	}
	if s.Internal == nil {
		return nil
	}
	expectedEngineMax := hull.InternalEngineMax
	if hasTacticalSpecial(s, "reinforced_hull") {
		expectedEngineMax *= 3
	}
	expectedShieldGeneratorMax := 0
	if s.ShieldID != "" {
		expectedShieldGeneratorMax = hull.InternalShieldGeneratorMax
	}
	if s.Internal.EngineMax != expectedEngineMax || s.Internal.SecondaryMax != hull.InternalSecondaryMax || s.Internal.ShieldGeneratorMax != expectedShieldGeneratorMax {
		return fmt.Errorf("tactical ship %d fixed internal maxima %d/%d/%d do not match derived %d/%d/%d", s.ShipID, s.Internal.EngineMax, s.Internal.SecondaryMax, s.Internal.ShieldGeneratorMax, expectedEngineMax, hull.InternalSecondaryMax, expectedShieldGeneratorMax)
	}
	if len(s.Internal.Weapons) != len(s.Weapons) {
		return fmt.Errorf("tactical ship %d internal weapon-pool count %d does not match mounts %d", s.ShipID, len(s.Internal.Weapons), len(s.Weapons))
	}
	for i, weapon := range s.Weapons {
		pool := s.Internal.Weapons[i]
		if pool.Slot != weapon.Slot || pool.Max != weapon.Count*pool.UnitSize {
			return fmt.Errorf("tactical ship %d internal weapon pool %d does not match weapon slot %d/count", s.ShipID, pool.Slot, weapon.Slot)
		}
	}
	wantSpecialPools := make(map[int]struct{})
	for _, special := range s.Specials {
		if !tacticalNonDamageableSpecial(special.SpecialID) {
			wantSpecialPools[special.Slot] = struct{}{}
		}
	}
	if len(s.Internal.Specials) != len(wantSpecialPools) {
		return fmt.Errorf("tactical ship %d internal Special-pool count %d does not match damageable installed Specials %d", s.ShipID, len(s.Internal.Specials), len(wantSpecialPools))
	}
	for _, pool := range s.Internal.Specials {
		if _, ok := wantSpecialPools[pool.Slot]; !ok || pool.Max <= 0 {
			return fmt.Errorf("tactical ship %d internal Special pool slot %d is not a damageable installed Special", s.ShipID, pool.Slot)
		}
	}
	return nil
}

func validateBaselineRules(r TacticalRulesSnapshot) error {
	if r.SchemaVersion != 1 || r.InitiativeBeamOffenseDivisor != 10 || r.RNGMultiplier != 0x41C64E6D || r.RNGIncrement != 0x3039 || r.BeamBaseHitThreshold != 40 || r.BeamMaxHitThreshold != 95 || r.BeamEffectiveRollCap != 100 {
		return fmt.Errorf("unsupported Slice 07 tactical rule constants")
	}
	wantHit := []int{0, 0, -10, -20, -30, -40, -55, -70, -85}
	wantDamage := []int{0, 0, -10, -20, -30, -40, -50, -60, -65}
	if !equalInts(r.BeamToHitRangeModifiers, wantHit) || !equalInts(r.BeamDamageRangeModifiers, wantDamage) {
		return fmt.Errorf("unsupported Slice 07 Beam range tables")
	}
	if r.Ordnance != nil {
		if err := validateTacticalOrdnanceRulesSnapshot(*r.Ordnance); err != nil {
			return err
		}
	}
	if r.Fighter != nil {
		f := r.Fighter
		if f.MovementUnitsPerField != 20 || f.CraftPerBay != 4 ||
			f.FighterBaseSpeed != 10 || f.BomberBaseSpeed != 8 || f.HeavyFighterBaseSpeed != 8 ||
			f.FighterBaseHits != 2 || f.BomberBaseHits != 4 || f.HeavyFighterBaseHits != 5 ||
			f.FighterAttackStock != 4 || f.BomberAttackStock != 1 || f.HeavyFighterAttackStock != 2 ||
			f.WarpDriveSpeedStep != 2 || f.TransDimensionalSpeedBonus != 4 ||
			f.RecoveryDivisor != 4 || !equalInts(f.ArmorDurabilityMultiples, []int{1, 2, 4, 6, 8, 10}) {
			return fmt.Errorf("unsupported Slice 17.8 Fighter rule constants")
		}
	}
	if r.Boarding != nil {
		b := r.Boarding
		if b.PhysicalRange != 1 || b.TransporterRange != 4 || b.DamperTransporterDivisor != 3 ||
			b.SecurityStationsDefenseBonus != 20 || b.RaidInternalDamageThreshold != 20 ||
			b.RaidDefenderLossThreshold != 50 || b.RaidAttackerLossThreshold != -50 ||
			b.CaptureQuantumRandomBound != 10 || b.CaptureQuantumDestroyMax != 5 ||
			!equalInts(b.CrewQualityBonuses, []int{0, 5, 10, 15}) {
			return fmt.Errorf("unsupported Slice 17.8 Boarding rule constants")
		}
	}
	if r.Specials != nil {
		x := r.Specials
		if x.AugmentedEnginesSpeedBonus != 5 || x.BattleScannerOffenseBonus != 50 ||
			x.ECMJammerDefenseBonus != 70 || x.MultiWaveECMJammerDefenseBonus != 100 ||
			x.WideAreaJammerSelfDefenseBonus != 130 || x.WideAreaJammerFleetDefenseBonus != 70 ||
			x.CloakBeamDefenseBonus != 80 || x.CloakMissileForcedMissThreshold != 50 ||
			x.PhasingInitialLifetime != 10 || x.EnergyAbsorberDivisor != 4 || x.DamperDamageDivisor != 4 ||
			x.DisplacementOrdnanceMissThreshold != 30 || x.DisplacementBeamMissThreshold != 30 ||
			x.ShieldRechargeNormalFactor != 3 || x.ShieldRechargeCapacitorFactor != 10 ||
			x.HardShieldsAdditionalAbsorption != 3 || x.MultiPhasedShieldNumerator != 3 ||
			x.MultiPhasedShieldDenominator != 2 || x.LightningDestroyThreshold != 50 ||
			x.QuantumExplosionMultiplier != 3 || x.AutomatedRepairBasePercent != 20 ||
			x.SelfTeleportMaxRange != 20 {
			return fmt.Errorf("unsupported Slice 17.8 Special rule constants")
		}
	}
	return nil
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func newTacticalRuntime(spec TacticalSpec) (tacticalRuntime, error) {
	r := tacticalRuntime{
		state: TacticalState{
			Round:               1,
			NextCommandSequence: 1,
			RNGState:            spec.InitialRNGState,
			PassKind:            TacticalPassKindNormal,
			Ships:               make([]TacticalShipState, len(spec.Ships)),
		},
		nextEventSequence:    1,
		nextMovingActorOrder: 1,
		ordnance:             initialTacticalOrdnanceState(spec),
		fighters:             initialTacticalFighterState(spec),
	}
	for i, ship := range spec.Ships {
		state := TacticalShipState{ShipID: ship.ShipID, X: ship.X, Y: ship.Y, Facing: ship.Facing, CurrentCombatSpeed: ship.CurrentCombatSpeed, MovementCurrent: ship.CurrentCombatSpeed, MovementMax: ship.CurrentCombatSpeed, ArmorCurrent: ship.ArmorMax, ShieldArcCurrent: fullTacticalShieldArcs(ship.ShieldArcMax), MarineCurrent: ship.MarineMax, OwnerEmpireID: ship.EmpireID, ControllerSeatID: ship.SeatID, Weapons: make([]TacticalWeaponState, len(ship.Weapons)), Internal: initialTacticalInternalState(ship.Internal)}
		state.CloakState, state.PhasingTurnsRemaining = tacticalInitialCloakState(spec, ship, state)
		for j, weapon := range ship.Weapons {
			state.Weapons[j] = TacticalWeaponState{Slot: weapon.Slot, Ready: tacticalBeamFireSupported(weapon), CountCurrent: weapon.Count, SpecialReadiness: tacticalInitialSpecialReadiness(ship, state, weapon)}
		}
		r.state.Ships[i] = state
	}
	refreshTacticalOrdnanceReadiness(spec, &r)
	refreshTacticalFighterReadiness(spec, &r)
	r.recomputeInitiative(spec)
	if len(r.state.InitiativeOrder) == 0 {
		return tacticalRuntime{}, fmt.Errorf("tactical battle has no live ships")
	}
	r.state.ActiveShipID = r.state.InitiativeOrder[0]
	if err := r.appendEvent("round_started", 0, 0, map[string]any{"round": r.state.Round, "initiative_order": r.state.InitiativeOrder, "active_ship_id": r.state.ActiveShipID}); err != nil {
		return tacticalRuntime{}, err
	}
	return r, nil
}

func (r *tacticalRuntime) recomputeInitiative(spec TacticalSpec) {
	type item struct {
		id    core.ID
		value int
	}
	items := make([]item, 0, len(spec.Ships))
	for _, ship := range spec.Ships {
		state := r.shipState(ship.ShipID)
		if state == nil || state.Destroyed {
			continue
		}
		items = append(items, item{id: ship.ShipID, value: state.CurrentCombatSpeed + tacticalBeamOffenseForState(spec, ship, *state, r.legacy17_8Disabled)/spec.Rules.InitiativeBeamOffenseDivisor})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].value != items[j].value {
			return items[i].value > items[j].value
		}
		return items[i].id < items[j].id
	})
	r.state.InitiativeOrder = r.state.InitiativeOrder[:0]
	for _, item := range items {
		r.state.InitiativeOrder = append(r.state.InitiativeOrder, item.id)
	}
}

func (r *tacticalRuntime) appendEvent(kind string, seatID protocol.SeatID, commandSequence uint32, data any) error {
	var raw json.RawMessage
	if data != nil {
		encoded, err := json.Marshal(data)
		if err != nil {
			return err
		}
		raw = encoded
	}
	r.events = append(r.events, TacticalEvent{Sequence: r.nextEventSequence, Kind: kind, SeatID: seatID, CommandSequence: commandSequence, Data: raw})
	r.nextEventSequence++
	return nil
}

func (r *tacticalRuntime) shipState(id core.ID) *TacticalShipState {
	for i := range r.state.Ships {
		if r.state.Ships[i].ShipID == id {
			return &r.state.Ships[i]
		}
	}
	return nil
}

func tacticalShipSpec(spec TacticalSpec, id core.ID) *TacticalShipSpec {
	for i := range spec.Ships {
		if spec.Ships[i].ShipID == id {
			return &spec.Ships[i]
		}
	}
	return nil
}

func weaponSpec(ship TacticalShipSpec, slot int) *TacticalWeaponSpec {
	for i := range ship.Weapons {
		if ship.Weapons[i].Slot == slot {
			return &ship.Weapons[i]
		}
	}
	return nil
}

func weaponState(ship *TacticalShipState, slot int) *TacticalWeaponState {
	if ship == nil {
		return nil
	}
	for i := range ship.Weapons {
		if ship.Weapons[i].Slot == slot {
			return &ship.Weapons[i]
		}
	}
	return nil
}

func (s *Session) PrepareCommand(seatID protocol.SeatID, command protocol.Command) (*PreparedCommand, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.phase != PhaseActive {
		return nil, fmt.Errorf("cannot submit tactical command in phase %q", s.phase)
	}
	if s.spec.Tactical == nil || s.tactical == nil {
		if s.spec.TacticalUnsupportedReason != "" {
			return nil, fmt.Errorf("tactical combat unsupported: %s", s.spec.TacticalUnsupportedReason)
		}
		return nil, fmt.Errorf("battle %d has no tactical command surface", s.spec.ID)
	}
	if err := command.Validate(s.tactical.state.NextCommandSequence); err != nil {
		return nil, err
	}
	prepared := &PreparedCommand{baseRevision: s.tacticalRevision, runtime: cloneTacticalRuntime(*s.tactical)}
	var result *Result
	var err error
	switch command.Kind {
	case CommandMoveShip:
		var payload MoveShipPayload
		if err = decodeBattlePayload(command, &payload); err == nil {
			err = prepareMoveShip(s.spec, &prepared.runtime, seatID, command.Sequence, payload)
		}
	case CommandFireBeam:
		var payload FireBeamPayload
		if err = decodeBattlePayload(command, &payload); err == nil {
			result, err = prepareFireBeam(s.spec, &prepared.runtime, seatID, command.Sequence, payload)
		}
	case CommandFireOrdnance:
		var payload FireOrdnancePayload
		if err = decodeBattlePayload(command, &payload); err == nil {
			err = prepareFireOrdnance(s.spec, &prepared.runtime, seatID, command.Sequence, payload)
		}
	case CommandLaunchFighters:
		var payload LaunchFightersPayload
		if err = decodeBattlePayload(command, &payload); err == nil {
			err = prepareLaunchFighters(s.spec, &prepared.runtime, seatID, command.Sequence, payload)
		}

	case CommandBoardShip:
		var payload BoardShipPayload
		if err = decodeBattlePayload(command, &payload); err == nil {
			result, err = prepareBoardShip(s.spec, &prepared.runtime, seatID, command.Sequence, payload)
		}

	case CommandEndActivation:
		var payload EndActivationPayload
		if err = decodeBattlePayload(command, &payload); err == nil {
			err = prepareEndActivation(s.spec, &prepared.runtime, seatID, command.Sequence, payload)
		}
	case CommandWaitActivation:
		var payload WaitActivationPayload
		if err = decodeBattlePayload(command, &payload); err == nil {
			err = prepareWaitActivation(s.spec, &prepared.runtime, seatID, command.Sequence, payload)
		}
	case CommandRetreat:
		var payload RetreatPayload
		if err = decodeBattlePayload(command, &payload); err == nil {
			result, err = prepareRetreat(s.spec, &prepared.runtime, seatID, command.Sequence, payload)
		}
	default:
		err = fmt.Errorf("unsupported tactical command %q", command.Kind)
	}
	if err != nil {
		return nil, err
	}
	if result == nil {
		result, err = tacticalTerminalResultAfterMutation(s.spec, &prepared.runtime, seatID, command.Sequence)
		if err != nil {
			return nil, err
		}
	}
	if result == nil && (command.Kind == CommandMoveShip || command.Kind == CommandFireBeam || command.Kind == CommandFireOrdnance || command.Kind == CommandLaunchFighters || command.Kind == CommandBoardShip) {
		if err := autoCompleteExhaustedActiveShip(s.spec, &prepared.runtime, command.Sequence); err != nil {
			return nil, err
		}
		result, err = tacticalTerminalResultAfterMutation(s.spec, &prepared.runtime, seatID, command.Sequence)
		if err != nil {
			return nil, err
		}
	}
	prepared.runtime.state.NextCommandSequence++
	if result != nil {
		normalized, err := normalizeResult(s.spec, *result)
		if err != nil {
			return nil, err
		}
		prepared.result = &normalized
	}
	return prepared, nil
}
func tacticalTerminalResultAfterMutation(spec Spec, r *tacticalRuntime, seatID protocol.SeatID, commandSequence uint32) (*Result, error) {
	winner := remainingWinner(spec, r)
	if winner == 0 {
		return nil, nil
	}
	if err := r.appendEvent("winner_determined", seatID, commandSequence, map[string]any{"winner_seat": winner}); err != nil {
		return nil, err
	}
	return &Result{
		WinnerSeat:       winner,
		Outcome:          TacticalOutcomeVictory,
		DestroyedShipIDs: destroyedShipIDs(r),
	}, nil
}

func (s *Session) ResolveReactiveBeamPointDefense(request ReactiveBeamPointDefenseRequest) (ReactiveBeamPointDefenseResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhaseActive {
		return ReactiveBeamPointDefenseResult{}, fmt.Errorf("cannot resolve reactive Beam Point Defense in phase %q", s.phase)
	}
	if s.spec.Tactical == nil || s.tactical == nil {
		return ReactiveBeamPointDefenseResult{}, fmt.Errorf("battle %d has no tactical runtime", s.spec.ID)
	}
	runtime := cloneTacticalRuntime(*s.tactical)
	result, err := resolveReactiveBeamPointDefense(s.spec, &runtime, request)
	if err != nil {
		return ReactiveBeamPointDefenseResult{}, err
	}
	s.tactical = &runtime
	s.tacticalRevision++
	return result, nil
}

func resolveReactiveBeamPointDefense(spec Spec, r *tacticalRuntime, request ReactiveBeamPointDefenseRequest) (ReactiveBeamPointDefenseResult, error) {
	if request.ShipID == 0 {
		return ReactiveBeamPointDefenseResult{}, fmt.Errorf("reactive Beam Point Defense ship id must be non-zero")
	}
	source := tacticalShipSpec(*spec.Tactical, request.ShipID)
	if source == nil {
		return ReactiveBeamPointDefenseResult{}, fmt.Errorf("reactive Beam Point Defense ship %d is missing from spec", request.ShipID)
	}
	sourceState := r.shipState(request.ShipID)
	if sourceState == nil || sourceState.Destroyed {
		return ReactiveBeamPointDefenseResult{}, fmt.Errorf("reactive Beam Point Defense ship %d is unavailable", request.ShipID)
	}
	weapon := weaponSpec(*source, request.WeaponSlot)
	if weapon == nil {
		return ReactiveBeamPointDefenseResult{}, fmt.Errorf("reactive Beam Point Defense weapon slot %d is missing", request.WeaponSlot)
	}
	if !tacticalBeamFireSupported(*weapon) {
		return ReactiveBeamPointDefenseResult{}, fmt.Errorf("weapon slot %d is outside the executable Beam firing contract", request.WeaponSlot)
	}
	if canonicalTacticalWeaponMountType(weapon.MountType) != core.ShipWeaponMountPointDefense {
		return ReactiveBeamPointDefenseResult{}, fmt.Errorf("weapon slot %d is not a Point Defense Beam mount", request.WeaponSlot)
	}
	configured, err := tacticalBeamConfigured(*weapon)
	if err != nil {
		return ReactiveBeamPointDefenseResult{}, err
	}
	targetPoint := TacticalShipState{X: request.TargetX, Y: request.TargetY}
	if !tacticalBeamTargetInArc(*sourceState, targetPoint, weapon.FiringArcID) {
		return ReactiveBeamPointDefenseResult{}, fmt.Errorf("reactive target coordinate %d,%d is outside weapon slot %d firing arc %q", request.TargetX, request.TargetY, request.WeaponSlot, weapon.FiringArcID)
	}
	geometricRangeIndex := tacticalRangeIndex(sourceState.X, sourceState.Y, request.TargetX, request.TargetY)
	rangeIndex := tacticalBeamEffectiveRangeIndex(geometricRangeIndex, len(spec.Tactical.Rules.BeamToHitRangeModifiers), configured.RangeProfile)
	if rangeIndex < 0 || rangeIndex >= len(spec.Tactical.Rules.BeamToHitRangeModifiers) {
		return ReactiveBeamPointDefenseResult{}, fmt.Errorf("reactive Beam range index %d is outside the evidenced table", rangeIndex)
	}
	if rangeIndex >= len(spec.Tactical.Rules.BeamDamageRangeModifiers) {
		return ReactiveBeamPointDefenseResult{}, fmt.Errorf("reactive Beam damage range index %d is outside the evidenced table", rangeIndex)
	}

	readinessBefore := append([]TacticalWeaponState(nil), sourceState.Weapons...)
	defer func() {
		sourceState.Weapons = readinessBefore
	}()

	rangeHitModifier := spec.Tactical.Rules.BeamToHitRangeModifiers[rangeIndex]
	if tacticalBeamHasEffect(configured, "doubled_range_to_hit") {
		rangeHitModifier *= 2
	}
	threshold := spec.Tactical.Rules.BeamBaseHitThreshold - rangeHitModifier
	if threshold > spec.Tactical.Rules.BeamMaxHitThreshold {
		threshold = spec.Tactical.Rules.BeamMaxHitThreshold
	}
	dependencies := neutralTacticalBeamDependencyContext()
	scaledMinDamage, scaledMaxDamage, damagePercent := beam.EffectiveDamageBounds(
		configured,
		spec.Tactical.Rules.BeamDamageRangeModifiers[rangeIndex],
		dependencies.PlayerLeaderDamagePercent,
		dependencies.HighEnergyFocusBonus,
	)
	attackDifferential := tacticalBeamOffenseForState(*spec.Tactical, *source, *sourceState, r.legacy17_8Disabled) - request.TargetBeamDefense + configured.StaticToHitAdjustment
	alwaysHit := tacticalBeamHasEffect(configured, "always_hit")
	liveWeaponCount := tacticalCurrentWeaponCount(*source, *sourceState, weapon.Slot)
	totalShots := liveWeaponCount * configured.ShotsPerWeapon
	result := ReactiveBeamPointDefenseResult{
		ShipID: request.ShipID, WeaponSlot: request.WeaponSlot, WeaponID: weapon.WeaponID,
		MountType:       canonicalTacticalWeaponMountType(weapon.MountType),
		ModifierIDs:     append([]core.ShipWeaponModifierID(nil), weapon.ModifierIDs...),
		InnateEffectIDs: append([]string(nil), weapon.InnateEffectIDs...),
		TargetX:         request.TargetX, TargetY: request.TargetY,
		Shots: make([]ReactiveBeamPointDefenseShot, 0, totalShots),
	}
	hitCount := 0
	totalDamage := 0
	flatShotIndex := 0
	for physicalIndex := 0; physicalIndex < liveWeaponCount; physicalIndex++ {
		for burstIndex := 0; burstIndex < configured.ShotsPerWeapon; burstIndex++ {
			flatShotIndex++
			rawRoll, err := r.random(100, spec.Tactical.Rules)
			if err != nil {
				return ReactiveBeamPointDefenseResult{}, err
			}
			hitRoll := rawRoll
			if alwaysHit {
				hitRoll = 100
			}
			effectiveRoll := hitRoll
			if hitRoll > spec.Tactical.Rules.BeamMaxHitThreshold {
				effectiveRoll = spec.Tactical.Rules.BeamEffectiveRollCap
			} else {
				effectiveRoll += attackDifferential
				if effectiveRoll > spec.Tactical.Rules.BeamEffectiveRollCap {
					effectiveRoll = spec.Tactical.Rules.BeamEffectiveRollCap
				}
			}
			hit := effectiveRoll >= threshold
			damage := 0
			if hit {
				hitCount++
				damage = beam.InterpolateDamage(scaledMinDamage, scaledMaxDamage, effectiveRoll, threshold, spec.Tactical.Rules.BeamEffectiveRollCap)
				totalDamage += damage
			}
			shot := ReactiveBeamPointDefenseShot{
				PhysicalWeaponIndex: physicalIndex + 1, PhysicalWeaponCount: liveWeaponCount,
				BurstShotIndex: burstIndex + 1, BurstShotCount: configured.ShotsPerWeapon,
				ShotIndex: flatShotIndex, ShotCount: totalShots, RangeIndex: rangeIndex,
				RawHitRoll: rawRoll, EffectiveHitRoll: effectiveRoll, Threshold: threshold,
				Hit: hit, Damage: damage,
			}
			result.Shots = append(result.Shots, shot)
			if err := r.appendEvent("beam_fired", source.SeatID, 0, map[string]any{
				"ship_id": source.ShipID, "target_ship_id": core.ID(0), "target_x": request.TargetX, "target_y": request.TargetY,
				"weapon_slot": request.WeaponSlot, "weapon_id": weapon.WeaponID, "mount_type": canonicalTacticalWeaponMountType(weapon.MountType),
				"modifier_ids":          append([]core.ShipWeaponModifierID(nil), weapon.ModifierIDs...),
				"physical_weapon_index": physicalIndex + 1, "physical_weapon_count": liveWeaponCount,
				"burst_shot_index": burstIndex + 1, "burst_shot_count": configured.ShotsPerWeapon,
				"shot_index": flatShotIndex, "shot_count": totalShots,
				"range_index": rangeIndex, "raw_hit_roll": rawRoll, "effective_hit_roll": effectiveRoll,
				"threshold": threshold, "hit": hit, "damage": damage,
				"damage_percent": damagePercent, "scaled_min_damage": scaledMinDamage, "scaled_max_damage": scaledMaxDamage,
				"reactive_point_defense": true,
			}); err != nil {
				return ReactiveBeamPointDefenseResult{}, err
			}
		}
	}
	if totalShots > 1 {
		if err := r.appendEvent("beam_volley_resolved", source.SeatID, 0, map[string]any{
			"ship_id": source.ShipID, "target_ship_id": core.ID(0), "target_x": request.TargetX, "target_y": request.TargetY,
			"weapon_slot": request.WeaponSlot, "weapon_id": weapon.WeaponID, "mount_type": canonicalTacticalWeaponMountType(weapon.MountType),
			"modifier_ids":          append([]core.ShipWeaponModifierID(nil), weapon.ModifierIDs...),
			"physical_weapon_count": liveWeaponCount, "burst_shot_count": configured.ShotsPerWeapon,
			"shot_count": totalShots, "hit_count": hitCount, "total_damage": totalDamage,
			"reactive_point_defense": true,
		}); err != nil {
			return ReactiveBeamPointDefenseResult{}, err
		}
	}
	return result, nil
}

func decodeBattlePayload(command protocol.Command, dst any) error {
	decoder := json.NewDecoder(bytes.NewReader(command.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("decode %s payload: %w", command.Kind, err)
	}
	if decoder.More() {
		return fmt.Errorf("decode %s payload: trailing JSON values", command.Kind)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return fmt.Errorf("decode %s payload: trailing JSON value", command.Kind)
	}
	return nil
}

func prepareRetreat(spec Spec, r *tacticalRuntime, seatID protocol.SeatID, commandSequence uint32, payload RetreatPayload) (*Result, error) {
	activeSpec := tacticalShipSpec(*spec.Tactical, r.state.ActiveShipID)
	activeState := r.shipState(r.state.ActiveShipID)
	if activeSpec == nil || activeState == nil {
		return nil, fmt.Errorf("active tactical ship %d is missing", r.state.ActiveShipID)
	}
	if seatID == 0 || tacticalControllerSeat(*activeSpec, activeState) != seatID {
		return nil, fmt.Errorf("seat %d does not control active ship %d", seatID, r.state.ActiveShipID)
	}
	if payload.ShipID != r.state.ActiveShipID {
		return nil, fmt.Errorf("retreat ship %d is not active ship %d", payload.ShipID, r.state.ActiveShipID)
	}
	if activeState.Destroyed || activeState.ActivationComplete {
		return nil, fmt.Errorf("active ship %d cannot retreat", payload.ShipID)
	}
	activeOwner := tacticalOwnerEmpire(*activeSpec, activeState)
	if tacticalRetreatBlockedByWarpDissipator(*spec.Tactical, r, activeOwner) {
		return nil, fmt.Errorf("retreat is blocked by an enemy Warp Dissipator")
	}
	winner := protocol.SeatID(0)
	for _, participant := range spec.Participants {
		if participant != seatID {
			winner = participant
			break
		}
	}
	if winner == 0 {
		return nil, fmt.Errorf("battle has no opposing participant for retreat")
	}
	if err := r.appendEvent("side_retreated", seatID, commandSequence, map[string]any{"seat_id": seatID, "active_ship_id": payload.ShipID, "winner_seat": winner}); err != nil {
		return nil, err
	}
	return &Result{WinnerSeat: winner, Outcome: TacticalOutcomeRetreat}, nil
}
func prepareMoveShip(spec Spec, r *tacticalRuntime, seatID protocol.SeatID, commandSequence uint32, payload MoveShipPayload) error {
	activeSpec := tacticalShipSpec(*spec.Tactical, r.state.ActiveShipID)
	activeState := r.shipState(r.state.ActiveShipID)
	if activeSpec == nil || activeState == nil {
		return fmt.Errorf("active tactical ship %d is missing", r.state.ActiveShipID)
	}
	if seatID == 0 || tacticalControllerSeat(*activeSpec, activeState) != seatID {
		return fmt.Errorf("seat %d does not control active ship %d", seatID, r.state.ActiveShipID)
	}
	if payload.ShipID != r.state.ActiveShipID {
		return fmt.Errorf("move ship %d is not active ship %d", payload.ShipID, r.state.ActiveShipID)
	}
	if activeState.Destroyed || activeState.ActivationComplete {
		return fmt.Errorf("active ship %d cannot move", payload.ShipID)
	}
	option, err := tacticalMoveOption(*spec.Tactical, r, *activeSpec, *activeState, payload.X, payload.Y)
	if err != nil {
		return err
	}
	fromX, fromY, fromFacing := activeState.X, activeState.Y, activeState.Facing
	if option.MovementKind == TacticalMovementKindTeleport {
		activeState.TeleporterUsed = true
	}
	activeState.X, activeState.Y, activeState.Facing, activeState.MovementCurrent = option.X, option.Y, option.ResultingFacing, option.MovementRemainingAfter
	return r.appendEvent("ship_moved", seatID, commandSequence, map[string]any{"ship_id": payload.ShipID, "from_x": fromX, "from_y": fromY, "to_x": option.X, "to_y": option.Y, "facing_before": fromFacing, "facing_after": option.ResultingFacing, "move_cost": option.MoveCost, "movement_remaining": option.MovementRemainingAfter, "movement_kind": option.MovementKind})
}

func tacticalMoveOption(spec TacticalSpec, r *tacticalRuntime, shipSpec TacticalShipSpec, shipState TacticalShipState, x, y int) (TacticalMoveOption, error) {
	if x == shipState.X && y == shipState.Y {
		return TacticalMoveOption{}, fmt.Errorf("same-coordinate Tactical move is not a rotate-only command")
	}
	if absInt(x) > TacticalCoordinateSafetyLimit || absInt(y) > TacticalCoordinateSafetyLimit {
		return TacticalMoveOption{}, fmt.Errorf("Tactical destination (%d,%d) exceeds technical coordinate safety envelope", x, y)
	}
	normalErr := fmt.Errorf("Tactical destination (%d,%d) exceeds ship %d movement budget", x, y, shipState.ShipID)
	if absInt(x-shipState.X) <= shipState.MovementCurrent && absInt(y-shipState.Y) <= shipState.MovementCurrent {
		// Preserve the pre-17.8 normal-move rejection order: an out-of-budget
		// destination remains a budget error, while an in-budget occupied cell
		// is rejected as occupied. Teleport owns its own occupancy check below.
		for _, other := range r.state.Ships {
			if other.ShipID != shipState.ShipID && !other.Destroyed && other.X == x && other.Y == y {
				return TacticalMoveOption{}, fmt.Errorf("Tactical destination (%d,%d) is occupied by ship %d", x, y, other.ShipID)
			}
		}
		cost, facing := tacticalMoveCost(shipState.X, shipState.Y, shipState.Facing, x, y, turnCostPerFacing(shipSpec))
		if cost <= shipState.MovementCurrent {
			return TacticalMoveOption{MovementKind: TacticalMovementKindNormal, X: x, Y: y, MoveCost: cost, ResultingFacing: facing, MovementRemainingAfter: shipState.MovementCurrent - cost}, nil
		}
		normalErr = fmt.Errorf("Tactical move costs %d movement points but ship %d has %d remaining", cost, shipState.ShipID, shipState.MovementCurrent)
	}
	if option, err := tacticalTeleportMoveOption(spec, r, shipSpec, shipState, x, y); err == nil {
		return option, nil
	}
	return TacticalMoveOption{}, normalErr
}

func legalMoves(spec TacticalSpec, r *tacticalRuntime) []TacticalMoveOption {
	state := r.shipState(r.state.ActiveShipID)
	ship := tacticalShipSpec(spec, r.state.ActiveShipID)
	if state == nil || ship == nil || state.Destroyed || state.ActivationComplete {
		return nil
	}
	radius := state.MovementCurrent
	if tacticalSelfTeleportFunctional(spec, r, *ship, *state) && spec.Rules.Specials != nil && spec.Rules.Specials.SelfTeleportMaxRange > radius {
		radius = spec.Rules.Specials.SelfTeleportMaxRange * 3
	}
	if radius <= 0 {
		return nil
	}
	out := make([]TacticalMoveOption, 0, (radius*2+1)*(radius*2+1)/2)
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			option, err := tacticalMoveOption(spec, r, *ship, *state, state.X+dx, state.Y+dy)
			if err == nil {
				out = append(out, option)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].MovementKind != out[j].MovementKind {
			return out[i].MovementKind < out[j].MovementKind
		}
		if out[i].MoveCost != out[j].MoveCost {
			return out[i].MoveCost < out[j].MoveCost
		}
		if out[i].Y != out[j].Y {
			return out[i].Y < out[j].Y
		}
		return out[i].X < out[j].X
	})
	return out
}

func tacticalMoveCost(fromX, fromY, currentFacing, toX, toY, turnCost int) (int, int) {
	dx, dy := toX-fromX, toY-fromY
	if dx == 0 && dy == 0 {
		return 0, currentFacing
	}
	distanceSquared := int64(dx)*int64(dx) + int64(dy)*int64(dy)
	translation := 0
	for int64(translation)*int64(translation) < distanceSquared {
		translation++
	}
	resultingFacing := tacticalFacing(dx, dy)
	facingSteps := absInt(resultingFacing - currentFacing)
	if facingSteps > 8 {
		facingSteps = 16 - facingSteps
	}
	return translation + facingSteps*turnCost, resultingFacing
}

func tacticalFacing(dx, dy int) int {
	angle := math.Atan2(float64(dy), float64(dx))
	if angle < 0 {
		angle += 2 * math.Pi
	}
	return int(math.Floor(angle/(2*math.Pi)*16+0.5)) % 16
}
func turnCostPerFacing(ship TacticalShipSpec) int {
	switch ship.TurningMode {
	case TacticalTurningInertialStabilizer:
		return 1
	case TacticalTurningInertialNullifier:
		return 0
	default:
		return 2
	}
}

type tacticalDamageApplication struct {
	IncomingDamage       int
	PostSpecialDamage    int
	EnergyAbsorbed       int
	DamperReducedDamage  int
	ShieldArc            int
	ShieldBefore         int
	ShieldAfter          int
	PostBlockDamage      int
	ShieldAbsorbed       int
	LowerLayerOverflow   int
	ArmorBefore          int
	ArmorAfter           int
	StructureBefore      int
	StructureAfter       int
	ExtraStructureDamage int
	MarineBefore         int
	MarineAfter          int
	MarineKilled         int
	Layer                string
}

type tacticalBeamDependencyContext struct {
	PlayerLeaderDamagePercent int
	HighEnergyFocusBonus      int
	HardShields               bool
	ArmorBypassNeutralized    bool
	MarineKillSuppressed      bool
	TargetMonsterAntaranBand  bool
}

func neutralTacticalBeamDependencyContext() tacticalBeamDependencyContext {
	return tacticalBeamDependencyContext{PlayerLeaderDamagePercent: 100}
}

type tacticalDamageOptions struct {
	ShieldPiercing        bool
	ArmorFirstBypass      bool
	ExtraStructureDamage  bool
	MarineKill            bool
	MarineKillSuppressed  bool
	ShieldArcOverride     int
	HasShieldArcOverride  bool
	ShieldAdditionalBlock int
}

func fullTacticalShieldArcs(max int) [4]int {
	if max < 0 {
		max = 0
	}
	return [4]int{max, max, max, max}
}

func tacticalShieldImpactArc(target, source TacticalShipState) int {
	bearing := math.Atan2(float64(source.Y-target.Y), float64(source.X-target.X)) * 180 / math.Pi
	if bearing < 0 {
		bearing += 360
	}
	relative := math.Mod(bearing-float64(target.Facing)*22.5+360, 360)
	switch {
	case relative <= 45 || relative > 315:
		return 0
	case relative <= 135:
		return 3
	case relative <= 225:
		return 2
	default:
		return 1
	}
}

func applyTacticalDamage(targetSpec TacticalShipSpec, targetState *TacticalShipState, sourceState TacticalShipState, incomingDamage int) tacticalDamageApplication {
	return applyTacticalDamageWithOptions(targetSpec, targetState, sourceState, incomingDamage, tacticalDamageOptions{})
}

func applyTacticalDamageWithOptions(targetSpec TacticalShipSpec, targetState *TacticalShipState, sourceState TacticalShipState, incomingDamage int, options tacticalDamageOptions) tacticalDamageApplication {
	shieldArc := tacticalShieldImpactArc(*targetState, sourceState)
	if options.HasShieldArcOverride {
		shieldArc = options.ShieldArcOverride
	}
	result := tacticalDamageApplication{
		IncomingDamage:    incomingDamage,
		PostSpecialDamage: incomingDamage,
		ShieldArc:         shieldArc,
		ArmorBefore:       targetState.ArmorCurrent,
		ArmorAfter:        targetState.ArmorCurrent,
		StructureBefore:   targetState.StructureDamage,
		StructureAfter:    targetState.StructureDamage,
		MarineBefore:      targetState.MarineCurrent,
		MarineAfter:       targetState.MarineCurrent,
	}
	result.ShieldBefore = targetState.ShieldArcCurrent[result.ShieldArc]
	result.ShieldAfter = result.ShieldBefore

	if options.ShieldPiercing {
		result.PostBlockDamage = incomingDamage
	} else {
		shieldBlock := targetSpec.ShieldStrength
		if targetState.ShieldArcCurrent[result.ShieldArc] > 0 {
			shieldBlock += options.ShieldAdditionalBlock
		}
		result.PostBlockDamage = incomingDamage - shieldBlock
		if result.PostBlockDamage < 0 {
			result.PostBlockDamage = 0
		}
	}
	remainingDamage := result.PostBlockDamage
	if !options.ShieldPiercing && remainingDamage > 0 && targetState.ShieldArcCurrent[result.ShieldArc] > 0 {
		result.ShieldAbsorbed = remainingDamage
		if result.ShieldAbsorbed > targetState.ShieldArcCurrent[result.ShieldArc] {
			result.ShieldAbsorbed = targetState.ShieldArcCurrent[result.ShieldArc]
		}
		targetState.ShieldArcCurrent[result.ShieldArc] -= result.ShieldAbsorbed
		remainingDamage -= result.ShieldAbsorbed
		result.ShieldAfter = targetState.ShieldArcCurrent[result.ShieldArc]
	}
	result.LowerLayerOverflow = remainingDamage

	if !options.ArmorFirstBypass && remainingDamage > 0 && targetState.ArmorCurrent > 0 {
		absorbed := remainingDamage
		if absorbed > targetState.ArmorCurrent {
			absorbed = targetState.ArmorCurrent
		}
		targetState.ArmorCurrent -= absorbed
		remainingDamage -= absorbed
		result.ArmorAfter = targetState.ArmorCurrent
	}
	if remainingDamage > 0 {
		targetState.StructureDamage += remainingDamage
		if targetState.StructureDamage > targetSpec.StructureMax {
			targetState.StructureDamage = targetSpec.StructureMax
		}
		result.StructureAfter = targetState.StructureDamage
	}

	if options.ExtraStructureDamage && result.LowerLayerOverflow > 0 && targetState.StructureDamage < targetSpec.StructureMax {
		before := targetState.StructureDamage
		targetState.StructureDamage += result.LowerLayerOverflow
		if targetState.StructureDamage > targetSpec.StructureMax {
			targetState.StructureDamage = targetSpec.StructureMax
		}
		result.ExtraStructureDamage = targetState.StructureDamage - before
		result.StructureAfter = targetState.StructureDamage
	}

	if options.MarineKill && !options.MarineKillSuppressed && result.LowerLayerOverflow > 0 {
		killed := result.LowerLayerOverflow / 5
		if killed > targetState.MarineCurrent {
			killed = targetState.MarineCurrent
		}
		if killed > 0 {
			targetState.MarineCurrent -= killed
			result.MarineKilled = killed
			result.MarineAfter = targetState.MarineCurrent
		}
	}

	if targetState.StructureDamage >= targetSpec.StructureMax {
		targetState.Destroyed = true
	}

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
	case targetSpec.ShieldStrength > 0 && incomingDamage > 0 && result.PostBlockDamage == 0:
		result.Layer = "shield_block"
	case armorChanged && structureChanged:
		result.Layer = "armor_structure"
	case armorChanged:
		result.Layer = "armor"
	case structureChanged:
		result.Layer = "structure"
	case targetState.StructureDamage >= targetSpec.StructureMax && incomingDamage > 0:
		result.Layer = "structure_overkill"
	}
	return result
}

func rechargeTacticalShieldArcs(arcs *[4]int, max int) {
	rechargeTacticalShieldArcsWithBudget(arcs, max, max, 3)
}

func rechargeTacticalShields(spec TacticalSpec, r *tacticalRuntime) {
	if r == nil {
		return
	}
	for i := range r.state.Ships {
		state := &r.state.Ships[i]
		if state.Destroyed {
			continue
		}
		ship := tacticalShipSpec(spec, state.ShipID)
		if ship == nil || ship.ShieldArcMax <= 0 {
			continue
		}
		effectiveMax := tacticalShieldArcMaxForState(spec, *ship, *state, r.legacy17_8Disabled)
		if effectiveMax <= 0 {
			continue
		}
		factor := 3
		if spec.Rules.Specials != nil {
			factor = spec.Rules.Specials.ShieldRechargeNormalFactor
			if TacticalSpecialFunctional(*ship, *state, r.legacy17_8Disabled, "shield_capacitors") {
				factor = spec.Rules.Specials.ShieldRechargeCapacitorFactor
			}
		}
		rechargeTacticalShieldArcsWithBudget(&state.ShieldArcCurrent, effectiveMax, ship.ShieldArcMax, factor)
	}
}

func buildTacticalView(spec TacticalSpec, r tacticalRuntime) TacticalView {
	fighters := cloneTacticalFighterState(r.fighters)
	view := TacticalView{State: cloneTacticalState(r.state), Ordnance: cloneTacticalOrdnanceState(r.ordnance), Fighters: TacticalFighterView{BayStates: fighters.BayStates, Groups: fighters.Groups}, Events: cloneTacticalEvents(r.events), Ships: tacticalShipViews(spec, &r)}
	state := r.shipState(r.state.ActiveShipID)
	if state != nil && !state.Destroyed && !state.ActivationComplete {
		view.CanEndActivation = true
		if activeSpec := tacticalShipSpec(spec, r.state.ActiveShipID); activeSpec != nil {
			controller := tacticalControllerSeat(*activeSpec, state)
			view.WaitTargetShipIDs = waitTargetShipIDs(spec, &r, controller, r.state.ActiveShipID)
			view.CanWaitActivation = len(view.WaitTargetShipIDs) != 0
			view.CanRetreat = !tacticalRetreatBlockedByWarpDissipator(spec, &r, tacticalOwnerEmpire(*activeSpec, state))
		}
		view.LegalMoves = legalMoves(spec, &r)
		view.WeaponActions = tacticalWeaponActions(spec, &r)
		view.LegalFireActions = legalFireActions(view.WeaponActions)
		if activeSpec := tacticalShipSpec(spec, r.state.ActiveShipID); activeSpec != nil {
			view.BoardingActions = tacticalBoardingActions(spec, &r, activeSpec, state)
		}
	}
	return view
}

func tacticalShipViews(spec TacticalSpec, r *tacticalRuntime) []TacticalShipView {
	out := make([]TacticalShipView, 0, len(spec.Ships))
	for _, shipSpec := range spec.Ships {
		state := r.shipState(shipSpec.ShipID)
		if state == nil {
			continue
		}
		view := TacticalShipView{
			ShipID: shipSpec.ShipID, EmpireID: tacticalOwnerEmpire(shipSpec, state), SeatID: tacticalControllerSeat(shipSpec, state),
			HullID: shipSpec.HullID, WarpDriveID: shipSpec.WarpDriveID, ComputerID: shipSpec.ComputerID, ArmorID: shipSpec.ArmorID, ShieldID: shipSpec.ShieldID,
			SourceDesignID: shipSpec.SourceDesignID, SourceDesignRevision: shipSpec.SourceDesignRevision, StrategicPictureID: shipSpec.StrategicPictureID,
			SourceVisualRevision: shipSpec.SourceVisualRevision,
			X:                    state.X, Y: state.Y, Facing: state.Facing, CurrentCombatSpeed: state.CurrentCombatSpeed, MovementCurrent: state.MovementCurrent, MovementMax: state.MovementMax,
			ActivationComplete: state.ActivationComplete, ArmorCurrent: state.ArmorCurrent, ArmorMax: shipSpec.ArmorMax,
			MarineCurrent: state.MarineCurrent, MarineMax: shipSpec.MarineMax,
			TeleporterUsed: state.TeleporterUsed, AlreadyBoarded: state.AlreadyBoarded, BeingCaptured: state.BeingCaptured, PostCaptureState: state.PostCaptureState,
			CloakState: state.CloakState, PhasingTurnsRemaining: state.PhasingTurnsRemaining,
			ShieldStrength: shipSpec.ShieldStrength, ShieldArcCurrent: state.ShieldArcCurrent, ShieldArcMax: tacticalShieldArcMaxForState(spec, shipSpec, *state, r.legacy17_8Disabled),
			StructureCurrent: shipSpec.StructureMax - state.StructureDamage, StructureMax: shipSpec.StructureMax,
			BeamOffense: tacticalBeamOffenseForState(spec, shipSpec, *state, r.legacy17_8Disabled), BeamDefense: tacticalBeamDefenseForState(spec, shipSpec, *state, r.legacy17_8Disabled), MissileDefense: tacticalMissileDefenseForState(spec, r, shipSpec, *state), MissileScannerReduction: shipSpec.MissileScannerReduction, Destroyed: state.Destroyed,
			Weapons:  make([]TacticalWeaponView, 0, len(shipSpec.Weapons)),
			Internal: tacticalInternalView(shipSpec.Internal, state.Internal),
		}
		if shipSpec.VisualGenome != nil {
			genome := core.CloneShipVisualGenome(*shipSpec.VisualGenome)
			view.VisualGenome = &genome
		}
		if view.StructureCurrent < 0 {
			view.StructureCurrent = 0
		}
		for _, weaponSpec := range shipSpec.Weapons {
			ready := false
			specialReadiness := TacticalSpecialReadinessNone
			if weapon := weaponState(state, weaponSpec.Slot); weapon != nil {
				ready = weapon.Ready
				specialReadiness = weapon.SpecialReadiness
			}
			view.Weapons = append(view.Weapons, TacticalWeaponView{Slot: weaponSpec.Slot, WeaponID: weaponSpec.WeaponID, Count: tacticalCurrentWeaponCount(shipSpec, *state, weaponSpec.Slot), MountType: weaponSpec.MountType, ModifierIDs: append([]core.ShipWeaponModifierID(nil), weaponSpec.ModifierIDs...), Ammo: weaponSpec.Ammo, FiringArcID: weaponSpec.FiringArcID, MinDamage: weaponSpec.MinDamage, MaxDamage: weaponSpec.MaxDamage, InnateEffectIDs: append([]string(nil), weaponSpec.InnateEffectIDs...), TacticalSupport: weaponSpec.TacticalSupport, Ready: ready, SpecialReadiness: specialReadiness})
		}
		out = append(out, view)
	}
	return out
}

func tacticalWeaponFamily(w TacticalWeaponSpec) string {
	if w.WeaponFamily != "" {
		return w.WeaponFamily
	}
	if w.WeaponID == "laser_cannon" {
		return "beam"
	}
	return ""
}

func tacticalBeamHasEffect(config beam.ConfiguredPreview, effectID string) bool {
	for _, id := range config.EffectIDs {
		if id == effectID {
			return true
		}
	}
	return false
}

func tacticalBeamEffectiveRangeIndex(geometricIndex, tableLen int, profile beam.RangeProfile) int {
	index := geometricIndex
	switch profile {
	case beam.RangeProfileHeavy:
		index /= 2
	case beam.RangeProfilePointDefense:
		index *= 2
		if tableLen > 0 && index >= tableLen {
			index = tableLen - 1
		}
	}
	return index
}

func tacticalBeamTargetInArc(source, target TacticalShipState, arcID core.ShipWeaponFiringArcID) bool {
	if arcID == "" {
		// Preserve pre-17.2A tactical snapshots; newly projected Beam rows always persist a canonical arc.
		return true
	}
	if arcID == core.ShipWeaponFiringArc360 {
		return true
	}
	bearing := math.Atan2(float64(target.Y-source.Y), float64(target.X-source.X)) * 180 / math.Pi
	if bearing < 0 {
		bearing += 360
	}
	relative := math.Mod(bearing-float64(source.Facing)*22.5+360, 360)
	switch arcID {
	case core.ShipWeaponFiringArcForward:
		return relative <= 60 || relative >= 300
	case core.ShipWeaponFiringArcForwardExtended:
		return relative <= 120 || relative >= 240
	case core.ShipWeaponFiringArcBackExtended:
		return relative >= 60 && relative <= 300
	case core.ShipWeaponFiringArcRear:
		return relative >= 120 && relative <= 240
	default:
		return false
	}
}

func tacticalBeamTargets(spec TacticalSpec, r *tacticalRuntime, activeSpec *TacticalShipSpec, activeState *TacticalShipState, weapon TacticalWeaponSpec) []TacticalFireTarget {
	if activeSpec == nil || activeState == nil {
		return nil
	}
	configured, err := tacticalBeamConfigured(weapon)
	if err != nil {
		return nil
	}
	var targets []TacticalFireTarget
	sourceOwner := tacticalOwnerEmpire(*activeSpec, activeState)
	for _, targetSpec := range spec.Ships {
		targetState := r.shipState(targetSpec.ShipID)
		if !tacticalShipTargetable(targetState) {
			continue
		}
		if tacticalOwnerEmpire(targetSpec, targetState) == sourceOwner {
			continue
		}
		if !tacticalBeamTargetInArc(*activeState, *targetState, weapon.FiringArcID) {
			continue
		}
		geometricRangeIndex := tacticalRangeIndex(activeState.X, activeState.Y, targetState.X, targetState.Y)
		rangeIndex := tacticalBeamEffectiveRangeIndex(geometricRangeIndex, len(spec.Rules.BeamToHitRangeModifiers), configured.RangeProfile)
		if rangeIndex < 0 || rangeIndex >= len(spec.Rules.BeamToHitRangeModifiers) {
			continue
		}
		targets = append(targets, TacticalFireTarget{TargetShipID: targetSpec.ShipID, RangeIndex: rangeIndex})
	}
	targets = append(targets, tacticalBeamProjectileTargets(spec, r, activeSpec, activeState, weapon)...)
	targets = append(targets, tacticalBeamFighterTargets(spec, r, activeSpec, activeState, weapon)...)
	return targets
}

func tacticalWeaponActions(spec TacticalSpec, r *tacticalRuntime) []TacticalWeaponActionAvailability {
	activeSpec := tacticalShipSpec(spec, r.state.ActiveShipID)
	activeState := r.shipState(r.state.ActiveShipID)
	if activeSpec == nil || activeState == nil || activeState.Destroyed || activeState.ActivationComplete {
		return nil
	}
	out := make([]TacticalWeaponActionAvailability, 0, len(activeSpec.Weapons))
	for _, weaponSpec := range activeSpec.Weapons {
		if weaponSpec.TacticalSupport == "ordnance" {
			out = append(out, tacticalOrdnanceWeaponAction(spec, r, activeSpec, activeState, weaponSpec))
			continue
		}
		if weaponSpec.TacticalSupport == "fighter" {
			out = append(out, tacticalFighterWeaponAction(spec, r, activeSpec, activeState, weaponSpec))
			continue
		}
		action := TacticalWeaponActionAvailability{
			ShipID:       activeSpec.ShipID,
			WeaponSlot:   weaponSpec.Slot,
			WeaponID:     weaponSpec.WeaponID,
			WeaponFamily: tacticalWeaponFamily(weaponSpec),
		}
		if !tacticalBeamFireSupported(weaponSpec) {
			action.UnavailableReason = TacticalWeaponUnavailableConsumer
			out = append(out, action)
			continue
		}
		action.CommandKind = CommandFireBeam
		ready := weaponState(activeState, weaponSpec.Slot)
		if ready == nil || !ready.Ready {
			action.UnavailableReason = TacticalWeaponUnavailableNotReady
			out = append(out, action)
			continue
		}
		targets := tacticalBeamTargets(spec, r, activeSpec, activeState, weaponSpec)
		if len(targets) == 0 {
			action.UnavailableReason = TacticalWeaponUnavailableNoTargets
			out = append(out, action)
			continue
		}
		action.Available = true
		action.Targets = targets
		out = append(out, action)
	}
	return out
}

func legalFireActions(actions []TacticalWeaponActionAvailability) []TacticalFireAction {
	out := make([]TacticalFireAction, 0, len(actions))
	for _, action := range actions {
		if !action.Available || action.CommandKind != CommandFireBeam {
			continue
		}
		out = append(out, TacticalFireAction{
			ShipID:     action.ShipID,
			WeaponSlot: action.WeaponSlot,
			WeaponID:   action.WeaponID,
			Targets:    append([]TacticalFireTarget(nil), action.Targets...),
		})
	}
	return out
}

func prepareFireBeam(spec Spec, r *tacticalRuntime, seatID protocol.SeatID, commandSequence uint32, payload FireBeamPayload) (*Result, error) {
	active := tacticalShipSpec(*spec.Tactical, r.state.ActiveShipID)
	if active == nil {
		return nil, fmt.Errorf("active tactical ship %d is missing from spec", r.state.ActiveShipID)
	}
	activeState := r.shipState(active.ShipID)
	if activeState == nil || seatID == 0 || tacticalControllerSeat(*active, activeState) != seatID {
		return nil, fmt.Errorf("seat %d does not control active ship %d", seatID, r.state.ActiveShipID)
	}
	if payload.ShipID != r.state.ActiveShipID {
		return nil, fmt.Errorf("fire ship %d is not active ship %d", payload.ShipID, r.state.ActiveShipID)
	}
	targetCount := 0
	if payload.TargetShipID != 0 {
		targetCount++
	}
	if payload.TargetProjectileID != 0 {
		targetCount++
	}
	if payload.TargetFighterGroupID != 0 {
		targetCount++
	}
	if targetCount != 1 {
		return nil, fmt.Errorf("Beam fire requires exactly one ship, projectile or Fighter-group target")
	}
	if payload.TargetProjectileID != 0 {
		return prepareFireBeamAtProjectile(spec, r, seatID, commandSequence, payload)
	}
	if payload.TargetFighterGroupID != 0 {
		return prepareFireBeamAtFighterGroup(spec, r, seatID, commandSequence, payload)
	}
	targetSpec := tacticalShipSpec(*spec.Tactical, payload.TargetShipID)
	targetState := r.shipState(payload.TargetShipID)
	if targetSpec == nil || !tacticalShipTargetable(targetState) {
		return nil, fmt.Errorf("target ship %d is not a live tactical ship", payload.TargetShipID)
	}
	if tacticalOwnerEmpire(*targetSpec, targetState) == tacticalOwnerEmpire(*active, activeState) {
		return nil, fmt.Errorf("target ship %d is not an opponent", payload.TargetShipID)
	}

	weapon := weaponSpec(*active, payload.WeaponSlot)
	ready := weaponState(activeState, payload.WeaponSlot)
	if weapon == nil || ready == nil {
		return nil, fmt.Errorf("ship %d has no weapon slot %d", active.ShipID, payload.WeaponSlot)
	}
	if !ready.Ready {
		return nil, fmt.Errorf("ship %d weapon slot %d is not ready", active.ShipID, payload.WeaponSlot)
	}
	if !tacticalBeamFireSupported(*weapon) {
		return nil, fmt.Errorf("weapon slot %d is outside the executable Beam firing contract", payload.WeaponSlot)
	}
	if activeState == nil {
		return nil, fmt.Errorf("active tactical ship %d has no runtime position", active.ShipID)
	}

	configured, err := tacticalBeamConfigured(*weapon)
	if err != nil {
		return nil, err
	}
	if !tacticalBeamTargetInArc(*activeState, *targetState, weapon.FiringArcID) {
		return nil, fmt.Errorf("target ship %d is outside weapon slot %d firing arc %q", targetSpec.ShipID, payload.WeaponSlot, weapon.FiringArcID)
	}

	geometricRangeIndex := tacticalRangeIndex(activeState.X, activeState.Y, targetState.X, targetState.Y)
	rangeIndex := tacticalBeamEffectiveRangeIndex(geometricRangeIndex, len(spec.Tactical.Rules.BeamToHitRangeModifiers), configured.RangeProfile)
	if rangeIndex < 0 || rangeIndex >= len(spec.Tactical.Rules.BeamToHitRangeModifiers) {
		return nil, fmt.Errorf("Beam range index %d is outside the evidenced table", rangeIndex)
	}
	if rangeIndex >= len(spec.Tactical.Rules.BeamDamageRangeModifiers) {
		return nil, fmt.Errorf("Beam damage range index %d is outside the evidenced table", rangeIndex)
	}

	if destroyedByAbsorber, err := tacticalResolveEnergyAbsorberPreWeapon(*spec.Tactical, r, *active, activeState, *targetSpec, targetState, seatID, commandSequence); err != nil {
		return nil, err
	} else if destroyedByAbsorber {
		tacticalBreakCloakOnAction(activeState)
		return nil, nil
	}
	rangeHitModifier := spec.Tactical.Rules.BeamToHitRangeModifiers[rangeIndex]
	if tacticalBeamHasEffect(configured, "doubled_range_to_hit") {
		rangeHitModifier *= 2
	}
	threshold := spec.Tactical.Rules.BeamBaseHitThreshold - rangeHitModifier
	if threshold > spec.Tactical.Rules.BeamMaxHitThreshold {
		threshold = spec.Tactical.Rules.BeamMaxHitThreshold
	}

	dependencies := neutralTacticalBeamDependencyContext()
	dependencies.HardShields = TacticalSpecialFunctional(*targetSpec, *targetState, r.legacy17_8Disabled, "hard_shields")
	dependencies.MarineKillSuppressed = TacticalSpecialFunctional(*targetSpec, *targetState, r.legacy17_8Disabled, "damper_field")
	dependencies.ArmorBypassNeutralized = targetSpec.ArmorID == "xentronium_armor" || TacticalSpecialFunctional(*targetSpec, *targetState, r.legacy17_8Disabled, "heavy_armor")
	rangeDamageModifier := spec.Tactical.Rules.BeamDamageRangeModifiers[rangeIndex]
	scaledMinDamage, scaledMaxDamage, damagePercent := beam.EffectiveDamageBounds(
		configured,
		rangeDamageModifier,
		dependencies.PlayerLeaderDamagePercent,
		dependencies.HighEnergyFocusBonus,
	)

	attackDifferential := tacticalBeamOffenseForState(*spec.Tactical, *active, *activeState, r.legacy17_8Disabled) - tacticalBeamDefenseForState(*spec.Tactical, *targetSpec, *targetState, r.legacy17_8Disabled) + configured.StaticToHitAdjustment
	alwaysHit := tacticalBeamHasEffect(configured, "always_hit")
	shieldPiercing := tacticalBeamHasEffect(configured, "shield_piercing") && !dependencies.HardShields
	armorFirstBypass := (tacticalBeamHasEffect(configured, "armor_piercing") || tacticalBeamHasEffect(configured, "armor_first_bypass")) && !dependencies.ArmorBypassNeutralized
	extraStructureDamage := tacticalBeamHasEffect(configured, "extra_structure_damage")
	marineKill := tacticalBeamHasEffect(configured, "marine_kill")
	enveloping := tacticalBeamHasEffect(configured, "enveloping")
	monsterAntaranQuarter := tacticalBeamHasEffect(configured, "monster_antaran_quarter_damage") && dependencies.TargetMonsterAntaranBand
	liveWeaponCount := tacticalCurrentWeaponCount(*active, *activeState, weapon.Slot)
	totalShots := liveWeaponCount * configured.ShotsPerWeapon

	tacticalBreakCloakOnAction(activeState)
	hitCount := 0
	totalDamage := 0
	flatShotIndex := 0
	for physicalIndex := 0; physicalIndex < liveWeaponCount; physicalIndex++ {
		for burstIndex := 0; burstIndex < configured.ShotsPerWeapon; burstIndex++ {
			flatShotIndex++
			rawRoll, err := r.random(100, spec.Tactical.Rules)
			if err != nil {
				return nil, err
			}
			hitRoll := rawRoll
			if alwaysHit {
				hitRoll = 100
			}
			effectiveRoll := hitRoll
			if hitRoll > spec.Tactical.Rules.BeamMaxHitThreshold {
				effectiveRoll = spec.Tactical.Rules.BeamEffectiveRollCap
			} else {
				effectiveRoll += attackDifferential
				if effectiveRoll > spec.Tactical.Rules.BeamEffectiveRollCap {
					effectiveRoll = spec.Tactical.Rules.BeamEffectiveRollCap
				}
			}
			forcedMiss, displacementRoll, err := tacticalBeamForcedMiss(*spec.Tactical, r, *targetSpec, *targetState)
			if err != nil {
				return nil, err
			}
			hit := !forcedMiss && effectiveRoll >= threshold
			damage := 0
			selectionRoll := 0
			var damageApplications []tacticalDamageApplication
			reflectionRoll := 0
			reflectedDamage := 0
			var reflectedApplication *tacticalDamageApplication
			if hit {
				hitCount++
				damage = beam.InterpolateDamage(scaledMinDamage, scaledMaxDamage, effectiveRoll, threshold, spec.Tactical.Rules.BeamEffectiveRollCap)
				if monsterAntaranQuarter {
					damage /= 4
				}
				damage, reflectedDamage, reflectionRoll, err = tacticalResolveBeamReflection(*spec.Tactical, r, *targetSpec, *targetState, damage)
				if err != nil {
					return nil, err
				}
				if reflectedDamage > 0 {
					reflectionOptions := tacticalDamageOptions{}
					if TacticalSpecialFunctional(*active, *activeState, r.legacy17_8Disabled, "hard_shields") && spec.Tactical.Rules.Specials != nil {
						reflectionOptions.ShieldAdditionalBlock = spec.Tactical.Rules.Specials.HardShieldsAdditionalAbsorption
					}
					applied := applyTacticalDamageWithSpecials(*spec.Tactical, r, *active, activeState, *targetState, reflectedDamage, reflectionOptions)
					reflectedApplication = &applied
				}
				totalDamage += damage
				if damage > 0 {
					selectionRoll, err = r.random(100, spec.Tactical.Rules)
					if err != nil {
						return nil, err
					}

					applicationCount := 1
					if enveloping {
						applicationCount = 4
					}
					damageApplications = make([]tacticalDamageApplication, 0, applicationCount)
					for applicationIndex := 0; applicationIndex < applicationCount; applicationIndex++ {
						options := tacticalDamageOptions{
							ShieldPiercing:       shieldPiercing,
							ArmorFirstBypass:     armorFirstBypass,
							ExtraStructureDamage: extraStructureDamage,
							MarineKill:           marineKill,
							MarineKillSuppressed: dependencies.MarineKillSuppressed,
						}
						if dependencies.HardShields && spec.Tactical.Rules.Specials != nil {
							options.ShieldAdditionalBlock = spec.Tactical.Rules.Specials.HardShieldsAdditionalAbsorption
						}
						if enveloping {
							options.HasShieldArcOverride = true
							options.ShieldArcOverride = applicationIndex
						}
						damageApplications = append(damageApplications, applyTacticalDamageWithSpecials(*spec.Tactical, r, *targetSpec, targetState, *activeState, damage, options))
					}
				}
			}
			if err := r.appendEvent("beam_fired", seatID, commandSequence, map[string]any{
				"ship_id": active.ShipID, "target_ship_id": targetSpec.ShipID, "weapon_slot": payload.WeaponSlot,
				"weapon_id": weapon.WeaponID, "mount_type": canonicalTacticalWeaponMountType(weapon.MountType),
				"modifier_ids":          append([]core.ShipWeaponModifierID(nil), weapon.ModifierIDs...),
				"physical_weapon_index": physicalIndex + 1, "physical_weapon_count": liveWeaponCount,
				"burst_shot_index": burstIndex + 1, "burst_shot_count": configured.ShotsPerWeapon,
				"shot_index": flatShotIndex, "shot_count": totalShots,
				"range_index": rangeIndex, "raw_hit_roll": rawRoll, "effective_hit_roll": effectiveRoll,
				"threshold": threshold, "hit": hit, "forced_miss": forcedMiss, "displacement_roll": displacementRoll, "damage": damage, "reflection_roll": reflectionRoll, "reflected_damage": reflectedDamage,
				"damage_percent": damagePercent, "scaled_min_damage": scaledMinDamage, "scaled_max_damage": scaledMaxDamage,
			}); err != nil {
				return nil, err
			}
			if reflectedApplication != nil {
				if err := r.appendEvent("beam_reflected", seatID, commandSequence, map[string]any{
					"source_ship_id": targetSpec.ShipID, "target_ship_id": active.ShipID, "weapon_slot": payload.WeaponSlot,
					"reflection_roll": reflectionRoll, "reflected_damage": reflectedDamage,
					"incoming_damage": reflectedApplication.IncomingDamage, "post_special_damage": reflectedApplication.PostSpecialDamage,
					"energy_absorbed": reflectedApplication.EnergyAbsorbed, "damper_reduced_damage": reflectedApplication.DamperReducedDamage,
					"shield_arc": reflectedApplication.ShieldArc, "shield_before": reflectedApplication.ShieldBefore, "shield_after": reflectedApplication.ShieldAfter,
					"armor_before": reflectedApplication.ArmorBefore, "armor_after": reflectedApplication.ArmorAfter,
					"structure_damage_before": reflectedApplication.StructureBefore, "structure_damage_after": reflectedApplication.StructureAfter,
				}); err != nil {
					return nil, err
				}
				if activeState.StructureDamage >= active.StructureMax {
					activeState.Destroyed = true
					if err := r.appendEvent("ship_destroyed", seatID, commandSequence, map[string]any{"ship_id": active.ShipID, "reason": "reflection"}); err != nil {
						return nil, err
					}
					winner := remainingWinner(spec, r)
					if winner != 0 {
						return &Result{WinnerSeat: winner, Outcome: TacticalOutcomeVictory, DestroyedShipIDs: destroyedShipIDs(r)}, nil
					}
				}
			}
			for applicationIndex, damageApplied := range damageApplications {
				event := map[string]any{
					"source_ship_id": active.ShipID, "target_ship_id": targetSpec.ShipID, "weapon_slot": payload.WeaponSlot,
					"weapon_id": weapon.WeaponID, "mount_type": canonicalTacticalWeaponMountType(weapon.MountType),
					"modifier_ids":          append([]core.ShipWeaponModifierID(nil), weapon.ModifierIDs...),
					"physical_weapon_index": physicalIndex + 1, "physical_weapon_count": liveWeaponCount,
					"burst_shot_index": burstIndex + 1, "burst_shot_count": configured.ShotsPerWeapon,
					"shot_index": flatShotIndex, "shot_count": totalShots,
					"application_index": applicationIndex + 1, "application_count": len(damageApplications),
					"layer": damageApplied.Layer, "damage": damage, "incoming_damage": damageApplied.IncomingDamage, "selection_roll": selectionRoll,
					"shield_arc": damageApplied.ShieldArc, "shield_before": damageApplied.ShieldBefore, "shield_after": damageApplied.ShieldAfter,
					"post_block_damage": damageApplied.PostBlockDamage, "shield_absorbed": damageApplied.ShieldAbsorbed, "lower_layer_overflow": damageApplied.LowerLayerOverflow,
					"armor_before": damageApplied.ArmorBefore, "armor_after": damageApplied.ArmorAfter,
					"structure_damage_before": damageApplied.StructureBefore, "structure_damage_after": damageApplied.StructureAfter,
					"extra_structure_damage": damageApplied.ExtraStructureDamage,
				}
				if marineKill {
					event["marine_before"] = damageApplied.MarineBefore
					event["marine_after"] = damageApplied.MarineAfter
					event["marine_killed"] = damageApplied.MarineKilled
				}
				if err := r.appendEvent("battle_damage_applied", seatID, commandSequence, event); err != nil {
					return nil, err
				}
			}
		}
	}

	tacticalCompleteWeaponFire(*active, activeState, *weapon, ready)
	if totalShots > 1 {
		if err := r.appendEvent("beam_volley_resolved", seatID, commandSequence, map[string]any{
			"ship_id": active.ShipID, "target_ship_id": targetSpec.ShipID, "weapon_slot": payload.WeaponSlot,
			"weapon_id": weapon.WeaponID, "mount_type": canonicalTacticalWeaponMountType(weapon.MountType),
			"modifier_ids":          append([]core.ShipWeaponModifierID(nil), weapon.ModifierIDs...),
			"physical_weapon_count": liveWeaponCount, "burst_shot_count": configured.ShotsPerWeapon,
			"shot_count": totalShots, "hit_count": hitCount, "total_damage": totalDamage,
		}); err != nil {
			return nil, err
		}
	}
	if hitCount > 0 && targetState.StructureDamage >= targetSpec.StructureMax {
		targetState.Destroyed = true
		if err := r.appendEvent("ship_destroyed", seatID, commandSequence, map[string]any{"ship_id": targetSpec.ShipID}); err != nil {
			return nil, err
		}
		winner := remainingWinner(spec, r)
		if winner != 0 {
			if err := r.appendEvent("winner_determined", seatID, commandSequence, map[string]any{"winner_seat": winner}); err != nil {
				return nil, err
			}
			return &Result{WinnerSeat: winner, Outcome: TacticalOutcomeVictory, DestroyedShipIDs: destroyedShipIDs(r)}, nil
		}
	}
	return nil, nil
}
func waitTargetShipIDs(spec TacticalSpec, r *tacticalRuntime, seatID protocol.SeatID, activeShipID core.ID) []core.ID {
	order := tacticalActivationOrder(r)
	if seatID == 0 || len(order) < 2 {
		return nil
	}
	start := 0
	for i, id := range order {
		if id == activeShipID {
			start = (i + 1) % len(order)
			break
		}
	}
	out := make([]core.ID, 0)
	for offset := 0; offset < len(order); offset++ {
		id := order[(start+offset)%len(order)]
		if id == activeShipID {
			continue
		}
		state := r.shipState(id)
		ship := tacticalShipSpec(spec, id)
		if state == nil || ship == nil || state.Destroyed || state.ActivationComplete || tacticalControllerSeat(*ship, state) != seatID {
			continue
		}
		out = append(out, id)
	}
	return out
}

func nextUnfinishedShipAfter(r *tacticalRuntime, activeShipID core.ID) core.ID {
	order := tacticalActivationOrder(r)
	if len(order) == 0 {
		return 0
	}
	start := 0
	for i, id := range order {
		if id == activeShipID {
			start = (i + 1) % len(order)
			break
		}
	}
	for offset := 0; offset < len(order); offset++ {
		id := order[(start+offset)%len(order)]
		if id == activeShipID {
			continue
		}
		state := r.shipState(id)
		if state != nil && !state.Destroyed && !state.ActivationComplete {
			return id
		}
	}
	return 0
}

func prepareWaitActivation(spec Spec, r *tacticalRuntime, seatID protocol.SeatID, commandSequence uint32, payload WaitActivationPayload) error {
	if spec.Tactical == nil {
		return fmt.Errorf("battle has no tactical spec")
	}
	active := tacticalShipSpec(*spec.Tactical, r.state.ActiveShipID)
	activeState := r.shipState(r.state.ActiveShipID)
	if active == nil || activeState == nil {
		return fmt.Errorf("active tactical ship %d is missing from spec", r.state.ActiveShipID)
	}
	if seatID == 0 || tacticalControllerSeat(*active, activeState) != seatID {
		return fmt.Errorf("seat %d does not control active ship %d", seatID, r.state.ActiveShipID)
	}
	if payload.ShipID != r.state.ActiveShipID {
		return fmt.Errorf("wait ship %d is not active ship %d", payload.ShipID, r.state.ActiveShipID)
	}
	if activeState == nil || activeState.Destroyed || activeState.ActivationComplete {
		return fmt.Errorf("ship %d cannot wait", payload.ShipID)
	}
	targets := waitTargetShipIDs(*spec.Tactical, r, seatID, payload.ShipID)
	if len(targets) == 0 {
		return fmt.Errorf("ship %d has no unfinished friendly wait target", payload.ShipID)
	}
	targetShipID := payload.TargetShipID
	if targetShipID == 0 {
		targetShipID = targets[0]
	} else {
		allowed := false
		for _, id := range targets {
			if id == targetShipID {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("ship %d is not an unfinished friendly wait target for ship %d", targetShipID, payload.ShipID)
		}
	}
	r.state.ActiveShipID = targetShipID
	return r.appendEvent("activation_waited", seatID, commandSequence, map[string]any{"ship_id": payload.ShipID, "target_ship_id": targetShipID, "round": r.state.Round})
}

func tacticalShipResourcesExhausted(state *TacticalShipState) bool {
	if state == nil || state.Destroyed || state.ActivationComplete || state.MovementCurrent > 0 {
		return false
	}
	for _, weapon := range state.Weapons {
		if weapon.Ready {
			return false
		}
	}
	return true
}

func autoCompleteExhaustedActiveShip(spec Spec, r *tacticalRuntime, commandSequence uint32) error {
	active := tacticalShipSpec(*spec.Tactical, r.state.ActiveShipID)
	activeState := r.shipState(r.state.ActiveShipID)
	if active == nil || !tacticalShipResourcesExhausted(activeState) {
		return nil
	}
	activeState.ActivationComplete = true
	eventSeat := tacticalControllerSeat(*active, activeState)
	if err := r.appendEvent("activation_auto_ended", eventSeat, commandSequence, map[string]any{
		"ship_id": active.ShipID, "round": r.state.Round, "reason": "resources_exhausted",
	}); err != nil {
		return err
	}
	return advanceAfterActivationCompletion(spec, r)
}

func advanceAfterActivationCompletion(spec Spec, r *tacticalRuntime) error {
	if spec.Tactical == nil || r == nil {
		return fmt.Errorf("tactical scheduler requires battle state")
	}
	active := tacticalShipSpec(*spec.Tactical, r.state.ActiveShipID)
	activeState := r.shipState(r.state.ActiveShipID)
	completedSeat := protocol.SeatID(0)
	if active != nil && activeState != nil {
		completedSeat = tacticalControllerSeat(*active, activeState)
	}

	if r.state.PassKind == TacticalPassKindTimeWarp {
		if nextShipID := nextUnfinishedShipAfter(r, r.state.ActiveShipID); nextShipID != 0 {
			r.state.ActiveShipID = nextShipID
			return nil
		}
		if err := tacticalEndOwnPassSpecials(*spec.Tactical, r, r.state.TimeWarpSeatID, true, r.state.NextCommandSequence-1); err != nil {
			return err
		}
		afterSeat := r.state.TimeWarpSeatID
		r.state.TimeWarpQueue = nil
		r.state.TimeWarpSeatID = 0
		r.state.PassKind = TacticalPassKindNormal
		if started, err := tacticalBeginNextTimeWarpPass(spec, r, afterSeat, r.state.NextCommandSequence-1); err != nil {
			return err
		} else if started {
			return nil
		}
		return finalizeTacticalRound(spec, r)
	}

	if completedSeat != 0 && tacticalSeatPassComplete(*spec.Tactical, r, completedSeat, false) {
		if err := tacticalEndOwnPassSpecials(*spec.Tactical, r, completedSeat, false, r.state.NextCommandSequence-1); err != nil {
			return err
		}
	}
	if nextShipID := nextUnfinishedShipAfter(r, r.state.ActiveShipID); nextShipID != 0 {
		r.state.ActiveShipID = nextShipID
		return nil
	}
	if started, err := tacticalBeginNextTimeWarpPass(spec, r, 0, r.state.NextCommandSequence-1); err != nil {
		return err
	} else if started {
		return nil
	}
	return finalizeTacticalRound(spec, r)
}

func finalizeTacticalRound(spec Spec, r *tacticalRuntime) error {
	if spec.Tactical == nil || r == nil {
		return fmt.Errorf("tactical round finalization requires battle state")
	}
	if err := rechargeTacticalShieldsWithEvents(*spec.Tactical, r, r.state.NextCommandSequence-1); err != nil {
		return err
	}
	if err := tacticalRepairShips(*spec.Tactical, r, r.state.NextCommandSequence-1); err != nil {
		return err
	}
	r.state.Round++
	r.state.PassKind = TacticalPassKindNormal
	r.state.TimeWarpSeatID = 0
	r.state.TimeWarpQueue = nil
	for i := range r.state.Ships {
		if r.state.Ships[i].Destroyed {
			continue
		}
		r.state.Ships[i].ActivationComplete = false
		r.state.Ships[i].MovementCurrent = r.state.Ships[i].MovementMax
		for j := range r.state.Ships[i].Weapons {
			r.state.Ships[i].Weapons[j].Ready = true
		}
		if ship := tacticalShipSpec(*spec.Tactical, r.state.Ships[i].ShipID); ship != nil {
			tacticalStartShipTurnSpecials(*spec.Tactical, *ship, &r.state.Ships[i])
		}
	}
	if err := advanceTacticalOrdnanceRound(*spec.Tactical, r); err != nil {
		return err
	}
	r.recomputeInitiative(*spec.Tactical)
	if len(r.state.InitiativeOrder) == 0 {
		return fmt.Errorf("new tactical round has no live ships")
	}
	r.state.ActiveShipID = r.state.InitiativeOrder[0]
	return r.appendEvent("round_started", 0, 0, map[string]any{"round": r.state.Round, "initiative_order": r.state.InitiativeOrder, "active_ship_id": r.state.ActiveShipID})
}

func prepareEndActivation(spec Spec, r *tacticalRuntime, seatID protocol.SeatID, commandSequence uint32, payload EndActivationPayload) error {
	active := tacticalShipSpec(*spec.Tactical, r.state.ActiveShipID)
	activeState := r.shipState(r.state.ActiveShipID)
	if active == nil || activeState == nil {
		return fmt.Errorf("active tactical ship %d is missing from spec", r.state.ActiveShipID)
	}
	if seatID == 0 || tacticalControllerSeat(*active, activeState) != seatID {
		return fmt.Errorf("seat %d does not control active ship %d", seatID, r.state.ActiveShipID)
	}
	if payload.ShipID != r.state.ActiveShipID {
		return fmt.Errorf("end-activation ship %d is not active ship %d", payload.ShipID, r.state.ActiveShipID)
	}
	if activeState == nil || activeState.Destroyed || activeState.ActivationComplete {
		return fmt.Errorf("ship %d cannot end activation", payload.ShipID)
	}
	activeState.ActivationComplete = true
	if err := r.appendEvent("activation_ended", seatID, commandSequence, map[string]any{"ship_id": payload.ShipID, "round": r.state.Round}); err != nil {
		return err
	}
	return advanceAfterActivationCompletion(spec, r)
}

func tacticalRangeIndex(ax, ay, bx, by int) int {
	dx := absInt(ax - bx)
	dy := absInt(ay - by)
	major, minor := dx, dy
	if dy > dx {
		major, minor = dy, dx
	}
	raw := major + minor/2
	return (raw + 2) / 3
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func (r *tacticalRuntime) random(n uint32, rules TacticalRulesSnapshot) (int, error) {
	if n == 0 {
		return 0, fmt.Errorf("tactical Random requires positive bound")
	}
	q := ^uint32(0) / n
	if q == 0 {
		return 0, fmt.Errorf("tactical Random bound %d is too large", n)
	}
	cutoff := q * n
	for {
		r.state.RNGState = r.state.RNGState*rules.RNGMultiplier + rules.RNGIncrement
		if r.state.RNGState < cutoff {
			return int(r.state.RNGState/q) + 1, nil
		}
	}
}

func destroyedShipIDs(r *tacticalRuntime) []core.ID {
	ids := make([]core.ID, 0)
	for _, ship := range r.state.Ships {
		if ship.Destroyed {
			ids = append(ids, ship.ShipID)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func remainingWinner(spec Spec, r *tacticalRuntime) protocol.SeatID {
	seatAlive := make(map[protocol.SeatID]bool)
	for _, ship := range spec.Tactical.Ships {
		state := r.shipState(ship.ShipID)
		if state != nil && !state.Destroyed {
			seatAlive[tacticalControllerSeat(ship, state)] = true
		}
	}
	if len(seatAlive) != 1 {
		return 0
	}
	for seat := range seatAlive {
		return seat
	}
	return 0
}

func (s *Session) CommitPreparedCommand(prepared *PreparedCommand) error {
	if prepared == nil {
		return fmt.Errorf("prepared tactical command is nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.phase != PhaseActive {
		return fmt.Errorf("cannot commit tactical command in phase %q", s.phase)
	}
	if s.tactical == nil {
		return fmt.Errorf("battle has no tactical runtime")
	}
	if s.tacticalRevision != prepared.baseRevision {
		return fmt.Errorf("stale prepared tactical command")
	}
	runtime := cloneTacticalRuntime(prepared.runtime)
	s.tactical = &runtime
	s.tacticalRevision++
	if prepared.result != nil {
		result := *prepared.result
		result.WinnerSeats = append([]protocol.SeatID(nil), prepared.result.WinnerSeats...)
		result.DestroyedShipIDs = append([]core.ID(nil), prepared.result.DestroyedShipIDs...)
		s.result = &result
		s.phase = PhaseCompleted
	}
	return nil
}

func cloneTacticalSpec(spec TacticalSpec) TacticalSpec {
	out := spec
	out.Rules.BeamToHitRangeModifiers = append([]int(nil), spec.Rules.BeamToHitRangeModifiers...)
	out.Rules.BeamDamageRangeModifiers = append([]int(nil), spec.Rules.BeamDamageRangeModifiers...)
	out.Rules.Ordnance = cloneTacticalOrdnanceRulesSnapshot(spec.Rules.Ordnance)
	if spec.Rules.Fighter != nil {
		fighter := *spec.Rules.Fighter
		fighter.ArmorDurabilityMultiples = append([]int(nil), spec.Rules.Fighter.ArmorDurabilityMultiples...)
		out.Rules.Fighter = &fighter
	}
	if spec.Rules.Boarding != nil {
		boarding := *spec.Rules.Boarding
		boarding.CrewQualityBonuses = append([]int(nil), spec.Rules.Boarding.CrewQualityBonuses...)
		out.Rules.Boarding = &boarding
	}
	if spec.Rules.Specials != nil {
		specials := *spec.Rules.Specials
		out.Rules.Specials = &specials
	}
	out.BoardingDoctrines = append([]TacticalBoardingDoctrineSpec(nil), spec.BoardingDoctrines...)
	out.FighterDoctrines = make([]TacticalFighterDoctrineSpec, len(spec.FighterDoctrines))
	for i := range spec.FighterDoctrines {
		out.FighterDoctrines[i] = spec.FighterDoctrines[i]
		out.FighterDoctrines[i].BeamPayload = cloneTacticalFighterPayload(spec.FighterDoctrines[i].BeamPayload)
		out.FighterDoctrines[i].BombPayload = cloneTacticalFighterPayload(spec.FighterDoctrines[i].BombPayload)
	}
	out.Ships = make([]TacticalShipSpec, len(spec.Ships))
	for i := range spec.Ships {
		out.Ships[i] = spec.Ships[i]
		out.Ships[i].Weapons = make([]TacticalWeaponSpec, len(spec.Ships[i].Weapons))
		for j := range spec.Ships[i].Weapons {
			out.Ships[i].Weapons[j] = spec.Ships[i].Weapons[j]
			out.Ships[i].Weapons[j].ModifierIDs = append([]core.ShipWeaponModifierID(nil), spec.Ships[i].Weapons[j].ModifierIDs...)
			out.Ships[i].Weapons[j].InnateEffectIDs = append([]string(nil), spec.Ships[i].Weapons[j].InnateEffectIDs...)
		}
		if spec.Ships[i].VisualGenome != nil {
			genome := core.CloneShipVisualGenome(*spec.Ships[i].VisualGenome)
			out.Ships[i].VisualGenome = &genome
		}
		if spec.Ships[i].Internal != nil {
			internal := *spec.Ships[i].Internal
			internal.Weapons = append([]TacticalInternalWeaponSpec(nil), spec.Ships[i].Internal.Weapons...)
			internal.Specials = append([]TacticalInternalSpecialSpec(nil), spec.Ships[i].Internal.Specials...)
			out.Ships[i].Internal = &internal
		}
	}
	return out
}

func cloneTacticalFighterPayload(in *TacticalFighterPayloadSpec) *TacticalFighterPayloadSpec {
	if in == nil {
		return nil
	}
	out := *in
	out.InnateEffectIDs = append([]string(nil), in.InnateEffectIDs...)
	return &out
}

func cloneTacticalState(state TacticalState) TacticalState {
	out := state
	out.InitiativeOrder = append([]core.ID(nil), state.InitiativeOrder...)
	out.TimeWarpQueue = append([]core.ID(nil), state.TimeWarpQueue...)
	out.Ships = make([]TacticalShipState, len(state.Ships))
	for i := range state.Ships {
		out.Ships[i] = state.Ships[i]
		out.Ships[i].Weapons = append([]TacticalWeaponState(nil), state.Ships[i].Weapons...)
		out.Ships[i].Internal = cloneTacticalInternalState(state.Ships[i].Internal)
	}
	return out
}

func cloneTacticalEvents(events []TacticalEvent) []TacticalEvent {
	out := make([]TacticalEvent, len(events))
	for i := range events {
		out[i] = events[i]
		out[i].Data = append(json.RawMessage(nil), events[i].Data...)
	}
	return out
}

func cloneTacticalRuntime(r tacticalRuntime) tacticalRuntime {
	return tacticalRuntime{state: cloneTacticalState(r.state), ordnance: cloneTacticalOrdnanceState(r.ordnance), fighters: cloneTacticalFighterState(r.fighters), events: cloneTacticalEvents(r.events), nextEventSequence: r.nextEventSequence, nextMovingActorOrder: r.nextMovingActorOrder, legacy17_8Disabled: r.legacy17_8Disabled}
}

// CloneTacticalSpec returns a detached immutable tactical battle snapshot.
func CloneTacticalSpec(spec TacticalSpec) TacticalSpec {
	return cloneTacticalSpec(spec)
}
