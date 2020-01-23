package sagas

import (
	"os"
	"path/filepath"
)

// Config for saga coordinator
type Config struct {
	Path        string
	HotelsAddr  string
	AutoRecover bool
}

// DefaultConfig provides default config for saga coordinator
func DefaultConfig() *Config {
	return &Config{
		Path:        filepath.Join(os.TempDir(), "sagas"),
		HotelsAddr:  "localhost:50051",
		AutoRecover: true,
	}
}
