package ruleset

import (
	"encoding/json"
	"fmt"
	"os"
)

const AssetsSchemaVersion = 2

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
	Variants         []AssetVariant  `json:"variants,omitempty"`
	UnresolvedReason string          `json:"unresolved_reason,omitempty"`
}

type AssetVariant struct {
	ID        string         `json:"id"`
	Reference AssetReference `json:"reference"`
	Metadata  map[string]int `json:"metadata,omitempty"`
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

func validateAssetReference(key string, ref AssetReference) error {
	if ref.Archive == "" || ref.Block < 0 || ref.Frame < 0 || ref.BlockSHA256 == "" || ref.Width <= 0 || ref.Height <= 0 {
		return fmt.Errorf("asset %q has incomplete reference", key)
	}
	return nil
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
			if asset.Reference == nil && len(asset.Variants) == 0 {
				return fmt.Errorf("confirmed asset %q requires a reference or variants", asset.Key)
			}
			if asset.Reference != nil && len(asset.Variants) != 0 {
				return fmt.Errorf("confirmed asset %q cannot mix a direct reference and variants", asset.Key)
			}
			if asset.Reference != nil {
				if err := validateAssetReference(asset.Key, *asset.Reference); err != nil {
					return err
				}
			}
			variantIDs := make(map[string]struct{}, len(asset.Variants))
			for _, variant := range asset.Variants {
				if variant.ID == "" {
					return fmt.Errorf("asset %q has a variant without an id", asset.Key)
				}
				if _, exists := variantIDs[variant.ID]; exists {
					return fmt.Errorf("asset %q has duplicate variant id %q", asset.Key, variant.ID)
				}
				variantIDs[variant.ID] = struct{}{}
				if err := validateAssetReference(asset.Key+"/"+variant.ID, variant.Reference); err != nil {
					return err
				}
			}
			if asset.UnresolvedReason != "" {
				return fmt.Errorf("confirmed asset %q cannot have unresolved_reason", asset.Key)
			}
		} else {
			if asset.Reference != nil || len(asset.Variants) != 0 {
				return fmt.Errorf("pending asset %q must not pretend to have resolved references", asset.Key)
			}
			if asset.UnresolvedReason == "" {
				return fmt.Errorf("pending asset %q requires unresolved_reason", asset.Key)
			}
		}
	}
	return nil
}

func (f *AssetsFile) assetByKey() map[string]Asset {
	byKey := make(map[string]Asset, len(f.Assets))
	for _, asset := range f.Assets {
		byKey[asset.Key] = asset
	}
	return byKey
}

func (f *AssetsFile) ValidateAgainstRaces(races *RacesFile) error {
	if err := f.Validate(); err != nil {
		return err
	}
	if races == nil {
		return fmt.Errorf("races are required")
	}
	byKey := f.assetByKey()
	for _, race := range races.Races {
		portrait, ok := byKey[race.PortraitAssetKey]
		if !ok {
			return fmt.Errorf("race %q portrait asset %q is missing", race.ID, race.PortraitAssetKey)
		}
		if portrait.Status != "confirmed" || portrait.Kind != "race_portrait" || portrait.Reference == nil {
			return fmt.Errorf("race %q portrait asset %q is not a confirmed direct race_portrait", race.ID, race.PortraitAssetKey)
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
			if !ok || asset.Status != "confirmed" || asset.Kind != "race_role_icon" || asset.Reference == nil {
				return fmt.Errorf("race %q requires confirmed direct role icon %q", race.ID, key)
			}
		}
	}
	return nil
}

func (f *AssetsFile) ValidateAgainstBuildings(buildings *BuildingsFile) error {
	if err := f.Validate(); err != nil {
		return err
	}
	if buildings == nil {
		return fmt.Errorf("buildings are required")
	}
	if err := buildings.Validate(); err != nil {
		return fmt.Errorf("validate buildings: %w", err)
	}
	byKey := f.assetByKey()
	for _, building := range buildings.Buildings {
		if building.ColonyReferenceAssetKey == "" {
			return fmt.Errorf("building %q has no colony reference asset key", building.ID)
		}
		asset, ok := byKey[building.ColonyReferenceAssetKey]
		if !ok {
			return fmt.Errorf("building %q colony asset %q is missing", building.ID, building.ColonyReferenceAssetKey)
		}
		if asset.Status != "confirmed" || asset.Kind != "building_colony_set" || len(asset.Variants) != 36 {
			return fmt.Errorf("building %q colony asset %q is not a confirmed 36-variant building_colony_set", building.ID, building.ColonyReferenceAssetKey)
		}
		seenFrames := make(map[int]struct{}, 36)
		for _, variant := range asset.Variants {
			effectiveFrame, ok := variant.Metadata["effective_frame"]
			if !ok || effectiveFrame < 0 || effectiveFrame >= 36 {
				return fmt.Errorf("building %q variant %q has invalid effective_frame metadata", building.ID, variant.ID)
			}
			if _, exists := seenFrames[effectiveFrame]; exists {
				return fmt.Errorf("building %q repeats effective frame %d", building.ID, effectiveFrame)
			}
			seenFrames[effectiveFrame] = struct{}{}
		}
	}
	return nil
}

func (f *AssetsFile) ValidateAgainstShipHulls(hulls *ShipHullsFile) error {
	if err := f.Validate(); err != nil {
		return err
	}
	if hulls == nil {
		return fmt.Errorf("ship hulls are required")
	}
	if err := hulls.Validate(); err != nil {
		return fmt.Errorf("validate ship hulls: %w", err)
	}
	byKey := f.assetByKey()
	for _, hull := range hulls.Hulls {
		asset, ok := byKey[hull.StrategicAssetKey]
		if !ok {
			return fmt.Errorf("ship hull %q strategic asset %q is missing", hull.ID, hull.StrategicAssetKey)
		}
		wantVariants := len(hull.StrategicPictureIDs) * 8
		if asset.Status != "confirmed" || asset.Kind != "ship_hull_strategic_set" || len(asset.Variants) != wantVariants {
			return fmt.Errorf("ship hull %q strategic asset %q is not a confirmed %d-variant ship_hull_strategic_set", hull.ID, hull.StrategicAssetKey, wantVariants)
		}
		allowed := make(map[int]int, len(hull.StrategicPictureIDs))
		for styleIndex, pictureID := range hull.StrategicPictureIDs {
			allowed[pictureID] = styleIndex
		}
		seen := make(map[string]struct{}, wantVariants)
		for _, variant := range asset.Variants {
			colorIndex, okColor := variant.Metadata["color_index"]
			pictureID, okPicture := variant.Metadata["picture_id"]
			styleIndex, okStyle := variant.Metadata["style_index"]
			if !okColor || !okPicture || !okStyle || colorIndex < 0 || colorIndex >= 8 {
				return fmt.Errorf("ship hull %q variant %q has invalid strategic metadata", hull.ID, variant.ID)
			}
			wantStyle, ok := allowed[pictureID]
			if !ok || styleIndex != wantStyle {
				return fmt.Errorf("ship hull %q variant %q has picture/style %d/%d outside hull mapping", hull.ID, variant.ID, pictureID, styleIndex)
			}
			key := fmt.Sprintf("%d/%d", colorIndex, pictureID)
			if _, exists := seen[key]; exists {
				return fmt.Errorf("ship hull %q repeats color/picture %s", hull.ID, key)
			}
			seen[key] = struct{}{}
		}
	}
	return nil
}
