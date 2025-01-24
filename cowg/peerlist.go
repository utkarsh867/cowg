package cowg

import (
	"bytes"
	"log"
	"net"
	"sort"

	"github.com/charmbracelet/bubbles/key"
	list "github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/utkarsh867/cowg/db"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var docStyle = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("63")).Margin(2, 0)

type PeerListModel struct {
	app              *Cowg
	wgClient         *wgctrl.Client
	wgDevice         *wgtypes.Device
	peers            []wgtypes.Peer
	additionalKeyMap PeerListKeyMap

	list     list.Model
	peerForm *huh.Form

	configView viewport.Model

	inputFlag      bool
	configViewFlag bool
}

type PeerListItem struct {
	title     string
	desc      string
	publicKey string
}

func (p PeerListItem) Title() string {
	return p.title
}

func (p PeerListItem) Description() string {
	return p.desc
}

func (p PeerListItem) PublicKey() string {
	return p.publicKey
}

func (p PeerListItem) FilterValue() string {
	return p.title
}

type PeerListKeyMap struct {
	addPeer        key.Binding
	deletePeer     key.Binding
	showConfig     key.Binding
	downloadConfig key.Binding
}

func GetPeersListItems(c *Cowg) []list.Item {
	wgDevice, err := c.WgClient.Device(c.WgDevice.Name)
	if err != nil {
		log.Print("Could not query device")
		log.Printf("%s", c.WgDevice.Name)
		log.Fatal(err)
	}
	peers := wgDevice.Peers

	sort.Slice(peers, func(i, j int) bool {
		return bytes.Compare(peers[i].AllowedIPs[0].IP, peers[j].AllowedIPs[0].IP) < 0
	})

	var listItems []list.Item
	for _, p := range peers {
		result := &db.Peer{}
		_ = c.Db.Client.Where(&db.Peer{PublicKey: p.PublicKey.String()}).First(&result)
		var title string
		if result.Name == "" {
			title = p.PublicKey.String()
		} else {
			title = result.Name
		}
		listItem := PeerListItem{
			title:     title,
			desc:      p.AllowedIPs[0].String(),
			publicKey: p.PublicKey.String(),
		}
		listItems = append(listItems, listItem)
	}

	return listItems
}

func NewPeerForm() *huh.Form {
	peerForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Peer Name").Key("name"),
		),
		huh.NewGroup(
			huh.NewInput().Title("Peer IPv4").Key("address").Placeholder("10.8.0.0"),
		),
	)
	return peerForm
}

func NewPeerListModel(c *Cowg) PeerListModel {
	listItems := GetPeersListItems(c)

	additionalKeyMap := PeerListKeyMap{
		addPeer: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add peer"),
		),
		deletePeer: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "delete peer"),
		),
		showConfig: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "view config"),
		),
    downloadConfig: key.NewBinding(
      key.WithKeys("d"),
      key.WithHelp("d", "download config"),
    ),
	}

	l := list.New(listItems, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Peers"
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			additionalKeyMap.addPeer,
			additionalKeyMap.deletePeer,
			additionalKeyMap.showConfig,
		}
	}
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			additionalKeyMap.addPeer,
			additionalKeyMap.deletePeer,
		}
	}
	l.DisableQuitKeybindings()
	l.KeyMap.NextPage.SetEnabled(false)

	return PeerListModel{
		app:              c,
		wgClient:         c.WgClient,
		wgDevice:         c.WgDevice,
		peers:            c.WgDevice.Peers,
		list:             l,
		additionalKeyMap: additionalKeyMap,
		peerForm:         nil,
		inputFlag:        false,
		configViewFlag:   false,
	}
}

func (m PeerListModel) Init() tea.Cmd {
	if m.peerForm == nil {
		return nil
	}
	return m.peerForm.Init()
}

func (m PeerListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.inputFlag {
		form, cmd := m.peerForm.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.peerForm = f
			cmds = append(cmds, cmd)
		}
		if m.peerForm.State == huh.StateAborted {
			m.inputFlag = false
		}

		if m.peerForm.State == huh.StateCompleted {
			peerName := m.peerForm.GetString("name")
			ipAddr := net.ParseIP(m.peerForm.GetString("address"))
			err := CreatePeer(m.app, peerName, ipAddr)
			if err != nil {
				log.Print("Error in creating peer")
				log.Fatal(err)
			}
			m.inputFlag = false
			cmd := m.list.SetItems(GetPeersListItems(m.app))
			cmds = append(cmds, cmd)
		}
	} else {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case msg.String() == "ctrl+c":
				return m, tea.Quit
			case key.Matches(msg, m.additionalKeyMap.addPeer):
				m.peerForm = NewPeerForm()
				cmd := m.peerForm.Init()
				m.inputFlag = true
				cmds = append(cmds, cmd)
			case key.Matches(msg, m.additionalKeyMap.deletePeer):
				DeletePeer(m.app, m.list.SelectedItem().(PeerListItem).PublicKey())
				cmd := m.list.SetItems(GetPeersListItems(m.app))
				cmds = append(cmds, cmd)
			case key.Matches(msg, m.additionalKeyMap.showConfig):
				s, _ := PeerConfig(m.app, m.list.SelectedItem().(PeerListItem))
				m.configView.SetContent(s)
				m.configViewFlag = true
			case key.Matches(msg, m.additionalKeyMap.downloadConfig):
				s, _ := PeerConfig(m.app, m.list.SelectedItem().(PeerListItem))
				m.configView.SetContent(s)
				m.configViewFlag = true
        m.app.Config = s
			}
		case tea.WindowSizeMsg:
			h, v := docStyle.GetFrameSize()
			m.list.SetSize(msg.Width-h, msg.Height-v)
			m.configView = viewport.New(60, msg.Height-v)
			m.configView.YPosition = 0
		}
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m PeerListModel) View() string {
	listString := docStyle.Render(m.list.View())
	if m.inputFlag {
		return m.peerForm.View()
	}

	if m.configViewFlag {
		return lipgloss.JoinHorizontal(lipgloss.Center, listString, docStyle.Render(m.configView.View()))
	}
	return listString
}
