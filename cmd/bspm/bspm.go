package main

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/diogox/bspm/internal/cli"
)

var Version string

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(fmt.Errorf("failed to initialise logger: %v", err))
	}
	defer func() {
		_ = logger.Sync()
	}()

	if err := cli.New(logger, Version).Run(); err != nil {
		logger.Error("failed to run desired command", zap.Error(err))
		return
	}
}
