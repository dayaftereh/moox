package main

import (
	"errors"
	"flag"
	"fmt"

	"moox/internal/supportcatalog"
)

func supportCmd(args []string) error {
	fs := flag.NewFlagSet("support", flag.ContinueOnError)
	out := fs.String("out", "", "output directory for non-LBX reference files (required)")
	clean := fs.Bool("clean", false, "remove output directory before rebuilding")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("support requires exactly one MOO2 installation directory")
	}
	if *out == "" {
		return errors.New("support requires -out <directory>")
	}
	manifest, err := supportcatalog.Build(fs.Arg(0), *out, supportcatalog.Options{Clean: *clean})
	if err != nil {
		return err
	}
	fmt.Printf("copied %d non-LBX reference files, %.2f MiB -> %s\n", manifest.FileCount, float64(manifest.TotalBytes)/(1024*1024), *out)
	return nil
}
