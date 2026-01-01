// Package render provides OpenGL rendering for the voxel engine
package render

import (
	"fmt"
	"math"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// ... (init function)

// Engine is the main rendering engine
type Engine struct {
	window *glfw.Window
	width  int
	height int

	// Camera
	camera *Camera

	// Shaders
	voxelShader *Shader

	// Texture Manager
	textureManager *TextureManager

	// Particle System
	particleSystem *ParticleSystem

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
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.FrontFace(gl.CW)
	gl.Enable(gl.MULTISAMPLE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.ClearColor(0.6, 0.8, 1.0, 1.0)

	tm := NewTextureManager()
	// Define standard texture list - Mapping to IDs 0, 1, 2...
	textureFiles := []string{
		"assets/textures/dirt.png",       // 0
		"assets/textures/grass_top.png",  // 1
		"assets/textures/grass_side.png", // 2
		"assets/textures/stone.png",      // 3
		"assets/textures/wood.png",       // 4
		"assets/textures/leaves.png",     // 5
		"assets/textures/water.png",      // 6
		"assets/textures/ice.png",        // 7
		"assets/textures/sand.png",       // 8
		"assets/textures/snow.png",       // 9
		"assets/textures/glass.png",      // 10
	}
	err = tm.LoadBlockTextures(textureFiles)
	if err != nil {
		fmt.Printf("Error loading textures: %v\n", err)
	}

	ps, err := NewParticleSystem(1000)
	if err != nil {
		fmt.Printf("Error initializing particle system: %v\n", err)
	}

	engine := &Engine{
		window:         window,
		width:          config.Width,
		height:         config.Height,
		camera:         NewCamera(mgl32.Vec3{0, 50, 0}),
		input:          NewInput(),
		textureManager: tm,
		particleSystem: ps,
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
	// Read shader files
	vShaderFile := "assets/shaders/voxel.vert"
	fShaderFile := "assets/shaders/voxel.frag"

	vSource, err := os.ReadFile(vShaderFile)
	if err != nil {
		return fmt.Errorf("failed to read vertex shader %s: %w", vShaderFile, err)
	}

	fSource, err := os.ReadFile(fShaderFile)
	if err != nil {
		return fmt.Errorf("failed to read fragment shader %s: %w", fShaderFile, err)
	}

	shader, err := NewShader(string(vSource), string(fSource))
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

		// Update Particles
		if e.particleSystem != nil {
			e.particleSystem.Update(e.deltaTime)
		}

		// Clear buffers
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Render Voxels
		if e.onRender != nil {
			e.onRender()
		}

		// Render Particles (Last for transparency)
		if e.particleSystem != nil {
			// vp := e.GetViewProjection() // Unused
			view := e.camera.GetViewMatrix()
			projection := mgl32.Perspective(
				mgl32.DegToRad(e.camera.FOV),
				float32(e.width)/float32(e.height),
				0.1, 1000.0,
			)
			e.particleSystem.Render(view, projection)
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

// GetParticleSystem returns the particle system
func (e *Engine) GetParticleSystem() *ParticleSystem {
	return e.particleSystem
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

	// Wind uniforms
	// Simple wind direction variation
	time := float64(glfw.GetTime())
	windStr := float32(0.3) + float32(0.1*math.Sin(time*0.5)) // Fluctuating wind
	windDir := mgl32.Vec3{1.0, 0.0, 0.5}.Normalize()
	e.voxelShader.SetFloat("uWindStrength", windStr)
	e.voxelShader.SetVec3("uWindDir", windDir)

	// Bind Textures
	e.textureManager.BindBlockTextures(0)
	e.voxelShader.SetInt("uBlockAtlas", 0) // Texture Unit 0
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

// Embedded shaders removed - now loaded from assets/shaders/
