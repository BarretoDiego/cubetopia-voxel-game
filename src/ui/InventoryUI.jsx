/**
 * Full Inventory UI Component
 * Opens with 'I' key - displays 36 inventory slots
 */

import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { itemRegistry } from '../core/items/ItemRegistry.js';
import { BlockDefinitions } from '../core/blocks/BlockTypes.js';
import { Preview3D } from './Preview3D.jsx';

export function InventoryUI({ isOpen, onClose, inventory, onSlotClick, onSlotDrop }) {
  const [draggedSlot, setDraggedSlot] = useState(null);
  const [hoveredSlot, setHoveredSlot] = useState(null);
  
  // Get all 36 slots from inventory - must be before any conditional returns
  const slots = useMemo(() => {
    return inventory ? inventory.slots : new Array(36).fill(null);
  }, [inventory]);
  
  const hotbarSize = 9;
  
  // Get item info for a slot
  const getSlotInfo = useCallback((slot) => {
    if (!slot) return null;
    
    // Check if it's an item
    const itemDef = itemRegistry.get(slot.itemId);
    if (itemDef) {
       return {
         type: 'item',
         variant: resolveItemVariant(slot.itemId), // Helper to map ID to string variant if needed
         name: itemDef.name,
         color: itemDef.color,
         count: slot.count,
         durability: slot.durability,
         itemId: slot.itemId
       };
    }
    
    // Check if it's a block (fallback)
    const blockDef = BlockDefinitions[slot.itemId];
    if (blockDef) {
      return {
        type: 'block',
        name: blockDef.name,
        color: blockDef.color,
        count: slot.count,
        durability: null,
        itemId: slot.itemId
      };
    }

    return {
      type: 'unknown',
      name: 'Desconhecido',
      color: '#888',
      count: slot.count,
      durability: null
    };
  }, []);

  // Helper to map item IDs to 3D model variants
  const resolveItemVariant = (id) => {
     // Map IDs from ItemTypes.js to strings expected by ItemModel in Preview3D.jsx
     // Weapons
     if (id >= 100 && id <= 102) return 'sword';
     // Pickaxes
     if (id >= 200 && id <= 202) return 'pickaxe';
     // Axes
     if (id >= 210 && id <= 211) return 'axe'; // Mapped to axe? Preview3D has 'axe'? Preview3D has 'axe' (added in my thought process, need to verify implementation)
     // Consumables
     if (id === 300) return 'apple';
     if (id === 402) return 'diamond'; // Diamond material
     
     return 'default';
  };
  
  // Handle slot click
  const handleSlotClick = useCallback((index) => {
    if (onSlotClick) {
      onSlotClick(index);
    }
  }, [onSlotClick]);
  
  // Handle drag start
  const handleDragStart = useCallback((index) => {
    setDraggedSlot(index);
  }, []);
  
  // Handle drag end
  const handleDragEnd = useCallback(() => {
    setDraggedSlot(null);
  }, []);
  
  // Handle drop
  const handleDrop = useCallback((targetIndex) => {
    if (draggedSlot !== null && draggedSlot !== targetIndex) {
      if (onSlotDrop) {
        onSlotDrop(draggedSlot, targetIndex);
      }
    }
    setDraggedSlot(null);
  }, [draggedSlot, onSlotDrop]);
  
  // Close on Escape key - MUST be before any conditional returns
  useEffect(() => {
    if (!isOpen) return;
    
    const handleKeyDown = (e) => {
      if (e.code === 'Escape' || e.code === 'KeyI') {
        e.preventDefault();
        onClose();
      }
    };
    
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);
  
  // Reset state when closing
  useEffect(() => {
    if (!isOpen) {
      setDraggedSlot(null);
      setHoveredSlot(null);
    }
  }, [isOpen]);
  
  // Don't render if not open - AFTER all hooks
  if (!isOpen) return null;
  
  // Render a single slot
  const renderSlot = (slot, index) => {
    const slotInfo = getSlotInfo(slot);
    const isHotbar = index < hotbarSize;
    const isDragging = draggedSlot === index;
    const isDropTarget = hoveredSlot === index && draggedSlot !== null && draggedSlot !== index;
    
    return (
      <div
        key={index}
        className={`inventory-slot ${isHotbar ? 'hotbar-row' : ''} ${isDragging ? 'dragging' : ''} ${isDropTarget ? 'drop-target' : ''}`}
        onClick={() => handleSlotClick(index)}
        draggable={slot !== null}
        onDragStart={() => handleDragStart(index)}
        onDragEnd={handleDragEnd}
        onDragOver={(e) => { e.preventDefault(); setHoveredSlot(index); }}
        onDragLeave={() => setHoveredSlot(null)}
        onDrop={() => handleDrop(index)}
      >
        {slotInfo && (
          <>
            <div className="item-preview-container">
               {/* 3D Preview */}
               <Preview3D 
                 type={slotInfo.type} 
                 variant={slotInfo.variant} 
                 color={slotInfo.color} 
                 size="100%" 
               />
            </div>

            {slotInfo.count > 1 && (
              <span className="item-count">{slotInfo.count}</span>
            )}
            {slotInfo.durability !== null && (
              <div className="durability-bar">
                <div 
                  className="durability-fill"
                  style={{ 
                    width: `${(slotInfo.durability / 100) * 100}%`,
                    backgroundColor: slotInfo.durability > 50 ? '#22c55e' : 
                                    slotInfo.durability > 25 ? '#f59e0b' : '#ef4444'
                  }}
                />
              </div>
            )}
            <div className="tooltip">{slotInfo.name}</div>
          </>
        )}
        {isHotbar && (
          <span className="slot-number">{index + 1}</span>
        )}
      </div>
    );
  };
  
  return (
    <div className="inventory-overlay" onClick={(e) => e.target === e.currentTarget && onClose()}>
      <div className="inventory-container">
        <div className="inventory-header">
          <h2>📦 Inventário</h2>
          <button className="inventory-close" onClick={onClose}>
            ✕
          </button>
        </div>
        
        {/* Main inventory grid (27 slots, 3 rows) */}
        <div className="inventory-section">
          <div className="inventory-label">Inventário</div>
          <div className="inventory-grid main">
            {slots.slice(hotbarSize, 36).map((slot, i) => renderSlot(slot, i + hotbarSize))}
          </div>
        </div>
        
        {/* Hotbar (9 slots) */}
        <div className="inventory-section">
          <div className="inventory-label">Hotbar</div>
          <div className="inventory-grid hotbar">
            {slots.slice(0, hotbarSize).map((slot, i) => renderSlot(slot, i))}
          </div>
        </div>
        
        {/* Tooltip */}
        {hoveredSlot !== null && slots[hoveredSlot] && (
          <div className="inventory-tooltip">
            {getSlotInfo(slots[hoveredSlot])?.name}
          </div>
        )}
        
        <div className="inventory-help">
          Pressione <kbd>I</kbd> ou <kbd>ESC</kbd> para fechar
        </div>
      </div>
    </div>
  );
}

export default InventoryUI;
