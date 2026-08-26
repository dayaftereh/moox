package main

import (
	"errors"
	"flag"
	"fmt"
	"sort"

	"moox/internal/blockcatalog"
)

func classifyCmd(args []string) error {
	fs := flag.NewFlagSet("classify", flag.ContinueOnError)
	out := fs.String("out", "", "output JSON catalog path (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("classify requires exactly one MOO2 installation directory")
	}
	if *out == "" {
		return errors.New("classify requires -out <file.json>")
	}
	manifest, err := blockcatalog.Build(fs.Arg(0))
	if err != nil {
		return err
	}
	if err := blockcatalog.WriteJSON(*out, manifest); err != nil {
		return err
	}
	keys := make([]string, 0, len(manifest.Classes))
	for key := range manifest.Classes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	fmt.Printf("classified %d LBX blocks\n", manifest.BlockCount)
	for _, key := range keys {
		fmt.Printf("  %-24s %d\n", key, manifest.Classes[key])
	}
	fmt.Printf("-> %s\n", *out)
	return nil
}
