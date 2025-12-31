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
import { HUD, Crosshair, DebugInfo, Hotbar, Instructions, Loading } from './ui/index.js';

// ============================================================================
// Effects Component - DISABLED for memory optimization
// ============================================================================

function Effects() {
  // Post-processing disabled to reduce memory usage
  // SSAO alone was consuming several GB of VRAM
  return null;
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

function World({ world, onPositionChange, selectedBlock, onWorldChange, initialPosition }) {
  const { camera } = useThree();
  
  // Player controls - pass initial spawn position
  usePlayerControls(world, onPositionChange, initialPosition);
  
  // Block interaction
  const { breakBlock, placeBlock } = useBlockInteraction(
    world, 
    selectedBlock, 
    onWorldChange
  );

  // Keyboard handlers for block interaction - Q = break, E = place
  useEffect(() => {
    const handleKeyDown = (e) => {
      if (e.code === 'KeyQ') {
        // Q = break block
        const result = breakBlock();
        console.log('Break block (Q):', result);
      } else if (e.code === 'KeyE') {
        // E = place block
        const result = placeBlock();
        console.log('Place block (E):', result);
      }
    };

    window.addEventListener('keydown', handleKeyDown);

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [breakBlock, placeBlock]);

  return null;
}

// ============================================================================
// Chunk Renderer - Optimized with deferred meshing
// ============================================================================

function ChunkRenderer({ chunk, world, updateTrigger }) {
  const meshRef = useRef();
  const geometryRef = useRef(null);
  const mesher = useMemo(() => new ChunkMesher(), []);
  const pendingUpdate = useRef(false);

  useEffect(() => {
    if (!chunk || !meshRef.current) return;
    if (pendingUpdate.current) return; // Skip if update pending
    
    // Use requestIdleCallback for non-blocking mesh generation
    const generateMesh = () => {
      pendingUpdate.current = false;
      
      if (!chunk || !meshRef.current) return;
      
      // Gera geometria
      const geometry = mesher.generateMesh(chunk, (x, y, z) => world.getBlock(x, y, z));
      
      if (geometry && meshRef.current) {
        // Limpa geometria antiga
        if (geometryRef.current) {
          geometryRef.current.dispose();
        }
        geometryRef.current = geometry;
        meshRef.current.geometry = geometry;
      }
      
      // Reset dirty flag after rendering
      if (chunk) {
        chunk.isDirty = false;
      }
    };
    
    pendingUpdate.current = true;
    
    // Use requestIdleCallback if available, otherwise setTimeout
    if (typeof requestIdleCallback !== 'undefined') {
      requestIdleCallback(generateMesh, { timeout: 50 });
    } else {
      setTimeout(generateMesh, 0);
    }
    
    // Cleanup on unmount
    return () => {
      if (geometryRef.current) {
        geometryRef.current.dispose();
        geometryRef.current = null;
      }
    };
  }, [chunk, chunk?.isDirty, mesher, world, updateTrigger]);

  return (
    <mesh ref={meshRef}>
      <meshStandardMaterial 
        vertexColors 
        side={THREE.DoubleSide}
      />
    </mesh>
  );
}

// ============================================================================
// Chunks Manager Component
// ============================================================================

function ChunksDisplay({ world, playerPosition, updateTrigger }) {
  const [visibleChunks, setVisibleChunks] = useState([]);

  useEffect(() => {
    if (!world || !playerPosition) return;

    const playerCx = Math.floor(playerPosition.x / CHUNK_SIZE);
    const playerCz = Math.floor(playerPosition.z / CHUNK_SIZE);

    const chunks = [];
    
    for (let dx = -RENDER_DISTANCE; dx <= RENDER_DISTANCE; dx++) {
      for (let dz = -RENDER_DISTANCE; dz <= RENDER_DISTANCE; dz++) {
        const cx = playerCx + dx;
        const cz = playerCz + dz;
        
        // Carrega/obtém chunk
        let chunk = world.chunkManager.getChunk(cx, cz);
        if (!chunk) {
          chunk = new Chunk(cx, cz);
          world.terrainGenerator.generateChunk(chunk);
          world.chunkManager.chunks.set(chunk.id, chunk);
        }
        
        chunks.push(chunk);
      }
    }

    setVisibleChunks(chunks);
  }, [world, playerPosition?.x, playerPosition?.z, updateTrigger]);

  return (
    <group>
      {visibleChunks.map(chunk => (
        <ChunkRenderer 
          key={chunk.id} 
          chunk={chunk}
          world={world}
          updateTrigger={updateTrigger}
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
  const [selectedBlockIndex, setSelectedBlockIndex] = useState(0);
  const [updateTrigger, setUpdateTrigger] = useState(0);
  const [fps, setFps] = useState(60);
  const [isUnderwater, setIsUnderwater] = useState(false);
  
  const worldRef = useRef(null);
  const fpsRef = useRef({ frames: 0, lastTime: Date.now() });
  
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

  // Inicializa mundo
  useEffect(() => {
    const world = new VoxelWorld(12345);
    worldRef.current = world;
    
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

  // Key handlers para seleção de bloco
  useEffect(() => {
    const handleKeyDown = (e) => {
      const num = parseInt(e.key);
      if (num >= 1 && num <= 9 && num <= availableBlocks.length) {
        setSelectedBlockIndex(num - 1);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [availableBlocks.length]);

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
        
        {/* Post Processing */}
        <Effects />
        
        {/* Mundo */}
        {worldRef.current && (
          <>
            <World 
              world={worldRef.current}
              onPositionChange={setPlayerPosition}
              selectedBlock={availableBlocks[selectedBlockIndex].id}
              onWorldChange={handleWorldChange}
              initialPosition={playerPosition}
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
        />
        <Hotbar 
          blocks={availableBlocks}
          selectedIndex={selectedBlockIndex}
          onSelect={setSelectedBlockIndex}
        />
      </HUD>

      {/* Instruções iniciais */}
      {showInstructions && (
        <Instructions onStart={handleStart} />
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
