package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"moox/internal/i18n"
	"moox/internal/moo2data"
	"moox/internal/ruleset"
)

func normalizeCmd(args []string) error {
	if len(args) == 0 {
		return errors.New("normalize requires a dataset name; currently supported: race-traits, races, assets")
	}
	switch args[0] {
	case "race-traits":
		return normalizeRaceTraitsCmd(args[1:])
	case "races":
		return normalizeRacesCmd(args[1:])
	case "assets":
		return normalizeAssetsCmd(args[1:])
	default:
		return fmt.Errorf("unknown normalize dataset %q; currently supported: race-traits, races, assets", args[0])
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

func normalizeRacesCmd(args []string) error {
	fs := flag.NewFlagSet("normalize races", flag.ContinueOnError)
	out := fs.String("out", "", "races ruleset JSON output path (required)")
	raceTraitsPath := fs.String("race-traits", "", "normalized race_traits.json path (required)")
	languagesDir := fs.String("languages-dir", "", "merge canonical English race names into en.json in this directory (optional)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("normalize races requires exactly one MOO2 installation directory")
	}
	if *out == "" || *raceTraitsPath == "" {
		return errors.New("normalize races requires -out <path> and -race-traits <race_traits.json>")
	}

	traits, err := ruleset.LoadRaceTraits(*raceTraitsPath)
	if err != nil {
		return fmt.Errorf("load race traits: %w", err)
	}
	bundle, err := moo2data.DecodeRaces(fs.Arg(0), traits)
	if err != nil {
		return err
	}
	path, err := writeJSONAtomic(*out, bundle.Rules)
	if err != nil {
		return err
	}

	if *languagesDir != "" {
		enPath := filepath.Join(*languagesDir, "en.json")
		var english *i18n.File
		if _, statErr := os.Stat(enPath); statErr == nil {
			english, err = i18n.Load(enPath)
			if err != nil {
				return fmt.Errorf("load English language file: %w", err)
			}
		} else if os.IsNotExist(statErr) {
			english = &i18n.File{SchemaVersion: i18n.SchemaVersion, Locale: "en", Strings: map[string]string{}}
		} else {
			return statErr
		}
		if err := english.Merge(bundle.English); err != nil {
			return err
		}
		if _, err := writeJSONAtomic(enPath, english); err != nil {
			return err
		}
	}

	fmt.Printf("normalized %d preset races -> %s\n", len(bundle.Rules.Races), path)
	for _, race := range bundle.Rules.Races {
		fmt.Printf("  %-10s traits=%d derived_picks=%d\n", race.ID, len(race.TraitSelections), race.DerivedPickTotal)
	}
	return nil
}

func normalizeAssetsCmd(args []string) error {
	fs := flag.NewFlagSet("normalize assets", flag.ContinueOnError)
	out := fs.String("out", "", "semantic assets JSON output path (required)")
	racesPath := fs.String("races", "", "normalized races.json path (required)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return errors.New("normalize assets requires exactly one MOO2 installation directory")
	}
	if *out == "" || *racesPath == "" {
		return errors.New("normalize assets requires -out <path> and -races <races.json>")
	}

	races, err := ruleset.LoadRaces(*racesPath)
	if err != nil {
		return fmt.Errorf("load races: %w", err)
	}
	assets, err := moo2data.DecodeAssets(fs.Arg(0), races)
	if err != nil {
		return err
	}
	path, err := writeJSONAtomic(*out, assets)
	if err != nil {
		return err
	}

	confirmed := 0
	pending := 0
	for _, asset := range assets.Assets {
		switch asset.Status {
		case "confirmed":
			confirmed++
		case "pending":
			pending++
		}
	}
	fmt.Printf("normalized %d semantic assets (%d confirmed, %d pending) -> %s\n", len(assets.Assets), confirmed, pending, path)
	return nil
}
