package sagas

// ProcessT runs a Vertex's  T
func (c *Coordinator) ProcessT(sagaID string, vertex Vertex) {
	// Sanity check on vertex's status
	if vertex.Status == Status_END_T {
		return
	}
	if !(vertex.Status == Status_START_T || vertex.Status == Status_NOT_REACHED) {
		panic(ErrInvalidSaga)
	}

	// Append to log
	data := encodeVertex(vertex)
	c.logs.AppendLog(sagaID, VertexLog, data)

	// Evaluate vertex's function
	f := vertex.T

	resp, err := HTTPReq(f.GetUrl(), f.GetMethod(), f.GetRequestId(), f.GetBody())
	status := Status_END_T
	if err != nil {
		// Error.Println(err)
		// Store error to output
		f.Resp["error"] = err.Error()
		// Set status to abort
		status = Status_ABORT
	} else {
		for k, v := range resp {
			f.Resp[k] = v
		}
		for _, k := range vertex.TransferFields {
			vertex.C.Body[k] = f.Resp[k]
		}
	}
	vertex.Status = status

	// Append to log
	c.logs.AppendLog(sagaID, VertexLog, encodeVertex(vertex))

	// Send newVertex to update chan for coordinator to update its map of sagas
	c.updateCh <- updateMsg{
		sagaID: sagaID,
		vertex: vertex,
	}
}

// ProcessC runs a Vertex's C
func (c *Coordinator) ProcessC(sagaID string, vertex Vertex) {
	// Sanity check on vertex's status
	if vertex.Status == Status_END_C {
		return
	}
	if !(vertex.Status == Status_START_T || vertex.Status == Status_END_T || vertex.Status == Status_START_C) {
		panic(ErrInvalidSaga)
	}
	// If vertex status is  Status_START_T, execute Status_END_T first
	if vertex.Status == Status_START_T {
		c.ProcessT(sagaID, vertex)
		return
	}

	// Now vertex must either be Status_END_T or Status_START_C. Append to log
	data := encodeVertex(vertex)
	c.logs.AppendLog(sagaID, VertexLog, data)

	// Evaluate vertex's function
	f := vertex.C

	resp, err := HTTPReq(f.GetUrl(), f.GetMethod(), f.GetRequestId(), f.GetBody())
	status := Status_END_C
	if err != nil {
		// Store error to output
		f.Resp["error"] = err.Error()
		// set status to startC because did not succeed
		status = Status_START_C
	} else {
		for k, v := range resp {
			f.Resp[k] = v
		}
	}
	vertex.Status = status

	// Append to log
	c.logs.AppendLog(sagaID, VertexLog, encodeVertex(vertex))

	// Send newVertex to update chan for coordinator to continue saga
	c.updateCh <- updateMsg{
		sagaID: sagaID,
		vertex: vertex,
	}
}
