package server

import (
	"errors"
	"io"
	"net/http"
)

const maxLiveSnapshotBodyBytes int64 = 64 << 20

func (s *apiServer) handleLiveSnapshotExport(w http.ResponseWriter, r *http.Request) {
	if !s.persistenceEnabled {
		writeAPIError(w, http.StatusForbidden, "forbidden", "persistence API is disabled")
		return
	}
	data, err := s.host.ExportLiveSnapshot(r.PathValue("gameID"))
	if err != nil {
		writeHostError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (s *apiServer) handleLiveSnapshotImport(w http.ResponseWriter, r *http.Request) {
	if !s.persistenceEnabled {
		writeAPIError(w, http.StatusForbidden, "forbidden", "persistence API is disabled")
		return
	}
	data, ok := readLiveSnapshotBody(w, r)
	if !ok {
		return
	}
	summary, err := s.host.ImportLiveSnapshot(data)
	if err != nil {
		writeHostError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, summary)
}

func (s *apiServer) handleLiveSnapshotRestore(w http.ResponseWriter, r *http.Request) {
	if !s.persistenceEnabled {
		writeAPIError(w, http.StatusForbidden, "forbidden", "persistence API is disabled")
		return
	}
	data, ok := readLiveSnapshotBody(w, r)
	if !ok {
		return
	}
	receipt, err := s.host.RestoreLiveSnapshot(r.PathValue("gameID"), data)
	if err != nil {
		writeHostError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, receipt)
}

func readLiveSnapshotBody(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxLiveSnapshotBodyBytes)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeAPIError(w, http.StatusRequestEntityTooLarge, "save_too_large", "live snapshot exceeds 64 MiB limit")
			return nil, false
		}
		writeAPIError(w, http.StatusBadRequest, "invalid_save", "read live snapshot: "+err.Error())
		return nil, false
	}
	if len(data) == 0 {
		writeAPIError(w, http.StatusBadRequest, "invalid_save", "live snapshot body is empty")
		return nil, false
	}
	return data, true
}
