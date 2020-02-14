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
func (c *Coordinator) StartSagaRPC(ctx context.Context, req *SagaReq) (*SagaResp, error) {
	vertices := make(map[VertexID]SagaVertex, len(req.GetVertices()))
	dag := make(map[VertexID]map[VertexID]SagaEdge, 0)

	// Populate vertices
	for id, v := range req.GetVertices() {
		tf := v.GetT()
		cf := v.GetC()
		vtx := SagaVertex{
			VertexID: VertexID(id),
			TFunc: SagaFunc{
				URL:       tf.GetUrl(),
				Method:    tf.GetMethod(),
				RequestID: c.logs.NewRequestID(),
				Body:      tf.GetBody(),
				Resp:      make(map[string]string, 0),
			},
			CFunc: SagaFunc{
				URL:       cf.GetUrl(),
				Method:    cf.GetMethod(),
				RequestID: c.logs.NewRequestID(),
				Body:      cf.GetBody(),
				Resp:      make(map[string]string, 0),
			},
			TransferFields: v.GetTransferFields(),
			Status:         NotReached,
		}
		if vtx.TFunc.Body == nil {
			vtx.TFunc.Body = make(map[string]string, 0)
		}
		if vtx.CFunc.Body == nil {
			vtx.CFunc.Body = make(map[string]string, 0)
		}
		vertices[VertexID(id)] = vtx
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

	sagaResp := &SagaResp{
		Vertices: verticesToProto(saga.Vertices),
	}

	if aborted {
		return sagaResp, ErrSagaAborted
	}
	if !finished {
		return sagaResp, ErrSagaUnfinished
	}
	return sagaResp, nil
}

func verticesToProto(vertices map[VertexID]SagaVertex) map[string]*Vertex {
	res := make(map[string]*Vertex, len(vertices))

	for id, v := range vertices {
		res[string(id)] = &Vertex{
			Id: string(id),
			T:  funcToProto(v.TFunc),
			C:  funcToProto(v.CFunc),
		}
	}

	return res
}

func protoToVertices(vertices map[string]*Vertex) map[VertexID]SagaVertex {
	res := make(map[VertexID]SagaVertex, len(vertices))

	for id, v := range vertices {
		res[VertexID(id)] = SagaVertex{
			VertexID: VertexID(id),
			TFunc:    protoToFunc(v.T),
			CFunc:    protoToFunc(v.C),
		}
	}

	return res
}

func funcToProto(f SagaFunc) *Func {
	body := make(map[string]string, len(f.Body)+len(f.Resp))

	for k, v := range f.Body {
		body[k] = v
	}

	for k, v := range f.Resp {
		body[k] = v
	}

	return &Func{
		Url:    f.URL,
		Method: f.Method,
		Body:   body,
	}
}

func protoToFunc(f *Func) SagaFunc {
	body := make(map[string]string, len(f.Body))

	for k, v := range f.Body {
		body[k] = v
	}

	return SagaFunc{
		URL:    f.Url,
		Method: f.Method,
		Body:   body,
	}
}
