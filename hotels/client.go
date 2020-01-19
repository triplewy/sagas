package hotels

import (
	context "context"
	"time"

	"github.com/triplewy/sagas/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func NewClient(addr string) HotelsClient {
	// Establishes connection with gRPC server
	cc, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	// Registers client to call rpcs defined in proto file
	return NewHotelsClient(cc)
}

func BookHotel(client HotelsClient, userID uint64) *BookReply {
	md := metadata.Pairs("id", "idempotent-key")
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Retry RPC until we get valid response
	reply, err := client.BookRPC(ctx, &BookReq{UserID: userID})
	retries := 0
	for err != nil {
		sleep := utils.Backoff(retries)
		time.Sleep(time.Duration(sleep) * time.Second)
		reply, err = client.BookRPC(ctx, &BookReq{UserID: userID})
	}
	return reply
}
