package modules

import (
	"errors"
	"io"

	"github.com/code-game-project/cli-utils/versions"
)

var ErrVersionNotFound = errors.New("version not found")

var providers = map[string]provider{
	"github": &ProviderGithub{},
}

type provider interface {
	ValidateProviderVars(providerVars map[string]string) (errs []string)
	FindExactVersion(providerVars map[string]string, version versions.Version) (versions.Version, error)
	DownloadModuleBinary(target io.Writer, providerVars map[string]string, version versions.Version) error
}
