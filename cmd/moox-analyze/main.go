package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"moox/internal/catalog"
	"moox/internal/lbx"
	"moox/internal/textscan"
)

const version = "0.4.0"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		usage(os.Stdout)
		return nil
	}

	switch args[0] {
	case "help", "-h", "--help":
		usage(os.Stdout)
		return nil
	case "version", "--version":
		fmt.Println(version)
		return nil
	case "inventory":
		return inventoryCmd(args[1:])
	case "normalize":
		return normalizeCmd(args[1:])
	case "inspect":
		return inspectCmd(args[1:])
	case "image":
		return imageCmd(args[1:])
	case "extract":
		return extractCmd(args[1:])
	case "strings":
		return stringsCmd(args[1:])
	default:
		return fmt.Errorf("unknown command %q; run moox-analyze help", args[0])
	}
}

func usage(w io.Writer) {
	fmt.Fprintf(w, `moox-analyze %s - read-only Master of Orion II research tool

Usage:
  moox-analyze inventory [options] <installation-directory>
  moox-analyze normalize race-traits -out <file.json> [-language-out <en.json>] <installation-directory>
  moox-analyze inspect [options] <file.lbx>
  moox-analyze image -out <file.png> <file.lbx> <block-index>
  moox-analyze graphics -out <directory> [options] <installation-directory>
  moox-analyze extract [options] <file.lbx> <block-index>
  moox-analyze strings [options] <file.lbx>
  moox-analyze version

Commands:
  inventory  Catalog files, hashes, LBX metadata and Smacker files as JSON.
  normalize  Convert verified MOO2 source data into loadable MOOX ruleset/language JSON.
  inspect    Validate one LBX and show its block table.
  image      Decode a MOO2 graphic block with an embedded palette to PNG.
  graphics   Catalog all MOO2 graphics and batch-export internally paletted frames.
  extract    Copy one raw LBX block to a file.
  strings    Show printable ASCII strings from one or all LBX blocks.

The tool never writes to the source installation.
`, version)
}

func inventoryCmd(args []string) error {
	fs := flag.NewFlagSet("inventory", flag.ContinueOnError)
	out := fs.String("out", "", "write JSON to this path instead of stdout")
	entryHashes := fs.Bool("entry-hashes", false, "also SHA-256 every LBX block")
	compact := fs.Bool("compact", false, "emit compact JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("inventory requires exactly one installation directory")
	}

	inv, err := catalog.Build(fs.Arg(0), catalog.Options{EntryHashes: *entryHashes})
	if err != nil {
		return err
	}

	var w io.Writer = os.Stdout
	var f *os.File
	if *out != "" {
		abs, err := filepath.Abs(*out)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return err
		}
		f, err = os.Create(abs)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	enc := json.NewEncoder(w)
	if !*compact {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(inv); err != nil {
		return err
	}
	if *out != "" {
		fmt.Fprintf(os.Stderr, "cataloged %d files (%d LBX containers, %d Smacker files) -> %s\n", inv.FileCount, inv.LBXContainers, inv.SmackerFiles, *out)
	}
	return nil
}

func inspectCmd(args []string) error {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "emit JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("inspect requires exactly one file")
	}
	path := fs.Arg(0)

	kind, err := lbx.Detect(path)
	if err != nil {
		return err
	}
	if kind != lbx.KindLBX {
		if *asJSON {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{"path": path, "kind": kind})
		}
		fmt.Printf("%s: %s\n", path, kind)
		return nil
	}

	parsed, err := lbx.Open(path)
	if err != nil {
		return err
	}
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(parsed)
	}

	fmt.Printf("LBX: %s\n", path)
	fmt.Printf("size: %d bytes\nentries: %d\nreserved: 0x%08X\ndata offset: 0x%X (%d)\n", parsed.Size, parsed.EntryCount, parsed.Reserved, parsed.DataOffset, parsed.DataOffset)
	fmt.Println("index\toffset\tsize")
	for _, entry := range parsed.Entries {
		fmt.Printf("%d\t%d\t%d\n", entry.Index, entry.Offset, entry.Size)
	}
	return nil
}

func extractCmd(args []string) error {
	fs := flag.NewFlagSet("extract", flag.ContinueOnError)
	out := fs.String("out", "", "output file path (required)")
	force := fs.Bool("force", false, "overwrite an existing output file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 2 {
		return errors.New("extract requires <file.lbx> <block-index>")
	}
	if *out == "" {
		return errors.New("extract requires -out <path>")
	}

	index, err := strconv.Atoi(fs.Arg(1))
	if err != nil {
		return fmt.Errorf("invalid block index: %w", err)
	}
	parsed, err := lbx.Open(fs.Arg(0))
	if err != nil {
		return err
	}

	flags := os.O_CREATE | os.O_WRONLY
	if *force {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}
	abs, err := filepath.Abs(*out)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	dst, err := os.OpenFile(abs, flags, 0o644)
	if err != nil {
		return err
	}
	defer dst.Close()

	if err := parsed.CopyEntry(index, dst); err != nil {
		return err
	}
	fmt.Printf("extracted block %d (%d bytes) -> %s\n", index, parsed.Entries[index].Size, abs)
	return nil
}

type blockStrings struct {
	Block   int               `json:"block"`
	Strings []textscan.String `json:"strings"`
}

func stringsCmd(args []string) error {
	fs := flag.NewFlagSet("strings", flag.ContinueOnError)
	block := fs.Int("block", -1, "inspect only this LBX block; default is all blocks")
	minLen := fs.Int("min", 4, "minimum printable ASCII run length")
	asJSON := fs.Bool("json", false, "emit JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("strings requires exactly one LBX file")
	}

	parsed, err := lbx.Open(fs.Arg(0))
	if err != nil {
		return err
	}
	if *block >= len(parsed.Entries) {
		return fmt.Errorf("block index %d out of range [0,%d)", *block, len(parsed.Entries))
	}

	start, end := 0, len(parsed.Entries)
	if *block >= 0 {
		start, end = *block, *block+1
	}

	results := make([]blockStrings, 0, end-start)
	for i := start; i < end; i++ {
		data, err := parsed.ReadEntry(i)
		if err != nil {
			return err
		}
		strings := textscan.ASCII(data, *minLen)
		if *asJSON {
			results = append(results, blockStrings{Block: i, Strings: strings})
			continue
		}
		if len(strings) == 0 {
			continue
		}
		fmt.Printf("# block %d (%d bytes)\n", i, len(data))
		for _, item := range strings {
			fmt.Printf("0x%X\t%s\n", item.Offset, item.Value)
		}
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}
	return nil
}
