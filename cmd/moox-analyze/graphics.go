package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"moox/internal/graphiccatalog"
)

func graphicsCmd(args []string) error {
	fs := flag.NewFlagSet("graphics", flag.ContinueOnError)
	out := fs.String("out", "", "output directory for manifest and PNG references (required)")
	manifestOnly := fs.Bool("manifest-only", false, "catalog graphics but do not write PNG frames")
	clean := fs.Bool("clean", false, "remove the output directory before rebuilding it")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("graphics requires exactly one MOO2 installation directory")
	}
	if *out == "" {
		return errors.New("graphics requires -out <directory>")
	}

	lastPrinted := 0
	manifest, err := graphiccatalog.Build(fs.Arg(0), *out, graphiccatalog.Options{
		ExportPNG: !*manifestOnly,
		Clean:     *clean,
		Progress: func(p graphiccatalog.Progress) {
			if p.ArchiveIndex-lastPrinted >= 20 || p.ArchiveIndex == p.ArchiveCount {
				fmt.Fprintf(os.Stderr, "graphics: %d/%d archives, %d blocks, %d frames exported\n", p.ArchiveIndex, p.ArchiveCount, p.Graphics, p.Frames)
				lastPrinted = p.ArchiveIndex
			}
		},
	})
	if err != nil {
		return err
	}

	fmt.Printf("cataloged %d graphic blocks across %d LBX archives\n", manifest.GraphicBlocks, manifest.LBXArchives)
	fmt.Printf("palette: %d internal, %d external; contexts: %d resolved, %d pending\n", manifest.InternalPalette, manifest.ExternalPalette, manifest.PaletteContextsResolved, manifest.PaletteContextsPending)
	fmt.Printf("frames: %d total, %d exported (%d complete, %d partial), %d failed\n", manifest.FramesTotal, manifest.FramesExported, manifest.FramesComplete, manifest.FramesPartial, manifest.FramesFailed)
	return nil
}
