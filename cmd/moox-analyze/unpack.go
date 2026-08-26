package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"moox/internal/rawextract"
)

func unpackCmd(args []string) error {
	fs := flag.NewFlagSet("unpack", flag.ContinueOnError)
	out := fs.String("out", "", "output directory for raw LBX blocks and Smacker files (required)")
	clean := fs.Bool("clean", false, "remove the output directory before rebuilding it")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("unpack requires exactly one MOO2 installation directory")
	}
	if *out == "" {
		return errors.New("unpack requires -out <directory>")
	}

	lastPrinted := 0
	manifest, err := rawextract.Build(fs.Arg(0), *out, rawextract.Options{
		Clean: *clean,
		Progress: func(p rawextract.Progress) {
			if p.FileIndex-lastPrinted >= 25 || p.FileIndex == p.FileCount {
				fmt.Fprintf(os.Stderr, "unpack: %d/%d LBX-named files, %d blocks, %.1f MiB written\n", p.FileIndex, p.FileCount, p.Blocks, float64(p.BytesWritten)/(1024*1024))
				lastPrinted = p.FileIndex
			}
		},
	})
	if err != nil {
		return err
	}
	fmt.Printf("unpacked %d blocks from %d LBX archives and copied %d Smacker files\n", manifest.BlockCount, manifest.LBXArchives, manifest.SmackerFiles)
	fmt.Printf("payload %.2f MiB; total copied %.2f MiB -> %s\n", float64(manifest.PayloadBytes)/(1024*1024), float64(manifest.CopiedBytes)/(1024*1024), *out)
	return nil
}
