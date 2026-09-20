package game

import (
	"strings"

	"moox/internal/core"
)

type MilitaryDesignDraft struct {
	HullID             string                 `json:"hull_id"`
	StrategicPictureID int                    `json:"strategic_picture_id"`
	ComputerID         string                 `json:"computer_id,omitempty"`
	ShieldID           *string                `json:"shield_id,omitempty"`
	Weapons            []core.ShipWeaponMount `json:"weapons,omitempty"`
}

type MilitaryDesignWeaponRowPreview struct {
	Slot                 int                         `json:"slot"`
	WeaponID             string                      `json:"weapon_id"`
	Count                int                         `json:"count"`
	MountType            core.ShipWeaponMountType    `json:"mount_type"`
	ModifierIDs          []core.ShipWeaponModifierID `json:"modifier_ids,omitempty"`
	MiniaturizationLevel int                         `json:"miniaturization_level"`
	UnitSpace            int                         `json:"unit_space"`
	UnitCostPP           int                         `json:"unit_cost_pp"`
	RowSpace             int                         `json:"row_space"`
	RowCostPP            int                         `json:"row_cost_pp"`
}

type MilitaryDesignDraftPreview struct {
	SaveAvailable    bool                             `json:"save_available"`
	RejectionReason  string                           `json:"rejection_reason,omitempty"`
	RejectionMessage string                           `json:"rejection_message,omitempty"`
	HullSpace        int                              `json:"hull_space"`
	SpaceUsed        int                              `json:"space_used"`
	RemainingSpace   int                              `json:"remaining_space"`
	BaseDesignCostPP int                              `json:"base_design_cost_pp"`
	ProductionCostPP int                              `json:"production_cost_pp"`
	Weapons          []MilitaryDesignWeaponRowPreview `json:"weapons,omitempty"`
	Spec             *core.ShipDesignSpec             `json:"spec,omitempty"`
}

func militaryDesignReasonFromError(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	reason, _, ok := strings.Cut(message, ":")
	if !ok {
		return ""
	}
	switch reason {
	case MilitaryDesignReasonTechnologyRequired,
		MilitaryDesignReasonMandatorySystemsRequired,
		MilitaryDesignReasonInvalidComponent,
		MilitaryDesignReasonInvalidStrategicPicture,
		MilitaryDesignReasonSpaceExceeded,
		MilitaryDesignReasonOrbitalDockRequired,
		MilitaryDesignReasonRevisionInUse,
		MilitaryDesignReasonWeaponUnknown,
		MilitaryDesignReasonWeaponTechnologyRequired,
		MilitaryDesignReasonWeaponSlotOutOfRange,
		MilitaryDesignReasonWeaponSlotDuplicate,
		MilitaryDesignReasonWeaponCountInvalid,
		MilitaryDesignReasonWeaponMountNotAllowed,
		MilitaryDesignReasonWeaponModifierUnknown,
		MilitaryDesignReasonWeaponModifierNotAllowed,
		MilitaryDesignReasonWeaponModifierRefinementRequired,
		MilitaryDesignReasonWeaponModifierDuplicate:
		return reason
	default:
		return ""
	}
}

func rejectedMilitaryDesignPreview(err error) MilitaryDesignDraftPreview {
	return MilitaryDesignDraftPreview{
		SaveAvailable:    false,
		RejectionReason:  militaryDesignReasonFromError(err),
		RejectionMessage: err.Error(),
	}
}

func (r *EconomyRules) PreviewMilitaryDesign(empire *core.Empire, draft MilitaryDesignDraft) MilitaryDesignDraftPreview {
	selections := militaryDesignSelections{
		ComputerID: draft.ComputerID,
		ShieldID:   draft.ShieldID,
	}
	baseSpec, err := r.compileMilitaryDesignSpec(empire, draft.HullID, draft.StrategicPictureID, selections, nil)
	if err != nil {
		return rejectedMilitaryDesignPreview(err)
	}

	mounts, err := canonicalWeaponMounts(draft.Weapons)
	if err != nil {
		preview := rejectedMilitaryDesignPreview(err)
		preview.HullSpace = baseSpec.HullSpace
		preview.SpaceUsed = baseSpec.SpaceUsed
		preview.RemainingSpace = baseSpec.HullSpace - baseSpec.SpaceUsed
		preview.BaseDesignCostPP = baseSpec.BaseDesignCostPP
		preview.ProductionCostPP = baseSpec.ProductionCostPP
		return preview
	}

	rows := make([]MilitaryDesignWeaponRowPreview, 0, len(mounts))
	weaponSpace := 0
	weaponCost := 0
	for _, mount := range mounts {
		quote, quoteErr := r.quoteBeamWeapon(empire, mount)
		if quoteErr != nil {
			preview := rejectedMilitaryDesignPreview(quoteErr)
			preview.HullSpace = baseSpec.HullSpace
			preview.SpaceUsed = baseSpec.SpaceUsed + weaponSpace
			preview.RemainingSpace = baseSpec.HullSpace - preview.SpaceUsed
			preview.BaseDesignCostPP = baseSpec.BaseDesignCostPP + weaponCost
			preview.ProductionCostPP = militaryShipProductionCostPP(empire, preview.BaseDesignCostPP, r.RaceModifiers)
			preview.Weapons = rows
			return preview
		}
		weaponSpace += quote.RowSpace
		weaponCost += quote.RowCostPP
		rows = append(rows, MilitaryDesignWeaponRowPreview{
			Slot: quote.Mount.Slot, WeaponID: quote.Mount.WeaponID, Count: quote.Mount.Count,
			MountType: quote.Mount.MountType, ModifierIDs: append([]core.ShipWeaponModifierID(nil), quote.Mount.ModifierIDs...),
			MiniaturizationLevel: quote.MiniaturizationLevel,
			UnitSpace:            quote.UnitSpace, UnitCostPP: quote.UnitCostPP,
			RowSpace: quote.RowSpace, RowCostPP: quote.RowCostPP,
		})
	}

	spaceUsed := baseSpec.SpaceUsed + weaponSpace
	baseCost := baseSpec.BaseDesignCostPP + weaponCost
	productionCost := militaryShipProductionCostPP(empire, baseCost, r.RaceModifiers)
	provisionalSpec := baseSpec
	provisionalSpec.Weapons = core.CloneShipWeaponMounts(mounts)
	provisionalSpec.SpaceUsed = spaceUsed
	provisionalSpec.BaseDesignCostPP = baseCost
	provisionalSpec.ProductionCostPP = productionCost
	preview := MilitaryDesignDraftPreview{
		HullSpace: baseSpec.HullSpace, SpaceUsed: spaceUsed, RemainingSpace: baseSpec.HullSpace - spaceUsed,
		BaseDesignCostPP: baseCost,
		ProductionCostPP: productionCost,
		Weapons:          rows, Spec: &provisionalSpec,
	}
	if spaceUsed > baseSpec.HullSpace {
		err := militaryDesignError(MilitaryDesignReasonSpaceExceeded, "military design uses %d space but hull %q has only %d", spaceUsed, draft.HullID, baseSpec.HullSpace)
		preview.RejectionReason = MilitaryDesignReasonSpaceExceeded
		preview.RejectionMessage = err.Error()
		return preview
	}

	spec, err := r.compileMilitaryDesignSpec(empire, draft.HullID, draft.StrategicPictureID, selections, mounts)
	if err != nil {
		preview.RejectionReason = militaryDesignReasonFromError(err)
		preview.RejectionMessage = err.Error()
		return preview
	}
	preview.SaveAvailable = true
	preview.Spec = &spec
	preview.HullSpace = spec.HullSpace
	preview.SpaceUsed = spec.SpaceUsed
	preview.RemainingSpace = spec.HullSpace - spec.SpaceUsed
	preview.BaseDesignCostPP = spec.BaseDesignCostPP
	preview.ProductionCostPP = spec.ProductionCostPP
	return preview
}
