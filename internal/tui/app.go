package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"

	"github.com/jparkerweb/shrinkray/internal/config"
	"github.com/jparkerweb/shrinkray/internal/engine"
	"github.com/jparkerweb/shrinkray/internal/presets"
	"github.com/jparkerweb/shrinkray/internal/tui/messages"
	"github.com/jparkerweb/shrinkray/internal/tui/screens"
	"github.com/jparkerweb/shrinkray/internal/tui/style"
)

// App is the top-level Bubble Tea model for the shrinkray TUI.
type App struct {
	currentScreen messages.Screen
	screenHistory []messages.Screen
	screenModels  map[messages.Screen]ScreenModel
	width         int
	height        int
	keyMap        KeyMap
	showHelp      bool // whether the help overlay is visible

	// Shared state passed between screens
	videoInfo      *engine.VideoInfo
	selectedPreset *presets.Preset
	encodeOpts     *engine.EncodeOptions
	inputPath      string
	encodeStart    time.Time
	hwEncoders     []engine.HWEncoder
	batchQueue     *engine.JobQueue
	batchStart     time.Time
}

// AppOptions holds configuration for creating a new App.
type AppOptions struct {
	InputPath string
	VideoInfo *engine.VideoInfo
}

// NewApp creates a new App model.
func NewApp(opts AppOptions) App {
	app := App{
		currentScreen: messages.ScreenSplash,
		keyMap:        DefaultKeyMap(),
		screenModels:  make(map[messages.Screen]ScreenModel),
		inputPath:     opts.InputPath,
		videoInfo:     opts.VideoInfo,
	}

	// Initialize all screen models
	app.screenModels[messages.ScreenSplash] = screens.NewSplashModel()
	app.screenModels[messages.ScreenFilePicker] = screens.NewFilePickerModel()
	app.screenModels[messages.ScreenInfo] = screens.NewInfoModel()
	app.screenModels[messages.ScreenPresets] = screens.NewPresetsModel()
	app.screenModels[messages.ScreenAdvanced] = screens.NewAdvancedModel()
	app.screenModels[messages.ScreenPreview] = screens.NewPreviewModel()
	app.screenModels[messages.ScreenEncoding] = screens.NewEncodingModel()
	app.screenModels[messages.ScreenComplete] = screens.NewCompleteModel()
	app.screenModels[messages.ScreenBatchQueue] = screens.NewBatchQueueModel()
	app.screenModels[messages.ScreenBatchProgress] = screens.NewBatchProgressModel()
	app.screenModels[messages.ScreenBatchComplete] = screens.NewBatchCompleteModel()

	// If a file was pre-selected, set video info on relevant screens
	if opts.VideoInfo != nil {
		app.videoInfo = opts.VideoInfo
		app.setVideoInfoOnScreens()
	}

	return app
}

// Init initializes the app and the splash screen.
func (a App) Init() tea.Cmd {
	// If we have pre-loaded video info, skip splash and go to info
	if a.videoInfo != nil {
		return tea.Batch(
			a.screenModels[messages.ScreenSplash].Init(),
			func() tea.Msg {
				return messages.NavigateMsg{Screen: messages.ScreenInfo}
			},
		)
	}
	return a.screenModels[messages.ScreenSplash].Init()
}

// Update handles messages for the app.
func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// Propagate to all screens
		var cmds []tea.Cmd
		for screen, model := range a.screenModels {
			updated, cmd := model.Update(msg)
			a.screenModels[screen] = updated
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return a, tea.Batch(cmds...)

	case messages.NavigateMsg:
		return a.navigateTo(msg.Screen)

	case messages.BackMsg:
		return a.goBack()

	case messages.FileSelectedMsg:
		a.inputPath = msg.Path
		return a, nil

	case messages.VideoProbeCompleteMsg:
		if msg.Err != nil {
			// Pass error to current screen
			updated, cmd := a.screenModels[a.currentScreen].Update(msg)
			a.screenModels[a.currentScreen] = updated
			return a, cmd
		}
		a.videoInfo = msg.Info
		a.setVideoInfoOnScreens()
		return a.navigateTo(messages.ScreenInfo)

	case messages.PresetSelectedMsg:
		a.selectedPreset = &msg.Preset
		// Update preview screen with the new preset
		if preview, ok := a.screenModels[messages.ScreenPreview].(screens.PreviewModel); ok {
			preview.SetPreset(a.selectedPreset)
			if len(a.hwEncoders) > 0 {
				preview.SetHWEncoders(a.hwEncoders)
			}
			a.screenModels[messages.ScreenPreview] = preview
		}
		// Update advanced screen with the new preset
		if adv, ok := a.screenModels[messages.ScreenAdvanced].(screens.AdvancedModel); ok {
			adv.SetPreset(a.selectedPreset)
			if len(a.hwEncoders) > 0 {
				adv.SetHWEncoders(a.hwEncoders)
			}
			a.screenModels[messages.ScreenAdvanced] = adv
		}
		return a.navigateTo(messages.ScreenPreview)

	case messages.HWDetectedMsg:
		a.hwEncoders = msg.Encoders
		// Propagate to preview screen
		if preview, ok := a.screenModels[messages.ScreenPreview].(screens.PreviewModel); ok {
			preview.SetHWEncoders(msg.Encoders)
			a.screenModels[messages.ScreenPreview] = preview
		}
		// Propagate to advanced screen
		if adv, ok := a.screenModels[messages.ScreenAdvanced].(screens.AdvancedModel); ok {
			adv.SetHWEncoders(msg.Encoders)
			a.screenModels[messages.ScreenAdvanced] = adv
		}
		// Also let the splash screen handle it
		if model, ok := a.screenModels[a.currentScreen]; ok {
			updated, cmd := model.Update(msg)
			a.screenModels[a.currentScreen] = updated
			return a, cmd
		}
		return a, nil

	case messages.AdvancedOptionsMsg:
		// Apply advanced options and go back to preview
		a.applyAdvancedOptions(msg.Opts)
		return a.goBack()

	case messages.EncodeStartMsg:
		a.encodeOpts = &msg.Opts
		a.encodeStart = time.Now()
		nav, navCmd := a.navigateTo(messages.ScreenEncoding)
		app := nav.(App)

		// Start encoding
		if enc, ok := app.screenModels[messages.ScreenEncoding].(screens.EncodingModel); ok {
			startCmd := enc.StartEncode(msg.Opts)
			app.screenModels[messages.ScreenEncoding] = enc
			return app, tea.Batch(navCmd, startCmd)
		}
		return app, navCmd

	case messages.EncodeCompleteMsg:
		elapsed := time.Since(a.encodeStart)
		inputSize := msg.InputSize
		if inputSize == 0 && a.videoInfo != nil {
			inputSize = a.videoInfo.Size
		}
		if complete, ok := a.screenModels[messages.ScreenComplete].(screens.CompleteModel); ok {
			complete.SetResults(msg.OutputPath, inputSize, msg.OutputSize, elapsed)
			a.screenModels[messages.ScreenComplete] = complete
		}
		return a.navigateTo(messages.ScreenComplete)

	case messages.EncodeCancelMsg:
		return a.navigateTo(messages.ScreenPreview)

	case messages.FilesSelectedMsg:
		// Multiple files selected — go to batch queue
		presetKey := "balanced"
		if a.selectedPreset != nil {
			presetKey = a.selectedPreset.Key
		}
		nav, navCmd := a.navigateTo(messages.ScreenBatchQueue)
		app := nav.(App)
		probeCmd := screens.ProbeAndAddFiles(msg.Paths, presetKey)
		return app, tea.Batch(navCmd, probeCmd)

	case messages.BatchQueueReadyMsg:
		a.batchQueue = msg.Queue
		if bq, ok := a.screenModels[messages.ScreenBatchQueue].(screens.BatchQueueModel); ok {
			bq.SetQueue(msg.Queue)
			if a.selectedPreset != nil {
				bq.SetPreset(a.selectedPreset)
			}
			a.screenModels[messages.ScreenBatchQueue] = bq
		}
		return a, nil

	case messages.BatchStartMsg:
		a.batchQueue = msg.Queue
		a.batchStart = time.Now()
		nav, navCmd := a.navigateTo(messages.ScreenBatchProgress)
		app := nav.(App)
		if bp, ok := app.screenModels[messages.ScreenBatchProgress].(screens.BatchProgressModel); ok {
			preset := presets.Preset{}
			if a.selectedPreset != nil {
				preset = *a.selectedPreset
			} else {
				// Default
				if p, found := presets.Lookup("balanced"); found {
					preset = p
				}
			}
			bp.SetOptions(preset, "", engine.OutputOptions{
				Mode:   engine.OutputModeSuffix,
				Suffix: "_shrunk",
			}, engine.SkipOptions{}, 1, 2)
			startCmd := bp.StartBatch(msg.Queue)
			app.screenModels[messages.ScreenBatchProgress] = bp
			return app, tea.Batch(navCmd, startCmd)
		}
		return app, navCmd

	case messages.BatchEventMsg:
		// Delegate to batch progress screen
		if model, ok := a.screenModels[messages.ScreenBatchProgress]; ok {
			updated, cmd := model.Update(msg)
			a.screenModels[messages.ScreenBatchProgress] = updated
			return a, cmd
		}
		return a, nil

	case messages.ThemeToggleMsg:
		// Toggle theme and persist to config
		newTheme := style.ToggleTheme()
		go func() {
			cfg, err := config.Load("")
			if err == nil {
				cfg.UI.Theme = string(newTheme)
				_ = cfg.Save()
			}
		}()
		return a, nil

	case tea.KeyPressMsg:
		// If help overlay is showing, dismiss it on ?, Esc, or navigation keys
		if a.showHelp {
			switch msg.String() {
			case "?", "esc", "enter", "up", "down", "left", "right", "tab":
				a.showHelp = false
				return a, nil
			}
			return a, nil
		}

		// Global key bindings
		switch {
		case key.Matches(msg, a.keyMap.Help):
			a.showHelp = true
			return a, nil
		case key.Matches(msg, a.keyMap.Quit):
			// On 'q', let the complete screen handle it (it has q=quit)
			if msg.String() == "q" && a.currentScreen == messages.ScreenComplete {
				break
			}
			if msg.String() == "q" && a.currentScreen == messages.ScreenBatchComplete {
				break
			}
			// ctrl+c always quits
			if msg.String() != "q" {
				return a, tea.Quit
			}
		case key.Matches(msg, a.keyMap.ThemeToggle):
			return a, func() tea.Msg { return messages.ThemeToggleMsg{} }
		}
	}

	// Delegate to current screen
	if model, ok := a.screenModels[a.currentScreen]; ok {
		updated, cmd := model.Update(msg)
		a.screenModels[a.currentScreen] = updated
		return a, cmd
	}

	return a, nil
}

// applyAdvancedOptions applies the user's advanced choices to the selected preset.
func (a *App) applyAdvancedOptions(opts messages.AdvancedOptions) {
	if a.selectedPreset == nil {
		return
	}

	// Apply overrides to preset
	a.selectedPreset.Codec = opts.Codec
	a.selectedPreset.CRF = opts.CRF
	if opts.Resolution != "" {
		a.selectedPreset.Resolution = opts.Resolution
	}
	if opts.MaxFPS > 0 {
		a.selectedPreset.MaxFPS = opts.MaxFPS
	}
	a.selectedPreset.AudioCodec = opts.AudioCodec
	a.selectedPreset.AudioBitrate = opts.AudioBitrate
	if opts.AudioChannels > 0 {
		a.selectedPreset.AudioChannels = opts.AudioChannels
	}

	// VP9 forces webm container
	if opts.Codec == "vp9" {
		a.selectedPreset.Container = "webm"
	}

	// Update preview screen
	if preview, ok := a.screenModels[messages.ScreenPreview].(screens.PreviewModel); ok {
		preview.SetPreset(a.selectedPreset)
		// Set HW encoder on preview if specified
		if opts.HWEncoderName != "" {
			hwEncName := engine.HWEncoderName(opts.HWEncoderName, opts.Codec)
			if hwEncName != "" {
				preview.SetHWEncoders(a.hwEncoders)
			}
		}
		a.screenModels[messages.ScreenPreview] = preview
	}
}

// View renders the app.
func (a App) View() tea.View {
	var b strings.Builder

	// Header bar
	header := a.renderHeader()
	b.WriteString(header)
	b.WriteString("\n")

	// Current screen
	if model, ok := a.screenModels[a.currentScreen]; ok {
		b.WriteString(model.View())
	}

	// Footer bar
	b.WriteString("\n")
	footer := a.renderFooter()
	b.WriteString(footer)

	content := b.String()

	// Help overlay
	if a.showHelp {
		content = a.renderHelpOverlay(content)
	}

	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// renderHelpOverlay renders a centered help box on top of the current screen.
func (a App) renderHelpOverlay(base string) string {
	// Build help content
	var help strings.Builder

	help.WriteString(style.TitleStyle().Render("Keyboard Shortcuts"))
	help.WriteString("\n\n")

	// Global bindings
	help.WriteString(style.AccentStyle().Render("Global"))
	help.WriteString("\n")
	globalBindings := [][2]string{
		{"Ctrl+C", "Quit"},
		{"Ctrl+T", "Toggle theme"},
		{"?", "Show/hide help"},
		{"Esc", "Go back"},
	}
	for _, bind := range globalBindings {
		fmt.Fprintf(&help, "  %s%s%s\n",
			style.KeyHintStyle().Render(fmt.Sprintf("%-12s", bind[0])),
			"  ",
			bind[1])
	}

	// Context-sensitive bindings
	help.WriteString("\n")
	help.WriteString(style.AccentStyle().Render(a.currentScreen.String()))
	help.WriteString("\n")

	screenBindings := a.screenBindings()
	for _, bind := range screenBindings {
		fmt.Fprintf(&help, "  %s%s%s\n",
			style.KeyHintStyle().Render(fmt.Sprintf("%-12s", bind[0])),
			"  ",
			bind[1])
	}

	help.WriteString("\n")
	help.WriteString(style.MutedStyle().Render("Press ? or Esc to close"))

	helpText := help.String()

	// Render as a centered bordered box
	boxWidth := 44
	if a.width > 0 && boxWidth > a.width-4 {
		boxWidth = a.width - 4
	}

	box := style.CardStyle().
		Width(boxWidth).
		Render(helpText)

	// Center the box over the base content
	lines := strings.Split(base, "\n")
	boxLines := strings.Split(box, "\n")

	// Calculate vertical position (centered)
	startRow := (len(lines) - len(boxLines)) / 2
	if startRow < 1 {
		startRow = 1
	}

	// Calculate horizontal padding for centering
	padLeft := 0
	if a.width > boxWidth+4 {
		padLeft = (a.width - boxWidth - 4) / 2
	}
	leftPad := strings.Repeat(" ", padLeft)

	// Overlay the box onto the base
	for i, boxLine := range boxLines {
		row := startRow + i
		if row < len(lines) {
			lines[row] = leftPad + boxLine
		}
	}

	return strings.Join(lines, "\n")
}

// screenBindings returns the key bindings specific to the current screen.
func (a App) screenBindings() [][2]string {
	switch a.currentScreen {
	case messages.ScreenSplash:
		return [][2]string{
			{"Any key", "Continue"},
		}
	case messages.ScreenFilePicker:
		return [][2]string{
			{"Tab", "Toggle input mode"},
			{"Enter", "Select file"},
			{"Esc", "Go back"},
		}
	case messages.ScreenInfo:
		return [][2]string{
			{"Enter", "Choose preset"},
		}
	case messages.ScreenPresets:
		return [][2]string{
			{"Arrows", "Navigate presets"},
			{"Enter", "Select preset"},
		}
	case messages.ScreenAdvanced:
		return [][2]string{
			{"Arrows", "Change values"},
			{"Enter", "Apply settings"},
			{"Esc", "Cancel"},
		}
	case messages.ScreenPreview:
		return [][2]string{
			{"Enter", "Start encoding"},
			{"a", "Advanced options"},
		}
	case messages.ScreenEncoding:
		return [][2]string{
			{"Esc/c", "Cancel encoding"},
		}
	case messages.ScreenComplete:
		return [][2]string{
			{"o", "Open output folder"},
			{"r", "Re-encode"},
			{"n", "New file"},
			{"q", "Quit"},
		}
	case messages.ScreenBatchQueue:
		return [][2]string{
			{"Enter", "Start batch"},
			{"d", "Remove selected"},
			{"Shift+Up/Down", "Reorder"},
		}
	case messages.ScreenBatchProgress:
		return [][2]string{
			{"Esc", "Cancel batch"},
		}
	case messages.ScreenBatchComplete:
		return [][2]string{
			{"r", "Retry failed"},
			{"o", "Open output folder"},
			{"n", "New batch"},
			{"q", "Quit"},
		}
	default:
		return nil
	}
}

func (a App) renderHeader() string {
	left := HeaderStyle().Render(" shrinkray ")
	right := HeaderStyle().Render(fmt.Sprintf(" %s ", a.currentScreen.String()))

	gap := a.width - visibleLen(left) - visibleLen(right)
	if gap < 0 {
		gap = 0
	}
	fill := HeaderStyle().Render(strings.Repeat(" ", gap))

	return left + fill + right
}

func (a App) renderFooter() string {
	hints := a.footerHints()
	return FooterStyle().Width(a.width).Render(hints)
}

func (a App) footerHints() string {
	switch a.currentScreen {
	case messages.ScreenSplash:
		return "Press any key to continue"
	case messages.ScreenFilePicker:
		return "Tab: toggle input | Enter: select | Esc: back | Ctrl+C: quit"
	case messages.ScreenInfo:
		return "Enter: choose preset | Esc: back | Ctrl+C: quit"
	case messages.ScreenPresets:
		return "Arrows: navigate | Enter: select | Esc: back | Ctrl+C: quit"
	case messages.ScreenAdvanced:
		return "Arrows: change values | Enter: apply | Esc: cancel | Ctrl+C: quit"
	case messages.ScreenPreview:
		return "Enter: start | a: advanced | Esc: back | Ctrl+C: quit"
	case messages.ScreenEncoding:
		return "Esc/c: cancel | Ctrl+C: quit"
	case messages.ScreenComplete:
		return "o: open folder | r: re-encode | n: new file | q: quit"
	case messages.ScreenBatchQueue:
		return "Enter: start | d: remove | Shift+Up/Down: reorder | Esc: back"
	case messages.ScreenBatchProgress:
		return "Esc: cancel | Ctrl+C: quit"
	case messages.ScreenBatchComplete:
		return "r: retry failed | o: open folder | n: new batch | q: quit"
	default:
		return "Ctrl+C: quit"
	}
}

func (a App) navigateTo(screen messages.Screen) (tea.Model, tea.Cmd) {
	a.screenHistory = append(a.screenHistory, a.currentScreen)
	a.currentScreen = screen

	var cmds []tea.Cmd

	// Propagate current dimensions
	if a.width > 0 && a.height > 0 {
		sizeMsg := tea.WindowSizeMsg{Width: a.width, Height: a.height}
		if model, ok := a.screenModels[screen]; ok {
			updated, cmd := model.Update(sizeMsg)
			a.screenModels[screen] = updated
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	// Set batch results when navigating to batch complete
	if screen == messages.ScreenBatchComplete && a.batchQueue != nil {
		if bc, ok := a.screenModels[messages.ScreenBatchComplete].(screens.BatchCompleteModel); ok {
			bc.SetResults(a.batchQueue, time.Since(a.batchStart))
			a.screenModels[messages.ScreenBatchComplete] = bc
		}
	}

	// Call Init on the new screen
	if model, ok := a.screenModels[screen]; ok {
		initCmd := model.Init()
		if initCmd != nil {
			cmds = append(cmds, initCmd)
		}
	}

	return a, tea.Batch(cmds...)
}

func (a App) goBack() (tea.Model, tea.Cmd) {
	if len(a.screenHistory) == 0 {
		return a, nil
	}

	prev := a.screenHistory[len(a.screenHistory)-1]
	a.screenHistory = a.screenHistory[:len(a.screenHistory)-1]
	a.currentScreen = prev

	return a, nil
}

func (a App) setVideoInfoOnScreens() {
	// Info screen
	if info, ok := a.screenModels[messages.ScreenInfo].(screens.InfoModel); ok {
		info.SetVideoInfo(a.videoInfo)
		a.screenModels[messages.ScreenInfo] = info
	}

	// Presets screen
	if ps, ok := a.screenModels[messages.ScreenPresets].(screens.PresetsModel); ok {
		ps.SetVideoInfo(a.videoInfo)
		a.screenModels[messages.ScreenPresets] = ps
	}

	// Preview screen
	if pv, ok := a.screenModels[messages.ScreenPreview].(screens.PreviewModel); ok {
		pv.SetVideoInfo(a.videoInfo)
		a.screenModels[messages.ScreenPreview] = pv
	}

	// Advanced screen
	if adv, ok := a.screenModels[messages.ScreenAdvanced].(screens.AdvancedModel); ok {
		adv.SetVideoInfo(a.videoInfo)
		a.screenModels[messages.ScreenAdvanced] = adv
	}
}

// visibleLen returns the visible length of a string (stripping ANSI codes).
func visibleLen(s string) int {
	inEscape := false
	n := 0
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		n++
	}
	return n
}
