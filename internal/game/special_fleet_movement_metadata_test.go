package game

import (
	"testing"

	"moox/internal/core"
)

func TestProjectSpecialFleetMovementMetadataMatchesAuthoritativeRules(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	empire := core.Empire{RaceID: "human", KnownTechnologyIDs: []int{120, standardFuelCellsTechnologyID}}

	for _, kind := range []core.StrategicFleetSpecialKind{
		core.StrategicFleetSpecialColonyShip,
		core.StrategicFleetSpecialOutpostShip,
		core.StrategicFleetSpecialTroopTransport,
	} {
		fleet := core.StrategicFleet{Role: core.StrategicFleetRoleCivilian, SpecialKind: kind, FTLSpeed: 2}
		projected := rules.ProjectSpecialFleetMovementMetadata(fleet, empire)
		if projected.WarpDriveID != "nuclear_drive" || projected.FTLSpeed != 2 || projected.FuelCellID != "standard_fuel_cells" || projected.FuelRangeParsecs != 4 {
			t.Fatalf("kind=%s projected movement metadata=%+v", kind, projected)
		}
	}

	// Fuel range for fixed strategic support ships follows current empire fuel
	// technology. The per-instance FTL speed remains unchanged, so a later fuel
	// upgrade must not silently replace the installed drive/ETA profile.
	empire.KnownTechnologyIDs = append(empire.KnownTechnologyIDs, deuteriumFuelCellsTechnologyID)
	fleet := core.StrategicFleet{Role: core.StrategicFleetRoleCivilian, SpecialKind: core.StrategicFleetSpecialColonyShip, FTLSpeed: 2}
	projected := rules.ProjectSpecialFleetMovementMetadata(fleet, empire)
	if projected.WarpDriveID != "nuclear_drive" || projected.FTLSpeed != 2 || projected.FuelCellID != "deuterium_fuel_cells" || projected.FuelRangeParsecs != 6 {
		t.Fatalf("upgraded projected movement metadata=%+v", projected)
	}
	if got := colonyShipFuelRangeParsecs(empire); got != projected.FuelRangeParsecs {
		t.Fatalf("authoritative special-fleet range=%d projected=%d", got, projected.FuelRangeParsecs)
	}
}
