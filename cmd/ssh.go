package cmd

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/spf13/cobra"
	"github.com/utkarsh867/cowg/api"
	"github.com/utkarsh867/cowg/cowg"
	"github.com/utkarsh867/cowg/db"
)

var sshCommand = &cobra.Command{
	Use:   "ssh",
	Short: "Run a Wish server",
	Long:  "Start the SSH server so that user can connect to it.",
	Run: func(cmd *cobra.Command, args []string) {
		WishServer()
	},
}

func teaHandler(c *cowg.Cowg, s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()

	model, err := cowg.InitialModel(c)
	model.Term = pty.Term

	if err != nil {
		log.Print("Error initialModel")
		log.Print(err)
	}

	return model, []tea.ProgramOption{tea.WithAltScreen()}
}

func TeaMiddlewareWithContext(c *cowg.Cowg) wish.Middleware {
	return bubbletea.Middleware(func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
    // Do a cleanup when the session ends
		go func() {
			select {
			case <-s.Context().Done():
        c.Config = ""
			}
		}()
		return teaHandler(c, s)
	})
}

func WishServer() {
	wgClient := cowg.CreateWgClient()

	devices, err := wgClient.Devices()
	if err != nil {
		log.Fatalf("Could not connect to db %s", err)
	}

	db, err := db.Connect()

	if err != nil {
		log.Fatal(err)
	}

	c := &cowg.Cowg{
		Db:       db,
		WgClient: wgClient,
		WgDevice: devices[0],
		Config:   "",
	}

	go api.RunHTTPServer(c)

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort("0.0.0.0", "22022")),
		wish.WithMiddleware(
			TeaMiddlewareWithContext(c),
			activeterm.Middleware(),
		),
	)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Wish Server started...")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			done <- nil
		}
	}()

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()

	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Fatalf("Could not stop server %s", err)
	}
}
