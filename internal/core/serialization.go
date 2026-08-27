package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func MarshalState(state *GameState) ([]byte, error) {
	if state == nil {
		return nil, fmt.Errorf("state is nil")
	}
	if err := state.Validate(); err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(state); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func UnmarshalState(data []byte) (*GameState, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var state GameState
	if err := decoder.Decode(&state); err != nil {
		return nil, err
	}
	if err := state.Validate(); err != nil {
		return nil, err
	}
	return &state, nil
}

func SaveState(path string, state *GameState) error {
	data, err := MarshalState(state)
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(abs), ".moox-state-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	committed := false
	defer func() {
		_ = tmp.Close()
		if !committed {
			_ = os.Remove(tmpPath)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, abs); err != nil {
		_ = os.Remove(abs)
		if err := os.Rename(tmpPath, abs); err != nil {
			return err
		}
	}
	committed = true
	return nil
}

func LoadState(path string) (*GameState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return UnmarshalState(data)
}
