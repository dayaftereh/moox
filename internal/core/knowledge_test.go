package core

import "testing"

func TestEmpireVisitedSystemsRemainSortedUnique(t *testing.T) {
	empire := Empire{}
	if !empire.MarkSystemVisited(9) || !empire.MarkSystemVisited(4) || !empire.MarkSystemVisited(7) {
		t.Fatal("new visited systems were not recorded")
	}
	if empire.MarkSystemVisited(7) {
		t.Fatal("duplicate visited system reported as new")
	}
	want := []ID{4, 7, 9}
	if len(empire.VisitedSystemIDs) != len(want) {
		t.Fatalf("visited=%v want=%v", empire.VisitedSystemIDs, want)
	}
	for i := range want {
		if empire.VisitedSystemIDs[i] != want[i] {
			t.Fatalf("visited=%v want=%v", empire.VisitedSystemIDs, want)
		}
	}
	if !empire.HasVisitedSystem(4) || !empire.HasVisitedSystem(7) || empire.HasVisitedSystem(8) {
		t.Fatalf("visited lookup inconsistent: %v", empire.VisitedSystemIDs)
	}
}
func TestKnownEmpirePairsRemainSortedAndSymmetric(t *testing.T) {
	state := NewSmallFixture(0xC017AC7)
	first := state.Empires[0].ID
	second := state.NewID()
	third := state.NewID()
	state.Empires = append(state.Empires,
		Empire{ID: second, Name: "Second", RaceID: "darlok"},
		Empire{ID: third, Name: "Third", RaceID: "alkari"},
	)
	if !state.MarkEmpiresKnown(first, third) || !state.MarkEmpiresKnown(first, second) {
		t.Fatal("new first-contact pairs were not recorded")
	}
	if state.MarkEmpiresKnown(first, second) {
		t.Fatal("duplicate first contact reported as new")
	}
	if !state.EmpiresHaveContact(first, second) || !state.EmpiresHaveContact(second, first) || !state.EmpiresHaveContact(first, third) {
		t.Fatalf("known empire lookup inconsistent: %+v", state.Empires)
	}
	if got := state.Empires[0].KnownEmpireIDs; len(got) != 2 || got[0] != second || got[1] != third {
		t.Fatalf("known empire ids=%v want [%d %d]", got, second, third)
	}
	if err := state.Validate(); err != nil {
		t.Fatal(err)
	}
}
