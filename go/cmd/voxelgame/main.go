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
	sky              *render.Sky
	postProcess      *render.PostProcess
	creatureRenderer *render.CreatureRenderer

	// UI
	uiRenderer *ui.Renderer
	inventory  *ui.Inventory

	// Save system
	saveManager *save.Manager

	// Game state
	paused    bool
	showDebug bool

	// Stats
	fps         int
	frameCount  int
	lastFPSTime time.Time

	// Block interaction
	targetBlock *physics.RaycastResult
}

func main() {
	fmt.Println("=================================")
	fmt.Println("  Voxel Engine - Go Edition")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("Controls:")
	fmt.Println("  WASD     - Move")
	fmt.Println("  Mouse    - Look around")
	fmt.Println("  Shift    - Sprint")
	fmt.Println("  Space    - Jump")
	fmt.Println("  F        - Toggle fly mode")
	fmt.Println("  1-9      - Select hotbar slot")
	fmt.Println("  Scroll   - Cycle hotbar")
	fmt.Println("  LMB      - Break block")
	fmt.Println("  RMB      - Place block")
	fmt.Println("  F3       - Toggle debug")
	fmt.Println("  F5       - Quick save")
	fmt.Println("  F9       - Quick load")
	fmt.Println("  ESC      - Exit")
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
		showDebug:   true,
		lastFPSTime: time.Now(),
	}

	// Create rendering engine
	config := render.DefaultConfig()
	config.Title = "Voxel Engine - Go Edition"
	config.Width = 1280
	config.Height = 720

	engine, err := render.NewEngine(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create engine: %w", err)
	}
	g.engine = engine

	// Load shaders
	if err := engine.LoadShaders(); err != nil {
		return nil, fmt.Errorf("failed to load shaders: %w", err)
	}

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

	// Create UI renderer
	uiRenderer, err := ui.NewRenderer(config.Width, config.Height)
	if err != nil {
		fmt.Printf("Warning: Failed to create UI renderer: %v\n", err)
	}
	g.uiRenderer = uiRenderer

	// Create inventory
	g.inventory = ui.NewInventory()

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

	// Create creature renderer
	creatureRenderer, err := render.NewCreatureRenderer()
	if err != nil {
		fmt.Printf("Warning: Failed to create creature renderer: %v\n", err)
	}
	g.creatureRenderer = creatureRenderer

	return g, nil
}

// Run starts the game loop
func (g *Game) Run() {
	g.engine.Run(g.Update, g.Render)
}

// Update updates the game state
func (g *Game) Update(dt float32) {
	input := g.engine.GetInput()

	// Handle special keys
	g.handleInput(input, dt)

	if g.paused {
		return
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

	sprint := input.IsKeyPressed(glfw.KeyLeftShift)
	jump := input.IsKeyPressed(glfw.KeySpace)

	// Handle mouse look
	dx, dy := input.GetMouseDelta()
	if dx != 0 || dy != 0 {
		g.player.SetRotation(g.player.Yaw+float32(dx)*0.1, g.player.Pitch+float32(dy)*0.1)
	}

	// Update player physics
	g.player.SetMovement(forward, right, sprint, jump)
	g.player.Update(dt)

	// Update world around player
	g.world.Update(
		float64(g.player.Position.X()),
		float64(g.player.Position.Y()),
		float64(g.player.Position.Z()),
	)

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
		// Break block
		g.world.SetBlock(g.targetBlock.BlockPos[0], g.targetBlock.BlockPos[1], g.targetBlock.BlockPos[2], block.Air)
	}
	if input.IsMouseButtonPressed(glfw.MouseButtonRight) && g.targetBlock != nil {
		// Place block
		placePos := physics.GetPlacementPosition(*g.targetBlock)
		selectedBlock := g.inventory.GetSelectedBlock()
		if selectedBlock != block.Air {
			g.world.SetBlock(placePos[0], placePos[1], placePos[2], selectedBlock)
		}
	}

	// FPS counter
	g.frameCount++
	if time.Since(g.lastFPSTime) >= time.Second {
		g.fps = g.frameCount
		g.frameCount = 0
		g.lastFPSTime = time.Now()
	}

	// Update sky
	if g.sky != nil {
		g.sky.Update(dt)
	}
}

// handleInput handles special key inputs
func (g *Game) handleInput(input *render.Input, dt float32) {
	// Number keys for hotbar selection
	for i := 0; i < 9; i++ {
		if input.IsKeyPressed(glfw.Key(int(glfw.Key1) + i)) {
			g.inventory.SelectSlot(i)
		}
	}

	// Scroll for hotbar cycling
	_, scrollY := input.GetScroll()
	if scrollY != 0 {
		g.inventory.ScrollSelection(int(-scrollY))
	}

	// Toggle fly mode (F)
	if input.IsKeyPressed(glfw.KeyF) {
		g.player.ToggleFlyMode()
		if g.player.IsFlying {
			fmt.Println("Fly mode enabled")
		} else {
			fmt.Println("Fly mode disabled")
		}
	}

	// Toggle debug (F3)
	if input.IsKeyPressed(glfw.KeyF3) {
		g.showDebug = !g.showDebug
	}

	// Quick save (F5)
	if input.IsKeyPressed(glfw.KeyF5) {
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

	// Quick load (F9)
	if input.IsKeyPressed(glfw.KeyF9) {
		data, err := g.saveManager.QuickLoad()
		if err != nil {
			fmt.Printf("Failed to load: %v\n", err)
		} else {
			g.player.Position = mgl32.Vec3{data.Player.PositionX, data.Player.PositionY, data.Player.PositionZ}
			g.player.Yaw = data.Player.Yaw
			g.player.Pitch = data.Player.Pitch
			fmt.Println("Game loaded!")
		}
	}
}

// Render renders the game
func (g *Game) Render() {
	// Render sky first (background)
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
			mgl32.DegToRad(camera.FOV),
			float32(1280)/float32(720),
			0.1, 1000.0,
		)
		sunDir := mgl32.Vec3{0.5, 0.8, 0.3}.Normalize()
		if g.sky != nil {
			sunDir = g.sky.GetSunDirection()
		}
		g.creatureRenderer.RenderCreatures(g.world.GetCreatures(), view, projection, sunDir)
	}

	// Render UI
	if g.uiRenderer != nil {
		g.uiRenderer.BeginFrame()

		// Crosshair
		g.uiRenderer.DrawCrosshair()

		// Hotbar
		g.uiRenderer.DrawHotbar(g.inventory.SelectedIndex, g.inventory.GetHotbarColors())

		// Debug panel
		if g.showDebug {
			stats := g.world.GetStats()
			g.uiRenderer.DrawDebugPanel(ui.DebugInfo{
				Position:     g.player.Position,
				ChunksLoaded: stats.ChunksLoaded,
				FPS:          g.fps,
				Biome:        g.world.GetBiomeAt(int(g.player.Position.X()), int(g.player.Position.Z())),
			})
		}

		g.uiRenderer.EndFrame()
	}
}

// Cleanup releases resources
func (g *Game) Cleanup() {
	fmt.Println("Cleaning up...")

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
