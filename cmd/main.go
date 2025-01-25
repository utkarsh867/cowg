package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/utkarsh867/cowg"
	"github.com/utkarsh867/cowg/db"
	"golang.zx2c4.com/wireguard/wgctrl"
)


type Model struct {
  peerlist tea.Model

  width int
  height int
}

var docStyle = lipgloss.NewStyle().AlignHorizontal(lipgloss.Center)

func createWgClient() *wgctrl.Client {
  client, err := wgctrl.New()
  if err != nil {
    log.Fatal(err)
  }
  return client
}

func initialModel(c *cowg.Cowg) (Model, error) {
  return Model {
    peerlist: cowg.NewPeerListModel(c),
  }, nil
}

func (m Model) Init() tea.Cmd {
  return m.peerlist.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  var cmds []tea.Cmd
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
  cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
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

func main() {
  wgClient := createWgClient()
  devices, err := wgClient.Devices()
  if err != nil {
    log.Fatalf("Could not connect to db %s", err)
  }
  
  db, err := db.Connect()
  if err != nil {
    log.Fatal(err)
  }

  c := cowg.Cowg{
    Db: db,
    WgClient: wgClient,
    WgDevice: devices[0],
  }

  model, err := initialModel(&c)
  if err != nil {
    log.Print("Error initialModel")
    log.Print(err)
  }

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
