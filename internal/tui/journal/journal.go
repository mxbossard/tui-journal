package journal

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	input   *textareaModel
	browser *mdBrowser
	keymap  keymap
	help    help.Model
}

func NewModel() (*model, error) {
	browser, err := newMdBrowser()
	if err != nil {
		return nil, err
	}
	input := newTextarea(browser)

	m := &model{
		input:   input,
		browser: browser,
		help:    help.New(),
		keymap: keymap{
			next: key.NewBinding(
				key.WithKeys("tab"),
				key.WithHelp("tab", "next"),
			),
			prev: key.NewBinding(
				key.WithKeys("shift+tab"),
				key.WithHelp("shift+tab", "prev"),
			),
			add: key.NewBinding(
				key.WithKeys("ctrl+n"),
				key.WithHelp("ctrl+n", "add an editor"),
			),
			remove: key.NewBinding(
				key.WithKeys("ctrl+w"),
				key.WithHelp("ctrl+w", "remove an editor"),
			),
			quit: key.NewBinding(
				key.WithKeys("esc", "ctrl+c"),
				key.WithHelp("esc", "quit"),
			),
		},
	}

	input.Focus()
	return m, nil
}

func (m model) updateSizes(width, height int) (err error) {
	m.input.updateSizes(width, height)
	err = m.browser.updateSizes(width, height)
	return
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.quit):
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		err := m.updateSizes(msg.Width, msg.Height)
		if err != nil {
			panic(err)
		}
	}

	newModel, cmd := m.input.Update(msg)
	m.input = newModel
	cmds = append(cmds, cmd)

	newModel2, cmd := m.browser.Update(msg)
	m.browser = newModel2
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	help := m.help.ShortHelpView([]key.Binding{
		m.keymap.quit,
	})

	var views []string
	views = append(views, m.input.View())
	views = append(views, m.browser.View())
	return "\n\n" + lipgloss.JoinHorizontal(lipgloss.Top, views...) + "\n\n" + help
}
