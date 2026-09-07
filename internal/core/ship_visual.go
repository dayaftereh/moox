package core

import (
	"fmt"
	"math"
	"strings"
)

const ShipVisualGenomeVersion = 4

// ShipVisualCutout stores one canonical half-cutout. The renderer mirrors it
// across the ship's longitudinal X axis unless Offset is zero (centerline).
type ShipVisualCutout struct {
	T      float64 `json:"t"`
	Offset float64 `json:"offset"`
	RX     float64 `json:"rx"`
	RY     float64 `json:"ry"`
	Angle  float64 `json:"angle"`
}

// ShipVisualPrimitive stores one canonical half-primitive. The renderer emits
// the exact mirrored partner and never persists a second side copy.
type ShipVisualPrimitive struct {
	Kind   string  `json:"kind"`
	T      float64 `json:"t"`
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Sweep  float64 `json:"sweep"`
}

// ShipVisualGenome is presentation-only authoritative data. It intentionally
// stores the fully resolved generated geometry rather than only a seed so a
// chosen ship remains pixel/topology stable across generator upgrades.
type ShipVisualGenome struct {
	Version       int                   `json:"version"`
	HullID        string                `json:"hull_id"`
	StyleID       string                `json:"style_id"`
	MorphologyID  string                `json:"morphology_id"`
	Seed          string                `json:"seed"`
	Length        float64               `json:"length"`
	Beam          float64               `json:"beam"`
	StationCount  int                   `json:"station_count"`
	StationWidths []float64             `json:"station_widths"`
	NotchDepths   []float64             `json:"notch_depths"`
	EngineCount   int                   `json:"engine_count"`
	DetailCount   int                   `json:"detail_count"`
	Cutouts       []ShipVisualCutout    `json:"cutouts,omitempty"`
	Primitives    []ShipVisualPrimitive `json:"primitives,omitempty"`
}

func finiteVisual(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func validVisualHull(id string) bool {
	switch id {
	case "scout", "frigate", "destroyer", "cruiser", "battleship", "titan", "doom_star":
		return true
	default:
		return false
	}
}

func validVisualStyle(id string) bool {
	switch id {
	case "spear", "sleek", "organic":
		return true
	default:
		return false
	}
}

func validVisualMorphology(id string) bool {
	switch id {
	case "needle", "barge", "manta", "fork", "chevron", "hammer", "bulb":
		return true
	default:
		return false
	}
}

// ValidateShipVisualGenome validates transport/save safety and the v4 symmetry
// contract. It does not make gameplay claims about the selected visual hull;
// until Slice 17 that presentation choice is intentionally independent from
// the currently narrow authoritative military-hull rules.
func ValidateShipVisualGenome(genome ShipVisualGenome) error {
	if genome.Version != ShipVisualGenomeVersion {
		return fmt.Errorf("visual genome version %d, expected %d", genome.Version, ShipVisualGenomeVersion)
	}
	if !validVisualHull(genome.HullID) {
		return fmt.Errorf("unsupported visual hull %q", genome.HullID)
	}
	if !validVisualStyle(genome.StyleID) {
		return fmt.Errorf("unsupported visual style %q", genome.StyleID)
	}
	if !validVisualMorphology(genome.MorphologyID) {
		return fmt.Errorf("unsupported visual morphology %q", genome.MorphologyID)
	}
	if strings.TrimSpace(genome.Seed) == "" || len(genome.Seed) > 512 {
		return fmt.Errorf("visual genome seed must be non-empty and at most 512 bytes")
	}
	if !finiteVisual(genome.Length) || !finiteVisual(genome.Beam) || genome.Length <= 0 || genome.Length > 1000 || genome.Beam <= 0 || genome.Beam > 1000 {
		return fmt.Errorf("visual genome length/beam are outside supported bounds")
	}
	if genome.StationCount < 3 || genome.StationCount > 32 || len(genome.StationWidths) != genome.StationCount || len(genome.NotchDepths) != genome.StationCount {
		return fmt.Errorf("visual genome station arrays do not match station_count %d", genome.StationCount)
	}
	for i := range genome.StationWidths {
		width := genome.StationWidths[i]
		notch := genome.NotchDepths[i]
		if !finiteVisual(width) || width < 0 || width > 2000 {
			return fmt.Errorf("visual genome station_widths[%d] is invalid", i)
		}
		if !finiteVisual(notch) || notch < 0 || notch > width+1e-9 {
			return fmt.Errorf("visual genome notch_depths[%d] is invalid", i)
		}
	}
	if math.Abs(genome.StationWidths[len(genome.StationWidths)-1]) > 1e-9 || math.Abs(genome.NotchDepths[len(genome.NotchDepths)-1]) > 1e-9 {
		return fmt.Errorf("visual genome nose station must terminate on the longitudinal axis")
	}
	if genome.EngineCount < 1 || genome.EngineCount > 16 || genome.DetailCount < 0 || genome.DetailCount > 32 {
		return fmt.Errorf("visual genome engine/detail count is outside supported bounds")
	}
	if len(genome.Cutouts) > 8 || len(genome.Primitives) > 64 {
		return fmt.Errorf("visual genome contains too many cutouts/primitives")
	}
	for i, cutout := range genome.Cutouts {
		if !finiteVisual(cutout.T) || cutout.T < 0 || cutout.T > 1 || !finiteVisual(cutout.Offset) || cutout.Offset < 0 || cutout.Offset > 2000 || !finiteVisual(cutout.RX) || cutout.RX <= 0 || cutout.RX > 1000 || !finiteVisual(cutout.RY) || cutout.RY <= 0 || cutout.RY > 1000 || !finiteVisual(cutout.Angle) || cutout.Angle < -180 || cutout.Angle > 180 {
			return fmt.Errorf("visual genome cutouts[%d] is invalid", i)
		}
	}
	for i, primitive := range genome.Primitives {
		switch primitive.Kind {
		case "wedge", "spike", "pod":
		default:
			return fmt.Errorf("visual genome primitives[%d] has unsupported kind %q", i, primitive.Kind)
		}
		if !finiteVisual(primitive.T) || primitive.T < 0 || primitive.T > 1 || !finiteVisual(primitive.Length) || primitive.Length <= 0 || primitive.Length > 2000 || !finiteVisual(primitive.Width) || primitive.Width <= 0 || primitive.Width > 2000 || !finiteVisual(primitive.Sweep) || primitive.Sweep < -4 || primitive.Sweep > 4 {
			return fmt.Errorf("visual genome primitives[%d] is invalid", i)
		}
	}
	return nil
}

func CloneShipVisualGenome(genome ShipVisualGenome) ShipVisualGenome {
	clone := genome
	clone.StationWidths = append([]float64(nil), genome.StationWidths...)
	clone.NotchDepths = append([]float64(nil), genome.NotchDepths...)
	clone.Cutouts = append([]ShipVisualCutout(nil), genome.Cutouts...)
	clone.Primitives = append([]ShipVisualPrimitive(nil), genome.Primitives...)
	return clone
}

func cloneShipVisualGenomePtr(genome *ShipVisualGenome) *ShipVisualGenome {
	if genome == nil {
		return nil
	}
	clone := CloneShipVisualGenome(*genome)
	return &clone
}
