package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
)

func main() {
	host := os.Getenv("CLIENT_HOST")
	port := os.Getenv("CLIENT_PORT")

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()

	renderer := bubbletea.MakeRenderer(s)
	style := style{
		renderer: renderer,
		width:    pty.Window.Width,
		height:   pty.Window.Height,
	}

	state := sharedState{}
	state.apiUrl = os.Getenv("API_URL")
	state.wsUrl = os.Getenv("WS_URL")
	
	// Fallback to derive WebSocket URL from API URL if WS_URL is not set
	if state.wsUrl == "" && state.apiUrl != "" {
		state.wsUrl = strings.Replace(state.apiUrl, "http", "ws", 1) + "/ws"
	}
	
	state.style = &style
	state.currentPage = Login

	m := baseModel{}
	m.sharedState = &state
	m.login = InitLoginModel(&state)
	m.signup = InitSignupModel(&state)
	m.ui = InitUIModel(&state)

	return m, []tea.ProgramOption{tea.WithAltScreen()}
}
