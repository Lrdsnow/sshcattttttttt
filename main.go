package main

import (
	"context"
	"errors"
  "fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "embed"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/logging"
)

const (
	host = "localhost"
	port = "23234"
)

//go:embed cat.txt
var cat string

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		wish.WithHostKeyPath(fmt.Sprintf("%s/.ssh/id_ed25519", homeDir)),
		wish.WithMiddleware(
			func(next ssh.Handler) ssh.Handler {
				return func(sess ssh.Session) {
          wish.Println(sess, cat)
          dat, err := os.ReadFile("motd.txt")
          if err != nil {
            log.Error("Could not read motd.txt", "error", err)
          }
          fileInfo, err := os.Stat("motd.txt")
          if err != nil {
            log.Error("Could not stat motd.txt", "error", err)
          }
          wish.Println(sess, string(dat))
          next(sess)
				}
			},
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
