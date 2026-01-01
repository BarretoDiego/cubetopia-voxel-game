// Voxel Game - Main entry point
// A high-performance voxel engine written in Go with OpenGL rendering
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"voxelgame/internal/core/block"
	"voxelgame/internal/physics"
	"voxelgame/internal/render"
	"voxelgame/internal/save"
	"voxelgame/internal/ui"
	"voxelgame/internal/world"
)

// Game holds all game state
type Game struct {
	// Core systems
	engine           *render.Engine
	world            *world.World
	player           *physics.Player
	movement         *physics.EnhancedMovement
	sky              *render.Sky
	postProcess      *render.PostProcess
	creatureRenderer *render.CreatureRenderer
	raytracer        *render.RaytracingRenderer
	underwater       *render.UnderwaterEffect

	// UI
	uiRenderer   *ui.Renderer
	inventory    *ui.Inventory
	stateManager *ui.GameStateManager
	mainMenu     *ui.Menu
	pauseMenu    *ui.Menu
	menuRenderer *ui.MenuRenderer
	settings     *ui.Settings
	settingsMenu *ui.SettingsMenu

	// Save system
	saveManager *save.Manager

	// Config
	screenWidth  int
	screenHeight int

	// Stats
	fps         int
	frameCount  int
	lastFPSTime time.Time

	// Block interaction
	targetBlock *physics.RaycastResult

	// Key state for single press detection
	lastKeyStates map[glfw.Key]bool
}

func main() {
	fmt.Println("=================================")
	fmt.Println("  Voxel Engine - Go Edition")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("Controls:")
	fmt.Println("  WASD       - Move")
	fmt.Println("  Mouse      - Look around")
	fmt.Println("  Shift      - Sprint")
	fmt.Println("  Space      - Jump")
	fmt.Println("  Ctrl       - Crouch")
	fmt.Println("  F          - Toggle fly mode")
	fmt.Println("  R          - Toggle raytracing")
	fmt.Println("  1-9        - Select hotbar slot")
	fmt.Println("  Scroll     - Cycle hotbar")
	fmt.Println("  LMB        - Break block")
	fmt.Println("  RMB        - Place block")
	fmt.Println("  F3         - Toggle debug")
	fmt.Println("  F5         - Quick save")
	fmt.Println("  F9         - Quick load")
	fmt.Println("  ESC/P      - Pause/Menu")
	fmt.Println()

	game, err := NewGame()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create game: %v\n", err)
		os.Exit(1)
	}
	defer game.Cleanup()

	game.Run()
}

// NewGame creates a new game instance
func NewGame() (*Game, error) {
	g := &Game{
		lastFPSTime:   time.Now(),
		screenWidth:   1280,
		screenHeight:  720,
		lastKeyStates: make(map[glfw.Key]bool),
	}

	// Create settings
	g.settings = ui.DefaultSettings()

	// Create rendering engine
	config := render.DefaultConfig()
	config.Title = "Voxel Engine - Go Edition"
	config.Width = g.screenWidth
	config.Height = g.screenHeight

	engine, err := render.NewEngine(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create engine: %w", err)
	}
	g.engine = engine

	// Load shaders
	if err := engine.LoadShaders(); err != nil {
		return nil, fmt.Errorf("failed to load shaders: %w", err)
	}

	// Create UI renderer first (needed for menus)
	uiRenderer, err := ui.NewRenderer(config.Width, config.Height)
	if err != nil {
		fmt.Printf("Warning: Failed to create UI renderer: %v\n", err)
	}
	g.uiRenderer = uiRenderer

	// Create game state manager
	g.stateManager = ui.NewGameStateManager()

	// Create menus
	g.setupMenus()

	// Create menu renderer
	g.menuRenderer = ui.NewMenuRenderer(g.uiRenderer)

	// Create settings menu
	g.settingsMenu = ui.NewSettingsMenu(g.settings)

	// Create save manager
	g.saveManager = save.NewManager()

	// Create sky renderer
	sky, err := render.NewSky()
	if err != nil {
		fmt.Printf("Warning: Failed to create sky renderer: %v\n", err)
	}
	g.sky = sky

	// Create post-processing
	postProcess, err := render.NewPostProcess(config.Width, config.Height)
	if err != nil {
		fmt.Printf("Warning: Failed to create post-processing: %v\n", err)
	}
	g.postProcess = postProcess

	// Create raytracing renderer
	raytracer := render.NewRaytracingRenderer(config.Width, config.Height)
	raytracer.SetEnabled(g.settings.EnableRaytracing)
	g.raytracer = raytracer

	// Create creature renderer
	creatureRenderer, err := render.NewCreatureRenderer()
	if err != nil {
		fmt.Printf("Warning: Failed to create creature renderer: %v\n", err)
	}
	g.creatureRenderer = creatureRenderer

	// Create underwater effect
	underwater, err := render.NewUnderwaterEffect()
	if err != nil {
		fmt.Printf("Warning: Failed to create underwater effect: %v\n", err)
	}
	g.underwater = underwater

	// Create inventory
	g.inventory = ui.NewInventory()

	// World will be created when starting new game
	g.world = nil
	g.player = nil

	return g, nil
}

func (g *Game) setupMenus() {
	g.mainMenu = ui.NewMainMenu(
		g.startNewGame, // New Game
		g.loadGame,     // Load Game
		g.openSettings, // Settings
		g.quitGame,     // Quit
	)

	g.pauseMenu = ui.NewPauseMenu(
		g.resumeGame,       // Resume
		g.openSettings,     // Settings
		g.saveGame,         // Save
		g.returnToMainMenu, // Main Menu
	)
}

// Menu callbacks
func (g *Game) startNewGame() {
	fmt.Println("[DEBUG] startNewGame called")

	// Create world with random seed
	seed := time.Now().UnixNano()
	fmt.Printf("World seed: %d\n", seed)

	g.world = world.NewWorld(seed)

	// Get spawn position
	spawnX, spawnY, spawnZ := g.world.GetSpawnPosition()
	spawnPos := mgl32.Vec3{float32(spawnX), float32(spawnY), float32(spawnZ)}
	fmt.Printf("Spawn position: %.1f, %.1f, %.1f\n", spawnX, spawnY, spawnZ)

	// Create player with physics
	g.player = physics.NewPlayer(spawnPos, func(x, y, z int) block.Type {
		return g.world.GetBlock(x, y, z)
	})

	// Create enhanced movement
	g.movement = physics.NewEnhancedMovement()

	// Switch to playing state
	g.stateManager.SetState(ui.StatePlaying)
	g.mainMenu.IsVisible = false

	// Capture mouse for FPS controls
	g.engine.SetCursorMode(true)
	fmt.Println("[DEBUG] Entered playing state")
}

func (g *Game) loadGame() {
	data, err := g.saveManager.QuickLoad()
	if err != nil {
		fmt.Printf("Failed to load: %v\n", err)
		return
	}

	// Create world with saved seed
	g.world = world.NewWorld(data.World.Seed)

	// Create player at saved position
	spawnPos := mgl32.Vec3{data.Player.PositionX, data.Player.PositionY, data.Player.PositionZ}
	g.player = physics.NewPlayer(spawnPos, func(x, y, z int) block.Type {
		return g.world.GetBlock(x, y, z)
	})
	g.player.Yaw = data.Player.Yaw
	g.player.Pitch = data.Player.Pitch

	// Create enhanced movement
	g.movement = physics.NewEnhancedMovement()

	// Switch to playing state
	g.stateManager.SetState(ui.StatePlaying)
	g.mainMenu.IsVisible = false

	// Capture mouse for FPS controls
	g.engine.SetCursorMode(true)
	fmt.Println("[DEBUG] Game loaded, entered playing state")
}

func (g *Game) openSettings() {
	g.stateManager.SetState(ui.StateSettings)
	g.settingsMenu.IsVisible = true
}

func (g *Game) quitGame() {
	fmt.Println("[DEBUG] quitGame called")
	g.engine.CloseWindow()
}

func (g *Game) resumeGame() {
	fmt.Println("[DEBUG] resumeGame called")
	g.stateManager.SetState(ui.StatePlaying)
	g.pauseMenu.IsVisible = false
	g.engine.SetCursorMode(true) // Capture mouse again
}

func (g *Game) saveGame() {
	if g.player == nil || g.world == nil {
		return
	}

	err := g.saveManager.QuickSave(save.PlayerSave{
		PositionX: g.player.Position.X(),
		PositionY: g.player.Position.Y(),
		PositionZ: g.player.Position.Z(),
		Yaw:       g.player.Yaw,
		Pitch:     g.player.Pitch,
	}, g.world.Seed)

	if err != nil {
		fmt.Printf("Failed to save: %v\n", err)
	} else {
		fmt.Println("Game saved!")
	}
}

func (g *Game) returnToMainMenu() {
	fmt.Println("[DEBUG] returnToMainMenu called")
	// Cleanup world
	if g.world != nil {
		g.world.Cleanup()
		g.world = nil
	}
	g.player = nil

	g.stateManager.SetState(ui.StateMainMenu)
	g.mainMenu.IsVisible = true
	g.pauseMenu.IsVisible = false

	// Show cursor for menu
	g.engine.SetCursorMode(false)
}

// Run starts the game loop
func (g *Game) Run() {
	g.engine.Run(g.Update, g.Render)
}

// Update updates the game state
func (g *Game) Update(dt float32) {
	input := g.engine.GetInput()

	// FPS counter
	g.frameCount++
	if time.Since(g.lastFPSTime) >= time.Second {
		g.fps = g.frameCount
		g.frameCount = 0
		g.lastFPSTime = time.Now()
	}

	// Handle state-specific input
	switch g.stateManager.CurrentState {
	case ui.StateMainMenu:
		g.updateMainMenu(input)
	case ui.StatePlaying:
		g.updatePlaying(input, dt)
	case ui.StatePaused:
		g.updatePaused(input)
	case ui.StateSettings:
		g.updateSettings(input)
	}
}

func (g *Game) updateMainMenu(input *render.Input) {
	if g.wasKeyJustPressed(input, glfw.KeyUp) || g.wasKeyJustPressed(input, glfw.KeyW) {
		g.mainMenu.SelectPrevious()
	}
	if g.wasKeyJustPressed(input, glfw.KeyDown) || g.wasKeyJustPressed(input, glfw.KeyS) {
		g.mainMenu.SelectNext()
	}
	if g.wasKeyJustPressed(input, glfw.KeyEnter) || g.wasKeyJustPressed(input, glfw.KeySpace) {
		g.mainMenu.Confirm()
	}
}

func (g *Game) updatePlaying(input *render.Input, dt float32) {
	if g.player == nil || g.world == nil {
		return
	}

	// Pause
	if g.wasKeyJustPressed(input, glfw.KeyEscape) || g.wasKeyJustPressed(input, glfw.KeyP) {
		fmt.Println("[DEBUG] Pausing game")
		g.stateManager.SetState(ui.StatePaused)
		g.pauseMenu.IsVisible = true
		g.engine.SetCursorMode(false) // Show cursor for menu
		return
	}

	// Toggle debug
	if g.wasKeyJustPressed(input, glfw.KeyF3) {
		g.settings.EnablePostProcess = !g.settings.EnablePostProcess
	}

	// Toggle raytracing
	if g.wasKeyJustPressed(input, glfw.KeyR) {
		g.settings.EnableRaytracing = !g.settings.EnableRaytracing
		g.raytracer.SetEnabled(g.settings.EnableRaytracing)
		if g.settings.EnableRaytracing {
			fmt.Println("Raytracing enabled")
		} else {
			fmt.Println("Raytracing disabled")
		}
	}

	// Toggle fly mode
	if g.wasKeyJustPressed(input, glfw.KeyF) {
		g.player.ToggleFlyMode()
		if g.player.IsFlying {
			fmt.Println("Fly mode enabled")
		} else {
			fmt.Println("Fly mode disabled")
		}
	}

	// Quick save (F5)
	if g.wasKeyJustPressed(input, glfw.KeyF5) {
		g.saveGame()
	}

	// Quick load (F9)
	if g.wasKeyJustPressed(input, glfw.KeyF9) {
		g.loadGame()
	}

	// Hotbar selection
	for i := 0; i < 9; i++ {
		if input.IsKeyPressed(glfw.Key(int(glfw.Key1) + i)) {
			g.inventory.SelectSlot(i)
		}
	}

	// Scroll for hotbar
	_, scrollY := input.GetScroll()
	if scrollY != 0 {
		g.inventory.ScrollSelection(int(-scrollY))
	}

	// Toggle crouch
	if g.wasKeyJustPressed(input, glfw.KeyLeftControl) {
		g.movement.ToggleCrouch()
	}

	// Update camera from player rotation
	camera := g.engine.GetCamera()
	camera.SetPosition(g.player.Position)
	camera.SetRotation(g.player.Yaw, g.player.Pitch)

	// Get movement input
	var forward, right float32
	if input.IsKeyPressed(glfw.KeyW) {
		forward = 1
	}
	if input.IsKeyPressed(glfw.KeyS) {
		forward = -1
	}
	if input.IsKeyPressed(glfw.KeyA) {
		right = -1
	}
	if input.IsKeyPressed(glfw.KeyD) {
		right = 1
	}

	// Set camera lean based on strafe
	g.movement.SetLean(right * 0.5)

	sprint := input.IsKeyPressed(glfw.KeyLeftShift) && g.movement.CanSprint()
	jump := input.IsKeyPressed(glfw.KeySpace)

	// Handle mouse look
	dx, dy := input.GetMouseDelta()
	sens := g.settings.MouseSensitivity
	if g.settings.InvertY {
		dy = -dy
	}
	if dx != 0 || dy != 0 {
		g.player.SetRotation(g.player.Yaw+float32(dx)*sens, g.player.Pitch+float32(dy)*sens)
	}

	// Check if underwater
	playerBlockY := int(g.player.Position.Y())
	playerBlock := g.world.GetBlock(int(g.player.Position.X()), playerBlockY, int(g.player.Position.Z()))
	isUnderwater := playerBlock == block.Water
	g.movement.SetUnderwater(isUnderwater)
	if g.underwater != nil {
		g.underwater.IsUnderwater = isUnderwater
	}

	// Update enhanced movement
	isMoving := forward != 0 || right != 0
	g.movement.Update(dt, g.player, isMoving)

	// Apply movement modifiers
	g.player.SetMovement(forward*g.movement.GetSpeedMultiplier(), right*g.movement.GetSpeedMultiplier(), sprint, jump)

	// Apply swim physics
	if isUnderwater {
		g.movement.ApplySwimPhysics(&g.player.Velocity, dt)
	}

	// Update player physics
	g.player.Update(dt)

	// Update world around player
	g.world.Update(
		float64(g.player.Position.X()),
		float64(g.player.Position.Y()),
		float64(g.player.Position.Z()),
	)

	// Update sky
	if g.sky != nil {
		g.sky.Update(dt)
	}

	// Update underwater effect
	if g.underwater != nil {
		g.underwater.Update(dt)
	}

	// Raycast for block selection
	lookDir := g.player.GetLookDirection()
	result := physics.Raycast(g.player.Position, lookDir, 5.0, func(x, y, z int) block.Type {
		return g.world.GetBlock(x, y, z)
	})
	if result.Hit {
		g.targetBlock = &result
	} else {
		g.targetBlock = nil
	}

	// Handle block interaction
	if input.IsMouseButtonPressed(glfw.MouseButtonLeft) && g.targetBlock != nil {
		g.world.SetBlock(g.targetBlock.BlockPos[0], g.targetBlock.BlockPos[1], g.targetBlock.BlockPos[2], block.Air)
	}
	if input.IsMouseButtonPressed(glfw.MouseButtonRight) && g.targetBlock != nil {
		placePos := physics.GetPlacementPosition(*g.targetBlock)
		selectedBlock := g.inventory.GetSelectedBlock()
		if selectedBlock != block.Air {
			g.world.SetBlock(placePos[0], placePos[1], placePos[2], selectedBlock)
		}
	}
}

func (g *Game) updatePaused(input *render.Input) {
	if g.wasKeyJustPressed(input, glfw.KeyEscape) || g.wasKeyJustPressed(input, glfw.KeyP) {
		g.resumeGame()
		return
	}
	if g.wasKeyJustPressed(input, glfw.KeyUp) || g.wasKeyJustPressed(input, glfw.KeyW) {
		g.pauseMenu.SelectPrevious()
	}
	if g.wasKeyJustPressed(input, glfw.KeyDown) || g.wasKeyJustPressed(input, glfw.KeyS) {
		g.pauseMenu.SelectNext()
	}
	if g.wasKeyJustPressed(input, glfw.KeyEnter) || g.wasKeyJustPressed(input, glfw.KeySpace) {
		g.pauseMenu.Confirm()
	}
}

func (g *Game) updateSettings(input *render.Input) {
	if g.wasKeyJustPressed(input, glfw.KeyEscape) {
		g.settingsMenu.IsVisible = false
		g.stateManager.SetState(g.stateManager.PreviousState)
		return
	}
	if g.wasKeyJustPressed(input, glfw.KeyUp) || g.wasKeyJustPressed(input, glfw.KeyW) {
		g.settingsMenu.SelectPrevious()
	}
	if g.wasKeyJustPressed(input, glfw.KeyDown) || g.wasKeyJustPressed(input, glfw.KeyS) {
		g.settingsMenu.SelectNext()
	}
	if g.wasKeyJustPressed(input, glfw.KeyLeft) || g.wasKeyJustPressed(input, glfw.KeyA) {
		g.settingsMenu.ToggleCurrentSetting(-1)
	}
	if g.wasKeyJustPressed(input, glfw.KeyRight) || g.wasKeyJustPressed(input, glfw.KeyD) {
		g.settingsMenu.ToggleCurrentSetting(1)
	}
	if g.wasKeyJustPressed(input, glfw.KeyEnter) || g.wasKeyJustPressed(input, glfw.KeySpace) {
		g.settingsMenu.ToggleCurrentSetting(1)
	}

	// Apply settings changes
	if g.raytracer != nil {
		g.raytracer.SetEnabled(g.settings.EnableRaytracing)
	}
	if g.postProcess != nil {
		g.postProcess.EnableFXAA = g.settings.EnableFXAA
		g.postProcess.EnableBloom = g.settings.EnableBloom
		g.postProcess.BloomStrength = g.settings.BloomStrength
	}
}

func (g *Game) wasKeyJustPressed(input *render.Input, key glfw.Key) bool {
	current := input.IsKeyPressed(key)
	last := g.lastKeyStates[key]
	g.lastKeyStates[key] = current
	return current && !last
}

// Render renders the game
func (g *Game) Render() {
	switch g.stateManager.CurrentState {
	case ui.StateMainMenu:
		g.renderMainMenu()
	case ui.StatePlaying:
		g.renderPlaying()
	case ui.StatePaused:
		g.renderPlaying() // Render world behind pause menu
		g.renderPauseMenu()
	case ui.StateSettings:
		if g.stateManager.PreviousState == ui.StatePlaying {
			g.renderPlaying()
		}
		g.renderSettings()
	}
}

func (g *Game) renderMainMenu() {
	// Simple background to verify rendering works
	if g.uiRenderer != nil {
		g.uiRenderer.BeginFrame()

		// Draw a large visible rectangle to test
		g.uiRenderer.DrawRect(100, 100, float32(g.screenWidth-200), float32(g.screenHeight-200), [4]float32{0.2, 0.3, 0.5, 1.0})

		// Draw title area
		g.uiRenderer.DrawRect(float32(g.screenWidth)/2-150, 150, 300, 60, [4]float32{0.3, 0.4, 0.7, 1.0})

		// Draw menu items
		for i := 0; i < 4; i++ {
			y := float32(250 + i*70)
			bgColor := [4]float32{0.15, 0.15, 0.25, 0.9}
			if i == g.mainMenu.SelectedIndex {
				bgColor = [4]float32{0.4, 0.5, 0.8, 1.0}
			}
			g.uiRenderer.DrawRect(float32(g.screenWidth)/2-120, y, 240, 50, bgColor)
		}

		// Instructions at bottom
		g.uiRenderer.DrawRect(float32(g.screenWidth)/2-200, float32(g.screenHeight)-80, 400, 30, [4]float32{0.1, 0.1, 0.15, 0.8})

		g.uiRenderer.EndFrame()
	}
}

func (g *Game) renderPlaying() {
	if g.world == nil {
		return
	}

	// Raytracing mode
	if g.settings.EnableRaytracing && g.raytracer != nil {
		sunDir := mgl32.Vec3{0.5, 0.8, 0.3}.Normalize()
		if g.sky != nil {
			sunDir = g.sky.GetSunDirection()
		}
		g.raytracer.Render(g.engine.GetCamera(), sunDir)
	} else {
		// Normal rendering
		// Render sky first
		if g.sky != nil {
			viewProj := g.engine.GetViewProjection()
			invViewProj := viewProj.Inv()
			g.sky.Render(invViewProj, g.engine.GetCamera().Position)
		}

		// Render world
		g.engine.UseVoxelShader()
		g.world.Render()

		// Render creatures
		if g.creatureRenderer != nil {
			camera := g.engine.GetCamera()
			view := camera.GetViewMatrix()
			projection := mgl32.Perspective(
				mgl32.DegToRad(g.settings.FOV),
				float32(g.screenWidth)/float32(g.screenHeight),
				0.1, 1000.0,
			)
			sunDir := mgl32.Vec3{0.5, 0.8, 0.3}.Normalize()
			if g.sky != nil {
				sunDir = g.sky.GetSunDirection()
			}
			g.creatureRenderer.RenderCreatures(g.world.GetCreatures(), view, projection, sunDir)
		}
	}

	// Render UI
	if g.uiRenderer != nil {
		g.uiRenderer.BeginFrame()

		// Crosshair
		g.uiRenderer.DrawCrosshair()

		// Hotbar
		g.uiRenderer.DrawHotbar(g.inventory.SelectedIndex, g.inventory.GetHotbarColors())

		// Debug panel
		if g.settings.EnablePostProcess {
			stats := g.world.GetStats()
			g.uiRenderer.DrawDebugPanel(ui.DebugInfo{
				Position:     g.player.Position,
				ChunksLoaded: stats.ChunksLoaded,
				FPS:          g.fps,
				Biome:        g.world.GetBiomeAt(int(g.player.Position.X()), int(g.player.Position.Z())),
			})
		}

		// Raytracing indicator
		if g.settings.EnableRaytracing {
			g.uiRenderer.DrawRect(10, float32(g.screenHeight-40), 120, 25, [4]float32{1, 0.5, 0, 0.8})
		}

		g.uiRenderer.EndFrame()
	}
}

func (g *Game) renderPauseMenu() {
	if g.uiRenderer != nil {
		g.uiRenderer.BeginFrame()
		g.menuRenderer.RenderMenu(g.pauseMenu, g.screenWidth, g.screenHeight)
		g.uiRenderer.EndFrame()
	}
}

func (g *Game) renderSettings() {
	if g.uiRenderer != nil {
		g.uiRenderer.BeginFrame()
		// Render settings as a simple menu-like display
		g.uiRenderer.DrawRect(0, 0, float32(g.screenWidth), float32(g.screenHeight), [4]float32{0, 0, 0, 0.8})

		// Settings title bar
		g.uiRenderer.DrawRect(float32(g.screenWidth)/2-200, 100, 400, 50, [4]float32{0.2, 0.3, 0.5, 1})

		// Settings items
		for i, item := range g.settingsMenu.Items {
			y := float32(180 + i*45)
			bgColor := [4]float32{0.15, 0.15, 0.2, 0.8}
			if i == g.settingsMenu.SelectedIndex {
				bgColor = [4]float32{0.3, 0.4, 0.6, 0.9}
			}
			g.uiRenderer.DrawRect(float32(g.screenWidth)/2-200, y, 400, 40, bgColor)

			// Setting name indicator
			nameWidth := float32(len(item.Name) * 8)
			g.uiRenderer.DrawRect(float32(g.screenWidth)/2-180, y+10, nameWidth, 20, [4]float32{0.7, 0.7, 0.7, 1})

			// Setting value indicator
			valueColor := [4]float32{0.3, 0.8, 0.3, 1}
			g.uiRenderer.DrawRect(float32(g.screenWidth)/2+100, y+10, 80, 20, valueColor)
		}

		g.uiRenderer.EndFrame()
	}
}

// Cleanup releases resources
func (g *Game) Cleanup() {
	fmt.Println("Cleaning up...")

	if g.raytracer != nil {
		g.raytracer.Cleanup()
	}
	if g.underwater != nil {
		g.underwater.Cleanup()
	}
	if g.creatureRenderer != nil {
		g.creatureRenderer.Cleanup()
	}
	if g.postProcess != nil {
		g.postProcess.Cleanup()
	}
	if g.sky != nil {
		g.sky.Cleanup()
	}
	if g.uiRenderer != nil {
		g.uiRenderer.Cleanup()
	}
	if g.world != nil {
		g.world.Cleanup()
	}
	if g.engine != nil {
		g.engine.Cleanup()
	}

	fmt.Println("Goodbye!")
}
