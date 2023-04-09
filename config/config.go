package config

import (
	"path/filepath"

	"github.com/adrg/xdg"
)

func ConfigDir() string {
	return filepath.Join(xdg.ConfigHome, "codegame")
}
