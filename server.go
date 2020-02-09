package sagas

import (
	"context"
	"errors"
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
	s := grpc.NewServer()

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	RegisterCoordinatorServer(s, c)

	go s.Serve(lis)

	return s
}

// StartSagaRPC starts a saga
func (c *Coordinator) StartSagaRPC(ctx context.Context, req *SagaReq) (*SagaResp, error) {
	vertices := make(map[VertexID]SagaVertex, len(req.GetVertices()))
	dag := make(map[VertexID]map[VertexID]SagaEdge, 0)

	// Populate vertices
	for id, v := range req.GetVertices() {
		tf := v.GetT()
		cf := v.GetC()
		vertices[VertexID(id)] = SagaVertex{
			VertexID: VertexID(id),
			TFunc: SagaFunc{
				URL:       tf.GetUrl(),
				Method:    tf.GetUrl(),
				RequestID: c.logs.NewRequestID(),
				Body:      tf.GetBody(),
				Resp:      make(map[string]string, 0),
			},
			CFunc: SagaFunc{
				URL:       cf.GetUrl(),
				Method:    cf.GetUrl(),
				RequestID: c.logs.NewRequestID(),
				Body:      cf.GetBody(),
				Resp:      make(map[string]string, 0),
			},
			TransferFields: v.GetTransferFields(),
			Status:         NotReached,
		}
		dag[VertexID(id)] = make(map[VertexID]SagaEdge, 0)
	}

	// Populate dag
	for _, edge := range req.GetEdges() {
		dag[VertexID(edge.GetStartId())][VertexID(edge.GetEndId())] = SagaEdge{Fields: edge.GetTransferFields()}
	}

	saga := Saga{DAG: dag, Vertices: vertices}
	sagaID := c.logs.NewSagaID()

	replyCh := make(chan Saga, 1)
	c.createCh <- createMsg{
		sagaID:  sagaID,
		saga:    saga,
		replyCh: replyCh,
	}
	replySaga := <-replyCh

	finished, aborted := CheckFinishedOrAbort(replySaga.Vertices)

	if aborted {
		return nil, ErrSagaAborted
	}
	if !finished {
		return nil, ErrSagaUnfinished
	}
	return &SagaResp{}, nil
}
