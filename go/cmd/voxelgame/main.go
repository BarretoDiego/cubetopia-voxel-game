// Voxel Game - Main entry point
// A high-performance voxel engine written in Go with OpenGL rendering
package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"voxelgame/internal/core/block"
	"voxelgame/internal/generation/entity"
	"voxelgame/internal/generation/terrain"
	"voxelgame/internal/physics"
	"voxelgame/internal/render"
	"voxelgame/internal/save"
	"voxelgame/internal/ui"
	"voxelgame/internal/world"
)

// Build metadata - injected at build time via ldflags
var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
	GameName  = "Voxel Engine"
)

// Game holds all game state
type Game struct {
	// Core systems
	engine           *render.Engine
	world            *world.World
	player           *physics.Player
	playerModel      *entity.Creature
	movement         *physics.EnhancedMovement
	sky              *render.Sky
	postProcess      *render.PostProcess
	creatureRenderer *render.CreatureRenderer
	raytracer        *render.RaytracingRenderer
	underwater       *render.UnderwaterEffect
	blockOutline     *render.BlockOutlineRenderer
	blockBreaker     *render.BlockBreaker

	// UI
	uiRenderer   *ui.Renderer
	inventory    *ui.Inventory
	stateManager *ui.GameStateManager
	mainMenu     *ui.Menu
	pauseMenu    *ui.Menu
	menuRenderer *ui.MenuRenderer
	settings     *ui.Settings
	settingsMenu *ui.SettingsMenu
	minimap      *ui.Minimap

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

	// UI State
	showControls  bool
	showInventory bool
	controlsList  []string

	// Key state for single press detection
	lastKeyStates map[glfw.Key]bool

	// Mouse button state for single click detection
	lastMouseButtonStates map[glfw.MouseButton]bool
}

func main() {
	// core: crucial for OpenGL on macOS
	runtime.LockOSThread()

	fmt.Println("═══════════════════════════════════════════")
	fmt.Printf("  %s  v%s\n", GameName, Version)
	fmt.Println("═══════════════════════════════════════════")
	fmt.Printf("  Build:  %s\n", BuildDate)
	fmt.Printf("  Commit: %s\n", GitCommit)
	fmt.Println("═══════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Controls:")
	fmt.Println("  WASD       - Move")
	fmt.Println("  Mouse      - Look around")
	fmt.Println("  Shift      - Sprint")
	fmt.Println("  Space      - Jump")
	fmt.Println("  Ctrl       - Crouch")
	fmt.Println("  F          - Toggle fly mode")
	fmt.Println("  C          - Toggle camera (1st/3rd person)")
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
		lastFPSTime:           time.Now(),
		screenWidth:           1920,
		screenHeight:          1080,
		lastKeyStates:         make(map[glfw.Key]bool),
		lastMouseButtonStates: make(map[glfw.MouseButton]bool),
		showControls:          false,
		controlsList: []string{
			"WASD - Move",
			"Mouse - Look",
			"Space - Jump",
			"Shift - Sprint",
			"Ctrl - Crouch",
			"F - Fly Mode",
			"C - Camera View",
			"R - Raytracing",
			"LMB - Break",
			"RMB - Place",
			"1-9 - Hotbar",
			"I - Inventario",
			"H - Toggle Help",
			"F3 - Debug Info",
			"F5 - Quick Save",
			"F9 - Quick Load",
			"ESC - Pause",
		},
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

	// Create minimap
	g.minimap = ui.NewMinimap(128)

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

	// Create block outline renderer
	blockOutline, err := render.NewBlockOutlineRenderer()
	if err != nil {
		fmt.Printf("Warning: Failed to create block outline renderer: %v\n", err)
	}
	g.blockOutline = blockOutline

	// Create block breaker
	blockBreaker, err := render.NewBlockBreaker()
	if err != nil {
		fmt.Printf("Warning: Failed to create block breaker: %v\n", err)
	}
	g.blockBreaker = blockBreaker

	// Create inventory
	g.inventory = ui.NewInventory()

	// Start at main menu
	g.stateManager.SetState(ui.StateMainMenu)
	g.mainMenu.IsVisible = true
	g.engine.SetCursorMode(false) // Show cursor for menu navigation

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

	// Create player model for 3rd person view
	gen := entity.NewGenerator(seed)
	g.playerModel = gen.Create(entity.TemplateBiped, "plains", spawnPos, 1.0)
	g.playerModel.PrimaryColor = [3]float32{0.2, 0.2, 0.8}
	g.playerModel.SecondaryColor = [3]float32{0.1, 0.1, 0.1}

	// Create demo campfire structure
	// Ground height at 5,5
	fireX, fireZ := 5, 5
	fireY := g.world.GetHeight(fireX, fireZ)
	// Clear area
	g.world.SetBlock(fireX, fireY+1, fireZ, block.Air)
	g.world.SetBlock(fireX, fireY+2, fireZ, block.Air)
	// Base
	g.world.SetBlock(fireX, fireY, fireZ, block.Cobblestone)
	// "Logs" for fire (using OakLog horizontally if possible, but just block for now)
	// Since we don't have rotation yet, just place a log.
	// Or maybe 4 logs around a center fire?
	// Simple: Just a log block with fire on top.
	g.world.SetBlock(fireX, fireY+1, fireZ, block.OakLog)

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

	// Create player model for 3rd person view
	gen := entity.NewGenerator(data.World.Seed)
	g.playerModel = gen.Create(entity.TemplateBiped, "plains", spawnPos, 1.0)
	g.playerModel.PrimaryColor = [3]float32{0.2, 0.2, 0.8}
	g.playerModel.SecondaryColor = [3]float32{0.1, 0.1, 0.1}

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

	// Toggle camera mode
	if g.wasKeyJustPressed(input, glfw.KeyC) {
		camera := g.engine.GetCamera()
		camera.ThirdPerson = !camera.ThirdPerson
		if camera.ThirdPerson {
			fmt.Println("Third person view enabled")
		} else {
			fmt.Println("First person view enabled")
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

	// Toggle Controls Overlay (H)
	if g.wasKeyJustPressed(input, glfw.KeyH) {
		g.showControls = !g.showControls
	}

	// Toggle Inventory Panel (I)
	if g.wasKeyJustPressed(input, glfw.KeyI) {
		g.showInventory = !g.showInventory
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

	// Handle mouse look
	dx, dy := input.GetMouseDelta()
	// Adjusted sensitivity
	sens := g.settings.MouseSensitivity * 0.5
	// Flip Y-axis logic as requested (Standard: Up=Up. User wants opposite? Or maybe my previous 'InvertY' check was confusing)
	// Current logic: dy negative -> Pitch Increase -> Look Up.
	// If User says "Reverse it", I will negation.
	// New Input: Mouse Up -> dy negative.
	// We want Mouse Up -> Look Down? (If that's what "Reverse" means)
	// Or maybe "Reverse" means "Standard controls".
	// Let's just flip the sign of dy to be opposite of what it was.
	dy = -dy

	if g.settings.InvertY {
		dy = -dy
	}
	if dx != 0 || dy != 0 {
		g.player.SetRotation(g.player.Yaw+float32(dx)*sens, g.player.Pitch-float32(dy)*sens)
	}

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

	// === UPDATE CAMERA FROM PLAYER AFTER ALL UPDATES ===
	camera := g.engine.GetCamera()
	camera.SetPosition(g.player.Position)
	camera.SetRotation(g.player.Yaw, g.player.Pitch)

	// Update Minimap
	if g.minimap != nil && g.world != nil {
		creatures := make([]mgl32.Vec3, 0)
		for _, c := range g.world.CreatureManager.GetCreatures() {
			creatures = append(creatures, c.Position)
		}
		g.minimap.Update(g.player.Position, g.world.GetBiomeAt, g.world.GetHeight, creatures)
	}

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

	// Update atmospheric particles based on biome and time of day
	if g.sky != nil && g.world != nil {
		playerPos := g.player.Position
		biome := g.world.GetBiomeAt(int(playerPos.X()), int(playerPos.Z()))
		timeOfDay := g.sky.GetTimeOfDayNormalized()
		g.engine.UpdateAtmosphericParticles(playerPos, biome, timeOfDay)
	}

	// Update water particles (bubbles, fish)
	if g.world != nil {
		waterSurfaceY := float32(12) // Sea level
		g.engine.UpdateWaterParticles(g.player.Position, isUnderwater, waterSurfaceY)
	}

	// Demo campfire at origin (always visible as a test)
	// Base is at height, Log is at height+1, so fire is at height+2
	campfirePos := mgl32.Vec3{5, float32(g.world.GetHeight(5, 5)) + 2.0, 5}
	g.engine.EmitCampfire(campfirePos)

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
	isBreaking := input.IsMouseButtonPressed(glfw.MouseButtonLeft) && g.targetBlock != nil

	if g.blockBreaker != nil {
		if isBreaking {
			// Start or continue breaking
			targetPos := g.targetBlock.BlockPos
			targetType := g.targetBlock.BlockType
			def := block.GetDefinition(targetType)

			// If we switched blocks, the breaker handles it (resets progress)
			g.blockBreaker.StartBreaking(targetPos, def.BreakTime)

			// Update progress
			broken := g.blockBreaker.Update(dt)

			if broken {
				// Block destroyed!
				// Get block color for particles before destroying
				destroyedBlock := g.world.GetBlock(targetPos[0], targetPos[1], targetPos[2])
				blockColor := destroyedBlock.GetColor()

				// Destroy block
				g.world.SetBlock(targetPos[0], targetPos[1], targetPos[2], block.Air)

				// Stop breaking animation
				g.blockBreaker.StopBreaking()

				// Emit destruction particles
				if ps := g.engine.GetParticleSystem(); ps != nil {
					blockCenter := mgl32.Vec3{
						float32(targetPos[0]) + 0.5,
						float32(targetPos[1]) + 0.5,
						float32(targetPos[2]) + 0.5,
					}
					particleColor := mgl32.Vec4{blockColor[0], blockColor[1], blockColor[2], 1.0}
					ps.EmitExplosion(blockCenter, 12, particleColor)
				}
			}
		} else {
			// Stop breaking if button released or looking away
			g.blockBreaker.StopBreaking()
			g.blockBreaker.Update(dt) // Just to reset state if needed
		}
	} else if g.wasMouseButtonJustPressed(input, glfw.MouseButtonLeft) && g.targetBlock != nil {
		// Fallback for click-to-break if breaker not initialized
		// Get block color for particles before destroying
		destroyedBlock := g.world.GetBlock(g.targetBlock.BlockPos[0], g.targetBlock.BlockPos[1], g.targetBlock.BlockPos[2])
		blockColor := destroyedBlock.GetColor()

		// Destroy block
		g.world.SetBlock(g.targetBlock.BlockPos[0], g.targetBlock.BlockPos[1], g.targetBlock.BlockPos[2], block.Air)

		// Emit destruction particles
		if ps := g.engine.GetParticleSystem(); ps != nil {
			blockCenter := mgl32.Vec3{
				float32(g.targetBlock.BlockPos[0]) + 0.5,
				float32(g.targetBlock.BlockPos[1]) + 0.5,
				float32(g.targetBlock.BlockPos[2]) + 0.5,
			}
			particleColor := mgl32.Vec4{blockColor[0], blockColor[1], blockColor[2], 1.0}
			ps.EmitExplosion(blockCenter, 12, particleColor)
		}
	}

	if g.wasMouseButtonJustPressed(input, glfw.MouseButtonRight) && g.targetBlock != nil {
		placePos := physics.GetPlacementPosition(*g.targetBlock)
		selectedBlock := g.inventory.GetSelectedBlock()
		if selectedBlock != block.Air {
			g.world.SetBlock(placePos[0], placePos[1], placePos[2], selectedBlock)

			// Emit placement particles
			if ps := g.engine.GetParticleSystem(); ps != nil {
				blockCenter := mgl32.Vec3{
					float32(placePos[0]) + 0.5,
					float32(placePos[1]) + 0.5,
					float32(placePos[2]) + 0.5,
				}
				blockColor := selectedBlock.GetColor()
				particleColor := mgl32.Vec4{blockColor[0], blockColor[1], blockColor[2], 0.8}
				ps.EmitExplosion(blockCenter, 8, particleColor)
			}
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
	if g.world != nil {
		terrainConfig := terrain.GeneratorConfig{
			SeaLevel:         g.settings.SeaLevel,
			TerrainAmplitude: g.settings.TerrainAmplitude,
			TreeDensity:      g.settings.TreeDensity,
			CaveFrequency:    g.settings.CaveFrequency,
		}
		g.world.ApplySettings(g.settings.DayDuration, g.settings.NightBrightness, terrainConfig)
	}
}

func (g *Game) wasKeyJustPressed(input *render.Input, key glfw.Key) bool {
	current := input.IsKeyPressed(key)
	last := g.lastKeyStates[key]
	g.lastKeyStates[key] = current
	return current && !last
}

func (g *Game) wasMouseButtonJustPressed(input *render.Input, button glfw.MouseButton) bool {
	current := input.IsMouseButtonPressed(button)
	last := g.lastMouseButtonStates[button]
	g.lastMouseButtonStates[button] = current
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
	if g.uiRenderer != nil && g.mainMenu != nil {
		g.uiRenderer.BeginFrame()
		g.menuRenderer.RenderMenu(g.mainMenu, g.screenWidth, g.screenHeight)
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
		sunDir := mgl32.Vec3{0.5, 0.8, 0.3}.Normalize()
		sunIntensity := float32(1.0)
		skyColor := mgl32.Vec3{0.53, 0.81, 0.98}
		ambientColor := mgl32.Vec3{1.0, 0.95, 0.9}

		if g.sky != nil {
			sunDir = g.sky.GetSunDirection()
			// Simple day/night logic for intensity and color
			timeOfDay := g.sky.GetTimeOfDayNormalized() // 0-1

			// Day/Night cycle
			// Day: 0.25 - 0.75 (approx)
			if g.sky.IsNight() {
				sunIntensity = 0.1
				skyColor = mgl32.Vec3{0.02, 0.02, 0.05}   // Dark night sky
				ambientColor = mgl32.Vec3{0.1, 0.1, 0.15} // Dim blueish ambient
			} else {
				// Sunset/Sunrise transition could be added here
				if timeOfDay < 0.3 || timeOfDay > 0.7 {
					// Golden hour
					sunIntensity = 0.6
					skyColor = mgl32.Vec3{0.8, 0.5, 0.3}
					ambientColor = mgl32.Vec3{0.8, 0.7, 0.6}
				}
			}
		}

		g.engine.UseVoxelShaderWithTime(render.TimeOfDayData{
			SunDirection: sunDir,
			SunIntensity: sunIntensity,
			SkyColor:     skyColor,
			AmbientColor: ambientColor,
			FogColor:     skyColor,
		})
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

			if g.engine.GetCamera().ThirdPerson && g.playerModel != nil {
				// Update player model position and rotation
				// Player position is eye height, model position is feet
				feetPos := g.player.GetFeetPosition()
				g.playerModel.Position = feetPos
				// Convert yaw to radians and invert/adjust as needed keying off standard math
				// Player Yaw: 0 = +X, 90 = +Z?
				// MathGL RotateY: CCW rotation around Y
				// We need to match the camera yaw
				g.playerModel.Rotation = mgl32.DegToRad(-g.player.Yaw + 90)

				// Update held item
				g.playerModel.HeldItem = g.inventory.GetSelectedBlock()

				g.creatureRenderer.RenderCreature(g.playerModel, view, projection, sunDir)
			}
		}

		// Render block outline
		if g.blockOutline != nil && g.targetBlock != nil {
			viewProj := g.engine.GetViewProjection()
			g.blockOutline.Render(g.targetBlock.BlockPos, viewProj)
		}

		// Render block breaking animation
		if g.blockBreaker != nil {
			viewProj := g.engine.GetViewProjection()
			g.blockBreaker.Render(viewProj)
		}
	}

	// Render UI
	if g.uiRenderer != nil {
		g.uiRenderer.BeginFrame()

		// Crosshair
		g.uiRenderer.DrawCrosshair()

		// Target Info
		if g.targetBlock != nil {
			def := block.GetDefinition(g.targetBlock.BlockType)
			g.uiRenderer.DrawTargetInfo(def.Name)
		}

		// Hotbar with item counts
		hotbarSlots := g.inventory.GetHotbarSlots()
		counts := make([]int, len(hotbarSlots))
		for i, slot := range hotbarSlots {
			counts[i] = slot.Count
		}
		g.uiRenderer.DrawHotbarWithCounts(g.inventory.SelectedIndex, g.inventory.GetHotbarColors(), counts)

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

		// Draw Minimap
		if g.minimap != nil {
			g.uiRenderer.DrawMinimap(g.minimap.GetTextureID())
		}

		// Draw Time Indicator
		if g.world != nil && g.world.TimeOfDay != nil {
			g.uiRenderer.DrawTimeIndicator(g.world.TimeOfDay.GetTimeString())
		}

		// Raytracing indicator
		if g.settings.EnableRaytracing {
			g.uiRenderer.DrawRect(10, float32(g.screenHeight-40), 120, 25, [4]float32{1, 0.5, 0, 0.8})
		}

		// Controls Overlay
		if g.showControls {
			g.uiRenderer.DrawControlsOverlay(g.controlsList)
		}

		// Inventory Panel
		if g.showInventory {
			blocks := g.inventory.GetAllBlocksWithCounts()
			displayBlocks := make([]ui.BlockDisplayInfo, len(blocks))
			for i, b := range blocks {
				displayBlocks[i] = ui.BlockDisplayInfo{
					Color:      b.Color,
					Name:       b.Name,
					Count:      b.TotalCount,
					HotbarSlot: b.HotbarSlot,
				}
			}
			g.uiRenderer.DrawInventoryPanel(displayBlocks, g.inventory.SelectedIndex)
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
		g.uiRenderer.DrawRect(0, 0, float32(g.screenWidth), float32(g.screenHeight), [4]float32{0.05, 0.05, 0.1, 0.95})

		// Settings panel
		panelWidth := float32(600)
		panelHeight := float32(600)
		panelX := (float32(g.screenWidth) - panelWidth) / 2
		panelY := (float32(g.screenHeight) - panelHeight) / 2

		// Panel background
		g.uiRenderer.DrawRect(panelX, panelY, panelWidth, panelHeight, [4]float32{0.1, 0.1, 0.15, 0.9})
		// Border
		g.uiRenderer.DrawRect(panelX, panelY, panelWidth, 3, [4]float32{0.3, 0.5, 0.8, 1.0})
		g.uiRenderer.DrawRect(panelX, panelY+panelHeight-3, panelWidth, 3, [4]float32{0.3, 0.5, 0.8, 1.0})

		// Settings title bar
		titleHeight := float32(60)
		g.uiRenderer.DrawRect(panelX, panelY, panelWidth, titleHeight, [4]float32{0.2, 0.3, 0.5, 1})
		g.uiRenderer.DrawText(panelX+panelWidth/2-80, panelY+15, 3.0, "SETTINGS", [4]float32{1, 1, 1, 1})

		// Settings items
		itemY := panelY + titleHeight + 20
		itemHeight := float32(40)
		for i, item := range g.settingsMenu.Items {
			// Item background
			bgColor := [4]float32{0.15, 0.15, 0.2, 0.5}
			if i == g.settingsMenu.SelectedIndex {
				bgColor = [4]float32{0.3, 0.4, 0.6, 0.8}
			}
			g.uiRenderer.DrawRect(panelX+20, itemY, panelWidth-40, itemHeight, bgColor)

			// Selection indicator
			if i == g.settingsMenu.SelectedIndex {
				g.uiRenderer.DrawRect(panelX+20, itemY, 4, itemHeight, [4]float32{1, 0.8, 0.2, 1})
			}

			// Setting Name
			nameColor := [4]float32{0.8, 0.8, 0.8, 1}
			if i == g.settingsMenu.SelectedIndex {
				nameColor = [4]float32{1, 1, 1, 1}
			}
			g.uiRenderer.DrawText(panelX+40, itemY+10, 2.0, item.Name, nameColor)

			// Setting Value
			valueStr := ""
			switch item.Type {
			case ui.SettingBool:
				val := item.Value
				if v, ok := g.getSettingValue(item.Name).(bool); ok {
					val = v
				}
				if val == true {
					valueStr = "ON"
				} else {
					valueStr = "OFF"
				}
			case ui.SettingInt:
				if v, ok := g.getSettingValue(item.Name).(int); ok {
					valueStr = fmt.Sprintf("%d", v)
				}
			case ui.SettingFloat:
				if v, ok := g.getSettingValue(item.Name).(float32); ok {
					valueStr = fmt.Sprintf("%.2f", v)
				}
			}

			valueColor := [4]float32{0.3, 0.8, 0.3, 1} // Greenish
			if valueStr == "OFF" {
				valueColor = [4]float32{0.8, 0.3, 0.3, 1} // Reddish
			}

			// Align value to right
			valWidth := float32(len(valueStr) * 10) // Approx
			g.uiRenderer.DrawText(panelX+panelWidth-60-valWidth, itemY+10, 2.0, valueStr, valueColor)

			itemY += itemHeight + 5
		}

		// Instructions
		g.uiRenderer.DrawText(panelX+20, panelY+panelHeight-40, 1.5, "ARROWS to change, ESC to back", [4]float32{0.6, 0.6, 0.6, 1})

		g.uiRenderer.EndFrame()
	}
}

// Helper to get current value from settings struct
func (g *Game) getSettingValue(name string) interface{} {
	// This duplicates logic in SettingsMenu but is needed for display
	// Ideally SettingsMenu would hold the current value directly
	switch name {
	case "Render Distance":
		return g.settings.RenderDistance
	case "FXAA":
		return g.settings.EnableFXAA
	case "Bloom":
		return g.settings.EnableBloom
	case "Post Processing":
		return g.settings.EnablePostProcess
	case "Raytracing":
		return g.settings.EnableRaytracing
	case "Bloom Strength":
		return g.settings.BloomStrength
	case "FOV":
		return g.settings.FOV
	case "Mouse Sensitivity":
		return g.settings.MouseSensitivity
	case "Invert Y":
		return g.settings.InvertY
	case "VSync":
		return g.settings.VSync
	}
	return nil
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
	if g.blockOutline != nil {
		g.blockOutline.Cleanup()
	}
	if g.blockBreaker != nil {
		g.blockBreaker.Cleanup()
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
