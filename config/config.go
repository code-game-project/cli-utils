package config

import (
	"path/filepath"

	"github.com/adrg/xdg"

	"github.com/code-game-project/cli-utils/feedback"
)

const FeedbackPkg = feedback.Package("config")

func ConfigDir() string {
	return filepath.Join(xdg.ConfigHome, "codegame")
}
