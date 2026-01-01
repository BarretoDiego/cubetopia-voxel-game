// Package physics provides raycasting for block interaction
package physics

import (
	"math"

	"voxelgame/internal/core/block"

	"github.com/go-gl/mathgl/mgl32"
)

// RaycastResult contains information about a raycast hit
type RaycastResult struct {
	Hit       bool
	Position  mgl32.Vec3 // Hit position
	BlockPos  [3]int     // Block coordinates
	Normal    mgl32.Vec3 // Surface normal
	Face      string     // Face name: "top", "bottom", "left", "right", "front", "back"
	BlockType block.Type // Type of block hit
	Distance  float32    // Distance to hit
}

// Raycast performs a raycast from origin in direction, up to maxDistance
func Raycast(origin, direction mgl32.Vec3, maxDistance float32, getBlock BlockGetter) RaycastResult {
	result := RaycastResult{}

	if getBlock == nil {
		return result
	}

	// Normalize direction
	dir := direction.Normalize()

	// Starting block
	x := int(math.Floor(float64(origin.X())))
	y := int(math.Floor(float64(origin.Y())))
	z := int(math.Floor(float64(origin.Z())))

	// Direction signs
	stepX := 1
	if dir.X() < 0 {
		stepX = -1
	}
	stepY := 1
	if dir.Y() < 0 {
		stepY = -1
	}
	stepZ := 1
	if dir.Z() < 0 {
		stepZ = -1
	}

	// Calculate t values for each axis
	// How far along the ray until we hit the next block boundary
	var tMaxX, tMaxY, tMaxZ float32
	var tDeltaX, tDeltaY, tDeltaZ float32

	if dir.X() != 0 {
		if stepX > 0 {
			tMaxX = (float32(x+1) - origin.X()) / dir.X()
		} else {
			tMaxX = (float32(x) - origin.X()) / dir.X()
		}
		tDeltaX = float32(math.Abs(1.0 / float64(dir.X())))
	} else {
		tMaxX = 1e30
		tDeltaX = 1e30
	}

	if dir.Y() != 0 {
		if stepY > 0 {
			tMaxY = (float32(y+1) - origin.Y()) / dir.Y()
		} else {
			tMaxY = (float32(y) - origin.Y()) / dir.Y()
		}
		tDeltaY = float32(math.Abs(1.0 / float64(dir.Y())))
	} else {
		tMaxY = 1e30
		tDeltaY = 1e30
	}

	if dir.Z() != 0 {
		if stepZ > 0 {
			tMaxZ = (float32(z+1) - origin.Z()) / dir.Z()
		} else {
			tMaxZ = (float32(z) - origin.Z()) / dir.Z()
		}
		tDeltaZ = float32(math.Abs(1.0 / float64(dir.Z())))
	} else {
		tMaxZ = 1e30
		tDeltaZ = 1e30
	}

	// Track which face we entered through
	var lastFace string
	var lastNormal mgl32.Vec3

	// Traverse the grid
	distance := float32(0)
	for distance < maxDistance {
		// Check current block
		blockType := getBlock(x, y, z)
		if blockType != block.Air && blockType.IsSolid() {
			result.Hit = true
			result.BlockPos = [3]int{x, y, z}
			result.Position = origin.Add(dir.Mul(distance))
			result.Normal = lastNormal
			result.Face = lastFace
			result.BlockType = blockType
			result.Distance = distance
			return result
		}

		// Step to next block
		if tMaxX < tMaxY {
			if tMaxX < tMaxZ {
				x += stepX
				distance = tMaxX
				tMaxX += tDeltaX
				if stepX > 0 {
					lastFace = "left"
					lastNormal = mgl32.Vec3{-1, 0, 0}
				} else {
					lastFace = "right"
					lastNormal = mgl32.Vec3{1, 0, 0}
				}
			} else {
				z += stepZ
				distance = tMaxZ
				tMaxZ += tDeltaZ
				if stepZ > 0 {
					lastFace = "back"
					lastNormal = mgl32.Vec3{0, 0, -1}
				} else {
					lastFace = "front"
					lastNormal = mgl32.Vec3{0, 0, 1}
				}
			}
		} else {
			if tMaxY < tMaxZ {
				y += stepY
				distance = tMaxY
				tMaxY += tDeltaY
				if stepY > 0 {
					lastFace = "bottom"
					lastNormal = mgl32.Vec3{0, -1, 0}
				} else {
					lastFace = "top"
					lastNormal = mgl32.Vec3{0, 1, 0}
				}
			} else {
				z += stepZ
				distance = tMaxZ
				tMaxZ += tDeltaZ
				if stepZ > 0 {
					lastFace = "back"
					lastNormal = mgl32.Vec3{0, 0, -1}
				} else {
					lastFace = "front"
					lastNormal = mgl32.Vec3{0, 0, 1}
				}
			}
		}
	}

	return result
}

// GetPlacementPosition returns the position where a block would be placed
func GetPlacementPosition(hit RaycastResult) [3]int {
	pos := hit.BlockPos

	switch hit.Face {
	case "top":
		pos[1]++
	case "bottom":
		pos[1]--
	case "left":
		pos[0]--
	case "right":
		pos[0]++
	case "front":
		pos[2]++
	case "back":
		pos[2]--
	}

	return pos
}
