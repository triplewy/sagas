package sagas

import (
	"net"
)

// AvailableAddr gets an available address on localhost
func AvailableAddr() string {
	// Create a new server without specifying a port which will result in an open port being chosen
	server, err := net.Listen("tcp", ":0")
	// If there's an error it likely means no ports are available or something else prevented finding an open port
	if err != nil {
		panic(err)
	}
	defer server.Close()
	// Get the host string in the format "127.0.0.1:4444"
	return server.Addr().String()
}
