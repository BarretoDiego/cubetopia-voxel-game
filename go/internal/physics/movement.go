// Package physics provides enhanced player movement
package physics

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// MovementMode defines the current movement mode
type MovementMode int

const (
	ModeWalking MovementMode = iota
	ModeSwimming
	ModeFlying
	ModeSprinting
	ModeCrouching
)

// EnhancedMovement provides improved player movement
type EnhancedMovement struct {
	// State
	Mode         MovementMode
	IsUnderwater bool
	IsCrouching  bool
	HeadBob      float32
	StepTime     float32

	// Movement modifiers
	SprintMultiplier float32
	CrouchMultiplier float32
	SwimMultiplier   float32
	FlyMultiplier    float32

	// Smooth camera
	CameraSway float32
	CameraLean float32
	TargetLean float32
	LeanSpeed  float32

	// Stamina
	Stamina        float32
	MaxStamina     float32
	StaminaDrain   float32
	StaminaRecover float32

	// Movement physics
	Acceleration float32
	Deceleration float32
	AirControl   float32
	MaxVelocity  float32
}

// NewEnhancedMovement creates enhanced movement controller
func NewEnhancedMovement() *EnhancedMovement {
	return &EnhancedMovement{
		Mode:             ModeWalking,
		SprintMultiplier: 1.8,
		CrouchMultiplier: 0.5,
		SwimMultiplier:   0.7,
		FlyMultiplier:    2.0,
		LeanSpeed:        8.0,
		Stamina:          100.0,
		MaxStamina:       100.0,
		StaminaDrain:     15.0,
		StaminaRecover:   10.0,
		Acceleration:     50.0,
		Deceleration:     30.0,
		AirControl:       0.3,
		MaxVelocity:      20.0,
	}
}

// Update updates movement state
func (em *EnhancedMovement) Update(dt float32, player *Player, isMoving bool) {
	// Update mode
	em.updateMode(player)

	// Update stamina
	em.updateStamina(dt, player.IsSprinting)

	// Update head bob
	em.updateHeadBob(dt, isMoving, player)

	// Update camera lean
	em.updateCameraLean(dt)
}

func (em *EnhancedMovement) updateMode(player *Player) {
	if player.IsFlying {
		em.Mode = ModeFlying
	} else if em.IsUnderwater {
		em.Mode = ModeSwimming
	} else if em.IsCrouching {
		em.Mode = ModeCrouching
	} else if player.IsSprinting && em.Stamina > 0 {
		em.Mode = ModeSprinting
	} else {
		em.Mode = ModeWalking
	}
}

func (em *EnhancedMovement) updateStamina(dt float32, isSprinting bool) {
	if isSprinting && em.Mode == ModeSprinting {
		em.Stamina -= em.StaminaDrain * dt
		if em.Stamina < 0 {
			em.Stamina = 0
		}
	} else {
		em.Stamina += em.StaminaRecover * dt
		if em.Stamina > em.MaxStamina {
			em.Stamina = em.MaxStamina
		}
	}
}

func (em *EnhancedMovement) updateHeadBob(dt float32, isMoving bool, player *Player) {
	if !isMoving || player.IsFlying || !player.IsOnGround {
		em.HeadBob *= 0.9 // Smooth decay
		return
	}

	bobSpeed := float32(8.0)
	if em.Mode == ModeSprinting {
		bobSpeed = 12.0
	} else if em.Mode == ModeCrouching {
		bobSpeed = 5.0
	}

	em.StepTime += dt * bobSpeed
	em.HeadBob = float32(math.Sin(float64(em.StepTime))) * 0.05
}

func (em *EnhancedMovement) updateCameraLean(dt float32) {
	// Smooth lerp to target lean
	lerpSpeed := em.LeanSpeed * dt
	em.CameraLean += (em.TargetLean - em.CameraLean) * lerpSpeed
}

// GetSpeedMultiplier returns the current speed multiplier
func (em *EnhancedMovement) GetSpeedMultiplier() float32 {
	switch em.Mode {
	case ModeSprinting:
		return em.SprintMultiplier
	case ModeCrouching:
		return em.CrouchMultiplier
	case ModeSwimming:
		return em.SwimMultiplier
	case ModeFlying:
		return em.FlyMultiplier
	default:
		return 1.0
	}
}

// SetLean sets the target camera lean (-1 to 1)
func (em *EnhancedMovement) SetLean(lean float32) {
	em.TargetLean = clampFloat(lean, -1.0, 1.0) * 5.0 // 5 degrees max
}

// GetHeadBobOffset returns the current head bob Y offset
func (em *EnhancedMovement) GetHeadBobOffset() float32 {
	return em.HeadBob
}

// GetCameraLean returns the current camera lean in degrees
func (em *EnhancedMovement) GetCameraLean() float32 {
	return em.CameraLean
}

// CanSprint returns true if the player can sprint
func (em *EnhancedMovement) CanSprint() bool {
	return em.Stamina > 10.0 && !em.IsUnderwater && !em.IsCrouching
}

// GetStaminaPercent returns stamina as percentage
func (em *EnhancedMovement) GetStaminaPercent() float32 {
	return em.Stamina / em.MaxStamina * 100.0
}

// SetUnderwater sets underwater state
func (em *EnhancedMovement) SetUnderwater(underwater bool) {
	em.IsUnderwater = underwater
}

// ToggleCrouch toggles crouching
func (em *EnhancedMovement) ToggleCrouch() {
	em.IsCrouching = !em.IsCrouching
}

// ApplySwimPhysics applies swimming physics modifiers
func (em *EnhancedMovement) ApplySwimPhysics(velocity *mgl32.Vec3, dt float32) {
	if !em.IsUnderwater {
		return
	}

	// Reduce gravity underwater
	if velocity.Y() < 0 {
		newY := velocity.Y() * 0.95 // Slow fall
		*velocity = mgl32.Vec3{velocity.X(), newY, velocity.Z()}
	}

	// Add buoyancy force
	buoyancy := float32(2.0) * dt
	*velocity = velocity.Add(mgl32.Vec3{0, buoyancy, 0})

	// Water resistance
	*velocity = velocity.Mul(0.98)
}

func clampFloat(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
