package core

// DifficultyID is the stable authoritative identifier stored in game state and
// exchanged by New Game APIs. The display vocabulary is localized by clients.
type DifficultyID string

const (
	DifficultyEasy       DifficultyID = "easy"
	DifficultyNormal     DifficultyID = "normal"
	DifficultyHard       DifficultyID = "hard"
	DifficultyVeryHard   DifficultyID = "very_hard"
	DifficultyImpossible DifficultyID = "impossible"
)

var SupportedDifficultyIDs = []DifficultyID{
	DifficultyEasy,
	DifficultyNormal,
	DifficultyHard,
	DifficultyVeryHard,
	DifficultyImpossible,
}

func IsSupportedDifficultyID(id DifficultyID) bool {
	switch id {
	case DifficultyEasy, DifficultyNormal, DifficultyHard, DifficultyVeryHard, DifficultyImpossible:
		return true
	default:
		return false
	}
}

func (s *GameState) EffectiveDifficultyID() DifficultyID {
	if s == nil || s.DifficultyID == "" {
		return DifficultyNormal
	}
	return s.DifficultyID
}
