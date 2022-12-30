package server

import (
	"context"

	"github.com/stockmq/stockmq-server/pb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Backend is used to implement pb.Monitor.
type Backend struct {
	pb.UnimplementedMonitorServer
	s *Server
}

// IsRunning returns whether service is in running state.
func (b *Backend) IsRunning(ctx context.Context, in *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	return wrapperspb.Bool(b.s.healthStatus().Error == ""), nil
}
