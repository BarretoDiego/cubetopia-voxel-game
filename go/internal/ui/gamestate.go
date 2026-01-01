// Package ui provides game state and menu management
package ui

// GameState represents the current game state
type GameState int

const (
	StateMainMenu GameState = iota
	StatePlaying
	StatePaused
	StateSettings
	StateLoading
)

// GameStateManager manages game states and transitions
type GameStateManager struct {
	CurrentState  GameState
	PreviousState GameState

	// Callbacks
	OnStateChange func(oldState, newState GameState)
}

// NewGameStateManager creates a new game state manager
func NewGameStateManager() *GameStateManager {
	return &GameStateManager{
		CurrentState:  StateMainMenu,
		PreviousState: StateMainMenu,
	}
}

// SetState changes the game state
func (gsm *GameStateManager) SetState(newState GameState) {
	if gsm.CurrentState == newState {
		return
	}

	gsm.PreviousState = gsm.CurrentState
	gsm.CurrentState = newState

	if gsm.OnStateChange != nil {
		gsm.OnStateChange(gsm.PreviousState, gsm.CurrentState)
	}
}

// TogglePause toggles between playing and paused states
func (gsm *GameStateManager) TogglePause() {
	if gsm.CurrentState == StatePlaying {
		gsm.SetState(StatePaused)
	} else if gsm.CurrentState == StatePaused {
		gsm.SetState(StatePlaying)
	}
}

// IsPlaying returns true if the game is actively playing
func (gsm *GameStateManager) IsPlaying() bool {
	return gsm.CurrentState == StatePlaying
}

// IsPaused returns true if the game is paused
func (gsm *GameStateManager) IsPaused() bool {
	return gsm.CurrentState == StatePaused
}

// IsInMenu returns true if the game is in a menu
func (gsm *GameStateManager) IsInMenu() bool {
	return gsm.CurrentState == StateMainMenu ||
		gsm.CurrentState == StatePaused ||
		gsm.CurrentState == StateSettings
}
