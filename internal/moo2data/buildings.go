package moo2data

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"path/filepath"

	"moox/internal/i18n"
	"moox/internal/lbx"
	"moox/internal/moo2exe"
	"moox/internal/ruleset"
	"moox/internal/textscan"
)

const (
	buildingHelpNamesSourceID = "moo2-1.31-help-building-names"
	buildingTechNamesSourceID = "moo2-1.31-techname-building-names"
	buildingIDsSourceID       = "openmoo2-building-id-order"
	buildingTableSourceID     = "moo2-1.31-orion2-building-table"
)

type BuildingsBundle struct {
	Rules   *ruleset.BuildingsFile
	English *i18n.File
}

type buildingSpec struct {
	ID   string
	Name string
}

func DecodeBuildings(installationRoot string) (*BuildingsBundle, error) {
	helpPath := filepath.Join(installationRoot, "HELP.LBX")
	helpArchive, err := lbx.Open(helpPath)
	if err != nil {
		return nil, fmt.Errorf("open HELP.LBX: %w", err)
	}
	helpBlock, err := helpArchive.ReadEntry(0)
	if err != nil {
		return nil, fmt.Errorf("read HELP.LBX block 0: %w", err)
	}
	helpFileHash, err := sha256File(helpPath)
	if err != nil {
		return nil, err
	}
	helpBlockHash := sha256.Sum256(helpBlock)
	helpBlockIndex := 0
	count, recordSize, err := parseFixedArrayHeader(helpBlock)
	if err != nil {
		return nil, fmt.Errorf("HELP.LBX block 0: %w", err)
	}

	techNamePath := filepath.Join(installationRoot, "TECHNAME.LBX")
	techNameArchive, err := lbx.Open(techNamePath)
	if err != nil {
		return nil, fmt.Errorf("open TECHNAME.LBX: %w", err)
	}
	techNameBlock, err := techNameArchive.ReadEntry(0)
	if err != nil {
		return nil, fmt.Errorf("read TECHNAME.LBX block 0: %w", err)
	}
	techNameFileHash, err := sha256File(techNamePath)
	if err != nil {
		return nil, err
	}
	techNameBlockHash := sha256.Sum256(techNameBlock)
	techNameBlockIndex := 0
	allTechNameRuns := textscan.ASCII(techNameBlock, 3)
	techRuns, err := concreteTechnologyRuns(allTechNameRuns)
	if err != nil {
		return nil, err
	}
	buildingTable, orion2Hash, err := readOriginalBuildingTable(installationRoot)
	if err != nil {
		return nil, err
	}

	specs := buildingSpecs()
	out := &ruleset.BuildingsFile{
		SchemaVersion: ruleset.BuildingsSchemaVersion,
		Ruleset:       "moo2-1.31",
		Sources: []ruleset.Source{
			{
				ID:          buildingHelpNamesSourceID,
				Type:        "original-observed",
				Description: "Canonical colony-building names observed as HELP.LBX block 0 record headings in Master of Orion II 1.31",
				Archive:     "HELP.LBX",
				Block:       &helpBlockIndex,
				SHA256:      helpFileHash,
				BlockSHA256: hex.EncodeToString(helpBlockHash[:]),
			},
			{
				ID:          buildingTechNamesSourceID,
				Type:        "original-observed",
				Description: "Exact original building/technology names in English TECHNAME.LBX block 0, used for names not present as HELP headings",
				Archive:     "TECHNAME.LBX",
				Block:       &techNameBlockIndex,
				SHA256:      techNameFileHash,
				BlockSHA256: hex.EncodeToString(techNameBlockHash[:]),
			},
			{
				ID:           buildingIDsSourceID,
				Type:         "secondary-community",
				Description:  "OpenMOO2 building table originally supplied the 48-name ordering; MOOX now cross-checks that ordering against the original Orion2.exe building table and retains this source only where the display-name identity is not independently encoded by the technology name",
				URL:          "https://github.com/mimi1vx/openmoo2/blob/2cd3c344aed24380390caaaa819bf7a010b8f4a2/oldmess/_buildings.py",
				AccessedDate: "2026-08-26",
			},
			{
				ID:          buildingTableSourceID,
				Type:        "original-observed",
				Description: "Orion2.exe 1.31 _buildings table in LE object 2 at offset 0x6B3D: 49 records x 19 bytes; records 1..48 contain original building ID at +4, technology ID at +6, production cost (PP) at +8 and maintenance (BC/turn) at +12. N_Bldgs_ iterates IDs 1..48.",
				Archive:     "Orion2.exe",
				SHA256:      orion2Hash,
			},
		},
	}
	english := &i18n.File{
		SchemaVersion: i18n.SchemaVersion,
		Locale:        "en",
		Sources: []i18n.Source{
			{
				ID:          buildingHelpNamesSourceID,
				Type:        "original-observed",
				Description: "Canonical colony-building names observed in HELP.LBX block 0",
				Archive:     "HELP.LBX",
				Block:       &helpBlockIndex,
				SHA256:      helpFileHash,
				BlockSHA256: hex.EncodeToString(helpBlockHash[:]),
			},
			{
				ID:          buildingTechNamesSourceID,
				Type:        "original-observed",
				Description: "Canonical original names observed in TECHNAME.LBX block 0 for HELP-heading exceptions",
				Archive:     "TECHNAME.LBX",
				Block:       &techNameBlockIndex,
				SHA256:      techNameFileHash,
				BlockSHA256: hex.EncodeToString(techNameBlockHash[:]),
			},
		},
		Strings: make(map[string]string, len(specs)),
	}

	for order, spec := range specs {
		nameSourceID := buildingHelpNamesSourceID
		nameVerification := "original-help-heading"
		_, _, nameOffset, _, helpErr := findUniqueHelpHeading(helpBlock, count, recordSize, spec.Name)
		if helpErr != nil {
			if spec.ID != "pollution_processor" && spec.ID != "artificial_planet" {
				return nil, fmt.Errorf("building %s: %w", spec.ID, helpErr)
			}
			var occurrences int
			nameOffset, occurrences = findFirstExactStringOffset(allTechNameRuns, spec.Name)
			if nameOffset < 0 {
				return nil, fmt.Errorf("building %s: original name %q not found in HELP heading or TECHNAME block 0", spec.ID, spec.Name)
			}
			nameSourceID = buildingTechNamesSourceID
			if occurrences == 1 {
				nameVerification = "original-techname-string"
			} else {
				nameVerification = fmt.Sprintf("original-techname-string-%d-occurrences", occurrences)
			}
		}

		entry := buildingTable[order+1]
		technologyID, technologyKey, technologyVerification, technologySource, productionVerification, productionSource, err := buildingTechnologyLink(spec, techRuns, entry)
		if err != nil {
			return nil, err
		}

		nameKey := "building." + spec.ID + ".name"
		building := ruleset.Building{
			ID:                         spec.ID,
			Order:                      order,
			ProductionID:               entry.ID,
			ProductionIDVerification:   productionVerification,
			ProductionIDSource:         productionSource,
			ProductionCostPP:           entry.ProductionCostPP,
			ProductionCostVerification: "original-exe-table-production-cost",
			ProductionCostSource:       ruleset.FieldProvenance{SourceID: buildingTableSourceID, Offset: intPtr(0x6B3D + (order+1)*0x13 + 8)},
			MaintenanceBC:              entry.MaintenanceBC,
			MaintenanceVerification:    "original-exe-table-maintenance",
			MaintenanceSource:          ruleset.FieldProvenance{SourceID: buildingTableSourceID, Offset: intPtr(0x6B3D + (order+1)*0x13 + 12)},
			TechnologyID:               technologyID,
			TechnologyKey:              technologyKey,
			TechnologyLinkVerification: technologyVerification,
			TechnologySource:           technologySource,
			NameKey:                    nameKey,
			NameVerification:           nameVerification,
			NameSource:                 ruleset.FieldProvenance{SourceID: nameSourceID, Offset: &nameOffset},
		}
		building.ColonyReferenceAssetKey = "building." + spec.ID + ".colony"
		out.Buildings = append(out.Buildings, building)
		english.Strings[nameKey] = spec.Name
	}

	if err := out.Validate(); err != nil {
		return nil, err
	}
	if err := english.Validate(); err != nil {
		return nil, err
	}
	return &BuildingsBundle{Rules: out, English: english}, nil
}

type originalBuildingEntry struct {
	ID               int
	TechnologyID     int
	ProductionCostPP int
	MaintenanceBC    int
}

func readOriginalBuildingTable(installationRoot string) ([]originalBuildingEntry, string, error) {
	path := filepath.Join(installationRoot, "Orion2.exe")
	exe, err := moo2exe.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("open Orion2.exe LE module: %w", err)
	}
	const (
		objectNumber = 2
		tableOffset  = 0x6B3D
		entrySize    = 0x13
		entryCount   = 49
	)
	data, err := exe.ReadObject(objectNumber, tableOffset, entrySize*entryCount)
	if err != nil {
		return nil, "", fmt.Errorf("read Orion2.exe _buildings table: %w", err)
	}
	entries := make([]originalBuildingEntry, entryCount)
	for index := 0; index < entryCount; index++ {
		record := data[index*entrySize : (index+1)*entrySize]
		entries[index] = originalBuildingEntry{
			ID:               int(binary.LittleEndian.Uint16(record[4:6])),
			TechnologyID:     int(binary.LittleEndian.Uint16(record[6:8])),
			ProductionCostPP: int(binary.LittleEndian.Uint16(record[8:10])),
			MaintenanceBC:    int(binary.LittleEndian.Uint16(record[12:14])),
		}
		if entries[index].ID != index {
			return nil, "", fmt.Errorf("Orion2.exe _buildings record %d carries id %d", index, entries[index].ID)
		}
	}
	if entries[0].TechnologyID != 0 {
		return nil, "", fmt.Errorf("Orion2.exe _buildings dummy record technology_id=%d, expected 0", entries[0].TechnologyID)
	}
	return entries, exe.SHA256(), nil
}

func buildingTechnologyLink(spec buildingSpec, techRuns []textscan.String, entry originalBuildingEntry) (technologyID int, technologyKey, verification string, source ruleset.FieldProvenance, productionVerification string, productionSource ruleset.FieldProvenance, err error) {
	if len(techRuns) != technologyCount {
		return 0, "", "", source, "", productionSource, fmt.Errorf("building %s: expected %d concrete technology runs, got %d", spec.ID, technologyCount, len(techRuns))
	}
	if entry.ID < 1 || entry.ID > 48 {
		return 0, "", "", source, "", productionSource, fmt.Errorf("building %s: original building id %d outside [1,48]", spec.ID, entry.ID)
	}
	if entry.TechnologyID < 1 || entry.TechnologyID > len(techRuns) {
		return 0, "", "", source, "", productionSource, fmt.Errorf("building %s: original technology id %d outside [1,%d]", spec.ID, entry.TechnologyID, len(techRuns))
	}

	run := techRuns[entry.TechnologyID-1]
	expectedTechnologyName := spec.Name
	verification = "original-exe-table-tech-id-and-techname-exact"
	productionVerification = "original-exe-table-id-techname-crosschecked"
	if spec.ID == "hydroponic_farms" {
		expectedTechnologyName = "Hydroponic Farm"
		verification = "original-exe-table-tech-id-and-singular-techname-alias"
		productionVerification = "original-exe-table-id-singular-techname-crosschecked"
	}
	if spec.ID == "artificial_planet" {
		expectedTechnologyName = "Planet Construction"
		verification = "original-exe-table-tech-id-planet-construction"
		productionVerification = "original-exe-table-id-secondary-display-name-link"
	}
	if run.Value != expectedTechnologyName {
		return 0, "", "", source, "", productionSource, fmt.Errorf("building %s: original building id %d references tech %d %q, expected %q", spec.ID, entry.ID, entry.TechnologyID, run.Value, expectedTechnologyName)
	}
	source = ruleset.FieldProvenance{SourceID: buildingTableSourceID}
	productionSource = ruleset.FieldProvenance{SourceID: buildingTableSourceID}
	return entry.TechnologyID, stableTechnologyID(run.Value), verification, source, productionVerification, productionSource, nil
}
func findUniqueHelpHeading(data []byte, count, recordSize int, heading string) (recordIndex, recordOffset, nameOffset int, recordHash string, err error) {
	found := -1
	for i := 0; i < count; i++ {
		offset := 4 + i*recordSize
		record := data[offset : offset+recordSize]
		runs := textscan.ASCII(record, 4)
		if len(runs) == 0 || runs[0].Value != heading {
			continue
		}
		if found >= 0 {
			return 0, 0, 0, "", fmt.Errorf("HELP heading %q occurs in more than one record", heading)
		}
		found = i
		recordOffset = offset
		nameOffset = offset + runs[0].Offset
		sum := sha256.Sum256(record)
		recordHash = hex.EncodeToString(sum[:])
	}
	if found < 0 {
		return 0, 0, 0, "", fmt.Errorf("HELP heading %q not found", heading)
	}
	return found, recordOffset, nameOffset, recordHash, nil
}

func findFirstExactStringOffset(runs []textscan.String, value string) (offset, occurrences int) {
	offset = -1
	for _, run := range runs {
		if run.Value != value {
			continue
		}
		if offset < 0 {
			offset = run.Offset
		}
		occurrences++
	}
	return offset, occurrences
}

func buildingSpecs() []buildingSpec {
	return []buildingSpec{
		{ID: "alien_management_center", Name: "Alien Management Center"},
		{ID: "armor_barracks", Name: "Armor Barracks"},
		{ID: "artemis_system_net", Name: "Artemis System Net"},
		{ID: "astro_university", Name: "Astro University"},
		{ID: "atmospheric_renewer", Name: "Atmospheric Renewer"},
		{ID: "autolab", Name: "Autolab"},
		{ID: "automated_factories", Name: "Automated Factories"},
		{ID: "battlestation", Name: "Battlestation"},
		{ID: "capitol", Name: "Capitol"},
		{ID: "cloning_center", Name: "Cloning Center"},
		{ID: "colony_base", Name: "Colony Base"},
		{ID: "deep_core_mining", Name: "Deep Core Mining"},
		{ID: "core_waste_dumps", Name: "Core Waste Dumps"},
		{ID: "dimensional_portal", Name: "Dimensional Portal"},
		{ID: "biospheres", Name: "Biospheres"},
		{ID: "food_replicators", Name: "Food Replicators"},
		{ID: "gaia_transformation", Name: "Gaia Transformation"},
		{ID: "galactic_currency_exchange", Name: "Galactic Currency Exchange"},
		{ID: "galactic_cybernet", Name: "Galactic Cybernet"},
		{ID: "holo_simulator", Name: "Holo Simulator"},
		{ID: "hydroponic_farms", Name: "Hydroponic Farms"},
		{ID: "marine_barracks", Name: "Marine Barracks"},
		{ID: "planetary_barrier_shield", Name: "Planetary Barrier Shield"},
		{ID: "planetary_flux_shield", Name: "Planetary Flux Shield"},
		{ID: "planetary_gravity_generator", Name: "Planetary Gravity Generator"},
		{ID: "planetary_missile_base", Name: "Planetary Missile Base"},
		{ID: "ground_batteries", Name: "Ground Batteries"},
		{ID: "planetary_radiation_shield", Name: "Planetary Radiation Shield"},
		{ID: "planetary_stock_exchange", Name: "Planetary Stock Exchange"},
		{ID: "planetary_supercomputer", Name: "Planetary Supercomputer"},
		{ID: "pleasure_dome", Name: "Pleasure Dome"},
		{ID: "pollution_processor", Name: "Pollution Processor"},
		{ID: "recyclotron", Name: "Recyclotron"},
		{ID: "robotic_factory", Name: "Robotic Factory"},
		{ID: "research_laboratory", Name: "Research Laboratory"},
		{ID: "robo_miners", Name: "Robo-Miners"},
		{ID: "soil_enrichment", Name: "Soil Enrichment"},
		{ID: "space_academy", Name: "Space Academy"},
		{ID: "spaceport", Name: "Spaceport"},
		{ID: "star_base", Name: "Star Base"},
		{ID: "star_fortress", Name: "Star Fortress"},
		{ID: "stellar_converter", Name: "Stellar Converter"},
		{ID: "subterranean_farms", Name: "Subterranean Farms"},
		{ID: "terraforming", Name: "Terraforming"},
		{ID: "warp_interdictor", Name: "Warp Interdictor"},
		{ID: "weather_control_system", Name: "Weather Control System"},
		{ID: "fighter_garrison", Name: "Fighter Garrison"},
		{ID: "artificial_planet", Name: "Artificial Planet"},
	}
}
