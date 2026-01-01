/**
 * Minimap Component with Radar
 * Shows top-down view of nearby terrain and entities
 */

import React, { useRef, useEffect, useMemo, useCallback } from 'react';
import { BlockTypes, BlockDefinitions } from '../core/blocks/index.js';

// Minimap configuration
const MINIMAP_SIZE = 150;
const MINIMAP_RANGE = 32; // blocks radius
const RADAR_UPDATE_INTERVAL = 200; // ms

export function Minimap({ 
  world, 
  playerPosition, 
  playerRotation = 0,
  creatures = [],
  showCreatures = true,
  showTerrain = true 
}) {
  const canvasRef = useRef(null);
  const lastUpdateRef = useRef(0);
  
  // Block colors for minimap
  const blockColors = useMemo(() => ({
    [BlockTypes.AIR]: 'transparent',
    [BlockTypes.GRASS]: '#4a7c3f',
    [BlockTypes.DIRT]: '#6b5344',
    [BlockTypes.STONE]: '#6d6d6d',
    [BlockTypes.SAND]: '#d4c584',
    [BlockTypes.WATER]: '#3b7cb8',
    [BlockTypes.SNOW]: '#e8e8e8',
    [BlockTypes.WOOD]: '#5a3825',
    [BlockTypes.LEAVES]: '#2d5a27',
    [BlockTypes.OAK_LOG]: '#5a3825',
    [BlockTypes.OAK_LEAVES]: '#2d5a27',
    [BlockTypes.BIRCH_LOG]: '#c4b89a',
    [BlockTypes.BIRCH_LEAVES]: '#6b8c45',
    [BlockTypes.SPRUCE_LOG]: '#3d2a1f',
    [BlockTypes.SPRUCE_LEAVES]: '#1a3d25',
    [BlockTypes.COAL_ORE]: '#2a2a2a',
    [BlockTypes.IRON_ORE]: '#998877',
    [BlockTypes.GOLD_ORE]: '#ccaa44',
    [BlockTypes.DIAMOND_ORE]: '#55dddd',
    [BlockTypes.BEDROCK]: '#111111',
    [BlockTypes.COBBLESTONE]: '#555555',
    [BlockTypes.BRICK]: '#9c5a42',
    [BlockTypes.GLASS]: '#b8d4d8',
    [BlockTypes.ICE]: '#a5d5f5',
    [BlockTypes.CACTUS]: '#2a5a2a',
  }), []);
  
  // Get color for a block type
  const getBlockColor = useCallback((blockType) => {
    return blockColors[blockType] || '#888888';
  }, [blockColors]);
  
  // Draw the minimap
  const drawMinimap = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas || !world || !playerPosition) return;
    
    const ctx = canvas.getContext('2d');
    const centerX = MINIMAP_SIZE / 2;
    const centerY = MINIMAP_SIZE / 2;
    const scale = MINIMAP_SIZE / (MINIMAP_RANGE * 2);
    
    // Clear canvas with dark background
    ctx.fillStyle = 'rgba(10, 10, 20, 0.9)';
    ctx.fillRect(0, 0, MINIMAP_SIZE, MINIMAP_SIZE);
    
    // Draw terrain
    if (showTerrain) {
      const px = Math.floor(playerPosition.x);
      const pz = Math.floor(playerPosition.z);
      
      for (let dx = -MINIMAP_RANGE; dx <= MINIMAP_RANGE; dx += 2) {
        for (let dz = -MINIMAP_RANGE; dz <= MINIMAP_RANGE; dz += 2) {
          const wx = px + dx;
          const wz = pz + dz;
          
          // Get surface block
          const height = world.getHeight(wx, wz);
          const block = world.getBlock(wx, height, wz);
          
          if (block !== BlockTypes.AIR) {
            const color = getBlockColor(block);
            
            // Convert to screen coordinates (rotated by player direction)
            const angle = -playerRotation;
            const rotX = dx * Math.cos(angle) - dz * Math.sin(angle);
            const rotZ = dx * Math.sin(angle) + dz * Math.cos(angle);
            
            const screenX = centerX + rotX * scale;
            const screenY = centerY + rotZ * scale;
            
            // Height-based brightness
            const brightness = Math.min(1, 0.5 + (height / 80));
            
            ctx.fillStyle = color;
            ctx.globalAlpha = brightness;
            ctx.fillRect(screenX - 1, screenY - 1, 3, 3);
          }
        }
      }
      ctx.globalAlpha = 1;
    }
    
    // Draw creatures as radar blips
    if (showCreatures && creatures.length > 0) {
      creatures.forEach(creature => {
        const dx = creature.position.x - playerPosition.x;
        const dz = creature.position.z - playerPosition.z;
        const distance = Math.sqrt(dx * dx + dz * dz);
        
        if (distance <= MINIMAP_RANGE) {
          // Rotate relative to player
          const angle = -playerRotation;
          const rotX = dx * Math.cos(angle) - dz * Math.sin(angle);
          const rotZ = dx * Math.sin(angle) + dz * Math.cos(angle);
          
          const screenX = centerX + rotX * scale;
          const screenY = centerY + rotZ * scale;
          
          // Draw creature blip
          ctx.beginPath();
          ctx.arc(screenX, screenY, 4, 0, Math.PI * 2);
          ctx.fillStyle = creature.hostile ? '#ff4444' : '#44ff44';
          ctx.fill();
          
          // Pulse effect
          ctx.beginPath();
          ctx.arc(screenX, screenY, 6, 0, Math.PI * 2);
          ctx.strokeStyle = creature.hostile ? 'rgba(255, 68, 68, 0.5)' : 'rgba(68, 255, 68, 0.5)';
          ctx.lineWidth = 2;
          ctx.stroke();
        }
      });
    }
    
    // Draw player indicator (arrow in center)
    ctx.save();
    ctx.translate(centerX, centerY);
    
    // Player arrow
    ctx.beginPath();
    ctx.moveTo(0, -8);
    ctx.lineTo(5, 6);
    ctx.lineTo(0, 3);
    ctx.lineTo(-5, 6);
    ctx.closePath();
    
    ctx.fillStyle = '#ffffff';
    ctx.fill();
    ctx.strokeStyle = '#000000';
    ctx.lineWidth = 1;
    ctx.stroke();
    
    ctx.restore();
    
    // Draw compass directions
    ctx.font = 'bold 10px Arial';
    ctx.textAlign = 'center';
    ctx.fillStyle = 'rgba(255, 255, 255, 0.7)';
    
    // Rotate compass based on player rotation
    const dirs = [
      { label: 'N', angle: 0 },
      { label: 'E', angle: Math.PI / 2 },
      { label: 'S', angle: Math.PI },
      { label: 'W', angle: -Math.PI / 2 },
    ];
    
    dirs.forEach(({ label, angle }) => {
      const compassAngle = angle - playerRotation;
      const compassX = centerX + Math.sin(compassAngle) * (MINIMAP_SIZE / 2 - 12);
      const compassY = centerY - Math.cos(compassAngle) * (MINIMAP_SIZE / 2 - 12);
      ctx.fillText(label, compassX, compassY + 3);
    });
    
    // Draw border with glow
    ctx.strokeStyle = 'rgba(100, 150, 255, 0.6)';
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.arc(centerX, centerY, MINIMAP_SIZE / 2 - 2, 0, Math.PI * 2);
    ctx.stroke();
    
    // Inner border
    ctx.strokeStyle = 'rgba(255, 255, 255, 0.2)';
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.arc(centerX, centerY, MINIMAP_SIZE / 2 - 5, 0, Math.PI * 2);
    ctx.stroke();
    
  }, [world, playerPosition, playerRotation, creatures, showCreatures, showTerrain, getBlockColor]);
  
  // Update minimap periodically
  useEffect(() => {
    const updateMinimap = () => {
      const now = Date.now();
      if (now - lastUpdateRef.current >= RADAR_UPDATE_INTERVAL) {
        lastUpdateRef.current = now;
        drawMinimap();
      }
    };
    
    // Initial draw
    drawMinimap();
    
    // Set up interval
    const interval = setInterval(updateMinimap, RADAR_UPDATE_INTERVAL);
    
    return () => clearInterval(interval);
  }, [drawMinimap]);
  
  return (
    <div className="minimap-container">
      <canvas
        ref={canvasRef}
        width={MINIMAP_SIZE}
        height={MINIMAP_SIZE}
        className="minimap-canvas"
      />
      <div className="minimap-label">
        <span className="minimap-coords">
          X: {Math.floor(playerPosition?.x || 0)} Z: {Math.floor(playerPosition?.z || 0)}
        </span>
      </div>
    </div>
  );
}

export default Minimap;
