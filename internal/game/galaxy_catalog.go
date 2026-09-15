package game

import "fmt"

type GalaxyBias string

const (
	GalaxyBiasLower    GalaxyBias = "lower"
	GalaxyBiasBaseline GalaxyBias = "baseline"
	GalaxyBiasHigher   GalaxyBias = "higher"

	DefaultGalaxySizeID GalaxySize = GalaxySizeSmall
	DefaultGalaxyAgeID  GalaxyAge  = GalaxyAgeNormal
)

type GalaxySizeProfile struct {
	ID        GalaxySize `json:"id"`
	StarCount int        `json:"star_count"`
}

type GalaxyAgeProfile struct {
	ID                  GalaxyAge  `json:"id"`
	MineralResourceBias GalaxyBias `json:"mineral_resource_bias"`
	FoodWorldBias       GalaxyBias `json:"food_world_bias"`
}

type GalaxyCatalog struct {
	DefaultSizeID GalaxySize          `json:"default_size_id"`
	DefaultAgeID  GalaxyAge           `json:"default_age_id"`
	Sizes         []GalaxySizeProfile `json:"sizes"`
	Ages          []GalaxyAgeProfile  `json:"ages"`
}

func GalaxyCatalogFromRules(rules *EconomyRules) (GalaxyCatalog, error) {
	if rules == nil || rules.NewGameGalaxy == nil {
		return GalaxyCatalog{}, fmt.Errorf("new game galaxy rules are unavailable")
	}
	catalog := GalaxyCatalog{
		DefaultSizeID: DefaultGalaxySizeID,
		DefaultAgeID:  DefaultGalaxyAgeID,
		Sizes:         make([]GalaxySizeProfile, 0, len(rules.NewGameGalaxy.GalaxySizes)),
		Ages:          make([]GalaxyAgeProfile, 0, len(rules.NewGameGalaxy.GalaxyAges)),
	}
	for _, size := range rules.NewGameGalaxy.GalaxySizes {
		catalog.Sizes = append(catalog.Sizes, GalaxySizeProfile{ID: GalaxySize(size.ID), StarCount: size.Stars})
	}
	for _, age := range rules.NewGameGalaxy.GalaxyAges {
		profile := GalaxyAgeProfile{ID: GalaxyAge(age.ID)}
		switch profile.ID {
		case GalaxyAgeMineralRich:
			profile.MineralResourceBias = GalaxyBiasHigher
			profile.FoodWorldBias = GalaxyBiasLower
		case GalaxyAgeNormal:
			profile.MineralResourceBias = GalaxyBiasBaseline
			profile.FoodWorldBias = GalaxyBiasBaseline
		case GalaxyAgeOrganicRich:
			profile.MineralResourceBias = GalaxyBiasLower
			profile.FoodWorldBias = GalaxyBiasHigher
		default:
			return GalaxyCatalog{}, fmt.Errorf("unsupported galaxy age profile %q", age.ID)
		}
		catalog.Ages = append(catalog.Ages, profile)
	}
	return catalog, nil
}
