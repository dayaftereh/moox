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
	out := fs.String("out", "", "ruleset JSON output path (required)")
	languageOut := fs.String("language-out", "", "English language JSON output path (optional)")
	languagesDir := fs.String("languages-dir", "", "write all extracted race-trait locales as <locale>.json into this directory (optional)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("normalize race-traits requires exactly one MOO2 installation directory")
	}
	if *out == "" {
		return errors.New("normalize race-traits requires -out <path>")
	}

	bundle, err := moo2data.DecodeRaceTraitsBundle(fs.Arg(0))
	if err != nil {
		return err
	}
	rulesPath, err := writeJSONAtomic(*out, bundle.Rules)
	if err != nil {
		return err
	}
	languagePath := ""
	if *languageOut != "" {
		languagePath, err = writeJSONAtomic(*languageOut, bundle.English)
		if err != nil {
			return err
		}
	}
	languagePaths := make(map[string]string)
	if *languagesDir != "" {
		for _, locale := range []string{"en", "de", "fr", "es", "it"} {
			language := bundle.Languages[locale]
			if language == nil {
				return fmt.Errorf("missing extracted locale %q", locale)
			}
			path, err := writeJSONAtomic(filepath.Join(*languagesDir, locale+".json"), language)
			if err != nil {
				return err
			}
			languagePaths[locale] = path
		}
	}

	options := 0
	for _, group := range bundle.Rules.Groups {
		options += len(group.Options)
	}
	fmt.Printf("normalized %d race-design groups / %d options -> %s\n", len(bundle.Rules.Groups), options, rulesPath)
	if languagePath != "" {
		fmt.Printf("wrote %d English language keys -> %s\n", len(bundle.English.Strings), languagePath)
	}
	if len(languagePaths) > 0 {
		fmt.Printf("wrote %d race-trait locales with %d keys each -> %s\n", len(languagePaths), len(bundle.English.Strings), *languagesDir)
	}
	return nil
}

func writeJSONAtomic(path string, value any) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(filepath.Dir(abs), ".moox-json-*.tmp")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	committed := false
	defer func() {
		_ = tmp.Close()
		if !committed {
			_ = os.Remove(tmpPath)
		}
	}()
	enc := json.NewEncoder(tmp)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		return "", err
	}
	if err := tmp.Sync(); err != nil {
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmpPath, abs); err != nil {
		_ = os.Remove(abs)
		if err := os.Rename(tmpPath, abs); err != nil {
			return "", err
		}
	}
	committed = true
	return abs, nil
}
