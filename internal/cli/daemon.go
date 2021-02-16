package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/diogox/bspc-go"
	"github.com/fatih/color"
	"go.uber.org/zap"

	"github.com/diogox/bspm/internal/feature"
	"github.com/diogox/bspm/internal/feature/state"
	"github.com/diogox/bspm/internal/ipc"
	"github.com/diogox/bspm/internal/log"
)

const (
	DaemonCommandMonocleToggle   = "monocle-toggle"
	DaemonCommandMonocleNext     = "monocle-next"
	DaemonCommandMonoclePrevious = "monocle-prev"
)

func runDaemon(logger *log.Logger) error {
	server, err := ipc.NewServer()
	if err != nil {
		return err
	}

	bspwmClient, err := bspc.New(logger.WithoutFields())
	if err != nil {
		return fmt.Errorf("failed to initialise bspwm client: %v", err)
	}

	monocle, cancel, err := feature.StartTransparentMonocle(logger, state.NewTransparentMonocle(), bspwmClient)
	if err != nil {
		return err
	}
	defer cancel()

	msgCh, errCh := server.Listen()
	defer server.Close()

	color.Blue("Daemon Running...")
	logger.Info("daemon started")

	//  TODO: Unrelated, am I closing the socket in bspc-go when subscribing to events?

	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case msg := <-msgCh: // TODO: Use a JSON struct as a message instead for versatility
			switch msg {
			case DaemonCommandMonocleToggle:
				logger.Info("Toggling transparent monocle mode")
				if err := monocle.ToggleCurrentDesktop(); err != nil {
					color.Red("Failed to toggle transparent monocle mode")
					logger.Error("failed to toggle transparent monocle mode", zap.Error(err))
				}
			case DaemonCommandMonocleNext:
				if err := monocle.FocusNextHiddenNode(); err != nil {
					color.Red("Failed to focus next node in transparent monocle mode")
					logger.Error("failed to focus next node in transparent monocle mode", zap.Error(err))
				}
			case DaemonCommandMonoclePrevious:
				if err := monocle.FocusPreviousHiddenNode(); err != nil {
					color.Red("Failed to focus previous node in transparent monocle mode")
					logger.Error("failed to focus previous node in transparent monocle mode", zap.Error(err))
				}
			}
		case err := <-errCh:
			color.Red("Error: %v", err)
			logger.Error("error while receiving ipc message from client", zap.Error(err))
		case <-exitCh:
			color.Blue("Daemon Stopped!")
			logger.Info("daemon stopped")
			return nil
		}
	}
}

func (a app) Run() error {
	if err := a.cli.Run(os.Args); err != nil {
		return err
	}

	return nil
}
