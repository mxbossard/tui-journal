package main

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
	const width = 78

	vp := viewport.New(width, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	// We need to adjust the width of the glamour render from our main width
	// to account for a few things:
	//
	//  * The viewport border width
	//  * The viewport padding
	//  * The viewport margins
	//  * The gutter glamour applies to the left side of the content
	//
	glamourRenderWidth := width - vp.Style.GetHorizontalFrameSize() - glamourGutter

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(glamourRenderWidth),
	)
	if err != nil {
		return nil, err
	}

	//content = ""
	str, err := renderer.Render(content)
	if err != nil {
		return nil, err
	}

	vp.SetContent(str)

	return &mdBrowser{
		viewport: vp,
		renderer: renderer,
	}, nil
}

func (e mdBrowser) Init() tea.Cmd {
	return nil
}

type UpdateMsg struct {
	tea.Msg
	content string
}

func (e *mdBrowser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return e, tea.Quit
		default:
			var cmd tea.Cmd
			e.viewport, cmd = e.viewport.Update(msg)
			return e, cmd
		}
	case UpdateMsg:
		//fmt.Printf("updated with msg content: %s\n", msg.content)
		str, err := e.renderer.Render(msg.content)
		if err != nil {
			panic(err)
		}

		e.viewport.SetContent(str)
	case tea.WindowSizeMsg:
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			//glamour.WithWordWrap(msg.Width/2-e.viewport.Style.GetHorizontalFrameSize()-glamourGutter),
			glamour.WithWordWrap(e.viewport.Width),
		)
		if err != nil {
			panic(err)
		}
		e.renderer = renderer
	}
	return e, nil
}

func (e mdBrowser) View() string {
	return e.viewport.View() //+ e.helpView()
}

func (e mdBrowser) helpView() string {
	return helpStyle("\n  ↑/↓: Navigate • q: Quit\n")
}
