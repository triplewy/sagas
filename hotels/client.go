package hotels

import (
	context "context"
	fmt "fmt"
	"time"

	"github.com/lithammer/shortuuid"
	"github.com/triplewy/sagas/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const ClientTimeout = 2

func NewClient(addr string) HotelsClient {
	// Establishes connection with gRPC server
	cc, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithUnaryInterceptor(clientInterceptor))
	if err != nil {
		panic(err)
	}
	// Registers client to call rpcs defined in proto file
	return NewHotelsClient(cc)
}

func clientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx, cancel := context.WithTimeout(ctx, ClientTimeout*time.Second)
	defer cancel()

	return invoker(ctx, method, req, reply, cc, opts...)
}

func contextWithRequestID(userID string) context.Context {
	uuid := shortuuid.New()
	id := fmt.Sprintf("%v:%v", userID, uuid)
	md := metadata.Pairs("request-id", id)
	return metadata.NewOutgoingContext(context.Background(), md)
}

func BookRoom(client HotelsClient, userID, roomID string) (string, error) {
	// Retry RPC until we get valid response
	ctx := contextWithRequestID(userID)
	req := &BookReq{
		UserID: userID,
		RoomID: roomID,
	}
	reply, err := client.BookRPC(ctx, req)
	retries := 0
	for err != nil {
		sleep := utils.Backoff(retries)
		time.Sleep(time.Duration(sleep) * time.Second)
		reply, err = client.BookRPC(ctx, req)
		if err != nil {
			st := status.Convert(err)
			if st.Code() != codes.Unavailable && st.Code() != codes.DeadlineExceeded {
				return "", err
			}
		}
	}
	return reply.GetReservationID(), nil
}

func CancelRoom(client HotelsClient, userID, reservationID string) error {
	ctx := contextWithRequestID(userID)
	req := &CancelReq{ReservationID: reservationID}
	_, err := client.CancelRPC(ctx, req)
	retries := 0
	for err != nil {
		sleep := utils.Backoff(retries)
		time.Sleep(time.Duration(sleep) * time.Second)
		_, err = client.CancelRPC(ctx, req)
		if err != nil {
			st := status.Convert(err)
			if st.Code() != codes.Unavailable && st.Code() != codes.DeadlineExceeded {
				return err
			}
		}
	}
	return nil
}
