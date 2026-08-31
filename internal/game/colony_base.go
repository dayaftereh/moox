package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"moox/internal/core"
	"moox/internal/protocol"
)

const (
	ColonyBaseBuildingID    = "colony_base"
	ColonyBaseProductionID  = 11
	ColonyBaseTechnologyID  = 40
	ColonyBaseBaseCostPP    = 200.0
	ColonyBaseTrashRefundBC = 100.0
	CommandColonizeWithBase = "colony.colonize_with_base"
	CommandTrashColonyBase  = "colony.trash_colony_base"
)

type ColonyBaseResolution struct {
	EmpireID        core.ID   `json:"empire_id"`
	SourceColonyID  core.ID   `json:"source_colony_id"`
	SystemID        core.ID   `json:"system_id"`
	TargetPlanetIDs []core.ID `json:"target_planet_ids"`
	TrashRefundBC   float64   `json:"trash_refund_bc"`
}

type ColonizeWithBasePayload struct {
	SourceColonyID core.ID `json:"source_colony_id"`
	PlanetID       core.ID `json:"planet_id"`
}

type TrashColonyBasePayload struct {
	SourceColonyID core.ID `json:"source_colony_id"`
}

type ColonyBaseColonizedEvent struct {
	EmpireID       core.ID `json:"empire_id"`
	SourceColonyID core.ID `json:"source_colony_id"`
	SystemID       core.ID `json:"system_id"`
	PlanetID       core.ID `json:"planet_id"`
	ColonyID       core.ID `json:"colony_id"`
	BuildingID     string  `json:"building_id"`
}

type ColonyBaseTrashedEvent struct {
	EmpireID          core.ID `json:"empire_id"`
	SourceColonyID    core.ID `json:"source_colony_id"`
	SystemID          core.ID `json:"system_id"`
	BuildingID        string  `json:"building_id"`
	RefundBC          float64 `json:"refund_bc"`
	PreviousBalanceBC float64 `json:"previous_balance_bc"`
	CurrentBalanceBC  float64 `json:"current_balance_bc"`
}

func NewColonizeWithBaseCommand(sequence uint32, payload ColonizeWithBasePayload) (protocol.Command, error) {
	if err := validateColonizeWithBasePayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandColonizeWithBase, payload)
}

func NewTrashColonyBaseCommand(sequence uint32, payload TrashColonyBasePayload) (protocol.Command, error) {
	if err := validateTrashColonyBasePayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandTrashColonyBase, payload)
}

func decodeColonizeWithBase(command protocol.Command) (ColonizeWithBasePayload, error) {
	var payload ColonizeWithBasePayload
	if err := decodeStrictCommandPayload(command, CommandColonizeWithBase, &payload); err != nil {
		return ColonizeWithBasePayload{}, err
	}
	if err := validateColonizeWithBasePayload(payload); err != nil {
		return ColonizeWithBasePayload{}, err
	}
	return payload, nil
}

func decodeTrashColonyBase(command protocol.Command) (TrashColonyBasePayload, error) {
	var payload TrashColonyBasePayload
	if err := decodeStrictCommandPayload(command, CommandTrashColonyBase, &payload); err != nil {
		return TrashColonyBasePayload{}, err
	}
	if err := validateTrashColonyBasePayload(payload); err != nil {
		return TrashColonyBasePayload{}, err
	}
	return payload, nil
}

func decodeStrictCommandPayload(command protocol.Command, kind string, dst any) error {
	if command.Kind != kind {
		return fmt.Errorf("command kind %q, expected %q", command.Kind, kind)
	}
	decoder := json.NewDecoder(bytes.NewReader(command.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("decode %s: %w", kind, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("decode %s: trailing JSON value", kind)
		}
		return fmt.Errorf("decode %s trailing data: %w", kind, err)
	}
	return nil
}

func validateColonizeWithBasePayload(payload ColonizeWithBasePayload) error {
	if payload.SourceColonyID == 0 {
		return fmt.Errorf("source_colony_id must be non-zero")
	}
	if payload.PlanetID == 0 {
		return fmt.Errorf("planet_id must be non-zero")
	}
	return nil
}

func validateTrashColonyBasePayload(payload TrashColonyBasePayload) error {
	if payload.SourceColonyID == 0 {
		return fmt.Errorf("source_colony_id must be non-zero")
	}
	return nil
}

// PendingColonyBaseResolutions projects completed Colony Bases that must be
// resolved at the post-Production boundary. empireID==0 returns all empires.
func PendingColonyBaseResolutions(state *core.GameState, empireID core.ID) ([]ColonyBaseResolution, error) {
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}
	if empireID != 0 && empireByID(state, empireID) == nil {
		return nil, fmt.Errorf("unknown empire %d", empireID)
	}

	resolutions := make([]ColonyBaseResolution, 0)
	for i := range state.Colonies {
		colony := &state.Colonies[i]
		if empireID != 0 && colony.EmpireID != empireID {
			continue
		}
		if !colonyOwnsBuilding(colony, ColonyBaseBuildingID) {
			continue
		}
		system := systemForPlanetID(state, colony.PlanetID)
		if system == nil {
			return nil, fmt.Errorf("colony %d planet %d is not assigned to a star system", colony.ID, colony.PlanetID)
		}
		targets, err := colonyBaseTargetPlanetIDs(state, colony)
		if err != nil {
			return nil, err
		}
		resolutions = append(resolutions, ColonyBaseResolution{
			EmpireID:        colony.EmpireID,
			SourceColonyID:  colony.ID,
			SystemID:        system.ID,
			TargetPlanetIDs: targets,
			TrashRefundBC:   ColonyBaseTrashRefundBC,
		})
	}
	sort.Slice(resolutions, func(i, j int) bool {
		if resolutions[i].EmpireID != resolutions[j].EmpireID {
			return resolutions[i].EmpireID < resolutions[j].EmpireID
		}
		return resolutions[i].SourceColonyID < resolutions[j].SourceColonyID
	})
	return resolutions, nil
}

func colonyBaseTargetPlanetIDs(state *core.GameState, source *core.Colony) ([]core.ID, error) {
	if state == nil || source == nil {
		return nil, fmt.Errorf("state and source colony must not be nil")
	}
	system := systemForPlanetID(state, source.PlanetID)
	if system == nil {
		return nil, fmt.Errorf("colony %d planet %d is not assigned to a star system", source.ID, source.PlanetID)
	}
	targets := make([]core.ID, 0, len(system.Planets))
	for i := range system.Planets {
		planet := &system.Planets[i]
		if planet.ColonyID != 0 || colonyReferencesPlanet(state, planet.ID) {
			continue
		}
		targets = append(targets, planet.ID)
	}
	return targets, nil
}

func colonyReferencesPlanet(state *core.GameState, planetID core.ID) bool {
	for i := range state.Colonies {
		if state.Colonies[i].PlanetID == planetID {
			return true
		}
	}
	return false
}

func colonyOwnsBuilding(colony *core.Colony, buildingID string) bool {
	if colony == nil {
		return false
	}
	for _, owned := range colony.Buildings {
		if owned == buildingID {
			return true
		}
	}
	return false
}

func removeColonyBuilding(colony *core.Colony, buildingID string) bool {
	if colony == nil {
		return false
	}
	for i, owned := range colony.Buildings {
		if owned != buildingID {
			continue
		}
		colony.Buildings = append(colony.Buildings[:i], colony.Buildings[i+1:]...)
		return true
	}
	return false
}

func (r *EconomyResolver) ResolveColonyBaseCommand(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) ([]DomainEvent, error) {
	if r == nil || r.Rules == nil {
		return nil, fmt.Errorf("economy resolver has no rules")
	}
	if state == nil {
		return nil, fmt.Errorf("game state must not be nil")
	}
	if err := command.Validate(command.Sequence); err != nil {
		return nil, err
	}
	switch command.Kind {
	case CommandColonizeWithBase:
		return r.colonizeWithBase(state, empireID, seatID, command)
	case CommandTrashColonyBase:
		event, err := r.trashColonyBase(state, empireID, seatID, command)
		if err != nil {
			return nil, err
		}
		return []DomainEvent{event}, nil
	default:
		return nil, fmt.Errorf("unsupported Colony Base command kind %q", command.Kind)
	}
}

func (r *EconomyResolver) colonizeWithBase(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) ([]DomainEvent, error) {
	payload, err := decodeColonizeWithBase(command)
	if err != nil {
		return nil, err
	}
	source := colonyByID(state, payload.SourceColonyID)
	if source == nil {
		return nil, fmt.Errorf("references unknown source colony %d", payload.SourceColonyID)
	}
	if source.EmpireID != empireID {
		return nil, fmt.Errorf("seat %d cannot use Colony Base on colony %d owned by empire %d", seatID, source.ID, source.EmpireID)
	}
	if !colonyOwnsBuilding(source, ColonyBaseBuildingID) {
		return nil, fmt.Errorf("colony %d does not own %q", source.ID, ColonyBaseBuildingID)
	}
	sourceSystem := systemForPlanetID(state, source.PlanetID)
	if sourceSystem == nil {
		return nil, fmt.Errorf("source colony %d planet %d is not assigned to a star system", source.ID, source.PlanetID)
	}
	targets, err := colonyBaseTargetPlanetIDs(state, source)
	if err != nil {
		return nil, err
	}
	legal := false
	for _, id := range targets {
		if id == payload.PlanetID {
			legal = true
			break
		}
	}
	if !legal {
		return nil, fmt.Errorf("planet %d is not a legal same-system Colony Base target for colony %d", payload.PlanetID, source.ID)
	}
	newColony, planet, targetSystem, err := r.prepareFoundedColony(state, empireID, payload.PlanetID)
	if err != nil {
		return nil, err
	}
	if targetSystem.ID != sourceSystem.ID {
		return nil, fmt.Errorf("planet %d is in system %d, not source colony %d system %d", planet.ID, targetSystem.ID, source.ID, sourceSystem.ID)
	}
	generic, err := NewDomainEvent("empire.planet_colonized", seatID, command.Sequence, PlanetColonizedEvent{
		EmpireID: empireID, FleetID: 0, SystemID: targetSystem.ID, PlanetID: planet.ID, ColonyID: newColony.ID,
	})
	if err != nil {
		return nil, err
	}
	baseEvent, err := NewDomainEvent("colony.colony_base_colonized", seatID, command.Sequence, ColonyBaseColonizedEvent{
		EmpireID: empireID, SourceColonyID: source.ID, SystemID: targetSystem.ID, PlanetID: planet.ID, ColonyID: newColony.ID, BuildingID: ColonyBaseBuildingID,
	})
	if err != nil {
		return nil, err
	}
	if allocated := state.NewID(); allocated != newColony.ID {
		return nil, fmt.Errorf("allocated Colony ID %d does not match expected next_id %d", allocated, newColony.ID)
	}
	planet.ColonyID = newColony.ID
	if !removeColonyBuilding(source, ColonyBaseBuildingID) {
		return nil, fmt.Errorf("source colony %d lost %q during Colony Base colonization", source.ID, ColonyBaseBuildingID)
	}
	state.Colonies = append(state.Colonies, newColony)
	return []DomainEvent{generic, baseEvent}, nil
}

func (r *EconomyResolver) trashColonyBase(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeTrashColonyBase(command)
	if err != nil {
		return DomainEvent{}, err
	}
	source := colonyByID(state, payload.SourceColonyID)
	if source == nil {
		return DomainEvent{}, fmt.Errorf("references unknown source colony %d", payload.SourceColonyID)
	}
	if source.EmpireID != empireID {
		return DomainEvent{}, fmt.Errorf("seat %d cannot trash Colony Base on colony %d owned by empire %d", seatID, source.ID, source.EmpireID)
	}
	if !colonyOwnsBuilding(source, ColonyBaseBuildingID) {
		return DomainEvent{}, fmt.Errorf("colony %d does not own %q", source.ID, ColonyBaseBuildingID)
	}
	empire := empireByID(state, empireID)
	if empire == nil {
		return DomainEvent{}, fmt.Errorf("seat %d references unknown empire %d", seatID, empireID)
	}
	system := systemForPlanetID(state, source.PlanetID)
	if system == nil {
		return DomainEvent{}, fmt.Errorf("source colony %d planet %d is not assigned to a star system", source.ID, source.PlanetID)
	}
	previous := empire.Treasury.BalanceBC
	if !removeColonyBuilding(source, ColonyBaseBuildingID) {
		return DomainEvent{}, fmt.Errorf("source colony %d lost %q during trash resolution", source.ID, ColonyBaseBuildingID)
	}
	empire.Treasury.BalanceBC += ColonyBaseTrashRefundBC
	return NewDomainEvent("colony.colony_base_trashed", seatID, command.Sequence, ColonyBaseTrashedEvent{
		EmpireID: empireID, SourceColonyID: source.ID, SystemID: system.ID, BuildingID: ColonyBaseBuildingID,
		RefundBC: ColonyBaseTrashRefundBC, PreviousBalanceBC: previous, CurrentBalanceBC: empire.Treasury.BalanceBC,
	})
}

func (r *EconomyResolver) prepareFoundedColony(state *core.GameState, empireID, planetID core.ID) (core.Colony, *core.Planet, *core.StarSystem, error) {
	if empireByID(state, empireID) == nil {
		return core.Colony{}, nil, nil, fmt.Errorf("unknown empire %d", empireID)
	}
	planet := planetByID(state, planetID)
	if planet == nil {
		return core.Colony{}, nil, nil, fmt.Errorf("references unknown planet %d", planetID)
	}
	targetSystem := systemForPlanetID(state, planet.ID)
	if targetSystem == nil {
		return core.Colony{}, nil, nil, fmt.Errorf("planet %d is not assigned to a star system", planet.ID)
	}
	if planet.ColonyID != 0 {
		return core.Colony{}, nil, nil, fmt.Errorf("planet %d is already colonized by colony %d", planet.ID, planet.ColonyID)
	}
	if colonyReferencesPlanet(state, planet.ID) {
		return core.Colony{}, nil, nil, fmt.Errorf("planet %d is already referenced by a colony", planet.ID)
	}
	newColonyID := state.NextID
	if newColonyID == 0 {
		return core.Colony{}, nil, nil, fmt.Errorf("cannot allocate Colony ID from zero next_id")
	}
	newColony := core.Colony{
		ID:         newColonyID,
		EmpireID:   empireID,
		PlanetID:   planet.ID,
		Population: core.NewAssimilatedPopulation(empireID, 1, 0, 0),
	}
	if err := r.recalculateColony(state, &newColony); err != nil {
		return core.Colony{}, nil, nil, fmt.Errorf("initialize colony %d on planet %d: %w", newColonyID, planet.ID, err)
	}
	return newColony, planet, targetSystem, nil
}
