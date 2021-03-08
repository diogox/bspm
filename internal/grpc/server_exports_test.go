package grpc

import (
	transparentmonocle "github.com/diogox/bspm/internal/feature/transparent_monocle"
	"github.com/diogox/bspm/internal/log"
)

func NewTestServer(logger *log.Logger, monocleService transparentmonocle.Feature) *server {
	return &server{
		logger:         logger,
		monocleService: monocleService,
	}
}
