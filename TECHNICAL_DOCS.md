# Technical Documentation

This document provides a deep dive into the technical implementation of the Voxel Engine. It covers the rendering pipeline, meshing algorithms, physics simulation, and shader logic.

## 🖥 Rendering Architecture

The engine uses **OpenGL 4.1 Core Profile**. The rendering pipeline is designed for high-performance voxel rendering with support for transparency, dynamic lighting, and post-processing.

### Shader Pipeline (`assets/shaders/`)

#### Vertex Shader (`voxel.vert`)

The vertex shader handles transformation and vertex manipulation effects.

- **Inputs**: Position, Normal, Color, AO, TexCoord, MaterialID, TextureLayerID.
- **Wind Simulation**: Vertices for foliage (MaterialID 1) and water (MaterialID 2) are displaced using a sine wave function based on `uTime` and world position.
  - _Optimization_: Only the top vertices (checked via `fract(pos.y) > 0.01`) are swayed to anchor the base of the mesh.
- **Fog Calculation**: Distance-based fog factor is prepared for the fragment stage.

#### Fragment Shader (`voxel.frag`)

The fragment shader implements a modified **Blinn-Phong lighting model** with several enhancements:

- **Lighting**:
  - **Directional Sun**: Diffuse lighting based on `dot(normal, sunDir)`.
  - **Ambient Occlusion (AO)**: Non-linear curve `1.0 - pow(vAO, 1.5) * 0.5` applied to vertex-calculated AO values for softer shadows in corners.
  - **Fresnel Effect**: increased reflectivity at glancing angles for Water and Ice/Glass.
- **Material Effects**:
  - **Lava**: Animated pulses and heat distortion using `sin(uTime)`.
  - **Water**: Surface animation and standard transparency (alpha 0.65).
  - **Emissive Blocks**: Ores (Diamond) and Campfire texturess have calculated emissivity to glow at night.
- **Fog**: Exponential squared fog (`fogFactor * fogFactor`) blending geometry into `uFogColor`.

### Chunk Meshing (`internal/core/chunk/mesher.go`)

The engine uses a **Face Culling** algorithm rather than Greedy Meshing, optimized with custom geometry support.

1.  **Face Culling**: The `Mesher` iterates through every block in a chunk. For each face (Top, Bottom, N/S/E/W), it checks the neighbor. If the neighbor is Air or Transparent, the face is added.
2.  **Ambient Occlusion**: Calculated per-vertex during mesh generation.
    - The mesher checks the 3 neighbors adjacent to a vertex (corner, side 1, side 2).
    - AO level (0-3) is determined by how many of these neighbors are solid.
3.  **Custom Geometry**:
    - **Cross Mesh**: Used for flowers and tall grass. Generates two intersecting quads diagonally.
    - **Grass Blades**: Procedural geometry added to the top of standard Grass blocks. The mesher generates ~5 small random quads on top of the block to simulate 3D grass blades swaying in the wind.

## ⚛️ Physics System

The physics engine (`internal/physics`) manages collision detection and improved player movement.

### Movement Controller (`movement.go`)

Implemented in `EnhancedMovement`, utilizing a state machine:

- **States**: Walking, Sprinting, Crouching, Swimming, Flying.
- **Stamina System**: Limits sprinting duration.
- **Camera Dynamics**:
  - **Head Bob**: Sine wave applied to camera Y-position while walking.
  - **Dynamic Lean**: Camera rolls slightly when strafing left/right to simulate weight transfer.

### Swimming Physics

When `IsUnderwater` is true:

- **Gravity Reduction**: Vertical velocity is damped to simulate buoyancy (`velocity.Y * 0.95`).
- **Buoyancy**: Constant upward force applied.
- **Drag**: Horizontal velocity is multiplied by `0.98` per frame to simulate water resistance.

### Raycasting (`raycast.go`)

Used for block selection:

- Implements a **DDA (Digital Differential Analyzer)** algorithm (or similar step-based ray traversal) to traverse the voxel grid efficiently.
- Returns normal vector of the hit face, allowing for precise block placement (placing a block against the face you are looking at).

## 🌍 World Management

- **Chunk Data**: Stored as a flat 1D array or 3D slice (implementation dependent, commonly 1D for cache locality in Go) of `BlockType` (uint8).
- **Storage**: Chunks are loading/unloaded dynamically based on render distance.
- **Generation**: Uses a noise cascade (likely Perlin/Simplex) to generate heightmaps, followed by biome decoration (trees, vegetation).

## 📦 Asset Management

- **Textures**: Loaded into a `sampler2DArray` (Texture Array). This allows the shader to access all block textures using a single texture unit and a `layerID`, preventing frequent texture binding switches during rendering.
- **Embedding**: Assets are embedded into the Go binary using `//go:embed`, simplifying distribution.
