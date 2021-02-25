package cli

import (
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

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

func New(logger *zap.Logger, version string) app {
	return app{
		cli: &cli.App{
			Name:    "bspm",
			Usage:   "the bspwm manager",
			Version: version,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    flagKeyDaemon,
					Aliases: []string{"d"},
					Usage:   "Run the manager deamon",
				},
				&cli.BoolFlag{
					Name:  flagKeyVerbose,
					Usage: "Verbose logging",
				},
			},
			Commands: []*cli.Command{
				{
					Name:  "monocle",
					Usage: "Manages the transparent monocle workflow",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:  flagKeyMonocleToggle,
							Usage: "Toggles the transparent monocle workflow",
						},
						&cli.BoolFlag{
							Name:  flagKeyMonocleNext,
							Usage: "Shows the next node in the transparent monocle workflow",
						},
						&cli.BoolFlag{
							Name:  flagKeyMonoclePrev,
							Usage: "Shows the previous node in the transparent monocle workflow",
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
							msg = DaemonCommandMonocleToggle
						case isPrev:
							msg = DaemonCommandMonoclePrevious
						case isNext:
							msg = DaemonCommandMonocleNext
						default:
							return errors.New("unexpected error")
						}

						if err := c.Send(msg); err != nil {
							color.Red("Failed: %v", err)
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

					return runDaemon(l)
				}

				return errors.New("invalid arguments")
			},
		},
	}
}
