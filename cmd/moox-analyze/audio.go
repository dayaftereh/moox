package main

import (
	"errors"
	"flag"
	"fmt"

	"moox/internal/audiocatalog"
)

func audioCmd(args []string) error {
	fs := flag.NewFlagSet("audio", flag.ContinueOnError)
	out := fs.String("out", "", "output directory for verified WAVE blocks (required)")
	clean := fs.Bool("clean", false, "remove output directory before rebuilding")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("audio requires exactly one MOO2 installation directory")
	}
	if *out == "" {
		return errors.New("audio requires -out <directory>")
	}
	manifest, err := audiocatalog.Build(fs.Arg(0), *out, audiocatalog.Options{Clean: *clean})
	if err != nil {
		return err
	}
	fmt.Printf("exported %d verified RIFF/WAVE blocks, %.2f MiB -> %s\n", manifest.WAVCount, float64(manifest.TotalBytes)/(1024*1024), *out)
	return nil
}
