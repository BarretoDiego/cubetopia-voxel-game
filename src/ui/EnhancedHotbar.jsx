/**
 * Enhanced Hotbar with 3D Block Previews
 * Modern, visual hotbar with animated 3D block previews
 */

import React, { useRef, useEffect, useState, useMemo } from 'react';
import * as THREE from 'three';

// Create a simple 3D block preview using canvas
function BlockPreviewCanvas({ color, size = 36, isSelected }) {
  const canvasRef = useRef(null);
  const animationRef = useRef(0);
  const rotationRef = useRef(0);
  
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    const centerX = size / 2;
    const centerY = size / 2;
    const blockSize = size * 0.4;
    
    // Parse color
    const baseColor = color || '#888888';
    
    // Create lighter and darker variants
    const lighten = (hex, percent) => {
      const num = parseInt(hex.slice(1), 16);
      const amt = Math.round(2.55 * percent);
      const R = Math.min(255, (num >> 16) + amt);
      const G = Math.min(255, ((num >> 8) & 0x00FF) + amt);
      const B = Math.min(255, (num & 0x0000FF) + amt);
      return `rgb(${R},${G},${B})`;
    };
    
    const darken = (hex, percent) => {
      const num = parseInt(hex.slice(1), 16);
      const amt = Math.round(2.55 * percent);
      const R = Math.max(0, (num >> 16) - amt);
      const G = Math.max(0, ((num >> 8) & 0x00FF) - amt);
      const B = Math.max(0, (num & 0x0000FF) - amt);
      return `rgb(${R},${G},${B})`;
    };
    
    const draw = () => {
      ctx.clearRect(0, 0, size, size);
      
      const rotation = rotationRef.current;
      const offsetX = Math.sin(rotation) * 2;
      const offsetY = Math.cos(rotation * 0.5) * 1;
      
      // Draw isometric cube
      const x = centerX + offsetX;
      const y = centerY + offsetY;
      const h = blockSize * 0.5;
      const w = blockSize * 0.866;
      
      // Top face (brightest)
      ctx.beginPath();
      ctx.moveTo(x, y - h);
      ctx.lineTo(x + w, y - h * 0.5);
      ctx.lineTo(x, y);
      ctx.lineTo(x - w, y - h * 0.5);
      ctx.closePath();
      ctx.fillStyle = lighten(baseColor, 30);
      ctx.fill();
      
      // Right face
      ctx.beginPath();
      ctx.moveTo(x, y);
      ctx.lineTo(x + w, y - h * 0.5);
      ctx.lineTo(x + w, y + h * 0.5);
      ctx.lineTo(x, y + h);
      ctx.closePath();
      ctx.fillStyle = baseColor;
      ctx.fill();
      
      // Left face (darkest)
      ctx.beginPath();
      ctx.moveTo(x, y);
      ctx.lineTo(x - w, y - h * 0.5);
      ctx.lineTo(x - w, y + h * 0.5);
      ctx.lineTo(x, y + h);
      ctx.closePath();
      ctx.fillStyle = darken(baseColor, 25);
      ctx.fill();
      
      // Add subtle outline
      ctx.strokeStyle = 'rgba(0,0,0,0.3)';
      ctx.lineWidth = 1;
      
      // Top outline
      ctx.beginPath();
      ctx.moveTo(x, y - h);
      ctx.lineTo(x + w, y - h * 0.5);
      ctx.lineTo(x, y);
      ctx.lineTo(x - w, y - h * 0.5);
      ctx.closePath();
      ctx.stroke();
      
      // Update rotation for animation
      if (isSelected) {
        rotationRef.current += 0.05;
        animationRef.current = requestAnimationFrame(draw);
      }
    };
    
    draw();
    
    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
    };
  }, [color, size, isSelected]);
  
  return (
    <canvas
      ref={canvasRef}
      width={size}
      height={size}
      style={{ display: 'block' }}
    />
  );
}

export function EnhancedHotbar({ blocks, selectedIndex, onSelect, inventory = {} }) {
  const [hoverIndex, setHoverIndex] = useState(null);
  
  return (
    <div className="enhanced-hotbar">
      {blocks.map((block, index) => {
        const count = inventory[block.id] || 99;
        const isSelected = index === selectedIndex;
        const isHovered = index === hoverIndex;
        
        return (
          <div
            key={block.id}
            className={`enhanced-hotbar-slot ${isSelected ? 'selected' : ''} ${isHovered ? 'hovered' : ''}`}
            onClick={() => onSelect(index)}
            onMouseEnter={() => setHoverIndex(index)}
            onMouseLeave={() => setHoverIndex(null)}
          >
            {/* Key hint */}
            <span className="slot-key">{index + 1}</span>
            
            {/* 3D Block preview */}
            <div className="slot-preview">
              <BlockPreviewCanvas 
                color={block.color} 
                size={40} 
                isSelected={isSelected}
              />
            </div>
            
            {/* Item count */}
            <span className="slot-count">{count}</span>
            
            {/* Block name tooltip */}
            {(isSelected || isHovered) && (
              <div className="slot-tooltip">
                {block.name}
              </div>
            )}
            
            {/* Selection glow */}
            {isSelected && <div className="slot-glow" />}
          </div>
        );
      })}
    </div>
  );
}

export default EnhancedHotbar;
