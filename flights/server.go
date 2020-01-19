package flights

import (
	context "context"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	server *grpc.Server
}

// NewServer starts gRPC server and returns server for graceful stop in future
func NewServer(addr string) *Server {
	// Creates a new gRPC server
	s := grpc.NewServer()
	// gRPC uses TCP as its underlying network protocol so we need to open a tcp listener
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	server := &Server{
		server: s,
	}
	// Registers gRPC server for rpcs defined in proto file
	RegisterFlightsServer(s, server)
	// Begins serving requests on separate goroutine
	go s.Serve(lis)

	return server
}

func (s *Server) BookRPC(ctx context.Context, req *BookReq) (*BookReply, error) {
	return nil, nil
}

func (s *Server) CancelRPC(ctx context.Context, req *CancelReq) (*CancelReply, error) {
	return nil, nil
}

// GracefulStop stops the gRPC server
func (s *Server) GracefulStop() {
	s.server.GracefulStop()
}
