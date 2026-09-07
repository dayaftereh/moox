package core

import "fmt"

// ShipWeaponMount is the persisted authoritative weapon identity carried by a
// design and by every built Ship snapshot. Slice 07 deliberately supports only
// the structural slot/id/count subset; weapon modifiers remain future work.
type ShipWeaponMount struct {
	Slot     int    `json:"slot"`
	WeaponID string `json:"weapon_id"`
	Count    int    `json:"count"`
}

// ShipDesignSpec is the authoritative supported design snapshot.
type ShipDesignSpec struct {
	HullID             string            `json:"hull_id"`
	StrategicPictureID int               `json:"strategic_picture_id"`
	WarpDriveID        string            `json:"warp_drive_id"`
	FTLSpeed           int               `json:"ftl_speed"`
	ComputerID         string            `json:"computer_id"`
	ArmorID            string            `json:"armor_id"`
	ShieldID           string            `json:"shield_id,omitempty"`
	FuelCellID         string            `json:"fuel_cell_id"`
	FuelRangeParsecs   int               `json:"fuel_range_parsecs"`
	HullBaseCostPP     int               `json:"hull_base_cost_pp"`
	HullSpace          int               `json:"hull_space"`
	SpaceUsed          int               `json:"space_used"`
	BaseDesignCostPP   int               `json:"base_design_cost_pp"`
	ProductionCostPP   int               `json:"production_cost_pp"`
	Weapons            []ShipWeaponMount `json:"weapons,omitempty"`
}

type ShipDesign struct {
	ID             ID                `json:"id"`
	EmpireID       ID                `json:"empire_id"`
	Revision       uint32            `json:"revision"`
	VisualRevision uint32            `json:"visual_revision,omitempty"`
	Name           string            `json:"name"`
	Spec           ShipDesignSpec    `json:"spec"`
	VisualGenome   *ShipVisualGenome `json:"visual_genome,omitempty"`
}

type Ship struct {
	ID                   ID                `json:"id"`
	EmpireID             ID                `json:"empire_id"`
	SourceDesignID       ID                `json:"source_design_id"`
	SourceDesignRevision uint32            `json:"source_design_revision"`
	SourceVisualRevision uint32            `json:"source_visual_revision,omitempty"`
	Name                 string            `json:"name"`
	Spec                 ShipDesignSpec    `json:"spec"`
	VisualGenome         *ShipVisualGenome `json:"visual_genome,omitempty"`
}

func validateShipDesignSpec(spec ShipDesignSpec, label string) error {
	if spec.HullID == "" || spec.WarpDriveID == "" || spec.ComputerID == "" || spec.ArmorID == "" || spec.FuelCellID == "" {
		return fmt.Errorf("%s has incomplete mandatory equipment", label)
	}
	if spec.StrategicPictureID < 0 || spec.FTLSpeed < 2 || spec.FuelRangeParsecs <= 0 {
		return fmt.Errorf("%s has invalid strategic picture/drive/fuel values", label)
	}
	if spec.HullBaseCostPP <= 0 || spec.HullSpace <= 0 || spec.SpaceUsed < 0 || spec.SpaceUsed > spec.HullSpace {
		return fmt.Errorf("%s has invalid hull cost/space values", label)
	}
	if spec.BaseDesignCostPP < spec.HullBaseCostPP || spec.ProductionCostPP <= 0 {
		return fmt.Errorf("%s has invalid design production cost", label)
	}
	if len(spec.Weapons) > 8 {
		return fmt.Errorf("%s has %d weapon mounts, maximum is 8", label, len(spec.Weapons))
	}
	previousSlot := -1
	for i, mount := range spec.Weapons {
		if mount.Slot < 0 || mount.Slot > 7 {
			return fmt.Errorf("%s weapon[%d] slot %d is outside 0..7", label, i, mount.Slot)
		}
		if i > 0 && mount.Slot <= previousSlot {
			return fmt.Errorf("%s weapon mounts must be strictly ascending by slot", label)
		}
		if mount.WeaponID == "" || mount.Count <= 0 {
			return fmt.Errorf("%s weapon[%d] requires weapon_id and positive count", label, i)
		}
		previousSlot = mount.Slot
	}
	return nil
}

func validateMilitaryState(state *GameState, empireIDs map[ID]struct{}, checkID func(ID, string) error) (map[ID]Ship, error) {
	designs := make(map[ID]ShipDesign, len(state.ShipDesigns))
	lastDesignID := ID(0)
	for i, design := range state.ShipDesigns {
		label := fmt.Sprintf("ship_design[%d]", i)
		if i > 0 && design.ID <= lastDesignID {
			return nil, fmt.Errorf("ship designs must be strictly ascending by id")
		}
		if err := checkID(design.ID, label); err != nil {
			return nil, err
		}
		if _, ok := empireIDs[design.EmpireID]; !ok {
			return nil, fmt.Errorf("%s references unknown empire %d", label, design.EmpireID)
		}
		if design.Revision == 0 || design.Name == "" {
			return nil, fmt.Errorf("%s requires positive revision and name", label)
		}
		if err := validateShipDesignSpec(design.Spec, label); err != nil {
			return nil, err
		}
		if design.VisualGenome == nil {
			if design.VisualRevision != 0 {
				return nil, fmt.Errorf("%s has visual_revision %d without visual_genome", label, design.VisualRevision)
			}
		} else {
			if design.VisualRevision == 0 {
				return nil, fmt.Errorf("%s visual_genome requires positive visual_revision", label)
			}
			if err := ValidateShipVisualGenome(*design.VisualGenome); err != nil {
				return nil, fmt.Errorf("%s visual_genome: %w", label, err)
			}
		}
		designs[design.ID] = design
		lastDesignID = design.ID
	}

	ships := make(map[ID]Ship, len(state.Ships))
	lastShipID := ID(0)
	for i, ship := range state.Ships {
		label := fmt.Sprintf("ship[%d]", i)
		if i > 0 && ship.ID <= lastShipID {
			return nil, fmt.Errorf("ships must be strictly ascending by id")
		}
		if err := checkID(ship.ID, label); err != nil {
			return nil, err
		}
		if _, ok := empireIDs[ship.EmpireID]; !ok {
			return nil, fmt.Errorf("%s references unknown empire %d", label, ship.EmpireID)
		}
		design, ok := designs[ship.SourceDesignID]
		if !ok {
			return nil, fmt.Errorf("%s references unknown source design %d", label, ship.SourceDesignID)
		}
		if design.EmpireID != ship.EmpireID {
			return nil, fmt.Errorf("%s owner %d differs from source design owner %d", label, ship.EmpireID, design.EmpireID)
		}
		if ship.SourceDesignRevision == 0 || ship.SourceDesignRevision > design.Revision {
			return nil, fmt.Errorf("%s source revision %d is invalid for current design revision %d", label, ship.SourceDesignRevision, design.Revision)
		}
		if ship.Name == "" {
			return nil, fmt.Errorf("%s requires a name", label)
		}
		if err := validateShipDesignSpec(ship.Spec, label); err != nil {
			return nil, err
		}
		if ship.VisualGenome == nil {
			if ship.SourceVisualRevision != 0 {
				return nil, fmt.Errorf("%s has source_visual_revision %d without visual_genome", label, ship.SourceVisualRevision)
			}
		} else {
			if ship.SourceVisualRevision == 0 {
				return nil, fmt.Errorf("%s visual_genome requires positive source_visual_revision", label)
			}
			if design.VisualRevision != 0 && ship.SourceVisualRevision > design.VisualRevision {
				return nil, fmt.Errorf("%s source visual revision %d exceeds current design visual revision %d", label, ship.SourceVisualRevision, design.VisualRevision)
			}
			if err := ValidateShipVisualGenome(*ship.VisualGenome); err != nil {
				return nil, fmt.Errorf("%s visual_genome: %w", label, err)
			}
		}
		ships[ship.ID] = ship
		lastShipID = ship.ID
	}
	return ships, nil
}
