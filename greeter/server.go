package greeter

import (
	context "context"
	"net"

	"google.golang.org/grpc"
)

type Server struct{}

func (s *Server) GreetRPC(ctx context.Context, req *GreetReq) (*GreetReply, error) {
	return &GreetReply{Reply: "Yo"}, nil
}

// NewServer starts gRPC server and returns server for graceful stop in future
func NewServer(addr string) *grpc.Server {
	s := grpc.NewServer()

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	srv := &Server{}

	RegisterGreeterServer(s, srv)

	go s.Serve(lis)

	return s
}
