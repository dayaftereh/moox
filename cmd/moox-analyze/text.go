package main

import (
	"errors"
	"flag"
	"fmt"

	"moox/internal/textcatalog"
)

func textCmd(args []string) error {
	fs := flag.NewFlagSet("text", flag.ContinueOnError)
	out := fs.String("out", "", "output directory for private text catalog/previews (required)")
	clean := fs.Bool("clean", false, "remove output directory before rebuilding")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("text requires exactly one MOO2 installation directory")
	}
	if *out == "" {
		return errors.New("text requires -out <directory>")
	}
	manifest, err := textcatalog.Build(fs.Arg(0), *out, textcatalog.Options{Clean: *clean})
	if err != nil {
		return err
	}
	fmt.Printf("cataloged %d fixed records + %d string-table candidates with %d ASCII runs -> %s\n", manifest.FixedRecords, manifest.StringTableCandidates, manifest.ASCIIRuns, *out)
	return nil
}
