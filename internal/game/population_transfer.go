package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"

	"moox/internal/core"
	"moox/internal/protocol"
)

const (
	CommandTransferPopulation               = "colony.transfer_population"
	populationTransferFreighters            = 5
	maxActivePopulationTransfers            = 25
	populationTransferCoordinateUnitsParsec = 30.0
	maxPopulationTransferETA                = 15
)

type TransferPopulationPayload struct {
	SourceColonyID      core.ID            `json:"source_colony_id"`
	DestinationColonyID core.ID            `json:"destination_colony_id"`
	Job                 core.PopulationJob `json:"job"`
}

type PopulationTransferredEvent struct {
	EmpireID            core.ID            `json:"empire_id"`
	SourceColonyID      core.ID            `json:"source_colony_id"`
	DestinationColonyID core.ID            `json:"destination_colony_id"`
	Job                 core.PopulationJob `json:"job"`
	Amount              float64            `json:"amount"`
	SameSystem          bool               `json:"same_system"`
}

type PopulationTransferStartedEvent struct {
	TransferID          core.ID            `json:"transfer_id"`
	EmpireID            core.ID            `json:"empire_id"`
	SourceColonyID      core.ID            `json:"source_colony_id"`
	DestinationColonyID core.ID            `json:"destination_colony_id"`
	Job                 core.PopulationJob `json:"job"`
	Amount              float64            `json:"amount"`
	ETA                 int                `json:"eta"`
	FreightersReserved  int                `json:"freighters_reserved"`
}

type PopulationTransferProgressedEvent struct {
	TransferID     core.ID `json:"transfer_id"`
	EmpireID       core.ID `json:"empire_id"`
	RemainingTurns int     `json:"remaining_turns"`
}

type PopulationTransferArrivedEvent struct {
	TransferID          core.ID            `json:"transfer_id"`
	EmpireID            core.ID            `json:"empire_id"`
	SourceColonyID      core.ID            `json:"source_colony_id"`
	DestinationColonyID core.ID            `json:"destination_colony_id"`
	Job                 core.PopulationJob `json:"job"`
	Amount              float64            `json:"amount"`
	FreightersReleased  int                `json:"freighters_released"`
}

type PopulationTransferLostEvent struct {
	TransferID          core.ID            `json:"transfer_id"`
	EmpireID            core.ID            `json:"empire_id"`
	SourceColonyID      core.ID            `json:"source_colony_id"`
	DestinationColonyID core.ID            `json:"destination_colony_id"`
	Job                 core.PopulationJob `json:"job"`
	Amount              float64            `json:"amount"`
	Reason              string             `json:"reason"`
	FreightersReleased  int                `json:"freighters_released"`
}

func NewTransferPopulationCommand(sequence uint32, payload TransferPopulationPayload) (protocol.Command, error) {
	if err := validateTransferPopulationPayload(payload); err != nil {
		return protocol.Command{}, err
	}
	return protocol.NewCommand(sequence, CommandTransferPopulation, payload)
}

func decodeTransferPopulation(command protocol.Command) (TransferPopulationPayload, error) {
	var payload TransferPopulationPayload
	decoder := json.NewDecoder(bytes.NewReader(command.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return TransferPopulationPayload{}, fmt.Errorf("decode %s: %w", CommandTransferPopulation, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return TransferPopulationPayload{}, fmt.Errorf("decode %s: trailing JSON value", CommandTransferPopulation)
		}
		return TransferPopulationPayload{}, fmt.Errorf("decode %s trailing data: %w", CommandTransferPopulation, err)
	}
	if err := validateTransferPopulationPayload(payload); err != nil {
		return TransferPopulationPayload{}, err
	}
	return payload, nil
}

func validateTransferPopulationPayload(payload TransferPopulationPayload) error {
	if payload.SourceColonyID == 0 || payload.DestinationColonyID == 0 {
		return fmt.Errorf("source_colony_id and destination_colony_id must be positive")
	}
	if payload.SourceColonyID == payload.DestinationColonyID {
		return fmt.Errorf("source_colony_id and destination_colony_id must differ")
	}
	switch payload.Job {
	case core.PopulationJobFarmer, core.PopulationJobWorker, core.PopulationJobScientist:
		return nil
	default:
		return fmt.Errorf("unsupported population job %q", payload.Job)
	}
}

func (r *EconomyResolver) transferPopulation(state *core.GameState, empireID core.ID, seatID protocol.SeatID, command protocol.Command) (DomainEvent, error) {
	payload, err := decodeTransferPopulation(command)
	if err != nil {
		return DomainEvent{}, err
	}
	source := colonyByID(state, payload.SourceColonyID)
	destination := colonyByID(state, payload.DestinationColonyID)
	if source == nil || destination == nil {
		return DomainEvent{}, fmt.Errorf("population transfer references unknown colony source=%d destination=%d", payload.SourceColonyID, payload.DestinationColonyID)
	}
	if source.EmpireID != empireID || destination.EmpireID != empireID {
		return DomainEvent{}, fmt.Errorf("empire %d must own both population-transfer colonies", empireID)
	}
	if source.Population.Total < 2-populationEpsilon {
		return DomainEvent{}, fmt.Errorf("colony %d must retain at least one Population unit", source.ID)
	}
	if populationJobValue(source.Population, payload.Job) < 1-populationEpsilon {
		return DomainEvent{}, fmt.Errorf("colony %d has less than one %s Population available", source.ID, payload.Job)
	}
	if err := r.recalculateColony(state, destination); err != nil {
		return DomainEvent{}, err
	}
	inbound := inboundPopulationTransfers(state, empireID, destination.ID)
	if destination.Population.Total+float64(inbound)+1 > destination.PopulationDynamics.Capacity+populationEpsilon {
		return DomainEvent{}, fmt.Errorf("colony %d has no capacity for another Population unit including inbound transfers", destination.ID)
	}

	sourceSystem := systemForPlanetID(state, source.PlanetID)
	destinationSystem := systemForPlanetID(state, destination.PlanetID)
	if sourceSystem == nil || destinationSystem == nil {
		return DomainEvent{}, fmt.Errorf("population-transfer colony is not attached to a star system")
	}

	if sourceSystem.ID == destinationSystem.ID {
		adjustPopulationJob(&source.Population, payload.Job, -1)
		adjustPopulationJob(&destination.Population, payload.Job, 1)
		return NewDomainEvent("colony.population_transferred", seatID, command.Sequence, PopulationTransferredEvent{
			EmpireID:            empireID,
			SourceColonyID:      source.ID,
			DestinationColonyID: destination.ID,
			Job:                 payload.Job,
			Amount:              1,
			SameSystem:          true,
		})
	}

	empire := empireByID(state, empireID)
	if empire == nil {
		return DomainEvent{}, fmt.Errorf("unknown empire %d", empireID)
	}
	active := activePopulationTransfers(state, empireID)
	if active >= maxActivePopulationTransfers {
		return DomainEvent{}, fmt.Errorf("empire %d already has the original maximum of %d active Population transfers", empireID, maxActivePopulationTransfers)
	}
	reservedAfterLaunch := (active + 1) * populationTransferFreighters
	if empire.Freighters < reservedAfterLaunch {
		return DomainEvent{}, fmt.Errorf("empire %d needs %d Freighters for Population transfers but owns %d", empireID, reservedAfterLaunch, empire.Freighters)
	}

	eta := populationTransferETA(*sourceSystem, *destinationSystem, r.Rules.populationTransferFTLSpeed(*empire))
	if eta < 1 {
		return DomainEvent{}, fmt.Errorf("invalid interstellar Population-transfer ETA %d", eta)
	}
	adjustPopulationJob(&source.Population, payload.Job, -1)
	transfer := core.PopulationTransfer{
		ID:                  state.NewID(),
		EmpireID:            empireID,
		SourceColonyID:      source.ID,
		DestinationColonyID: destination.ID,
		Job:                 payload.Job,
		RemainingTurns:      eta,
	}
	state.PopulationTransfers = append(state.PopulationTransfers, transfer)
	return NewDomainEvent("empire.population_transfer_started", seatID, command.Sequence, PopulationTransferStartedEvent{
		TransferID:          transfer.ID,
		EmpireID:            empireID,
		SourceColonyID:      source.ID,
		DestinationColonyID: destination.ID,
		Job:                 payload.Job,
		Amount:              1,
		ETA:                 eta,
		FreightersReserved:  populationTransferFreighters,
	})
}

func (r *EconomyResolver) advancePopulationTransfers(state *core.GameState) ([]DomainEvent, error) {
	if len(state.PopulationTransfers) == 0 {
		return nil, nil
	}
	kept := make([]core.PopulationTransfer, 0, len(state.PopulationTransfers))
	var events []DomainEvent
	for _, original := range state.PopulationTransfers {
		transfer := original
		transfer.RemainingTurns--
		if transfer.RemainingTurns > 0 {
			kept = append(kept, transfer)
			event, err := NewDomainEvent("empire.population_transfer_progressed", 0, 0, PopulationTransferProgressedEvent{
				TransferID:     transfer.ID,
				EmpireID:       transfer.EmpireID,
				RemainingTurns: transfer.RemainingTurns,
			})
			if err != nil {
				return nil, err
			}
			events = append(events, event)
			continue
		}

		destination := colonyByID(state, transfer.DestinationColonyID)
		reason := ""
		if destination == nil || destination.EmpireID != transfer.EmpireID {
			reason = "destination_unavailable"
		} else {
			system := systemForPlanetID(state, destination.PlanetID)
			if system == nil {
				reason = "destination_unavailable"
			} else if systemBlockadesEmpire(system, transfer.EmpireID) {
				reason = "destination_blockaded"
			} else {
				if err := r.recalculateColony(state, destination); err != nil {
					return nil, err
				}
				if destination.Population.Total+1 > destination.PopulationDynamics.Capacity+populationEpsilon {
					reason = "destination_full"
				}
			}
		}

		if reason != "" {
			event, err := NewDomainEvent("empire.population_transfer_lost", 0, 0, PopulationTransferLostEvent{
				TransferID:          transfer.ID,
				EmpireID:            transfer.EmpireID,
				SourceColonyID:      transfer.SourceColonyID,
				DestinationColonyID: transfer.DestinationColonyID,
				Job:                 transfer.Job,
				Amount:              1,
				Reason:              reason,
				FreightersReleased:  populationTransferFreighters,
			})
			if err != nil {
				return nil, err
			}
			events = append(events, event)
			continue
		}

		adjustPopulationJob(&destination.Population, transfer.Job, 1)
		event, err := NewDomainEvent("empire.population_transfer_arrived", 0, 0, PopulationTransferArrivedEvent{
			TransferID:          transfer.ID,
			EmpireID:            transfer.EmpireID,
			SourceColonyID:      transfer.SourceColonyID,
			DestinationColonyID: transfer.DestinationColonyID,
			Job:                 transfer.Job,
			Amount:              1,
			FreightersReleased:  populationTransferFreighters,
		})
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	state.PopulationTransfers = kept
	return events, nil
}

func populationJobValue(population core.PopulationState, job core.PopulationJob) float64 {
	switch job {
	case core.PopulationJobFarmer:
		return population.Farmers
	case core.PopulationJobWorker:
		return population.Workers
	case core.PopulationJobScientist:
		return population.Scientists
	default:
		return 0
	}
}

func adjustPopulationJob(population *core.PopulationState, job core.PopulationJob, delta float64) {
	population.Total += delta
	switch job {
	case core.PopulationJobFarmer:
		population.Farmers += delta
	case core.PopulationJobWorker:
		population.Workers += delta
	case core.PopulationJobScientist:
		population.Scientists += delta
	}
}

func activePopulationTransfers(state *core.GameState, empireID core.ID) int {
	count := 0
	for _, transfer := range state.PopulationTransfers {
		if transfer.EmpireID == empireID {
			count++
		}
	}
	return count
}

func inboundPopulationTransfers(state *core.GameState, empireID, colonyID core.ID) int {
	count := 0
	for _, transfer := range state.PopulationTransfers {
		if transfer.EmpireID == empireID && transfer.DestinationColonyID == colonyID {
			count++
		}
	}
	return count
}

func populationTransferFreightersReserved(state *core.GameState, empireID core.ID) int {
	return activePopulationTransfers(state, empireID) * populationTransferFreighters
}

func systemForPlanetID(state *core.GameState, planetID core.ID) *core.StarSystem {
	for si := range state.Galaxy.Systems {
		system := &state.Galaxy.Systems[si]
		for pi := range system.Planets {
			if system.Planets[pi].ID == planetID {
				return system
			}
		}
	}
	return nil
}

func populationTransferETA(source, destination core.StarSystem, ftlSpeed int) int {
	dx := float64(source.X - destination.X)
	dy := float64(source.Y - destination.Y)
	distance := math.Hypot(dx, dy)
	if distance <= 0 {
		return 0
	}
	parsecs := int(math.Ceil(distance / populationTransferCoordinateUnitsParsec))
	if ftlSpeed < 2 {
		ftlSpeed = 2
	}
	eta := (parsecs + ftlSpeed - 1) / ftlSpeed
	if eta > maxPopulationTransferETA {
		eta = maxPopulationTransferETA
	}
	if eta < 1 {
		eta = 1
	}
	return eta
}

func (r *EconomyRules) populationTransferFTLSpeed(empire core.Empire) int {
	known := make(map[int]struct{}, len(empire.KnownTechnologyIDs))
	for _, technologyID := range empire.KnownTechnologyIDs {
		known[technologyID] = struct{}{}
	}
	speed := 0
	for _, drive := range []struct {
		technologyID int
		speed        int
	}{
		{120, 2}, // Nuclear Drive
		{72, 3},  // Fusion Drive
		{96, 4},  // Ion Drive
		{11, 5},  // Anti-Matter Drive
		{88, 6},  // Hyper Drive
		{95, 7},  // Interphased Drive
	} {
		if _, ok := known[drive.technologyID]; ok && drive.speed > speed {
			speed = drive.speed
		}
	}
	if modifiers, ok := r.RaceModifiers[empire.RaceID]; ok && modifiers.TransDimensional {
		speed += 2
	}
	return speed
}
