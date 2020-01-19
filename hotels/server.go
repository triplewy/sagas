package hotels

import (
	context "context"
	"net"

	"google.golang.org/grpc"
)

// NewServer starts gRPC server and returns server for graceful stop in future
func NewServer(addr string) (*grpc.Server, *Hotels) {
	server := grpc.NewServer(grpc.UnaryInterceptor(unaryInterceptor))

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	h := NewHotels()

	RegisterHotelsServer(server, h)

	go server.Serve(lis)

	return server, h
}

// Authorization unary interceptor function to handle authorize per RPC call
func unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if h, ok := info.Server.(*Hotels); ok {
		// Check if blocking network
		if h.blockNetwork.Load() {
			return nil, ErrNetworkBlocked
		}
	} else {
		panic("Should always be server")
	}

	// Calls the handler
	return handler(ctx, req)
}
