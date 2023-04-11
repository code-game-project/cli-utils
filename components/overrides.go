package components

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/code-game-project/cli-utils/config"
)

// component name -> version (e.g. CG/CGE/...) -> path to compatible binary
var componentOverrides map[string]map[string]string

func loadOverrides(componentName string) map[string]string {
	if componentOverrides == nil {
		overridesFile, err := os.Open(filepath.Join(config.ConfigDir(), "component_overrides.json"))
		if err != nil {
			return nil
		}
		defer overridesFile.Close()
		err = json.NewDecoder(overridesFile).Decode(&componentOverrides)
		if err != nil {
			return nil
		}
	}
	return componentOverrides[componentName]
}
