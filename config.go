package sagas

import (
	"os"
	"path/filepath"
)

// Config for saga coordinator
type Config struct {
	Path            string
	HotelsAddr      string
	CoordinatorAddr string
	AutoRecover     bool
	InMemory        bool
}

// DefaultConfig provides default config for saga coordinator
func DefaultConfig() *Config {
	return &Config{
		Path:            filepath.Join(os.TempDir(), "sagas"),
		HotelsAddr:      ":50051",
		CoordinatorAddr: ":50050",
		AutoRecover:     true,
		InMemory:        true,
	}
}
