package game

import (
	"math"

	"moox/internal/core"
)

var startingScoutVisualStyles = []string{"spear", "sleek", "organic"}
var startingScoutVisualMorphologies = []string{"needle", "barge", "manta", "fork", "chevron", "hammer", "bulb"}

type startingScoutMorphologyProfile struct {
	lengthScale    float64
	beamScale      float64
	coreDensity    float64
	frontBias      float64
	rearBias       float64
	primitiveScale float64
	cutoutBonus    int
}

type visualPRNG struct{ state uint32 }

func visualHashSeed(value string) uint32 {
	hash := uint32(0x811c9dc5)
	for i := 0; i < len(value); i++ {
		hash ^= uint32(value[i])
		hash *= 0x01000193
	}
	return hash
}

func newVisualPRNG(seed uint32) *visualPRNG {
	if seed == 0 {
		seed = 0x6d2b79f5
	}
	return &visualPRNG{state: seed}
}

func (r *visualPRNG) next() float64 {
	r.state += 0x6d2b79f5
	value := r.state
	value = (value ^ (value >> 15)) * (value | 1)
	value ^= value + (value^(value>>7))*(value|61)
	value ^= value >> 14
	return float64(value) / 4294967296.0
}

func startingScoutStyleScale(styleID string) (float64, float64) {
	switch styleID {
	case "sleek":
		return 1.06, .9
	case "organic":
		return .96, 1.12
	default:
		return 1.02, 1
	}
}

func startingScoutMorphology(id string) startingScoutMorphologyProfile {
	switch id {
	case "needle":
		return startingScoutMorphologyProfile{1.28, .58, .72, .88, .72, .78, 0}
	case "barge":
		return startingScoutMorphologyProfile{.76, 1.62, 1.18, .78, .94, 1.08, 0}
	case "manta":
		return startingScoutMorphologyProfile{.84, 1.82, .86, .58, .52, 1.34, 1}
	case "fork":
		return startingScoutMorphologyProfile{.98, 1.26, .64, .52, .82, 1.24, 1}
	case "chevron":
		return startingScoutMorphologyProfile{.82, 1.7, .6, .44, .7, 1.5, 0}
	case "hammer":
		return startingScoutMorphologyProfile{.78, 1.42, .92, 1.38, .62, 1.2, 0}
	case "bulb":
		return startingScoutMorphologyProfile{.68, 1.9, 1.34, .92, .92, 1.06, 1}
	default:
		return startingScoutMorphologyProfile{1, 1, 1, 1, 1, 1, 0}
	}
}

func startingScoutEnvelope(morphologyID string, t float64) float64 {
	sin := math.Max(0, math.Sin(math.Pi*math.Min(.995, t+.02)))
	switch morphologyID {
	case "needle":
		return math.Pow(sin, .9)
	case "barge":
		return .58 + .42*math.Pow(sin, .32)
	case "manta":
		return math.Pow(sin, .24)
	case "fork":
		return .4 + .6*math.Pow(sin, .62)
	case "chevron":
		return .28 + .72*math.Pow(sin, .72)
	case "hammer":
		return .38 + .62*math.Pow(t, .82)
	case "bulb":
		return .72 + .28*math.Pow(sin, .16)
	default:
		return sin
	}
}

func startingScoutSignaturePrimitives(morphologyID string, beam, length float64) []core.ShipVisualPrimitive {
	switch morphologyID {
	case "manta":
		return []core.ShipVisualPrimitive{
			{Kind: "wedge", T: .46, Length: length * .28, Width: beam * .8, Sweep: -.9},
			{Kind: "wedge", T: .58, Length: length * .22, Width: beam * .66, Sweep: .55},
		}
	case "fork":
		return []core.ShipVisualPrimitive{
			{Kind: "spike", T: .7, Length: length * .34, Width: beam * .42, Sweep: .12},
			{Kind: "pod", T: .36, Length: length * .14, Width: beam * .46, Sweep: -.25},
		}
	case "chevron":
		return []core.ShipVisualPrimitive{
			{Kind: "wedge", T: .5, Length: length * .4, Width: beam * .86, Sweep: -1.2},
			{Kind: "spike", T: .34, Length: length * .26, Width: beam * .46, Sweep: -.5},
		}
	case "hammer":
		return []core.ShipVisualPrimitive{
			{Kind: "pod", T: .72, Length: length * .22, Width: beam * .7, Sweep: .08},
			{Kind: "wedge", T: .68, Length: length * .18, Width: beam * .58, Sweep: .48},
		}
	case "bulb":
		return []core.ShipVisualPrimitive{
			{Kind: "pod", T: .46, Length: length * .28, Width: beam * .72, Sweep: 0},
			{Kind: "pod", T: .62, Length: length * .22, Width: beam * .62, Sweep: .18},
		}
	case "barge":
		return []core.ShipVisualPrimitive{{Kind: "pod", T: .48, Length: length * .2, Width: beam * .42, Sweep: 0}}
	case "needle":
		return []core.ShipVisualPrimitive{{Kind: "spike", T: .74, Length: length * .18, Width: beam * .24, Sweep: .2}}
	default:
		return nil
	}
}

func startingScoutRandomPrimitive(random *visualPRNG, beam, length float64, styleID, morphologyID string) core.ShipVisualPrimitive {
	roll := random.next()
	kind := "wedge"
	switch {
	case morphologyID == "bulb" || styleID == "organic":
		if roll < .18 {
			kind = "wedge"
		} else if roll < .32 {
			kind = "spike"
		} else {
			kind = "pod"
		}
	case morphologyID == "needle":
		if roll < .34 {
			kind = "wedge"
		} else if roll < .88 {
			kind = "spike"
		} else {
			kind = "pod"
		}
	case morphologyID == "manta" || morphologyID == "chevron":
		if roll < .72 {
			kind = "wedge"
		} else if roll < .84 {
			kind = "spike"
		} else {
			kind = "pod"
		}
	case styleID == "sleek":
		if roll < .62 {
			kind = "wedge"
		} else if roll < .72 {
			kind = "spike"
		} else {
			kind = "pod"
		}
	default:
		if roll < .48 {
			kind = "wedge"
		} else if roll < .82 {
			kind = "spike"
		} else {
			kind = "pod"
		}
	}
	morph := startingScoutMorphology(morphologyID)
	styleLength := 1.02
	if styleID == "spear" {
		styleLength = 1.15
	} else if styleID == "organic" {
		styleLength = .88
	}
	styleWidth := 1.0
	if styleID == "organic" {
		styleWidth = 1.18
	} else if styleID == "sleek" {
		styleWidth = .8
	}
	sweepRange := 1.0
	if morphologyID == "chevron" {
		sweepRange = 1.45
	} else if morphologyID == "manta" {
		sweepRange = 1.1
	} else if morphologyID == "needle" {
		sweepRange = .55
	}
	return core.ShipVisualPrimitive{
		Kind:   kind,
		T:      .12 + random.next()*.76,
		Length: math.Max(4.5, length*(.045+random.next()*.11)*styleLength*morph.primitiveScale),
		Width:  math.Max(2.5, beam*(.1+random.next()*.28)*styleWidth),
		Sweep:  -sweepRange + random.next()*sweepRange*2,
	}
}

// generateStartingScoutVisualGenome mirrors the browser v4 generator's Scout
// profile, but resolves and persists the complete geometry at game creation.
// This prevents later generator changes from visually changing an existing ship.
func generateStartingScoutVisualGenome(seed string) core.ShipVisualGenome {
	chooser := newVisualPRNG(visualHashSeed(seed + "|full-random-v4"))
	morphologyID := startingScoutVisualMorphologies[int(math.Floor(chooser.next()*float64(len(startingScoutVisualMorphologies))))]
	styleID := startingScoutVisualStyles[int(math.Floor(chooser.next()*float64(len(startingScoutVisualStyles))))]

	styleLength, styleBeam := startingScoutStyleScale(styleID)
	morph := startingScoutMorphology(morphologyID)
	length := 54.0 * styleLength * morph.lengthScale
	beam := 8.5 * styleBeam * morph.beamScale
	random := newVisualPRNG(visualHashSeed(seed + "|scout|" + styleID + "|" + morphologyID + "|v4"))

	stationWidths := make([]float64, 7)
	for station := range stationWidths {
		t := float64(station) / 6
		if station == 6 {
			stationWidths[station] = 0
			continue
		}
		envelope := startingScoutEnvelope(morphologyID, t)
		endBias := (1-t)*morph.rearBias + t*morph.frontBias
		jitter := .8 + random.next()*.4
		startScale := 1.0
		if station == 0 {
			startScale = .72
		}
		stationWidths[station] = math.Max(2.2, beam*(.18+.82*envelope)*morph.coreDensity*endBias*jitter*startScale)
	}
	notchDepths := make([]float64, 7)
	engineCount := 1
	primitiveCount := 3 + int(math.Floor(random.next()*3))
	primitives := startingScoutSignaturePrimitives(morphologyID, beam, length)
	for i := 0; i < primitiveCount; i++ {
		primitives = append(primitives, startingScoutRandomPrimitive(random, beam, length, styleID, morphologyID))
	}

	cutoutCount := morph.cutoutBonus
	if cutoutCount > 3 {
		cutoutCount = 3
	}
	cutouts := make([]core.ShipVisualCutout, 0, cutoutCount)
	for index := 0; index < cutoutCount; index++ {
		isForkCenter := morphologyID == "fork" && index == 0
		cutout := core.ShipVisualCutout{}
		if isForkCenter {
			cutout = core.ShipVisualCutout{T: .66, Offset: 0, RX: math.Max(5, length*.09), RY: math.Max(3, beam*.22), Angle: 0}
		} else {
			cutout = core.ShipVisualCutout{
				T:      .34 + (float64(index+1)/float64(cutoutCount+1))*.4 + (random.next()-.5)*.05,
				Offset: beam * (.08 + random.next()*.16),
				RX:     math.Max(3.2, length*(.03+random.next()*.024)),
				RY:     math.Max(2, beam*(.1+random.next()*.08)),
				Angle:  -28 + random.next()*56,
			}
		}
		cutouts = append(cutouts, cutout)
	}

	return core.ShipVisualGenome{
		Version:       core.ShipVisualGenomeVersion,
		HullID:        "scout",
		StyleID:       styleID,
		MorphologyID:  morphologyID,
		Seed:          seed,
		Length:        length,
		Beam:          beam,
		StationCount:  7,
		StationWidths: stationWidths,
		NotchDepths:   notchDepths,
		EngineCount:   engineCount,
		DetailCount:   2,
		Cutouts:       cutouts,
		Primitives:    primitives,
	}
}
