package journal

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var content = `
# Today’s Menu

## Appetizers

| Name        | Price | Notes                           |
| ---         | ---   | ---                             |
| Tsukemono   | $2    | Just an appetizer               |
| Tomato Soup | $4    | Made with San Marzano tomatoes  |
| Okonomiyaki | $4    | Takes a few minutes to make     |
| Curry       | $3    | We can add squash if you’d like |

## Seasonal Dishes

| Name                 | Price | Notes              |
| ---                  | ---   | ---                |
| Steamed bitter melon | $2    | Not so bitter      |
| Takoyaki             | $3    | Fun to eat         |
| Winter squash        | $3    | Today it's pumpkin |

## Desserts

| Name         | Price | Notes                 |
| ---          | ---   | ---                   |
| Dorayaki     | $4    | Looks good on rabbits |
| Banana Split | $5    | A classic             |
| Cream Puff   | $3    | Pretty creamy!        |

All our dishes are made in-house by Karen, our chef. Most of our ingredients
are from our garden or the fish market down the street.

Some famous people that have eaten here lately:

* [x] René Redzepi
* [x] David Chang
* [ ] Jiro Ono (maybe some day)

Bon appétit!
`

const glamourGutter = 2

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render

type mdBrowser struct {
	viewport viewport.Model
	renderer *glamour.TermRenderer
}

func newMdBrowser() (*mdBrowser, error) {
	width := 10
	height := 5
	vp := viewport.New(width, height)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	m := &mdBrowser{
		viewport: vp,
	}

	glamourRenderWidth := width - m.viewport.Style.GetHorizontalFrameSize() - glamourGutter
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(glamourRenderWidth),
	)
	if err != nil {
		return nil, err
	}
	m.renderer = renderer

	err = m.updateSizes(width, height)
	if err != nil {
		return nil, err
	}

	err = m.render(content)

	return m, nil
}

func (m *mdBrowser) updateSizes(width, height int) error {
	m.viewport.Width = width
	m.viewport.Height = height

	// We need to adjust the width of the glamour render from our main width
	// to account for a few things:
	//
	//  * The viewport border width
	//  * The viewport padding
	//  * The viewport margins
	//  * The gutter glamour applies to the left side of the content
	//

	// glamourRenderWidth := width - m.viewport.Style.GetHorizontalFrameSize() - glamourGutter
	// renderer, err := glamour.NewTermRenderer(
	// 	glamour.WithAutoStyle(),
	// 	glamour.WithWordWrap(glamourRenderWidth),
	// )
	// if err != nil {
	// 	return err
	// }
	// m.renderer = renderer
	return nil
}

func (m *mdBrowser) render(content string) error {
	str, err := m.renderer.Render(content)
	if err != nil {
		return err
	}
	m.viewport.SetContent(str)
	return nil
}

func (m mdBrowser) Init() tea.Cmd {
	return nil
}

func (m *mdBrowser) Update(msg tea.Msg) (*mdBrowser, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m mdBrowser) View() string {
	return m.viewport.View() //+ e.helpView()
}

func (m mdBrowser) helpView() string {
	return helpStyle("\n  ↑/↓: Navigate • q: Quit\n")
}
