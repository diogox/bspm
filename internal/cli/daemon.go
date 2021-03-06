package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/diogox/bspc-go"
	"github.com/fatih/color"

	"github.com/diogox/bspm/internal/bspwm"
	bspwmdesktop "github.com/diogox/bspm/internal/bspwm/desktop"
	bspwmevent "github.com/diogox/bspm/internal/bspwm/event"
	bspwmnode "github.com/diogox/bspm/internal/bspwm/node"
	transparentmonocle "github.com/diogox/bspm/internal/feature/transparent_monocle"
	"github.com/diogox/bspm/internal/feature/transparent_monocle/state"
	"github.com/diogox/bspm/internal/grpc"
	"github.com/diogox/bspm/internal/log"
	"github.com/diogox/bspm/internal/subscription"
)

func runDaemon(logger *log.Logger, subscriptionManager subscription.Manager) error {
	bspwmClient, err := bspc.New(logger.WithoutFields())
	if err != nil {
		return fmt.Errorf("failed to initialise bspwm client: %v", err)
	}

	monocle, cancel, err := transparentmonocle.Start(
		logger,
		state.NewTransparentMonocle(subscriptionManager),
		bspwm.NewService(
			bspwmClient,
			bspwmdesktop.NewService(bspwmClient),
			bspwmnode.NewService(bspwmClient),
			bspwmevent.NewManager(logger, bspwmClient),
		),
		subscriptionManager,
	)
	if err != nil {
		return err
	}
	defer cancel()

	color.Blue("Daemon Running...")
	logger.Info("daemon started")

	startServer, stopServer := grpc.NewServer(logger, monocle)

	go func() {
		exitCh := make(chan os.Signal, 1)
		signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM)

		// Wait for Ctrl-C
		<-exitCh

		stopServer()
		color.Blue("Daemon Stopped!")
		logger.Info("daemon stopped")
	}()

	if err := startServer(); err != nil {
		color.Red("Daemon Failed to Start!!")
		logger.Info("daemon failed to start")

		return err
	}

	return nil
}
