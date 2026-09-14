package battle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"sort"

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

	BaselineTacticalRNGState uint32 = 0x001500BD

	TacticalTurningNormal             = "normal"
	TacticalTurningInertialStabilizer = "inertial_stabilizer"
	TacticalTurningInertialNullifier  = "inertial_nullifier"
	TacticalCoordinateSafetyLimit     = 1_000_000
)

type TacticalRulesSnapshot struct {
	SchemaVersion                int    `json:"schema_version"`
	InitiativeBeamOffenseDivisor int    `json:"initiative_beam_offense_divisor"`
	RNGMultiplier                uint32 `json:"rng_multiplier"`
	RNGIncrement                 uint32 `json:"rng_increment"`
	BeamBaseHitThreshold         int    `json:"beam_base_hit_threshold"`
	BeamMaxHitThreshold          int    `json:"beam_max_hit_threshold"`
	BeamEffectiveRollCap         int    `json:"beam_effective_roll_cap"`
	BeamToHitRangeModifiers      []int  `json:"beam_to_hit_range_modifiers"`
	BeamDamageRangeModifiers     []int  `json:"beam_damage_range_modifiers"`
}

type TacticalWeaponSpec struct {
	Slot      int    `json:"slot"`
	WeaponID  string `json:"weapon_id"`
	Count     int    `json:"count"`
	MinDamage int    `json:"min_damage"`
	MaxDamage int    `json:"max_damage"`
}

type TacticalShipSpec struct {
	ShipID               core.ID                `json:"ship_id"`
	EmpireID             core.ID                `json:"empire_id"`
	SeatID               protocol.SeatID        `json:"seat_id"`
	X                    int                    `json:"x"`
	Y                    int                    `json:"y"`
	Facing               int                    `json:"facing"`
	TurningMode          string                 `json:"turning_mode,omitempty"`
	HullID               string                 `json:"hull_id"`
	WarpDriveID          string                 `json:"warp_drive_id"`
	ComputerID           string                 `json:"computer_id"`
	ArmorID              string                 `json:"armor_id"`
	SourceDesignID       core.ID                `json:"source_design_id,omitempty"`
	SourceDesignRevision uint32                 `json:"source_design_revision,omitempty"`
	StrategicPictureID   int                    `json:"strategic_picture_id,omitempty"`
	SourceVisualRevision uint32                 `json:"source_visual_revision,omitempty"`
	VisualGenome         *core.ShipVisualGenome `json:"visual_genome,omitempty"`
	Weapons              []TacticalWeaponSpec   `json:"weapons,omitempty"`
	CurrentCombatSpeed   int                    `json:"current_combat_speed"`
	BeamOffense          int                    `json:"beam_offense"`
	BeamDefense          int                    `json:"beam_defense"`
	ArmorMax             int                    `json:"armor_max"`
	StructureMax         int                    `json:"structure_max"`
}

type TacticalSpec struct {
	Rules             TacticalRulesSnapshot `json:"rules"`
	InitiativeEnabled bool                  `json:"initiative_enabled"`
	InitialRNGState   uint32                `json:"initial_rng_state,omitempty"`
	Ships             []TacticalShipSpec    `json:"ships"`
}

type TacticalWeaponState struct {
	Slot  int  `json:"slot"`
	Ready bool `json:"ready"`
}

type TacticalShipState struct {
	ShipID             core.ID               `json:"ship_id"`
	X                  int                   `json:"x"`
	Y                  int                   `json:"y"`
	Facing             int                   `json:"facing"`
	MovementCurrent    int                   `json:"movement_current"`
	MovementMax        int                   `json:"movement_max"`
	ActivationComplete bool                  `json:"activation_complete"`
	ArmorCurrent       int                   `json:"armor_current"`
	StructureDamage    int                   `json:"structure_damage"`
	Weapons            []TacticalWeaponState `json:"weapons,omitempty"`
	Destroyed          bool                  `json:"destroyed"`
}

type TacticalState struct {
	Round               uint32              `json:"round"`
	InitiativeOrder     []core.ID           `json:"initiative_order"`
	ActiveShipID        core.ID             `json:"active_ship_id"`
	NextCommandSequence uint32              `json:"next_command_sequence"`
	RNGState            uint32              `json:"rng_state,omitempty"`
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
	X                      int `json:"x"`
	Y                      int `json:"y"`
	MoveCost               int `json:"move_cost"`
	ResultingFacing        int `json:"resulting_facing"`
	MovementRemainingAfter int `json:"movement_remaining_after"`
}

type TacticalWeaponView struct {
	Slot      int    `json:"slot"`
	WeaponID  string `json:"weapon_id"`
	Count     int    `json:"count"`
	MinDamage int    `json:"min_damage"`
	MaxDamage int    `json:"max_damage"`
	Ready     bool   `json:"ready"`
}

type TacticalShipView struct {
	ShipID               core.ID                `json:"ship_id"`
	EmpireID             core.ID                `json:"empire_id"`
	SeatID               protocol.SeatID        `json:"seat_id"`
	HullID               string                 `json:"hull_id"`
	WarpDriveID          string                 `json:"warp_drive_id"`
	ComputerID           string                 `json:"computer_id"`
	ArmorID              string                 `json:"armor_id"`
	SourceDesignID       core.ID                `json:"source_design_id,omitempty"`
	SourceDesignRevision uint32                 `json:"source_design_revision,omitempty"`
	StrategicPictureID   int                    `json:"strategic_picture_id,omitempty"`
	SourceVisualRevision uint32                 `json:"source_visual_revision,omitempty"`
	VisualGenome         *core.ShipVisualGenome `json:"visual_genome,omitempty"`
	X                    int                    `json:"x"`
	Y                    int                    `json:"y"`
	Facing               int                    `json:"facing"`
	MovementCurrent      int                    `json:"movement_current"`
	MovementMax          int                    `json:"movement_max"`
	ActivationComplete   bool                   `json:"activation_complete"`
	ArmorCurrent         int                    `json:"armor_current"`
	ArmorMax             int                    `json:"armor_max"`
	StructureCurrent     int                    `json:"structure_current"`
	StructureMax         int                    `json:"structure_max"`
	BeamOffense          int                    `json:"beam_offense"`
	BeamDefense          int                    `json:"beam_defense"`
	Weapons              []TacticalWeaponView   `json:"weapons,omitempty"`
	Destroyed            bool                   `json:"destroyed"`
}

type TacticalFireTarget struct {
	TargetShipID core.ID `json:"target_ship_id"`
	RangeIndex   int     `json:"range_index"`
}

type TacticalFireAction struct {
	ShipID     core.ID              `json:"ship_id"`
	WeaponSlot int                  `json:"weapon_slot"`
	WeaponID   string               `json:"weapon_id"`
	Targets    []TacticalFireTarget `json:"targets"`
}

type TacticalView struct {
	State             TacticalState        `json:"state"`
	Events            []TacticalEvent      `json:"events"`
	Ships             []TacticalShipView   `json:"ships"`
	LegalMoves        []TacticalMoveOption `json:"legal_moves,omitempty"`
	LegalFireActions  []TacticalFireAction `json:"legal_fire_actions,omitempty"`
	CanEndActivation  bool                 `json:"can_end_activation"`
	CanWaitActivation bool                 `json:"can_wait_activation"`
	WaitTargetShipIDs []core.ID            `json:"wait_target_ship_ids,omitempty"`
}

type MoveShipPayload struct {
	ShipID core.ID `json:"ship_id"`
	X      int     `json:"x"`
	Y      int     `json:"y"`
}

type FireBeamPayload struct {
	ShipID       core.ID `json:"ship_id"`
	TargetShipID core.ID `json:"target_ship_id"`
	WeaponSlot   int     `json:"weapon_slot"`
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
	state             TacticalState
	events            []TacticalEvent
	nextEventSequence uint64
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
		if err := validateBaselineCombatant(ship); err != nil {
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

func validateBaselineCombatant(s TacticalShipSpec) error {
	if s.HullID != "frigate" || s.ComputerID != "electronic_computer" || s.ArmorID != "titanium_armor" || s.ArmorMax != 4 || s.StructureMax != 4 {
		return fmt.Errorf("tactical ship %d is outside the Slice 15.5 Frigate/Electronic/Titanium baseline", s.ShipID)
	}
	switch s.WarpDriveID {
	case "fusion_drive":
		if s.CurrentCombatSpeed != 22 {
			return fmt.Errorf("fusion tactical ship %d requires pristine combat speed 22", s.ShipID)
		}
	case "nuclear_drive":
		if s.CurrentCombatSpeed != 20 {
			return fmt.Errorf("nuclear tactical ship %d requires pristine combat speed 20", s.ShipID)
		}
	default:
		return fmt.Errorf("tactical ship %d drive %q is outside the Slice 15.5 baseline", s.ShipID, s.WarpDriveID)
	}
	if s.BeamOffense != 25 || s.BeamDefense != 0 {
		return fmt.Errorf("tactical ship %d offense/defense is outside the Slice 15.5 baseline", s.ShipID)
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
		if w.WeaponID != "laser_cannon" || w.Count <= 0 || w.MinDamage != 1 || w.MaxDamage != 4 {
			return fmt.Errorf("tactical ship %d weapon slot %d is outside the supported Laser baseline", s.ShipID, w.Slot)
		}
		previousSlot = w.Slot
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
			Ships:               make([]TacticalShipState, len(spec.Ships)),
		},
		nextEventSequence: 1,
	}
	for i, ship := range spec.Ships {
		state := TacticalShipState{ShipID: ship.ShipID, X: ship.X, Y: ship.Y, Facing: ship.Facing, MovementCurrent: ship.CurrentCombatSpeed, MovementMax: ship.CurrentCombatSpeed, ArmorCurrent: ship.ArmorMax, Weapons: make([]TacticalWeaponState, len(ship.Weapons))}
		for j, weapon := range ship.Weapons {
			state.Weapons[j] = TacticalWeaponState{Slot: weapon.Slot, Ready: true}
		}
		r.state.Ships[i] = state
	}
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
		items = append(items, item{id: ship.ShipID, value: ship.CurrentCombatSpeed + ship.BeamOffense/spec.Rules.InitiativeBeamOffenseDivisor})
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
	if result == nil && (command.Kind == CommandMoveShip || command.Kind == CommandFireBeam) {
		if err := autoCompleteExhaustedActiveShip(s.spec, &prepared.runtime, command.Sequence); err != nil {
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
	if seatID == 0 || activeSpec.SeatID != seatID {
		return nil, fmt.Errorf("seat %d does not control active ship %d", seatID, r.state.ActiveShipID)
	}
	if payload.ShipID != r.state.ActiveShipID {
		return nil, fmt.Errorf("retreat ship %d is not active ship %d", payload.ShipID, r.state.ActiveShipID)
	}
	if activeState.Destroyed || activeState.ActivationComplete {
		return nil, fmt.Errorf("active ship %d cannot retreat", payload.ShipID)
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
	if seatID == 0 || activeSpec.SeatID != seatID {
		return fmt.Errorf("seat %d does not control active ship %d", seatID, r.state.ActiveShipID)
	}
	if payload.ShipID != r.state.ActiveShipID {
		return fmt.Errorf("move ship %d is not active ship %d", payload.ShipID, r.state.ActiveShipID)
	}
	if activeState.Destroyed || activeState.ActivationComplete {
		return fmt.Errorf("active ship %d cannot move", payload.ShipID)
	}
	option, err := tacticalMoveOption(r, *activeSpec, *activeState, payload.X, payload.Y)
	if err != nil {
		return err
	}
	fromX, fromY, fromFacing := activeState.X, activeState.Y, activeState.Facing
	activeState.X, activeState.Y, activeState.Facing, activeState.MovementCurrent = option.X, option.Y, option.ResultingFacing, option.MovementRemainingAfter
	return r.appendEvent("ship_moved", seatID, commandSequence, map[string]any{"ship_id": payload.ShipID, "from_x": fromX, "from_y": fromY, "to_x": option.X, "to_y": option.Y, "facing_before": fromFacing, "facing_after": option.ResultingFacing, "move_cost": option.MoveCost, "movement_remaining": option.MovementRemainingAfter})
}

func tacticalMoveOption(r *tacticalRuntime, shipSpec TacticalShipSpec, shipState TacticalShipState, x, y int) (TacticalMoveOption, error) {
	if x == shipState.X && y == shipState.Y {
		return TacticalMoveOption{}, fmt.Errorf("same-coordinate Tactical move is not a rotate-only command")
	}
	if absInt(x) > TacticalCoordinateSafetyLimit || absInt(y) > TacticalCoordinateSafetyLimit {
		return TacticalMoveOption{}, fmt.Errorf("Tactical destination (%d,%d) exceeds technical coordinate safety envelope", x, y)
	}
	if absInt(x-shipState.X) > shipState.MovementCurrent || absInt(y-shipState.Y) > shipState.MovementCurrent {
		return TacticalMoveOption{}, fmt.Errorf("Tactical destination (%d,%d) exceeds ship %d movement budget", x, y, shipState.ShipID)
	}
	for _, other := range r.state.Ships {
		if other.ShipID != shipState.ShipID && !other.Destroyed && other.X == x && other.Y == y {
			return TacticalMoveOption{}, fmt.Errorf("Tactical destination (%d,%d) is occupied by ship %d", x, y, other.ShipID)
		}
	}
	cost, facing := tacticalMoveCost(shipState.X, shipState.Y, shipState.Facing, x, y, turnCostPerFacing(shipSpec))
	if cost > shipState.MovementCurrent {
		return TacticalMoveOption{}, fmt.Errorf("Tactical move costs %d movement points but ship %d has %d remaining", cost, shipState.ShipID, shipState.MovementCurrent)
	}
	return TacticalMoveOption{X: x, Y: y, MoveCost: cost, ResultingFacing: facing, MovementRemainingAfter: shipState.MovementCurrent - cost}, nil
}

func legalMoves(spec TacticalSpec, r *tacticalRuntime) []TacticalMoveOption {
	state := r.shipState(r.state.ActiveShipID)
	ship := tacticalShipSpec(spec, r.state.ActiveShipID)
	if state == nil || ship == nil || state.Destroyed || state.ActivationComplete || state.MovementCurrent <= 0 {
		return nil
	}
	radius := state.MovementCurrent
	out := make([]TacticalMoveOption, 0, (radius*2+1)*(radius*2+1)/2)
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			option, err := tacticalMoveOption(r, *ship, *state, state.X+dx, state.Y+dy)
			if err == nil {
				out = append(out, option)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
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
func buildTacticalView(spec TacticalSpec, r tacticalRuntime) TacticalView {
	view := TacticalView{State: cloneTacticalState(r.state), Events: cloneTacticalEvents(r.events), Ships: tacticalShipViews(spec, &r)}
	state := r.shipState(r.state.ActiveShipID)
	if state != nil && !state.Destroyed && !state.ActivationComplete {
		view.CanEndActivation = true
		if activeSpec := tacticalShipSpec(spec, r.state.ActiveShipID); activeSpec != nil {
			view.WaitTargetShipIDs = waitTargetShipIDs(spec, &r, activeSpec.SeatID, r.state.ActiveShipID)
			view.CanWaitActivation = len(view.WaitTargetShipIDs) != 0
		}
		view.LegalMoves = legalMoves(spec, &r)
		view.LegalFireActions = legalFireActions(spec, &r)
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
			ShipID: shipSpec.ShipID, EmpireID: shipSpec.EmpireID, SeatID: shipSpec.SeatID,
			HullID: shipSpec.HullID, WarpDriveID: shipSpec.WarpDriveID, ComputerID: shipSpec.ComputerID, ArmorID: shipSpec.ArmorID,
			SourceDesignID: shipSpec.SourceDesignID, SourceDesignRevision: shipSpec.SourceDesignRevision, StrategicPictureID: shipSpec.StrategicPictureID,
			SourceVisualRevision: shipSpec.SourceVisualRevision,
			X:                    state.X, Y: state.Y, Facing: state.Facing, MovementCurrent: state.MovementCurrent, MovementMax: state.MovementMax,
			ActivationComplete: state.ActivationComplete, ArmorCurrent: state.ArmorCurrent, ArmorMax: shipSpec.ArmorMax,
			StructureCurrent: shipSpec.StructureMax - state.StructureDamage, StructureMax: shipSpec.StructureMax,
			BeamOffense: shipSpec.BeamOffense, BeamDefense: shipSpec.BeamDefense, Destroyed: state.Destroyed,
			Weapons: make([]TacticalWeaponView, 0, len(shipSpec.Weapons)),
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
			if weapon := weaponState(state, weaponSpec.Slot); weapon != nil {
				ready = weapon.Ready
			}
			view.Weapons = append(view.Weapons, TacticalWeaponView{Slot: weaponSpec.Slot, WeaponID: weaponSpec.WeaponID, Count: weaponSpec.Count, MinDamage: weaponSpec.MinDamage, MaxDamage: weaponSpec.MaxDamage, Ready: ready})
		}
		out = append(out, view)
	}
	return out
}

func legalFireActions(spec TacticalSpec, r *tacticalRuntime) []TacticalFireAction {
	activeSpec := tacticalShipSpec(spec, r.state.ActiveShipID)
	activeState := r.shipState(r.state.ActiveShipID)
	if activeSpec == nil || activeState == nil || activeState.Destroyed || activeState.ActivationComplete {
		return nil
	}
	out := make([]TacticalFireAction, 0, len(activeSpec.Weapons))
	for _, weaponSpec := range activeSpec.Weapons {
		ready := weaponState(activeState, weaponSpec.Slot)
		if ready == nil || !ready.Ready || weaponSpec.WeaponID != "laser_cannon" {
			continue
		}
		action := TacticalFireAction{ShipID: activeSpec.ShipID, WeaponSlot: weaponSpec.Slot, WeaponID: weaponSpec.WeaponID}
		for _, targetSpec := range spec.Ships {
			if targetSpec.EmpireID == activeSpec.EmpireID {
				continue
			}
			targetState := r.shipState(targetSpec.ShipID)
			if targetState == nil || targetState.Destroyed {
				continue
			}
			rangeIndex := tacticalRangeIndex(activeState.X, activeState.Y, targetState.X, targetState.Y)
			if rangeIndex < 0 || rangeIndex >= len(spec.Rules.BeamToHitRangeModifiers) {
				continue
			}
			action.Targets = append(action.Targets, TacticalFireTarget{TargetShipID: targetSpec.ShipID, RangeIndex: rangeIndex})
		}
		if len(action.Targets) > 0 {
			out = append(out, action)
		}
	}
	return out
}

func prepareFireBeam(spec Spec, r *tacticalRuntime, seatID protocol.SeatID, commandSequence uint32, payload FireBeamPayload) (*Result, error) {
	active := tacticalShipSpec(*spec.Tactical, r.state.ActiveShipID)
	if active == nil {
		return nil, fmt.Errorf("active tactical ship %d is missing from spec", r.state.ActiveShipID)
	}
	if seatID == 0 || active.SeatID != seatID {
		return nil, fmt.Errorf("seat %d does not control active ship %d", seatID, r.state.ActiveShipID)
	}
	if payload.ShipID != r.state.ActiveShipID {
		return nil, fmt.Errorf("fire ship %d is not active ship %d", payload.ShipID, r.state.ActiveShipID)
	}
	targetSpec := tacticalShipSpec(*spec.Tactical, payload.TargetShipID)
	targetState := r.shipState(payload.TargetShipID)
	if targetSpec == nil || targetState == nil || targetState.Destroyed {
		return nil, fmt.Errorf("target ship %d is not a live tactical ship", payload.TargetShipID)
	}
	if targetSpec.EmpireID == active.EmpireID {
		return nil, fmt.Errorf("target ship %d is not an opponent", payload.TargetShipID)
	}
	weapon := weaponSpec(*active, payload.WeaponSlot)
	activeState := r.shipState(active.ShipID)
	ready := weaponState(activeState, payload.WeaponSlot)
	if weapon == nil || ready == nil {
		return nil, fmt.Errorf("ship %d has no weapon slot %d", active.ShipID, payload.WeaponSlot)
	}
	if !ready.Ready {
		return nil, fmt.Errorf("ship %d weapon slot %d is not ready", active.ShipID, payload.WeaponSlot)
	}
	if weapon.WeaponID != "laser_cannon" || weapon.Count <= 0 || weapon.MinDamage != 1 || weapon.MaxDamage != 4 {
		return nil, fmt.Errorf("weapon slot %d is outside the supported Laser baseline", payload.WeaponSlot)
	}

	activePosition := r.shipState(active.ShipID)
	if activePosition == nil {
		return nil, fmt.Errorf("active tactical ship %d has no runtime position", active.ShipID)
	}
	rangeIndex := tacticalRangeIndex(activePosition.X, activePosition.Y, targetState.X, targetState.Y)
	if rangeIndex < 0 || rangeIndex >= len(spec.Tactical.Rules.BeamToHitRangeModifiers) {
		return nil, fmt.Errorf("Beam range index %d is outside the evidenced table", rangeIndex)
	}
	rangeHitModifier := spec.Tactical.Rules.BeamToHitRangeModifiers[rangeIndex]
	threshold := spec.Tactical.Rules.BeamBaseHitThreshold - rangeHitModifier
	if threshold > spec.Tactical.Rules.BeamMaxHitThreshold {
		threshold = spec.Tactical.Rules.BeamMaxHitThreshold
	}
	hitCount := 0
	totalDamage := 0
	for shotIndex := 0; shotIndex < weapon.Count; shotIndex++ {
		rawRoll, err := r.random(100, spec.Tactical.Rules)
		if err != nil {
			return nil, err
		}
		effectiveRoll := rawRoll
		if rawRoll > spec.Tactical.Rules.BeamMaxHitThreshold {
			effectiveRoll = spec.Tactical.Rules.BeamEffectiveRollCap
		} else {
			effectiveRoll += active.BeamOffense - targetSpec.BeamDefense
			if effectiveRoll > spec.Tactical.Rules.BeamEffectiveRollCap {
				effectiveRoll = spec.Tactical.Rules.BeamEffectiveRollCap
			}
		}
		hit := effectiveRoll >= threshold
		damage := 0
		selectionRoll := 0
		layer := ""
		beforeArmor, afterArmor := targetState.ArmorCurrent, targetState.ArmorCurrent
		beforeStructure, afterStructure := targetState.StructureDamage, targetState.StructureDamage
		if hit {
			hitCount++
			damage = beamDamage(*weapon, spec.Tactical.Rules.BeamDamageRangeModifiers[rangeIndex], effectiveRoll, threshold, spec.Tactical.Rules.BeamEffectiveRollCap)
			totalDamage += damage
			selectionRoll, err = r.random(100, spec.Tactical.Rules)
			if err != nil {
				return nil, err
			}
			remainingDamage := damage
			if targetState.ArmorCurrent > 0 {
				absorbed := remainingDamage
				if absorbed > targetState.ArmorCurrent {
					absorbed = targetState.ArmorCurrent
				}
				targetState.ArmorCurrent -= absorbed
				remainingDamage -= absorbed
				afterArmor = targetState.ArmorCurrent
			}
			if remainingDamage > 0 {
				// Internal subsystem damage remains deferred. Grouped volleys apply
				// each physical weapon sequentially to aggregate Armor/Structure.
				targetState.StructureDamage += remainingDamage
				if targetState.StructureDamage > targetSpec.StructureMax {
					targetState.StructureDamage = targetSpec.StructureMax
				}
				afterStructure = targetState.StructureDamage
			}
			switch {
			case beforeArmor > 0 && afterStructure > beforeStructure:
				layer = "armor_structure"
			case beforeArmor > afterArmor:
				layer = "armor"
			case afterStructure > beforeStructure:
				layer = "structure"
			case targetState.StructureDamage >= targetSpec.StructureMax:
				layer = "structure_overkill"
			}
		}
		if err := r.appendEvent("beam_fired", seatID, commandSequence, map[string]any{
			"ship_id": active.ShipID, "target_ship_id": targetSpec.ShipID, "weapon_slot": payload.WeaponSlot,
			"shot_index": shotIndex + 1, "shot_count": weapon.Count,
			"range_index": rangeIndex, "raw_hit_roll": rawRoll, "effective_hit_roll": effectiveRoll,
			"threshold": threshold, "hit": hit, "damage": damage,
		}); err != nil {
			return nil, err
		}
		if hit {
			if err := r.appendEvent("battle_damage_applied", seatID, commandSequence, map[string]any{
				"target_ship_id": targetSpec.ShipID, "weapon_slot": payload.WeaponSlot,
				"shot_index": shotIndex + 1, "shot_count": weapon.Count,
				"layer": layer, "damage": damage, "selection_roll": selectionRoll,
				"armor_before": beforeArmor, "armor_after": afterArmor,
				"structure_damage_before": beforeStructure, "structure_damage_after": afterStructure,
			}); err != nil {
				return nil, err
			}
		}
	}
	ready.Ready = false
	if weapon.Count > 1 {
		if err := r.appendEvent("beam_volley_resolved", seatID, commandSequence, map[string]any{
			"ship_id": active.ShipID, "target_ship_id": targetSpec.ShipID, "weapon_slot": payload.WeaponSlot,
			"shot_count": weapon.Count, "hit_count": hitCount, "total_damage": totalDamage,
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
	if seatID == 0 || len(r.state.InitiativeOrder) < 2 {
		return nil
	}
	start := 0
	for i, id := range r.state.InitiativeOrder {
		if id == activeShipID {
			start = (i + 1) % len(r.state.InitiativeOrder)
			break
		}
	}
	out := make([]core.ID, 0)
	for offset := 0; offset < len(r.state.InitiativeOrder); offset++ {
		id := r.state.InitiativeOrder[(start+offset)%len(r.state.InitiativeOrder)]
		if id == activeShipID {
			continue
		}
		state := r.shipState(id)
		ship := tacticalShipSpec(spec, id)
		if state == nil || ship == nil || state.Destroyed || state.ActivationComplete || ship.SeatID != seatID {
			continue
		}
		out = append(out, id)
	}
	return out
}

func nextUnfinishedShipAfter(r *tacticalRuntime, activeShipID core.ID) core.ID {
	if len(r.state.InitiativeOrder) == 0 {
		return 0
	}
	start := 0
	for i, id := range r.state.InitiativeOrder {
		if id == activeShipID {
			start = (i + 1) % len(r.state.InitiativeOrder)
			break
		}
	}
	for offset := 0; offset < len(r.state.InitiativeOrder); offset++ {
		id := r.state.InitiativeOrder[(start+offset)%len(r.state.InitiativeOrder)]
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
	if active == nil {
		return fmt.Errorf("active tactical ship %d is missing from spec", r.state.ActiveShipID)
	}
	if seatID == 0 || active.SeatID != seatID {
		return fmt.Errorf("seat %d does not control active ship %d", seatID, r.state.ActiveShipID)
	}
	if payload.ShipID != r.state.ActiveShipID {
		return fmt.Errorf("wait ship %d is not active ship %d", payload.ShipID, r.state.ActiveShipID)
	}
	activeState := r.shipState(payload.ShipID)
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
	if err := r.appendEvent("activation_auto_ended", active.SeatID, commandSequence, map[string]any{
		"ship_id": active.ShipID, "round": r.state.Round, "reason": "resources_exhausted",
	}); err != nil {
		return err
	}
	return advanceAfterActivationCompletion(spec, r)
}

func advanceAfterActivationCompletion(spec Spec, r *tacticalRuntime) error {
	if nextShipID := nextUnfinishedShipAfter(r, r.state.ActiveShipID); nextShipID != 0 {
		r.state.ActiveShipID = nextShipID
		return nil
	}
	r.state.Round++
	for i := range r.state.Ships {
		if r.state.Ships[i].Destroyed {
			continue
		}
		r.state.Ships[i].ActivationComplete = false
		r.state.Ships[i].MovementCurrent = r.state.Ships[i].MovementMax
		for j := range r.state.Ships[i].Weapons {
			r.state.Ships[i].Weapons[j].Ready = true
		}
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
	if active == nil {
		return fmt.Errorf("active tactical ship %d is missing from spec", r.state.ActiveShipID)
	}
	if seatID == 0 || active.SeatID != seatID {
		return fmt.Errorf("seat %d does not control active ship %d", seatID, r.state.ActiveShipID)
	}
	if payload.ShipID != r.state.ActiveShipID {
		return fmt.Errorf("end-activation ship %d is not active ship %d", payload.ShipID, r.state.ActiveShipID)
	}
	activeState := r.shipState(payload.ShipID)
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

func beamDamage(weapon TacticalWeaponSpec, rangeModifier, effectiveRoll, threshold, rollCap int) int {
	multiplier := 100 + rangeModifier
	minDamage := weapon.MinDamage * multiplier / 100
	maxDamage := weapon.MaxDamage * multiplier / 100
	if maxDamage < minDamage {
		maxDamage = minDamage
	}
	if maxDamage <= minDamage || rollCap <= threshold {
		return minDamage
	}
	damage := minDamage + ((effectiveRoll-threshold)*(maxDamage-minDamage+1))/(rollCap-threshold)
	if damage > maxDamage {
		damage = maxDamage
	}
	if damage < minDamage {
		damage = minDamage
	}
	return damage
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
			seatAlive[ship.SeatID] = true
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
	out.Ships = make([]TacticalShipSpec, len(spec.Ships))
	for i := range spec.Ships {
		out.Ships[i] = spec.Ships[i]
		out.Ships[i].Weapons = append([]TacticalWeaponSpec(nil), spec.Ships[i].Weapons...)
		if spec.Ships[i].VisualGenome != nil {
			genome := core.CloneShipVisualGenome(*spec.Ships[i].VisualGenome)
			out.Ships[i].VisualGenome = &genome
		}
	}
	return out
}

func cloneTacticalState(state TacticalState) TacticalState {
	out := state
	out.InitiativeOrder = append([]core.ID(nil), state.InitiativeOrder...)
	out.Ships = make([]TacticalShipState, len(state.Ships))
	for i := range state.Ships {
		out.Ships[i] = state.Ships[i]
		out.Ships[i].Weapons = append([]TacticalWeaponState(nil), state.Ships[i].Weapons...)
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
	return tacticalRuntime{state: cloneTacticalState(r.state), events: cloneTacticalEvents(r.events), nextEventSequence: r.nextEventSequence}
}

// CloneTacticalSpec returns a detached immutable tactical battle snapshot.
func CloneTacticalSpec(spec TacticalSpec) TacticalSpec {
	return cloneTacticalSpec(spec)
}
