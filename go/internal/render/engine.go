// Package render provides OpenGL rendering for the voxel engine
package render

import (
	"fmt"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

func init() {
	// GLFW requires the main thread
	runtime.LockOSThread()
}

// Engine is the main rendering engine
type Engine struct {
	window *glfw.Window
	width  int
	height int

	// Camera
	camera *Camera

	// Shaders
	voxelShader *Shader

	// Input state
	input *Input

	// Timing
	lastFrame float64
	deltaTime float32

	// Callbacks
	onUpdate func(dt float32)
	onRender func()
	onResize func(width, height int)
}

// Config contains engine configuration
type Config struct {
	Width      int
	Height     int
	Title      string
	Fullscreen bool
	VSync      bool
}

// DefaultConfig returns default engine configuration
func DefaultConfig() Config {
	return Config{
		Width:      1280,
		Height:     720,
		Title:      "Voxel Engine",
		Fullscreen: false,
		VSync:      true,
	}
}

// NewEngine creates a new rendering engine
func NewEngine(config Config) (*Engine, error) {
	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize GLFW: %w", err)
	}

	// Configure GLFW
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Anti-aliasing
	glfw.WindowHint(glfw.Samples, 4)

	// Create window
	var monitor *glfw.Monitor
	if config.Fullscreen {
		monitor = glfw.GetPrimaryMonitor()
	}

	window, err := glfw.CreateWindow(config.Width, config.Height, config.Title, monitor, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}

	window.MakeContextCurrent()

	// VSync
	if config.VSync {
		glfw.SwapInterval(1)
	} else {
		glfw.SwapInterval(0)
	}

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize OpenGL: %w", err)
	}

	// Print OpenGL version
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Printf("OpenGL version: %s\n", version)

	// Configure OpenGL
	gl.Enable(gl.DEPTH_TEST)
	// Enable face culling for performance
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.Enable(gl.MULTISAMPLE)

	// Blending for transparent blocks
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Clear color (sky blue)
	gl.ClearColor(0.6, 0.8, 1.0, 1.0)

	engine := &Engine{
		window: window,
		width:  config.Width,
		height: config.Height,
		camera: NewCamera(mgl32.Vec3{0, 50, 0}),
		input:  NewInput(),
	}

	// Set up callbacks
	window.SetFramebufferSizeCallback(engine.framebufferSizeCallback)
	window.SetKeyCallback(engine.keyCallback)
	window.SetCursorPosCallback(engine.cursorPosCallback)
	window.SetMouseButtonCallback(engine.mouseButtonCallback)
	window.SetScrollCallback(engine.scrollCallback)

	// Start with cursor visible for menus (will be captured when game starts)
	window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)

	return engine, nil
}

// LoadShaders loads the voxel shader
func (e *Engine) LoadShaders() error {
	shader, err := NewShader(voxelVertexShader, voxelFragmentShader)
	if err != nil {
		return fmt.Errorf("failed to create voxel shader: %w", err)
	}
	e.voxelShader = shader
	return nil
}

// Run starts the main game loop
func (e *Engine) Run(onUpdate func(dt float32), onRender func()) {
	e.onUpdate = onUpdate
	e.onRender = onRender
	e.lastFrame = glfw.GetTime()

	for !e.window.ShouldClose() {
		// Calculate delta time
		currentFrame := glfw.GetTime()
		e.deltaTime = float32(currentFrame - e.lastFrame)
		e.lastFrame = currentFrame

		// Limit delta time to prevent physics issues
		if e.deltaTime > 0.1 {
			e.deltaTime = 0.1
		}

		// Poll events
		glfw.PollEvents()

		// Process input
		// Process input - REMOVED to avoid conflict with Game input handling
		// e.processInput()

		// Update
		if e.onUpdate != nil {
			e.onUpdate(e.deltaTime)
		}

		// Clear buffers
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Render
		if e.onRender != nil {
			e.onRender()
		}

		// Swap buffers
		e.window.SwapBuffers()
	}
}

// Cleanup releases resources
func (e *Engine) Cleanup() {
	if e.voxelShader != nil {
		e.voxelShader.Delete()
	}
	glfw.Terminate()
}

// GetCamera returns the camera
func (e *Engine) GetCamera() *Camera {
	return e.camera
}

// GetInput returns the input state
func (e *Engine) GetInput() *Input {
	return e.input
}

// GetDeltaTime returns the current frame delta time
func (e *Engine) GetDeltaTime() float32 {
	return e.deltaTime
}

// GetViewProjection returns the combined view-projection matrix
func (e *Engine) GetViewProjection() mgl32.Mat4 {
	view := e.camera.GetViewMatrix()
	projection := mgl32.Perspective(
		mgl32.DegToRad(e.camera.FOV),
		float32(e.width)/float32(e.height),
		0.1, 1000.0,
	)
	return projection.Mul4(view)
}

// UseVoxelShader activates the voxel shader with uniforms
func (e *Engine) UseVoxelShader() {
	if e.voxelShader == nil {
		return
	}

	e.voxelShader.Use()

	view := e.camera.GetViewMatrix()
	projection := mgl32.Perspective(
		mgl32.DegToRad(e.camera.FOV),
		float32(e.width)/float32(e.height),
		0.1, 1000.0,
	)

	e.voxelShader.SetMat4("uView", view)
	e.voxelShader.SetMat4("uProjection", projection)
	e.voxelShader.SetVec3("uCameraPos", e.camera.Position)
	e.voxelShader.SetVec3("uSunDirection", mgl32.Vec3{0.5, 0.8, 0.3}.Normalize())
	e.voxelShader.SetFloat("uTime", float32(glfw.GetTime()))
}

// Callbacks

func (e *Engine) framebufferSizeCallback(w *glfw.Window, width, height int) {
	e.width = width
	e.height = height
	gl.Viewport(0, 0, int32(width), int32(height))

	if e.onResize != nil {
		e.onResize(width, height)
	}
}

func (e *Engine) keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	// Don't auto-close on ESC - let game handle it
	e.input.HandleKey(key, action)
}

func (e *Engine) cursorPosCallback(w *glfw.Window, xpos, ypos float64) {
	e.input.HandleMouseMove(xpos, ypos)
}

func (e *Engine) mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	e.input.HandleMouseButton(button, action)
}

func (e *Engine) scrollCallback(w *glfw.Window, xoff, yoff float64) {
	e.input.HandleScroll(xoff, yoff)
}

// SetCursorMode sets the cursor mode (normal for menus, disabled for gameplay)
func (e *Engine) SetCursorMode(disabled bool) {
	if disabled {
		e.window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	} else {
		e.window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	}
}

// CloseWindow closes the game window
func (e *Engine) CloseWindow() {
	e.window.SetShouldClose(true)
}

func (e *Engine) processInput() {
	// Camera movement
	moveDir := mgl32.Vec3{0, 0, 0}

	if e.input.IsKeyPressed(glfw.KeyW) {
		moveDir = moveDir.Add(e.camera.Front)
	}
	if e.input.IsKeyPressed(glfw.KeyS) {
		moveDir = moveDir.Sub(e.camera.Front)
	}
	if e.input.IsKeyPressed(glfw.KeyA) {
		moveDir = moveDir.Sub(e.camera.Right)
	}
	if e.input.IsKeyPressed(glfw.KeyD) {
		moveDir = moveDir.Add(e.camera.Right)
	}

	// Normalize horizontal movement
	if moveDir.Len() > 0 {
		moveDir = mgl32.Vec3{moveDir.X(), 0, moveDir.Z()}.Normalize()
	}

	speed := float32(10.0)
	if e.input.IsKeyPressed(glfw.KeyLeftShift) {
		speed *= 1.8
	}

	e.camera.Position = e.camera.Position.Add(moveDir.Mul(speed * e.deltaTime))

	// Mouse look
	dx, dy := e.input.GetMouseDelta()
	if dx != 0 || dy != 0 {
		e.camera.ProcessMouseMovement(float32(dx), float32(-dy))
	}
}

// Embedded shaders

var voxelVertexShader = `
#version 410 core

layout(location = 0) in vec3 aPosition;
layout(location = 1) in vec3 aNormal;
layout(location = 2) in vec3 aColor;
layout(location = 3) in float aAO;

uniform mat4 uProjection;
uniform mat4 uView;

out vec3 vColor;
out vec3 vNormal;
out float vAO;
out vec3 vWorldPos;

void main() {
    vColor = aColor;
    vNormal = aNormal;
    vAO = aAO;
    vWorldPos = aPosition;
    gl_Position = uProjection * uView * vec4(aPosition, 1.0);
}
` + "\x00"

var voxelFragmentShader = `
#version 410 core

in vec3 vColor;
in vec3 vNormal;
in float vAO;
in vec3 vWorldPos;

uniform vec3 uSunDirection;
uniform vec3 uCameraPos;
uniform float uTime;

out vec4 fragColor;

void main() {
    // Basic directional lighting
    float diffuse = max(dot(vNormal, uSunDirection), 0.0);
    float ambient = 0.3;
    
    // Apply AO
    float ao = 1.0 - vAO * 0.25;
    
    vec3 lighting = vColor * (ambient + diffuse * 0.7) * ao;
    
    // Distance fog (Linear)
    float dist = length(uCameraPos - vWorldPos);
    float fogStart = 30.0;
    float fogEnd = 75.0; // Slightly less than 5 chunks (80)
    float fogFactor = clamp((dist - fogStart) / (fogEnd - fogStart), 0.0, 1.0);
    
    vec3 fogColor = vec3(0.6, 0.8, 1.0); // Match clear color
    
    vec3 finalColor = mix(lighting, fogColor, fogFactor);
    
    fragColor = vec4(finalColor, 1.0);
}
` + "\x00"
