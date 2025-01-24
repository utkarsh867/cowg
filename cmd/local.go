package cmd

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/utkarsh867/cowg/cowg"
	"github.com/utkarsh867/cowg/db"
)

var localCommand = &cobra.Command{
	Use: "local",
	Short: "Run the local server",
	Long: "Start the local CLI.",
	Run: func(cmd *cobra.Command, args []string) {
		RunLocal()
	},
}

func RunLocal() {
	wgClient := cowg.CreateWgClient()
	defer wgClient.Close()

	devices, err := wgClient.Devices()
	if err != nil {
		log.Fatalf("Could not connect to db %s", err)
	}

	db, err := db.Connect()

	if err != nil {
		log.Fatal(err)
	}

	c := cowg.Cowg{
		Db:       db,
		WgClient: wgClient,
		WgDevice: devices[0],
	}

	model, err := cowg.InitialModel(&c)
	if err != nil {
		log.Print("Error initialModel")
		log.Print(err)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
