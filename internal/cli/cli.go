package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/diogox/bspc-go"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/diogox/bspm/internal/feature"
	"github.com/diogox/bspm/internal/feature/state"
	"github.com/diogox/bspm/internal/ipc"
	"github.com/diogox/bspm/internal/log"
)

const (
	flagKeyDaemon        = "daemon"
	flagKeyVerbose       = "verbose"
	flagKeyMonocleToggle = "toggle"
	flagKeyMonocleNext   = "next"
	flagKeyMonoclePrev   = "prev"
)

type app struct {
	cli *cli.App
}

func New(logger *zap.Logger) app {
	return app{
		cli: &cli.App{
			Name:  "bspm",
			Usage: "the bspwm manager",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    flagKeyDaemon,
					Aliases: []string{"d"},
					Usage:   "run the manager deamon",
				},
				&cli.BoolFlag{
					Name:    flagKeyVerbose,
					Aliases: []string{"v"},
					Usage:   "verbose logging",
				},
			},
			Commands: []*cli.Command{
				{
					Name:  "monocle",
					Usage: "manages the transparent monocle workflow",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  flagKeyMonocleToggle,
							Usage: "toggles the transparent monocle workflow",
						},
						&cli.BoolFlag{
							Name:  flagKeyMonocleNext,
							Usage: "shows the next node in the transparent monocle workflow",
						},
						&cli.BoolFlag{
							Name:  flagKeyMonoclePrev,
							Usage: "shows the previous node in the transparent monocle workflow",
						},
					},
					Action: func(ctx *cli.Context) error {
						c, err := ipc.NewClient()
						if err != nil {
							return err
						}
						defer c.Close()

						if ctx.NumFlags() != 1 {
							return errors.New("only one flag is expected")
						}

						var (
							isToggle = ctx.Bool(flagKeyMonocleToggle)
							isNext   = ctx.Bool(flagKeyMonocleNext)
							isPrev   = ctx.Bool(flagKeyMonoclePrev)
						)

						var msg ipc.Message
						switch {
						case isToggle:
							msg = "monocle"
						case isPrev:
							msg = "prev"
						case isNext:
							msg = "next"
						default:
							return errors.New("unexpected error")
						}

						if err := c.Send(msg); err != nil {
							return fmt.Errorf("failed to communicate with manager: %v", err)
						}

						return nil
					},
				},
			},
			Action: func(ctx *cli.Context) error {
				if isDaemon := ctx.Bool(flagKeyDaemon); isDaemon {
					var isVerbose bool
					if ctx.Bool(flagKeyVerbose) {
						isVerbose = true
					}

					l, err := log.New(logger, isVerbose)
					if err != nil {
						return fmt.Errorf("failed to initialize logger: %v", err)
					}

					return runServerDaemon(l)
				}

				return errors.New("invalid arguments")
			},
		},
	}
}

func runServerDaemon(logger *log.Logger) error {
	server, err := ipc.NewServer()
	if err != nil {
		return err
	}

	bspwmClient, err := bspc.New(logger.WithoutFields())
	if err != nil {
		return fmt.Errorf("failed to initialise bspwm client: %v", err)
	}

	monocle, _, err := feature.StartTransparentMonocle(logger, state.NewTransparentMonocle(), bspwmClient)
	if err != nil {
		return err
	}

	// TODO: Use cancel() function above to listen for Ctrl-C interruptions.

	msgCh, errCh := server.Listen()
	defer server.Close()

	logger.Info("Running daemon")

	// TODO: Either override existing socket file, or add a listener to close the server when the user uses Ctrl-C
	//  Closing it, in the line above, will remove the file. But since I'm killing the process, it never gets to run that.
	//  Also, unrelated, am I closing the socket in bspc-go when subscribing to events?

	for {
		select {
		case msg := <-msgCh: // TODO: Use a JSON struct as a message instead for versatility
			switch msg {
			case "monocle":
				logger.Info("Toggling transparent monocle mode")
				if err := monocle.ToggleCurrentDesktop(); err != nil {
					logger.Error("failed to toggle transparent monocle mode", zap.Error(err))
				}
			case "next":
				if err := monocle.FocusNextHiddenNode(); err != nil {
					logger.Error("failed to focus next node in transparent monocle mode", zap.Error(err))
				}
			case "prev":
				if err := monocle.FocusPreviousHiddenNode(); err != nil {
					logger.Error("failed to focus previous node in transparent monocle mode", zap.Error(err))
				}
			}
		case err := <-errCh:
			logger.Error("error while receiving ipc message from client", zap.Error(err))
		}
	}
}

func (a app) Run() error {
	if err := a.cli.Run(os.Args); err != nil {
		return err
	}

	return nil
}
