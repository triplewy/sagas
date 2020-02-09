package sagas

import (
	context "context"
	"fmt"
	"strconv"

	"google.golang.org/grpc"
)

// NewClient creates a new gRPC client to coordinator
func NewClient(addr string) CoordinatorClient {
	cc, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return NewCoordinatorClient(cc)
}

// BookRoom starts a new saga that makes a single transaction to book a hotel room
func BookRoom(c CoordinatorClient, userID, roomID string) error {
	vertices := map[string]*Vertex{
		"book": &Vertex{
			Id: "book",
			T: &Func{
				Url:    "http://localhost:51051/book",
				Method: "POST",
				Body: map[string]string{
					"userID": userID,
					"roomID": roomID,
				},
			},
			C: &Func{
				Url:    "http://localhost:51051/cancel",
				Method: "POST",
				Body: map[string]string{
					"userID": userID,
				},
			},
			TransferFields: []string{"reservationID"},
		},
	}

	edges := []*Edge{}

	msg := &SagaReq{
		Vertices: vertices,
		Edges:    edges,
	}

	ctx := context.Background()

	_, err := c.StartSagaRPC(ctx, msg)

	return err
}

// BookMultipleRooms starts a new saga that makes a multiple transactions to book hotel rooms
func BookMultipleRooms(c CoordinatorClient, userID string, roomIDs []string) error {
	vertices := make(map[string]*Vertex, len(roomIDs))

	for i, roomID := range roomIDs {
		id := fmt.Sprintf("book%s", strconv.Itoa(i))
		vertices[id] = &Vertex{
			Id: id,
			T: &Func{
				Url:    "http://localhost:51051/book",
				Method: "POST",
				Body: map[string]string{
					"userID": userID,
					"roomID": roomID,
				},
			},
			C: &Func{
				Url:    "http://localhost:51051/cancel",
				Method: "POST",
				Body: map[string]string{
					"userID": userID,
				},
			},
			TransferFields: []string{"reservationID"},
		}
	}

	edges := []*Edge{}

	msg := &SagaReq{
		Vertices: vertices,
		Edges:    edges,
	}

	ctx := context.Background()

	_, err := c.StartSagaRPC(ctx, msg)

	return err
}
