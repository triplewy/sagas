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

// LocalSaga takes in a DAG with only 'true' or 'false' nodes that represent nodes
// that fail in the saga. To determine 'true' or 'false' nodes, the vertexes of the dag
// should be named '<id>1' to signify true or '<id>0' to signify false.
// When coordinator receives LocalSaga req, it performs all vertex requests locally
func LocalSaga(c CoordinatorClient, dag map[string]map[string]struct{}) error {
	vertices := make(map[string]*Vertex, len(dag))
	var edges []*Edge

	for id, set := range dag {
		status := id[len(id)-1:]
		vertices[id] = &Vertex{
			Id: id,
			T: &Func{
				Method: "LOCAL",
				Body: map[string]string{
					"success": status,
				},
			},
			C: &Func{
				Method: "LOCAL",
			},
			TransferFields: []string{"success"},
		}
		for v := range set {
			edges = append(edges, &Edge{
				StartId: id,
				EndId:   v,
			})
		}
	}

	msg := &SagaReq{
		Vertices: vertices,
		Edges:    edges,
	}

	ctx := context.Background()

	resp, err := c.StartSagaRPC(ctx, msg)

	fmt.Printf("%#v\n", protoToVertices(resp.GetVertices()))

	return err
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
