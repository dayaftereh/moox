package core

type OrbitalBodyKind string

const (
	OrbitalBodyAsteroidBelt OrbitalBodyKind = "asteroid_belt"
	OrbitalBodyGasGiant     OrbitalBodyKind = "gas_giant"
	OrbitalBodyPlanet       OrbitalBodyKind = "planet"
)

// OrbitalBody gives every visible generated system body a stable authoritative
// identity. Normal planets intentionally reuse their Planet ID as Body ID so
// existing Planet/Fleet/Colony identifiers remain stable; non-colonizable
// bodies receive their own IDs.
type OrbitalBody struct {
	ID        ID              `json:"id"`
	Name      string          `json:"name"`
	Orbit     int             `json:"orbit"`
	Kind      OrbitalBodyKind `json:"kind"`
	PlanetID  ID              `json:"planet_id,omitempty"`
	OutpostID ID              `json:"outpost_id,omitempty"`
}
