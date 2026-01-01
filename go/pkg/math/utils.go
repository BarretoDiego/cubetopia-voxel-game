// Package math provides mathematical utilities for procedural generation
package math

import (
	"math"
)

// Clamp restricts value between min and max
func Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// ClampInt restricts integer value between min and max
func ClampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Lerp performs linear interpolation between a and b
func Lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// Smoothstep performs smooth interpolation
func Smoothstep(edge0, edge1, x float64) float64 {
	t := Clamp((x-edge0)/(edge1-edge0), 0, 1)
	return t * t * (3 - 2*t)
}

// Smootherstep performs even smoother interpolation
func Smootherstep(edge0, edge1, x float64) float64 {
	t := Clamp((x-edge0)/(edge1-edge0), 0, 1)
	return t * t * t * (t*(t*6-15) + 10)
}

// Mod performs modulo that works correctly with negative numbers
func Mod(n, m int) int {
	return ((n % m) + m) % m
}

// ModFloat performs modulo for float64 that works correctly with negative numbers
func ModFloat(n, m float64) float64 {
	return math.Mod(math.Mod(n, m)+m, m)
}

// WorldToChunk converts world coordinates to chunk coordinates
func WorldToChunk(x, z float64, chunkSize int) (cx, cz int) {
	cx = int(math.Floor(x / float64(chunkSize)))
	cz = int(math.Floor(z / float64(chunkSize)))
	return
}

// WorldToLocal converts world coordinates to local chunk coordinates
func WorldToLocal(x, z float64, chunkSize int) (lx, lz int) {
	lx = Mod(int(math.Floor(x)), chunkSize)
	lz = Mod(int(math.Floor(z)), chunkSize)
	return
}

// Distance2D calculates euclidean distance in 2D
func Distance2D(x1, z1, x2, z2 float64) float64 {
	dx := x2 - x1
	dz := z2 - z1
	return math.Sqrt(dx*dx + dz*dz)
}

// Distance3D calculates euclidean distance in 3D
func Distance3D(x1, y1, z1, x2, y2, z2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	dz := z2 - z1
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// HashCoords generates a simple hash from coordinates for seeding
func HashCoords(x, y, z int) int {
	hash := 17
	hash = hash*31 + x
	hash = hash*31 + y
	hash = hash*31 + z
	return hash
}
