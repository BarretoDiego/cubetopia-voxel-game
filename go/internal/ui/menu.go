// Package ui provides menu rendering
package ui

// MenuItem represents a menu item
type MenuItem struct {
	Text     string
	Action   func()
	Selected bool
	Enabled  bool
	Color    [4]float32 // Specific color for this item
}

// Menu represents a menu screen
type Menu struct {
	Title         string
	Items         []MenuItem
	SelectedIndex int
	IsVisible     bool

	// Styling
	BackgroundColor [4]float32
	TitleColor      [4]float32
	ItemColor       [4]float32
	SelectedColor   [4]float32
	DisabledColor   [4]float32
}

// NewMainMenu creates the main menu
func NewMainMenu(onNewGame, onLoadGame, onSettings, onQuit func()) *Menu {
	return &Menu{
		Title: "VOXEL ENGINE",
		Items: []MenuItem{
			{Text: "New Game", Action: onNewGame, Enabled: true, Color: [4]float32{0.2, 0.8, 0.2, 1.0}},   // Green
			{Text: "Load Game", Action: onLoadGame, Enabled: true, Color: [4]float32{0.2, 0.4, 0.8, 1.0}}, // Blue
			{Text: "Settings", Action: onSettings, Enabled: true, Color: [4]float32{0.5, 0.5, 0.5, 1.0}},  // Gray
			{Text: "Quit", Action: onQuit, Enabled: true, Color: [4]float32{0.8, 0.2, 0.2, 1.0}},          // Red
		},
		SelectedIndex:   0,
		IsVisible:       true,
		BackgroundColor: [4]float32{0.1, 0.1, 0.15, 0.95},
		TitleColor:      [4]float32{1.0, 1.0, 1.0, 1.0},
		ItemColor:       [4]float32{0.7, 0.7, 0.7, 1.0},
		SelectedColor:   [4]float32{1.0, 0.8, 0.2, 1.0},
		DisabledColor:   [4]float32{0.4, 0.4, 0.4, 0.5},
	}
}

// NewPauseMenu creates the pause menu
func NewPauseMenu(onResume, onSettings, onSave, onMainMenu func()) *Menu {
	return &Menu{
		Title: "PAUSED",
		Items: []MenuItem{
			{Text: "Resume", Action: onResume, Enabled: true, Color: [4]float32{0.2, 0.8, 0.2, 1.0}},      // Green
			{Text: "Settings", Action: onSettings, Enabled: true, Color: [4]float32{0.5, 0.5, 0.5, 1.0}},  // Gray
			{Text: "Save Game", Action: onSave, Enabled: true, Color: [4]float32{0.8, 0.6, 0.2, 1.0}},     // Orange
			{Text: "Main Menu", Action: onMainMenu, Enabled: true, Color: [4]float32{0.8, 0.2, 0.2, 1.0}}, // Red
		},
		SelectedIndex:   0,
		IsVisible:       true,
		BackgroundColor: [4]float32{0.0, 0.0, 0.0, 0.7},
		TitleColor:      [4]float32{1.0, 1.0, 1.0, 1.0},
		ItemColor:       [4]float32{0.8, 0.8, 0.8, 1.0},
		SelectedColor:   [4]float32{1.0, 0.8, 0.2, 1.0},
		DisabledColor:   [4]float32{0.4, 0.4, 0.4, 0.5},
	}
}

// SelectNext selects the next menu item
func (m *Menu) SelectNext() {
	for i := 0; i < len(m.Items); i++ {
		m.SelectedIndex = (m.SelectedIndex + 1) % len(m.Items)
		if m.Items[m.SelectedIndex].Enabled {
			break
		}
	}
}

// SelectPrevious selects the previous menu item
func (m *Menu) SelectPrevious() {
	for i := 0; i < len(m.Items); i++ {
		m.SelectedIndex--
		if m.SelectedIndex < 0 {
			m.SelectedIndex = len(m.Items) - 1
		}
		if m.Items[m.SelectedIndex].Enabled {
			break
		}
	}
}

// Confirm executes the selected menu item
func (m *Menu) Confirm() {
	if m.SelectedIndex >= 0 && m.SelectedIndex < len(m.Items) {
		item := m.Items[m.SelectedIndex]
		if item.Enabled && item.Action != nil {
			item.Action()
		}
	}
}

// MenuRenderer renders menus
type MenuRenderer struct {
	uiRenderer *Renderer
}

// NewMenuRenderer creates a new menu renderer
func NewMenuRenderer(uiRenderer *Renderer) *MenuRenderer {
	return &MenuRenderer{
		uiRenderer: uiRenderer,
	}
}

// RenderMenu renders a menu
func (mr *MenuRenderer) RenderMenu(menu *Menu, screenWidth, screenHeight int) {
	if menu == nil || !menu.IsVisible || mr.uiRenderer == nil {
		return
	}

	// Background overlay
	mr.uiRenderer.DrawRect(0, 0, float32(screenWidth), float32(screenHeight), menu.BackgroundColor)

	// Menu panel
	panelWidth := float32(400)
	panelHeight := float32(50 + len(menu.Items)*60 + 100)
	panelX := (float32(screenWidth) - panelWidth) / 2
	panelY := (float32(screenHeight) - panelHeight) / 2

	// Panel background
	mr.uiRenderer.DrawRect(panelX, panelY, panelWidth, panelHeight, [4]float32{0.15, 0.15, 0.2, 0.9})

	// Panel border
	mr.uiRenderer.DrawRect(panelX, panelY, panelWidth, 3, [4]float32{0.3, 0.5, 0.8, 1.0})
	mr.uiRenderer.DrawRect(panelX, panelY+panelHeight-3, panelWidth, 3, [4]float32{0.3, 0.5, 0.8, 1.0})

	// Title bar
	titleBarHeight := float32(60)
	mr.uiRenderer.DrawRect(panelX, panelY, panelWidth, titleBarHeight, [4]float32{0.2, 0.3, 0.5, 1.0})

	// Title indicator
	titleScale := float32(3.0)
	titleWidth := float32(len(menu.Title)*6) * titleScale
	mr.uiRenderer.DrawText(
		panelX+(panelWidth-titleWidth)/2,
		panelY+15,
		titleScale,
		menu.Title,
		menu.TitleColor,
	)

	// Menu items
	itemY := panelY + titleBarHeight + 30
	itemHeight := float32(40)
	itemSpacing := float32(10)

	for i, item := range menu.Items {
		itemX := panelX + 40
		itemWidth := panelWidth - 80

		// Item background
		bgColor := [4]float32{0.1, 0.1, 0.15, 0.5}
		if i == menu.SelectedIndex {
			bgColor = [4]float32{0.3, 0.4, 0.6, 0.8}
		}
		mr.uiRenderer.DrawRect(itemX, itemY, itemWidth, itemHeight, bgColor)

		// Selection indicator
		if i == menu.SelectedIndex {
			mr.uiRenderer.DrawRect(itemX, itemY, 4, itemHeight, menu.SelectedColor)
		}

		// Item Text
		var textColor [4]float32
		if !item.Enabled {
			textColor = menu.DisabledColor
		} else if i == menu.SelectedIndex {
			textColor = menu.SelectedColor
		} else if item.Color[3] > 0 { // Use specific color if set
			textColor = item.Color
		} else {
			textColor = menu.ItemColor
		}

		mr.uiRenderer.DrawText(
			itemX+20,
			itemY+10,
			2.0, // Scale 2.0
			item.Text,
			textColor,
		)

		itemY += itemHeight + itemSpacing
	}

	// Instructions indicator at bottom
	instructY := panelY + panelHeight - 35
	mr.uiRenderer.DrawText(panelX+20, instructY, 1.5, "Use ARROWS to select, ENTER to confirm", [4]float32{0.7, 0.7, 0.7, 1.0})
}
