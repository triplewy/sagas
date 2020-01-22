package sagas

import (
	"os"
	"path/filepath"
)

type Config struct {
	Path        string
	HotelsAddr  string
	AutoRecover bool
}

func DefaultConfig() *Config {
	return &Config{
		Path:        filepath.Join(os.TempDir(), "sagas"),
		HotelsAddr:  "localhost:50051",
		AutoRecover: true,
	}
}
