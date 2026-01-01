// Package noise provides Fractal Brownian Motion for natural terrain
package noise

import (
	"math"
)

// FBMConfig contains configuration for FBM noise
type FBMConfig struct {
	Octaves     int     // Number of noise layers
	Lacunarity  float64 // Frequency multiplier per octave
	Persistence float64 // Amplitude multiplier per octave
	Scale       float64 // Base scale
	OffsetX     float64 // X offset
	OffsetZ     float64 // Z offset
}

// DefaultFBMConfig returns a default FBM configuration
func DefaultFBMConfig() FBMConfig {
	return FBMConfig{
		Octaves:     6,
		Lacunarity:  2.0,
		Persistence: 0.5,
		Scale:       1.0,
		OffsetX:     0,
		OffsetZ:     0,
	}
}

// FBM implements Fractal Brownian Motion for natural-looking terrain
type FBM struct {
	Config FBMConfig
}

// NewFBM creates a new FBM generator with the given configuration
func NewFBM(config FBMConfig) *FBM {
	return &FBM{Config: config}
}

// Sample2D samples FBM noise in 2D
// Returns a value in the approximate range [-1, 1]
func (f *FBM) Sample2D(noise *SimplexNoise, x, z float64) float64 {
	value := 0.0
	amplitude := 1.0
	frequency := f.Config.Scale
	maxValue := 0.0

	for i := 0; i < f.Config.Octaves; i++ {
		value += amplitude * noise.Noise2D(
			(x+f.Config.OffsetX)*frequency,
			(z+f.Config.OffsetZ)*frequency,
		)
		maxValue += amplitude
		amplitude *= f.Config.Persistence
		frequency *= f.Config.Lacunarity
	}

	return value / maxValue
}

// Sample3D samples FBM noise in 3D
// Returns a value in the approximate range [-1, 1]
func (f *FBM) Sample3D(noise *SimplexNoise, x, y, z float64) float64 {
	value := 0.0
	amplitude := 1.0
	frequency := f.Config.Scale
	maxValue := 0.0

	for i := 0; i < f.Config.Octaves; i++ {
		value += amplitude * noise.Noise3D(
			(x+f.Config.OffsetX)*frequency,
			y*frequency,
			(z+f.Config.OffsetZ)*frequency,
		)
		maxValue += amplitude
		amplitude *= f.Config.Persistence
		frequency *= f.Config.Lacunarity
	}

	return value / maxValue
}

// Ridged2D samples ridged FBM noise (for mountains)
// Creates sharp ridges by inverting and squaring the absolute value
func (f *FBM) Ridged2D(noise *SimplexNoise, x, z float64) float64 {
	value := 0.0
	amplitude := 1.0
	frequency := f.Config.Scale
	maxValue := 0.0

	for i := 0; i < f.Config.Octaves; i++ {
		n := noise.Noise2D(
			(x+f.Config.OffsetX)*frequency,
			(z+f.Config.OffsetZ)*frequency,
		)
		n = 1 - math.Abs(n) // Ridge
		n = n * n           // Sharpen
		value += amplitude * n
		maxValue += amplitude
		amplitude *= f.Config.Persistence
		frequency *= f.Config.Lacunarity
	}

	return value / maxValue
}

// Turbulence2D samples turbulent FBM noise (for clouds, erosion)
// Uses absolute value of noise for always-positive contribution
func (f *FBM) Turbulence2D(noise *SimplexNoise, x, z float64) float64 {
	value := 0.0
	amplitude := 1.0
	frequency := f.Config.Scale
	maxValue := 0.0

	for i := 0; i < f.Config.Octaves; i++ {
		value += amplitude * math.Abs(noise.Noise2D(
			(x+f.Config.OffsetX)*frequency,
			(z+f.Config.OffsetZ)*frequency,
		))
		maxValue += amplitude
		amplitude *= f.Config.Persistence
		frequency *= f.Config.Lacunarity
	}

	return value / maxValue
}

// Warped2D samples domain-warped FBM for more interesting terrain
// Uses FBM to distort the input coordinates before sampling
func (f *FBM) Warped2D(noise *SimplexNoise, x, z, warpAmount float64) float64 {
	warpX := f.Sample2D(noise, x*0.5, z*0.5) * warpAmount
	warpZ := f.Sample2D(noise, x*0.5+100, z*0.5+100) * warpAmount
	return f.Sample2D(noise, x+warpX, z+warpZ)
}
