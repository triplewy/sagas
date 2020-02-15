package sagas

import (
	"context"
	"errors"
	"log"
	"net"

	"google.golang.org/grpc"
)

// Errors encountered by client API
var (
	ErrSagaAborted    = errors.New("saga aborted during execution")
	ErrSagaUnfinished = errors.New("saga did not finish during execution. Should never happen")
)

// NewServer starts a new gRPC server for sagas
func NewServer(addr string, c *Coordinator) *grpc.Server {
	s := grpc.NewServer(grpc.UnaryInterceptor(serverInterceptor))

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	RegisterCoordinatorServer(s, c)

	go s.Serve(lis)

	return s
}

// Authorization unary interceptor function to handle authorize per RPC call
func serverInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Println(info.FullMethod)

	// Calls the handler
	return handler(ctx, req)
}

// StartSagaRPC starts a saga
func (c *Coordinator) StartSagaRPC(ctx context.Context, req *SagaMsg) (*SagaMsg, error) {
	saga := protoToSaga(req)
	// Don't forget to set sagaID
	sagaID := c.logs.NewSagaID()
	saga.ID = sagaID

	replyCh := make(chan Saga, 1)
	c.createCh <- createMsg{
		saga:    saga,
		replyCh: replyCh,
	}

	replySaga := <-replyCh

	sagaResp := sagaToProto(replySaga)

	return sagaResp, nil
}

func protoToSaga(req *SagaMsg) Saga {
	vertices := make(map[string]Vertex, len(req.GetVertices()))
	dag := make(map[string]map[string][]string, 0)

	// Populate vertices
	for id, vtx := range req.GetVertices() {
		if vtx == nil {
			continue
		}
		if vtx.T.Body == nil {
			vtx.T.Body = make(map[string]string, 0)
		}
		if vtx.T.Resp == nil {
			vtx.T.Resp = make(map[string]string, 0)
		}
		if vtx.C.Body == nil {
			vtx.C.Body = make(map[string]string, 0)
		}
		if vtx.C.Resp == nil {
			vtx.C.Resp = make(map[string]string, 0)
		}
		vertices[id] = *vtx
		dag[id] = make(map[string][]string, 0)
	}

	// Populate dag
	for _, edge := range req.GetEdges() {
		dag[edge.GetStartId()][edge.GetEndId()] = edge.GetTransferFields()
	}

	return NewSaga(vertices, dag)
}

func sagaToProto(saga Saga) *SagaMsg {
	// Populate vertices
	vertices := make(map[string]*Vertex, len(saga.Vertices))
	saga.Vertices.IterCb(func(k string, v interface{}) {
		vtx := v.(Vertex)
		vertices[k] = &vtx
	})

	// Populate edges
	saga.dagMtx.RLock()
	defer saga.dagMtx.RUnlock()

	var edges []*Edge
	for k, children := range saga.DAG {
		for c, fields := range children {
			edges = append(edges, &Edge{
				StartId:        k,
				EndId:          c,
				TransferFields: fields,
			})
		}
	}
	return &SagaMsg{
		Id:       saga.ID,
		Vertices: vertices,
		Edges:    edges,
	}
}
