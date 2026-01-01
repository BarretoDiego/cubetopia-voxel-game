// Package physics provides physics simulation for the voxel engine
package physics

import (
	"math"

	"voxelgame/internal/core/block"

	"github.com/go-gl/mathgl/mgl32"
)

// Constants for player physics
const (
	Gravity          = 32.0
	JumpForce        = 12.0
	PlayerSpeed      = 10.0
	SprintMultiplier = 1.8

	// Player dimensions (AABB)
	PlayerWidth     = 0.6
	PlayerHeight    = 1.8
	PlayerEyeHeight = 1.6
)

// BlockGetter is a function that returns the block at world coordinates
type BlockGetter func(x, y, z int) block.Type

// Player represents the player with physics
type Player struct {
	// Position (eye position)
	Position mgl32.Vec3

	// Velocity
	Velocity mgl32.Vec3

	// Rotation (yaw, pitch in degrees)
	Yaw   float32
	Pitch float32

	// State
	IsOnGround  bool
	IsSprinting bool
	IsFlying    bool // Noclip mode

	// Movement input
	moveForward float32
	moveRight   float32
	wantJump    bool

	// Block getter for collision
	getBlock BlockGetter
}

// NewPlayer creates a new player at the given position
func NewPlayer(position mgl32.Vec3, getBlock BlockGetter) *Player {
	return &Player{
		Position: position,
		Yaw:      -90.0,
		getBlock: getBlock,
	}
}

// SetMovement sets the movement input
func (p *Player) SetMovement(forward, right float32, sprint, jump bool) {
	p.moveForward = forward
	p.moveRight = right
	p.IsSprinting = sprint
	p.wantJump = jump
}

// SetRotation sets the player rotation
func (p *Player) SetRotation(yaw, pitch float32) {
	p.Yaw = yaw
	p.Pitch = pitch

	// Clamp pitch
	if p.Pitch > 89.0 {
		p.Pitch = 89.0
	}
	if p.Pitch < -89.0 {
		p.Pitch = -89.0
	}
}

// Update updates player physics
func (p *Player) Update(dt float32) {
	if p.IsFlying {
		p.updateFlying(dt)
	} else {
		p.updateNormal(dt)
	}
}

func (p *Player) updateNormal(dt float32) {
	// Calculate movement direction based on yaw
	yawRad := float64(p.Yaw) * math.Pi / 180.0

	// Forward and right vectors (horizontal only)
	forward := mgl32.Vec3{
		float32(math.Cos(yawRad)),
		0,
		float32(math.Sin(yawRad)),
	}
	right := mgl32.Vec3{
		float32(-math.Sin(yawRad)),
		0,
		float32(math.Cos(yawRad)),
	}

	// Calculate desired horizontal velocity
	moveDir := forward.Mul(p.moveForward).Add(right.Mul(p.moveRight))
	if moveDir.Len() > 0 {
		moveDir = moveDir.Normalize()
	}

	speed := float32(PlayerSpeed)
	if p.IsSprinting {
		speed *= SprintMultiplier
	}

	// Set horizontal velocity
	p.Velocity[0] = moveDir.X() * speed
	p.Velocity[2] = moveDir.Z() * speed

	// Apply gravity
	p.Velocity[1] -= Gravity * dt

	// Jumping
	if p.wantJump && p.IsOnGround {
		p.Velocity[1] = JumpForce
		p.IsOnGround = false
	}

	// Move with collision
	p.moveWithCollision(dt)
}

func (p *Player) updateFlying(dt float32) {
	// Calculate movement direction based on yaw and pitch
	yawRad := float64(p.Yaw) * math.Pi / 180.0
	pitchRad := float64(p.Pitch) * math.Pi / 180.0

	// Full 3D forward vector
	forward := mgl32.Vec3{
		float32(math.Cos(yawRad) * math.Cos(pitchRad)),
		float32(math.Sin(pitchRad)),
		float32(math.Sin(yawRad) * math.Cos(pitchRad)),
	}
	right := mgl32.Vec3{
		float32(-math.Sin(yawRad)),
		0,
		float32(math.Cos(yawRad)),
	}

	moveDir := forward.Mul(p.moveForward).Add(right.Mul(p.moveRight))

	speed := float32(PlayerSpeed * 2) // Faster in fly mode
	if p.IsSprinting {
		speed *= SprintMultiplier
	}

	// Direct position update (no collision)
	p.Position = p.Position.Add(moveDir.Mul(speed * dt))
}

func (p *Player) moveWithCollision(dt float32) {
	// Try to move in each axis separately
	newPos := p.Position

	// Move X
	newPos[0] += p.Velocity[0] * dt
	if p.checkCollision(newPos) {
		newPos[0] = p.Position[0]
		p.Velocity[0] = 0
	}

	// Move Z
	newPos[2] += p.Velocity[2] * dt
	if p.checkCollision(newPos) {
		newPos[2] = p.Position[2]
		p.Velocity[2] = 0
	}

	// Move Y
	newPos[1] += p.Velocity[1] * dt

	// Check ground collision
	if p.Velocity[1] < 0 {
		if p.checkGroundCollision(newPos) {
			// Snap to ground
			groundY := float32(math.Floor(float64(newPos[1]-PlayerEyeHeight)+0.5)) + PlayerEyeHeight
			newPos[1] = groundY
			p.Velocity[1] = 0
			p.IsOnGround = true
		} else {
			p.IsOnGround = false
		}
	}

	// Check ceiling collision
	if p.Velocity[1] > 0 {
		if p.checkCeilingCollision(newPos) {
			p.Velocity[1] = 0
		}
	}

	// Clamp to world bounds
	if newPos[1] < 2 {
		newPos[1] = 2
		p.Velocity[1] = 0
		p.IsOnGround = true
	}
	if newPos[1] > 62 {
		newPos[1] = 62
	}

	p.Position = newPos
}

// checkCollision checks if the player would collide at the given position
func (p *Player) checkCollision(pos mgl32.Vec3) bool {
	if p.getBlock == nil {
		return false
	}

	// Check multiple points around the player
	halfWidth := float32(PlayerWidth / 2)
	checkPoints := [][3]float32{
		{0, 0, 0},
		{-halfWidth, 0, -halfWidth},
		{halfWidth, 0, -halfWidth},
		{-halfWidth, 0, halfWidth},
		{halfWidth, 0, halfWidth},
	}

	for _, offset := range checkPoints {
		bx := int(math.Floor(float64(pos.X() + offset[0])))
		by := int(math.Floor(float64(pos.Y() - PlayerEyeHeight + 0.5 + offset[1])))
		bz := int(math.Floor(float64(pos.Z() + offset[2])))

		blockType := p.getBlock(bx, by, bz)
		if blockType.IsCollidable() {
			return true
		}

		// Also check at body height
		by = int(math.Floor(float64(pos.Y() - PlayerEyeHeight + 1.0 + offset[1])))
		blockType = p.getBlock(bx, by, bz)
		if blockType.IsCollidable() {
			return true
		}
	}

	return false
}

// checkGroundCollision checks for ground below the player
func (p *Player) checkGroundCollision(pos mgl32.Vec3) bool {
	if p.getBlock == nil {
		return false
	}

	halfWidth := float32(PlayerWidth / 2)
	feetY := pos.Y() - PlayerEyeHeight

	checkPoints := [][2]float32{
		{0, 0},
		{-halfWidth, -halfWidth},
		{halfWidth, -halfWidth},
		{-halfWidth, halfWidth},
		{halfWidth, halfWidth},
	}

	for _, offset := range checkPoints {
		bx := int(math.Floor(float64(pos.X() + offset[0])))
		by := int(math.Floor(float64(feetY)))
		bz := int(math.Floor(float64(pos.Z() + offset[1])))

		blockType := p.getBlock(bx, by, bz)
		if blockType.IsCollidable() {
			return true
		}
	}

	return false
}

// checkCeilingCollision checks for ceiling above the player
func (p *Player) checkCeilingCollision(pos mgl32.Vec3) bool {
	if p.getBlock == nil {
		return false
	}

	bx := int(math.Floor(float64(pos.X())))
	by := int(math.Floor(float64(pos.Y() + 0.2)))
	bz := int(math.Floor(float64(pos.Z())))

	blockType := p.getBlock(bx, by, bz)
	return blockType.IsCollidable()
}

// GetFeetPosition returns the position of the player's feet
func (p *Player) GetFeetPosition() mgl32.Vec3 {
	return mgl32.Vec3{p.Position.X(), p.Position.Y() - PlayerEyeHeight, p.Position.Z()}
}

// GetLookDirection returns the direction the player is looking
func (p *Player) GetLookDirection() mgl32.Vec3 {
	yawRad := float64(p.Yaw) * math.Pi / 180.0
	pitchRad := float64(p.Pitch) * math.Pi / 180.0

	return mgl32.Vec3{
		float32(math.Cos(yawRad) * math.Cos(pitchRad)),
		float32(math.Sin(pitchRad)),
		float32(math.Sin(yawRad) * math.Cos(pitchRad)),
	}
}

// ToggleFlyMode toggles fly/noclip mode
func (p *Player) ToggleFlyMode() {
	p.IsFlying = !p.IsFlying
	if p.IsFlying {
		p.Velocity = mgl32.Vec3{0, 0, 0}
	}
}
