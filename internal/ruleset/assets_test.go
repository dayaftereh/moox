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
	assets, err := LoadAssets(filepath.Join(root, "data", "rulesets", "moo2-1.31", "assets.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := assets.ValidateAgainstRaces(races); err != nil {
		t.Fatal(err)
	}
	if len(assets.Assets) != 79 {
		t.Fatalf("assets=%d want=79", len(assets.Assets))
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
	if confirmed != 66 || pending != 13 {
		t.Fatalf("confirmed=%d pending=%d want=66/13", confirmed, pending)
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
