package main

import (
	"errors"
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strconv"

	"moox/internal/lbx"
	"moox/internal/moo2gfx"
)

func imageCmd(args []string) error {
	fs := flag.NewFlagSet("image", flag.ContinueOnError)
	out := fs.String("out", "", "output PNG path (required)")
	frame := fs.Int("frame", 0, "animation frame index")
	force := fs.Bool("force", false, "overwrite an existing output file")
	paletteFile := fs.String("palette-file", "", "optional LBX containing an external palette")
	paletteBlock := fs.Int("palette-block", -1, "block index in -palette-file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 2 {
		return errors.New("image requires <file.lbx> <block-index>")
	}
	if *out == "" {
		return errors.New("image requires -out <file.png>")
	}
	blockIndex, err := strconv.Atoi(fs.Arg(1))
	if err != nil {
		return fmt.Errorf("invalid block index: %w", err)
	}
	archive, err := lbx.Open(fs.Arg(0))
	if err != nil {
		return err
	}
	data, err := archive.ReadEntry(blockIndex)
	if err != nil {
		return err
	}
	graphic, err := moo2gfx.Parse(data)
	if err != nil {
		return err
	}

	if *paletteFile != "" {
		if *paletteBlock < 0 {
			return errors.New("-palette-file requires -palette-block")
		}
		paletteArchive, err := lbx.Open(*paletteFile)
		if err != nil {
			return fmt.Errorf("open palette archive: %w", err)
		}
		paletteData, err := paletteArchive.ReadEntry(*paletteBlock)
		if err != nil {
			return fmt.Errorf("read palette block: %w", err)
		}
		palette, err := moo2gfx.ParseExternalPalette(paletteData)
		if err != nil {
			return fmt.Errorf("parse external palette: %w", err)
		}
		palette.Apply(graphic)
	}

	if graphic.Flags&moo2gfx.FlagInternalPalette == 0 && *paletteFile == "" {
		return errors.New("graphic uses an external palette; provide -palette-file and -palette-block")
	}
	img, info, err := graphic.DecodeFrame(*frame)
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
	flags := os.O_CREATE | os.O_WRONLY
	if *force {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}
	f, err := os.OpenFile(abs, flags, 0o644)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	fmt.Printf("image block %d frame %d: %dx%d, drawn=%d, missing_palette=%d -> %s\n", blockIndex, *frame, graphic.Width, graphic.Height, info.DrawnPixels, info.MissingPalettePixels, abs)
	return nil
}
