// Package noise provides procedural noise algorithms for terrain generation
package noise

import (
	"math"
)

// SimplexNoise implements 2D and 3D Simplex Noise
// Based on Ken Perlin's and Stefan Gustavson's algorithm
type SimplexNoise struct {
	seed      int64
	perm      [512]uint8
	permMod12 [512]uint8

	// Constants for 2D
	f2 float64
	g2 float64

	// Constants for 3D
	f3 float64
	g3 float64
}

// Gradients for 2D and 3D
var grad3 = [12][3]float64{
	{1, 1, 0}, {-1, 1, 0}, {1, -1, 0}, {-1, -1, 0},
	{1, 0, 1}, {-1, 0, 1}, {1, 0, -1}, {-1, 0, -1},
	{0, 1, 1}, {0, -1, 1}, {0, 1, -1}, {0, -1, -1},
}

// NewSimplexNoise creates a new Simplex Noise generator with the given seed
func NewSimplexNoise(seed int64) *SimplexNoise {
	s := &SimplexNoise{
		seed: seed,
		f2:   0.5 * (math.Sqrt(3.0) - 1.0),
		g2:   (3.0 - math.Sqrt(3.0)) / 6.0,
		f3:   1.0 / 3.0,
		g3:   1.0 / 6.0,
	}
	s.initPermutation()
	return s
}

func (s *SimplexNoise) initPermutation() {
	p := make([]uint8, 256)

	// Initialize with identity
	for i := 0; i < 256; i++ {
		p[i] = uint8(i)
	}

	// Fisher-Yates shuffle with seed
	seed := s.seed
	for i := 255; i > 0; i-- {
		seed = (seed * 16807) % 2147483647
		j := int(seed) % (i + 1)
		p[i], p[j] = p[j], p[i]
	}

	// Duplicate to avoid overflow
	for i := 0; i < 512; i++ {
		s.perm[i] = p[i&255]
		s.permMod12[i] = s.perm[i] % 12
	}
}

// Noise2D generates 2D Simplex Noise at the given coordinates
// Returns a value in the range [-1, 1]
func (s *SimplexNoise) Noise2D(xin, yin float64) float64 {
	var n0, n1, n2 float64

	// Skew input space
	t := (xin + yin) * s.f2
	i := int(math.Floor(xin + t))
	j := int(math.Floor(yin + t))

	// Unskew back
	t2 := float64(i+j) * s.g2
	x0 := xin - (float64(i) - t2)
	y0 := yin - (float64(j) - t2)

	// Determine which simplex
	var i1, j1 int
	if x0 > y0 {
		i1, j1 = 1, 0
	} else {
		i1, j1 = 0, 1
	}

	x1 := x0 - float64(i1) + s.g2
	y1 := y0 - float64(j1) + s.g2
	x2 := x0 - 1.0 + 2.0*s.g2
	y2 := y0 - 1.0 + 2.0*s.g2

	// Hash coordinates
	ii := i & 255
	jj := j & 255
	gi0 := int(s.permMod12[ii+int(s.perm[jj])])
	gi1 := int(s.permMod12[ii+i1+int(s.perm[jj+j1])])
	gi2 := int(s.permMod12[ii+1+int(s.perm[jj+1])])

	// Calculate contribution from each corner
	t0 := 0.5 - x0*x0 - y0*y0
	if t0 < 0 {
		n0 = 0.0
	} else {
		t0 *= t0
		n0 = t0 * t0 * (grad3[gi0][0]*x0 + grad3[gi0][1]*y0)
	}

	t1 := 0.5 - x1*x1 - y1*y1
	if t1 < 0 {
		n1 = 0.0
	} else {
		t1 *= t1
		n1 = t1 * t1 * (grad3[gi1][0]*x1 + grad3[gi1][1]*y1)
	}

	t3 := 0.5 - x2*x2 - y2*y2
	if t3 < 0 {
		n2 = 0.0
	} else {
		t3 *= t3
		n2 = t3 * t3 * (grad3[gi2][0]*x2 + grad3[gi2][1]*y2)
	}

	// Return value in range [-1, 1]
	return 70.0 * (n0 + n1 + n2)
}

// Noise3D generates 3D Simplex Noise at the given coordinates
// Returns a value in the range [-1, 1]
func (s *SimplexNoise) Noise3D(xin, yin, zin float64) float64 {
	var n0, n1, n2, n3 float64

	// Skew input space
	t := (xin + yin + zin) * s.f3
	i := int(math.Floor(xin + t))
	j := int(math.Floor(yin + t))
	k := int(math.Floor(zin + t))

	// Unskew back
	t2 := float64(i+j+k) * s.g3
	x0 := xin - (float64(i) - t2)
	y0 := yin - (float64(j) - t2)
	z0 := zin - (float64(k) - t2)

	// Determine which simplex
	var i1, j1, k1, i2, j2, k2 int

	if x0 >= y0 {
		if y0 >= z0 {
			i1, j1, k1 = 1, 0, 0
			i2, j2, k2 = 1, 1, 0
		} else if x0 >= z0 {
			i1, j1, k1 = 1, 0, 0
			i2, j2, k2 = 1, 0, 1
		} else {
			i1, j1, k1 = 0, 0, 1
			i2, j2, k2 = 1, 0, 1
		}
	} else {
		if y0 < z0 {
			i1, j1, k1 = 0, 0, 1
			i2, j2, k2 = 0, 1, 1
		} else if x0 < z0 {
			i1, j1, k1 = 0, 1, 0
			i2, j2, k2 = 0, 1, 1
		} else {
			i1, j1, k1 = 0, 1, 0
			i2, j2, k2 = 1, 1, 0
		}
	}

	x1 := x0 - float64(i1) + s.g3
	y1 := y0 - float64(j1) + s.g3
	z1 := z0 - float64(k1) + s.g3
	x2 := x0 - float64(i2) + 2.0*s.g3
	y2 := y0 - float64(j2) + 2.0*s.g3
	z2 := z0 - float64(k2) + 2.0*s.g3
	x3 := x0 - 1.0 + 3.0*s.g3
	y3 := y0 - 1.0 + 3.0*s.g3
	z3 := z0 - 1.0 + 3.0*s.g3

	// Hash
	ii := i & 255
	jj := j & 255
	kk := k & 255
	gi0 := int(s.permMod12[ii+int(s.perm[jj+int(s.perm[kk])])])
	gi1 := int(s.permMod12[ii+i1+int(s.perm[jj+j1+int(s.perm[kk+k1])])])
	gi2 := int(s.permMod12[ii+i2+int(s.perm[jj+j2+int(s.perm[kk+k2])])])
	gi3 := int(s.permMod12[ii+1+int(s.perm[jj+1+int(s.perm[kk+1])])])

	// Contributions
	t0 := 0.6 - x0*x0 - y0*y0 - z0*z0
	if t0 < 0 {
		n0 = 0.0
	} else {
		t0 *= t0
		n0 = t0 * t0 * (grad3[gi0][0]*x0 + grad3[gi0][1]*y0 + grad3[gi0][2]*z0)
	}

	t1 := 0.6 - x1*x1 - y1*y1 - z1*z1
	if t1 < 0 {
		n1 = 0.0
	} else {
		t1 *= t1
		n1 = t1 * t1 * (grad3[gi1][0]*x1 + grad3[gi1][1]*y1 + grad3[gi1][2]*z1)
	}

	t2val := 0.6 - x2*x2 - y2*y2 - z2*z2
	if t2val < 0 {
		n2 = 0.0
	} else {
		t2val *= t2val
		n2 = t2val * t2val * (grad3[gi2][0]*x2 + grad3[gi2][1]*y2 + grad3[gi2][2]*z2)
	}

	t3 := 0.6 - x3*x3 - y3*y3 - z3*z3
	if t3 < 0 {
		n3 = 0.0
	} else {
		t3 *= t3
		n3 = t3 * t3 * (grad3[gi3][0]*x3 + grad3[gi3][1]*y3 + grad3[gi3][2]*z3)
	}

	return 32.0 * (n0 + n1 + n2 + n3)
}
