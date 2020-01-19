package trips

import (
	"net"

	"google.golang.org/grpc"
)

// NewServer starts gRPC server and returns server for graceful stop in future
func NewServer(addr string) *grpc.Server {
	server := grpc.NewServer()

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	t := NewTrips()

	RegisterTripsServer(server, t)

	go server.Serve(lis)

	return server
}
