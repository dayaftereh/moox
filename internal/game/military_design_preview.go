package game

import (
	"strings"

	"moox/internal/beam"
	"moox/internal/core"
)

type MilitaryDesignDraft struct {
	HullID             string                  `json:"hull_id"`
	StrategicPictureID int                     `json:"strategic_picture_id"`
	ComputerID         string                  `json:"computer_id,omitempty"`
	ShieldID           *string                 `json:"shield_id,omitempty"`
	Specials           []core.ShipSpecialMount `json:"specials,omitempty"`
	Weapons            []core.ShipWeaponMount  `json:"weapons,omitempty"`
}

type MilitaryDesignWeaponRowPreview struct {
	Slot                 int                         `json:"slot"`
	WeaponID             string                      `json:"weapon_id"`
	Count                int                         `json:"count"`
	MountType            core.ShipWeaponMountType    `json:"mount_type"`
	ModifierIDs          []core.ShipWeaponModifierID `json:"modifier_ids,omitempty"`
	Ammo                 int                         `json:"ammo,omitempty"`
	FiringArcID          core.ShipWeaponFiringArcID  `json:"firing_arc_id,omitempty"`
	MiniaturizationLevel int                         `json:"miniaturization_level"`
	UnitSpace            int                         `json:"unit_space"`
	UnitCostPP           int                         `json:"unit_cost_pp"`
	RowSpace             int                         `json:"row_space"`
	RowCostPP            int                         `json:"row_cost_pp"`
	BeamTactical         *beam.ConfiguredPreview     `json:"beam_tactical,omitempty"`
}

type MilitaryDesignSpecialRowPreview struct {
	Slot                 int    `json:"slot"`
	SpecialID            string `json:"special_id"`
	MiniaturizationLevel int    `json:"miniaturization_level"`
	UnitSpace            int    `json:"unit_space"`
	UnitCostPP           int    `json:"unit_cost_pp"`
	SupportClass         int    `json:"support_class"`
	TacticalSupport      string `json:"tactical_support"`
}

type MilitaryDesignDraftPreview struct {
	SaveAvailable    bool                              `json:"save_available"`
	RejectionReason  string                            `json:"rejection_reason,omitempty"`
	RejectionMessage string                            `json:"rejection_message,omitempty"`
	HullSpace        int                               `json:"hull_space"`
	SpaceUsed        int                               `json:"space_used"`
	RemainingSpace   int                               `json:"remaining_space"`
	BaseDesignCostPP int                               `json:"base_design_cost_pp"`
	ProductionCostPP int                               `json:"production_cost_pp"`
	Weapons          []MilitaryDesignWeaponRowPreview  `json:"weapons,omitempty"`
	Specials         []MilitaryDesignSpecialRowPreview `json:"specials,omitempty"`
	Spec             *core.ShipDesignSpec              `json:"spec,omitempty"`
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
		MilitaryDesignReasonWeaponModifierDuplicate,
		MilitaryDesignReasonWeaponAmmoRequired,
		MilitaryDesignReasonWeaponAmmoInvalid,
		MilitaryDesignReasonWeaponAmmoNotAllowed,
		MilitaryDesignReasonWeaponFiringArcUnknown,
		MilitaryDesignReasonWeaponFiringArcNotAllowed,
		MilitaryDesignReasonSpecialUnknown,
		MilitaryDesignReasonSpecialTechnologyRequired,
		MilitaryDesignReasonSpecialSlotOutOfRange,
		MilitaryDesignReasonSpecialSlotDuplicate,
		MilitaryDesignReasonSpecialDuplicate,
		MilitaryDesignReasonSpecialExclusionConflict,
		MilitaryDesignReasonSpecialRequiresShield,
		MilitaryDesignReasonSpecialNotSupported:
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
	baseSpec, err := r.compileMilitaryDesignSpec(empire, draft.HullID, draft.StrategicPictureID, selections, nil, nil)
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
	for i, mount := range mounts {
		quote, quoteErr := r.quoteWeapon(empire, mount)
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
		mounts[i] = quote.Mount
		weaponSpace += quote.RowSpace
		weaponCost += quote.RowCostPP
		rows = append(rows, MilitaryDesignWeaponRowPreview{
			Slot: quote.Mount.Slot, WeaponID: quote.Mount.WeaponID, Count: quote.Mount.Count,
			MountType: quote.Mount.MountType, ModifierIDs: append([]core.ShipWeaponModifierID(nil), quote.Mount.ModifierIDs...),
			Ammo: quote.Mount.Ammo, FiringArcID: quote.Mount.FiringArcID,
			MiniaturizationLevel: quote.MiniaturizationLevel,
			UnitSpace:            quote.UnitSpace, UnitCostPP: quote.UnitCostPP,
			RowSpace: quote.RowSpace, RowCostPP: quote.RowCostPP,
			BeamTactical: quote.BeamTactical,
		})
	}

	hull := r.ShipHulls[draft.HullID]
	specialMounts, specialQuotes, specialSpace, specialCost, specialErr := r.quoteShipSpecials(empire, hull.SizeIndex, baseSpec.ShieldID, draft.Specials)
	if specialErr != nil {
		preview := rejectedMilitaryDesignPreview(specialErr)
		preview.HullSpace = baseSpec.HullSpace
		preview.SpaceUsed = baseSpec.SpaceUsed + weaponSpace
		preview.RemainingSpace = baseSpec.HullSpace - preview.SpaceUsed
		preview.BaseDesignCostPP = baseSpec.BaseDesignCostPP + weaponCost
		preview.ProductionCostPP = militaryShipProductionCostPP(empire, preview.BaseDesignCostPP, r.RaceModifiers)
		preview.Weapons = rows
		return preview
	}
	specialRows := make([]MilitaryDesignSpecialRowPreview, 0, len(specialQuotes))
	for _, quote := range specialQuotes {
		specialRows = append(specialRows, MilitaryDesignSpecialRowPreview{
			Slot: quote.Mount.Slot, SpecialID: quote.Mount.SpecialID,
			MiniaturizationLevel: quote.MiniaturizationLevel,
			UnitSpace:            quote.UnitSpace, UnitCostPP: quote.UnitCostPP,
			SupportClass: quote.Special.SupportClass, TacticalSupport: quote.Special.TacticalSupport,
		})
	}

	spaceUsed := baseSpec.SpaceUsed + weaponSpace + specialSpace
	baseCost := baseSpec.BaseDesignCostPP + weaponCost + specialCost
	productionCost := militaryShipProductionCostPP(empire, baseCost, r.RaceModifiers)
	provisionalSpec := baseSpec
	provisionalSpec.Weapons = core.CloneShipWeaponMounts(mounts)
	provisionalSpec.Specials = core.CloneShipSpecialMounts(specialMounts)
	provisionalSpec.SpaceUsed = spaceUsed
	provisionalSpec.BaseDesignCostPP = baseCost
	provisionalSpec.ProductionCostPP = productionCost
	preview := MilitaryDesignDraftPreview{
		HullSpace: baseSpec.HullSpace, SpaceUsed: spaceUsed, RemainingSpace: baseSpec.HullSpace - spaceUsed,
		BaseDesignCostPP: baseCost,
		ProductionCostPP: productionCost,
		Weapons:          rows, Spec: &provisionalSpec,
		Specials: specialRows,
	}
	if spaceUsed > baseSpec.HullSpace {
		err := militaryDesignError(MilitaryDesignReasonSpaceExceeded, "military design uses %d space but hull %q has only %d", spaceUsed, draft.HullID, baseSpec.HullSpace)
		preview.RejectionReason = MilitaryDesignReasonSpaceExceeded
		preview.RejectionMessage = err.Error()
		return preview
	}

	spec, err := r.compileMilitaryDesignSpec(empire, draft.HullID, draft.StrategicPictureID, selections, mounts, specialMounts)
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
