package core

type PopulationJob string

const (
	PopulationJobFarmer    PopulationJob = "farmer"
	PopulationJobWorker    PopulationJob = "worker"
	PopulationJobScientist PopulationJob = "scientist"
)

type PopulationTransfer struct {
	ID                  ID            `json:"id"`
	EmpireID            ID            `json:"empire_id"`
	SourceColonyID      ID            `json:"source_colony_id"`
	DestinationColonyID ID            `json:"destination_colony_id"`
	Job                 PopulationJob `json:"job"`
	RemainingTurns      int           `json:"remaining_turns"`
}
