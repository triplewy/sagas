package utils

import (
	"math"
	"math/rand"
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

const BaseTimeout = 1
const MaxTimeout = 10

// Backoff implements an exponential backoff algorithm with jitter.
// https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
func Backoff(retries int) int {
	return rand.Intn(minInt(MaxTimeout, BaseTimeout*int(math.Pow(2, float64(retries))))) + 1
}

func minInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}
