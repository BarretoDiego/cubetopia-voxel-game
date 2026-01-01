/**
 * MainMenu - Modern main menu with animated background
 */

import React, { useState, useEffect } from 'react';
import { listSaves, loadFromFile } from '../systems/SaveManager.js';

export function MainMenu({ onNewGame, onContinue, onLoadSave }) {
  const [saves, setSaves] = useState([]);
  const [showSaves, setShowSaves] = useState(false);
  const [hasSave, setHasSave] = useState(false);

  useEffect(() => {
    // Check for existing saves
    loadSavesList();
  }, []);

  const loadSavesList = async () => {
    const saveList = await listSaves();
    setSaves(saveList);
    setHasSave(saveList.length > 0);
  };

  const handleContinue = async () => {
    const saveData = await loadFromFile('quicksave');
    if (saveData) {
      onLoadSave(saveData);
    } else if (saves.length > 0) {
      const lastSave = await loadFromFile(saves[0].name);
      if (lastSave) onLoadSave(lastSave);
    }
  };

  return (
    <div className="main-menu">
      {/* Animated background */}
      <div className="menu-background">
        <div className="floating-voxels">
          {[...Array(12)].map((_, i) => (
            <div 
              key={i} 
              className="floating-voxel"
              style={{
                '--delay': `${i * 0.5}s`,
                '--x': `${10 + Math.random() * 80}%`,
                '--duration': `${8 + Math.random() * 8}s`,
                backgroundColor: ['#4ade80', '#60a5fa', '#f472b6', '#fbbf24', '#a78bfa'][i % 5]
              }}
            />
          ))}
        </div>
      </div>

      {/* Main content */}
      <div className="menu-content">
        <h1 className="game-title">
          <span className="title-voxel">⬛</span>
          Voxel World
          <span className="title-voxel">⬛</span>
        </h1>
        <p className="subtitle">Explore • Build • Survive</p>

        <div className="menu-buttons">
          <button className="menu-btn primary" onClick={onNewGame}>
            <span className="btn-icon">🌍</span>
            Novo Jogo
          </button>
          
          {hasSave && (
            <button className="menu-btn" onClick={handleContinue}>
              <span className="btn-icon">▶️</span>
              Continuar
            </button>
          )}
          
          <button className="menu-btn" onClick={() => setShowSaves(!showSaves)}>
            <span className="btn-icon">💾</span>
            {showSaves ? 'Fechar Saves' : 'Gerenciar Saves'}
          </button>
        </div>

        {/* Save slots */}
        {showSaves && (
          <div className="saves-panel">
            <h3>Saves Disponíveis</h3>
            {saves.length === 0 ? (
              <p className="no-saves">Nenhum save encontrado</p>
            ) : (
              <div className="saves-list">
                {saves.map((save, i) => (
                  <div key={i} className="save-slot" onClick={() => onLoadSave(save)}>
                    <span className="save-name">{save.name}</span>
                    <span className="save-date">
                      {new Date(save.timestamp).toLocaleDateString()}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        <div className="menu-footer">
          <p>Controles: WASD = Mover | Mouse = Olhar | ESC = Menu</p>
          <p>Clique Esquerdo = Quebrar | Clique Direito = Colocar | V = Câmera</p>
        </div>
      </div>
    </div>
  );
}

export default MainMenu;
