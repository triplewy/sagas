package sagas

import (
	"os"
	"path/filepath"
)

type Config struct {
	Path       string
	HotelsAddr string
}

func DefaultConfig() *Config {
	return &Config{
		Path:       filepath.Join(os.TempDir(), "sagas"),
		HotelsAddr: "localhost:50051",
	}
}
