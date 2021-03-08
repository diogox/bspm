//go:generate mockgen -package grpc -destination bspm_proto_mock.go github.com/diogox/bspm/internal/grpc/bspm BSPM_MonocleModeSubscribeServer

package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"

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

	var (
		start = func() error { return startServer(s) }
		stop  = func() { s.GracefulStop() }
	)

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

func (s *server) MonocleModeToggle(context.Context, *empty.Empty) (*empty.Empty, error) {
	s.logger.Info("Toggling transparent monocle mode")

	if err := s.monocleService.ToggleCurrentDesktop(); err != nil {
		s.logger.Error("failed to toggle transparent monocle mode", zap.Error(err))
		return &empty.Empty{}, fmt.Errorf("failed to toggle transparent monocle mode: %w", err)
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

func (s *server) MonocleModeSubscribe(req *bspm.MonocleModeSubscribeRequest, stream bspm.BSPM_MonocleModeSubscribeServer) error {
	switch req.Type {
	case bspm.MonocleModeSubscriptionType_MONOCLE_MODE_SUBSCRIPTION_TYPE_NODE_COUNT:
		for newCount := range s.monocleService.SubscribeNodeCount() {
			err := stream.Send(&bspm.MonocleModeSubscribeResponse{
				SubscriptionType: &bspm.MonocleModeSubscribeResponse_NodeCount{
					NodeCount: int32(newCount),
				},
			})
			if err != nil {
				return fmt.Errorf("failed with to send subscription response: %w", err)
			}
		}

	default:
		return errors.New("invalid subscription type")
	}

	return nil
}
