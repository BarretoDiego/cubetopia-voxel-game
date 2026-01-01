// Package render provides FPS camera for the voxel engine
package render

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// Camera represents a first-person camera
type Camera struct {
	// Position
	Position mgl32.Vec3

	// Orientation
	Front mgl32.Vec3
	Up    mgl32.Vec3
	Right mgl32.Vec3

	// Euler angles (in degrees)
	Yaw   float32
	Pitch float32

	// Options
	FOV         float32
	Sensitivity float32
}

// NewCamera creates a new camera at the given position
func NewCamera(position mgl32.Vec3) *Camera {
	c := &Camera{
		Position:    position,
		Up:          mgl32.Vec3{0, 1, 0},
		Yaw:         -90.0, // Looking towards negative Z
		Pitch:       0.0,
		FOV:         75.0,
		Sensitivity: 0.1,
	}
	c.updateVectors()
	return c
}

// GetViewMatrix returns the view matrix for this camera
func (c *Camera) GetViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(c.Position, c.Position.Add(c.Front), c.Up)
}

// ProcessMouseMovement handles mouse movement for looking around
func (c *Camera) ProcessMouseMovement(xoffset, yoffset float32) {
	xoffset *= c.Sensitivity
	yoffset *= c.Sensitivity

	c.Yaw += xoffset
	c.Pitch += yoffset

	// Constrain pitch to prevent flipping
	if c.Pitch > 89.0 {
		c.Pitch = 89.0
	}
	if c.Pitch < -89.0 {
		c.Pitch = -89.0
	}

	c.updateVectors()
}

// ProcessScroll handles scroll wheel for FOV zoom
func (c *Camera) ProcessScroll(yoffset float32) {
	c.FOV -= yoffset
	if c.FOV < 1.0 {
		c.FOV = 1.0
	}
	if c.FOV > 120.0 {
		c.FOV = 120.0
	}
}

// SetPosition sets the camera position
func (c *Camera) SetPosition(pos mgl32.Vec3) {
	c.Position = pos
}

// SetRotation sets the camera rotation (yaw, pitch in degrees)
func (c *Camera) SetRotation(yaw, pitch float32) {
	c.Yaw = yaw
	c.Pitch = pitch

	if c.Pitch > 89.0 {
		c.Pitch = 89.0
	}
	if c.Pitch < -89.0 {
		c.Pitch = -89.0
	}

	c.updateVectors()
}

// GetForward returns the forward direction (X-Z plane only)
func (c *Camera) GetForward() mgl32.Vec3 {
	forward := mgl32.Vec3{c.Front.X(), 0, c.Front.Z()}
	if forward.Len() > 0 {
		return forward.Normalize()
	}
	return mgl32.Vec3{0, 0, -1}
}

func (c *Camera) updateVectors() {
	// Calculate front vector
	yawRad := float64(c.Yaw) * math.Pi / 180.0
	pitchRad := float64(c.Pitch) * math.Pi / 180.0

	front := mgl32.Vec3{
		float32(math.Cos(yawRad) * math.Cos(pitchRad)),
		float32(math.Sin(pitchRad)),
		float32(math.Sin(yawRad) * math.Cos(pitchRad)),
	}
	c.Front = front.Normalize()

	// Recalculate right and up vectors
	worldUp := mgl32.Vec3{0, 1, 0}
	c.Right = c.Front.Cross(worldUp).Normalize()
	c.Up = c.Right.Cross(c.Front).Normalize()
}
