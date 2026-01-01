/**
 * PauseMenu - In-game pause menu (ESC key)
 */

import React from 'react';

export function PauseMenu({ 
  onResume, 
  onSave, 
  onLoad, 
  onNewGame, 
  onMainMenu,
  saveMessage 
}) {
  return (
    <div className="pause-overlay">
      <div className="pause-menu">
        <h2>⏸️ Pausado</h2>
        
        <div className="pause-buttons">
          <button className="pause-btn primary" onClick={onResume}>
            <span className="btn-icon">▶️</span>
            Continuar
          </button>
          
          <button className="pause-btn" onClick={onSave}>
            <span className="btn-icon">💾</span>
            Salvar Jogo (F5)
          </button>
          
          <button className="pause-btn" onClick={onLoad}>
            <span className="btn-icon">📂</span>
            Carregar Jogo (F9)
          </button>
          
          <div className="pause-divider" />
          
          <button className="pause-btn warning" onClick={onNewGame}>
            <span className="btn-icon">🔄</span>
            Novo Jogo
          </button>
          
          <button className="pause-btn danger" onClick={onMainMenu}>
            <span className="btn-icon">🏠</span>
            Menu Principal
          </button>
        </div>

        {saveMessage && (
          <div className="save-message">{saveMessage}</div>
        )}
        
        <p className="pause-hint">Pressione ESC para voltar ao jogo</p>
      </div>
    </div>
  );
}

export default PauseMenu;
