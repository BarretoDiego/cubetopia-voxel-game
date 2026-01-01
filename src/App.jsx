/**
 * Componente principal do mundo voxel - integra todos os sistemas
 */

import React, { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import { Canvas, useThree, useFrame } from '@react-three/fiber';
import { PointerLockControls, Sky, Stars } from '@react-three/drei';
import * as THREE from 'three';

// Core
import { CHUNK_SIZE, CHUNK_HEIGHT, RENDER_DISTANCE } from './utils/constants.js';
import { BlockTypes, blockRegistry } from './core/blocks/index.js';
import { Chunk, ChunkManager, ChunkMesher } from './core/chunks/index.js';

// Generation
import { TerrainGenerator } from './generation/terrain/index.js';
import { CreatureGenerator, CreatureTemplates } from './generation/entities/index.js';

// Rendering
import { createBlockTextures } from './rendering/textures/index.js';
import { 
  createSlimeModel, 
  createPigModel, 
  createZombieModel, 
  createSpiderModel, 
  createBirdModel 
} from './rendering/voxelModels/EntityModels.js';

// Controls
import { usePlayerControls } from './controls/usePlayerControls.js';
import { useBlockInteraction } from './controls/useBlockInteraction.js';

// Post-processing
import { EffectComposer, Bloom, SSAO, Vignette } from '@react-three/postprocessing';
import { BlendFunction } from 'postprocessing';

// UI
import { 
  HUD, 
  Crosshair, 
  DebugInfo, 
  Hotbar, 
  Instructions, 
  Loading, 
  UnderwaterOverlay, 
  Gallery, 
  InventoryUI,
  Minimap,
  EnhancedHotbar,
  MainMenu,
  PauseMenu,
  GameCursor,
  PreviewCanvas // [NEW] Use shared preview canvas
} from './ui/index.js';

// Systems
import { MemoryManager } from './systems/MemoryManager.js';
import { WaterSimulator } from './systems/WaterSimulator.js';
import { createSaveData, saveToFile, loadFromFile, applySaveData } from './systems/SaveManager.js';

// Items
import { Inventory } from './core/items/index.js';

// Player Model
import { PlayerModel } from './rendering/PlayerModel.jsx';
import { EnvironmentParticles } from './effects/EnvironmentParticles';

// ============================================================================
// Effects Component - Optimized for performance
// ============================================================================


function Effects({ disabled }) {
  if (disabled) return null;

  return (
    <EffectComposer multisampling={0}>
      {/* SSAO - Light settings for performance */}
      <SSAO
        blendFunction={BlendFunction.MULTIPLY}
        samples={8}
        radius={2}
        intensity={8}
        luminanceInfluence={0.6}
        color="#000000"
        distanceScaling={true}
        depthAwareUpsampling={false}
      />
      {/* Subtle bloom for glow */}
      <Bloom
        intensity={0.3}
        luminanceThreshold={0.8}
        luminanceSmoothing={0.9}
        mipmapBlur={true}
      />
      {/* Vignette for depth */}
      <Vignette
        blendFunction={BlendFunction.NORMAL}
        offset={0.3}
        darkness={0.5}
      />
    </EffectComposer>
  );
}

// ============================================================================
// Auto-Lock Pointer Controls
// ============================================================================

function GameControls({ enabled }) {
  const controlsRef = useRef();
  const hasLocked = useRef(false);
  
  useEffect(() => {
    if (enabled && controlsRef.current && !hasLocked.current) {
      // Auto-lock pointer on first enable
      const timer = setTimeout(() => {
        if (controlsRef.current) {
          controlsRef.current.lock();
          hasLocked.current = true;
        }
      }, 100);
      return () => clearTimeout(timer);
    }
  }, [enabled]);
  
  if (!enabled) return null;
  
  return <PointerLockControls ref={controlsRef} />;
}

function World({ world, onPositionChange, onRotationChange, selectedBlock, onWorldChange, initialPosition, onScrollHotbar }) {
  const { camera } = useThree();
  
  // Player controls - pass initial spawn position
  usePlayerControls(world, onPositionChange, initialPosition);
  
  // Block interaction
  const { breakBlock, placeBlock } = useBlockInteraction(
    world, 
    selectedBlock, 
    onWorldChange
  );

  // Mouse handlers - Left click = break, Right click = place
  useEffect(() => {
    const handleMouseDown = (e) => {
      if (e.button === 0) {
        // Left click = break block
        breakBlock();
      } else if (e.button === 2) {
        // Right click = place block
        placeBlock();
      }
    };

    const handleContextMenu = (e) => {
      e.preventDefault(); // Prevent context menu on right click
    };

    const handleWheel = (e) => {
      if (onScrollHotbar) {
        // Scroll up = previous, scroll down = next
        onScrollHotbar(e.deltaY > 0 ? 1 : -1);
      }
    };

    window.addEventListener('mousedown', handleMouseDown);
    window.addEventListener('contextmenu', handleContextMenu);
    window.addEventListener('wheel', handleWheel);

    return () => {
      window.removeEventListener('mousedown', handleMouseDown);
      window.removeEventListener('contextmenu', handleContextMenu);
      window.removeEventListener('wheel', handleWheel);
    };
  }, [breakBlock, placeBlock, onScrollHotbar]);

  // Keyboard handlers for block interaction - Q = break, E = place (backup)
  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.code === 'KeyQ') {
        breakBlock();
      } else if (e.code === 'KeyE') {
        placeBlock();
      }
    };

    window.addEventListener('keydown', handleKeyDown);

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [breakBlock, placeBlock]);

  // Track camera rotation for minimap
  useFrame(() => {
    if (onRotationChange && camera) {
      const rotation = camera.rotation.y;
      onRotationChange(rotation);
    }
  });

  return null;
}

// ============================================================================
// ============================================================================
// Chunk Renderer - Optimized with deferred meshing and strict cleanup
// ============================================================================

import { sharedChunkMesher } from './core/chunks/ChunkMesher.js';

function ChunkRenderer({ chunk, world, updateTrigger, sharedMaterial }) {
  const meshRef = useRef();
  const geometryRef = useRef(null);
  // Remove per-instance mesher, use shared singleton
  
  const hasGenerated = useRef(false);

  useEffect(() => {
    if (!chunk || !meshRef.current) return;
    
    // Check if chunk already has valid geometry we can reuse
    if (chunk.mesh && chunk.mesh.geometry && !chunk.isDirty) {
      if (geometryRef.current !== chunk.mesh.geometry) {
        geometryRef.current = chunk.mesh.geometry;
        meshRef.current.geometry = chunk.mesh.geometry;
        
        if (sharedMaterial) {
          meshRef.current.material = sharedMaterial;
        }
      }
      return;
    }
    
    // Generate mesh synchronously using shared mesher
    const geometry = sharedChunkMesher.generateMesh(chunk, (x, y, z) => world.getBlock(x, y, z));
    
    if (geometry && meshRef.current) {
      // CRITICAL: If chunk already had geometry, dispose it before overwriting!
      if (chunk.mesh && chunk.mesh.geometry) {
         chunk.mesh.geometry.dispose();
      }
      
      geometryRef.current = geometry;
      meshRef.current.geometry = geometry;
      
      if (sharedMaterial) {
        meshRef.current.material = sharedMaterial;
      }
      
      // Link mesh to chunk for explicit disposal by ChunkManager
      // We create a dummy mesh object if needed just to hold the geometry ref for ChunkManager
      if (!chunk.mesh) {
        chunk.mesh = { geometry: geometry, material: null };
      } else {
        chunk.mesh.geometry = geometry;
      }
      
      hasGenerated.current = true;
    }
    
    // Reset dirty flag after rendering
    if (chunk) {
      chunk.isDirty = false;
    }
    
    // Cleanup on unmount
    return () => {
      // ChunkManager handles main disposal, but we clear local refs
      if (geometryRef.current) {
        // geometryRef.current.dispose(); // Let ChunkManager handle disposal to avoid double-free
        geometryRef.current = null;
      }
      hasGenerated.current = false;
    };
  }, [chunk, chunk?.isDirty, world, updateTrigger, sharedMaterial]);

  return (
    <mesh ref={meshRef}>
      {/* Material is assigned via sharedMaterial, fallback if not provided */}
      {!sharedMaterial && (
        <meshStandardMaterial 
          vertexColors 
          side={THREE.DoubleSide}
        />
      )}
    </mesh>
  );
}

// ============================================================================
// Chunks Manager Component - with strict memory limits
// ============================================================================


function ChunksDisplay({ world, playerPosition, updateTrigger }) {
  const [visibleChunks, setVisibleChunks] = useState([]);
  const lastUpdateRef = useRef({ x: null, z: null });

  // Shared material to reduce draw calls and memory overhead
  const sharedMaterial = useMemo(() => new THREE.MeshStandardMaterial({
    vertexColors: true,
    side: THREE.DoubleSide, // Reverted to DoubleSide to fix missing faces
    roughness: 0.8,
    metalness: 0.1,
  }), []);

  useEffect(() => {
    // ... (logic remains same)
    if (!world || !playerPosition) return;

    const playerCx = Math.floor(playerPosition.x / CHUNK_SIZE);
    const playerCz = Math.floor(playerPosition.z / CHUNK_SIZE);

    if (lastUpdateRef.current.x === playerCx && lastUpdateRef.current.z === playerCz) {
      return;
    }
    lastUpdateRef.current = { x: playerCx, z: playerCz };

    const chunks = [];
    
    // Load chunks with throttle to prevent memory spike
    let loadedCount = 0;
    const MAX_LOADS_PER_FRAME = 2; // Throttle loading speed
    
    const loadNextBatch = async () => {
      const loadPromises = [];
      
      for (let dx = -RENDER_DISTANCE; dx <= RENDER_DISTANCE; dx++) {
        for (let dz = -RENDER_DISTANCE; dz <= RENDER_DISTANCE; dz++) {
          const cx = playerCx + dx;
          const cz = playerCz + dz;
          
          loadPromises.push(world.chunkManager.loadChunk(cx, cz));
        }
      }

      // Wait for all current visible chunks to load
      const loadedChunks = await Promise.all(loadPromises);
      
      if (lastUpdateRef.current.x !== playerCx || lastUpdateRef.current.z !== playerCz) {
        return; // Player moved, abort
      }

      // Filter valid chunks
      const validChunks = loadedChunks.filter(c => c !== null);
      setVisibleChunks(validChunks);
      
      // Cleanup distant chunks
      const unloadDistance = RENDER_DISTANCE + 1;
      for (const [id, chunk] of world.chunkManager.chunks) {
        const dx = Math.abs(chunk.cx - playerCx);
        const dz = Math.abs(chunk.cz - playerCz);
        if (dx > unloadDistance || dz > unloadDistance) {
          world.chunkManager.unloadChunk(chunk.cx, chunk.cz);
        }
      }
    };

    loadNextBatch();

  }, [world, playerPosition?.x, playerPosition?.z, updateTrigger]);

  return (
    <group>
      {visibleChunks.map(chunk => (
        <ChunkRenderer 
          key={chunk.id} 
          chunk={chunk} 
          world={world}
          updateTrigger={updateTrigger}
          sharedMaterial={sharedMaterial}
        />
      ))}
    </group>
  );
}

// ============================================================================
// Creature Component - High Resolution Voxel Models
// ============================================================================

function Creature({ creature, world }) {
  const meshRef = useRef();
  const creatureRef = useRef(creature);
  const geometryRef = useRef(null);

  // Create geometry based on creature type
  const geometry = useMemo(() => {
    const size = creature.size || 1;
    const color = creature.colors?.primary || '#22CC22';
    
    switch (creature.template) {
      case CreatureTemplates.SLIME:
        return createSlimeModel(color, size * 0.5);
      case CreatureTemplates.QUADRUPED:
        // Use pig model for quadrupeds
        return createPigModel(size * 0.4);
      case CreatureTemplates.BIPED:
        // Use zombie model for bipeds
        return createZombieModel(size * 0.3);
      case CreatureTemplates.SPIDER:
        return createSpiderModel(size * 0.4);
      case CreatureTemplates.FLYING:
        return createBirdModel(color, size * 0.5);
      default:
        return createSlimeModel(color, size * 0.5);
    }
  }, [creature.template, creature.size, creature.colors?.primary]);

  useFrame((state, delta) => {
    if (!meshRef.current || !creatureRef.current) return;

    const c = creatureRef.current;
    
    // Atualiza posição
    c.position.x += c.velocity.x * delta;
    c.position.z += c.velocity.z * delta;
    
    // Gravidade
    c.velocity.y -= 20 * delta;
    c.position.y += c.velocity.y * delta;
    
    // Chão
    const groundY = world ? world.getHeight(Math.floor(c.position.x), Math.floor(c.position.z)) + 1 : 0;
    if (c.position.y < groundY) {
      c.position.y = groundY;
      c.velocity.y = 0;
    }
    
    // Timer para comportamento
    c.timer += delta;
    if (c.timer > 3) {
      c.velocity.x = (Math.random() - 0.5) * 2;
      c.velocity.z = (Math.random() - 0.5) * 2;
      c.timer = 0;
      
      // Slimes pulam
      if (c.template === CreatureTemplates.SLIME) {
        c.velocity.y = 6;
      }
    }

    // Atualiza mesh
    meshRef.current.position.set(c.position.x, c.position.y, c.position.z);
    meshRef.current.rotation.y = Math.atan2(c.velocity.x, c.velocity.z);
    
    // Animação de slime
    if (c.template === CreatureTemplates.SLIME) {
      const scale = 1 + Math.sin(state.clock.elapsedTime * 5) * 0.1;
      meshRef.current.scale.set(scale, 1/scale, scale);
    }
  });

  // Cleanup geometry on unmount
  useEffect(() => {
    return () => {
      if (geometryRef.current) {
        geometryRef.current.dispose();
      }
    };
  }, []);

  if (!geometry) return null;

  return (
    <mesh 
      ref={meshRef} 
      geometry={geometry}
      position={[creature.position.x, creature.position.y, creature.position.z]}
    >
      <meshStandardMaterial 
        vertexColors 
        transparent={creature.template === CreatureTemplates.SLIME}
        opacity={creature.template === CreatureTemplates.SLIME ? 0.7 : 1}
      />
    </mesh>
  );
}

// ============================================================================
// VoxelWorld Class
// ============================================================================

class VoxelWorld {
  constructor(seed = Date.now()) {
    this.seed = seed;
    this.terrainGenerator = new TerrainGenerator(seed);
    this.chunkManager = new ChunkManager(this.terrainGenerator);
    this.creatureGenerator = new CreatureGenerator(seed);
    this.creatures = [];
  }

  getBlock(x, y, z) {
    return this.chunkManager.getBlock(x, y, z);
  }

  setBlock(x, y, z, type) {
    return this.chunkManager.setBlock(x, y, z, type);
  }

  getHeight(x, z) {
    return this.chunkManager.getHeight(x, z);
  }

  spawnCreature(x, z, template, biome) {
    const y = this.getHeight(Math.floor(x), Math.floor(z)) + 1;
    const creature = this.creatureGenerator.create({
      template: template || this.creatureGenerator.rng.choose(Object.values(CreatureTemplates)),
      biome: biome || 'plains',
      position: { x, y, z }
    });
    this.creatures.push(creature);
    return creature;
  }
}

// ============================================================================
// Main App Component
// ============================================================================

function App() {
  const [isLoading, setIsLoading] = useState(true);
  const [showInstructions, setShowInstructions] = useState(true);
  const [playerPosition, setPlayerPosition] = useState({ x: 0, y: 40, z: 0 });
  const [playerRotation, setPlayerRotation] = useState(0);
  const [selectedBlockIndex, setSelectedBlockIndex] = useState(0);
  const [updateTrigger, setUpdateTrigger] = useState(0);
  const [fps, setFps] = useState(60);
  const [isUnderwater, setIsUnderwater] = useState(false);
  const [showGallery, setShowGallery] = useState(false);
  const [showInventory, setShowInventory] = useState(false);
  const [memoryUsage, setMemoryUsage] = useState(null);
  const [memoryWarning, setMemoryWarning] = useState(false);
  const [memoryCritical, setMemoryCritical] = useState(false);
  const [useEnhancedHotbar, setUseEnhancedHotbar] = useState(false); // Disabled for memory
  const [saveNotification, setSaveNotification] = useState(null); // { type: 'save'|'load', message: '' }
  const [isThirdPerson, setIsThirdPerson] = useState(false); // V key toggles
  const [isMoving, setIsMoving] = useState(false); // For walk animation
  const [gameState, setGameState] = useState('menu'); // 'menu', 'playing', 'paused'
  
  const worldRef = useRef(null);
  const fpsRef = useRef({ frames: 0, lastTime: Date.now() });
  const memoryManagerRef = useRef(null);
  const waterSimulatorRef = useRef(null);
  const inventoryRef = useRef(null);
  
  // Blocos disponíveis para construção (mais variedade)
  const availableBlocks = useMemo(() => [
    { id: BlockTypes.GRASS, name: 'Grama', color: '#567d46' },
    { id: BlockTypes.DIRT, name: 'Terra', color: '#8b6914' },
    { id: BlockTypes.STONE, name: 'Pedra', color: '#7a7a7a' },
    { id: BlockTypes.WOOD, name: 'Madeira', color: '#8b5a2b' },
    { id: BlockTypes.WATER, name: 'Água', color: '#3498db' },
    { id: BlockTypes.SAND, name: 'Areia', color: '#e0c090' },
    { id: BlockTypes.COBBLESTONE, name: 'Pedregulho', color: '#5a5a5a' },
    { id: BlockTypes.BRICK, name: 'Tijolo', color: '#b75a3c' },
    { id: BlockTypes.GLASS, name: 'Vidro', color: '#c8dbe0' }
  ], []);

  // Inicializa mundo e sistemas
  useEffect(() => {
    const world = new VoxelWorld(12345);
    worldRef.current = world;
    
    // Inicializa Water Simulator
    waterSimulatorRef.current = new WaterSimulator(world);
    
    // Inicializa Inventory with random starting items
    inventoryRef.current = new Inventory(36);
    
    // Add random starting blocks to inventory
    const startingItems = [
      { id: BlockTypes.DIRT, count: 20 + Math.floor(Math.random() * 30) },
      { id: BlockTypes.STONE, count: 15 + Math.floor(Math.random() * 25) },
      { id: BlockTypes.WOOD, count: 10 + Math.floor(Math.random() * 20) },
      { id: BlockTypes.GRASS, count: 10 + Math.floor(Math.random() * 15) },
      { id: BlockTypes.SAND, count: 5 + Math.floor(Math.random() * 15) },
      { id: BlockTypes.BRICK, count: 5 + Math.floor(Math.random() * 10) },
      { id: BlockTypes.GLASS, count: 5 + Math.floor(Math.random() * 10) },
    ];
    startingItems.forEach((item, index) => {
      if (inventoryRef.current.setSlot) {
        inventoryRef.current.setSlot(index, { type: item.id, count: item.count });
      }
    });
    
    // Pré-gera chunks ao redor do spawn para calcular altura correta
    const spawnChunk = new Chunk(0, 0);
    world.terrainGenerator.generateChunk(spawnChunk);
    world.chunkManager.chunks.set(spawnChunk.id, spawnChunk);
    
    // Calcula altura correta do spawn baseado no terreno
    const spawnHeight = spawnChunk.getHeight(8, 8) + 3; // Centro do chunk + margem
    setPlayerPosition({ x: 8, y: spawnHeight, z: 8 });
    
    // Spawn varied creatures - different types
    setTimeout(() => {
      const creatureTypes = [
        CreatureTemplates.SLIME,
        CreatureTemplates.QUADRUPED, // Pig/Animal
        CreatureTemplates.FLYING,    // Bird
        CreatureTemplates.QUADRUPED, // Cow
        CreatureTemplates.SPIDER,
      ];
      
      for (let i = 0; i < 5; i++) {
        const x = (Math.random() - 0.5) * 40;
        const z = (Math.random() - 0.5) * 40;
        const biomes = ['plains', 'forest', 'desert'];
        const biome = biomes[Math.floor(Math.random() * biomes.length)];
        world.spawnCreature(x, z, creatureTypes[i % creatureTypes.length], biome);
      }
    }, 2000);
    
    setIsLoading(false);
  }, []);

  // Inicializa Memory Manager
  useEffect(() => {
    const memoryManager = new MemoryManager({
      onUpdate: (usage) => {
        setMemoryUsage(usage);
      },
      onWarning: (usage) => {
        setMemoryWarning(true);
        console.warn('[App] Memory warning:', usage);
      },
      onCritical: (usage) => {
        setMemoryCritical(true);
        console.error('[App] Memory critical - stopping game:', usage);
      }
    });
    
    memoryManagerRef.current = memoryManager;
    memoryManager.start();
    
    return () => {
      memoryManager.stop();
    };
  }, []);

  // Water simulation tick
  useEffect(() => {
    if (!waterSimulatorRef.current) return;
    
    const tickInterval = setInterval(() => {
      waterSimulatorRef.current.tick(20);
    }, 200);
    
    return () => clearInterval(tickInterval);
  }, []);

  // Atualiza FPS
  useEffect(() => {
    const interval = setInterval(() => {
      const now = Date.now();
      const elapsed = (now - fpsRef.current.lastTime) / 1000;
      setFps(Math.round(fpsRef.current.frames / elapsed));
      fpsRef.current.frames = 0;
      fpsRef.current.lastTime = now;
    }, 1000);
    return () => clearInterval(interval);
  }, []);

  // Key handlers para seleção de bloco, galeria e inventário
  useEffect(() => {
    const handleKeyDown = async (e) => {
      // ESC = Toggle Pause / Menu
      if (e.code === 'Escape') {
        e.preventDefault();
        if (gameState === 'playing') {
          setGameState('paused');
          setShowInventory(false);
          setShowGallery(false);
        } else if (gameState === 'paused') {
          setGameState('playing');
        }
        return;
      }

      // Ignore keys if not playing
      if (gameState !== 'playing') return;

      // Ignore if inventory is open (let InventoryUI handle keys)
      if (showInventory) return;
      
      const num = parseInt(e.key);
      if (num >= 1 && num <= 9 && num <= availableBlocks.length) {
        setSelectedBlockIndex(num - 1);
      }
      // Tab key toggles gallery
      if (e.code === 'Tab') {
        e.preventDefault();
        setShowGallery(prev => !prev);
      }
      // I key toggles inventory
      if (e.code === 'KeyI') {
        e.preventDefault();
        setShowInventory(prev => !prev);
      }
      
      // F5 = Quick Save
      if (e.code === 'F5') {
        e.preventDefault();
        if (worldRef.current) {
          setSaveNotification({ type: 'save', message: 'Salvando...' });
          try {
            const saveData = createSaveData(
              worldRef.current,
              playerPosition,
              playerRotation,
              inventoryRef.current
            );
            await saveToFile(saveData, 'quicksave');
            setSaveNotification({ type: 'save', message: '✓ Jogo salvo!' });
            setTimeout(() => setSaveNotification(null), 2000);
          } catch (error) {
            console.error('Save failed:', error);
            setSaveNotification({ type: 'error', message: '✗ Falha ao salvar' });
            setTimeout(() => setSaveNotification(null), 3000);
          }
        }
      }
      
      // F9 = Quick Load
      if (e.code === 'F9') {
        e.preventDefault();
        setSaveNotification({ type: 'load', message: 'Carregando...' });
        try {
          const saveData = await loadFromFile('quicksave');
          if (saveData && worldRef.current) {
            applySaveData(saveData, worldRef.current, setPlayerPosition, inventoryRef.current);
            setUpdateTrigger(t => t + 1);
            setSaveNotification({ type: 'load', message: '✓ Jogo carregado!' });
            setTimeout(() => setSaveNotification(null), 2000);
          } else {
            setSaveNotification({ type: 'error', message: 'Nenhum save encontrado' });
            setTimeout(() => setSaveNotification(null), 3000);
          }
        } catch (error) {
          console.error('Load failed:', error);
          setSaveNotification({ type: 'error', message: '✗ Falha ao carregar' });
          setTimeout(() => setSaveNotification(null), 3000);
        }
      }
      
      // V = Toggle camera mode (1st/3rd person)
      if (e.code === 'KeyV') {
        e.preventDefault();
        setIsThirdPerson(prev => !prev);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [availableBlocks.length, showInventory, playerPosition, playerRotation, gameState]);

  // Check if player is underwater
  useEffect(() => {
    if (!worldRef.current || !playerPosition) return;
    
    const checkInterval = setInterval(() => {
      const block = worldRef.current.getBlock(
        Math.floor(playerPosition.x),
        Math.floor(playerPosition.y + 1.5), // Head level
        Math.floor(playerPosition.z)
      );
      setIsUnderwater(block === BlockTypes.WATER);
    }, 200);
    
    return () => clearInterval(checkInterval);
  }, [playerPosition]);

  const handleStart = useCallback(() => {
    setShowInstructions(false);
  }, []);

  const handleWorldChange = useCallback(() => {
    setUpdateTrigger(t => t + 1);
  }, []);

  if (isLoading) {
    return <Loading message="Gerando terreno..." />;
  }

  return (
    <div style={{ width: '100%', height: '100vh', background: '#000' }}>
      {/* Shared Canvas for UI 3D Elements */}
      <PreviewCanvas />

      <Canvas
        camera={{ fov: 75, near: 0.1, far: 500 }}
        onCreated={({ gl }) => {
          gl.setClearColor('#87CEEB');
        }}
      >
        {/* Céu */}
        <Sky 
          sunPosition={[100, 50, 100]} 
          turbidity={0.3}
          rayleigh={0.5}
          mieCoefficient={0.005}
          mieDirectionalG={0.8}
        />
        <Stars 
          radius={300} 
          depth={60} 
          count={2000} 
          factor={4} 
          saturation={0} 
          fade 
          speed={0.5} 
        />
        
        {/* Iluminação Realista */}
        <ambientLight intensity={0.2} />
        <directionalLight 
          position={[50, 100, 50]} 
          intensity={1.5} 
          castShadow
          shadow-mapSize-width={2048}
          shadow-mapSize-height={2048}
          shadow-camera-far={200}
          shadow-camera-left={-100}
          shadow-camera-right={100}
          shadow-camera-top={100}
          shadow-camera-bottom={-100}
        />
        <hemisphereLight 
          args={['#87CEEB', '#2a4a2a', 0.5]} 
          groundColor="#2a4a2a"
        />
        
        {/* Névoa mais densa para esconder o pop-up de chunks */}
        <fog attach="fog" args={['#87CEEB', 20, 100]} />
        
        {/* Environment Particles - disabled if low memory */}
        <EnvironmentParticles 
          playerPosition={playerPosition} 
          disabled={memoryWarning} 
        />

        {/* Post Processing - disabled if low memory */}
        <Effects disabled={memoryWarning} />
        
        {/* Mundo */}
        {worldRef.current && (
          <>
            <World 
              world={worldRef.current}
              onPositionChange={setPlayerPosition}
              onRotationChange={setPlayerRotation}
              selectedBlock={availableBlocks[selectedBlockIndex].id}
              onWorldChange={handleWorldChange}
              initialPosition={playerPosition}
              onScrollHotbar={(delta) => {
                setSelectedBlockIndex(prev => {
                  const newIndex = prev + delta;
                  if (newIndex < 0) return availableBlocks.length - 1;
                  if (newIndex >= availableBlocks.length) return 0;
                  return newIndex;
                });
              }}
            />
            
            <ChunksDisplay 
              world={worldRef.current}
              playerPosition={playerPosition}
              updateTrigger={updateTrigger}
            />
            
            {/* Criaturas */}
            {worldRef.current.creatures.map(creature => (
              <Creature 
                key={creature.id} 
                creature={creature}
                world={worldRef.current}
              />
            ))}
            
            {/* Player Model - only visible in 3rd person */}
            <PlayerModel 
              position={playerPosition}
              rotation={playerRotation}
              isMoving={isMoving}
              isThirdPerson={isThirdPerson}
            />
          </>
        )}
        
        {/* Controles de câmera - auto-lock when game starts */}
        <GameControls enabled={!showInstructions} />
        
        {/* Contador de FPS invisível */}
        <FrameCounter fpsRef={fpsRef} />
      </Canvas>

      {/* HUD */}
      <HUD>
        <Crosshair />
        <DebugInfo 
          position={playerPosition}
          chunks={(RENDER_DISTANCE * 2 + 1) ** 2}
          fps={fps}
          memoryUsage={memoryUsage}
        />
        
        {/* Minimap with Radar */}
        <Minimap
          world={worldRef.current}
          playerPosition={playerPosition}
          playerRotation={playerRotation}
          creatures={worldRef.current?.creatures || []}
          showCreatures={true}
          showTerrain={true}
        />
        
        {/* Radar Legend */}
        <div className="radar-legend">
          <div className="radar-legend-item">
            <span className="dot friendly"></span>
            <span>Amigável</span>
          </div>
          <div className="radar-legend-item">
            <span className="dot hostile"></span>
            <span>Hostil</span>
          </div>
        </div>
        
        {/* Enhanced Hotbar with 3D previews */}
        {useEnhancedHotbar ? (
          <EnhancedHotbar 
            blocks={availableBlocks}
            selectedIndex={selectedBlockIndex}
            onSelect={setSelectedBlockIndex}
          />
        ) : (
          <Hotbar 
            blocks={availableBlocks}
            selectedIndex={selectedBlockIndex}
            onSelect={setSelectedBlockIndex}
          />
        )}
      </HUD>

      {/* Underwater Effect */}
      <UnderwaterOverlay isUnderwater={isUnderwater} />
      
      {/* Gallery (Tab key) */}
      <Gallery 
        isOpen={showGallery} 
        onClose={() => setShowGallery(false)}
        blocks={availableBlocks}
      />
      
      {/* Inventory UI (I key) */}
      <InventoryUI
        isOpen={showInventory}
        onClose={() => setShowInventory(false)}
        inventory={inventoryRef.current}
      />
      
      {/* Save/Load Notification */}
      {saveNotification && (
        <div className="save-notification" style={{
          position: 'fixed',
          top: '50%',
          left: '50%',
          transform: 'translate(-50%, -50%)',
          background: saveNotification.type === 'error' ? 'rgba(239, 68, 68, 0.9)' : 'rgba(0, 0, 0, 0.9)',
          backdropFilter: 'blur(10px)',
          border: `2px solid ${saveNotification.type === 'save' ? '#22c55e' : saveNotification.type === 'load' ? '#3b82f6' : '#ef4444'}`,
          borderRadius: '16px',
          padding: '24px 48px',
          color: 'white',
          fontSize: '18px',
          fontWeight: '600',
          zIndex: 400,
          textAlign: 'center'
        }}>
          {saveNotification.message}
        </div>
      )}

      {/* Memory Warning */}
      {memoryWarning && !memoryCritical && (
        <div className="memory-warning">
          <h3>⚠️ Uso de Memória Alto</h3>
          <p>O jogo pode ficar lento. Considere salvar seu progresso.</p>
        </div>
      )}
      
      {/* Memory Critical - Game Over */}
      {memoryCritical && (
        <div className="memory-critical-overlay">
          <h1>⛔ Limite de Memória Atingido</h1>
          <p>O jogo foi pausado para evitar travamentos. Por favor, recarregue a página para continuar jogando.</p>
          <button onClick={() => window.location.reload()}>Recarregar Página</button>
        </div>
      )}

      {/* Custom Game Cursor */}
      <GameCursor />

      {/* Main Menu */}
      {gameState === 'menu' && (
        <MainMenu 
          onNewGame={() => {
            // Reset world logic here if needed
            setGameState('playing');
            setShowInstructions(false);
            if (worldRef.current) {
              // Reset player position for new game
              // ...
            }
          }}
          onContinue={() => {
            setGameState('playing');
            setShowInstructions(false);
          }}
          onLoadSave={(saveData) => {
            if (worldRef.current) {
              applySaveData(saveData, worldRef.current, setPlayerPosition, inventoryRef.current);
              setUpdateTrigger(t => t + 1);
              setGameState('playing');
              setShowInstructions(false);
              setSaveNotification({ type: 'load', message: '✓ Jogo carregado!' });
              setTimeout(() => setSaveNotification(null), 2000);
            }
          }}
        />
      )}

      {/* Pause Menu */}
      {gameState === 'paused' && (
        <PauseMenu 
          onResume={() => setGameState('playing')}
          onSave={() => window.dispatchEvent(new KeyboardEvent('keydown', { code: 'F5' }))}
          onLoad={() => window.dispatchEvent(new KeyboardEvent('keydown', { code: 'F9' }))}
          onNewGame={() => window.location.reload()}
          onMainMenu={() => setGameState('menu')}
          saveMessage={saveNotification?.message}
        />
      )}
      
      {/* Show cursor when in menus */}
      {gameState !== 'playing' && (
        <style>{`
          body { cursor: auto !important; }
          .crosshair { display: none !important; }
        `}</style>
      )}
    </div>
  );
}

// Componente helper para contar frames
function FrameCounter({ fpsRef }) {
  useFrame(() => {
    fpsRef.current.frames++;
  });
  return null;
}

export default App;
