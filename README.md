# Voxel Engine - Go Edition

A high-performance, procedurally generated voxel game engine written in Go from scratch, utilizing OpenGL 4.1 for rendering. This project demonstrates advanced graphics programming techniques, efficient memory management for infinite worlds, and a modular game architecture.

## 🌟 Features

### Core Engine

- **Custom OpenGL 4.1 Engine**: Built from the ground up without heavy game engine dependencies.
- **Infinite Voxel World**: Seamless chunk loading and unloading based on player position.
- **Efficient Meshing**: Greedy meshing (or similar optimization) to reduce vertex count.
- **Multi-threaded Generation**: World generation and mesh processing happen asynchronously to prevent frame drops.

### Graphics & Visuals

- **Dynamic Day/Night Cycle**: Realistic sky simulation with sun, moon, stars, and atmospheric scattering.
- **Advanced Lighting**:
  - Smooth ambient occlusion (AO).
  - Dynamic shadows.
  - Bloom and Post-processing effects.
  - **Experimental Raytracing**: Real-time raytraced effects (toggleable).
- **Water Rendering**: Reflections, refractions, and underwater visual effects (fog, distortion).
- **Particle System**:
  - Atmospheric particles (pollen, dust).
  - Block breaking debris.
  - Campfire smoke and fire.
  - Water bubbles and splashes.
- **Block Outlines**: Visual feedback for targeted blocks.

### Gameplay

- **Survival Mechanics**:
  - Player movement with physics (Walk, Sprint, Jump, Crouch).
  - Flying mode (Creative style).
  - Swimming physics with buoyancy.
- **Interaction**:
  - Block breaking with swing animations and progress visualization.
  - Block placing.
  - Range-limited selection (Raycasting).
- **Inventory System**:
  - Hotbar with slot selection (1-9).
  - Scrollable item selection.
  - Visual inventory UI.
- **Camera Modes**: First-Person and Third-Person views.

### World Generation

- **Procedural Terrain**: Noise-based generation logic.
- **Biomes**: Distinct environments with unique flora and terrain characteristics.
- **Structures**: Trees, caves, and other decorations.
- **Persistence**: Save/Load system to preserve world state and player progress.

## 🎮 Controls

| key              | Action                         |
| ---------------- | ------------------------------ |
| **W, A, S, D**   | Move Character                 |
| **Mouse**        | Look Around                    |
| **Space**        | Jump / Swim Up                 |
| **Shift**        | Sprint                         |
| **Ctrl**         | Crouch / Swim Down             |
| **Left Click**   | Break Block                    |
| **Right Click**  | Place Block                    |
| **Scroll / 1-9** | Select Item                    |
| **F**            | Toggle Fly Mode                |
| **C**            | Toggle Camera (1st/3rd Person) |
| **R**            | Toggle Raytracing              |
| **I**            | Toggle Inventory               |
| **H**            | Toggle Controls Overlay        |
| **F3**           | Toggle Debug Info              |
| **F5**           | Quick Save                     |
| **F9**           | Quick Load                     |
| **ESC / P**      | Pause Menu                     |

## 🏗 Architecture

The project is structured following standard Go patterns, keeping the codebase modular and testable.

### Directory Structure

```
├── cmd/
│   └── voxelgame/      # Entry point (main.go)
├── internal/           # Private application code
│   ├── core/           # Core data structures (Block types, Chunk implementation)
│   ├── generation/     # Procedural generation (Terrain, Decorators, Noise)
│   ├── physics/        # Physics engine (AABB, Movement, Raycasting)
│   ├── render/         # OpenGL rendering (Shaders, Textures, Meshes, Sky)
│   ├── ui/             # User Interface (Menus, HUD, Inventory)
│   ├── world/          # World state management (Chunk loading, Entities)
│   └── save/           # Serialization and file I/O for game saves
├── assets/             # Embedded resources (Shaders, Textures)
└── go.mod              # Module definition and dependencies
```

### Key Modules

- **Engine (`internal/render/engine.go`)**: Manages the GLFW window, OpenGL context, main game loop, and high-level rendering subsystems.
- **World (`internal/world/world.go`)**: The central hub for game state. It coordinates the `ChunkManager`, `TerrainGenerator`, and `CreatureManager`.
- **Chunk System (`internal/core/chunk`)**: Handles storage of voxel data. The `Mesher` converts this raw data into renderable OpenGL buffers.
- **Physics (`internal/physics`)**: Handles collision detection against the voxel grid and player movement simulation.

## 🚀 Getting Started

### Prerequisites

- **Go**: Version 1.21 or higher.
- **C Compiler**: GCC (MinGW on Windows, clang/gcc on macOS/Linux) is required for CGo (OpenGL bindings).
- **GPU**: Graphics card supporting OpenGL 4.1 core profile.

### Installation

1.  **Clone the repository**:

    ```bash
    git clone https://github.com/yourusername/voxel-sim.git
    cd voxel-sim
    ```

2.  **Navigate to the Go source directory**:

    ```bash
    cd go
    ```

3.  **Install dependencies**:
    ```bash
    go mod tidy
    ```

### Running the Game

To run the game directly from source:

```bash
go run cmd/voxelgame/main.go
```

### Building

To create a standalone executable:

```bash
go build -o voxel-game cmd/voxelgame/main.go
```
