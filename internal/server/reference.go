package server

import (
	"net/http"

	"moox/internal/app"
)

func (s *apiServer) handleReferenceAdvance(w http.ResponseWriter, r *http.Request) {
	if !s.referenceControlsEnabled {
		http.NotFound(w, r)
		return
	}
	if !validateMutationRequest(w, r) {
		return
	}
	var request app.ReferenceAdvanceRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	if request.SchemaVersion != app.SchemaVersion {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "unsupported schema_version")
		return
	}
	if request.SeatID == 0 || request.BaseRevision == 0 {
		writeAPIError(w, http.StatusBadRequest, "bad_request", "seat_id and base_revision must be non-zero")
		return
	}
	switch request.Mode {
	case app.ReferenceAdvanceTurns:
		if request.Turns < 1 || request.Turns > app.ReferenceMaxTurnsPerRequest || request.ColonyID != 0 {
			writeAPIError(w, http.StatusBadRequest, "bad_request", "turns mode requires turns within [1,25] and no colony_id")
			return
		}
	case app.ReferenceAdvanceUntilConstruction:
		if request.Turns != 0 || request.ColonyID == 0 {
			writeAPIError(w, http.StatusBadRequest, "bad_request", "until_construction_complete requires colony_id and no turns")
			return
		}
	default:
		writeAPIError(w, http.StatusBadRequest, "bad_request", "unsupported reference advance mode")
		return
	}
	result, err := s.host.AdvanceReference(r.PathValue("gameID"), request)
	if err != nil {
		writeHostError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
