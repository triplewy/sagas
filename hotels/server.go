package hotels

import (
	context "context"
	"errors"
	"log"
	"math/rand"
	"net"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc/metadata"

	"google.golang.org/grpc"
)

// Errors encountered by serving gRPC requests
var (
	ErrInvalidServer     = errors.New("unary interceptor encountered invalid server interface")
	ErrNoMetadata        = errors.New("failed to retrieve metadata from request")
	ErrTooManyRequestIDs = errors.New("request contains more than 1 requestID")
	ErrNoRequestID       = errors.New("no requestID found in request metadata")
)

// NewServer starts gRPC server and returns server for graceful stop in future
func NewServer(addr string) (*grpc.Server, *Hotels) {
	server := grpc.NewServer(grpc.UnaryInterceptor(serverInterceptor))

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
func serverInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Println(info.FullMethod)

	h, ok := info.Server.(*Hotels)
	if !ok {
		return nil, serverError(ErrInvalidServer)
	}

	// Check if blocking network
	if h.BlockNetwork.Load() {
		return nil, retryError(ErrNetworkBlocked)
	}

	// Check if slow network
	if h.SlowNetwork.Load() {
		// Sleep random time between 2-5s
		sleep := ClientTimeout + rand.Intn(4)
		time.Sleep(time.Duration(sleep) * time.Second)
	}

	// Check if requestID exists in cache
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, clientError(ErrNoMetadata)
	}
	arr, ok := md["request-id"]
	if !ok {
		return nil, clientError(ErrNoRequestID)
	}
	if len(arr) != 1 {
		return nil, clientError(ErrTooManyRequestIDs)
	}
	id := arr[0]

	if reply, ok := h.requests.Get(id); ok {
		if err, ok := reply.(error); ok {
			return nil, err
		}
		return reply, nil
	}

	// Calls the handler
	reply, err := handler(ctx, req)

	// Save handler response to requests cache
	if err != nil {
		err = serverError(err)
		h.requests.Set(id, err)
	} else {
		h.requests.Set(id, reply)
	}

	return reply, err
}

func retryError(err error) error {
	st := status.New(codes.Unavailable, err.Error())
	return st.Err()
}

func clientError(err error) error {
	st := status.New(codes.InvalidArgument, err.Error())
	return st.Err()
}

func serverError(err error) error {
	st := status.New(codes.Internal, err.Error())
	return st.Err()
}
