package ruleset

import (
	"path/filepath"
	"testing"
)

func TestSlice178Gate3Checkpoint1ShipSpecialSupportContract(t *testing.T) {
	file, err := LoadShipSpecials(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31", "ship_specials.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Validate(); err != nil {
		t.Fatal(err)
	}
	warp, ok := file.ByOriginalID(37)
	if !ok {
		t.Fatal("missing original ship special row 37")
	}
	if warp.ID != "warp_dissipator" || warp.TacticalSupport != "tactical_passive" {
		t.Fatalf("row 37=%+v", warp)
	}
	for _, tc := range []struct {
		id      string
		support string
	}{
		{"phase_shifter", "tactical_active"},
		{"sub_space_teleporter", "tactical_active"},
		{"achilles_targeting_unit", "tactical_passive"},
		{"automated_repair_unit", "tactical_passive"},
		{"transporters", "tactical_passive"},
		{"troop_pods", "tactical_passive"},
		{"wide_area_jammer", "tactical_passive"},
		{"inertial_stabilizer", "turning_mode"},
		{"battle_pods", "strategic_only"},
		{"scout_lab", "metadata_only"},
	} {
		got, ok := file.ByID(tc.id)
		if !ok {
			t.Fatalf("missing special %q", tc.id)
		}
		if got.TacticalSupport != tc.support {
			t.Fatalf("%s tactical_support=%q want=%q", tc.id, got.TacticalSupport, tc.support)
		}
	}
}

func TestSlice178Gate3Checkpoint1RejectsWrongSpecialSupport(t *testing.T) {
	file, err := LoadShipSpecials(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31", "ship_specials.json"))
	if err != nil {
		t.Fatal(err)
	}
	for i := range file.Specials {
		if file.Specials[i].ID == "phase_shifter" {
			file.Specials[i].TacticalSupport = "metadata_only"
			if err := file.Validate(); err == nil {
				t.Fatal("expected Phase Shifter support mismatch to reject")
			}
			return
		}
	}
	t.Fatal("phase_shifter not found")
}

func TestSlice178Gate3Checkpoint1TacticalRuleBlocksAndFighterSupport(t *testing.T) {
	file, err := LoadTacticalCombat(filepath.Join("..", "..", "data", "rulesets", "moo2-1.31", "tactical_combat.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Validate(); err != nil {
		t.Fatal(err)
	}
	if file.Fighter.CraftPerBay != 4 || file.Fighter.FighterBaseSpeed != 10 || file.Fighter.BomberBaseSpeed != 8 || file.Fighter.HeavyFighterBaseSpeed != 8 {
		t.Fatalf("fighter rules=%+v", file.Fighter)
	}
	if file.Boarding.PhysicalRange != 1 || file.Boarding.TransporterRange != 4 || file.Boarding.DamperTransporterDivisor != 3 {
		t.Fatalf("boarding rules=%+v", file.Boarding)
	}
	if file.Specials.SelfTeleportMaxRange != 20 || file.Specials.AutomatedRepairBasePercent != 20 {
		t.Fatalf("special rules=%+v", file.Specials)
	}
	for _, id := range []string{"fighter_bays", "bomber_bays", "heavy_fighter_bays"} {
		w, ok := file.WeaponByID(id)
		if !ok {
			t.Fatalf("missing Fighter row %q", id)
		}
		if w.Kind != "fighter" || w.TacticalSupport != "fighter" || w.CraftPerUnit != 4 || w.MinDamage != 0 || w.MaxDamage != 0 {
			t.Fatalf("%s=%+v", id, w)
		}
	}
}
