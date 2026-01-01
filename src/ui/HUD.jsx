/**
 * Componentes de UI - HUD
 */

import React from 'react';
import { BlockPreview3D, CreaturePreview3D, ItemPreview3D } from './Preview3D.jsx';


export function HUD({ children }) {
  return (
    <div className="hud">
      {/* Help Panel - Shows controls */}
      <div className="help-panel" style={{
        position: 'absolute',
        top: '16px',
        right: '16px',
        background: 'rgba(0, 0, 0, 0.6)',
        backdropFilter: 'blur(4px)',
        padding: '16px',
        borderRadius: '12px',
        color: 'white',
        border: '1px solid rgba(255, 255, 255, 0.1)',
        maxWidth: '250px'
      }}>
        <h3 style={{ margin: '0 0 12px 0', fontSize: '14px', color: '#fbbf24', display: 'flex', alignItems: 'center', gap: '8px' }}>
          <span>⌨️</span> Controles
        </h3>
        <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr', gap: '8px 16px', fontSize: '12px' }}>
          <span style={{ color: '#aaa' }}>W A S D</span> <span>Mover</span>
          <span style={{ color: '#aaa' }}>Espaço</span> <span>Pular</span>
          <span style={{ color: '#aaa' }}>Shift</span> <span>Correr</span>
          <span style={{ color: '#aaa' }}>🖱️ Esq</span> <span>Quebrar</span>
          <span style={{ color: '#aaa' }}>🖱️ Dir</span> <span>Colocar</span>
          <span style={{ color: '#aaa' }}>Scroll</span> <span>Selecionar</span>
          <span style={{ color: '#aaa' }}>V</span> <span>Câmera</span>
          <span style={{ color: '#aaa' }}>I</span> <span>Inventário</span>
          <span style={{ color: '#aaa' }}>TAB</span> <span>Galeria</span>
          <span style={{ color: '#aaa' }}>F5 / F9</span> <span>Save / Load</span>
          <span style={{ color: '#aaa' }}>ESC</span> <span>Menu / Pausa</span>
        </div>
      </div>

      {/* Menu Button - Quick access to pause menu */}
      <button 
        className="hud-menu-btn"
        onClick={() => window.dispatchEvent(new KeyboardEvent('keydown', { code: 'Escape' }))}
        style={{
          position: 'absolute',
          top: '16px',
          right: '280px', // Left of help panel
          background: 'rgba(255, 255, 255, 0.1)',
          border: '1px solid rgba(255, 255, 255, 0.2)',
          borderRadius: '8px',
          padding: '8px 16px',
          color: 'white',
          cursor: 'pointer',
          backdropFilter: 'blur(4px)',
          display: 'flex',
          alignItems: 'center',
          gap: '8px',
          transition: 'all 0.2s'
        }}
        onMouseEnter={e => e.currentTarget.style.background = 'rgba(255, 255, 255, 0.2)'}
        onMouseLeave={e => e.currentTarget.style.background = 'rgba(255, 255, 255, 0.1)'}
      >
        <span>⏸️</span> Menu
      </button>

      {children}
    </div>
  );
}

export function Crosshair() {
  return <div className="crosshair" />;
}

export function DebugInfo({ position, chunks, fps, biome, memoryUsage }) {
  return (
    <div className="debug-panel">
      <h1>Voxel Engine</h1>
      <div className="stat">
        <span className="label">Posição</span>
        <span className="value">{position.x}, {position.y}, {position.z}</span>
      </div>
      <div className="stat">
        <span className="label">Chunks</span>
        <span className="value">{chunks}</span>
      </div>
      <div className="stat">
        <span className="label">FPS</span>
        <span className="value">{fps}</span>
      </div>
      {biome && (
        <div className="stat">
          <span className="label">Bioma</span>
          <span className="value">{biome}</span>
        </div>
      )}
      {memoryUsage && (
        <div className="stat">
          <span className="label">Memória</span>
          <span className="value" style={{ 
            color: memoryUsage.percentage > 0.8 ? '#ef4444' : 
                   memoryUsage.percentage > 0.6 ? '#f59e0b' : '#60a5fa' 
          }}>
            {memoryUsage.used}MB / {memoryUsage.limit}MB
          </span>
        </div>
      )}
    </div>
  );
}

export function Hotbar({ blocks, selectedIndex, onSelect, inventory = {} }) {
  return (
    <div className="hotbar">
      {blocks.map((block, index) => {
        const count = inventory[block.id] || 99; // Default to 99 if no inventory system
        return (
          <div
            key={block.id}
            className={`hotbar-slot ${index === selectedIndex ? 'selected' : ''}`}
            onClick={() => onSelect(index)}
          >
            <span className="key-hint">{index + 1}</span>
            <div className="hotbar-preview-container" style={{ width: '40px', height: '40px', pointerEvents: 'none' }}>
              <BlockPreview3D color={block.color} size="100%" />
            </div>
            <span className="block-count">{count}</span>
            <span className="block-name">{block.name}</span>
          </div>
        );
      })}
    </div>
  );
}

export function Instructions({ onStart }) {
  return (
    <div className="instructions">
      <h2>🎮 Controles</h2>
      <p><span className="key">W A S D</span> - Movimento</p>
      <p><span className="key">Mouse</span> - Olhar</p>
      <p><span className="key">Espaço</span> - Pular</p>
      <p><span className="key">Shift</span> - Correr</p>
      <p><span className="key">Q</span> - Quebrar bloco</p>
      <p><span className="key">E</span> - Colocar bloco</p>
      <p><span className="key">1-9</span> - Selecionar bloco</p>
      <p><span className="key">I</span> - Inventário</p>
      <p><span className="key">Tab</span> - Galeria</p>
      <p><span className="key">G</span> - Destravar (se preso)</p>
      <button className="start-btn" onClick={onStart}>
        Clique para Jogar
      </button>
    </div>
  );
}

export function Loading({ message = 'Carregando...' }) {
  return (
    <div className="loading">
      <h1>🌍 Gerando Mundo...</h1>
      <div className="spinner" />
      <p style={{ marginTop: 16, opacity: 0.7 }}>{message}</p>
    </div>
  );
}

// Underwater visual effect overlay
export function UnderwaterOverlay({ isUnderwater }) {
  if (!isUnderwater) return null;
  
  return (
    <div className="underwater-overlay">
      <div style={{
        position: 'absolute',
        bottom: 20,
        left: '50%',
        transform: 'translateX(-50%)',
        color: 'rgba(255, 255, 255, 0.4)',
        fontSize: 12,
        fontFamily: 'monospace'
      }}>
        🌊 Subaquático
      </div>
    </div>
  );
}

// Gallery for viewing all items, blocks, and creatures
export function Gallery({ isOpen, onClose, blocks, creatures }) {
  const [activeTab, setActiveTab] = React.useState('blocks');
  
  if (!isOpen) return null;
  
  const creatures3D = [
    { type: 'slime', name: 'Slime', color: '#22CC22' },
    { type: 'pig', name: 'Porco', color: '#FFB6C1' },
    { type: 'zombie', name: 'Zumbi', color: '#5A8A5A' },
    { type: 'spider', name: 'Aranha', color: '#2F1F1F' },
    { type: 'bird', name: 'Pássaro', color: '#FF6347' },
    { type: 'cow', name: 'Vaca', color: '#FFFFFF' },
  ];
  
  const items3D = [
    { type: 'sword', name: 'Espada de Ferro', color: '#C0C0C0' },
    { type: 'pickaxe', name: 'Picareta de Pedra', color: '#808080' },
    { type: 'axe', name: 'Machado de Madeira', color: '#8B4513' },
    { type: 'apple', name: 'Maçã', color: '#FF0000' },
    { type: 'diamond', name: 'Diamante', color: '#00FFFF' },
  ];
  
  return (
    <div className="gallery-overlay">
      <div className="gallery-header">
        <h2>📦 Galeria</h2>
        <button className="gallery-close" onClick={onClose}>
          Fechar (Tab)
        </button>
      </div>
      
      <div className="gallery-tabs">
        <button 
          className={`gallery-tab ${activeTab === 'blocks' ? 'active' : ''}`}
          onClick={() => setActiveTab('blocks')}
        >
          🧱 Blocos
        </button>
        <button 
          className={`gallery-tab ${activeTab === 'creatures' ? 'active' : ''}`}
          onClick={() => setActiveTab('creatures')}
        >
          🐾 Criaturas
        </button>
        <button 
          className={`gallery-tab ${activeTab === 'items' ? 'active' : ''}`}
          onClick={() => setActiveTab('items')}
        >
          ⚔️ Itens
        </button>
      </div>
      
      <div className="gallery-grid">
        {activeTab === 'blocks' && blocks?.map((block, i) => (
          <div key={i} className="gallery-item">
            <BlockPreview3D color={block.color} size={120} />
            <div className="name">{block.name}</div>
          </div>
        ))}
        
        {activeTab === 'creatures' && creatures3D.map((creature, i) => (
          <div key={i} className="gallery-item">
            <CreaturePreview3D type={creature.type} color={creature.color} size={120} />
            <div className="name">{creature.name}</div>
          </div>
        ))}
        
        {activeTab === 'items' && items3D.map((item, i) => (
          <div key={i} className="gallery-item">
            <ItemPreview3D type={item.type} color={item.color} size={120} />
            <div className="name">{item.name}</div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default { HUD, Crosshair, DebugInfo, Hotbar, Instructions, Loading, UnderwaterOverlay, Gallery };
