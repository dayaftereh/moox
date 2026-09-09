package game

import "testing"

func TestResearchTechnologyEffectsExposeNormalizedRuntimeUnlocks(t *testing.T) {
	rules := loadCommittedEconomyRules(t)

	biospheres := rules.researchTechnologyEffects(61)
	if len(biospheres) == 0 || biospheres[0].Kind != "building_unlock" || biospheres[0].ID != "biospheres" {
		t.Fatalf("biospheres effects=%+v", biospheres)
	}
	if biospheres[0].ProductionCostPP != 60 || biospheres[0].MaintenanceBC != 1 {
		t.Fatalf("biospheres building metadata=%+v", biospheres[0])
	}

	fusionDrive := rules.researchTechnologyEffects(72)
	foundDrive := false
	for _, effect := range fusionDrive {
		if effect.Kind == "ship_drive_unlock" && effect.ID == "fusion_drive" {
			foundDrive = true
			if effect.FTLSpeed != 3 {
				t.Fatalf("fusion drive FTL speed=%d", effect.FTLSpeed)
			}
		}
	}
	if !foundDrive {
		t.Fatalf("fusion drive effects=%+v", fusionDrive)
	}
}
