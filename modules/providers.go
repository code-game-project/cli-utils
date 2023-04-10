package modules

import (
	"errors"
	"io"

	"github.com/code-game-project/cli-utils/versions"
)

var ErrVersionNotFound = errors.New("version not found")

var providers = map[string]provider{
	"github": &ProviderGithub{},
	"local":  &ProviderLocal{},
}

type provider interface {
	Name() string
	ValidateProviderVars(providerVars map[string]any) (errs []string)
	FindExactVersion(providerVars map[string]any, version versions.Version) (versions.Version, error)
	DownloadModuleBinary(target io.Writer, providerVars map[string]any, version versions.Version) error
}
