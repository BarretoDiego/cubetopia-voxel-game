// Package math provides mathematical utilities including seeded RNG
package math

// SeededRNG is a Linear Congruential Generator for deterministic random numbers
type SeededRNG struct {
	state uint64
	m     uint64
	a     uint64
	c     uint64
}

// NewSeededRNG creates a new seeded random number generator
func NewSeededRNG(seed int64) *SeededRNG {
	return &SeededRNG{
		state: uint64(seed),
		m:     0x80000000, // 2^31
		a:     1103515245,
		c:     12345,
	}
}

// Next returns a random float64 in [0, 1)
func (r *SeededRNG) Next() float64 {
	r.state = (r.a*r.state + r.c) % r.m
	return float64(r.state) / float64(r.m)
}

// NextInt returns a random integer in [min, max]
func (r *SeededRNG) NextInt(min, max int) int {
	return min + int(r.Next()*float64(max-min+1))
}

// NextFloat returns a random float64 in [min, max)
func (r *SeededRNG) NextFloat(min, max float64) float64 {
	return min + r.Next()*(max-min)
}

// NextBool returns true with the given probability
func (r *SeededRNG) NextBool(probability float64) bool {
	return r.Next() < probability
}

// Choose returns a random element from the slice
func Choose[T any](r *SeededRNG, items []T) T {
	return items[r.NextInt(0, len(items)-1)]
}

// Shuffle returns a shuffled copy of the slice using Fisher-Yates algorithm
func Shuffle[T any](r *SeededRNG, items []T) []T {
	result := make([]T, len(items))
	copy(result, items)
	for i := len(result) - 1; i > 0; i-- {
		j := r.NextInt(0, i)
		result[i], result[j] = result[j], result[i]
	}
	return result
}
