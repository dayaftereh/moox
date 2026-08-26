package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const AssetsSchemaVersion = 1

type AssetsFile struct {
	SchemaVersion int      `json:"schema_version"`
	Ruleset       string   `json:"ruleset"`
	Sources       []Source `json:"sources"`
	Assets        []Asset  `json:"assets"`
}

type Asset struct {
	Key              string          `json:"key"`
	Kind             string          `json:"kind"`
	Status           string          `json:"status"`
	Verification     string          `json:"verification"`
	Reference        *AssetReference `json:"reference,omitempty"`
	UnresolvedReason string          `json:"unresolved_reason,omitempty"`
}

type AssetReference struct {
	Archive     string `json:"archive"`
	Block       int    `json:"block"`
	Frame       int    `json:"frame"`
	BlockSHA256 string `json:"block_sha256"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

func LoadAssets(path string) (*AssetsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var file AssetsFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	return &file, nil
}

func (f *AssetsFile) Validate() error {
	if f.SchemaVersion != AssetsSchemaVersion {
		return fmt.Errorf("unsupported assets schema version %d", f.SchemaVersion)
	}
	if f.Ruleset == "" {
		return fmt.Errorf("ruleset is required")
	}
	seen := make(map[string]struct{}, len(f.Assets))
	for _, asset := range f.Assets {
		if asset.Key == "" || asset.Kind == "" || asset.Status == "" || asset.Verification == "" {
			return fmt.Errorf("asset key, kind, status and verification are required")
		}
		if _, exists := seen[asset.Key]; exists {
			return fmt.Errorf("duplicate asset key %q", asset.Key)
		}
		seen[asset.Key] = struct{}{}
		switch asset.Status {
		case "confirmed", "pending":
		default:
			return fmt.Errorf("asset %q has unsupported status %q", asset.Key, asset.Status)
		}
		if asset.Status == "confirmed" {
			if asset.Reference == nil {
				return fmt.Errorf("confirmed asset %q requires a reference", asset.Key)
			}
			if asset.Reference.Archive == "" || asset.Reference.Block < 0 || asset.Reference.Frame < 0 || asset.Reference.BlockSHA256 == "" || asset.Reference.Width <= 0 || asset.Reference.Height <= 0 {
				return fmt.Errorf("confirmed asset %q has incomplete reference", asset.Key)
			}
			if asset.UnresolvedReason != "" {
				return fmt.Errorf("confirmed asset %q cannot have unresolved_reason", asset.Key)
			}
		} else {
			if asset.Reference != nil {
				return fmt.Errorf("pending asset %q must not pretend to have a resolved reference", asset.Key)
			}
			if asset.UnresolvedReason == "" {
				return fmt.Errorf("pending asset %q requires unresolved_reason", asset.Key)
			}
		}
	}
	return nil
}

func (f *AssetsFile) ValidateAgainstRaces(races *RacesFile) error {
	if err := f.Validate(); err != nil {
		return err
	}
	if races == nil {
		return fmt.Errorf("races are required")
	}
	byKey := make(map[string]Asset, len(f.Assets))
	for _, asset := range f.Assets {
		byKey[asset.Key] = asset
	}
	for _, race := range races.Races {
		portrait, ok := byKey[race.PortraitAssetKey]
		if !ok {
			return fmt.Errorf("race %q portrait asset %q is missing", race.ID, race.PortraitAssetKey)
		}
		if portrait.Status != "confirmed" || portrait.Kind != "race_portrait" {
			return fmt.Errorf("race %q portrait asset %q is not a confirmed race_portrait", race.ID, race.PortraitAssetKey)
		}
		icon, ok := byKey[race.IconAssetKey]
		if !ok {
			return fmt.Errorf("race %q icon asset %q is missing", race.ID, race.IconAssetKey)
		}
		if icon.Kind != "race_icon" {
			return fmt.Errorf("race %q icon asset %q has kind %q", race.ID, race.IconAssetKey, icon.Kind)
		}
		for _, role := range []string{"farmer", "worker", "scientist", "marine"} {
			key := race.IconAssetKey + "." + role
			asset, ok := byKey[key]
			if !ok || asset.Status != "confirmed" || asset.Kind != "race_role_icon" {
				return fmt.Errorf("race %q requires confirmed role icon %q", race.ID, key)
			}
		}
	}
	return nil
}
