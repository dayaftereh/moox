package game

import (
	"fmt"

	"moox/internal/core"
)

const ReferenceTriangleCombatReviewScenarioID = "triangle-combat-review-all-tech-v1"

func (r *EconomyRules) NewReferenceTriangleCombatReviewGame(seed uint64) (NewGameResult, error) {
	generated, err := r.NewReferenceTriangleGame(seed)
	if err != nil {
		return NewGameResult{}, err
	}
	state := generated.State
	if err := r.ApplyReferenceTriangleTechnologyProfile(state, ReferenceTriangleProfileAllTech); err != nil {
		return NewGameResult{}, fmt.Errorf("apply combat review all-tech profile: %w", err)
	}
	fightPlan := newGameSystemPlan{
		ID: state.NewID(), Name: "Fight System", X: 459, Y: 347, Spectral: 2,
		Planets: []newGamePlanetPlan{
			{Orbit: 0, SizeID: r.NewGameGalaxy.Homeworld.SizeID, MineralID: r.NewGameGalaxy.Homeworld.MineralID, GravityID: r.NewGameGalaxy.Homeworld.GravityID, ClimateID: r.NewGameGalaxy.Homeworld.ClimateID},
			{Orbit: 2, SizeID: "small", MineralID: "poor", GravityID: "normal_g", ClimateID: "desert"},
			{Orbit: 4, SizeID: "large", MineralID: "abundant", GravityID: "normal_g", ClimateID: "barren"},
		},
		Bodies: []newGameBodyPlan{
			{Orbit: 0, Kind: core.OrbitalBodyPlanet},
			{Orbit: 2, Kind: core.OrbitalBodyPlanet},
			{Orbit: 4, Kind: core.OrbitalBodyPlanet},
		},
	}
	fightSystem := core.StarSystem{ID: fightPlan.ID, Name: fightPlan.Name, X: fightPlan.X, Y: fightPlan.Y, SpectralClass: fightPlan.Spectral}
	for pi, planetPlan := range fightPlan.Planets {
		if err := r.validateNewGamePlanetPlan(planetPlan); err != nil {
			return NewGameResult{}, fmt.Errorf("fight system planet %d: %w", pi, err)
		}
		fightSystem.Planets = append(fightSystem.Planets, core.Planet{
			ID: state.NewID(), Name: newGamePlanetName(fightPlan.Name, planetPlan.Orbit), Orbit: planetPlan.Orbit,
			SizeID: planetPlan.SizeID, MineralID: planetPlan.MineralID, GravityID: planetPlan.GravityID, ClimateID: planetPlan.ClimateID,
		})
	}
	for _, bodyPlan := range fightPlan.Bodies {
		var planet *core.Planet
		for pi := range fightSystem.Planets {
			if fightSystem.Planets[pi].Orbit == bodyPlan.Orbit {
				planet = &fightSystem.Planets[pi]
				break
			}
		}
		if planet == nil {
			return NewGameResult{}, fmt.Errorf("fight system orbit %d has no materialized planet", bodyPlan.Orbit)
		}
		fightSystem.Bodies = append(fightSystem.Bodies, core.OrbitalBody{
			ID: planet.ID, Name: planet.Name, Orbit: bodyPlan.Orbit, Kind: core.OrbitalBodyPlanet, PlanetID: planet.ID,
		})
	}
	state.Galaxy.Systems = append(state.Galaxy.Systems, fightSystem)
	humanEmpireID := generated.Players[0].EmpireID
	aiEmpireID := generated.Players[1].EmpireID
	setDiplomaticPair(state, humanEmpireID, aiEmpireID, core.DiplomaticStanceWar)
	for i := range state.Empires {
		switch state.Empires[i].ID {
		case humanEmpireID:
			state.Empires[i].KnownEmpireIDs = append(state.Empires[i].KnownEmpireIDs, aiEmpireID)
			state.Empires[i].VisitedSystemIDs = append(state.Empires[i].VisitedSystemIDs, fightPlan.ID)
		case aiEmpireID:
			state.Empires[i].KnownEmpireIDs = append(state.Empires[i].KnownEmpireIDs, humanEmpireID)
			state.Empires[i].VisitedSystemIDs = append(state.Empires[i].VisitedSystemIDs, fightPlan.ID)
		}
	}
	baseSpec := referenceCombatReviewFalconSpec()
	for _, empireID := range []core.ID{humanEmpireID, aiEmpireID} {
		designID := state.NewID()
		state.ShipDesigns = append(state.ShipDesigns, core.ShipDesign{
			ID: designID, EmpireID: empireID, Revision: 1, Name: "Falcon MK1", Spec: cloneReferenceCombatReviewSpec(baseSpec),
		})
		shipIDs := make([]core.ID, 0, 3)
		for n := 1; n <= 3; n++ {
			ship := core.Ship{
				ID: state.NewID(), EmpireID: empireID, SourceDesignID: designID, SourceDesignRevision: 1,
				Name: fmt.Sprintf("Falcon MK1 %d", n), Spec: cloneReferenceCombatReviewSpec(baseSpec),
			}
			state.Ships = append(state.Ships, ship)
			shipIDs = append(shipIDs, ship.ID)
		}
		state.StrategicFleets = append(state.StrategicFleets, core.StrategicFleet{
			ID: state.NewID(), EmpireID: empireID, Role: core.StrategicFleetRoleCombat, AtSystemID: fightPlan.ID, ShipIDs: shipIDs,
		})
	}
	state.AddEvent("combat_review_ready", "Falcon 3v3 combat review fixture created at Fight System")
	if err := state.Validate(); err != nil {
		return NewGameResult{}, fmt.Errorf("validate combat review game: %w", err)
	}
	return generated, nil
}

func referenceCombatReviewFalconSpec() core.ShipDesignSpec {
	return core.ShipDesignSpec{
		HullID: "battleship", StrategicPictureID: 24, WarpDriveID: "interphased_drive", FTLSpeed: 7,
		ComputerID: "cybertronic_computer", ArmorID: "xentronium_armor", ShieldID: "class_v_shield",
		FuelCellID: "thorium_fuel_cells", FuelRangeParsecs: 255, HullBaseCostPP: 600, HullSpace: 312,
		SpaceUsed: 306, BaseDesignCostPP: 1652, ProductionCostPP: 1652,
		Specials: []core.ShipSpecialMount{
			{Slot: 0, SpecialID: "battle_pods"}, {Slot: 1, SpecialID: "automated_repair_unit"},
			{Slot: 2, SpecialID: "battle_scanner"}, {Slot: 3, SpecialID: "transporters"},
		},
		Weapons: []core.ShipWeaponMount{
			{Slot: 0, WeaponID: "zeon_missile", Count: 2, MountType: core.ShipWeaponMountStandard, ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierArmored, core.ShipWeaponModifierFast}, Ammo: 5, FiringArcID: core.ShipWeaponFiringArc360},
			{Slot: 1, WeaponID: "particle_beam", Count: 3, MountType: core.ShipWeaponMountHeavy, FiringArcID: core.ShipWeaponFiringArc360},
			{Slot: 2, WeaponID: "particle_beam", Count: 1, MountType: core.ShipWeaponMountHeavy, FiringArcID: core.ShipWeaponFiringArc360},
			{Slot: 3, WeaponID: "particle_beam", Count: 4, MountType: core.ShipWeaponMountPointDefense, FiringArcID: core.ShipWeaponFiringArc360},
			{Slot: 4, WeaponID: "merculite_missile", Count: 2, MountType: core.ShipWeaponMountStandard, ModifierIDs: []core.ShipWeaponModifierID{core.ShipWeaponModifierArmored, core.ShipWeaponModifierFast}, Ammo: 5, FiringArcID: core.ShipWeaponFiringArc360},
			{Slot: 5, WeaponID: "fighter_bays", Count: 1, MountType: core.ShipWeaponMountStandard},
			{Slot: 6, WeaponID: "bomber_bays", Count: 1, MountType: core.ShipWeaponMountStandard},
		},
	}
}

func cloneReferenceCombatReviewSpec(spec core.ShipDesignSpec) core.ShipDesignSpec {
	clone := spec
	clone.Specials = core.CloneShipSpecialMounts(spec.Specials)
	clone.Weapons = core.CloneShipWeaponMounts(spec.Weapons)
	return clone
}
