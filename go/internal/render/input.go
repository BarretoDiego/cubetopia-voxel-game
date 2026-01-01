// Package render provides input handling
package render

import (
	"github.com/go-gl/glfw/v3.3/glfw"
)

// Input handles keyboard and mouse input
type Input struct {
	// Keyboard state
	keys map[glfw.Key]bool

	// Mouse state
	mouseButtons map[glfw.MouseButton]bool

	// Mouse position
	mouseX, mouseY         float64
	lastMouseX, lastMouseY float64
	firstMouse             bool

	// Mouse delta
	mouseDeltaX, mouseDeltaY float64

	// Scroll
	scrollX, scrollY float64
}

// NewInput creates a new input handler
func NewInput() *Input {
	return &Input{
		keys:         make(map[glfw.Key]bool),
		mouseButtons: make(map[glfw.MouseButton]bool),
		firstMouse:   true,
	}
}

// HandleKey processes keyboard events
func (i *Input) HandleKey(key glfw.Key, action glfw.Action) {
	if action == glfw.Press {
		i.keys[key] = true
	} else if action == glfw.Release {
		i.keys[key] = false
	}
}

// HandleMouseMove processes mouse movement
func (i *Input) HandleMouseMove(xpos, ypos float64) {
	if i.firstMouse {
		i.lastMouseX = xpos
		i.lastMouseY = ypos
		i.firstMouse = false
	}

	i.mouseDeltaX = xpos - i.lastMouseX
	i.mouseDeltaY = i.lastMouseY - ypos // Y is inverted

	i.lastMouseX = xpos
	i.lastMouseY = ypos
	i.mouseX = xpos
	i.mouseY = ypos
}

// HandleMouseButton processes mouse button events
func (i *Input) HandleMouseButton(button glfw.MouseButton, action glfw.Action) {
	if action == glfw.Press {
		i.mouseButtons[button] = true
	} else if action == glfw.Release {
		i.mouseButtons[button] = false
	}
}

// HandleScroll processes scroll events
func (i *Input) HandleScroll(xoff, yoff float64) {
	i.scrollX = xoff
	i.scrollY = yoff
}

// IsKeyPressed returns true if a key is currently pressed
func (i *Input) IsKeyPressed(key glfw.Key) bool {
	return i.keys[key]
}

// IsMouseButtonPressed returns true if a mouse button is pressed
func (i *Input) IsMouseButtonPressed(button glfw.MouseButton) bool {
	return i.mouseButtons[button]
}

// GetMousePosition returns current mouse position
func (i *Input) GetMousePosition() (x, y float64) {
	return i.mouseX, i.mouseY
}

// GetMouseDelta returns mouse movement since last frame and resets it
func (i *Input) GetMouseDelta() (dx, dy float64) {
	dx = i.mouseDeltaX
	dy = i.mouseDeltaY
	i.mouseDeltaX = 0
	i.mouseDeltaY = 0
	return
}

// GetScroll returns scroll wheel movement and resets it
func (i *Input) GetScroll() (x, y float64) {
	x = i.scrollX
	y = i.scrollY
	i.scrollX = 0
	i.scrollY = 0
	return
}

// ResetMouse resets mouse state (call when resuming from pause)
func (i *Input) ResetMouse() {
	i.firstMouse = true
	i.mouseDeltaX = 0
	i.mouseDeltaY = 0
}
