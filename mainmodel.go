package cowg

import (

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
  peerlist tea.Model

  width int
  height int

  Term string
}

func InitialModel(c *Cowg) (Model, error) {
  return Model {
    peerlist: NewPeerListModel(c),
  }, nil
}

func (m Model) Init() tea.Cmd {
  return m.peerlist.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		}
  case tea.WindowSizeMsg:
    m.height = msg.Height
    m.width = msg.Width
	}

  var cmd tea.Cmd
  m.peerlist, cmd = m.peerlist.Update(msg)

	return m, cmd
}

func (m Model) View() string {
  return lipgloss.Place(
    m.width,
    m.height,
    lipgloss.Center,
    lipgloss.Center,
    m.peerlist.View(),
  )
}
