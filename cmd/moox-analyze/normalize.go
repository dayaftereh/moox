package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"moox/internal/moo2data"
)

func normalizeCmd(args []string) error {
	if len(args) == 0 {
		return errors.New("normalize requires a dataset name; currently supported: race-traits")
	}
	switch args[0] {
	case "race-traits":
		return normalizeRaceTraitsCmd(args[1:])
	default:
		return fmt.Errorf("unknown normalize dataset %q; currently supported: race-traits", args[0])
	}
}

func normalizeRaceTraitsCmd(args []string) error {
	fs := flag.NewFlagSet("normalize race-traits", flag.ContinueOnError)
	out := fs.String("out", "", "output JSON path (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("normalize race-traits requires exactly one MOO2 installation directory")
	}
	if *out == "" {
		return errors.New("normalize race-traits requires -out <path>")
	}

	data, err := moo2data.DecodeRaceTraits(fs.Arg(0))
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(*out)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(abs), ".race-traits-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	committed := false
	defer func() {
		tmp.Close()
		if !committed {
			_ = os.Remove(tmpPath)
		}
	}()

	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
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

	options := 0
	for _, group := range data.Groups {
		options += len(group.Options)
	}
	fmt.Printf("normalized %d race-design groups / %d options -> %s\n", len(data.Groups), options, abs)
	return nil
}
