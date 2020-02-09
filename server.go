package sagas

import (
	"net"

	"google.golang.org/grpc"
)

// NewServer starts a new gRPC server for sagas
func NewServer(addr string, c *Coordinator) *grpc.Server {
	s := grpc.NewServer()

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	RegisterCoordinatorServer(s, c)

	go s.Serve(lis)

	return s
}
