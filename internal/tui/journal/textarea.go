package journal

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	helpHeight = 5
)

var (
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	cursorLineStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("230"))

	placeholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("238"))

	endOfBufferStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("235"))

	focusedPlaceholderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("99"))

	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("238"))

	blurredBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.HiddenBorder())
)

type keymap = struct {
	next, prev, add, remove, quit key.Binding
}

type textareaModel struct {
	textarea.Model
	keymap  keymap
	help    help.Model
	focus   int
	browser *mdBrowser
}

func newTextarea(browser *mdBrowser) *textareaModel {
	t := textarea.New()
	t.Prompt = ""
	t.Placeholder = "Type your daily journal"
	t.ShowLineNumbers = true
	t.Cursor.Style = cursorStyle
	t.FocusedStyle.Placeholder = focusedPlaceholderStyle
	t.BlurredStyle.Placeholder = placeholderStyle
	t.FocusedStyle.CursorLine = cursorLineStyle
	t.FocusedStyle.Base = focusedBorderStyle
	t.BlurredStyle.Base = blurredBorderStyle
	t.FocusedStyle.EndOfBuffer = endOfBufferStyle
	t.BlurredStyle.EndOfBuffer = endOfBufferStyle
	t.KeyMap.DeleteWordBackward.SetEnabled(false)
	t.KeyMap.LineNext = key.NewBinding(key.WithKeys("down"))
	t.KeyMap.LinePrevious = key.NewBinding(key.WithKeys("up"))
	//t.Blur()
	return &textareaModel{
		Model: t,
		keymap: keymap{
			quit: key.NewBinding(
				key.WithKeys("esc", "ctrl+c"),
				key.WithHelp("esc", "quit"),
			),
		},
		browser: browser,
	}
}

func (m textareaModel) updateSizes(width, height int) {
	m.Model.SetWidth(width / 2)
	m.Model.SetHeight(height / 2)
}

func (m textareaModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m *textareaModel) Update(msg tea.Msg) (*textareaModel, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		default:
			_ = msg
			err := m.browser.render(m.Model.Value())
			if err != nil {
				panic(err)
			}
			// var cmd tea.Cmd
			// m.browser, cmd = m.browser.Update(msg)
			// cmds = append(cmds, cmd)
		}
	}
	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m textareaModel) View() string {
	help := m.help.ShortHelpView([]key.Binding{
		m.keymap.next,
		m.keymap.prev,
		m.keymap.add,
		m.keymap.remove,
		m.keymap.quit,
	})

	view := m.Model.View()
	return "\n\n" + lipgloss.JoinVertical(lipgloss.Top, view) + "\n\n" + help
}
