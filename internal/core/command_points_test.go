package core

import (
	"math"
	"testing"
)

func TestCommandPointSnapshotValidationSchema21(t *testing.T) {
	state := NewSmallFixture(0xA530)
	state.Empires[0].CommandPoints = EmpireCommandPoints{Capacity: 5, Used: 6}
	state.Empires[0].Treasury.ShipCommandMaintenanceBC = 10
	if err := state.Validate(); err != nil {
		t.Fatalf("valid schema21 Command Point snapshot rejected: %v", err)
	}

	state.Empires[0].CommandPoints.Capacity = -1
	if err := state.Validate(); err == nil {
		t.Fatal("negative Command Point capacity unexpectedly accepted")
	}
	state.Empires[0].CommandPoints = EmpireCommandPoints{Capacity: 5, Used: -1}
	if err := state.Validate(); err == nil {
		t.Fatal("negative Command Point usage unexpectedly accepted")
	}
	state.Empires[0].CommandPoints = EmpireCommandPoints{Capacity: 5, Used: 6}
	state.Empires[0].Treasury.ShipCommandMaintenanceBC = math.NaN()
	if err := state.Validate(); err == nil {
		t.Fatal("non-finite ship Command Maintenance unexpectedly accepted")
	}
}
