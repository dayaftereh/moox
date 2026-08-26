package ruleset

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestAssetsValidateAgainstRaces(t *testing.T) {
	races := &RacesFile{Ruleset: "moo2-1.31", Races: []Race{{
		ID:               "alkari",
		Order:            0,
		PortraitAssetKey: "race.alkari.portrait",
		IconAssetKey:     "race.alkari.icon",
	}}}
	assets := &AssetsFile{
		SchemaVersion: AssetsSchemaVersion,
		Ruleset:       "moo2-1.31",
		Assets: []Asset{
			confirmedAsset("race.alkari.portrait", "race_portrait"),
			{Key: "race.alkari.icon", Kind: "race_icon", Status: "pending", Verification: "pending", UnresolvedReason: "generic variant unknown"},
			confirmedAsset("race.alkari.icon.farmer", "race_role_icon"),
			confirmedAsset("race.alkari.icon.worker", "race_role_icon"),
			confirmedAsset("race.alkari.icon.scientist", "race_role_icon"),
			confirmedAsset("race.alkari.icon.marine", "race_role_icon"),
		},
	}
	if err := assets.ValidateAgainstRaces(races); err != nil {
		t.Fatal(err)
	}
}

func TestAssetsValidateBuildingVariants(t *testing.T) {
	building := Building{
		ID:                         "alien_management_center",
		Order:                      0,
		ProductionID:               1,
		ProductionIDVerification:   "test",
		ProductionIDSource:         FieldProvenance{SourceID: "test"},
		TechnologyID:               5,
		TechnologyKey:              "alien_management_center",
		TechnologyLinkVerification: "test",
		TechnologySource:           FieldProvenance{SourceID: "test"},
		NameKey:                    "building.alien_management_center.name",
		NameVerification:           "test",
		NameSource:                 FieldProvenance{SourceID: "test"},
		ColonyReferenceAssetKey:    "building.alien_management_center.colony",
	}
	buildings := &BuildingsFile{SchemaVersion: BuildingsSchemaVersion, Ruleset: "moo2-1.31", Buildings: make([]Building, 48)}
	for i := 0; i < 48; i++ {
		b := building
		b.ID = "building_" + string(rune('a'+i%26)) + string(rune('A'+i/26))
		b.Order = i
		b.ProductionID = i + 1
		b.TechnologyID = i + 1
		b.TechnologyKey = "tech_" + b.ID
		b.NameKey = "name." + b.ID
		b.ColonyReferenceAssetKey = "asset." + b.ID
		buildings.Buildings[i] = b
	}
	assets := &AssetsFile{SchemaVersion: AssetsSchemaVersion, Ruleset: "moo2-1.31"}
	for _, b := range buildings.Buildings {
		asset := Asset{Key: b.ColonyReferenceAssetKey, Kind: "building_colony_set", Status: "confirmed", Verification: "test"}
		for frame := 0; frame < 36; frame++ {
			asset.Variants = append(asset.Variants, AssetVariant{
				ID:        "v" + string(rune('a'+frame)),
				Reference: AssetReference{Archive: "BLDG0.LBX", Block: frame, Frame: 0, BlockSHA256: "deadbeef", Width: 640, Height: 480},
				Metadata:  map[string]int{"effective_frame": frame},
			})
		}
		assets.Assets = append(assets.Assets, asset)
	}
	if err := assets.ValidateAgainstBuildings(buildings); err != nil {
		t.Fatal(err)
	}
}

func TestAssetsRejectConfirmedWithoutReference(t *testing.T) {
	assets := &AssetsFile{SchemaVersion: AssetsSchemaVersion, Ruleset: "moo2-1.31", Assets: []Asset{{
		Key: "race.alkari.portrait", Kind: "race_portrait", Status: "confirmed", Verification: "test",
	}}}
	if err := assets.Validate(); err == nil {
		t.Fatal("expected confirmed asset without reference to fail")
	}
}

func TestCommittedAssetsLoadAndValidate(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test source path")
	}
	root := filepath.Join(filepath.Dir(currentFile), "..", "..")
	races, err := LoadRaces(filepath.Join(root, "data", "rulesets", "moo2-1.31", "races.json"))
	if err != nil {
		t.Fatal(err)
	}
	buildings, err := LoadBuildings(filepath.Join(root, "data", "rulesets", "moo2-1.31", "buildings.json"))
	if err != nil {
		t.Fatal(err)
	}
	shipHulls, err := LoadShipHulls(filepath.Join(root, "data", "rulesets", "moo2-1.31", "ship_hulls.json"))
	if err != nil {
		t.Fatal(err)
	}
	assets, err := LoadAssets(filepath.Join(root, "data", "rulesets", "moo2-1.31", "assets.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := assets.ValidateAgainstRaces(races); err != nil {
		t.Fatal(err)
	}
	if err := assets.ValidateAgainstBuildings(buildings); err != nil {
		t.Fatal(err)
	}
	if err := assets.ValidateAgainstShipHulls(shipHulls); err != nil {
		t.Fatal(err)
	}
	if len(assets.Assets) != 156 {
		t.Fatalf("assets=%d want=156", len(assets.Assets))
	}
	confirmed := 0
	pending := 0
	buildingSets := 0
	buildingVariants := 0
	shipStrategicSets := 0
	shipStrategicVariants := 0
	shipTacticalSets := 0
	shipTacticalVariants := 0
	for _, asset := range assets.Assets {
		switch asset.Status {
		case "confirmed":
			confirmed++
		case "pending":
			pending++
		}
		if asset.Kind == "building_colony_set" {
			buildingSets++
			buildingVariants += len(asset.Variants)
		}
		if asset.Kind == "ship_hull_strategic_set" || asset.Kind == "ship_civilian_strategic_set" {
			shipStrategicSets++
			shipStrategicVariants += len(asset.Variants)
		}
		if asset.Kind == "ship_hull_tactical_set" {
			shipTacticalSets++
			shipTacticalVariants += len(asset.Variants)
		}
	}
	if confirmed != 143 || pending != 13 {
		t.Fatalf("confirmed=%d pending=%d want=143/13", confirmed, pending)
	}
	if buildingSets != 48 || buildingVariants != 1728 {
		t.Fatalf("building sets=%d variants=%d want=48/1728", buildingSets, buildingVariants)
	}
	if shipStrategicSets != 9 || shipStrategicVariants != 352 {
		t.Fatalf("ship strategic sets=%d variants=%d want=9/352", shipStrategicSets, shipStrategicVariants)
	}
	if shipTacticalSets != 6 || shipTacticalVariants != 6560 {
		t.Fatalf("ship tactical sets=%d variants=%d want=6/6560", shipTacticalSets, shipTacticalVariants)
	}
}

func confirmedAsset(key, kind string) Asset {
	return Asset{
		Key:          key,
		Kind:         kind,
		Status:       "confirmed",
		Verification: "test",
		Reference: &AssetReference{
			Archive:     "TEST.LBX",
			Block:       1,
			Frame:       0,
			BlockSHA256: "deadbeef",
			Width:       28,
			Height:      28,
		},
	}
}
