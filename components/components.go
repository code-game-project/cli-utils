package components

import (
	"errors"
	"fmt"

	"github.com/code-game-project/cli-utils/feedback"
	"github.com/code-game-project/cli-utils/versions"
)

const FeedbackPkg = feedback.Package("components")

func CGEParser(cgeVersion versions.Version) (binPath string, err error) {
	return component("cge-parser", cgeVersion)
}

func CGDebug(cgVersion versions.Version) (binPath string, err error) {
	return component("cg-debug", cgVersion)
}

func component(name string, version versions.Version) (string, error) {
	compatibleOverride, binPath, err := findLatestCompatibleVersionSupportedByComponentInOverrides(name, version)
	if err != nil && !errors.Is(err, ErrVersionNotFound) {
		return "", fmt.Errorf("find latest compatible version supported by %s in overrides: %s", name, err)
	}

	comp, sup, err := findLatestCompatibleVersionSupportedByComponent(name, version)
	if err != nil {
		if compatibleOverride != nil {
			return binPath, nil
		}
		return "", fmt.Errorf("find latest compatible version supported by '%s': %s", name, err)
	}
	if versions.Compare(compatibleOverride, sup) < 1 {
		return binPath, nil
	}

	return install(name, comp)
}
