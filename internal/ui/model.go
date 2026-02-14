package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dnl/zap/internal/container"
	"github.com/dnl/zap/internal/kill"
	"github.com/dnl/zap/internal/port"
	"github.com/dnl/zap/internal/process"
)

type state int

const (
	stateLoading state = iota
	stateList
	stateConfirm
	stateResult
)

type processItem struct {
	listener port.Listener
	context  process.Context
}

// Model is the Bubble Tea model for the zap TUI.
type Model struct {
	state     state
	query     *port.Query // nil means show all ports
	items     []processItem
	cursor    int
	force     bool
	message   string
	isError   bool
	width     int
	height    int
	quitting  bool
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

			ctx, err := process.GatherContext(l.PID)
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
	return loadProcesses(m.query)
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

	case loadedMsg:
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
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.state = stateConfirm
		case "r":
			m.state = stateLoading
			return m, loadProcesses(m.query)
		}

	case stateConfirm:
		switch msg.String() {
		case "y", "Y", "enter":
			item := m.items[m.cursor]
			m.state = stateLoading
			m.message = "Killing..."
			return m, executeKill(item, m.force)
		case "n", "N", "esc", "q":
			m.state = stateList
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}

	case stateResult:
		switch msg.String() {
		case "q", "ctrl+c", "esc", "enter":
			m.quitting = true
			return m, tea.Quit
		case "r":
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
	return "\n  Scanning ports...\n"
}

func (m Model) viewList() string {
	var b strings.Builder

	title := "Listening processes"
	if m.query != nil {
		if m.query.IsSinglePort() {
			title = fmt.Sprintf("Processes on port %d", m.query.StartPort)
		} else {
			title = fmt.Sprintf("Processes on ports %d-%d", m.query.StartPort, m.query.EndPort)
		}
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n")

	for i, item := range m.items {
		b.WriteString(m.renderItem(i, item))
		b.WriteString("\n")
	}

	help := "j/k navigate • enter select • r refresh • q quit"
	if m.force {
		help += " • FORCE mode"
	}
	b.WriteString(helpStyle.Render(help))
	b.WriteString("\n")

	return b.String()
}

func (m Model) renderItem(index int, item processItem) string {
	cursor := unselectedStyle.Render(" ")
	if index == m.cursor {
		cursor = selectedStyle.Render("")
	}

	pid := pidStyle.Render(fmt.Sprintf("PID %d", item.context.Info.PID))
	portStr := portStyle.Render(fmt.Sprintf(":%d/%s", item.listener.Port, item.listener.Protocol))

	cmd := item.context.Info.Command
	if len(cmd) > 50 {
		cmd = cmd[:47] + "..."
	}
	cmdStr := commandStyle.Render(cmd)

	var tags []string
	if item.context.IsContainerized() {
		name := item.context.Container.Name
		if name == "" {
			name = container.ShortID(item.context.Container.ID)
		}
		tag := tagContainerStyle.Render(fmt.Sprintf("%s:%s", item.context.Container.Runtime, name))
		tags = append(tags, tag)
	}
	if item.context.IsSystemdManaged() {
		tag := tagSystemdStyle.Render(item.context.SystemdUnit)
		tags = append(tags, tag)
	}

	line := fmt.Sprintf("%s %s %s  %s", cursor, pid, portStr, cmdStr)
	if len(tags) > 0 {
		line += "  " + strings.Join(tags, " ")
	}

	// Second line with details
	var details []string
	if item.context.Info.User != "" {
		details = append(details, fmt.Sprintf("user:%s", item.context.Info.User))
	}
	if item.context.Info.MemoryKB > 0 {
		mem := item.context.Info.MemoryKB
		if mem > 1024 {
			details = append(details, fmt.Sprintf("mem:%dMB", mem/1024))
		} else {
			details = append(details, fmt.Sprintf("mem:%dKB", mem))
		}
	}
	if uptime := item.context.Info.Uptime(); uptime > 0 {
		details = append(details, fmt.Sprintf("up:%s", formatDuration(uptime)))
	}
	if len(item.context.Info.Children) > 0 {
		details = append(details, fmt.Sprintf("children:%d", len(item.context.Info.Children)))
	}

	if len(details) > 0 {
		detailLine := lipgloss.NewStyle().
			PaddingLeft(5).
			Foreground(subtle).
			Render(strings.Join(details, " • "))
		line += "\n" + detailLine
	}

	return line
}

func (m Model) viewConfirm() string {
	var b strings.Builder

	item := m.items[m.cursor]
	strategy := kill.RecommendedStrategy(item.context)
	action := kill.Action{
		Strategy: strategy,
		Context:  item.context,
		Force:    m.force,
	}
	desc := kill.Describe(action)

	b.WriteString("\n")
	b.WriteString(confirmStyle.Render(fmt.Sprintf("  Kill? %s", desc)))
	b.WriteString("\n\n")

	// Show what we're about to kill
	b.WriteString(itemStyle.Render(m.renderItem(m.cursor, item)))
	b.WriteString("\n\n")

	if len(item.context.Info.Children) > 0 {
		b.WriteString(itemStyle.Render(
			errorStyle.Render(fmt.Sprintf("  Warning: %d child processes will also be affected", len(item.context.Info.Children))),
		))
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("  y/enter confirm • n/esc cancel"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) viewResult() string {
	var b strings.Builder
	b.WriteString("\n")
	if m.isError {
		b.WriteString(errorStyle.Render("  " + m.message))
	} else {
		b.WriteString(successStyle.Render("  " + m.message))
	}
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("  r retry • q/enter quit"))
	b.WriteString("\n")
	return b.String()
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
