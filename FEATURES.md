# Project Features

This document outlines the gameplay features and content available in the Voxel Engine.

## 🕹 Gameplay Modes

### Survival Mechanics

Currently, the game focuses on exploration and traversal mechanics:

- **Stamina System**: Sprinting consumes stamina. When stamina is depleted, the player cannot sprint until it regenerates.
- **Swimming**: Realistic buoyancy and water resistance. Gravity is reduced underwater, allowing "floating" up.
- **Oxygen**: (Planned/Upcoming) - Currently visualization for underwater state exists (fog/color).

### Creative Tools

- **Fly Mode**: Toggleable flight allows unrestricted movement through the air (Default key: `F`).
- **Block Selection**: Raycast-based interactions with a range of 5 blocks.

## 🎒 Inventory & Items

- **Hotbar**: 9 slots for quick access to blocks/items. Use `1-9` or Mouse Scroll to selection.
- **Inventory Panel**: Expandable grid (Press `I`) showing all available blocks including colorful variants.
- **Block Picking**: Middle click (if implemented) or simple logical selection via UI.
- **Items**:
  - **Building Blocks**: Dirt, Grass, Stone, Wood, Leaves, Sand, Glass, Bricks.
  - **Decorations**: Flowers (Red/Yellow), Mushrooms, Saplings.
  - **Interactive**: Campfire (animated), Water, Lava.

## 🌎 World Exploration

The world is infinite and procedurally generated.

### Biomes

1.  **Plains**: Flat, grassy areas with occasional trees and flowers.
2.  **Forest**: Denser tree coverage, rich vegetation.
3.  **Desert**: Sandy terrain, cacti, no trees.
4.  **Snow**: Snowy surface, spruce trees, ice lakes.
5.  **Mountains**: High elevation, stone cliffs, waterfalls.

### Structures

- **Caves**: Perlin-noise generated tunnel networks. Deep caves contain lava.
- **Dungeons**: Rare underground rooms made of Stone Bricks and Mossy Stone Bricks.
- **Waterfalls**: Natural water sources flowing from cliffs in mountain biomes.
- **Lakes**: Small pools of water generated on the surface.
- **Vegetation**: Validated tree types (Oak, Birch, Spruce) and Cacti.

## 🖥 User Interface (UI)

- **Main Menu**: Start New Game, Load Game, Settings.
- **HUD**:
  - **Crosshair**: Dynamic center pointer.
  - **Hotbar**: Visual rendering of selected blocks.
  - **Debug Panel (F3)**: Shows FPS, Position, Biome, and Memory usage.
  - **Minimap**: Radar-style map in the bottom-left showing terrain topography.
- **Settings**: Adjustable graphics (FOV, Render Distance, Raytracing toggle).

## ⚙️ Graphics Settings

- **Day/Night Cycle**: Dynamic sky blending.
- **Raytracing**: Toggleable real-time raytracing effects.
- **Post-Processing**: Bloom and optimized visual filters.
