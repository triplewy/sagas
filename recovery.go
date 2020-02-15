package sagas

// Recover reads logs from disks and reconstructs dags in memory
func Recover(logs LogStore) map[string]Saga {
	sagas := make(map[string]Saga, 0)

	// Repopulate all sagas into memory
	for i := uint64(1); i <= logs.LastIndex(); i++ {
		log, err := logs.GetLog(i)
		if err != nil {
			// Error must be ErrLogIndexNotFound so we just skip the index
			continue
		}
		switch log.LogType {
		case InitLog:
			continue
		case GraphLog:
			if _, ok := sagas[log.SagaID]; ok {
				panic("multiple graphs with same sagaID")
			}
			saga := decodeSaga(log.Data)
			sagas[log.SagaID] = saga
		case VertexLog:
			saga, ok := sagas[log.SagaID]
			if !ok {
				panic("log of vertex has sagaID that does not exist")
			}
			vertex := decodeVertex(log.Data)
			if _, ok := saga.getVtx(vertex.Id); !ok {
				panic(ErrIDNotFound)
			}
			saga.Vertices.Set(vertex.Id, vertex)
		default:
			panic("unrecognized log type")
		}
	}

	// Add all sagas into coordinator
	return sagas
}
