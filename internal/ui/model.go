package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/dnlvgl/zap/internal/container"
	"github.com/dnlvgl/zap/internal/kill"
	"github.com/dnlvgl/zap/internal/port"
	"github.com/dnlvgl/zap/internal/process"
)

type state int

const (
	stateLoading state = iota
	stateList
	stateConfirm
	stateResult
)

const autoRefreshInterval = 2 * time.Second

type tickMsg time.Time

type processItem struct {
	listener port.Listener
	context  process.Context
}

// Model is the Bubble Tea model for the zap TUI.
type Model struct {
	state       state
	query       *port.Query // nil means show all ports
	items       []processItem
	cursor      int
	selectedPID int // PID of selected row — used to restore cursor after refresh
	force       bool
	message     string
	isError     bool
	width       int
	height      int
	quitting    bool
	search      string
}

// visibleItems returns the filtered subset of items matching m.search.
func (m Model) visibleItems() []processItem {
	if m.search == "" {
		return m.items
	}
	var out []processItem
	for _, item := range m.items {
		if strings.Contains(strconv.Itoa(item.listener.Port), m.search) {
			out = append(out, item)
		}
	}
	return out
}

// New creates a new TUI model.
func New(query *port.Query, force bool) Model {
	return Model{
		state: stateLoading,
		query: query,
		force: force,
	}
}

// Messages

type loadedMsg struct {
	items []processItem
	err   error
}

type killResultMsg struct {
	desc string
	err  error
}

// Commands

func tickCmd() tea.Cmd {
	return tea.Tick(autoRefreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func loadProcesses(query *port.Query) tea.Cmd {
	return func() tea.Msg {
		var listeners []port.Listener
		var err error

		if query == nil {
			listeners, err = port.DetectAll()
		} else {
			listeners, err = port.Detect(*query)
		}
		if err != nil {
			return loadedMsg{err: err}
		}

		// Deduplicate by PID and gather context
		seen := make(map[int]bool)
		var items []processItem
		for _, l := range listeners {
			if seen[l.PID] {
				continue
			}
			seen[l.PID] = true

			ctx, err := process.GatherContext(l.PID, l.Port)
			if err != nil {
				continue
			}
			items = append(items, processItem{
				listener: l,
				context:  ctx,
			})
		}

		return loadedMsg{items: items}
	}
}

func executeKill(item processItem, force bool) tea.Cmd {
	return func() tea.Msg {
		strategy := kill.RecommendedStrategy(item.context)
		action := kill.Action{
			Strategy: strategy,
			Context:  item.context,
			Force:    force,
		}
		desc := kill.Describe(action)
		err := kill.Execute(action)
		return killResultMsg{desc: desc, err: err}
	}
}

// Init starts the initial loading.
func (m Model) Init() tea.Cmd {
	return tea.Batch(loadProcesses(m.query), tickCmd())
}

// Update handles events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tickMsg:
		if m.state == stateList {
			visible := m.visibleItems()
			if m.cursor < len(visible) {
				m.selectedPID = visible[m.cursor].context.Info.PID
			}
			return m, tea.Batch(loadProcesses(m.query), tickCmd())
		}
		return m, tickCmd()

	case loadedMsg:
		// Don't disrupt an active confirm dialog
		if m.state == stateConfirm {
			return m, nil
		}
		if msg.err != nil {
			m.state = stateResult
			m.message = msg.err.Error()
			m.isError = true
			return m, nil
		}
		if len(msg.items) == 0 {
			m.state = stateResult
			m.message = "No processes found."
			m.isError = false
			return m, nil
		}
		m.items = msg.items
		m.state = stateList
		// Restore cursor by PID within visible (filtered) items; fall back to first
		visible := m.visibleItems()
		if m.selectedPID != 0 {
			m.cursor = 0
			for i, item := range visible {
				if item.context.Info.PID == m.selectedPID {
					m.cursor = i
					break
				}
			}
		}
		if m.cursor >= len(visible) {
			m.cursor = max(0, len(visible)-1)
		}
		// Keep selectedPID in sync with actual cursor
		if m.cursor < len(visible) {
			m.selectedPID = visible[m.cursor].context.Info.PID
		}
		return m, nil

	case killResultMsg:
		m.state = stateResult
		if msg.err != nil {
			m.message = fmt.Sprintf("Failed: %s — %v", msg.desc, msg.err)
			m.isError = true
		} else {
			m.message = fmt.Sprintf("Done: %s", msg.desc)
			m.isError = false
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateList:
		switch msg.String() {
		case "ctrl+g", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			if m.search != "" {
				m.search = ""
				m.cursor = 0
			} else {
				m.quitting = true
				return m, tea.Quit
			}
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			m.search += msg.String()
			m.cursor = 0
		case "backspace":
			if len(m.search) > 0 {
				m.search = m.search[:len(m.search)-1]
				m.cursor = 0
			}
		case "up", "ctrl+p":
			visible := m.visibleItems()
			if m.cursor > 0 {
				m.cursor--
			}
			if m.cursor < len(visible) {
				m.selectedPID = visible[m.cursor].context.Info.PID
			}
		case "down", "ctrl+n":
			visible := m.visibleItems()
			if m.cursor < len(visible)-1 {
				m.cursor++
			}
			if m.cursor < len(visible) {
				m.selectedPID = visible[m.cursor].context.Info.PID
			}
		case "enter", " ":
			if m.cursor < len(m.visibleItems()) {
				m.state = stateConfirm
			}
		case "ctrl+r":
			visible := m.visibleItems()
			if m.cursor < len(visible) {
				m.selectedPID = visible[m.cursor].context.Info.PID
			}
			m.state = stateLoading
			return m, loadProcesses(m.query)
		}

	case stateConfirm:
		switch msg.String() {
		case "y", "Y", "enter":
			item := m.visibleItems()[m.cursor]
			m.state = stateLoading
			m.message = "Killing..."
			return m, executeKill(item, m.force)
		case "n", "N", "esc", "ctrl+g":
			m.state = stateList
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}

	case stateResult:
		switch msg.String() {
		case "ctrl+g", "ctrl+c", "esc", "enter":
			m.quitting = true
			return m, tea.Quit
		case "ctrl+b":
			m.state = stateLoading
			m.cursor = 0
			return m, loadProcesses(m.query)
		}
	}

	return m, nil
}

// View renders the UI.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	switch m.state {
	case stateLoading:
		return m.viewLoading()
	case stateList:
		return m.viewList()
	case stateConfirm:
		return m.viewConfirm()
	case stateResult:
		return m.viewResult()
	}
	return ""
}

func (m Model) viewLoading() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		"  Scanning ports...",
		"",
	)
}

func (m Model) viewList() string {
	title := m.buildTitle()
	search := m.buildSearchBar()
	tbl := m.buildTable()
	detail := m.buildDetailPanel()
	help := m.buildHelp()

	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(title),
		search,
		tbl,
		detail,
		help,
		"",
	)
}

func (m Model) viewConfirm() string {
	title := m.buildTitle()
	search := m.buildSearchBar()
	tbl := m.buildTable()
	confirm := m.buildConfirmPrompt()
	help := helpStyle.Render("y/enter confirm • n/esc cancel")

	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(title),
		search,
		tbl,
		confirm,
		help,
		"",
	)
}

func (m Model) viewResult() string {
	var msg string
	if m.isError {
		msg = errorStyle.Render("  " + m.message)
	} else {
		msg = successStyle.Render("  " + m.message)
	}
	help := helpStyle.Render("  C-b go back • C-g/enter quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		"",
		msg,
		help,
		"",
	)
}

// buildTitle returns the title string based on query.
func (m Model) buildTitle() string {
	if m.query != nil {
		if m.query.IsSinglePort() {
			return fmt.Sprintf("Processes on port %d", m.query.StartPort)
		}
		return fmt.Sprintf("Processes on ports %d-%d", m.query.StartPort, m.query.EndPort)
	}
	return "Listening Processes"
}

// buildSearchBar renders a persistent filter input line.
func (m Model) buildSearchBar() string {
	prompt := searchPromptStyle.Render("/ ")
	if m.search == "" {
		return prompt + searchPlaceholderStyle.Render("type digits to filter by port")
	}
	return prompt + searchStyle.Render(m.search+"█")
}

// buildHelp returns the help line.
func (m Model) buildHelp() string {
	help := "C-p/C-n navigate • enter select • C-r refresh • auto • C-g quit"
	if m.force {
		help += " • FORCE mode"
	}
	return helpStyle.Render(help)
}

// buildTable constructs a lipgloss table from the process items.
func (m Model) buildTable() string {
	width := m.width
	if width == 0 {
		width = 80
	}

	visible := m.visibleItems()
	rows := make([][]string, len(visible))
	for i, item := range visible {
		rows[i] = m.buildRow(i, item, width)
	}

	t := table.New().
		Headers("", "PORT", "PID", "COMMAND").
		Rows(rows...).
		Border(lipgloss.NormalBorder()).
		BorderHeader(true).
		BorderColumn(false).
		BorderRow(false).
		Width(width).
		StyleFunc(m.tableStyleFunc)

	return t.Render()
}

// tableStyleFunc returns the style for each cell based on row/col.
func (m Model) tableStyleFunc(row, col int) lipgloss.Style {
	if row == table.HeaderRow {
		s := tableHeaderStyle
		if col == 0 {
			return s.Width(2)
		}
		return s
	}

	if row == m.cursor {
		s := tableSelectedStyle
		if col == 0 {
			return s.Width(2)
		}
		return s
	}

	s := tableCellStyle
	if col == 0 {
		return s.Width(2)
	}
	switch col {
	case 1: // PORT
		return s.Foreground(colorAccent)
	case 2: // PID
		return s.Foreground(colorYellow)
	case 3: // COMMAND
		return s.Foreground(colorSubtle)
	}
	return s
}

// buildRow returns a row for the table.
func (m Model) buildRow(index int, item processItem, width int) []string {
	sel := " "
	if index == m.cursor {
		sel = ">"
	}

	portStr := fmt.Sprintf(":%d/%s", item.listener.Port, item.listener.Protocol)
	pidStr := strconv.Itoa(item.context.Info.PID)

	cmd := item.context.Info.Command
	// Reserve space for selector(2) + port(~12) + pid(~8) + borders/padding(~10)
	maxCmd := width - 32
	if maxCmd < 20 {
		maxCmd = 20
	}
	if len(cmd) > maxCmd && maxCmd > 3 {
		cmd = cmd[:maxCmd-3] + "..."
	}

	return []string{sel, portStr, pidStr, cmd}
}

// buildDetailPanel renders the detail panel for the selected item.
func (m Model) buildDetailPanel() string {
	visible := m.visibleItems()
	if m.cursor >= len(visible) {
		return ""
	}
	item := visible[m.cursor]
	info := item.context.Info

	var lines []string

	// User
	if info.User != "" {
		lines = append(lines, detailLabelStyle.Render("User")+detailValueStyle.Render(info.User))
	}

	// Memory
	if info.MemoryKB > 0 {
		var memStr string
		if info.MemoryKB > 1024 {
			memStr = fmt.Sprintf("%d MB", info.MemoryKB/1024)
		} else {
			memStr = fmt.Sprintf("%d KB", info.MemoryKB)
		}
		lines = append(lines, detailLabelStyle.Render("Memory")+detailValueStyle.Render(memStr))
	}

	// Uptime
	if uptime := info.Uptime(); uptime > 0 {
		lines = append(lines, detailLabelStyle.Render("Uptime")+detailValueStyle.Render(formatDuration(uptime)))
	}

	// Children
	if len(info.Children) > 0 {
		lines = append(lines, detailLabelStyle.Render("Children")+detailValueStyle.Render(strconv.Itoa(len(info.Children))))
	}

	// Kill strategy
	strategy := kill.RecommendedStrategy(item.context)
	action := kill.Action{
		Strategy: strategy,
		Context:  item.context,
		Force:    m.force,
	}
	desc := kill.Describe(action)
	lines = append(lines, detailLabelStyle.Render("Action")+strategyStyle.Render(desc))

	// Warnings
	var warnings []string
	if info.IsPrivileged() {
		warnings = append(warnings, "needs sudo")
	}
	if len(info.Children) > 0 {
		warnings = append(warnings, fmt.Sprintf("%d children affected", len(info.Children)))
	}
	if m.force {
		warnings = append(warnings, "FORCE mode")
	}
	if len(warnings) > 0 {
		lines = append(lines, detailLabelStyle.Render("Warning")+warningStyle.Render(strings.Join(warnings, ", ")))
	}

	// Tags
	var tags []string
	if item.context.IsContainerized() {
		name := item.context.Container.Name
		if name == "" {
			name = container.ShortID(item.context.Container.ID)
		}
		tags = append(tags, tagContainerStyle.Render(fmt.Sprintf("%s:%s", item.context.Container.Runtime, name)))
	}
	if item.context.IsSystemdManaged() {
		tags = append(tags, tagSystemdStyle.Render(item.context.SystemdUnit))
	}
	if info.IsPrivileged() {
		tags = append(tags, tagSudoStyle.Render("sudo"))
	}
	if len(tags) > 0 {
		lines = append(lines, detailLabelStyle.Render("")+strings.Join(tags, " "))
	}

	content := strings.Join(lines, "\n")
	const detailPanelLines = 7
	width := m.width
	if width > 0 {
		// Account for border (2 chars) and padding (2 chars)
		return detailPanelStyle.Width(width - 4).Height(detailPanelLines).Render(content)
	}
	return detailPanelStyle.Height(detailPanelLines).Render(content)
}

// buildConfirmPrompt renders the inline confirm prompt.
func (m Model) buildConfirmPrompt() string {
	visible := m.visibleItems()
	if m.cursor >= len(visible) {
		return ""
	}
	item := visible[m.cursor]
	strategy := kill.RecommendedStrategy(item.context)
	action := kill.Action{
		Strategy: strategy,
		Context:  item.context,
		Force:    m.force,
	}
	desc := kill.Describe(action)

	var lines []string
	lines = append(lines, confirmPromptStyle.Render("Kill? ")+confirmDescStyle.Render(desc+" [y/n]"))

	if len(item.context.Info.Children) > 0 {
		lines = append(lines, warningStyle.Render(
			fmt.Sprintf("Warning: %d child processes will be affected", len(item.context.Info.Children)),
		))
	}

	content := strings.Join(lines, "\n")
	width := m.width
	if width > 0 {
		return confirmPanelStyle.Width(width - 4).Render(content)
	}
	return confirmPanelStyle.Render(content)
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	if h >= 24 {
		return fmt.Sprintf("%dd", h/24)
	}
	if h > 0 {
		return fmt.Sprintf("%dh", h)
	}
	m := int(d.Minutes())
	if m > 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}
