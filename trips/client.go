package trips

import (
	"google.golang.org/grpc"
)

func NewClient(addr string) TripsClient {
	// Establishes connection with gRPC server
	cc, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	// Registers client to call rpcs defined in proto file
	return NewTripsClient(cc)
}
