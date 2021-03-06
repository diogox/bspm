package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/diogox/bspm/internal/subscription"

	"github.com/fatih/color"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/diogox/bspm/internal/grpc"
	"github.com/diogox/bspm/internal/grpc/bspm"
	"github.com/diogox/bspm/internal/log"
)

const (
	flagKeyDaemon                    = "daemon"
	flagKeyVerbose                   = "verbose"
	flagKeyMonocleToggle             = "toggle"
	flagKeyMonocleNext               = "next"
	flagKeyMonoclePrev               = "prev"
	flagKeyMonocleSubscribeNodeCount = "subscribe-node-count"
)

type app struct {
	cli *cli.App
}

func New(logger *zap.Logger, version string) app {
	subscriptionManager := subscription.NewManager()

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
			ExitErrHandler: func(context *cli.Context, err error) {
				color.Red("Failed: %v", err)
				os.Exit(1)
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
						&cli.BoolFlag{
							Name:  flagKeyMonocleSubscribeNodeCount,
							Usage: "Returns the number of nodes in the transparent monocle workflow, every time it changes",
						},
					},
					Action: func(ctx *cli.Context) error {
						c, err := grpc.NewClient()
						if err != nil {
							return err
						}

						if ctx.NumFlags() != 1 {
							return errors.New("only one flag is expected")
						}

						var (
							isToggle             = ctx.Bool(flagKeyMonocleToggle)
							isNext               = ctx.Bool(flagKeyMonocleNext)
							isPrev               = ctx.Bool(flagKeyMonoclePrev)
							isSubscribeNodeCount = ctx.Bool(flagKeyMonocleSubscribeNodeCount)
						)

						switch {
						case isToggle:
							if _, err := c.MonocleModeToggle(ctx.Context, &empty.Empty{}); err != nil {
								return fmt.Errorf("failed to toggle monocle mode: %w", err)
							}
						case isPrev:
							req := &bspm.MonocleModeCycleRequest{
								CycleDirection: bspm.CycleDir_CYCLE_DIR_PREV,
							}

							if _, err := c.MonocleModeCycle(ctx.Context, req); err != nil {
								return fmt.Errorf("failed to cycle to previous node in monocle mode: %w", err)
							}
						case isNext:
							req := &bspm.MonocleModeCycleRequest{
								CycleDirection: bspm.CycleDir_CYCLE_DIR_NEXT,
							}

							if _, err := c.MonocleModeCycle(ctx.Context, req); err != nil {
								return fmt.Errorf("failed to cycle to next node in monocle mode: %w", err)
							}
						case isSubscribeNodeCount:
							req := &bspm.MonocleModeSubscribeRequest{
								Type: bspm.MonocleModeSubscriptionType_MONOCLE_MODE_SUBSCRIPTION_TYPE_NODE_COUNT,
							}

							subscription, err := c.MonocleModeSubscribe(ctx.Context, req)
							if err != nil {
								return fmt.Errorf("failed to subscribe to node count in monocle mode: %w", err)
							}

							// TODO: Graceful shutdown
							for {
								msg, err := subscription.Recv()
								if err != nil {
									if errors.Is(err, io.EOF) {
										return nil
									}

									return fmt.Errorf("failed to receive message from monocle mode node count subscription: %w", err)
								}

								fmt.Println(msg.GetNodeCount())
							}
						default:
							return errors.New("unexpected error")
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

					return runDaemon(l, subscriptionManager)
				}

				return errors.New("invalid arguments")
			},
		},
	}
}

func (a app) Run() error {
	if err := a.cli.Run(os.Args); err != nil {
		return err
	}

	return nil
}
