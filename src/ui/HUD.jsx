/**
 * Componentes de UI - HUD
 */

import React from 'react';

export function HUD({ children }) {
  return (
    <div className="hud">
      {children}
    </div>
  );
}

export function Crosshair() {
  return <div className="crosshair" />;
}

export function DebugInfo({ position, chunks, fps, biome }) {
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
            <div 
              className="block-preview"
              style={{ backgroundColor: block.color }}
            />
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

export default { HUD, Crosshair, DebugInfo, Hotbar, Instructions, Loading };
