package main

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
	"github.com/utkarsh867/cowg"
	"github.com/utkarsh867/cowg/db"
)


func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
  pty, _, _ := s.Pty()

  wgClient := cowg.CreateWgClient()

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

  model, err := cowg.InitialModel(&c)
  model.Term = pty.Term

  if err != nil {
    log.Print("Error initialModel")
    log.Print(err)
  }

  return model, []tea.ProgramOption{tea.WithAltScreen()}
}

func main() {
  s, err := wish.NewServer(
    wish.WithAddress(net.JoinHostPort("localhost", "22022")),
    wish.WithMiddleware(
      bubbletea.Middleware(teaHandler),
      activeterm.Middleware(),
    ),
  )

  if err != nil {
    log.Fatal(err)
  }


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
		log.Fatalf("Could not stop server", "error", err)
	}
}
