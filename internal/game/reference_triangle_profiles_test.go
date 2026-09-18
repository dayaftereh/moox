package game

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"moox/internal/core"
)

func TestReferenceTriangleTechnologyProfilesFreeze(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	base, err := rules.NewReferenceTriangleGame(ReferenceTriangleSeed)
	if err != nil {
		t.Fatal(err)
	}
	mid, err := rules.NewReferenceTriangleGame(ReferenceTriangleSeed)
	if err != nil {
		t.Fatal(err)
	}
	all, err := rules.NewReferenceTriangleGame(ReferenceTriangleSeed)
	if err != nil {
		t.Fatal(err)
	}

	midNextID, midRNG := mid.State.NextID, mid.State.RNGState
	if err := rules.ApplyReferenceTriangleTechnologyProfile(mid.State, ReferenceTriangleProfileMidTech); err != nil {
		t.Fatal(err)
	}
	if mid.State.NextID != midNextID || mid.State.RNGState != midRNG {
		t.Fatalf("mid profile changed identity/RNG next=%d->%d rng=%d->%d", midNextID, mid.State.NextID, midRNG, mid.State.RNGState)
	}
	allNextID, allRNG := all.State.NextID, all.State.RNGState
	if err := rules.ApplyReferenceTriangleTechnologyProfile(all.State, ReferenceTriangleProfileAllTech); err != nil {
		t.Fatal(err)
	}
	if all.State.NextID != allNextID || all.State.RNGState != allRNG {
		t.Fatalf("all profile changed identity/RNG next=%d->%d rng=%d->%d", allNextID, all.State.NextID, allRNG, all.State.RNGState)
	}

	wantMidFields := ReferenceTriangleMidTechFieldIDs()
	if len(wantMidFields) != 36 {
		t.Fatalf("mid fields=%d want 36", len(wantMidFields))
	}
	if got := intFingerprint(wantMidFields); got != "5933dd42d93850acea1493767d23a988be40234c4f35097ccbd4d0d75dbbfc6b" {
		t.Fatalf("mid field fingerprint=%s", got)
	}
	wantAllFields := ReferenceTriangleAllTechFieldIDs()
	if len(wantAllFields) != 75 || wantAllFields[0] != 0 || wantAllFields[len(wantAllFields)-1] != 74 {
		t.Fatalf("all fields=%v", wantAllFields)
	}

	for i := range mid.State.Empires {
		empire := mid.State.Empires[i]
		if !reflect.DeepEqual(empire.KnownTechnologyFieldIDs, wantMidFields) || len(empire.KnownTechnologyIDs) != 94 {
			t.Fatalf("mid empire[%d] fields=%v techs=%d", i, empire.KnownTechnologyFieldIDs, len(empire.KnownTechnologyIDs))
		}
		if got := intFingerprint(empire.KnownTechnologyIDs); got != "90a821a5c4890c5ab3098990acbe86ad49dee74040f7ed167985540955f82d7c" {
			t.Fatalf("mid empire[%d] tech fingerprint=%s", i, got)
		}
	}
	for i := range all.State.Empires {
		empire := all.State.Empires[i]
		if !reflect.DeepEqual(empire.KnownTechnologyFieldIDs, wantAllFields) || len(empire.KnownTechnologyIDs) != 202 {
			t.Fatalf("all empire[%d] fields=%v techs=%d", i, empire.KnownTechnologyFieldIDs, len(empire.KnownTechnologyIDs))
		}
		if got := intFingerprint(empire.KnownTechnologyIDs); got != "ce320d510044b036a9e2920f817bcacdea5bc5d587e0787237454a7787a5efca" {
			t.Fatalf("all empire[%d] tech fingerprint=%s", i, got)
		}
		if containsInt(empire.KnownTechnologyIDs, 125) {
			t.Fatalf("all empire[%d] unexpectedly knows fieldless phase_shifter technology 125", i)
		}
	}
	if got := intFingerprint(wantAllFields); got != "0354bfea6fb292379fc7f5ffcf8754effad7496722f0c0152a435f694ccabff6" {
		t.Fatalf("all field fingerprint=%s", got)
	}

	assertReferenceStatesDifferOnlyByTechnology(t, base.State, mid.State)
	assertReferenceStatesDifferOnlyByTechnology(t, base.State, all.State)
}

func TestReferenceTriangleMidTechBoundaryIsPredecessorClosed(t *testing.T) {
	rules := loadCommittedEconomyRules(t)
	known := make(map[int]struct{})
	for _, fieldID := range ReferenceTriangleMidTechFieldIDs() {
		known[fieldID] = struct{}{}
	}
	for fieldID := range known {
		if fieldID == 0 {
			continue
		}
		previous := rules.TechnologyFieldPreviousID[fieldID]
		if previous == 0 {
			continue
		}
		if _, ok := known[previous]; !ok {
			t.Fatalf("mid field %d predecessor %d is outside profile", fieldID, previous)
		}
		if cost := rules.TechnologyFieldCostsRP[fieldID]; cost > 1150 {
			t.Fatalf("mid field %d cost=%g exceeds 1150", fieldID, cost)
		}
	}
}

func assertReferenceStatesDifferOnlyByTechnology(t *testing.T, a, b *core.GameState) {
	t.Helper()
	left := cloneReferenceStateForComparison(t, a)
	right := cloneReferenceStateForComparison(t, b)
	if !reflect.DeepEqual(left, right) {
		t.Fatal("reference profile changed non-technology Turn-1 state")
	}
}

func cloneReferenceStateForComparison(t *testing.T, state *core.GameState) core.GameState {
	t.Helper()
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	var clone core.GameState
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatal(err)
	}
	for i := range clone.Empires {
		clone.Empires[i].KnownTechnologyIDs = nil
		clone.Empires[i].KnownTechnologyFieldIDs = nil
		clone.Empires[i].Research = nil
		clone.Empires[i].HyperAdvancedResearch = nil
	}
	return clone
}

func intFingerprint(ids []int) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.Itoa(id)
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, ",")))
	return hex.EncodeToString(sum[:])
}
