package utils

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand"
	"net"

	"github.com/hashicorp/go-msgpack/codec"
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

// BaseTimeout is lowest timeout value
const BaseTimeout = 1

// MaxTimeout is maximum timeout value
const MaxTimeout = 10

// Backoff implements an exponential backoff algorithm with jitter.
// https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
func Backoff(retries int) int {
	return rand.Intn(min(MaxTimeout, BaseTimeout*int(math.Pow(2, float64(retries))))) + 1
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// DecodeMsgPack reverses the encode operation on a byte slice input
func DecodeMsgPack(buf []byte, out interface{}) error {
	r := bytes.NewBuffer(buf)
	hd := codec.MsgpackHandle{}
	dec := codec.NewDecoder(r, &hd)
	return dec.Decode(out)
}

// EncodeMsgPack writes an encoded object to a new bytes buffer
func EncodeMsgPack(in interface{}) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	hd := codec.MsgpackHandle{}
	enc := codec.NewEncoder(buf, &hd)
	err := enc.Encode(in)
	return buf, err
}

// BytesToUint64 converts bytes to an integer
func BytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// Uint64ToBytes converts a uint to a byte slice
func Uint64ToBytes(u uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, u)
	return buf
}
