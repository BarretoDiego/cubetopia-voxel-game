// Package ui provides settings management
package ui

// Settings holds all game settings
type Settings struct {
	// Graphics
	RenderDistance    int
	EnableFXAA        bool
	EnableBloom       bool
	EnablePostProcess bool
	EnableRaytracing  bool // Raytracing toggle
	BloomStrength     float32
	FOV               float32
	VSync             bool

	// Gameplay
	MouseSensitivity float32
	InvertY          bool

	// Audio
	MasterVolume float32
	MusicVolume  float32
	SFXVolume    float32

	// Performance
	MaxChunks         int
	ChunkLoadPerFrame int
}

// DefaultSettings returns default settings
func DefaultSettings() *Settings {
	return &Settings{
		// Graphics
		RenderDistance:    3,
		EnableFXAA:        true,
		EnableBloom:       true,
		EnablePostProcess: true,
		EnableRaytracing:  false, // Disabled by default (performance)
		BloomStrength:     0.15,
		FOV:               75.0,
		VSync:             true,

		// Gameplay
		MouseSensitivity: 0.1,
		InvertY:          false,

		// Audio
		MasterVolume: 1.0,
		MusicVolume:  0.7,
		SFXVolume:    1.0,

		// Performance
		MaxChunks:         100,
		ChunkLoadPerFrame: 2,
	}
}

// SettingItem represents a settings item
type SettingItem struct {
	Name     string
	Type     SettingType
	Value    interface{}
	Min, Max float32
	Options  []string
	OnChange func(value interface{})
}

// SettingType defines the type of setting
type SettingType int

const (
	SettingBool SettingType = iota
	SettingFloat
	SettingInt
	SettingOption
)

// SettingsMenu provides a settings interface
type SettingsMenu struct {
	Items         []SettingItem
	SelectedIndex int
	IsVisible     bool
	Settings      *Settings
}

// NewSettingsMenu creates a settings menu
func NewSettingsMenu(settings *Settings) *SettingsMenu {
	sm := &SettingsMenu{
		Settings:  settings,
		IsVisible: false,
	}

	sm.Items = []SettingItem{
		{
			Name: "Render Distance",
			Type: SettingInt,
			Min:  1, Max: 8,
			OnChange: func(v interface{}) {
				settings.RenderDistance = v.(int)
			},
		},
		{
			Name: "FXAA",
			Type: SettingBool,
			OnChange: func(v interface{}) {
				settings.EnableFXAA = v.(bool)
			},
		},
		{
			Name: "Bloom",
			Type: SettingBool,
			OnChange: func(v interface{}) {
				settings.EnableBloom = v.(bool)
			},
		},
		{
			Name: "Post Processing",
			Type: SettingBool,
			OnChange: func(v interface{}) {
				settings.EnablePostProcess = v.(bool)
			},
		},
		{
			Name: "Raytracing",
			Type: SettingBool,
			OnChange: func(v interface{}) {
				settings.EnableRaytracing = v.(bool)
			},
		},
		{
			Name: "Bloom Strength",
			Type: SettingFloat,
			Min:  0.0, Max: 1.0,
			OnChange: func(v interface{}) {
				settings.BloomStrength = v.(float32)
			},
		},
		{
			Name: "FOV",
			Type: SettingFloat,
			Min:  60, Max: 120,
			OnChange: func(v interface{}) {
				settings.FOV = v.(float32)
			},
		},
		{
			Name: "Mouse Sensitivity",
			Type: SettingFloat,
			Min:  0.01, Max: 0.5,
			OnChange: func(v interface{}) {
				settings.MouseSensitivity = v.(float32)
			},
		},
		{
			Name: "Invert Y",
			Type: SettingBool,
			OnChange: func(v interface{}) {
				settings.InvertY = v.(bool)
			},
		},
		{
			Name: "VSync",
			Type: SettingBool,
			OnChange: func(v interface{}) {
				settings.VSync = v.(bool)
			},
		},
	}

	return sm
}

// ToggleCurrentSetting toggles a bool setting or adjusts numeric
func (sm *SettingsMenu) ToggleCurrentSetting(delta float32) {
	if sm.SelectedIndex < 0 || sm.SelectedIndex >= len(sm.Items) {
		return
	}

	item := &sm.Items[sm.SelectedIndex]

	switch item.Type {
	case SettingBool:
		current := sm.getSettingValue(item.Name).(bool)
		item.OnChange(!current)

	case SettingFloat:
		current := sm.getSettingValue(item.Name).(float32)
		step := (item.Max - item.Min) / 20.0
		newVal := current + step*delta
		if newVal < item.Min {
			newVal = item.Min
		}
		if newVal > item.Max {
			newVal = item.Max
		}
		item.OnChange(newVal)

	case SettingInt:
		current := sm.getSettingValue(item.Name).(int)
		newVal := current + int(delta)
		if newVal < int(item.Min) {
			newVal = int(item.Min)
		}
		if newVal > int(item.Max) {
			newVal = int(item.Max)
		}
		item.OnChange(newVal)
	}
}

func (sm *SettingsMenu) getSettingValue(name string) interface{} {
	switch name {
	case "Render Distance":
		return sm.Settings.RenderDistance
	case "FXAA":
		return sm.Settings.EnableFXAA
	case "Bloom":
		return sm.Settings.EnableBloom
	case "Post Processing":
		return sm.Settings.EnablePostProcess
	case "Raytracing":
		return sm.Settings.EnableRaytracing
	case "Bloom Strength":
		return sm.Settings.BloomStrength
	case "FOV":
		return sm.Settings.FOV
	case "Mouse Sensitivity":
		return sm.Settings.MouseSensitivity
	case "Invert Y":
		return sm.Settings.InvertY
	case "VSync":
		return sm.Settings.VSync
	}
	return nil
}

// SelectNext selects the next setting
func (sm *SettingsMenu) SelectNext() {
	sm.SelectedIndex = (sm.SelectedIndex + 1) % len(sm.Items)
}

// SelectPrevious selects the previous setting
func (sm *SettingsMenu) SelectPrevious() {
	sm.SelectedIndex--
	if sm.SelectedIndex < 0 {
		sm.SelectedIndex = len(sm.Items) - 1
	}
}
