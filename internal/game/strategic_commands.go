package game

import (
	"fmt"

	"moox/internal/core"
	"moox/internal/protocol"
)

const CommandMoveFleet = "empire.move_fleet"
const CommandSplitFleet = "empire.split_fleet"
const CommandMergeFleets = "empire.merge_fleets"
const CommandColonizePlanet = "empire.colonize_planet"
const CommandDeployOutpost = "fleet.deploy_outpost"

type MoveFleetPayload struct {
	FleetID             core.ID   `json:"fleet_id"`
	DestinationSystemID core.ID   `json:"destination_system_id"`
	ShipIDs             []core.ID `json:"ship_ids,omitempty"`
}

func NewMoveFleetCommand(sequence uint32, payload MoveFleetPayload) (protocol.Command, error) {
	if err := validateMoveFleetPayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandMoveFleet, payload)
}

func decodeMoveFleet(command protocol.Command) (MoveFleetPayload, error) {
	var payload MoveFleetPayload
	if err := decodeStrictCommandPayload(command, CommandMoveFleet, &payload); err != nil {
		return MoveFleetPayload{}, err
	}
	if err := validateMoveFleetPayload(payload); err != nil {
		return MoveFleetPayload{}, err
	}
	return payload, nil
}

func validateMoveFleetPayload(payload MoveFleetPayload) error {
	if payload.FleetID == 0 {
		return fmt.Errorf("fleet_id must be non-zero")
	}
	if payload.DestinationSystemID == 0 {
		return fmt.Errorf("destination_system_id must be non-zero")
	}
	if err := validateSortedUniqueShipIDs(payload.ShipIDs, "ship_ids", false); err != nil {
		return err
	}
	return nil
}

type SplitFleetPayload struct {
	FleetID core.ID   `json:"fleet_id"`
	ShipIDs []core.ID `json:"ship_ids"`
}

func NewSplitFleetCommand(sequence uint32, payload SplitFleetPayload) (protocol.Command, error) {
	if err := validateSplitFleetPayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandSplitFleet, payload)
}

func decodeSplitFleet(command protocol.Command) (SplitFleetPayload, error) {
	var payload SplitFleetPayload
	if err := decodeStrictCommandPayload(command, CommandSplitFleet, &payload); err != nil {
		return SplitFleetPayload{}, err
	}
	if err := validateSplitFleetPayload(payload); err != nil {
		return SplitFleetPayload{}, err
	}
	return payload, nil
}

func validateSplitFleetPayload(payload SplitFleetPayload) error {
	if payload.FleetID == 0 {
		return fmt.Errorf("fleet_id must be non-zero")
	}
	return validateSortedUniqueShipIDs(payload.ShipIDs, "ship_ids", true)
}

type MergeFleetsPayload struct {
	TargetFleetID core.ID `json:"target_fleet_id"`
	SourceFleetID core.ID `json:"source_fleet_id"`
}

func NewMergeFleetsCommand(sequence uint32, payload MergeFleetsPayload) (protocol.Command, error) {
	if err := validateMergeFleetsPayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandMergeFleets, payload)
}

func decodeMergeFleets(command protocol.Command) (MergeFleetsPayload, error) {
	var payload MergeFleetsPayload
	if err := decodeStrictCommandPayload(command, CommandMergeFleets, &payload); err != nil {
		return MergeFleetsPayload{}, err
	}
	if err := validateMergeFleetsPayload(payload); err != nil {
		return MergeFleetsPayload{}, err
	}
	return payload, nil
}

func validateMergeFleetsPayload(payload MergeFleetsPayload) error {
	if payload.TargetFleetID == 0 {
		return fmt.Errorf("target_fleet_id must be non-zero")
	}
	if payload.SourceFleetID == 0 {
		return fmt.Errorf("source_fleet_id must be non-zero")
	}
	if payload.TargetFleetID == payload.SourceFleetID {
		return fmt.Errorf("target_fleet_id and source_fleet_id must differ")
	}
	return nil
}

func validateSortedUniqueShipIDs(ids []core.ID, label string, requireNonEmpty bool) error {
	if requireNonEmpty && len(ids) == 0 {
		return fmt.Errorf("%s must not be empty", label)
	}
	last := core.ID(0)
	for i, id := range ids {
		if id == 0 {
			return fmt.Errorf("%s[%d] must be non-zero", label, i)
		}
		if i > 0 && id <= last {
			return fmt.Errorf("%s must be strictly ascending", label)
		}
		last = id
	}
	return nil
}

type ColonizePlanetPayload struct {
	FleetID  core.ID `json:"fleet_id"`
	PlanetID core.ID `json:"planet_id"`
}

func NewColonizePlanetCommand(sequence uint32, payload ColonizePlanetPayload) (protocol.Command, error) {
	if err := validateColonizePlanetPayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandColonizePlanet, payload)
}

func decodeColonizePlanet(command protocol.Command) (ColonizePlanetPayload, error) {
	var payload ColonizePlanetPayload
	if err := decodeStrictCommandPayload(command, CommandColonizePlanet, &payload); err != nil {
		return ColonizePlanetPayload{}, err
	}
	if err := validateColonizePlanetPayload(payload); err != nil {
		return ColonizePlanetPayload{}, err
	}
	return payload, nil
}

func validateColonizePlanetPayload(payload ColonizePlanetPayload) error {
	if payload.FleetID == 0 {
		return fmt.Errorf("fleet_id must be non-zero")
	}
	if payload.PlanetID == 0 {
		return fmt.Errorf("planet_id must be non-zero")
	}
	return nil
}

type DeployOutpostPayload struct {
	FleetID  core.ID `json:"fleet_id"`
	BodyID   core.ID `json:"body_id,omitempty"`
	PlanetID core.ID `json:"planet_id,omitempty"` // legacy/normal-Planet compatibility
}

func (p DeployOutpostPayload) TargetBodyID() core.ID {
	if p.BodyID != 0 {
		return p.BodyID
	}
	return p.PlanetID
}

func NewDeployOutpostCommand(sequence uint32, payload DeployOutpostPayload) (protocol.Command, error) {
	if err := validateDeployOutpostPayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandDeployOutpost, payload)
}

func decodeDeployOutpost(command protocol.Command) (DeployOutpostPayload, error) {
	var payload DeployOutpostPayload
	if err := decodeStrictCommandPayload(command, CommandDeployOutpost, &payload); err != nil {
		return DeployOutpostPayload{}, err
	}
	if err := validateDeployOutpostPayload(payload); err != nil {
		return DeployOutpostPayload{}, err
	}
	return payload, nil
}

func validateDeployOutpostPayload(payload DeployOutpostPayload) error {
	if payload.FleetID == 0 {
		return fmt.Errorf("fleet_id must be non-zero")
	}
	if payload.BodyID == 0 && payload.PlanetID == 0 {
		return fmt.Errorf("body_id or legacy planet_id must be non-zero")
	}
	if payload.BodyID != 0 && payload.PlanetID != 0 && payload.BodyID != payload.PlanetID {
		return fmt.Errorf("body_id %d and planet_id %d must refer to the same normal-Planet body when both are provided", payload.BodyID, payload.PlanetID)
	}
	return nil
}
