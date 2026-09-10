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
