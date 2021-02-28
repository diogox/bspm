package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/fatih/color"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	transparentmonocle "github.com/diogox/bspm/internal/feature/transparent_monocle"
	"github.com/diogox/bspm/internal/grpc/bspm"
	"github.com/diogox/bspm/internal/log"
)

const unixSocketPath = "/tmp/bspm.socket"

func NewServer(logger *log.Logger, monocleService transparentmonocle.Feature) (func() error, func()) {
	s := grpc.NewServer()
	bspm.RegisterBSPMServer(s, &server{
		logger:         logger,
		monocleService: monocleService,
	})

	start := func() error { return startServer(s) }
	stop := func() { s.GracefulStop() }

	return start, stop
}

func startServer(s *grpc.Server) error {
	lis, err := net.Listen("unix", unixSocketPath)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

type server struct {
	logger         *log.Logger
	monocleService transparentmonocle.Feature
}

func (s *server) ToggleMonocleMode(context.Context, *empty.Empty) (*empty.Empty, error) {
	s.logger.Info("Toggling transparent monocle mode")

	if err := s.monocleService.ToggleCurrentDesktop(); err != nil {
		color.Red("Failed to toggle transparent monocle mode")
		s.logger.Error("failed to toggle transparent monocle mode", zap.Error(err))
	}

	return &empty.Empty{}, nil
}

func (s *server) MonocleModeCycle(_ context.Context, req *bspm.MonocleModeCycleRequest) (*empty.Empty, error) {
	switch req.CycleDirection {
	case bspm.CycleDir_CYCLE_DIR_PREV:
		if err := s.monocleService.FocusPreviousHiddenNode(); err != nil {
			return nil, fmt.Errorf("failed to focus previous node in transparent mode: %w", err)
		}
	case bspm.CycleDir_CYCLE_DIR_NEXT:
		if err := s.monocleService.FocusNextHiddenNode(); err != nil {
			return nil, fmt.Errorf("failed to focus next node in transparent mode: %w", err)
		}
	default:
		return nil, errors.New("invalid monocle mode cycling direction")
	}

	return &empty.Empty{}, nil
}