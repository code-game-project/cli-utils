package modules

import (
	"fmt"
	"io"
	"strconv"

	"github.com/code-game-project/cli-utils/versions"
)

type ProviderLocal struct{}

func (p *ProviderLocal) Name() string {
	return "local"
}

func (p *ProviderLocal) ValidateProviderVars(providerVars map[string]any) []string {
	var errs []string
	_, containsPath := providerVars["path"]
	_, containsPaths := providerVars["paths"]
	if !containsPath && !containsPaths {
		errs = append(errs, "missing 'path' or 'paths' field")
	}
	if containsPath {
		if _, ok := providerVars["path"].(string); !ok {
			errs = append(errs, "value of 'path' field must be a string")
		}
	}
	if containsPaths {
		if list, ok := providerVars["paths"].([]any); ok {
			for _, p := range list {
				if _, ok := p.(string); !ok {
					errs = append(errs, "value of 'paths' field must be a string list")
					break
				}
			}
		} else {
			errs = append(errs, "value of 'paths' field must be a string list")
		}
	}
	return errs
}

func (p *ProviderLocal) FindExactVersion(providerVars map[string]any, version versions.Version) (versions.Version, error) {
	return version, nil
}

func (p *ProviderLocal) DownloadModuleBinary(target io.Writer, providerVars map[string]any, version versions.Version) error {
	panic("cannot download local modules")
}

func (m *Module) loadLocalModules() error {
	path, ok := m.providerVars["path"]
	if ok {
		err := m.loadLocalModulePath(path.(string))
		if err != nil {
			return err
		}
	}
	rawPaths, ok := m.providerVars["paths"]
	if ok {
		for _, p := range rawPaths.([]any) {
			err := m.loadLocalModulePath(p.(string))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Module) loadLocalModulePath(path string) error {
	info, err := execInfo(path)
	if err != nil {
		return fmt.Errorf("failed to receive module version of '%s': %w", path, err)
	}

	id := strconv.Itoa(len(m.installedExecutables))
	for _, v := range info.LibraryVersions["client"] {
		if _, ok := m.clientLibToModVersions[v.String()]; !ok {
			m.clientLibToModVersions[v.String()] = id
		}
	}
	for _, v := range info.LibraryVersions["server"] {
		if _, ok := m.serverLibToModVersions[v.String()]; !ok {
			m.serverLibToModVersions[v.String()] = id
		}
	}
	m.installedExecutables[id] = path
	return nil
}
