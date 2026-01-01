# Systems Architecture

This document details the major subsystems of the Voxel Engine, focusing on logic, data management, and architectural flow.

## 🗺 Generation System (`internal/generation`)

The world generation pipeline executes in **6 passes** for each chunk:

1.  **Base Geometry**: Uses Simplex Noise (FBM) to determine heightmap. Fills column with Bedrock -> Stone/Ore -> Subsurface (Dirt/Sand) -> Surface (Grass/Snow).
2.  **Structures**: Places trees (Oak, Birch, Spruce) and Cacti based on Biome probability.
3.  **Decorations**: Adds grass blades, flowers, and mushrooms to the chunks surface.
4.  **Water Features**: procedural Waterfalls (only in Mountains) and Lakes.
5.  **Dungeons**: Rare underground rooms (Stone Bricks) generated in air pockets (caves).
6.  **Campfires**: Rare surface structures.

**Biome Logic**: Determined by 2D noise maps for **Temperature** and **Humidity**.

- _Hot + Dry_ = Desert
- _Cold_ = Snow
- _Wet + Moderate_ = Forest
- _Dry + Moderate_ = Plains/Mountains

## 💾 Save System (`internal/save`)

Data is persisted using JSON serialization to `~/.voxelgame/saves/`.

- **World Data**:
  - `Seed`: The specific seed used for generation.
  - `ModifiedChunks`: A map storing _only_ the blocks that have changed from the procedural baseline. This keeps save files small.
- **Player Data**: Position (X, Y, Z) and Rotation (Yaw, Pitch).
- **Format**: Human-readable JSON allows for easy debugging and hacking.

## 🎨 UI System (`internal/ui`)

The UI is rendered using a dedicated orthographic shader (`UIShader`), separate from the 3D world.

- **Font Rendering**: Custom bitmap font renderer. Supports text scaling and color.
- **Components**:
  - **Renderer**: Primitive drawing (Quads/Rects) and Text.
  - **3D Preview**: A specialized function `Render3DItemInBox` renders a rotating 3D mesh of a block _inside_ the 2D UI context for inventory slots.
- **Input Layer**:
  - **Game State Manager**: Handles transitions between `Playing`, `Paused`, `Menu`, and `Settings` states.
  - **Event Propagation**: Input events are routed to the active state controller.

## 🕹 Input System

- **GLFW Callbacks**: Wrapped in an `Input` struct that tracks key states and mouse deltas.
- **Polling**: State is polled every frame in the Game Loop.
- **Context**:
  - **Gameplay**: Mouse is captured (hidden).
  - **UI/Menu**: Mouse is released (visible) for interaction.

## 🏎 Game Loop (`internal/render/engine.go`)

The engine uses a fixed time-step or variable delta-time loop:

1.  **Poll Events**: Mouse/Keyboard.
2.  **Update (Logic)**: Physics, World Loading, Entity AI. Cap `dt` at 0.1s to prevent physics explosions during lag.
3.  **Render (Draw)**:
    - 3D World (Sky -> Terrain -> Water -> Particles).
    - User Interface (Ortho overlay).
4.  **Swap Buffers**.
