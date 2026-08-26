package main

import (
	"errors"
	"flag"
	"fmt"

	"moox/internal/palettecatalog"
)

func palettesCmd(args []string) error {
	fs := flag.NewFlagSet("palettes", flag.ContinueOnError)
	out := fs.String("out", "", "output directory for palette JSON/swatches (required)")
	clean := fs.Bool("clean", false, "remove the output directory before rebuilding it")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("palettes requires exactly one MOO2 installation directory")
	}
	if *out == "" {
		return errors.New("palettes requires -out <directory>")
	}
	manifest, err := palettecatalog.Build(fs.Arg(0), *out, palettecatalog.Options{Clean: *clean})
	if err != nil {
		return err
	}
	fmt.Printf("exported %d external MOO2 palettes -> %s\n", manifest.PaletteCount, *out)
	return nil
}
