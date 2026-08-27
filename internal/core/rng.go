package core

import "fmt"

// RNG is a small deterministic SplitMix64 generator owned by the simulation.
// Its algorithm is intentionally implemented here so save/replay behavior does
// not depend on changes in Go's standard-library random-number generators.
type RNG struct {
	state uint64
}

func NewRNG(seed uint64) *RNG {
	return &RNG{state: seed}
}

func (r *RNG) State() uint64 {
	return r.state
}

func (r *RNG) SetState(state uint64) {
	r.state = state
}

func (r *RNG) Uint64() uint64 {
	r.state += 0x9e3779b97f4a7c15
	z := r.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

func (r *RNG) Intn(n int) (int, error) {
	if n <= 0 {
		return 0, fmt.Errorf("random bound must be positive, got %d", n)
	}
	bound := uint64(n)
	threshold := -bound % bound
	for {
		value := r.Uint64()
		if value >= threshold {
			return int(value % bound), nil
		}
	}
}
